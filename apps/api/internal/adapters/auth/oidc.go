package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type VerifiedToken struct {
	Issuer        string
	Subject       string
	Email         string
	EmailVerified bool
}

type TokenVerifier interface {
	Verify(ctx context.Context, rawToken string) (VerifiedToken, error)
}

type OIDCAuthenticator struct {
	verifier TokenVerifier
}

func NewOIDCAuthenticator(verifier TokenVerifier) OIDCAuthenticator {
	return OIDCAuthenticator{verifier: verifier}
}

func NewOIDCAuthenticatorFromIssuer(ctx context.Context, issuer string, clientID string) (OIDCAuthenticator, error) {
	return NewOIDCAuthenticatorFromIssuerForClientIDs(ctx, issuer, []string{clientID})
}

func NewOIDCAuthenticatorFromIssuerForClientIDs(ctx context.Context, issuer string, clientIDs []string) (OIDCAuthenticator, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return OIDCAuthenticator{}, err
	}
	allowedClientIDs := normalizeClientIDs(clientIDs)
	if len(allowedClientIDs) == 0 {
		return OIDCAuthenticator{}, errors.New("at least one oidc client id is required")
	}

	return NewOIDCAuthenticator(oidcTokenVerifier{
		verifier:         provider.Verifier(&oidc.Config{SkipClientIDCheck: true}),
		allowedClientIDs: allowedClientIDs,
	}), nil
}

func (a OIDCAuthenticator) Authenticate(ctx context.Context, authorizationHeader string) (identity.Principal, error) {
	rawToken, ok := bearerToken(authorizationHeader)
	if !ok {
		return identity.Principal{}, ports.ErrUnauthenticated
	}

	token, err := a.verifier.Verify(ctx, rawToken)
	if err != nil {
		return identity.Principal{}, ports.ErrUnauthenticated
	}

	principalID, ok := identity.NewPrincipalID(oidcPrincipalID(token.Issuer, token.Subject))
	if !ok {
		return identity.Principal{}, ports.ErrUnauthenticated
	}

	principal := identity.Principal{ID: principalID}
	if token.EmailVerified {
		if email, ok := identity.NewEmail(token.Email); ok {
			principal.Email = email
		}
	}

	return principal, nil
}

func bearerToken(authorizationHeader string) (string, bool) {
	parts := strings.Fields(authorizationHeader)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}

	return parts[1], true
}

type oidcTokenVerifier struct {
	verifier         *oidc.IDTokenVerifier
	allowedClientIDs map[string]struct{}
}

func (v oidcTokenVerifier) Verify(ctx context.Context, rawToken string) (VerifiedToken, error) {
	token, err := v.verifier.Verify(ctx, rawToken)
	if err != nil {
		return VerifiedToken{}, err
	}
	if !v.allowsAudience(token.Audience) {
		return VerifiedToken{}, errors.New("oidc token audience is not allowed")
	}

	claims := struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}{}
	if err := token.Claims(&claims); err != nil {
		return VerifiedToken{}, err
	}

	return VerifiedToken{Issuer: token.Issuer, Subject: token.Subject, Email: claims.Email, EmailVerified: claims.EmailVerified}, nil
}

func (v oidcTokenVerifier) allowsAudience(audiences []string) bool {
	for _, audience := range audiences {
		if _, ok := v.allowedClientIDs[audience]; ok {
			return true
		}
	}
	return false
}

func normalizeClientIDs(clientIDs []string) map[string]struct{} {
	allowed := map[string]struct{}{}
	for _, clientID := range clientIDs {
		clientID = strings.TrimSpace(clientID)
		if clientID == "" {
			continue
		}
		allowed[clientID] = struct{}{}
	}
	return allowed
}

func oidcPrincipalID(issuer string, subject string) string {
	issuer = strings.TrimSpace(issuer)
	subject = strings.TrimSpace(subject)
	if issuer == "" || subject == "" {
		return ""
	}

	sum := sha256.Sum256([]byte(issuer + "\x00" + subject))
	return "oidc_" + base64.RawURLEncoding.EncodeToString(sum[:])
}
