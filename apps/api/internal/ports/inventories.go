package ports

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type InventoryRepository interface {
	InventoryByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (inventory.Inventory, bool, error)
	InventoryHasActiveAssets(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (bool, error)
	ListInventoriesByTenant(ctx context.Context, tenantID inventory.TenantID, page InventoryListPageRequest) ([]inventory.Inventory, error)
}

type InventoryUnitOfWork interface {
	SaveInventory(ctx context.Context, inventory inventory.Inventory) error
	UpdateInventory(ctx context.Context, inventory inventory.Inventory, auditRecord audit.Record) error
	UpdateInventoryLifecycle(ctx context.Context, inventory inventory.Inventory, auditRecord audit.Record) error
	DeleteInventory(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, auditRecord audit.Record) error
}

type InventoryListPageRequest struct {
	AfterInventoryID inventory.InventoryID
	Limit            int
}
