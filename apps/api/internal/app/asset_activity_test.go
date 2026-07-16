package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestListAssetActivityDefaultsToChangesProjectsSafeFieldsAndPaginates(t *testing.T) {
	base := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	created := assetActivityRecord(t, "audit-created", audit.ActionAssetCreated, base)
	updated := assetActivityRecord(t, "audit-updated", audit.ActionAssetUpdated, base.Add(time.Minute))
	updated.Metadata = map[string]string{
		"previous_title":           "Drill",
		"updated_title":            "Cordless drill",
		"description_changed":      "true",
		"previous_tag_count":       "1",
		"updated_tag_count":        "2",
		"previous_parent":          "Garage",
		"new_parent":               "Workshop",
		"previous_lifecycle_state": "archived",
		"new_lifecycle_state":      "active",
		"previous_checkout_state":  "available",
		"new_checkout_state":       "checked_out",
		"operation_id":             "operation-one",
		"original_action":          "asset.updated",
		"target_type":              "asset",
		"credential":               "must-not-leak",
	}
	viewed := assetActivityRecord(t, "audit-viewed", audit.ActionAssetViewed, base.Add(2*time.Minute))
	assets := &fakeAssetRepository{
		items: map[asset.ID]asset.Asset{
			asset.ID("asset-one"): assetItem("asset-one", "tenant-one", "inventory-one", asset.KindItem, ""),
		},
		undoables: map[string]ports.UndoableOperation{
			"operation-one": {
				ID: "operation-one", TenantID: tenant.ID("tenant-one"), InventoryID: inventory.InventoryID("inventory-one"),
				TargetType: audit.TargetAsset, TargetID: "asset-one", OriginalAction: audit.ActionAssetUpdated, Status: ports.UndoableOperationAvailable,
			},
		},
	}
	application := assetActivityApplication(assets, []audit.Record{created, updated, viewed})

	first, err := application.ListAssetActivity(context.Background(), ListAssetActivityInput{
		Principal: identity.Principal{ID: identity.PrincipalID("viewer")}, TenantID: tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"), AssetID: asset.ID("asset-one"), Limit: 1,
	})
	if err != nil {
		t.Fatalf("list first activity page: %v", err)
	}
	if len(first.Items) != 1 || first.Items[0].Action != audit.ActionAssetUpdated || !first.HasMore || first.NextCursor == nil {
		t.Fatalf("unexpected first page: %+v", first)
	}
	entry := first.Items[0]
	if entry.Category != audit.AssetActivityCategoryChange || entry.Undo == nil || entry.Undo.OperationID != "operation-one" || entry.Undo.Status != string(ports.UndoableOperationAvailable) {
		t.Fatalf("expected typed change with undo, got %+v", entry)
	}
	if len(entry.Changes) != 6 || entry.Changes[0].Field != audit.AssetActivityFieldTitle || entry.Changes[0].PreviousValue != "Drill" || entry.Changes[0].CurrentValue != "Cordless drill" || entry.Changes[1].Field != audit.AssetActivityFieldDescription || entry.Changes[2].Field != audit.AssetActivityFieldTags || entry.Changes[3].Field != audit.AssetActivityFieldParent || entry.Changes[4].Field != audit.AssetActivityFieldLifecycleState || entry.Changes[5].Field != audit.AssetActivityFieldCheckoutState {
		t.Fatalf("unexpected safe changes: %+v", entry.Changes)
	}
	if entry.TechnicalMetadata["credential"] != "" || entry.TechnicalMetadata["operation_id"] != "" {
		t.Fatalf("unsafe metadata leaked: %+v", entry.TechnicalMetadata)
	}
	if entry.TechnicalMetadata["original_action"] != "asset.updated" || entry.TechnicalMetadata["target_type"] != "asset" {
		t.Fatalf("expected validated safe technical metadata, got %+v", entry.TechnicalMetadata)
	}

	second, err := application.ListAssetActivity(context.Background(), ListAssetActivityInput{
		Principal: identity.Principal{ID: identity.PrincipalID("viewer")}, TenantID: tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"), AssetID: asset.ID("asset-one"), Limit: 1, Cursor: *first.NextCursor,
	})
	if err != nil {
		t.Fatalf("list second activity page: %v", err)
	}
	if len(second.Items) != 1 || second.Items[0].Action != audit.ActionAssetCreated {
		t.Fatalf("unexpected second page: %+v", second)
	}
}

func TestAssetActivityCursorBindsTenantInventoryAssetAndView(t *testing.T) {
	record := assetActivityRecord(t, "audit-one", audit.ActionAssetUpdated, time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC))
	cursor := encodeAssetActivityCursor("tenant-one", "inventory-one", "asset-one", AssetActivityViewChanges, record)
	if cursor == nil {
		t.Fatal("expected cursor")
	}
	tests := []struct {
		name      string
		tenant    tenant.ID
		inventory inventory.InventoryID
		asset     asset.ID
		view      AssetActivityView
	}{
		{name: "tenant", tenant: "tenant-two", inventory: "inventory-one", asset: "asset-one", view: AssetActivityViewChanges},
		{name: "inventory", tenant: "tenant-one", inventory: "inventory-two", asset: "asset-one", view: AssetActivityViewChanges},
		{name: "asset", tenant: "tenant-one", inventory: "inventory-one", asset: "asset-two", view: AssetActivityViewChanges},
		{name: "view", tenant: "tenant-one", inventory: "inventory-one", asset: "asset-one", view: AssetActivityViewAll},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, _, err := decodeAssetActivityCursor(test.tenant, test.inventory, test.asset, test.view, *cursor); err == nil {
				t.Fatal("expected scoped cursor rejection")
			}
		})
	}
}

func TestAssetActivityOmitsUndoOperationFromAnotherScope(t *testing.T) {
	record := assetActivityRecord(t, "audit-one", audit.ActionAssetUpdated, time.Now())
	record.Metadata = map[string]string{"operation_id": "operation-other"}
	repository := &fakeAssetRepository{undoables: map[string]ports.UndoableOperation{
		"operation-other": {ID: "operation-other", TenantID: "tenant-one", InventoryID: "inventory-two", TargetType: audit.TargetAsset, TargetID: "asset-one", OriginalAction: audit.ActionAssetUpdated, Status: ports.UndoableOperationAvailable},
	}}
	application := App{undoables: repository}
	entry := application.projectAssetActivityEntry(context.Background(), ListAssetActivityInput{TenantID: "tenant-one", InventoryID: "inventory-one", AssetID: "asset-one"}, record, true)
	if entry.Undo != nil {
		t.Fatalf("expected wrong-scope operation to be omitted, got %+v", entry.Undo)
	}
}

func TestListAssetActivityAllIncludesReadsAndRejectsWrongScopeCursor(t *testing.T) {
	base := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		asset.ID("asset-one"): assetItem("asset-one", "tenant-one", "inventory-one", asset.KindItem, ""),
		asset.ID("asset-two"): assetItem("asset-two", "tenant-one", "inventory-one", asset.KindItem, ""),
	}}
	application := assetActivityApplication(assets, []audit.Record{
		assetActivityRecord(t, "audit-created", audit.ActionAssetCreated, base),
		assetActivityRecord(t, "audit-viewed", audit.ActionAssetViewed, base.Add(time.Minute)),
	})
	page, err := application.ListAssetActivity(context.Background(), ListAssetActivityInput{
		Principal: identity.Principal{ID: identity.PrincipalID("viewer")}, TenantID: tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"), AssetID: asset.ID("asset-one"), Limit: 1, View: AssetActivityViewAll,
	})
	if err != nil || len(page.Items) != 1 || page.Items[0].Category != audit.AssetActivityCategoryRead || page.NextCursor == nil {
		t.Fatalf("unexpected all-events page: %+v, %v", page, err)
	}
	_, err = application.ListAssetActivity(context.Background(), ListAssetActivityInput{
		Principal: identity.Principal{ID: identity.PrincipalID("viewer")}, TenantID: tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"), AssetID: asset.ID("asset-two"), Cursor: *page.NextCursor,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected wrong-scope cursor to fail, got %v", err)
	}
}

func TestListAssetActivityRequiresConcreteAsset(t *testing.T) {
	application := assetActivityApplication(&fakeAssetRepository{items: map[asset.ID]asset.Asset{}}, nil)
	_, err := application.ListAssetActivity(context.Background(), ListAssetActivityInput{
		Principal: identity.Principal{ID: identity.PrincipalID("viewer")}, TenantID: tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"), AssetID: asset.ID("missing"),
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing asset to fail, got %v", err)
	}
}

func TestListAssetActivityIncludesOnlySameScopeUndoAndRedoRecords(t *testing.T) {
	base := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	undo := assetActivityRecord(t, "audit-undo", audit.ActionUndoableOperationUndone, base)
	undo.Metadata = map[string]string{"operation_id": "operation-one"}
	wrongInventory := undo
	wrongInventory.ID = "audit-wrong-inventory"
	wrongInventory.InventoryID = "inventory-two"
	wrongAsset := undo
	wrongAsset.ID = "audit-wrong-asset"
	wrongAsset.TargetID = "asset-two"
	application := assetActivityApplication(&fakeAssetRepository{items: map[asset.ID]asset.Asset{
		"asset-one": assetItem("asset-one", "tenant-one", "inventory-one", asset.KindItem, ""),
	}}, []audit.Record{undo, wrongInventory, wrongAsset})
	page, err := application.ListAssetActivity(context.Background(), ListAssetActivityInput{
		Principal: identity.Principal{ID: "viewer"}, TenantID: "tenant-one", InventoryID: "inventory-one", AssetID: "asset-one", Limit: 10,
	})
	if err != nil || len(page.Items) != 1 || page.Items[0].Action != audit.ActionUndoableOperationUndone {
		t.Fatalf("expected only same-scope undo record, page=%+v err=%v", page, err)
	}
}

func assetActivityApplication(assets *fakeAssetRepository, records []audit.Record) App {
	return New(Dependencies{
		Observer: &fakeObserver{}, Authorizer: &fakeAuthorizer{}, Assets: assets, Undoables: assets,
		Tenants: &fakeTenantRepository{exists: true}, TenantUnitOfWork: &fakeTenantRepository{exists: true},
		Inventories: &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		Audit:       &fakeAuditRepository{items: records}, Outbox: &fakeOutbox{},
	})
}

func assetActivityRecord(t *testing.T, id string, action audit.Action, occurredAt time.Time) audit.Record {
	t.Helper()
	record, ok := audit.NewRecord(audit.ID(id), audit.TenantID("tenant-one"), audit.InventoryID("inventory-one"), audit.PrincipalID("owner"), action, audit.SourceAPI, audit.TargetAsset, "asset-one", occurredAt, "request-one", map[string]string{})
	if !ok {
		t.Fatal("invalid audit record fixture")
	}
	return record
}
