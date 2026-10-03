package memory

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"testing"
)

func TestPrinterMutationIsScopedRevisionCheckedAndAuditAtomic(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	scope := printing.Scope{TenantID: "tenant", InventoryID: "inventory"}
	original := printing.Printer{ID: "printer", Scope: scope, Revision: 1, Name: "Garage"}
	if _, _, err := store.CreatePrinter(ctx, original, audit.Record{ID: "create-audit"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetPrinter(ctx, printing.Scope{TenantID: "other", InventoryID: scope.InventoryID}, original.ID); err == nil {
		t.Fatal("cross tenant read")
	}
	rename := func(p *printing.Printer) error { p.Name = "Changed"; p.Revision++; return nil }
	failedAudit := func(printing.Printer) (audit.Record, error) { return audit.Record{}, errors.New("audit unavailable") }
	if _, err := store.UpdatePrinter(ctx, scope, original.ID, 1, rename, failedAudit); err == nil {
		t.Fatal("ignored failed audit")
	}
	unchanged, err := store.GetPrinter(ctx, scope, original.ID)
	if err != nil || unchanged.Name != original.Name || unchanged.Revision != 1 {
		t.Fatalf("mutation leaked: %+v %v", unchanged, err)
	}
	acceptedAudit := func(printing.Printer) (audit.Record, error) { return audit.Record{ID: "update-audit"}, nil }
	if _, err := store.UpdatePrinter(ctx, scope, original.ID, 0, rename, acceptedAudit); err == nil {
		t.Fatal("accepted stale revision")
	}
	if _, err := store.UpdatePrinter(ctx, scope, original.ID, 1, rename, acceptedAudit); err != nil {
		t.Fatal(err)
	}
	result, err := store.GetPrinter(ctx, scope, original.ID)
	if err != nil || result.Name != "Changed" || len(store.auditRecords) != 2 {
		t.Fatalf("missing atomic result: %+v %v", result, err)
	}
}
