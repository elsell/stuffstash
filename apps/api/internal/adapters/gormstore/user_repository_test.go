package gormstore

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
)

func TestStorePersistsUsersByPrincipalID(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	email, ok := identity.NewEmail("owner@example.test")
	if !ok {
		t.Fatalf("expected valid email")
	}
	user, ok := identity.NewUser(identity.PrincipalID("owner"), email)
	if !ok {
		t.Fatalf("expected valid user")
	}
	if err := store.SaveUser(ctx, user); err != nil {
		t.Fatalf("save user: %v", err)
	}

	users, err := store.UsersByID(ctx, []identity.PrincipalID{"owner", "missing"})
	if err != nil {
		t.Fatalf("load users: %v", err)
	}
	if users["owner"].Email.String() != "owner@example.test" {
		t.Fatalf("expected user email to round trip, got %+v", users)
	}
	if _, ok := users["missing"]; ok {
		t.Fatalf("expected missing user to be omitted, got %+v", users)
	}

	withoutEmail, ok := identity.NewUser(identity.PrincipalID("owner"), identity.Email(""))
	if !ok {
		t.Fatalf("expected valid user without email")
	}
	if err := store.SaveUser(ctx, withoutEmail); err != nil {
		t.Fatalf("save user without email: %v", err)
	}
	users, err = store.UsersByID(ctx, []identity.PrincipalID{"owner"})
	if err != nil {
		t.Fatalf("reload users: %v", err)
	}
	if users["owner"].Email.String() != "owner@example.test" {
		t.Fatalf("expected empty profile update to preserve email, got %+v", users)
	}
}
