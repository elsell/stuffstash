package ports

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type TenantRepository interface {
	SaveTenant(ctx context.Context, tenant tenant.Tenant) error
	TenantExists(ctx context.Context, tenantID tenant.ID) (bool, error)
}

type InventoryRepository interface {
	SaveInventory(ctx context.Context, inventory inventory.Inventory) error
	ListInventoriesByTenant(ctx context.Context, tenantID inventory.TenantID) ([]inventory.Inventory, error)
}
