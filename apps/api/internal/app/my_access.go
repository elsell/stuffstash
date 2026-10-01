package app

import (
	"context"

	inventoryapp "github.com/stuffstash/stuff-stash/internal/app/inventories"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type AccessRelationship = inventoryapp.AccessRelationship
type AccessSummary = inventoryapp.AccessSummary
type MyTenantAccess = inventoryapp.MyTenantAccess
type ListMyTenantsInput = inventoryapp.ListMyTenantsInput
type ListMyTenantsResult = inventoryapp.ListMyTenantsResult

const AccessRelationshipOwner = inventoryapp.AccessRelationshipOwner
const AccessRelationshipEditor = inventoryapp.AccessRelationshipEditor
const AccessRelationshipViewer = inventoryapp.AccessRelationshipViewer

func (a App) ListMyTenants(ctx context.Context, input ListMyTenantsInput) (ListMyTenantsResult, error) {
	return a.inventoryService().ListMyTenants(ctx, input)
}

func (a App) TenantAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID) (AccessSummary, error) {
	return a.inventoryService().TenantAccess(ctx, principal, tenantID)
}

func (a App) InventoryAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) (AccessSummary, error) {
	return a.inventoryService().InventoryAccess(ctx, principal, tenantID, inventoryID)
}
