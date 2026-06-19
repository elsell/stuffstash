package app

import (
	"context"
	"errors"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestListInventoriesReturnsAuthorizationBackendFailures(t *testing.T) {
	expected := errors.New("authorization backend unavailable")
	application := New(Dependencies{
		Observer: &fakeObserver{},
		Authorizer: &fakeAuthorizer{
			checkInventoryErr: expected,
		},
		Tenants: &fakeTenantRepository{
			exists: true,
		},
		Inventories: &fakeInventoryRepository{
			items: []inventory.Inventory{
				inventoryItem("inventory-one", "tenant-one", "Tools"),
			},
		},
	})

	_, err := application.ListInventories(context.Background(), ListInventoriesInput{
		Principal: identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:  tenant.ID("tenant-one"),
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected backend error, got %v", err)
	}
}

func TestListInventoriesSkipsForbiddenInventories(t *testing.T) {
	application := New(Dependencies{
		Observer: &fakeObserver{},
		Authorizer: &fakeAuthorizer{
			checkInventoryErr: ports.ErrForbidden,
		},
		Tenants: &fakeTenantRepository{
			exists: true,
		},
		Inventories: &fakeInventoryRepository{
			items: []inventory.Inventory{
				inventoryItem("inventory-one", "tenant-one", "Tools"),
			},
		},
	})

	items, err := application.ListInventories(context.Background(), ListInventoriesInput{
		Principal: identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:  tenant.ID("tenant-one"),
	})
	if err != nil {
		t.Fatalf("list inventories: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected forbidden inventory to be hidden, got %+v", items)
	}
}

type fakeAuthorizer struct {
	checkInventoryErr error
}

func (f *fakeAuthorizer) CheckTenant(context.Context, identity.Principal, ports.TenantPermission, tenant.ID) error {
	return nil
}

func (f *fakeAuthorizer) CheckInventory(context.Context, identity.Principal, ports.InventoryPermission, inventory.InventoryID) error {
	return f.checkInventoryErr
}

func (f *fakeAuthorizer) GrantTenantOwner(context.Context, identity.Principal, tenant.ID) error {
	return nil
}

func (f *fakeAuthorizer) GrantInventoryOwner(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

type fakeTenantRepository struct {
	exists bool
}

func (f *fakeTenantRepository) SaveTenant(context.Context, tenant.Tenant) error {
	return nil
}

func (f *fakeTenantRepository) TenantExists(context.Context, tenant.ID) (bool, error) {
	return f.exists, nil
}

type fakeInventoryRepository struct {
	items []inventory.Inventory
}

func (f *fakeInventoryRepository) SaveInventory(context.Context, inventory.Inventory) error {
	return nil
}

func (f *fakeInventoryRepository) ListInventoriesByTenant(context.Context, inventory.TenantID) ([]inventory.Inventory, error) {
	return f.items, nil
}

type fakeObserver struct{}

func (f *fakeObserver) Record(context.Context, ports.Event) {}

func inventoryItem(id string, tenantID string, name string) inventory.Inventory {
	inventoryName, ok := inventory.NewName(name)
	if !ok {
		panic("invalid test inventory name")
	}
	return inventory.Inventory{
		ID:       inventory.InventoryID(id),
		TenantID: inventory.TenantID(tenantID),
		Name:     inventoryName,
	}
}
