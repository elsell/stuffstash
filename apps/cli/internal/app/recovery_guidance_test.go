package app_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/pairingkeys"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/presentation"
	"github.com/stuffstash/stuff-stash/cli/internal/app"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type endedPairing struct {
	*pairingFixture
	state ports.PairingState
}

func (f endedPairing) Poll(context.Context, ports.PairingChallenge) (ports.PairingState, error) {
	return f.state, nil
}

func TestPairingFailurePreservesRegistrationOrRotationRecovery(t *testing.T) {
	for _, rotate := range []bool{false, true} {
		for _, state := range []ports.PairingState{ports.PairingPending, ports.PairingConsumed} {
			f := &pairingFixture{now: time.Now()}
			old := ports.ConnectorRegistration{Server: "https://stash.example", TenantID: "t", InventoryID: "i", ConnectorID: "connector", Credential: "old-private"}
			if rotate {
				f.saved = &old
			}
			var diagnostic bytes.Buffer
			r := app.ConnectorRegistrar{API: endedPairing{f, state}, Credentials: f, Keys: pairingkeys.Keys{}, Clock: f, Waiter: f, Output: presentation.Output{Stdout: &diagnostic, Stderr: &diagnostic}, PollInterval: time.Second}
			var err error
			if rotate {
				err = r.Rotate(context.Background(), old.Server, old.ConnectorID)
			} else {
				err = r.Register(context.Background(), old.Server, "Garage", []ports.PairingCandidate{{ID: "printer"}})
			}
			var failure *ports.Error
			if !errors.As(err, &failure) || failure.Category != "pairing" {
				t.Fatalf("wrong pairing failure: %v", err)
			}
			wanted := "connectors print register"
			if rotate {
				wanted = `connectors print rotate --connector "connector"`
				if strings.Contains(failure.Message, "register") {
					t.Fatal("rotation recovery recommends new registration")
				}
			}
			if !strings.Contains(failure.Message, wanted) || !strings.Contains(failure.Message, `--server "https://stash.example"`) {
				t.Fatalf("lost recovery action or server: %s", failure.Message)
			}
			if f.active || f.consumed || (rotate && f.saved.Credential != old.Credential) || (!rotate && f.saved != nil) {
				t.Fatal("failed pairing changed stored credential or activated")
			}
		}
	}
}
func TestUnknownOptionDoesNotExposeSecretOrTerminalControls(t *testing.T) {
	for _, name := range []string{"private-api-token", "secret\n\x1b[2J"} {
		_, err := app.Parse([]string{"assets", "list", "--" + name}, func(string) string { return "" })
		var failure *ports.Error
		if !errors.As(err, &failure) || failure.Category != "usage" {
			t.Fatalf("wrong unknown-option category: %v", err)
		}
		if strings.Contains(failure.Message, name) || strings.ContainsAny(failure.Message, "\x1b\n") || !strings.Contains(failure.Message, "--help") {
			t.Fatalf("unsafe option diagnostic: %q", failure.Message)
		}
	}
}
