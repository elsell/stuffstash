package app

import (
	"context"
	"errors"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestUpdateAssetRejectsInvalidMovement(t *testing.T) {
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		asset.ID("garage"):   assetItem("garage", "tenant-one", "inventory-one", asset.KindLocation, ""),
		asset.ID("shelf"):    assetItem("shelf", "tenant-one", "inventory-one", asset.KindLocation, "garage"),
		asset.ID("box"):      assetItem("box", "tenant-one", "inventory-one", asset.KindContainer, "shelf"),
		asset.ID("wrench"):   assetItem("wrench", "tenant-one", "inventory-one", asset.KindItem, "box"),
		asset.ID("supplies"): assetItem("supplies", "tenant-one", "inventory-one", asset.KindItem, ""),
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

	for _, item := range []struct {
		name    string
		assetID asset.ID
		parent  string
	}{
		{name: "self parent", assetID: asset.ID("box"), parent: "box"},
		{name: "cycle through descendant", assetID: asset.ID("garage"), parent: "box"},
		{name: "item parent", assetID: asset.ID("wrench"), parent: "supplies"},
	} {
		t.Run(item.name, func(t *testing.T) {
			_, err := application.UpdateAsset(context.Background(), UpdateAssetInput{
				Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
				TenantID:    tenant.ID("tenant-one"),
				InventoryID: inventory.InventoryID("inventory-one"),
				AssetID:     item.assetID,
				ParentAssetID: AssetParentUpdate{
					Present: true,
					Value:   item.parent,
				},
			})
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected invalid movement rejection, got %v", err)
			}
		})
	}

	updated, err := application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("box"),
		ParentAssetID: AssetParentUpdate{
			Present: true,
			Null:    true,
		},
	})
	if err != nil {
		t.Fatalf("move container to root: %v", err)
	}
	if updated.ParentAssetID.String() != "" || assets.items[asset.ID("wrench")].ParentAssetID != asset.ID("box") {
		t.Fatalf("expected box moved to root with child preserved, box=%+v wrench=%+v", updated, assets.items[asset.ID("wrench")])
	}

	_, err = application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("box"),
		ParentAssetID: AssetParentUpdate{
			Present: true,
			Value:   " ",
		},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected blank parent rejection, got %v", err)
	}
}

func TestUpdateAssetRequiresEditPermission(t *testing.T) {
	application := New(Dependencies{
		Observer: &fakeObserver{},
		Authorizer: &fakeAuthorizer{
			checkInventoryErr: ports.ErrForbidden,
		},
		Tenants:               &fakeTenantRepository{exists: true},
		TenantUnitOfWork:      &fakeTenantRepository{exists: true},
		Inventories:           &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields:          &fakeCustomFieldRepository{},
		CustomFieldUnitOfWork: &fakeCustomFieldRepository{},
		Assets: &fakeAssetRepository{items: map[asset.ID]asset.Asset{
			asset.ID("drill"): assetItem("drill", "tenant-one", "inventory-one", asset.KindItem, ""),
		}},
		Audit:  &fakeAuditRepository{},
		Outbox: &fakeOutbox{},
		IDs:    &fakeIDGenerator{},
	})

	title := "Cordless Drill"
	_, err := application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("viewer")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("drill"),
		Title:       &title,
	})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized update, got %v", err)
	}
}
