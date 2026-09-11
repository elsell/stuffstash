package app

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func TestExpirationTypeCapabilityNoOpAndOmittedPreservation(t *testing.T) {
	repository := &fakeCustomAssetTypeRepository{}
	application := New(Dependencies{
		Authorizer: &fakeAuthorizer{}, Tenants: &fakeTenantRepository{exists: true},
		Inventories:      &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Home")}},
		CustomAssetTypes: repository, CustomAssetTypeUnitOfWork: repository,
		IDs: &fakeIDGenerator{ids: []string{"type-one", "audit-create", "audit-rename"}},
	})
	principal := identity.Principal{ID: identity.PrincipalID("owner")}
	value, err := application.CreateInventoryCustomAssetType(context.Background(), CreateCustomAssetTypeInput{
		Principal: principal, TenantID: tenant.ID("tenant-one"), InventoryID: inventory.InventoryID("inventory-one"), Key: "medicine", DisplayName: "Medicine", ExpirationEnabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	enabled := true
	input := UpdateCustomAssetTypeInput{Principal: principal, TenantID: tenant.ID("tenant-one"), InventoryID: inventory.InventoryID("inventory-one"), CustomAssetTypeID: value.ID, ExpirationEnabled: &enabled}
	if _, err := application.UpdateInventoryCustomAssetType(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	if len(repository.auditRecords) != 1 {
		t.Fatal("no-op generated an audit write")
	}
	name := "Medication"
	input.ExpirationEnabled = nil
	input.DisplayName = &name
	value, err = application.UpdateInventoryCustomAssetType(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if !value.ExpirationEnabled || len(repository.auditRecords) != 2 {
		t.Fatal("omitted capability was reset or rename not audited")
	}
}
