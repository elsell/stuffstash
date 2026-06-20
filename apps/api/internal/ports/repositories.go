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
	TenantByID(ctx context.Context, tenantID tenant.ID) (tenant.Tenant, bool, error)
	TenantExists(ctx context.Context, tenantID tenant.ID) (bool, error)
}

type TenantUnitOfWork interface {
	SaveTenant(ctx context.Context, tenant tenant.Tenant) error
	UpdateTenant(ctx context.Context, tenant tenant.Tenant, auditRecord audit.Record) error
	UpdateTenantLifecycle(ctx context.Context, tenant tenant.Tenant, auditRecord audit.Record) error
	DeleteTenant(ctx context.Context, tenantID tenant.ID, auditRecord audit.Record) error
}

type InventoryRepository interface {
	InventoryByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (inventory.Inventory, bool, error)
	InventoryHasActiveAssets(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (bool, error)
	ListInventoriesByTenant(ctx context.Context, tenantID inventory.TenantID, page InventoryListPageRequest) ([]inventory.Inventory, error)
}

type InventoryUnitOfWork interface {
	SaveInventory(ctx context.Context, inventory inventory.Inventory) error
	UpdateInventory(ctx context.Context, inventory inventory.Inventory, auditRecord audit.Record) error
	UpdateInventoryLifecycle(ctx context.Context, inventory inventory.Inventory, auditRecord audit.Record) error
	DeleteInventory(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, auditRecord audit.Record) error
}

type InventoryAccessRepository interface {
	InventoryAccessGrantByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, principalID identity.PrincipalID, relationship InventoryAccessRelationship) (InventoryAccessGrant, bool, error)
	ListInventoryAccessGrants(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page InventoryAccessGrantPageRequest) ([]InventoryAccessGrant, error)
	InventoryAccessInvitationByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string) (InventoryAccessInvitation, bool, error)
	ListInventoryAccessInvitations(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page InventoryAccessInvitationPageRequest) ([]InventoryAccessInvitation, error)
}

type InventoryAccessUnitOfWork interface {
	SaveInventoryAccessGrantAndEnqueue(ctx context.Context, eventID string, grant InventoryAccessGrant, auditRecord audit.Record) error
	DeleteInventoryAccessGrantAndClaimRevoke(ctx context.Context, eventID string, claimID string, leaseUntil time.Time, grant InventoryAccessGrant, auditRecord audit.Record) (AuthorizationOutboxEvent, bool, error)
	SaveInventoryAccessInvitation(ctx context.Context, invitation InventoryAccessInvitation, auditRecord audit.Record) (InventoryAccessInvitation, error)
	AcceptInventoryAccessInvitationAndEnqueue(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, tokenHash string, acceptor identity.Principal, eventID string, auditRecord audit.Record) (InventoryAccessInvitation, InventoryAccessGrant, error)
	RevokeInventoryAccessInvitation(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, auditRecord audit.Record) (bool, error)
	CancelInventoryAccessInvitation(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, auditRecord audit.Record) (bool, error)
	UpdateInventoryAccessInvitationExpiration(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, expiresAt time.Time, auditRecord audit.Record) (InventoryAccessInvitation, bool, error)
	DeleteInventoryAccessInvitation(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, auditRecord audit.Record) (bool, error)
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

type InventoryAccessInvitationStatus string

const (
	InventoryAccessInvitationPending   InventoryAccessInvitationStatus = "pending"
	InventoryAccessInvitationAccepted  InventoryAccessInvitationStatus = "accepted"
	InventoryAccessInvitationRevoked   InventoryAccessInvitationStatus = "revoked"
	InventoryAccessInvitationCancelled InventoryAccessInvitationStatus = "cancelled"
)

type InventoryAccessInvitationStatusFilter string

const (
	InventoryAccessInvitationStatusFilterAll       InventoryAccessInvitationStatusFilter = "all"
	InventoryAccessInvitationStatusFilterPending   InventoryAccessInvitationStatusFilter = "pending"
	InventoryAccessInvitationStatusFilterAccepted  InventoryAccessInvitationStatusFilter = "accepted"
	InventoryAccessInvitationStatusFilterRevoked   InventoryAccessInvitationStatusFilter = "revoked"
	InventoryAccessInvitationStatusFilterCancelled InventoryAccessInvitationStatusFilter = "cancelled"
	InventoryAccessInvitationStatusFilterExpired   InventoryAccessInvitationStatusFilter = "expired"
)

type InventoryAccessInvitation struct {
	ID                  string
	TenantID            tenant.ID
	InventoryID         inventory.InventoryID
	Email               identity.Email
	TokenHash           string
	Relationship        InventoryAccessRelationship
	Status              InventoryAccessInvitationStatus
	InviterPrincipalID  identity.PrincipalID
	AcceptedPrincipalID identity.PrincipalID
	CreatedAt           time.Time
	ExpiresAt           time.Time
	AcceptedAt          time.Time
	RevokedAt           time.Time
}

func (i InventoryAccessInvitation) IsExpired(now time.Time) bool {
	return i.Status == InventoryAccessInvitationPending && !i.ExpiresAt.IsZero() && !i.ExpiresAt.After(now)
}

func (i InventoryAccessInvitation) CursorKey() string {
	return i.ID
}

type InventoryAccessInvitationPageRequest struct {
	AfterInvitationID string
	Limit             int
	StatusFilter      InventoryAccessInvitationStatusFilter
	Now               time.Time
}

type CustomFieldDefinitionRepository interface {
	CustomFieldDefinitionHasActiveAssetValues(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, definition customfield.Definition) (bool, error)
	CustomFieldDefinitionByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, definitionID customfield.ID) (customfield.Definition, bool, error)
	ListTenantCustomFieldDefinitions(ctx context.Context, tenantID tenant.ID, page CustomFieldDefinitionPageRequest) ([]customfield.Definition, error)
	ListInventoryCustomFieldDefinitions(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page CustomFieldDefinitionPageRequest) ([]customfield.Definition, error)
	ListEffectiveCustomFieldDefinitions(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) ([]customfield.Definition, error)
}

type CustomFieldDefinitionUnitOfWork interface {
	SaveCustomFieldDefinition(ctx context.Context, definition customfield.Definition, auditRecord audit.Record) error
	UpdateCustomFieldDefinition(ctx context.Context, definition customfield.Definition, auditRecord audit.Record) error
	UpdateCustomFieldDefinitionLifecycle(ctx context.Context, definition customfield.Definition, auditRecord audit.Record) error
	DeleteCustomFieldDefinition(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, definitionID customfield.ID, auditRecord audit.Record) error
}

type CustomFieldDefinitionPageRequest struct {
	AfterDefinitionKey string
	Limit              int
}

type CustomAssetTypeRepository interface {
	CustomAssetTypeHasActiveReferences(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetTypeID customfield.AssetTypeID) (bool, error)
	CustomAssetTypeByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetTypeID customfield.AssetTypeID) (customfield.AssetType, bool, error)
	ListTenantCustomAssetTypes(ctx context.Context, tenantID tenant.ID, page CustomAssetTypePageRequest) ([]customfield.AssetType, error)
	ListInventoryCustomAssetTypes(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page CustomAssetTypePageRequest) ([]customfield.AssetType, error)
	CustomAssetTypesByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, ids []customfield.AssetTypeID) ([]customfield.AssetType, error)
}

type CustomAssetTypeUnitOfWork interface {
	SaveCustomAssetType(ctx context.Context, assetType customfield.AssetType, auditRecord audit.Record) error
	UpdateCustomAssetType(ctx context.Context, assetType customfield.AssetType, auditRecord audit.Record) error
	ArchiveCustomAssetType(ctx context.Context, assetType customfield.AssetType, auditRecord audit.Record) error
	RestoreCustomAssetType(ctx context.Context, assetType customfield.AssetType, auditRecord audit.Record) error
	DeleteCustomAssetType(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetTypeID customfield.AssetTypeID, auditRecord audit.Record) error
}

type CustomAssetTypePageRequest struct {
	AfterAssetTypeKey string
	Limit             int
}

type AssetRepository interface {
	AssetByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (asset.Asset, bool, error)
	AssetHasActiveChildren(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (bool, error)
	ListAssetsByInventory(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page AssetListPageRequest) ([]asset.Asset, error)
}

type AssetUnitOfWork interface {
	CreateAsset(ctx context.Context, asset asset.Asset, auditRecord audit.Record, undoableOperation *UndoableOperation) error
	UpdateAsset(ctx context.Context, asset asset.Asset, auditRecords []audit.Record, undoableOperation *UndoableOperation) error
	UpdateAssetLifecycle(ctx context.Context, asset asset.Asset, auditRecord audit.Record, undoableOperation *UndoableOperation) error
	DeleteAsset(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, auditRecord audit.Record) error
}

type UndoableOperationRepository interface {
	UndoableOperationByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, operationID string) (UndoableOperation, bool, error)
	ApplyAssetUndoableOperation(ctx context.Context, operationID string, direction UndoableOperationDirection, expectedCurrent asset.Asset, resulting asset.Asset, auditRecord audit.Record) (UndoableOperation, asset.Asset, error)
}

type UndoableOperationStatus string

const (
	UndoableOperationAvailable UndoableOperationStatus = "available"
	UndoableOperationUndone    UndoableOperationStatus = "undone"
	UndoableOperationRedone    UndoableOperationStatus = "redone"
)

type UndoableOperationDirection string

const (
	UndoableOperationDirectionUndo UndoableOperationDirection = "undo"
	UndoableOperationDirectionRedo UndoableOperationDirection = "redo"
)

type UndoableOperation struct {
	ID                string
	TenantID          tenant.ID
	InventoryID       inventory.InventoryID
	PrincipalID       identity.PrincipalID
	Source            audit.Source
	TargetType        audit.TargetType
	TargetID          string
	OriginalAction    audit.Action
	Status            UndoableOperationStatus
	CreatedAt         time.Time
	LastAppliedAt     time.Time
	BeforeAsset       *asset.Asset
	AfterAsset        asset.Asset
	UndoAuditRecordID audit.ID
	RedoAuditRecordID audit.ID
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
	AttachmentByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, attachmentID media.ID) (media.Attachment, bool, error)
	ListAttachmentsByAsset(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, page AttachmentListPageRequest) ([]media.Attachment, error)
}

type AttachmentUnitOfWork interface {
	SaveAttachment(ctx context.Context, attachment media.Attachment, auditRecord audit.Record) error
	UpdateAttachmentLifecycle(ctx context.Context, attachment media.Attachment, auditRecord audit.Record) error
	DeleteAttachmentAndEnqueueBlobDeletion(ctx context.Context, eventID string, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, attachmentID media.ID, auditRecord audit.Record) (media.Attachment, bool, error)
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

type BlobDeletionEvent struct {
	ID               string
	StorageKey       media.StorageKey
	Attempts         int
	LastError        string
	ClaimID          string
	ClaimedUntil     time.Time
	ProcessedAt      time.Time
	DeadLetteredAt   time.Time
	DeadLetterReason string
	CreatedAt        time.Time
}

type BlobDeletionOutbox interface {
	ClaimPendingBlobDeletionEvents(ctx context.Context, claimID string, limit int, leaseUntil time.Time) ([]BlobDeletionEvent, error)
	MarkBlobDeletionEventProcessed(ctx context.Context, eventID string, claimID string) error
	MarkBlobDeletionEventFailed(ctx context.Context, eventID string, claimID string, reason string) error
	MarkBlobDeletionEventDeadLettered(ctx context.Context, eventID string, claimID string, reason string) error
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
