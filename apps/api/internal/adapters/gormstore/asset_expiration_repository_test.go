package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

func TestStoreExpirationDateRoundTripAndUndo(t *testing.T) {
	ctx := context.Background()
	store := newUndoableOperationTestStore(t, ctx)
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	original := assetItem("asset-one", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	kind := customAssetType(t, "type-one", tenantID.String(), inventoryID.String(), customfield.ScopeInventory, "medicine")
	kind.ExpirationEnabled = true
	if err := saveCustomAssetType(t, ctx, store, kind); err != nil {
		t.Fatal(err)
	}
	original.CustomAssetTypeID = asset.CustomAssetTypeID(kind.ID.String())
	var err error
	original.Expiration, err = expirationdate.ParseDate("2028-02", expirationdate.Month)
	if err != nil {
		t.Fatal(err)
	}
	if err := createAsset(t, ctx, store, original); err != nil {
		t.Fatal(err)
	}
	saved, found, err := store.AssetByID(ctx, tenantID, inventoryID, original.ID)
	if err != nil || !found || saved.Expiration != original.Expiration {
		t.Fatalf("date round trip: %+v %v", saved, err)
	}
	updated := saved
	updated.Expiration, err = expirationdate.ParseDate("2028-02-28", expirationdate.Day)
	if err != nil {
		t.Fatal(err)
	}
	operation := undoableAssetOperation("operation-edit", tenantID, inventoryID, audit.ActionAssetUpdated, &saved, updated)
	if err := store.UpdateAsset(ctx, updated, []audit.Record{auditRecord(t, "audit-edit", tenantID, inventoryID, audit.ActionAssetUpdated)}, &operation); err != nil {
		t.Fatal(err)
	}
	loaded, found, err := store.UndoableOperationByID(ctx, tenantID, inventoryID, operation.ID)
	if err != nil || !found || loaded.BeforeAsset == nil || loaded.BeforeAsset.Expiration != original.Expiration || loaded.AfterAsset.Expiration != updated.Expiration {
		t.Fatalf("undo snapshot lost expiration: %+v %v", loaded, err)
	}
	if _, _, err := store.ApplyAssetUndoableOperation(ctx, operation.ID, ports.UndoableOperationDirectionUndo, updated, saved, auditRecord(t, "audit-undo", tenantID, inventoryID, audit.ActionUndoableOperationUndone)); err != nil {
		t.Fatal(err)
	}
	restored, found, err := store.AssetByID(ctx, tenantID, inventoryID, original.ID)
	if err != nil || !found || restored.Expiration != original.Expiration {
		t.Fatalf("undo lost date: %+v %v", restored, err)
	}
}

func TestDatedAssetScanIsScopedAndExcludesUndated(t *testing.T) {
	ctx := context.Background()
	store := newUndoableOperationTestStore(t, ctx)
	dated := assetItem("dated", "tenant-one", "inventory-one", asset.KindItem, "")
	dated.Expiration, _ = expirationdate.ParseDate("2028-02", expirationdate.Month)
	undated := assetItem("undated", "tenant-one", "inventory-one", asset.KindItem, "")
	for _, item := range []asset.Asset{dated, undated} {
		if err := createAsset(t, ctx, store, item); err != nil {
			t.Fatal(err)
		}
	}
	items, err := store.ListAssetsByInventory(ctx, "tenant-one", "inventory-one", ports.AssetListPageRequest{OnlyDated: true, Limit: 10})
	if err != nil || len(items) != 1 || items[0].ID != dated.ID {
		t.Fatalf("dated scan: %+v %v", items, err)
	}
	items, err = store.ListAssetsByInventory(ctx, "other", "inventory-one", ports.AssetListPageRequest{OnlyDated: true, Limit: 10})
	if err != nil || len(items) != 0 {
		t.Fatal("cross-tenant dated scan leaked")
	}
	items, err = store.ListAssetsByInventory(ctx, "tenant-one", "inventory-one", ports.AssetListPageRequest{OnlyDated: true, AfterAssetID: dated.ID, Limit: 10})
	if err != nil || len(items) != 0 {
		t.Fatal("dated cursor ignored")
	}
}
