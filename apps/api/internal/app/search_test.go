package app

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestSearchAssetsUsesAuthorizationVisibilityPort(t *testing.T) {
	search := &recordingAssetSearchRepository{}
	authorizer := &visibilityAuthorizer{
		t:       t,
		visible: []inventory.InventoryID{inventory.InventoryID("inventory-two")},
	}
	application := New(Dependencies{
		Observer:   &fakeObserver{},
		Authorizer: authorizer,
		Tenants:    &fakeTenantRepository{exists: true},
		Inventories: &fakeInventoryRepository{items: []inventory.Inventory{
			inventoryItem("inventory-one", "tenant-one", "Tools"),
			inventoryItem("inventory-two", "tenant-one", "Medicine"),
		}},
		Search:           search,
		DefaultPageLimit: 1,
		MaxPageLimit:     10,
	})

	_, err := application.SearchAssets(context.Background(), SearchAssetsInput{
		Principal: identity.Principal{ID: identity.PrincipalID("viewer")},
		TenantID:  tenant.ID("tenant-one"),
		Query:     "aspirin",
		Mode:      "exact",
	})
	if err != nil {
		t.Fatalf("search assets: %v", err)
	}

	if !authorizer.visibilityCalled {
		t.Fatalf("expected search to use authorization visibility port")
	}
	if len(authorizer.candidates) != 2 {
		t.Fatalf("expected two candidate inventories, got %+v", authorizer.candidates)
	}
	if len(search.inventoryIDs) != 1 || search.inventoryIDs[0] != inventory.InventoryID("inventory-two") {
		t.Fatalf("expected search repository to receive visible inventory IDs only, got %+v", search.inventoryIDs)
	}
}

type visibilityAuthorizer struct {
	t                *testing.T
	visible          []inventory.InventoryID
	candidates       []inventory.InventoryID
	visibilityCalled bool
}

func (v *visibilityAuthorizer) CheckTenant(context.Context, identity.Principal, ports.TenantPermission, tenant.ID) error {
	return nil
}

func (v *visibilityAuthorizer) CheckInventory(context.Context, identity.Principal, ports.InventoryPermission, inventory.InventoryID) error {
	v.t.Fatalf("search must use ListViewableInventoryIDs instead of per-inventory checks")
	return nil
}

func (v *visibilityAuthorizer) ListViewableInventoryIDs(_ context.Context, _ identity.Principal, _ tenant.ID, candidates []inventory.InventoryID) ([]inventory.InventoryID, error) {
	v.visibilityCalled = true
	v.candidates = append([]inventory.InventoryID{}, candidates...)
	return append([]inventory.InventoryID{}, v.visible...), nil
}

func (v *visibilityAuthorizer) GrantTenantOwner(context.Context, identity.Principal, tenant.ID) error {
	return nil
}

func (v *visibilityAuthorizer) GrantInventoryOwner(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (v *visibilityAuthorizer) GrantInventoryViewer(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (v *visibilityAuthorizer) GrantInventoryEditor(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (v *visibilityAuthorizer) RevokeInventoryViewer(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (v *visibilityAuthorizer) RevokeInventoryEditor(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

type recordingAssetSearchRepository struct {
	inventoryIDs []inventory.InventoryID
}

func (r *recordingAssetSearchRepository) SearchAssets(_ context.Context, _ tenant.ID, inventoryIDs []inventory.InventoryID, _ ports.AssetSearchPageRequest) ([]ports.AssetSearchResult, error) {
	r.inventoryIDs = append([]inventory.InventoryID{}, inventoryIDs...)
	return []ports.AssetSearchResult{}, nil
}
