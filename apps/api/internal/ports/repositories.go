package ports

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
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

type AuthorizationOutboxEventKind string

const (
	AuthorizationOutboxGrantTenantOwner    AuthorizationOutboxEventKind = "grant_tenant_owner"
	AuthorizationOutboxGrantInventoryOwner AuthorizationOutboxEventKind = "grant_inventory_owner"
)

type AuthorizationOutboxEvent struct {
	ID          string
	Kind        AuthorizationOutboxEventKind
	PrincipalID identity.PrincipalID
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Attempts    int
	LastError   string
	CreatedAt   time.Time
}

type AuthorizationOutbox interface {
	SaveTenantAndEnqueueOwnerGrant(ctx context.Context, eventID string, tenant tenant.Tenant, principal identity.Principal) error
	SaveInventoryAndEnqueueOwnerGrant(ctx context.Context, eventID string, inventory inventory.Inventory, tenantID tenant.ID, principal identity.Principal) error
	ListPendingAuthorizationOutboxEvents(ctx context.Context, limit int) ([]AuthorizationOutboxEvent, error)
	MarkAuthorizationOutboxEventProcessed(ctx context.Context, eventID string) error
	MarkAuthorizationOutboxEventFailed(ctx context.Context, eventID string, reason string) error
}
