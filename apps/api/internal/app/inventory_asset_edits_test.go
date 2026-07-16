package app

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func TestUpdateAssetPersistsFieldsAndTagsAsOneUndoableChangeAndSkipsNoOp(t *testing.T) {
	item := assetItem("drill", "tenant-one", "inventory-one", asset.KindItem, "")
	workshop := assetTag("tag-workshop", "tenant-one", "inventory-one", "workshop", "Workshop", "#2f80ed")
	camping := assetTag("tag-camping", "tenant-one", "inventory-one", "camping", "Camping", "#27ae60")
	assets := &fakeAssetRepository{
		items: map[asset.ID]asset.Asset{item.ID: item}, assetTags: map[assettag.ID]assettag.Tag{workshop.ID: workshop, camping.ID: camping},
		assetTagLinks: map[asset.ID]map[assettag.ID]struct{}{item.ID: {workshop.ID: {}}},
	}
	application := New(Dependencies{
		Observer: &fakeObserver{}, Authorizer: &fakeAuthorizer{}, Tenants: &fakeTenantRepository{exists: true},
		Inventories: &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		Assets:      assets, AssetTags: assets, AssetUnitOfWork: assets, AssetTagUnitOfWork: assets, AssetEditUnitOfWork: assets, Undoables: assets,
		Audit: &fakeAuditRepository{}, Outbox: &fakeOutbox{}, IDs: &fakeIDGenerator{ids: []string{"operation-update", "audit-update", "audit-undo"}},
	})
	title := "Cordless drill"
	tagIDs := []string{camping.ID.String()}
	if _, err := application.UpdateAsset(context.Background(), UpdateAssetInput{Principal: identity.Principal{ID: "editor"}, TenantID: "tenant-one", InventoryID: "inventory-one", AssetID: item.ID, Title: &title, TagIDs: &tagIDs}); err != nil {
		t.Fatalf("update fields and tags: %v", err)
	}
	if len(assets.auditRecords) != 1 || assets.auditRecords[0].Action != audit.ActionAssetUpdated {
		t.Fatalf("expected one coherent update audit, got %+v", assets.auditRecords)
	}
	record := assets.auditRecords[0]
	if record.Metadata["previous_title"] != item.Title.String() || record.Metadata["updated_title"] != title || record.Metadata["previous_tag_count"] != "1" || record.Metadata["updated_tag_count"] != "1" {
		t.Fatalf("expected safe field and tag metadata, got %+v", record.Metadata)
	}
	operation := assets.undoables[record.Metadata["operation_id"]]
	if operation.ID == "" || len(operation.BeforeTagIDs) != 1 || operation.BeforeTagIDs[0] != workshop.ID || len(operation.AfterTagIDs) != 1 || operation.AfterTagIDs[0] != camping.ID {
		t.Fatalf("expected coherent tag snapshots on undo, got %+v", operation)
	}
	auditCount, operationCount := len(assets.auditRecords), len(assets.undoables)
	if _, err := application.UpdateAsset(context.Background(), UpdateAssetInput{Principal: identity.Principal{ID: "editor"}, TenantID: "tenant-one", InventoryID: "inventory-one", AssetID: item.ID, Title: &title, TagIDs: &tagIDs}); err != nil {
		t.Fatalf("repeat no-op edit: %v", err)
	}
	if len(assets.auditRecords) != auditCount || len(assets.undoables) != operationCount {
		t.Fatalf("no-op edit created history: audits=%d operations=%d", len(assets.auditRecords), len(assets.undoables))
	}
	if _, err := application.UndoOperation(context.Background(), ApplyUndoableOperationInput{Principal: identity.Principal{ID: "editor"}, TenantID: "tenant-one", InventoryID: "inventory-one", OperationID: operation.ID}); err != nil {
		t.Fatalf("undo coherent edit: %v", err)
	}
	restoredTags, err := assets.AssetTagsByAsset(context.Background(), tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), item.ID)
	if err != nil {
		t.Fatalf("read restored tags: %v", err)
	}
	if assets.items[item.ID].Title != item.Title || len(restoredTags) != 1 || restoredTags[0].ID != workshop.ID {
		t.Fatalf("expected undo to restore fields and tags, item=%+v tags=%+v", assets.items[item.ID], restoredTags)
	}
}

func assetTag(id, tenantID, inventoryID, keyValue, displayNameValue, color string) assettag.Tag {
	key, ok := assettag.NewKey(keyValue)
	if !ok {
		panic("invalid test tag key")
	}
	displayName, ok := assettag.NewDisplayName(displayNameValue)
	if !ok {
		panic("invalid test tag display name")
	}
	tagColor, ok := assettag.NewColor(color)
	if !ok {
		panic("invalid test tag color")
	}
	tag, ok := assettag.NewTag(assettag.ID(id), assettag.TenantID(tenantID), assettag.InventoryID(inventoryID), key, displayName, tagColor, time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC))
	if !ok {
		panic("invalid test tag")
	}
	return tag
}
