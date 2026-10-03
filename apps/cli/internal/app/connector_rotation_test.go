package app_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/pairingkeys"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/presentation"
	"github.com/stuffstash/stuff-stash/cli/internal/app"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func TestRotationPreservesExistingIdentityBeforeSaving(t *testing.T) {
	for _, mismatch := range []string{"", "tenant", "inventory", "connector", "server", "storage"} {
		t.Run(mismatch, func(t *testing.T) {
			old := ports.ConnectorRegistration{Server: "https://stash.example", TenantID: "t", InventoryID: "i", ConnectorID: "connector", Credential: "old-private", ExpiresAt: time.Now().Add(-time.Hour)}
			switch mismatch {
			case "tenant":
				old.TenantID = "other"
			case "inventory":
				old.InventoryID = "other"
			case "connector":
				old.ConnectorID = "other"
			case "server":
				old.Server = "https://other.example"
			}
			f := &pairingFixture{now: time.Now(), saved: &old, storageFull: mismatch == "storage"}
			var output bytes.Buffer
			r := app.ConnectorRegistrar{API: f, Credentials: f, Keys: pairingkeys.Keys{}, Clock: f, Waiter: f, Output: presentation.Output{Stdout: &output, Stderr: &output}, PollInterval: time.Second}
			err := r.Rotate(context.Background(), old.Server, old.ConnectorID)
			if !f.request.Rotation || len(f.request.Candidates) != 0 {
				t.Fatal("rotation required hardware discovery")
			}
			if mismatch == "" {
				if err != nil || !f.active || f.saved.Credential == old.Credential {
					t.Fatalf("rotation failed: %v", err)
				}
			} else if err == nil || f.active || f.saved.Credential != old.Credential {
				t.Fatal("changed identity or failed save replaced credential")
			}
			for _, secret := range []string{"old-private", "private-poll", "machine-secret"} {
				if strings.Contains(output.String(), secret) {
					t.Fatal("rotation leaked secret")
				}
			}
			if !strings.Contains(output.String(), "connectorId="+old.ConnectorID) || !strings.Contains(output.String(), "inventoryId="+old.InventoryID) || !strings.Contains(output.String(), "tenantId="+old.TenantID) {
				t.Fatal("approval destination lost rotation identity")
			}
		})
	}
}
