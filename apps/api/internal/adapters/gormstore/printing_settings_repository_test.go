package gormstore

import (
	"context"
	"errors"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestPrintSettingsRevisionDestinationFenceAndAuditRollback(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, "tenant", "Home")
	saveInventory(t, ctx, store, "inventory", "tenant", "Garage")
	scope := printing.Scope{TenantID: "tenant", InventoryID: "inventory"}
	printer := printing.Printer{ID: "printer", Scope: scope, Name: "Garage", Revision: 1, RequestKey: "one", RequestFingerprint: "one", MediaFingerprint: "media-one"}
	registered := auditRecord(t, "registered", tenant.ID("tenant"), inventory.InventoryID("inventory"), audit.ActionPrinterRegistered)
	if _, _, err := store.CreatePrinter(ctx, printer, registered); err != nil {
		t.Fatal(err)
	}
	settings := printing.DefaultInventoryPrintSettings(scope)
	settings.Revision = 1
	settings.DefaultPrinterID = printer.ID
	settings.PrintOnCreateDefault = true
	fence := &printing.SettingsDestination{ID: printer.ID, Revision: 1, MediaFingerprint: "media-one"}
	record := auditRecord(t, "settings", tenant.ID("tenant"), inventory.InventoryID("inventory"), audit.ActionPrintSettingsUpdated)
	if _, err := store.SavePrintSettings(ctx, settings, 0, fence, record); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SavePrintSettings(ctx, settings, 0, fence, record); !errors.Is(err, ports.ErrPrintConflict) {
		t.Fatalf("stale write: %v", err)
	}
	settings.Revision = 2
	settings.PrintOnCreateDefault = false
	if _, err := store.SavePrintSettings(ctx, settings, 1, fence, record); err == nil {
		t.Fatal("duplicate audit should roll back settings")
	}
	current, err := store.GetPrintSettings(ctx, scope)
	if err != nil || current.Revision != 1 || !current.PrintOnCreateDefault {
		t.Fatalf("audit rollback lost settings: %+v %v", current, err)
	}
	fence.MediaFingerprint = "changed-after-render"
	replacement := auditRecord(t, "replacement", tenant.ID("tenant"), inventory.InventoryID("inventory"), audit.ActionPrintSettingsUpdated)
	if _, err := store.SavePrintSettings(ctx, settings, 1, fence, replacement); !errors.Is(err, ports.ErrPrintConflict) {
		t.Fatalf("changed media accepted: %v", err)
	}
	settings.DefaultPrinterID = ""
	settings.Template.Options.ShowReference = false
	if _, err = store.SavePrintSettings(ctx, settings, 1, nil, replacement); err != nil {
		t.Fatal(err)
	}
	current, err = store.GetPrintSettings(ctx, scope)
	if err != nil || current.Revision != 2 || current.PrintOnCreateDefault || current.DefaultPrinterID != "" || current.Template.Options.ShowReference {
		t.Fatal("clearing defaults did not persist zero values")
	}
	other, err := store.GetPrintSettings(ctx, printing.Scope{TenantID: "other", InventoryID: "inventory"})
	if err != nil || other.DefaultPrinterID != "" || other.Revision != 0 {
		t.Fatal("cross tenant settings exposed")
	}
}
