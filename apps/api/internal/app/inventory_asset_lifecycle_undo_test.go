package app

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func TestUndoAndRedoAssetLifecycleOperations(t *testing.T) {
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		asset.ID("drill"): assetItem("drill", "tenant-one", "inventory-one", asset.KindItem, ""),
	}}
	application := New(Dependencies{
		Observer:              &fakeObserver{},
		Authorizer:            &fakeAuthorizer{},
		Tenants:               &fakeTenantRepository{exists: true},
		TenantUnitOfWork:      &fakeTenantRepository{exists: true},
		Inventories:           &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields:          &fakeCustomFieldRepository{},
		CustomFieldUnitOfWork: &fakeCustomFieldRepository{},
		Assets:                assets,
		AssetUnitOfWork:       assets,
		Undoables:             assets,
		Audit:                 &fakeAuditRepository{},
		Outbox:                &fakeOutbox{},
		IDs:                   &fakeIDGenerator{},
	})

	archiveResult, err := application.ArchiveAssetWithOperation(context.Background(), UpdateAssetLifecycleInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("drill"),
	})
	if err != nil {
		t.Fatalf("archive asset: %v", err)
	}
	archived := archiveResult.Asset
	if archived.LifecycleState != asset.LifecycleStateArchived {
		t.Fatalf("expected archived asset, got %+v", archived)
	}
	archiveOperationID := assets.auditRecords[len(assets.auditRecords)-1].Metadata["operation_id"]

	undoArchive, err := application.UndoOperation(context.Background(), ApplyUndoableOperationInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		OperationID: archiveOperationID,
	})
	if err != nil {
		t.Fatalf("undo archive: %v", err)
	}
	if undoArchive.LifecycleState != asset.LifecycleStateActive {
		t.Fatalf("expected undo archive to restore active state, got %+v", undoArchive)
	}

	redoArchive, err := application.RedoOperation(context.Background(), ApplyUndoableOperationInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		OperationID: archiveOperationID,
	})
	if err != nil {
		t.Fatalf("redo archive: %v", err)
	}
	if redoArchive.LifecycleState != asset.LifecycleStateArchived {
		t.Fatalf("expected redo archive to archive asset, got %+v", redoArchive)
	}

	restoreResult, err := application.RestoreAssetWithOperation(context.Background(), UpdateAssetLifecycleInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("drill"),
	})
	if err != nil {
		t.Fatalf("restore asset: %v", err)
	}
	restored := restoreResult.Asset
	if restored.LifecycleState != asset.LifecycleStateActive {
		t.Fatalf("expected restored asset, got %+v", restored)
	}
	restoreOperationID := assets.auditRecords[len(assets.auditRecords)-1].Metadata["operation_id"]

	undoRestore, err := application.UndoOperation(context.Background(), ApplyUndoableOperationInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		OperationID: restoreOperationID,
	})
	if err != nil {
		t.Fatalf("undo restore: %v", err)
	}
	if undoRestore.LifecycleState != asset.LifecycleStateArchived {
		t.Fatalf("expected undo restore to archive asset, got %+v", undoRestore)
	}
}
