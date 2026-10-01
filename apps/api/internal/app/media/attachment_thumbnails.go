package media

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a AttachmentService) WarmPrimarySmallThumbnails(ctx context.Context, attachments []media.Attachment) {
	// Durable upload scheduling and backfill replace read-triggered warming.
	if a.deps.ThumbnailReader != nil {
		return
	}
	if a.deps.Blobs == nil || a.deps.ImageProcessor == nil || a.deps.ThumbnailWarmState == nil || len(attachments) == 0 {
		return
	}
	limit := len(attachments)
	if limit > a.deps.PrimaryThumbnailWarmLimit {
		limit = a.deps.PrimaryThumbnailWarmLimit
	}
	for _, attachment := range attachments[:limit] {
		if !attachment.ContentType.IsImage() {
			continue
		}
		if _, ok := ThumbnailStorageKey(attachment, media.ThumbnailVariantSmall); !ok {
			continue
		}
		if !a.deps.ThumbnailWarmState.tryAcquire() {
			continue
		}
		warmCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), a.deps.PrimaryThumbnailWarmTimeout)
		go func(attachment media.Attachment) {
			defer func() {
				cancel()
				a.deps.ThumbnailWarmState.release()
			}()
			a.warmPrimarySmallThumbnail(warmCtx, attachment)
		}(attachment)
	}
}

func (a AttachmentService) warmPrimarySmallThumbnail(ctx context.Context, attachment media.Attachment) {
	_, _ = a.getOrGenerateThumbnail(ctx, attachment, media.ThumbnailVariantSmall)
}

func (a AttachmentService) getOrGenerateThumbnail(ctx context.Context, attachment media.Attachment, variant media.ThumbnailVariant) (thumbnailGenerationResult, error) {
	if a.deps.ThumbnailReader != nil {
		derivative, cached, err := a.deps.ThumbnailReader.ReadThumbnail(ctx, attachment, variant)
		if err != nil {
			return thumbnailGenerationResult{}, err
		}
		source := "generated"
		if cached {
			source = "cache"
		}
		return thumbnailGenerationResult{contentType: derivative.ContentType, content: derivative.Content, source: source}, nil
	}

	cacheKey, ok := ThumbnailStorageKey(attachment, variant)
	if !ok {
		return thumbnailGenerationResult{}, apperrors.ErrInvalidInput
	}
	if a.deps.ThumbnailGenerationState == nil {
		return a.generateThumbnail(ctx, attachment, variant, cacheKey)
	}
	result, err := a.generateThumbnailSingleflight(ctx, attachment, variant, cacheKey)
	if err == nil || ctx.Err() != nil {
		return result, err
	}
	return a.generateThumbnailSingleflight(ctx, attachment, variant, cacheKey)
}

func (a AttachmentService) generateThumbnailSingleflight(ctx context.Context, attachment media.Attachment, variant media.ThumbnailVariant, cacheKey media.StorageKey) (thumbnailGenerationResult, error) {
	value, err, _ := a.deps.ThumbnailGenerationState.group.Do(cacheKey.String(), func() (any, error) {
		if cached, cachedContentType, err := a.cachedThumbnail(ctx, cacheKey); err == nil {
			return thumbnailGenerationResult{contentType: cachedContentType, content: cached, source: "cache"}, nil
		} else if !errors.Is(err, ports.ErrBlobNotFound) {
			a.deps.Observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "thumbnail cache read failed"})
			return thumbnailGenerationResult{}, err
		}
		return a.generateThumbnail(ctx, attachment, variant, cacheKey)
	})
	if err != nil {
		return thumbnailGenerationResult{}, err
	}
	result, ok := value.(thumbnailGenerationResult)
	if !ok {
		return thumbnailGenerationResult{}, apperrors.ErrInvalidInput
	}
	return result, nil
}

func (a AttachmentService) generateThumbnail(ctx context.Context, attachment media.Attachment, variant media.ThumbnailVariant, cacheKey media.StorageKey) (thumbnailGenerationResult, error) {
	content, err := a.deps.Blobs.GetBlob(ctx, attachment.StorageKey)
	if err != nil {
		a.deps.Observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "blob storage failed"})
		return thumbnailGenerationResult{}, err
	}
	thumbnail, err := a.deps.ImageProcessor.CreateThumbnail(ctx, ports.ImageDerivativeRequest{
		Attachment:  attachment,
		Variant:     variant,
		ContentType: attachment.ContentType,
		Content:     content,
	})
	if err != nil {
		a.deps.Observer.Record(ctx, ports.Event{
			Name:    ports.EventBlobStorageFailed,
			Message: "thumbnail generation failed",
			Fields: map[string]string{
				"content_type": attachment.ContentType.String(),
				"reason":       "image_processing_failed",
				"variant":      variant.String(),
			},
		})
		return thumbnailGenerationResult{}, err
	}
	if !thumbnail.ContentType.IsImage() || len(thumbnail.Content) == 0 {
		return thumbnailGenerationResult{}, apperrors.ErrInvalidInput
	}
	a.cacheThumbnailBestEffort(ctx, cacheKey, thumbnail)
	return thumbnailGenerationResult{contentType: thumbnail.ContentType, content: thumbnail.Content, source: "generated"}, nil
}

func (a AttachmentService) cachedThumbnail(ctx context.Context, cacheKey media.StorageKey) ([]byte, media.ContentType, error) {
	metadataKey, ok := ThumbnailMetadataStorageKey(cacheKey)
	if !ok {
		return nil, "", apperrors.ErrInvalidInput
	}
	metadata, err := a.deps.Blobs.GetBlob(ctx, metadataKey)
	if err != nil {
		return nil, "", err
	}
	contentType, ok := media.NewContentType(string(metadata))
	if !ok || !contentType.IsImage() {
		return nil, "", ports.ErrBlobNotFound
	}
	content, err := a.deps.Blobs.GetBlob(ctx, cacheKey)
	if err != nil {
		return nil, "", err
	}
	return content, contentType, nil
}

func (a AttachmentService) cacheThumbnailBestEffort(ctx context.Context, cacheKey media.StorageKey, thumbnail ports.ImageDerivative) {
	if err := a.deps.Blobs.PutBlob(ctx, cacheKey, thumbnail.ContentType, thumbnail.Content); err != nil {
		a.deps.Observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "thumbnail cache write failed"})
		return
	}
	metadataKey, ok := ThumbnailMetadataStorageKey(cacheKey)
	if !ok {
		a.deps.Observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "thumbnail cache metadata key invalid"})
		return
	}
	if err := a.deps.Blobs.PutBlob(ctx, metadataKey, media.ContentType("text/plain"), []byte(thumbnail.ContentType.String())); err != nil {
		a.deps.Observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "thumbnail cache metadata write failed"})
	}
}

func ThumbnailStorageKey(attachment media.Attachment, variant media.ThumbnailVariant) (media.StorageKey, bool) {
	return thumbnailStorageKeyForBlob(attachment.StorageKey, variant)
}

func thumbnailStorageKeyForBlob(storageKey media.StorageKey, variant media.ThumbnailVariant) (media.StorageKey, bool) {
	return media.NewStorageKey(storageKey.String() + ".thumb/" + variant.String())
}

func ThumbnailMetadataStorageKey(cacheKey media.StorageKey) (media.StorageKey, bool) {
	return media.NewStorageKey(cacheKey.String() + ".meta")
}

func ThumbnailStorageKeysForBlob(storageKey media.StorageKey) []media.StorageKey {
	variants := []media.ThumbnailVariant{
		media.ThumbnailVariantSmall,
		media.ThumbnailVariantMedium,
		media.ThumbnailVariantLarge,
	}
	keys := make([]media.StorageKey, 0, len(variants))
	for _, variant := range variants {
		key, ok := thumbnailStorageKeyForBlob(storageKey, variant)
		if ok {
			keys = append(keys, key)
			if metadataKey, ok := ThumbnailMetadataStorageKey(key); ok {
				keys = append(keys, metadataKey)
			}
		}
	}
	return keys
}

func NormalizePrimaryThumbnailWarmLimit(limit int) int {
	if limit <= 0 {
		return DefaultPrimarySmallThumbnailWarmLimit
	}
	return limit
}
func NormalizePrimaryThumbnailWarmConcurrency(concurrency int) int {
	if concurrency <= 0 {
		return DefaultPrimarySmallThumbnailWarmConcurrency
	}
	return concurrency
}
func NormalizePrimaryThumbnailWarmTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return DefaultPrimarySmallThumbnailWarmTimeout
	}
	return timeout
}
