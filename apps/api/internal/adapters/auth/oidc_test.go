package auth

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestOIDCAuthenticatorAcceptsVerifiedBearerToken(t *testing.T) {
	authenticator := NewOIDCAuthenticator(&fakeTokenVerifier{
		token: VerifiedToken{Issuer: "https://accounts.google.com", Subject: "google-user-123", Email: "Owner@Example.COM", EmailVerified: true},
	})

	principal, err := authenticator.Authenticate(context.Background(), "Bearer valid-id-token")
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}

	if principal.ID.String() != "oidc_os13uwZQU011TSXUcOEuemPs1E5sdAPRkHQFKcmVQ6w" {
		t.Fatalf("unexpected principal ID %q", principal.ID.String())
	}
	if principal.Email.String() != "owner@example.com" {
		t.Fatalf("unexpected principal email %q", principal.Email.String())
	}
}

func TestOIDCAuthenticatorIgnoresUnverifiedEmail(t *testing.T) {
	authenticator := NewOIDCAuthenticator(&fakeTokenVerifier{
		token: VerifiedToken{Issuer: "https://accounts.google.com", Subject: "google-user-123", Email: "owner@example.com"},
	})

	principal, err := authenticator.Authenticate(context.Background(), "Bearer valid-id-token")
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if principal.Email.String() != "" {
		t.Fatalf("expected unverified email to be omitted, got %q", principal.Email.String())
	}
}

func TestOIDCAuthenticatorRejectsMalformedAuthorizationHeader(t *testing.T) {
	authenticator := NewOIDCAuthenticator(&fakeTokenVerifier{
		token: VerifiedToken{Issuer: "https://accounts.google.com", Subject: "google-user-123"},
	})

	tests := []string{
		"",
		"Basic valid-id-token",
		"Bearer",
		"Bearer ",
		"Bearer one two",
	}

	for _, authorization := range tests {
		t.Run(authorization, func(t *testing.T) {
			_, err := authenticator.Authenticate(context.Background(), authorization)
			if !errors.Is(err, ports.ErrUnauthenticated) {
				t.Fatalf("expected unauthenticated, got %v", err)
			}
		})
	}
}

func TestOIDCAuthenticatorRejectsVerifierFailure(t *testing.T) {
	authenticator := NewOIDCAuthenticator(&fakeTokenVerifier{
		err: errors.New("verification failed"),
	})

	_, err := authenticator.Authenticate(context.Background(), "Bearer invalid-id-token")
	if !errors.Is(err, ports.ErrUnauthenticated) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}

func TestOIDCAuthenticatorSupportsProviderSpecificSubjectCharacters(t *testing.T) {
	authenticator := NewOIDCAuthenticator(&fakeTokenVerifier{
		token: VerifiedToken{Issuer: "https://issuer.example", Subject: "provider|subject/with:chars"},
	})

	principal, err := authenticator.Authenticate(context.Background(), "Bearer valid-id-token")
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if principal.ID.String() == "" || strings.ContainsAny(principal.ID.String(), "|/:") {
		t.Fatalf("expected safe internal principal ID, got %q", principal.ID.String())
	}
}

func TestOIDCAuthenticatorRejectsEmptyIssuerOrSubject(t *testing.T) {
	tests := []VerifiedToken{
		{Issuer: "", Subject: "subject"},
		{Issuer: "https://issuer.example", Subject: ""},
	}

	for _, token := range tests {
		authenticator := NewOIDCAuthenticator(&fakeTokenVerifier{token: token})
		_, err := authenticator.Authenticate(context.Background(), "Bearer valid-id-token")
		if !errors.Is(err, ports.ErrUnauthenticated) {
			t.Fatalf("expected unauthenticated, got %v", err)
		}
	}
}

type fakeTokenVerifier struct {
	token VerifiedToken
	err   error
}

func (f *fakeTokenVerifier) Verify(_ context.Context, _ string) (VerifiedToken, error) {
	return f.token, f.err
}
