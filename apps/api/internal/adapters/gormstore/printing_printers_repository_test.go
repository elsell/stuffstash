package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"testing"
)

func TestPrinterRepositoryReplaysAndRollsBackWithAudit(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, "tenant", "Home")
	saveInventory(t, ctx, store, "inventory", "tenant", "Garage")
	scope := printing.Scope{TenantID: "tenant", InventoryID: "inventory"}
	p := printing.Printer{ID: "printer", Scope: scope, Name: "Garage", Revision: 1, RequestKey: "key", RequestFingerprint: "fingerprint"}
	record := auditRecord(t, "printer-create", tenant.ID("tenant"), inventory.InventoryID("inventory"), audit.ActionPrinterRegistered)
	if _, created, err := store.CreatePrinter(ctx, p, record); err != nil || !created {
		t.Fatalf("create %v %v", created, err)
	}
	retry := p
	retry.ID = "retry-id"
	if result, created, err := store.CreatePrinter(ctx, retry, record); err != nil || created || result.ID != p.ID {
		t.Fatalf("replay %+v %v %v", result, created, err)
	}
	retry.RequestFingerprint = "different"
	if _, _, err := store.CreatePrinter(ctx, retry, record); err == nil {
		t.Fatal("changed replay accepted")
	}
	change := func(p *printing.Printer) error { p.Name = "Mutated"; p.Revision++; return nil }
	// A duplicate history primary key fails after the row update; the complete
	// transaction must roll back rather than leave a mutation without history.
	if _, err := store.UpdatePrinter(ctx, scope, p.ID, 1, change, func(printing.Printer) (audit.Record, error) { return record, nil }); err == nil {
		t.Fatal("duplicate audit accepted")
	}
	unchanged, err := store.GetPrinter(ctx, scope, p.ID)
	if err != nil || unchanged.Name != "Garage" || unchanged.Revision != 1 {
		t.Fatalf("rollback %+v %v", unchanged, err)
	}
	if _, err := store.GetPrinter(ctx, printing.Scope{TenantID: "other", InventoryID: "inventory"}, p.ID); err == nil {
		t.Fatal("cross tenant read")
	}
}
