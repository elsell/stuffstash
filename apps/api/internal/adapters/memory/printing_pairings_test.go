package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPairingExchangeIsSingleUseAndPendingGrantsDenyIssuance(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	now := time.Now().UTC()
	scope := printing.Scope{TenantID: "tenant", InventoryID: "inventory"}
	store.inventories["inventory"] = inventory.Inventory{ID: "inventory", TenantID: "tenant", LifecycleState: inventory.LifecycleStateActive}
	p := printing.Pairing{ID: "pair", State: printing.PairingPending, CodeHash: "code", ExpiresAt: now.Add(time.Minute)}
	if err := store.CreatePrintPairing(ctx, p); err != nil {
		t.Fatal(err)
	}
	c := printing.Connector{ID: "connector", Scope: scope, State: printing.ConnectorPending, Generation: 1, CredentialVersion: 1}
	approval := ports.PairingApproval{PairingID: p.ID, CodeHash: p.CodeHash, Now: now, Registration: ports.ConnectorRegistration{Connector: c}, Audit: audit.Record{ID: "approved"}}
	if _, err := store.ApprovePrintPairing(ctx, approval); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ConsumePrintPairing(ctx, ports.PairingExchange{PairingID: p.ID, Now: now, CredentialHash: "hash", CredentialExpiresAt: now.Add(time.Hour), ActivationDeadline: now.Add(time.Minute), Audit: audit.Record{ID: "issued"}}); err == nil {
		t.Fatal("pending grants issued credential")
	}
	if err := store.SynchronizePrintConnector(ctx, scope, c.ID, func(context.Context, ports.ConnectorRegistration, []printing.Printer) error { return nil }); err != nil {
		t.Fatal(err)
	}
	var successes atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := store.ConsumePrintPairing(ctx, ports.PairingExchange{PairingID: p.ID, Now: now, CredentialHash: "hash", CredentialExpiresAt: now.Add(time.Hour), ActivationDeadline: now.Add(time.Minute), Audit: audit.Record{ID: "issued"}}); err == nil {
				successes.Add(1)
			}
		}()
	}
	wg.Wait()
	if successes.Load() != 1 {
		t.Fatalf("credential issued %d times", successes.Load())
	}
	current, err := store.FindPrintConnectorCredential(ctx, "hash")
	if err != nil || current.State != printing.ConnectorAwaitingActivation {
		t.Fatalf("credential not awaiting activation: %+v %v", current, err)
	}
	if _, err := store.HeartbeatPrintConnector(ctx, current, now.Add(2*time.Minute), nil, nil); err == nil {
		t.Fatal("unconfirmed credential remained valid")
	}
}
