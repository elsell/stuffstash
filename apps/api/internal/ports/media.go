package ports

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type AttachmentRepository interface {
	AttachmentByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, attachmentID media.ID) (media.Attachment, bool, error)
	ListAttachmentsByAsset(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, page AttachmentListPageRequest) ([]media.Attachment, error)
	FirstImageAttachmentsByAssets(ctx context.Context, tenantID tenant.ID, assets []AttachmentAssetReference) (map[AttachmentAssetReference]media.Attachment, error)
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

type AttachmentAssetReference struct {
	InventoryID inventory.InventoryID
	AssetID     asset.ID
}

type BlobStorage interface {
	PutBlob(ctx context.Context, key media.StorageKey, contentType media.ContentType, data []byte) error
	GetBlob(ctx context.Context, key media.StorageKey) ([]byte, error)
	DeleteBlob(ctx context.Context, key media.StorageKey) error
}

type DirectAttachmentUploadRequest struct {
	UploadID     string
	AttachmentID media.ID
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	AssetID      asset.ID
	StorageKey   media.StorageKey
	FileName     media.FileName
	ContentType  media.ContentType
	SizeBytes    int64
	ExpiresAt    time.Time
}

type DirectAttachmentUpload struct {
	UploadID     string
	AttachmentID media.ID
	Method       string
	URL          string
	Headers      map[string]string
	FormFields   map[string]string
	ExpiresAt    time.Time
}

type CompletedDirectAttachmentUpload struct {
	UploadID     string
	AttachmentID media.ID
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	AssetID      asset.ID
	StorageKey   media.StorageKey
	FileName     media.FileName
	ContentType  media.ContentType
	SizeBytes    int64
	SHA256       media.SHA256
	ExpiresAt    time.Time
}

type DirectAttachmentUploader interface {
	CreateDirectAttachmentUpload(ctx context.Context, request DirectAttachmentUploadRequest) (DirectAttachmentUpload, error)
	CompleteDirectAttachmentUpload(ctx context.Context, uploadID string) (CompletedDirectAttachmentUpload, error)
}

type ImageDerivativeRequest struct {
	Attachment  media.Attachment
	Variant     media.ThumbnailVariant
	ContentType media.ContentType
	Content     []byte
}

type ImageDerivative struct {
	ContentType media.ContentType
	Content     []byte
}

type ModelImageRequest struct {
	Attachment  media.Attachment
	ContentType media.ContentType
	Content     []byte
}

type ModelImage struct {
	ContentType media.ContentType
	Content     []byte
	SizeBytes   int64
	SHA256      media.SHA256
	Width       int
	Height      int
}

type ImageProcessor interface {
	CreateThumbnail(ctx context.Context, request ImageDerivativeRequest) (ImageDerivative, error)
	PrepareImageForModelUse(ctx context.Context, request ModelImageRequest) (ModelImage, error)
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
	ClaimPendingBlobDeletionEvents(ctx context.Context, claimID string, limit int, now time.Time, leaseUntil time.Time) ([]BlobDeletionEvent, error)
	MarkBlobDeletionEventProcessed(ctx context.Context, eventID string, claimID string) error
	MarkBlobDeletionEventFailed(ctx context.Context, eventID string, claimID string, reason string) error
	MarkBlobDeletionEventDeadLettered(ctx context.Context, eventID string, claimID string, reason string) error
}
