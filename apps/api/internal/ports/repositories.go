package ports

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

var ErrAuthorizationOutboxClaimLost = errors.New("authorization outbox claim lost")
var ErrConflict = errors.New("conflict")

type TenantRepository interface {
	SaveTenant(ctx context.Context, tenant tenant.Tenant) error
	TenantExists(ctx context.Context, tenantID tenant.ID) (bool, error)
}

type InventoryRepository interface {
	SaveInventory(ctx context.Context, inventory inventory.Inventory) error
	InventoryByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (inventory.Inventory, bool, error)
	ListInventoriesByTenant(ctx context.Context, tenantID inventory.TenantID, page InventoryListPageRequest) ([]inventory.Inventory, error)
	SaveInventoryAccessGrantAndEnqueue(ctx context.Context, eventID string, grant InventoryAccessGrant, auditRecord audit.Record) error
	ListInventoryAccessGrants(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page InventoryAccessGrantPageRequest) ([]InventoryAccessGrant, error)
}

type InventoryListPageRequest struct {
	AfterInventoryID inventory.InventoryID
	Limit            int
}

type InventoryAccessGrant struct {
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	PrincipalID  identity.PrincipalID
	Relationship InventoryAccessRelationship
}

func (g InventoryAccessGrant) CursorKey() string {
	return g.PrincipalID.String() + ":" + string(g.Relationship)
}

type InventoryAccessGrantPageRequest struct {
	AfterGrantKey string
	Limit         int
}

type CustomFieldDefinitionRepository interface {
	SaveCustomFieldDefinition(ctx context.Context, definition customfield.Definition, auditRecord audit.Record) error
	ListTenantCustomFieldDefinitions(ctx context.Context, tenantID tenant.ID, page CustomFieldDefinitionPageRequest) ([]customfield.Definition, error)
	ListInventoryCustomFieldDefinitions(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page CustomFieldDefinitionPageRequest) ([]customfield.Definition, error)
	ListEffectiveCustomFieldDefinitions(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) ([]customfield.Definition, error)
}

type CustomFieldDefinitionPageRequest struct {
	AfterDefinitionKey string
	Limit              int
}

type AssetRepository interface {
	CreateAsset(ctx context.Context, asset asset.Asset, auditRecord audit.Record) error
	UpdateAsset(ctx context.Context, asset asset.Asset, auditRecords []audit.Record) error
	AssetByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (asset.Asset, bool, error)
	ListAssetsByInventory(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page AssetListPageRequest) ([]asset.Asset, error)
}

type AssetListPageRequest struct {
	AfterAssetID asset.ID
	Limit        int
}

type AuditRepository interface {
	SaveAuditRecord(ctx context.Context, record audit.Record) error
	ListTenantAuditRecords(ctx context.Context, tenantID tenant.ID, page AuditRecordPageRequest) ([]audit.Record, error)
	ListInventoryAuditRecords(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page AuditRecordPageRequest) ([]audit.Record, error)
}

type AuditRecordPageRequest struct {
	AfterOccurredAt time.Time
	AfterRecordID   audit.ID
	Limit           int
}

type AuthorizationOutboxEventKind string

const (
	AuthorizationOutboxGrantTenantOwner     AuthorizationOutboxEventKind = "grant_tenant_owner"
	AuthorizationOutboxGrantInventoryOwner  AuthorizationOutboxEventKind = "grant_inventory_owner"
	AuthorizationOutboxGrantInventoryViewer AuthorizationOutboxEventKind = "grant_inventory_viewer"
	AuthorizationOutboxGrantInventoryEditor AuthorizationOutboxEventKind = "grant_inventory_editor"
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
	ClaimPendingAuthorizationOutboxEvents(ctx context.Context, claimID string, limit int, leaseUntil time.Time) ([]AuthorizationOutboxEvent, error)
	MarkAuthorizationOutboxEventProcessed(ctx context.Context, eventID string, claimID string) error
	MarkAuthorizationOutboxEventFailed(ctx context.Context, eventID string, claimID string, reason string) error
	MarkAuthorizationOutboxEventDeadLettered(ctx context.Context, eventID string, claimID string, reason string) error
}
