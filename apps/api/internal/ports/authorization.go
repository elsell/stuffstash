package ports

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

var ErrForbidden = errors.New("forbidden")

type TenantPermission string

const (
	TenantPermissionView            TenantPermission = "view"
	TenantPermissionCreateInventory TenantPermission = "create_inventory"
)

type InventoryPermission string

const (
	InventoryPermissionView InventoryPermission = "view"
)

type Authorizer interface {
	CheckTenant(ctx context.Context, principal identity.Principal, permission TenantPermission, tenantID tenant.ID) error
	CheckInventory(ctx context.Context, principal identity.Principal, permission InventoryPermission, inventoryID inventory.InventoryID) error
	GrantTenantOwner(ctx context.Context, principal identity.Principal, tenantID tenant.ID) error
	GrantInventoryOwner(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error
}
