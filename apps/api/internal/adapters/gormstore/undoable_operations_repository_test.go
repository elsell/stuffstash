package gormstore

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestStoreCreatesAndAppliesUndoableAssetOperation(t *testing.T) {
	ctx := context.Background()
	store := newUndoableOperationTestStore(t, ctx)
	item := assetItem("asset-one", "tenant-one", "inventory-one", asset.KindItem, "")
	operation := undoableAssetOperation("operation-one", tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), audit.ActionAssetCreated, nil, item)
	if err := store.CreateAsset(ctx, item, auditRecord(t, "audit-create", tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), audit.ActionAssetCreated), &operation); err != nil {
		t.Fatalf("create asset with undoable operation: %v", err)
	}

	foundOperation, found, err := store.UndoableOperationByID(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), operation.ID)
	if err != nil {
		t.Fatalf("find undoable operation: %v", err)
	}
	if !found || foundOperation.AfterAsset.ID != item.ID || foundOperation.Status != ports.UndoableOperationAvailable {
		t.Fatalf("expected stored undoable operation, found=%t operation=%+v", found, foundOperation)
	}

	resulting := item
	resulting.LifecycleState = asset.LifecycleStateArchived
	applied, appliedAsset, err := store.ApplyAssetUndoableOperation(ctx, operation.ID, ports.UndoableOperationDirectionUndo, item, resulting, auditRecord(t, "audit-undo", tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), audit.ActionUndoableOperationUndone))
	if err != nil {
		t.Fatalf("apply undo operation: %v", err)
	}
	if applied.Status != ports.UndoableOperationUndone || applied.UndoAuditRecordID.String() == "" || appliedAsset.LifecycleState != asset.LifecycleStateArchived {
		t.Fatalf("expected applied undo state, operation=%+v asset=%+v", applied, appliedAsset)
	}
	persisted, ok, err := store.AssetByID(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), item.ID)
	if err != nil {
		t.Fatalf("find asset after undo: %v", err)
	}
	if !ok || persisted.LifecycleState != asset.LifecycleStateArchived {
		t.Fatalf("expected archived asset after undo, found=%t asset=%+v", ok, persisted)
	}
}

func TestStoreUndoRestoresAssetAndTagAssignmentsTogether(t *testing.T) {
	ctx := context.Background()
	store := newUndoableOperationTestStore(t, ctx)
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	original := assetItem("asset-one", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	if err := createAsset(t, ctx, store, original); err != nil {
		t.Fatalf("create asset: %v", err)
	}
	workshop := assetTag(t, "tag-workshop", tenantID, inventoryID, "workshop", "Workshop")
	camping := assetTag(t, "tag-camping", tenantID, inventoryID, "camping", "Camping")
	for index, tag := range []assettag.Tag{workshop, camping} {
		if err := store.CreateAssetTag(ctx, tag, auditRecord(t, "audit-tag-"+string(rune('a'+index)), tenantID, inventoryID, audit.ActionAssetTagCreated)); err != nil {
			t.Fatalf("create tag: %v", err)
		}
	}
	if err := store.SetAssetTags(ctx, tenantID, inventoryID, original.ID, []assettag.ID{workshop.ID}, auditRecord(t, "audit-initial-tags", tenantID, inventoryID, audit.ActionAssetUpdated)); err != nil {
		t.Fatalf("set initial tags: %v", err)
	}
	updated := original
	updated.Title, _ = asset.NewTitle("Cordless drill")
	operation := undoableAssetOperation("operation-edit", tenantID, inventoryID, audit.ActionAssetUpdated, &original, updated)
	operation.ReplacesTags = true
	operation.BeforeTagIDs = []assettag.ID{workshop.ID}
	operation.AfterTagIDs = []assettag.ID{camping.ID}
	if err := store.UpdateAssetAndTags(ctx, updated, operation.AfterTagIDs, []audit.Record{auditRecord(t, "audit-edit", tenantID, inventoryID, audit.ActionAssetUpdated)}, &operation); err != nil {
		t.Fatalf("update fields and tags: %v", err)
	}
	loaded, found, err := store.UndoableOperationByID(ctx, tenantID, inventoryID, operation.ID)
	if err != nil || !found || !loaded.ReplacesTags || len(loaded.BeforeTagIDs) != 1 || loaded.BeforeTagIDs[0] != workshop.ID {
		t.Fatalf("expected persisted tag snapshots, found=%t operation=%+v err=%v", found, loaded, err)
	}
	if err := store.SetAssetTags(ctx, tenantID, inventoryID, original.ID, []assettag.ID{workshop.ID}, auditRecord(t, "audit-newer-tags", tenantID, inventoryID, audit.ActionAssetUpdated)); err != nil {
		t.Fatalf("set newer tags: %v", err)
	}
	if _, _, err := store.ApplyAssetUndoableOperation(ctx, operation.ID, ports.UndoableOperationDirectionUndo, updated, original, auditRecord(t, "audit-stale-tag-undo", tenantID, inventoryID, audit.ActionUndoableOperationUndone)); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("expected newer tags to make undo stale, got %v", err)
	}
	if err := store.SetAssetTags(ctx, tenantID, inventoryID, original.ID, []assettag.ID{camping.ID}, auditRecord(t, "audit-reset-tags", tenantID, inventoryID, audit.ActionAssetUpdated)); err != nil {
		t.Fatalf("reset expected tags: %v", err)
	}
	if _, _, err := store.ApplyAssetUndoableOperation(ctx, operation.ID, ports.UndoableOperationDirectionUndo, updated, original, auditRecord(t, "audit-undo-edit", tenantID, inventoryID, audit.ActionUndoableOperationUndone)); err != nil {
		t.Fatalf("undo edit: %v", err)
	}
	assigned, err := store.AssetTagsByAsset(ctx, tenantID, inventoryID, original.ID)
	if err != nil {
		t.Fatalf("read restored tags: %v", err)
	}
	if len(assigned) != 1 || assigned[0].ID != workshop.ID {
		t.Fatalf("expected workshop tag restored, got %+v", assigned)
	}
	if err := store.SetAssetTags(ctx, tenantID, inventoryID, original.ID, []assettag.ID{camping.ID}, auditRecord(t, "audit-newer-tags-before-redo", tenantID, inventoryID, audit.ActionAssetUpdated)); err != nil {
		t.Fatalf("set newer tags before redo: %v", err)
	}
	if _, _, err := store.ApplyAssetUndoableOperation(ctx, operation.ID, ports.UndoableOperationDirectionRedo, original, updated, auditRecord(t, "audit-stale-tag-redo", tenantID, inventoryID, audit.ActionUndoableOperationRedone)); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("expected newer tags to make redo stale, got %v", err)
	}
}

func TestStoreRollsBackUndoableAssetApplyWhenAuditInsertFails(t *testing.T) {
	ctx := context.Background()
	store := newUndoableOperationTestStore(t, ctx)
	item := assetItem("asset-one", "tenant-one", "inventory-one", asset.KindItem, "")
	operation := undoableAssetOperation("operation-one", tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), audit.ActionAssetCreated, nil, item)
	createAudit := auditRecord(t, "audit-create", tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), audit.ActionAssetCreated)
	if err := store.CreateAsset(ctx, item, createAudit, &operation); err != nil {
		t.Fatalf("create asset with undoable operation: %v", err)
	}

	resulting := item
	resulting.LifecycleState = asset.LifecycleStateArchived
	_, _, err := store.ApplyAssetUndoableOperation(ctx, operation.ID, ports.UndoableOperationDirectionUndo, item, resulting, createAudit)
	if err == nil {
		t.Fatalf("expected duplicate audit ID failure")
	}

	persisted, ok, err := store.AssetByID(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), item.ID)
	if err != nil {
		t.Fatalf("find asset after failed undo: %v", err)
	}
	if !ok || persisted.LifecycleState != asset.LifecycleStateActive {
		t.Fatalf("expected asset rollback after failed undo, found=%t asset=%+v", ok, persisted)
	}
	persistedOperation, found, err := store.UndoableOperationByID(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), operation.ID)
	if err != nil {
		t.Fatalf("find operation after failed undo: %v", err)
	}
	if !found || persistedOperation.Status != ports.UndoableOperationAvailable {
		t.Fatalf("expected operation rollback after failed undo, found=%t operation=%+v", found, persistedOperation)
	}
}

func TestStoreRejectsStaleAndInvalidUndoableAssetOperations(t *testing.T) {
	ctx := context.Background()
	store := newUndoableOperationTestStore(t, ctx)
	item := assetItem("asset-one", "tenant-one", "inventory-one", asset.KindItem, "")
	operation := undoableAssetOperation("operation-one", tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), audit.ActionAssetCreated, nil, item)
	if err := store.CreateAsset(ctx, item, auditRecord(t, "audit-create", tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), audit.ActionAssetCreated), &operation); err != nil {
		t.Fatalf("create asset with undoable operation: %v", err)
	}

	resulting := item
	resulting.LifecycleState = asset.LifecycleStateArchived
	_, _, err := store.ApplyAssetUndoableOperation(ctx, operation.ID, ports.UndoableOperationDirectionRedo, item, resulting, auditRecord(t, "audit-redo-invalid", tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), audit.ActionUndoableOperationRedone))
	if !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("expected invalid redo transition conflict, got %v", err)
	}

	stale := item
	title, ok := asset.NewTitle("Changed")
	if !ok {
		t.Fatalf("invalid title")
	}
	stale.Title = title
	if err := updateAsset(t, ctx, store, stale); err != nil {
		t.Fatalf("make asset stale: %v", err)
	}
	_, _, err = store.ApplyAssetUndoableOperation(ctx, operation.ID, ports.UndoableOperationDirectionUndo, item, resulting, auditRecord(t, "audit-undo-stale", tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), audit.ActionUndoableOperationUndone))
	if !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("expected stale state conflict, got %v", err)
	}

	_, found, err := store.UndoableOperationByID(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-two"), operation.ID)
	if err != nil {
		t.Fatalf("find operation in wrong inventory: %v", err)
	}
	if found {
		t.Fatalf("expected scoped lookup to hide operation from wrong inventory")
	}
}

func newUndoableOperationTestStore(t *testing.T, ctx context.Context) Store {
	t.Helper()

	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-one"), "Home")
	saveInventory(t, ctx, store, "inventory-one", tenant.ID("tenant-one"), "Tools")
	saveInventory(t, ctx, store, "inventory-two", tenant.ID("tenant-one"), "Medicine")
	return store
}

func undoableAssetOperation(id string, tenantID tenant.ID, inventoryID inventory.InventoryID, action audit.Action, before *asset.Asset, after asset.Asset) ports.UndoableOperation {
	return ports.UndoableOperation{
		ID:             id,
		TenantID:       tenantID,
		InventoryID:    inventoryID,
		PrincipalID:    identity.PrincipalID("owner"),
		Source:         audit.SourceAPI,
		TargetType:     audit.TargetAsset,
		TargetID:       after.ID.String(),
		OriginalAction: action,
		Status:         ports.UndoableOperationAvailable,
		CreatedAt:      time.Now().UTC(),
		BeforeAsset:    before,
		AfterAsset:     after,
	}
}
