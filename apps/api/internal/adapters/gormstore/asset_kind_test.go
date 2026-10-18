package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestAssetKindConversionConstraintsAndUndo(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tid := tenant.ID("tenant-one")
	iid := inventory.InventoryID("inventory-one")
	saveTenant(t, ctx, store, tid, "Home")
	saveInventory(t, ctx, store, iid.String(), tid, "Main")
	original := assetItem("bin", tid.String(), iid.String(), asset.KindItem, "")
	if err := createAsset(t, ctx, store, original); err != nil {
		t.Fatal(err)
	}
	converted := original
	converted.Kind = asset.KindContainer
	op := undoableAssetOperation("conversion", tid, iid, audit.ActionAssetUpdated, &original, converted)
	if err := store.UpdateAsset(ctx, converted, []audit.Record{auditRecord(t, "convert", tid, iid, audit.ActionAssetUpdated)}, &op); err != nil {
		t.Fatal(err)
	}

	if _, _, err := store.ApplyAssetUndoableOperation(ctx, op.ID, ports.UndoableOperationDirectionUndo, converted, original, auditRecord(t, "successful-undo", tid, iid, audit.ActionUndoableOperationUndone)); err != nil {
		t.Fatal(err)
	}
	persisted, _, err := store.AssetByID(ctx, tid, iid, original.ID)
	if err != nil || persisted.Kind != asset.KindItem {
		t.Fatalf("kind undo not persisted: %v %s", err, persisted.Kind)
	}
	if _, _, err := store.ApplyAssetUndoableOperation(ctx, op.ID, ports.UndoableOperationDirectionRedo, original, converted, auditRecord(t, "successful-redo", tid, iid, audit.ActionUndoableOperationRedone)); err != nil {
		t.Fatal(err)
	}
	child := assetItem("child", tid.String(), iid.String(), asset.KindItem, "bin")
	if err := createAsset(t, ctx, store, child); err != nil {
		t.Fatal(err)
	}
	child.LifecycleState = asset.LifecycleStateArchived
	if err := store.UpdateAssetLifecycle(ctx, child, auditRecord(t, "archive", tid, iid, audit.ActionAssetArchived), nil); err != nil {
		t.Fatal(err)
	}
	if err := store.UpdateAsset(ctx, original, []audit.Record{auditRecord(t, "invalid", tid, iid, audit.ActionAssetUpdated)}, nil); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("archived child: %v", err)
	}
	if _, _, err := store.ApplyAssetUndoableOperation(ctx, op.ID, ports.UndoableOperationDirectionUndo, converted, original, auditRecord(t, "undo", tid, iid, audit.ActionAssetUpdated)); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("undo with contents: %v", err)
	}
	checkout := checkoutRecord("checkout", converted, time.Now())
	if err := store.CheckOutAsset(ctx, checkout, auditRecord(t, "checkout-audit", tid, iid, audit.ActionAssetCheckedOut), nil); err != nil {
		t.Fatal(err)
	}
	location := converted
	location.Kind = asset.KindLocation
	if err := store.UpdateAsset(ctx, location, []audit.Record{auditRecord(t, "invalid-location", tid, iid, audit.ActionAssetUpdated)}, nil); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("location with checkout: %v", err)
	}
}
