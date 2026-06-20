package gormstore

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
