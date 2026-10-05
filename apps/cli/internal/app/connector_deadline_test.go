package app_test

import (
	"bytes"
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/pairingkeys"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/presentation"
	"github.com/stuffstash/stuff-stash/cli/internal/app"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strings"
	"testing"
	"time"
)

func TestPairingDeadlineProtectsStoredCredentials(t *testing.T) {
	now := time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)
	for _, tc := range []struct {
		name          string
		deadline      time.Time
		saveDelay     time.Duration
		category      string
		saved, active bool
	}{
		{name: "missing", category: "protocol"},
		{name: "expired", deadline: now.Add(-time.Nanosecond), category: "pairing"},
		{name: "exact boundary", deadline: now, category: "pairing"},
		{name: "future", deadline: now.Add(time.Nanosecond), saved: true, active: true},
		{name: "expires during save", deadline: now.Add(time.Second), saveDelay: time.Second, category: "activation", saved: true},
	} {
		for _, rotate := range []bool{false, true} {
			name := tc.name + "/register"
			if rotate {
				name = tc.name + "/rotate"
			}
			t.Run(name, func(t *testing.T) {
				prior := ports.ConnectorRegistration{Server: "https://stash.example", TenantID: "t", InventoryID: "i", ConnectorID: "connector", Credential: "old-private", ExpiresAt: now.Add(time.Hour)}
				f := &pairingFixture{now: now, approved: true, activationDeadline: &tc.deadline, saveDelay: tc.saveDelay}
				if rotate {
					f.saved = &prior
				}
				var output bytes.Buffer
				registrar := app.ConnectorRegistrar{API: f, Credentials: f, Keys: pairingkeys.Keys{}, Clock: f, Waiter: f, Output: presentation.Output{Stdout: &output, Stderr: &output}, PollInterval: time.Second}
				var err error
				if rotate {
					err = registrar.Rotate(context.Background(), prior.Server, prior.ConnectorID)
				} else {
					err = registrar.Register(context.Background(), prior.Server, "Garage", []ports.PairingCandidate{{ID: "device"}})
				}
				var failure *ports.Error
				if tc.category == "" {
					if err != nil {
						t.Fatal(err)
					}
				} else if !errors.As(err, &failure) || failure.Category != tc.category {
					t.Fatalf("category: %v", err)
				}
				if f.active != tc.active {
					t.Fatalf("activation=%v", f.active)
				}
				if tc.saved {
					if f.saved == nil || f.saved.Credential != "machine-secret" || !f.saved.ActivationDeadline.Equal(tc.deadline) {
						t.Fatal("replacement or deadline was not saved")
					}
				} else if rotate {
					if f.saved == nil || *f.saved != prior {
						t.Fatal("old credential changed before save")
					}
				} else if f.saved != nil {
					t.Fatal("expired registration saved")
				}
				if tc.saveDelay > 0 && (err == nil || !strings.Contains(err.Error(), "saved") || !strings.Contains(err.Error(), "rotate")) {
					t.Fatalf("missing recovery after save: %v", err)
				}
				text := output.String()
				if err != nil {
					text += err.Error()
				}
				for _, secret := range []string{"old-private", "machine-secret", "private-poll"} {
					if strings.Contains(text, secret) {
						t.Fatal("secret exposed")
					}
				}
			})
		}
	}
}
