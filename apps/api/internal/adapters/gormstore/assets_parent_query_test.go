package gormstore

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestAssetParentPredicatePrecedesPaginationAndPreservesScope(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("home")
	inventoryID := inventory.InventoryID("main")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, "main", tenantID, "Main")
	saveInventory(t, ctx, store, "other", tenantID, "Other")
	for _, item := range []asset.Asset{
		assetItem("a-parent", "home", "main", asset.KindLocation, ""),
		assetItem("b-other-root", "home", "main", asset.KindItem, ""),
		assetItem("c-child", "home", "main", asset.KindItem, "a-parent"),
		assetItem("d-child", "home", "main", asset.KindItem, "a-parent"),
		assetItem("e-other-inventory", "home", "other", asset.KindItem, ""),
	} {
		if err := createAsset(t, ctx, store, item); err != nil {
			t.Fatal(err)
		}
	}
	children, err := store.ListAssetsByInventory(ctx, tenantID, inventoryID, ports.AssetListPageRequest{Limit: 1, Parent: ports.AssetParentFilter{Applied: true, ID: "a-parent"}})
	if err != nil || len(children) != 1 || children[0].ID != "c-child" {
		t.Fatalf("first child: %v %v", children, err)
	}
	next, err := store.ListAssetsByInventory(ctx, tenantID, inventoryID, ports.AssetListPageRequest{Limit: 1, AfterAssetID: children[0].ID, Parent: ports.AssetParentFilter{Applied: true, ID: "a-parent"}})
	if err != nil || len(next) != 1 || next[0].ID != "d-child" {
		t.Fatalf("next child: %v %v", next, err)
	}
	roots, err := store.ListAssetsByInventory(ctx, tenantID, inventoryID, ports.AssetListPageRequest{Limit: 10, Parent: ports.AssetParentFilter{Applied: true}})
	if err != nil || len(roots) != 2 {
		t.Fatalf("roots: %v %v", roots, err)
	}
	hidden, err := store.ListAssetsByInventory(ctx, "other-tenant", inventoryID, ports.AssetListPageRequest{Limit: 10, Parent: ports.AssetParentFilter{Applied: true}})
	if err != nil || len(hidden) != 0 {
		t.Fatalf("cross-tenant: %v %v", hidden, err)
	}
}
