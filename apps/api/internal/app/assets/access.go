package assets

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s Service) ensureActiveInventoryAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID, permission ports.InventoryPermission) error {
	if err := s.ensureInventoryAccessDependencies(); err != nil {
		return err
	}
	item, err := s.ensureInventoryAccessItem(ctx, principal, tenantID, inventoryID, permission)
	if err != nil {
		return err
	}
	if !item.IsActive() {
		return apperrors.ErrNotFound
	}
	return nil
}

func (s Service) ensureInventoryAccessItem(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID, permission ports.InventoryPermission) (inventory.Inventory, error) {
	if err := s.ensureInventoryAccessDependencies(); err != nil {
		return inventory.Inventory{}, err
	}
	exists, err := s.tenants.TenantExists(ctx, tenantID)
	if err != nil {
		return inventory.Inventory{}, err
	}
	if !exists {
		return inventory.Inventory{}, apperrors.ErrNotFound
	}

	item, found, err := s.inventories.InventoryByID(ctx, tenantID, inventoryID)
	if err != nil {
		return inventory.Inventory{}, err
	}
	if !found {
		return inventory.Inventory{}, apperrors.ErrNotFound
	}

	if err := s.authorizer.CheckInventory(ctx, principal, permission, inventoryID); err != nil {
		s.recordAuthorizationDenied(ctx, principal, tenantID)
		return inventory.Inventory{}, err
	}
	return item, nil
}

func (s Service) recordAuthorizationDenied(ctx context.Context, principal identity.Principal, tenantID tenant.ID) {
	s.observer.Record(ctx, ports.Event{
		Name:    ports.EventAuthorizationDenied,
		Message: "authorization denied",
		Fields: map[string]string{
			"tenant_id":    tenantID.String(),
			"principal_id": principal.ID.String(),
		},
	})
}
