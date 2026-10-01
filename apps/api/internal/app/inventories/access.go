package inventories

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a Service) EnsureActiveInventoryAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID, permission ports.InventoryPermission) error {
	item, err := a.EnsureInventoryAccessItem(ctx, principal, tenantID, inventoryID, permission)
	if err != nil {
		return err
	}
	if !item.IsActive() {
		return apperrors.ErrNotFound
	}
	return nil
}

func (a Service) EnsureInventoryAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID, permission ports.InventoryPermission) error {
	_, err := a.EnsureInventoryAccessItem(ctx, principal, tenantID, inventoryID, permission)
	return err
}

func (a Service) EnsureTenantExists(ctx context.Context, tenantID tenant.ID) error {
	exists, err := a.tenants.TenantExists(ctx, tenantID)
	if err != nil {
		return err
	}
	if !exists {
		return apperrors.ErrNotFound
	}
	return nil
}

func (a Service) EnsureInventoryAccessItem(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID, permission ports.InventoryPermission) (inventory.Inventory, error) {
	exists, err := a.tenants.TenantExists(ctx, tenantID)
	if err != nil {
		return inventory.Inventory{}, err
	}
	if !exists {
		return inventory.Inventory{}, apperrors.ErrNotFound
	}

	item, found, err := a.inventories.InventoryByID(ctx, tenantID, inventoryID)
	if err != nil {
		return inventory.Inventory{}, err
	}
	if !found {
		return inventory.Inventory{}, apperrors.ErrNotFound
	}

	if err := a.authorizer.CheckInventory(ctx, principal, permission, inventoryID); err != nil {
		a.RecordAuthorizationDenied(ctx, principal, tenantID)
		return inventory.Inventory{}, err
	}
	return item, nil
}

func (a Service) RecordAuthorizationDenied(ctx context.Context, principal identity.Principal, tenantID tenant.ID) {
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAuthorizationDenied,
		Message: "authorization denied",
		Fields: map[string]string{
			"tenant_id":    tenantID.String(),
			"principal_id": principal.ID.String(),
		},
	})
}
