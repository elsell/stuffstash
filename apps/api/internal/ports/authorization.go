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
	TenantPermissionConfigure       TenantPermission = "configure"
)

type InventoryPermission string

const (
	InventoryPermissionView        InventoryPermission = "view"
	InventoryPermissionCreateAsset InventoryPermission = "create_asset"
	InventoryPermissionEditAsset   InventoryPermission = "edit_asset"
	InventoryPermissionShare       InventoryPermission = "share"
	InventoryPermissionConfigure   InventoryPermission = "configure"
)

type InventoryAccessRelationship string

const (
	InventoryAccessViewer InventoryAccessRelationship = "viewer"
	InventoryAccessEditor InventoryAccessRelationship = "editor"
)

type Authorizer interface {
	CheckTenant(ctx context.Context, principal identity.Principal, permission TenantPermission, tenantID tenant.ID) error
	CheckInventory(ctx context.Context, principal identity.Principal, permission InventoryPermission, inventoryID inventory.InventoryID) error
	// ListViewableInventoryIDs receives inventory candidates that the application has already scoped to tenantID.
	// Adapters must intersect authorization results with candidates and ignore resources outside that set.
	ListViewableInventoryIDs(ctx context.Context, principal identity.Principal, tenantID tenant.ID, candidates []inventory.InventoryID) ([]inventory.InventoryID, error)
	// Grant methods must be idempotent because authorization outbox retries may replay them.
	GrantTenantOwner(ctx context.Context, principal identity.Principal, tenantID tenant.ID) error
	GrantInventoryOwner(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error
	GrantInventoryViewer(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error
	GrantInventoryEditor(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error
	RevokeInventoryViewer(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error
	RevokeInventoryEditor(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error
}
