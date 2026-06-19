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
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

var ErrAuthorizationOutboxClaimLost = errors.New("authorization outbox claim lost")
var ErrConflict = errors.New("conflict")
var ErrBlobNotFound = errors.New("blob not found")

type TenantRepository interface {
	SaveTenant(ctx context.Context, tenant tenant.Tenant) error
	TenantExists(ctx context.Context, tenantID tenant.ID) (bool, error)
}

type InventoryRepository interface {
	SaveInventory(ctx context.Context, inventory inventory.Inventory) error
	InventoryByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (inventory.Inventory, bool, error)
	ListInventoriesByTenant(ctx context.Context, tenantID inventory.TenantID, page InventoryListPageRequest) ([]inventory.Inventory, error)
	SaveInventoryAccessGrantAndEnqueue(ctx context.Context, eventID string, grant InventoryAccessGrant, auditRecord audit.Record) error
	DeleteInventoryAccessGrantAndClaimRevoke(ctx context.Context, eventID string, claimID string, leaseUntil time.Time, grant InventoryAccessGrant, auditRecord audit.Record) (AuthorizationOutboxEvent, bool, error)
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

func (r InventoryAccessRelationship) GrantOutboxKind() (AuthorizationOutboxEventKind, bool) {
	switch r {
	case InventoryAccessViewer:
		return AuthorizationOutboxGrantInventoryViewer, true
	case InventoryAccessEditor:
		return AuthorizationOutboxGrantInventoryEditor, true
	default:
		return "", false
	}
}

func (r InventoryAccessRelationship) RevokeOutboxKind() (AuthorizationOutboxEventKind, bool) {
	switch r {
	case InventoryAccessViewer:
		return AuthorizationOutboxRevokeInventoryViewer, true
	case InventoryAccessEditor:
		return AuthorizationOutboxRevokeInventoryEditor, true
	default:
		return "", false
	}
}

type InventoryAccessGrantPageRequest struct {
	AfterGrantKey string
	Limit         int
}

type CustomFieldDefinitionRepository interface {
	SaveCustomFieldDefinition(ctx context.Context, definition customfield.Definition, auditRecord audit.Record) error
	UpdateCustomFieldDefinition(ctx context.Context, definition customfield.Definition, auditRecord audit.Record) error
	CustomFieldDefinitionByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, definitionID customfield.ID) (customfield.Definition, bool, error)
	ListTenantCustomFieldDefinitions(ctx context.Context, tenantID tenant.ID, page CustomFieldDefinitionPageRequest) ([]customfield.Definition, error)
	ListInventoryCustomFieldDefinitions(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page CustomFieldDefinitionPageRequest) ([]customfield.Definition, error)
	ListEffectiveCustomFieldDefinitions(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) ([]customfield.Definition, error)
}

type CustomFieldDefinitionPageRequest struct {
	AfterDefinitionKey string
	Limit              int
}

type CustomAssetTypeRepository interface {
	SaveCustomAssetType(ctx context.Context, assetType customfield.AssetType, auditRecord audit.Record) error
	UpdateCustomAssetType(ctx context.Context, assetType customfield.AssetType, auditRecord audit.Record) error
	CustomAssetTypeByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetTypeID customfield.AssetTypeID) (customfield.AssetType, bool, error)
	ListTenantCustomAssetTypes(ctx context.Context, tenantID tenant.ID, page CustomAssetTypePageRequest) ([]customfield.AssetType, error)
	ListInventoryCustomAssetTypes(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page CustomAssetTypePageRequest) ([]customfield.AssetType, error)
	CustomAssetTypesByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, ids []customfield.AssetTypeID) ([]customfield.AssetType, error)
}

type CustomAssetTypePageRequest struct {
	AfterAssetTypeKey string
	Limit             int
}

type AssetRepository interface {
	CreateAsset(ctx context.Context, asset asset.Asset, auditRecord audit.Record) error
	UpdateAsset(ctx context.Context, asset asset.Asset, auditRecords []audit.Record) error
	UpdateAssetLifecycle(ctx context.Context, asset asset.Asset, auditRecord audit.Record) error
	AssetByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (asset.Asset, bool, error)
	AssetHasActiveChildren(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (bool, error)
	ListAssetsByInventory(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page AssetListPageRequest) ([]asset.Asset, error)
}

type AssetLifecycleFilter string

const (
	AssetLifecycleFilterActive   AssetLifecycleFilter = "active"
	AssetLifecycleFilterArchived AssetLifecycleFilter = "archived"
	AssetLifecycleFilterAll      AssetLifecycleFilter = "all"
)

type AssetListPageRequest struct {
	AfterAssetID    asset.ID
	Limit           int
	LifecycleFilter AssetLifecycleFilter
}

type AssetSearchRepository interface {
	SearchAssets(ctx context.Context, tenantID tenant.ID, inventoryIDs []inventory.InventoryID, page AssetSearchPageRequest) ([]AssetSearchResult, error)
}

type AssetSearchPageRequest struct {
	Query             search.Query
	Mode              search.Mode
	CustomAssetTypeID asset.CustomAssetTypeID
	AfterResultKey    string
	Limit             int
	LifecycleFilter   AssetLifecycleFilter
}

type AssetSearchResult struct {
	Type      search.ResultType
	TenantID  tenant.ID
	Inventory inventory.Inventory
	Asset     asset.Asset
	Matches   []search.Match
}

func (r AssetSearchResult) CursorKey() string {
	return r.Inventory.ID.String() + ":" + r.Asset.ID.String()
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

type AttachmentRepository interface {
	SaveAttachment(ctx context.Context, attachment media.Attachment, auditRecord audit.Record) error
	AttachmentByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, attachmentID media.ID) (media.Attachment, bool, error)
	ListAttachmentsByAsset(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, page AttachmentListPageRequest) ([]media.Attachment, error)
}

type AttachmentListPageRequest struct {
	AfterAttachmentID media.ID
	Limit             int
}

type BlobStorage interface {
	PutBlob(ctx context.Context, key media.StorageKey, contentType media.ContentType, data []byte) error
	GetBlob(ctx context.Context, key media.StorageKey) ([]byte, error)
	DeleteBlob(ctx context.Context, key media.StorageKey) error
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
	ClaimAuthorizationOutboxEvent(ctx context.Context, eventID string, claimID string, leaseUntil time.Time) (AuthorizationOutboxEvent, bool, error)
	ClaimPendingAuthorizationOutboxEvents(ctx context.Context, claimID string, limit int, leaseUntil time.Time) ([]AuthorizationOutboxEvent, error)
	MarkAuthorizationOutboxEventProcessed(ctx context.Context, eventID string, claimID string) error
	MarkAuthorizationOutboxEventFailed(ctx context.Context, eventID string, claimID string, reason string) error
	MarkAuthorizationOutboxEventDeadLettered(ctx context.Context, eventID string, claimID string, reason string) error
}
