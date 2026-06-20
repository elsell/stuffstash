package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestLocalDevAuthenticatorAcceptsOptionalEmailFixture(t *testing.T) {
	principal, err := NewLocalDevAuthenticator().Authenticate(context.Background(), "Bearer dev:user-one:User@One.Example")
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if principal.ID.String() != "user-one" || principal.Email.String() != "user@one.example" {
		t.Fatalf("unexpected principal: %+v", principal)
	}
}

func TestLocalDevAuthenticatorRejectsInvalidEmailFixture(t *testing.T) {
	_, err := NewLocalDevAuthenticator().Authenticate(context.Background(), "Bearer dev:user-one:not-email")
	if !errors.Is(err, ports.ErrUnauthenticated) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}
