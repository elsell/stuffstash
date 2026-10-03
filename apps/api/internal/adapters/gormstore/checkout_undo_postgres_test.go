package gormstore

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestPostgresCheckoutReturnUndoTimestampPrecision(t *testing.T) {
	dsn := os.Getenv("STUFF_STASH_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	db, err := OpenPostgres(dsn)
	if err != nil {
		t.Fatal(err)
	}
	connection, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	if err = runEmbeddedPostgresMigrations(db); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	store := NewStore(db)
	tid := tenant.ID("precision-tenant")
	iid := inventory.InventoryID("precision-inv")
	t.Cleanup(func() {
		for _, model := range []any{&undoableOperationModel{}, &auditRecordModel{}, &assetCheckoutModel{}, &assetModel{}, &inventoryModel{}, &tenantModel{}} {
			column := "tenant_id"
			if _, ok := model.(*tenantModel); ok {
				column = "id"
			}
			if err := db.Where(map[string]any{column: tid.String()}).Delete(model).Error; err != nil {
				t.Error(err)
			}
		}
	})
	saveTenant(t, ctx, store, tid, "Home")
	saveInventory(t, ctx, store, iid.String(), tid, "Tools")
	item := assetItem("precision-asset", tid.String(), iid.String(), asset.KindItem, "")
	if err = createAsset(t, ctx, store, item); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 3, 8, 0, 0, 123456789, time.UTC)
	before := checkoutRecord("precision-record", item, now)
	if err = store.CheckOutAsset(ctx, before, auditRecord(t, "precision-checkout", tid, iid, audit.ActionAssetCheckedOut), nil); err != nil {
		t.Fatal(err)
	}
	// Return reads the persisted checkout before capturing the undo snapshot.
	before, _, err = store.AssetCheckoutByID(ctx, tid, iid, before.ID)
	if err != nil {
		t.Fatal(err)
	}
	after := before
	after.State = asset.CheckoutStateReturned
	after.ReturnedAt = now.Add(time.Minute)
	after.UpdatedAt = after.ReturnedAt
	after.ReturnedByPrincipal = "editor-one"
	op := ports.UndoableOperation{ID: "precision-return-op", PrincipalID: identity.PrincipalID("editor-one"), Source: audit.SourceAPI, TenantID: tid, InventoryID: iid, OriginalAction: audit.ActionAssetReturned, TargetType: audit.TargetAsset, TargetID: item.ID.String(), Status: ports.UndoableOperationAvailable, BeforeCheckout: &before, AfterCheckout: &after, CreatedAt: now}
	if err = store.ReturnAsset(ctx, before, after, auditRecord(t, "precision-return", tid, iid, audit.ActionAssetReturned), &op); err != nil {
		t.Fatal(err)
	}
	stored, found, err := store.UndoableOperationByID(ctx, tid, iid, op.ID)
	if err != nil || !found {
		t.Fatalf("read operation: found=%t err=%v", found, err)
	}
	wrongScope := *stored.AfterCheckout
	wrongScope.InventoryID = asset.InventoryID("another-inventory")
	if _, _, err = store.ApplyAssetCheckoutUndoableOperation(ctx, op.ID, ports.UndoableOperationDirectionUndo, wrongScope, *stored.BeforeCheckout, auditRecord(t, "precision-wrong-scope", tid, iid, audit.ActionUndoableOperationUndone)); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("wrong inventory accepted: %v", err)
	}
	wrongTime := *stored.AfterCheckout
	wrongTime.UpdatedAt = wrongTime.UpdatedAt.Add(time.Microsecond)
	if _, _, err = store.ApplyAssetCheckoutUndoableOperation(ctx, op.ID, ports.UndoableOperationDirectionUndo, wrongTime, *stored.BeforeCheckout, auditRecord(t, "precision-wrong-time", tid, iid, audit.ActionUndoableOperationUndone)); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("timestamp-only mismatch accepted: %v", err)
	}
	wrongDetails := *stored.AfterCheckout
	wrongDetails.ReturnDetails, _ = asset.NewCheckoutDetails("not the stored details")
	if _, _, err = store.ApplyAssetCheckoutUndoableOperation(ctx, op.ID, ports.UndoableOperationDirectionUndo, wrongDetails, *stored.BeforeCheckout, auditRecord(t, "precision-wrong-details", tid, iid, audit.ActionUndoableOperationUndone)); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("changed details accepted: %v", err)
	}
	if _, _, err = store.ApplyAssetCheckoutUndoableOperation(ctx, op.ID, ports.UndoableOperationDirectionUndo, *stored.AfterCheckout, *stored.BeforeCheckout, auditRecord(t, "precision-undo", tid, iid, audit.ActionUndoableOperationUndone)); err != nil {
		t.Fatalf("cancel unchanged return: %v", err)
	}
	reopened, found, err := store.AssetCheckoutByID(ctx, tid, iid, before.ID)
	if err != nil || !found || reopened.State != asset.CheckoutStateOpen || !reopened.ReturnedAt.IsZero() {
		t.Fatalf("checkout not reopened: %+v %v", reopened, err)
	}
	if _, _, err = store.ApplyAssetCheckoutUndoableOperation(ctx, op.ID, ports.UndoableOperationDirectionRedo, *stored.BeforeCheckout, *stored.AfterCheckout, auditRecord(t, "precision-redo", tid, iid, audit.ActionUndoableOperationRedone)); err != nil {
		t.Fatalf("redo return: %v", err)
	}
	changed := after
	changed.ReturnDetails, _ = asset.NewCheckoutDetails("later edit")
	changed.UpdatedAt = after.UpdatedAt.Add(time.Microsecond)
	if err = store.UpdateAssetCheckoutReturnDetails(ctx, after, changed, auditRecord(t, "precision-details", tid, iid, audit.ActionAssetReturnDetailsUpdated)); err != nil {
		t.Fatal(err)
	}
	if _, _, err = store.ApplyAssetCheckoutUndoableOperation(ctx, op.ID, ports.UndoableOperationDirectionUndo, *stored.AfterCheckout, *stored.BeforeCheckout, auditRecord(t, "precision-stale", tid, iid, audit.ActionUndoableOperationUndone)); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("expected stale-edit rejection: %v", err)
	}
}
