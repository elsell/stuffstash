package app

import (
	"context"

	mediaapp "github.com/stuffstash/stuff-stash/internal/app/media"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
)

type CreateAttachmentInput = mediaapp.CreateAttachmentInput
type ListAttachmentsInput = mediaapp.ListAttachmentsInput
type GetAttachmentInput = mediaapp.GetAttachmentInput
type DownloadAttachmentInput = mediaapp.DownloadAttachmentInput
type UpdateAttachmentLifecycleInput = mediaapp.UpdateAttachmentLifecycleInput
type ListAttachmentsResult = mediaapp.ListAttachmentsResult
type AttachmentContentResult = mediaapp.AttachmentContentResult
type preparedAttachment = mediaapp.PreparedAttachment

func (a App) attachmentService() mediaapp.AttachmentService {
	return mediaapp.NewAttachmentService(mediaapp.AttachmentDependencies{Access: a.inventoryService(), Audit: a.audit,
		Assets:                      a.assets,
		AttachmentUnitOfWork:        a.attachmentUnitOfWork,
		Attachments:                 a.attachments,
		BlobDeletionClaimLease:      a.blobDeletionClaimLease,
		BlobDeletionMaxAttempts:     a.blobDeletionMaxAttempts,
		BlobDeletionOutbox:          a.blobDeletionOutbox,
		Blobs:                       a.blobs,
		Clock:                       a.clock,
		DefaultPageLimit:            a.defaultPageLimit,
		DirectUploadTTL:             a.directUploadTTL,
		DirectUploads:               a.directUploads,
		IDs:                         a.ids,
		ImageProcessor:              a.imageProcessor,
		MaxAttachmentBytes:          a.maxAttachmentBytes,
		MaxPageLimit:                a.maxPageLimit,
		Observer:                    a.observer,
		PrimaryThumbnailWarmLimit:   a.primaryThumbnailWarmLimit,
		PrimaryThumbnailWarmTimeout: a.primaryThumbnailWarmTimeout,
		ThumbnailGenerationState:    a.thumbnailGenerationState,
		ThumbnailReader:             a.thumbnailReader,
		ThumbnailWarmState:          a.thumbnailWarmState,
	})
}
func (a App) CreateAttachment(ctx context.Context, input CreateAttachmentInput) (media.Attachment, error) {
	return a.attachmentService().CreateAttachment(ctx, input)
}
func (a App) prepareAttachment(ctx context.Context, input CreateAttachmentInput) (preparedAttachment, error) {
	return a.attachmentService().PrepareAttachment(ctx, input)
}
func (a App) recordAttachmentCreated(ctx context.Context, input CreateAttachmentInput, attachment media.Attachment) {
	a.attachmentService().RecordAttachmentCreated(ctx, input, attachment)
}
func (a App) ListAttachments(ctx context.Context, input ListAttachmentsInput) (ListAttachmentsResult, error) {
	return a.attachmentService().ListAttachments(ctx, input)
}
func (a App) GetAttachment(ctx context.Context, input GetAttachmentInput) (media.Attachment, error) {
	return a.attachmentService().GetAttachment(ctx, input)
}
func (a App) DownloadAttachment(ctx context.Context, input DownloadAttachmentInput) (AttachmentContentResult, error) {
	return a.attachmentService().DownloadAttachment(ctx, input)
}
func (a App) ArchiveAttachment(ctx context.Context, input UpdateAttachmentLifecycleInput) (media.Attachment, error) {
	return a.attachmentService().ArchiveAttachment(ctx, input)
}
func (a App) RestoreAttachment(ctx context.Context, input UpdateAttachmentLifecycleInput) (media.Attachment, error) {
	return a.attachmentService().RestoreAttachment(ctx, input)
}
func (a App) DeleteAttachment(ctx context.Context, input UpdateAttachmentLifecycleInput) error {
	return a.attachmentService().DeleteAttachment(ctx, input)
}
func (a App) DrainBlobDeletionOutbox(ctx context.Context, limit int) error {
	return a.attachmentService().DrainBlobDeletionOutbox(ctx, limit)
}
