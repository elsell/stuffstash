//go:build windows

package credentials

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"github.com/zalando/go-keyring"
)

func TestWindowsKeyringBoundary(t *testing.T) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		t.Fatal("Cannot create test identity")
	}
	server := "https://cli-keyring-" + hex.EncodeToString(id[:]) + ".invalid"
	store := Keyring{}
	ctx := context.Background()
	t.Cleanup(func() {
		if err := store.Delete(ctx, server); err != nil {
			t.Error("Cannot remove synthetic test session")
		}
	})
	s := ports.Session{Server: server, Issuer: "https://issuer.invalid", Subject: "synthetic", ClientID: "test", IDToken: "synthetic-id", RefreshToken: "synthetic-refresh", ExpiresAt: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)}
	if err := store.Save(ctx, s); err != nil {
		t.Fatal("Cannot save synthetic session to Windows Credential Manager")
	}
	got, err := store.Load(ctx, server)
	if err != nil || got != s {
		t.Fatal("Stored session did not round trip")
	}
	if _, err := store.Load(ctx, server+"/other"); !errors.Is(err, ports.ErrNotLoggedIn) {
		t.Fatal("Other server session was not isolated")
	}
	s.Server = "https://other.invalid"
	wrong, _ := json.Marshal(s)
	for _, value := range []string{"invalid JSON", string(wrong)} {
		if err := keyring.Set(service, server, value); err != nil {
			t.Fatal("Cannot prepare adversarial stored session")
		}
		if _, err := store.Load(ctx, server); !errors.Is(err, ports.ErrNotLoggedIn) {
			t.Fatal("Accepted malformed or mismatched stored session")
		}
	}
	if err := store.Delete(ctx, server); err != nil {
		t.Fatal("Cannot delete synthetic session")
	}
	if _, err := store.Load(ctx, server); !errors.Is(err, ports.ErrNotLoggedIn) {
		t.Fatal("Deleted session remained available")
	}
	if err := store.Delete(ctx, server); err != nil {
		t.Fatal("Deleting absent session did not succeed")
	}
}
