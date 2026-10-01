package ports

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type ActiveInventoryAccess interface {
	EnsureActiveInventoryAccess(context.Context, identity.Principal, tenant.ID, inventory.InventoryID, InventoryPermission) error
}
