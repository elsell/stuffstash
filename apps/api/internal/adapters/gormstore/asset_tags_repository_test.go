package gormstore

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestStoreAssetTagsScopeAndAssignments(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("tenant-one")
	otherTenantID := tenant.ID("tenant-two")
	inventoryID := inventory.InventoryID("inventory-one")
	otherInventoryID := inventory.InventoryID("inventory-two")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveTenant(t, ctx, store, otherTenantID, "Cabin")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")
	saveInventory(t, ctx, store, otherInventoryID.String(), tenantID, "Supplies")

	item := assetItem("asset-one", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	if err := createAsset(t, ctx, store, item); err != nil {
		t.Fatalf("create asset: %v", err)
	}
	workshop := assetTag(t, "tag-workshop", tenantID, inventoryID, "workshop", "Workshop")
	if err := store.CreateAssetTag(ctx, workshop, auditRecord(t, "audit-tag-workshop", tenantID, inventoryID, audit.ActionAssetTagCreated)); err != nil {
		t.Fatalf("create workshop tag: %v", err)
	}
	if err := store.CreateAssetTag(ctx, assetTag(t, "tag-duplicate", tenantID, inventoryID, "workshop", "Workshop Duplicate"), auditRecord(t, "audit-tag-duplicate", tenantID, inventoryID, audit.ActionAssetTagCreated)); err == nil {
		t.Fatalf("expected duplicate tag key rejection")
	}
	otherInventoryTag := assetTag(t, "tag-other-inventory", tenantID, otherInventoryID, "other", "Other")
	if err := store.CreateAssetTag(ctx, otherInventoryTag, auditRecord(t, "audit-tag-other", tenantID, otherInventoryID, audit.ActionAssetTagCreated)); err != nil {
		t.Fatalf("create other inventory tag: %v", err)
	}
	otherTenantTag := assetTag(t, "tag-other-tenant", otherTenantID, inventoryID, "cabin", "Cabin")
	if err := store.CreateAssetTag(ctx, otherTenantTag, auditRecord(t, "audit-tag-cabin", otherTenantID, inventoryID, audit.ActionAssetTagCreated)); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected tenant/inventory scope rejection, got %v", err)
	}
	archived := assetTag(t, "tag-archived", tenantID, inventoryID, "archived", "Archived")
	if err := store.CreateAssetTag(ctx, archived, auditRecord(t, "audit-tag-archived", tenantID, inventoryID, audit.ActionAssetTagCreated)); err != nil {
		t.Fatalf("create archived tag: %v", err)
	}
	archived.LifecycleState = assettag.LifecycleStateArchived
	archived.UpdatedAt = time.Now().UTC()
	if err := store.UpdateAssetTagLifecycle(ctx, archived, auditRecord(t, "audit-tag-archived-update", tenantID, inventoryID, audit.ActionAssetTagArchived)); err != nil {
		t.Fatalf("archive tag: %v", err)
	}

	if err := store.SetAssetTags(ctx, tenantID, inventoryID, item.ID, []assettag.ID{workshop.ID}, auditRecord(t, "audit-set-tags", tenantID, inventoryID, audit.ActionAssetUpdated)); err != nil {
		t.Fatalf("assign workshop tag: %v", err)
	}
	assigned, err := store.AssetTagsByAsset(ctx, tenantID, inventoryID, item.ID)
	if err != nil {
		t.Fatalf("read assigned tags: %v", err)
	}
	if len(assigned) != 1 || assigned[0].ID != workshop.ID {
		t.Fatalf("expected workshop assignment, got %+v", assigned)
	}
	assignedByAsset, err := store.AssetTagsByAssets(ctx, tenantID, inventoryID, []asset.ID{item.ID})
	if err != nil {
		t.Fatalf("read assigned tags by assets: %v", err)
	}
	if len(assignedByAsset[item.ID]) != 1 || assignedByAsset[item.ID][0].ID != workshop.ID {
		t.Fatalf("expected workshop assignment by asset, got %+v", assignedByAsset)
	}
	listed, err := store.ListAssetTags(ctx, tenantID, inventoryID, ports.AssetTagPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != workshop.ID {
		t.Fatalf("expected active inventory tag only, got %+v", listed)
	}

	if err := store.SetAssetTags(ctx, tenantID, inventoryID, item.ID, []assettag.ID{otherInventoryTag.ID}, auditRecord(t, "audit-cross-inventory-tags", tenantID, inventoryID, audit.ActionAssetUpdated)); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected cross-inventory tag rejection, got %v", err)
	}
	if err := store.SetAssetTags(ctx, tenantID, inventoryID, item.ID, []assettag.ID{archived.ID}, auditRecord(t, "audit-archived-tags", tenantID, inventoryID, audit.ActionAssetUpdated)); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected archived tag rejection, got %v", err)
	}
	assigned, err = store.AssetTagsByAsset(ctx, tenantID, inventoryID, item.ID)
	if err != nil {
		t.Fatalf("read assigned tags after rejected updates: %v", err)
	}
	if len(assigned) != 1 || assigned[0].ID != workshop.ID {
		t.Fatalf("rejected assignments must not replace existing assignment, got %+v", assigned)
	}
}

func TestStoreRollsBackDirectAssetEditWhenTagReplacementFails(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")
	item := assetItem("asset-one", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	if err := createAsset(t, ctx, store, item); err != nil {
		t.Fatalf("create asset: %v", err)
	}
	updated := item
	updated.Title, _ = asset.NewTitle("Changed title")
	operation := undoableAssetOperation("operation-edit", tenantID, inventoryID, audit.ActionAssetUpdated, &item, updated)
	err := store.UpdateAssetAndTags(ctx, updated, []assettag.ID{"missing-tag"}, []audit.Record{
		auditRecord(t, "audit-edit", tenantID, inventoryID, audit.ActionAssetUpdated),
	}, &operation)
	if !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected invalid tag replacement rejection, got %v", err)
	}
	persisted, found, err := store.AssetByID(ctx, tenantID, inventoryID, item.ID)
	if err != nil || !found || persisted.Title != item.Title {
		t.Fatalf("expected asset rollback, found=%t asset=%+v err=%v", found, persisted, err)
	}
	if _, found, err := store.UndoableOperationByID(ctx, tenantID, inventoryID, operation.ID); err != nil || found {
		t.Fatalf("expected undo rollback, found=%t err=%v", found, err)
	}
}

func assetTag(t *testing.T, id string, tenantID tenant.ID, inventoryID inventory.InventoryID, keyValue string, displayNameValue string) assettag.Tag {
	t.Helper()
	tagID, ok := assettag.NewID(id)
	if !ok {
		t.Fatalf("expected valid tag id")
	}
	key, ok := assettag.NewKey(keyValue)
	if !ok {
		t.Fatalf("expected valid tag key")
	}
	displayName, ok := assettag.NewDisplayName(displayNameValue)
	if !ok {
		t.Fatalf("expected valid tag display name")
	}
	color, ok := assettag.NewColor("")
	if !ok {
		t.Fatalf("expected valid empty tag color")
	}
	tag, ok := assettag.NewTag(tagID, assettag.TenantID(tenantID.String()), assettag.InventoryID(inventoryID.String()), key, displayName, color, time.Now().UTC())
	if !ok {
		t.Fatalf("expected valid tag")
	}
	return tag
}
