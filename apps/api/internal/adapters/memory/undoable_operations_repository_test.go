package memory

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

func TestApplyAssetUndoableOperationRejectsNewerTagAssignments(t *testing.T) {
	for _, test := range []struct {
		name      string
		direction ports.UndoableOperationDirection
		status    ports.UndoableOperationStatus
		action    audit.Action
	}{
		{name: "undo", direction: ports.UndoableOperationDirectionUndo, status: ports.UndoableOperationAvailable, action: audit.ActionUndoableOperationUndone},
		{name: "redo", direction: ports.UndoableOperationDirectionRedo, status: ports.UndoableOperationUndone, action: audit.ActionUndoableOperationRedone},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := NewStore()
			tenantID := tenant.ID("tenant-one")
			inventoryID := inventory.InventoryID("inventory-one")
			assetID := asset.ID("asset-one")
			current := asset.Asset{ID: assetID, TenantID: asset.TenantID(tenantID), InventoryID: asset.InventoryID(inventoryID), Kind: asset.KindItem, LifecycleState: asset.LifecycleStateActive}
			store.assets[assetID] = current
			store.assetTagLinks[assetID] = map[assettag.ID]struct{}{assettag.ID("tag-newer"): {}}
			store.undoables["operation-one"] = ports.UndoableOperation{ID: "operation-one", TenantID: tenantID, InventoryID: inventoryID, TargetID: assetID.String(), Status: test.status, ReplacesTags: true, BeforeTagIDs: []assettag.ID{"tag-before"}, AfterTagIDs: []assettag.ID{"tag-after"}}
			record, ok := audit.NewRecord(audit.ID("audit-"+test.name), audit.TenantID(tenantID), audit.InventoryID(inventoryID), "principal-one", test.action, audit.SourceAPI, audit.TargetAsset, assetID.String(), time.Now(), "request-one", nil)
			if !ok {
				t.Fatal("invalid audit fixture")
			}
			_, _, err := store.ApplyAssetUndoableOperation(context.Background(), "operation-one", test.direction, current, current, record)
			if !errors.Is(err, ports.ErrConflict) {
				t.Fatalf("expected stale tag conflict, got %v", err)
			}
		})
	}
}
