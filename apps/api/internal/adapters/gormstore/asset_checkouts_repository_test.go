package gormstore

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestCheckOutAssetDuplicateOpenCheckoutReturnsConflict(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Home")
	item := assetItem("asset-one", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	if err := createAsset(t, ctx, store, item); err != nil {
		t.Fatalf("create asset: %v", err)
	}
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	first := checkoutRecord("checkout-one", item, now)
	if err := store.CheckOutAsset(ctx, first, auditRecord(t, "audit-checkout-one", tenantID, inventoryID, audit.ActionAssetCheckedOut), nil); err != nil {
		t.Fatalf("checkout first: %v", err)
	}

	second := checkoutRecord("checkout-two", item, now.Add(time.Minute))
	err := store.CheckOutAsset(ctx, second, auditRecord(t, "audit-checkout-two", tenantID, inventoryID, audit.ActionAssetCheckedOut), nil)
	if !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("expected duplicate checkout conflict, got %v", err)
	}
}

func checkoutRecord(id string, item asset.Asset, checkedOutAt time.Time) asset.Checkout {
	return asset.Checkout{
		ID:                    asset.CheckoutID(id),
		TenantID:              item.TenantID,
		InventoryID:           item.InventoryID,
		AssetID:               item.ID,
		State:                 asset.CheckoutStateOpen,
		CheckedOutAt:          checkedOutAt,
		CheckedOutByPrincipal: "editor-one",
		CreatedAt:             checkedOutAt,
		UpdatedAt:             checkedOutAt,
	}
}
