package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

// InventoryReadAccess preserves scope checks for history and other read consumers.
type InventoryReadAccess interface {
	EnsureTenantExists(context.Context, tenant.ID) error
	EnsureInventoryAccess(context.Context, identity.Principal, tenant.ID, inventory.InventoryID, InventoryPermission) error
}
