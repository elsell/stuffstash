package app_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/pairingkeys"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/presentation"
	"github.com/stuffstash/stuff-stash/cli/internal/app"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type pairingFixture struct {
	now                        time.Time
	request                    ports.PairingRequest
	approved, consumed, active bool
	saved                      *ports.ConnectorRegistration
	storageFull                bool
}

func (f *pairingFixture) Now() time.Time { return f.now }
func (f *pairingFixture) Wait(ctx context.Context, d time.Duration) error {
	f.now = f.now.Add(d)
	f.approved = true
	return ctx.Err()
}
func (f *pairingFixture) Start(_ context.Context, r ports.PairingRequest) (ports.PairingChallenge, error) {
	f.request = r
	return ports.PairingChallenge{ID: "pair", PollToken: "private-poll", UserCode: "ABCD1234", VerificationURL: "https://stash.example/print-connectors/pair/pair", ExpiresAt: f.now.Add(time.Minute)}, nil
}
func (f *pairingFixture) Poll(context.Context, ports.PairingChallenge) (ports.PairingState, error) {
	if f.consumed {
		return ports.PairingConsumed, nil
	}
	if f.approved {
		return ports.PairingApproved, nil
	}
	return ports.PairingPending, nil
}
func (f *pairingFixture) Exchange(_ context.Context, c ports.PairingChallenge, signature []byte) (ports.ConnectorRegistration, error) {
	if !f.approved || f.consumed || !ed25519.Verify(f.request.PublicKey, []byte("stuffstash-print-pairing-v1\n"+c.ID+"\n"+c.PollToken), signature) {
		return ports.ConnectorRegistration{}, errors.New("denied")
	}
	f.consumed = true
	return ports.ConnectorRegistration{Server: "https://stash.example", TenantID: "t", InventoryID: "i", ConnectorID: "connector", Credential: "machine-secret", ExpiresAt: f.now.Add(time.Hour)}, nil
}
func (f *pairingFixture) Activate(_ context.Context, r ports.ConnectorRegistration, _ string) error {
	if f.saved == nil || f.saved.Credential != r.Credential {
		return errors.New("credential not durably saved")
	}
	f.active = true
	return nil
}
func (f *pairingFixture) Save(_ context.Context, r ports.ConnectorRegistration) error {
	if f.storageFull {
		return errors.New("storage full")
	}
	f.saved = &r
	return nil
}
func (f *pairingFixture) Load(context.Context, string, string) (ports.ConnectorRegistration, error) {
	if f.saved == nil {
		return ports.ConnectorRegistration{}, ports.ErrConnectorNotRegistered
	}
	return *f.saved, nil
}
func (f *pairingFixture) Delete(context.Context, string, string) error { f.saved = nil; return nil }

func TestRegisterConnectorPersistsBeforeActivationAndNeverOutputsSecrets(t *testing.T) {
	for _, storageFull := range []bool{false, true} {
		t.Run(map[bool]string{false: "successful", true: "save-failure"}[storageFull], func(t *testing.T) {
			f := &pairingFixture{now: time.Now(), storageFull: storageFull}
			var output bytes.Buffer
			runner := app.ConnectorRegistrar{API: f, Credentials: f, Keys: pairingkeys.Keys{}, Clock: f, Waiter: f, Output: presentation.Output{Stdout: &output, Stderr: &output}, PollInterval: time.Second}
			err := runner.Register(context.Background(), "https://stash.example", "Garage", []ports.PairingCandidate{{ID: "device", Name: "Brother", AdapterID: "brother-ql800", DeviceID: "physical"}})
			if storageFull {
				if err == nil || f.active {
					t.Fatal("failed save activated connector")
				}
			} else if err != nil || !f.active {
				t.Fatalf("registration failed: %v", err)
			}
			for _, secret := range []string{"private-poll", "machine-secret"} {
				if strings.Contains(output.String(), secret) {
					t.Fatal("output leaked secret")
				}
			}
			if !strings.Contains(output.String(), "ABCD1234") {
				t.Fatal("approval code missing")
			}
		})
	}
}
