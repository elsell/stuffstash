package ports

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type TenantRepository interface {
	TenantByID(ctx context.Context, tenantID tenant.ID) (tenant.Tenant, bool, error)
	TenantExists(ctx context.Context, tenantID tenant.ID) (bool, error)
	ListTenants(ctx context.Context, page TenantListPageRequest) ([]tenant.Tenant, error)
}

type TenantListPageRequest struct {
	AfterTenantID tenant.ID
	Limit         int
}

type TenantUnitOfWork interface {
	SaveTenant(ctx context.Context, tenant tenant.Tenant) error
	UpdateTenant(ctx context.Context, tenant tenant.Tenant, auditRecord audit.Record) error
	UpdateTenantLifecycle(ctx context.Context, tenant tenant.Tenant, auditRecord audit.Record) error
	DeleteTenant(ctx context.Context, tenantID tenant.ID, auditRecord audit.Record) error
}

type AuthorizationOutboxEventKind string

const (
	AuthorizationOutboxGrantTenantOwner      AuthorizationOutboxEventKind = "grant_tenant_owner"
	AuthorizationOutboxGrantInventoryOwner   AuthorizationOutboxEventKind = "grant_inventory_owner"
	AuthorizationOutboxGrantInventoryViewer  AuthorizationOutboxEventKind = "grant_inventory_viewer"
	AuthorizationOutboxGrantInventoryEditor  AuthorizationOutboxEventKind = "grant_inventory_editor"
	AuthorizationOutboxRevokeInventoryViewer AuthorizationOutboxEventKind = "revoke_inventory_viewer"
	AuthorizationOutboxRevokeInventoryEditor AuthorizationOutboxEventKind = "revoke_inventory_editor"
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
	SaveTenantAndEnqueueOwnerGrant(ctx context.Context, eventID string, tenant tenant.Tenant, principal identity.Principal, auditRecord audit.Record) error
	SaveInventoryAndEnqueueOwnerGrant(ctx context.Context, eventID string, inventory inventory.Inventory, tenantID tenant.ID, principal identity.Principal, auditRecord audit.Record) error
	ListAuthorizationOutboxReplayEvents(ctx context.Context) ([]AuthorizationOutboxEvent, error)
	ClaimAuthorizationOutboxEvent(ctx context.Context, eventID string, claimID string, leaseUntil time.Time) (AuthorizationOutboxEvent, bool, error)
	ClaimPendingAuthorizationOutboxEvents(ctx context.Context, claimID string, limit int, now time.Time, leaseUntil time.Time) ([]AuthorizationOutboxEvent, error)
	MarkAuthorizationOutboxEventProcessed(ctx context.Context, eventID string, claimID string) error
	MarkAuthorizationOutboxEventFailed(ctx context.Context, eventID string, claimID string, reason string) error
	MarkAuthorizationOutboxEventDeadLettered(ctx context.Context, eventID string, claimID string, reason string) error
}
