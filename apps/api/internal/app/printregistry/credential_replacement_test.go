package printregistry_test

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/adapters/pairingcrypto"
	"github.com/stuffstash/stuff-stash/internal/app/printregistry"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

type credentialClock struct{ now time.Time }

func (c *credentialClock) Now() time.Time { return c.now }
func TestExpiredPendingCredentialCannotReuseReplacementAuthority(t *testing.T) {
	ctx := context.Background()
	clock := &credentialClock{now: time.Now().UTC()}
	store := memory.NewStore()
	authorization := memory.NewAuthorizer()
	secrets := pairingcrypto.Secrets{}
	scope := printing.Scope{TenantID: "tenant", InventoryID: "inventory"}
	printer := printing.Printer{ID: "printer", Scope: scope, Revision: 1}
	if _, _, err := store.CreatePrinter(ctx, printer, audit.Record{ID: "printer-created"}); err != nil {
		t.Fatal(err)
	}
	pair := printing.Pairing{ID: "initial", State: printing.PairingPending, CodeHash: "initial-code", ExpiresAt: clock.now.Add(time.Hour)}
	if err := store.CreatePrintPairing(ctx, pair); err != nil {
		t.Fatal(err)
	}
	c := printing.Connector{ID: "connector", Scope: scope, ServiceAccountID: "service", State: printing.ConnectorPending, Generation: 1, CredentialVersion: 1}
	bindings := []printing.PrinterBinding{{Scope: scope, ConnectorID: c.ID, PrinterID: printer.ID, DeviceID: "usb-device", Generation: 1}}
	if _, err := store.ApprovePrintPairing(ctx, ports.PairingApproval{PairingID: pair.ID, CodeHash: pair.CodeHash, Now: clock.now, Registration: ports.ConnectorRegistration{Connector: c, Bindings: bindings}, Audit: audit.Record{ID: "approved"}}); err != nil {
		t.Fatal(err)
	}
	if err := store.SynchronizePrintConnector(ctx, scope, c.ID, func(ctx context.Context, r ports.ConnectorRegistration, printers []printing.Printer) error {
		return authorization.SyncPrintConnector(ctx, r.Connector, r.Bindings)
	}); err != nil {
		t.Fatal(err)
	}
	issued, err := store.ConsumePrintPairing(ctx, ports.PairingExchange{PairingID: pair.ID, Now: clock.now, CredentialHash: secrets.Digest("old"), CredentialExpiresAt: clock.now.Add(time.Hour), ActivationDeadline: clock.now.Add(time.Minute), Audit: audit.Record{ID: "issued"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.HeartbeatPrintConnector(ctx, issued, clock.now, func(printing.Connector, bool) (audit.Record, error) { return audit.Record{ID: "activated"}, nil }, nil); err != nil {
		t.Fatal(err)
	}
	service := printregistry.ConnectorService{Registry: printregistry.Service{Clock: clock}, Repository: store, Authorization: authorization, Secrets: secrets}
	rotate := func(id string) printing.Connector {
		t.Helper()
		p := printing.Pairing{ID: printing.PairingID(id), State: printing.PairingPending, CodeHash: id, ExpiresAt: clock.now.Add(time.Hour)}
		if err := store.CreatePrintPairing(ctx, p); err != nil {
			t.Fatal(err)
		}
		if _, err := store.ApprovePrintCredentialRotation(ctx, ports.RotationApproval{PairingID: p.ID, CodeHash: p.CodeHash, Scope: scope, ConnectorID: c.ID, Generation: 1, Now: clock.now, Audit: audit.Record{ID: audit.ID(id + "-approved")}}); err != nil {
			t.Fatal(err)
		}
		result, err := store.ConsumePrintPairing(ctx, ports.PairingExchange{PairingID: p.ID, Now: clock.now, CredentialHash: secrets.Digest(id), CredentialExpiresAt: clock.now.Add(time.Hour), ActivationDeadline: clock.now.Add(time.Minute), Audit: audit.Record{ID: audit.ID(id + "-issued")}})
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	rotate("A")
	delayedA, err := service.AuthenticateConsumer(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	clock.now = clock.now.Add(2 * time.Minute)
	next := rotate("B")
	if next.CredentialVersion != delayedA.CredentialVersion {
		t.Fatal("fixture must exercise the same pending version")
	}
	if _, err := store.HeartbeatPrintConnector(ctx, next, clock.now, func(printing.Connector, bool) (audit.Record, error) { return audit.Record{ID: "B-activated"}, nil }, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := store.HeartbeatPrintConnector(ctx, delayedA, clock.now, nil, nil); err == nil {
		t.Error("delayed A heartbeat authenticated as replacement B")
	}
	if _, err := service.AuthorizePrinter(ctx, delayedA, printer.ID, ports.PrinterPermissionConsume); err == nil {
		t.Error("delayed pending A authorized work after B activated")
	}
}
