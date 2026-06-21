package app

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) ensureActiveInventoryAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID, permission ports.InventoryPermission) error {
	item, err := a.ensureInventoryAccessItem(ctx, principal, tenantID, inventoryID, permission)
	if err != nil {
		return err
	}
	if !item.IsActive() {
		return ErrNotFound
	}
	return nil
}

func (a App) ensureInventoryAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID, permission ports.InventoryPermission) error {
	_, err := a.ensureInventoryAccessItem(ctx, principal, tenantID, inventoryID, permission)
	return err
}

func (a App) ensureInventoryAccessItem(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID, permission ports.InventoryPermission) (inventory.Inventory, error) {
	exists, err := a.tenants.TenantExists(ctx, tenantID)
	if err != nil {
		return inventory.Inventory{}, err
	}
	if !exists {
		return inventory.Inventory{}, ErrNotFound
	}

	item, found, err := a.inventories.InventoryByID(ctx, tenantID, inventoryID)
	if err != nil {
		return inventory.Inventory{}, err
	}
	if !found {
		return inventory.Inventory{}, ErrNotFound
	}

	if err := a.authorizer.CheckInventory(ctx, principal, permission, inventoryID); err != nil {
		a.recordAuthorizationDenied(ctx, principal, tenantID)
		return inventory.Inventory{}, err
	}
	return item, nil
}
