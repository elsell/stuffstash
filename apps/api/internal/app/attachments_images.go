package app

import (
	"context"

	mediaapp "github.com/stuffstash/stuff-stash/internal/app/media"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type appNoopObserver = mediaapp.AppNoopObserver
type primaryThumbnailWarmState = mediaapp.PrimaryThumbnailWarmState
type thumbnailGenerationState = mediaapp.ThumbnailGenerationState
type DownloadAttachmentThumbnailInput = mediaapp.DownloadAttachmentThumbnailInput
type PrepareAttachmentForModelUseInput = mediaapp.PrepareAttachmentForModelUseInput
type AttachmentThumbnailResult = mediaapp.AttachmentThumbnailResult

const defaultPrimarySmallThumbnailWarmLimit = mediaapp.DefaultPrimarySmallThumbnailWarmLimit
const defaultPrimarySmallThumbnailWarmConcurrency = mediaapp.DefaultPrimarySmallThumbnailWarmConcurrency
const defaultPrimarySmallThumbnailWarmTimeout = mediaapp.DefaultPrimarySmallThumbnailWarmTimeout

func newPrimaryThumbnailWarmState(concurrency int) *primaryThumbnailWarmState {
	return mediaapp.NewPrimaryThumbnailWarmState(concurrency)
}
func newThumbnailGenerationState() *thumbnailGenerationState {
	return mediaapp.NewThumbnailGenerationState()
}
func (a App) DownloadAttachmentThumbnail(ctx context.Context, input DownloadAttachmentThumbnailInput) (AttachmentThumbnailResult, error) {
	return a.attachmentService().DownloadAttachmentThumbnail(ctx, input)
}
func (a App) PrepareAttachmentForModelUse(ctx context.Context, input PrepareAttachmentForModelUseInput) (ports.ModelImage, error) {
	return a.attachmentService().PrepareAttachmentForModelUse(ctx, input)
}
func (a App) warmPrimarySmallThumbnails(ctx context.Context, attachments []media.Attachment) {
	a.attachmentService().WarmPrimarySmallThumbnails(ctx, attachments)
}
func thumbnailStorageKey(attachment media.Attachment, variant media.ThumbnailVariant) (media.StorageKey, bool) {
	return mediaapp.ThumbnailStorageKey(attachment, variant)
}
func thumbnailMetadataStorageKey(cacheKey media.StorageKey) (media.StorageKey, bool) {
	return mediaapp.ThumbnailMetadataStorageKey(cacheKey)
}
func thumbnailStorageKeysForBlob(storageKey media.StorageKey) []media.StorageKey {
	return mediaapp.ThumbnailStorageKeysForBlob(storageKey)
}
