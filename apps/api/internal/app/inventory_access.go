package app

import (
	"context"

	inventoryapp "github.com/stuffstash/stuff-stash/internal/app/inventories"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type GrantInventoryAccessInput = inventoryapp.GrantInventoryAccessInput
type ListInventoryAccessGrantsInput = inventoryapp.ListInventoryAccessGrantsInput
type GetInventoryAccessGrantInput = inventoryapp.GetInventoryAccessGrantInput
type ListInventoryAccessGrantsResult = inventoryapp.ListInventoryAccessGrantsResult
type RevokeInventoryAccessInput = inventoryapp.RevokeInventoryAccessInput

func (a App) GrantInventoryAccess(ctx context.Context, input GrantInventoryAccessInput) (ports.InventoryAccessGrant, error) {
	return a.inventoryService().GrantInventoryAccess(ctx, input)
}

func (a App) RevokeInventoryAccess(ctx context.Context, input RevokeInventoryAccessInput) (bool, error) {
	return a.inventoryService().RevokeInventoryAccess(ctx, input)
}

func (a App) ListInventoryAccessGrants(ctx context.Context, input ListInventoryAccessGrantsInput) (ListInventoryAccessGrantsResult, error) {
	return a.inventoryService().ListInventoryAccessGrants(ctx, input)
}

func (a App) GetInventoryAccessGrant(ctx context.Context, input GetInventoryAccessGrantInput) (ports.InventoryAccessGrant, error) {
	return a.inventoryService().GetInventoryAccessGrant(ctx, input)
}
