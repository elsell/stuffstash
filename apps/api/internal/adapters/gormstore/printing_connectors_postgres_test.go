package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm/clause"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPostgresPrintConnectorApprovalRollbackAndSingleCredentialExchange(t *testing.T) {
	dsn := os.Getenv("STUFF_STASH_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	db, err := OpenPostgres(dsn)
	if err != nil {
		t.Fatal(err)
	}
	connection, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	if err := runEmbeddedPostgresMigrations(db); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	store := NewStore(db)
	now := time.Now().UTC()
	const tid = "print-connector-tenant"
	const iid = "print-connector-inventory"
	saveTenant(t, ctx, store, tid, "Home")
	saveInventory(t, ctx, store, iid, tid, "Garage")
	t.Cleanup(func() {
		for _, model := range []any{&printingPairingModel{}, &printingReportModel{}, &printingBindingModel{}, &printingConnectorModel{}, &printingPrinterModel{}, &auditRecordModel{}, &inventoryModel{}} {
			if err := db.Where(clause.Eq{Column: clause.Column{Name: "tenant_id"}, Value: tid}).Delete(model).Error; err != nil {
				t.Errorf("cleanup: %v", err)
			}
		}
		if err := db.Delete(&tenantModel{ID: tid}).Error; err != nil {
			t.Errorf("cleanup tenant: %v", err)
		}
	})
	scope := printing.Scope{TenantID: tid, InventoryID: iid}
	pair := printing.Pairing{ID: "pg-print-pair", PublicKey: make([]byte, 32), PollHash: "poll", CodeHash: "pg-print-code", State: printing.PairingPending, ExpiresAt: now.Add(time.Hour), CreatedAt: now}
	if err := store.CreatePrintPairing(ctx, pair); err != nil {
		t.Fatal(err)
	}
	c := printing.Connector{ID: "pg-print-connector", Scope: scope, ServiceAccountID: "pg-print-service", Name: "Garage host", PublicKey: make([]byte, 32), State: printing.ConnectorPending, Generation: 1, CredentialVersion: 1, CreatedAt: now, UpdatedAt: now}
	record := auditRecord(t, "pg-print-approved", tenant.ID(tid), inventory.InventoryID(iid), audit.ActionPrintConnectorApproved)
	if err := store.SaveAuditRecord(ctx, record); err != nil {
		t.Fatal(err)
	}
	approval := ports.PairingApproval{PairingID: pair.ID, CodeHash: pair.CodeHash, Now: now, Registration: ports.ConnectorRegistration{Connector: c}, Audit: record}
	if _, err := store.ApprovePrintPairing(ctx, approval); err == nil {
		t.Fatal("duplicate audit accepted")
	}
	pending, err := store.GetPrintPairing(ctx, pair.ID)
	if err != nil || pending.State != printing.PairingPending {
		t.Fatalf("approval leaked after audit failure: %+v %v", pending, err)
	}
	if _, err := store.GetPrintConnector(ctx, scope, c.ID); err == nil {
		t.Fatal("connector leaked after audit failure")
	}
	approval.Audit = auditRecord(t, "pg-print-approved-2", tenant.ID(tid), inventory.InventoryID(iid), audit.ActionPrintConnectorApproved)
	if _, err := store.ApprovePrintPairing(ctx, approval); err != nil {
		t.Fatal(err)
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
			if _, err := store.ConsumePrintPairing(ctx, ports.PairingExchange{PairingID: pair.ID, Now: now, CredentialHash: "pg-print-hash", CredentialExpiresAt: now.Add(time.Hour), ActivationDeadline: now.Add(time.Minute), Audit: auditRecord(t, "pg-issued", tenant.ID(tid), inventory.InventoryID(iid), audit.ActionPrintConnectorCredentialIssued)}); err == nil {
				successes.Add(1)
			}
		}()
	}
	wg.Wait()
	if successes.Load() != 1 {
		t.Fatalf("credential issued %d times", successes.Load())
	}
	issued, err := store.FindPrintConnectorCredential(ctx, "pg-print-hash")
	if err != nil {
		t.Fatal(err)
	}
	activationAudit := func(printing.Connector, bool) (audit.Record, error) {
		return auditRecord(t, "pg-activated", tenant.ID(tid), inventory.InventoryID(iid), audit.ActionPrintConnectorActivated), nil
	}
	active, err := store.HeartbeatPrintConnector(ctx, issued, now, activationAudit)
	if err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"rotation", "replacement", "delayed"} {
		fresh := printing.Pairing{ID: printing.PairingID("pg-" + suffix), PublicKey: make([]byte, 32), PollHash: "poll-" + suffix, CodeHash: "code-" + suffix, State: printing.PairingPending, ExpiresAt: now.Add(time.Hour), CreatedAt: now}
		if err := store.CreatePrintPairing(ctx, fresh); err != nil {
			t.Fatal(err)
		}
		if _, err := store.ApprovePrintCredentialRotation(ctx, ports.RotationApproval{PairingID: fresh.ID, CodeHash: fresh.CodeHash, Scope: scope, ConnectorID: c.ID, Generation: 1, Now: now, Audit: auditRecord(t, "pg-approve-"+suffix, tenant.ID(tid), inventory.InventoryID(iid), audit.ActionPrintConnectorRotationRequested)}); err != nil {
			t.Fatal(err)
		}
	}
	next, err := store.ConsumePrintPairing(ctx, ports.PairingExchange{PairingID: "pg-rotation", Now: now, CredentialHash: "pg-next-hash", CredentialExpiresAt: now.Add(time.Hour), ActivationDeadline: now.Add(time.Minute), Audit: auditRecord(t, "pg-next-issued", tenant.ID(tid), inventory.InventoryID(iid), audit.ActionPrintConnectorCredentialIssued)})
	if err != nil {
		t.Fatal(err)
	}
	if next.ID != active.ID || next.ServiceAccountID != active.ServiceAccountID || next.CredentialVersion != active.CredentialVersion+1 {
		t.Fatalf("rotation changed identity/version: %+v", next)
	}
	expiredA := next
	now = now.Add(2 * time.Minute)
	replacement, err := store.ConsumePrintPairing(ctx, ports.PairingExchange{PairingID: "pg-replacement", Now: now, CredentialHash: "pg-replacement-hash", CredentialExpiresAt: now.Add(time.Hour), ActivationDeadline: now.Add(time.Minute), Audit: auditRecord(t, "pg-replacement-issued", tenant.ID(tid), inventory.InventoryID(iid), audit.ActionPrintConnectorCredentialIssued)})
	if err != nil {
		t.Fatal(err)
	}
	if replacement.CredentialVersion != expiredA.CredentialVersion {
		t.Fatal("fixture must exercise reused pending version")
	}
	next = replacement
	if _, err := store.HeartbeatPrintConnector(ctx, active, now, nil); err != nil {
		t.Fatalf("old credential invalid before activation: %v", err)
	}
	// A late audit failure must roll back activation and keep the old credential.
	if _, err := store.HeartbeatPrintConnector(ctx, next, now, activationAudit); err == nil {
		t.Fatal("activation committed despite duplicate audit")
	}
	stillOld, err := store.FindPrintConnectorCredential(ctx, "pg-print-hash")
	if err != nil || stillOld.CredentialVersion != active.CredentialVersion {
		t.Fatalf("old credential lost on audit rollback: %+v %v", stillOld, err)
	}
	rotated, err := store.HeartbeatPrintConnector(ctx, next, now, func(printing.Connector, bool) (audit.Record, error) {
		return auditRecord(t, "pg-rotated", tenant.ID(tid), inventory.InventoryID(iid), audit.ActionPrintConnectorCredentialRotated), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if rotated.CredentialVersion != next.CredentialVersion || rotated.PendingCredentialHash != "" {
		t.Fatal("pending credential not atomically activated")
	}
	if _, err := store.HeartbeatPrintConnector(ctx, expiredA, now, nil); err == nil {
		t.Fatal("expired pending A authenticated as replacement B")
	}
	if _, err := store.FindPrintConnectorCredential(ctx, "pg-print-hash"); err == nil {
		t.Fatal("old credential hash still authenticates")
	}
	if _, err := store.HeartbeatPrintConnector(ctx, active, now, nil); err == nil {
		t.Fatal("old authenticated request bypassed version fence")
	}
	if _, err := store.ConsumePrintPairing(ctx, ports.PairingExchange{PairingID: "pg-delayed", Now: now, CredentialHash: "pg-stale-hash", CredentialExpiresAt: now.Add(time.Hour), ActivationDeadline: now.Add(time.Minute), Audit: auditRecord(t, "pg-stale-issued", tenant.ID(tid), inventory.InventoryID(iid), audit.ActionPrintConnectorCredentialIssued)}); err == nil {
		t.Fatal("delayed old approval replaced new credential")
	}

}
