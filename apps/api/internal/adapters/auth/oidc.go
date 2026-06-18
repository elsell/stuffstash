package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type VerifiedToken struct {
	Issuer  string
	Subject string
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
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return OIDCAuthenticator{}, err
	}

	return NewOIDCAuthenticator(oidcTokenVerifier{
		verifier: provider.Verifier(&oidc.Config{ClientID: clientID}),
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

	return identity.Principal{ID: principalID}, nil
}

func bearerToken(authorizationHeader string) (string, bool) {
	parts := strings.Fields(authorizationHeader)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}

	return parts[1], true
}

type oidcTokenVerifier struct {
	verifier *oidc.IDTokenVerifier
}

func (v oidcTokenVerifier) Verify(ctx context.Context, rawToken string) (VerifiedToken, error) {
	token, err := v.verifier.Verify(ctx, rawToken)
	if err != nil {
		return VerifiedToken{}, err
	}

	return VerifiedToken{Issuer: token.Issuer, Subject: token.Subject}, nil
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
