package app

import (
	"context"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type CreateTenantInput struct {
	Principal identity.Principal
	Name      string
}

type CreateInventoryInput struct {
	Principal identity.Principal
	TenantID  tenant.ID
	Name      string
}

type ListInventoriesInput struct {
	Principal identity.Principal
	TenantID  tenant.ID
}

func (a App) CurrentPrincipal(principal identity.Principal) identity.Principal {
	return principal
}

func (a App) CreateTenant(ctx context.Context, input CreateTenantInput) (tenant.Tenant, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return tenant.Tenant{}, ErrInvalidInput
	}

	id := a.ids.NewID()

	tenantName, ok := tenant.NewName(name)
	if !ok {
		return tenant.Tenant{}, ErrInvalidInput
	}
	tenantID, ok := tenant.NewID(id)
	if !ok {
		return tenant.Tenant{}, ErrInvalidInput
	}

	item := tenant.Tenant{
		ID:   tenantID,
		Name: tenantName,
	}

	if err := a.tenants.SaveTenant(ctx, item); err != nil {
		return tenant.Tenant{}, err
	}
	if err := a.authorizer.GrantTenantOwner(ctx, input.Principal, item.ID); err != nil {
		return tenant.Tenant{}, err
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventTenantCreated,
		Message: "tenant created",
		Fields: map[string]string{
			"tenant_id":    item.ID.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})

	return item, nil
}

func (a App) CreateInventory(ctx context.Context, input CreateInventoryInput) (inventory.Inventory, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return inventory.Inventory{}, ErrInvalidInput
	}

	exists, err := a.tenants.TenantExists(ctx, input.TenantID)
	if err != nil {
		return inventory.Inventory{}, err
	}
	if !exists {
		return inventory.Inventory{}, ErrNotFound
	}

	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionCreateInventory, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return inventory.Inventory{}, err
	}

	id := a.ids.NewID()
	inventoryID, ok := inventory.NewID(id)
	if !ok {
		return inventory.Inventory{}, ErrInvalidInput
	}
	inventoryName, ok := inventory.NewName(name)
	if !ok {
		return inventory.Inventory{}, ErrInvalidInput
	}

	item := inventory.Inventory{
		ID:       inventoryID,
		TenantID: inventory.TenantID(input.TenantID.String()),
		Name:     inventoryName,
	}

	if err := a.inventories.SaveInventory(ctx, item); err != nil {
		return inventory.Inventory{}, err
	}
	if err := a.authorizer.GrantInventoryOwner(ctx, input.Principal, input.TenantID, item.ID); err != nil {
		return inventory.Inventory{}, err
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventInventoryCreated,
		Message: "inventory created",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": item.ID.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})

	return item, nil
}

func (a App) ListInventories(ctx context.Context, input ListInventoriesInput) ([]inventory.Inventory, error) {
	exists, err := a.tenants.TenantExists(ctx, input.TenantID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}

	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionView, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return nil, err
	}

	items, err := a.inventories.ListInventoriesByTenant(ctx, inventory.TenantID(input.TenantID.String()))
	if err != nil {
		return nil, err
	}

	visible := make([]inventory.Inventory, 0, len(items))
	for _, item := range items {
		if err := a.authorizer.CheckInventory(ctx, input.Principal, ports.InventoryPermissionView, item.ID); err == nil {
			visible = append(visible, item)
		}
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventInventoriesListed,
		Message: "inventories listed",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})

	return visible, nil
}

func (a App) recordAuthorizationDenied(ctx context.Context, principal identity.Principal, tenantID tenant.ID) {
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAuthorizationDenied,
		Message: "authorization denied",
		Fields: map[string]string{
			"tenant_id":    tenantID.String(),
			"principal_id": principal.ID.String(),
		},
	})
}
