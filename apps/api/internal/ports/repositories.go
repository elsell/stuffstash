package ports

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

var ErrAuthorizationOutboxClaimLost = errors.New("authorization outbox claim lost")

type TenantRepository interface {
	SaveTenant(ctx context.Context, tenant tenant.Tenant) error
	TenantExists(ctx context.Context, tenantID tenant.ID) (bool, error)
}

type InventoryRepository interface {
	SaveInventory(ctx context.Context, inventory inventory.Inventory) error
	InventoryByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (inventory.Inventory, bool, error)
	ListInventoriesByTenant(ctx context.Context, tenantID inventory.TenantID, page InventoryListPageRequest) ([]inventory.Inventory, error)
}

type InventoryListPageRequest struct {
	AfterInventoryID inventory.InventoryID
	Limit            int
}

type AssetRepository interface {
	CreateAsset(ctx context.Context, asset asset.Asset) error
	AssetByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (asset.Asset, bool, error)
	ListAssetsByInventory(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page AssetListPageRequest) ([]asset.Asset, error)
}

type AssetListPageRequest struct {
	AfterAssetID asset.ID
	Limit        int
}

type AuthorizationOutboxEventKind string

const (
	AuthorizationOutboxGrantTenantOwner    AuthorizationOutboxEventKind = "grant_tenant_owner"
	AuthorizationOutboxGrantInventoryOwner AuthorizationOutboxEventKind = "grant_inventory_owner"
)

type AuthorizationOutboxEvent struct {
	ID               string
	Kind             AuthorizationOutboxEventKind
	PrincipalID      identity.PrincipalID
	TenantID         tenant.ID
	InventoryID      inventory.InventoryID
	Attempts         int
	LastError        string
	ClaimID          string
	ClaimedUntil     time.Time
	DeadLetteredAt   time.Time
	DeadLetterReason string
	CreatedAt        time.Time
}

type AuthorizationOutbox interface {
	SaveTenantAndEnqueueOwnerGrant(ctx context.Context, eventID string, tenant tenant.Tenant, principal identity.Principal) error
	SaveInventoryAndEnqueueOwnerGrant(ctx context.Context, eventID string, inventory inventory.Inventory, tenantID tenant.ID, principal identity.Principal) error
	ClaimPendingAuthorizationOutboxEvents(ctx context.Context, claimID string, limit int, leaseUntil time.Time) ([]AuthorizationOutboxEvent, error)
	MarkAuthorizationOutboxEventProcessed(ctx context.Context, eventID string, claimID string) error
	MarkAuthorizationOutboxEventFailed(ctx context.Context, eventID string, claimID string, reason string) error
	MarkAuthorizationOutboxEventDeadLettered(ctx context.Context, eventID string, claimID string, reason string) error
}
