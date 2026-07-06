package ports

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type ImportSourceRequest struct {
	SourceType           importplan.SourceType
	BaseURL              string
	Username             string
	Password             string
	IncludeImages        bool
	FetchAttachmentBytes bool
	AllowInsecureTLS     bool
	AllowPrivateNetwork  bool
	MaxAttachmentBytes   int64
	FileName             string
	Content              []byte
}

type ImportSourceReader interface {
	ReadImportPlan(ctx context.Context, request ImportSourceRequest) (importplan.Plan, error)
}

type ImportJobPageRequest struct {
	Limit int
}

type ImportJobStatusPageRequest struct {
	Status importjob.Status
	Limit  int
}

type ImportJobCommand struct {
	Principal   identity.Principal
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	JobID       importjob.ID
}

type ImportJobRepository interface {
	SaveImportJob(ctx context.Context, job importjob.Record) error
	ImportJobByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID) (importjob.Record, bool, error)
	ListImportJobs(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ImportJobPageRequest) ([]importjob.Record, error)
	ListImportJobsByStatus(ctx context.Context, page ImportJobStatusPageRequest) ([]importjob.Record, error)
	UpdateImportJobIfStatus(ctx context.Context, job importjob.Record, expected importjob.Status) (bool, error)
	UpdateImportJobProgress(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID, progress importjob.Progress, expectedUpdatedAt time.Time) (bool, error)
	ClaimImportJob(ctx context.Context, job importjob.Record, expectedUpdatedAt time.Time) (bool, error)
	MarkImportJobHistoryRemoved(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID, removedAt time.Time, expectedUpdatedAt time.Time) (bool, error)
	UpdateImportJob(ctx context.Context, job importjob.Record) error
}

type ImportWorker interface {
	ExecuteImportJob(ctx context.Context, command ImportJobCommand) (importjob.Record, error)
	CancelImportJob(ctx context.Context, jobID importjob.ID, mode importjob.CancellationMode) error
}

type ImportJobSourceScope struct {
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	JobID       importjob.ID
}

type SealedImportJobSource struct {
	KeyID      string
	Algorithm  string
	Nonce      []byte
	Ciphertext []byte
}

type ImportJobSourceRecord struct {
	Scope     ImportJobSourceScope
	Sealed    SealedImportJobSource
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ImportJobSourceVault interface {
	StoreImportJobSource(ctx context.Context, scope ImportJobSourceScope, request ImportSourceRequest, expiresAt time.Time, now time.Time) error
	ImportJobSourceRequest(ctx context.Context, scope ImportJobSourceScope) (ImportSourceRequest, bool, error)
	DeleteImportJobSource(ctx context.Context, scope ImportJobSourceScope) (bool, error)
	VacuumImportJobSources(ctx context.Context, now time.Time) ([]ImportJobSourceScope, error)
}

type ImportJobSourceRepository interface {
	ReplaceImportJobSource(ctx context.Context, source ImportJobSourceRecord) error
	ImportJobSource(ctx context.Context, scope ImportJobSourceScope) (ImportJobSourceRecord, bool, error)
	DeleteImportJobSource(ctx context.Context, scope ImportJobSourceScope) (bool, error)
	DeleteExpiredImportJobSources(ctx context.Context, now time.Time) (int, error)
	DeleteVacuumableImportJobSources(ctx context.Context, terminalStatuses []importjob.Status, now time.Time) ([]ImportJobSourceScope, error)
}

type ImportSourceEntityType string

const (
	ImportSourceEntityAsset      ImportSourceEntityType = "asset"
	ImportSourceEntityAttachment ImportSourceEntityType = "attachment"
)

type ImportResourceType string

const (
	ImportResourceAsset      ImportResourceType = "asset"
	ImportResourceAttachment ImportResourceType = "attachment"
)

type ImportSourceLinkKey struct {
	TenantID          tenant.ID
	InventoryID       inventory.InventoryID
	SourceType        importplan.SourceType
	SourceInstanceKey string
	SourceEntityType  ImportSourceEntityType
	SourceEntityID    string
}

type ImportSourceLink struct {
	Key          ImportSourceLinkKey
	ResourceType ImportResourceType
	ResourceID   string
	JobID        importjob.ID
	CreatedAt    time.Time
}

type ImportJobResource struct {
	TenantID          tenant.ID
	InventoryID       inventory.InventoryID
	JobID             importjob.ID
	ResourceType      ImportResourceType
	ResourceID        string
	ResourceOwnerID   string
	SourceType        importplan.SourceType
	SourceInstanceKey string
	SourceEntityType  ImportSourceEntityType
	SourceEntityID    string
	CreatedAt         time.Time
}

type ImportJobResourcePageRequest struct {
	Limit int
}

type ImportLinkRepository interface {
	ImportSourceLinkByKey(ctx context.Context, key ImportSourceLinkKey) (ImportSourceLink, bool, error)
	SaveImportSourceLink(ctx context.Context, link ImportSourceLink) error
	SaveImportJobResource(ctx context.Context, record ImportJobResource) error
	ListImportJobResources(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID, page ImportJobResourcePageRequest) ([]ImportJobResource, error)
	ListAllImportJobResources(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID) ([]ImportJobResource, error)
	DeleteImportSourceLinksForJob(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID) (int, error)
}

type ImportAssetUnitOfWork interface {
	CreateImportedAsset(ctx context.Context, item asset.Asset, auditRecord audit.Record, undoableOperation *UndoableOperation, promotedParent *asset.Asset, parentAuditRecord *audit.Record, link ImportSourceLink, record ImportJobResource) error
}

type ImportAttachmentUnitOfWork interface {
	CreateImportedAttachment(ctx context.Context, attachment media.Attachment, auditRecord audit.Record, link ImportSourceLink, record ImportJobResource) error
}

type ImportSourceUserError struct {
	Detail string
}

func (e ImportSourceUserError) Error() string {
	return e.Detail
}

func NewImportSourceUserError(detail string) error {
	return ImportSourceUserError{Detail: detail}
}
