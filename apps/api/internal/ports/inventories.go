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

// InventoryDeletionEffects supplies application-owned history for printing
// tombstones committed in the inventory deletion transaction.
type InventoryDeletionEffects struct {
	Audit func(audit.Action, audit.TargetType, string) (audit.Record, error)
}

type InventoryUnitOfWork interface {
	SaveInventory(ctx context.Context, inventory inventory.Inventory) error
	UpdateInventory(ctx context.Context, inventory inventory.Inventory, auditRecord audit.Record) error
	UpdateInventoryLifecycle(ctx context.Context, inventory inventory.Inventory, auditRecord audit.Record) error
	DeleteInventory(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, auditRecord audit.Record, effects *InventoryDeletionEffects) error
}

type InventoryListPageRequest struct {
	AfterInventoryID inventory.InventoryID
	Limit            int
}
