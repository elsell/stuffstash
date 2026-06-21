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
	ClaimPendingBlobDeletionEvents(ctx context.Context, claimID string, limit int, now time.Time, leaseUntil time.Time) ([]BlobDeletionEvent, error)
	MarkBlobDeletionEventProcessed(ctx context.Context, eventID string, claimID string) error
	MarkBlobDeletionEventFailed(ctx context.Context, eventID string, claimID string, reason string) error
	MarkBlobDeletionEventDeadLettered(ctx context.Context, eventID string, claimID string, reason string) error
}
