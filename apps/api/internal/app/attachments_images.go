package app

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"golang.org/x/sync/singleflight"
)

const defaultPrimarySmallThumbnailWarmLimit = 12
const defaultPrimarySmallThumbnailWarmConcurrency = 4
const defaultPrimarySmallThumbnailWarmTimeout = 10 * time.Second

type appNoopObserver struct{}

func (appNoopObserver) Record(context.Context, ports.Event) {}

type primaryThumbnailWarmState struct {
	sem chan struct{}
}

func newPrimaryThumbnailWarmState(concurrency int) *primaryThumbnailWarmState {
	return &primaryThumbnailWarmState{
		sem: make(chan struct{}, normalizePrimaryThumbnailWarmConcurrency(concurrency)),
	}
}

func (s *primaryThumbnailWarmState) tryAcquire() bool {
	if s == nil {
		return false
	}
	select {
	case s.sem <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *primaryThumbnailWarmState) release() {
	if s == nil {
		return
	}
	<-s.sem
}

type thumbnailGenerationState struct {
	group singleflight.Group
}

func newThumbnailGenerationState() *thumbnailGenerationState {
	return &thumbnailGenerationState{}
}

type thumbnailGenerationResult struct {
	contentType media.ContentType
	content     []byte
	source      string
}

type DownloadAttachmentThumbnailInput struct {
	Principal    identity.Principal
	Source       audit.Source
	RequestID    string
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	AssetID      asset.ID
	AttachmentID media.ID
	Variant      string
}

type PrepareAttachmentForModelUseInput struct {
	Principal    identity.Principal
	Source       audit.Source
	RequestID    string
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	AssetID      asset.ID
	AttachmentID media.ID
}

type AttachmentThumbnailResult struct {
	Attachment  media.Attachment
	ContentType media.ContentType
	Content     []byte
}

func (a App) DownloadAttachmentThumbnail(ctx context.Context, input DownloadAttachmentThumbnailInput) (AttachmentThumbnailResult, error) {
	if a.imageProcessor == nil {
		return AttachmentThumbnailResult{}, ErrInvalidInput
	}
	variant, ok := media.NewThumbnailVariant(input.Variant)
	if !ok {
		return AttachmentThumbnailResult{}, ErrInvalidInput
	}
	attachment, err := a.authorizeImageAttachmentRead(ctx, input.Principal, input.Source, input.RequestID, input.TenantID, input.InventoryID, input.AssetID, input.AttachmentID)
	if err != nil {
		return AttachmentThumbnailResult{}, err
	}
	thumbnail, err := a.getOrGenerateThumbnail(ctx, attachment, variant)
	if err != nil {
		return AttachmentThumbnailResult{}, err
	}
	a.recordAttachmentThumbnailServed(ctx, input, attachment, variant, thumbnail.source)
	return AttachmentThumbnailResult{Attachment: attachment, ContentType: thumbnail.contentType, Content: thumbnail.content}, nil
}

func (a App) PrepareAttachmentForModelUse(ctx context.Context, input PrepareAttachmentForModelUseInput) (ports.ModelImage, error) {
	attachment, content, err := a.readAttachmentBlobForImageWork(ctx, input.Principal, input.Source, input.RequestID, input.TenantID, input.InventoryID, input.AssetID, input.AttachmentID)
	if err != nil {
		return ports.ModelImage{}, err
	}
	if a.imageProcessor == nil {
		return ports.ModelImage{}, ErrInvalidInput
	}
	prepared, err := a.imageProcessor.PrepareImageForModelUse(ctx, ports.ModelImageRequest{
		Attachment:  attachment,
		ContentType: attachment.ContentType,
		Content:     content,
	})
	if err != nil {
		a.observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "model image preparation failed"})
		return ports.ModelImage{}, err
	}
	if !prepared.ContentType.IsImage() || len(prepared.Content) == 0 || prepared.SizeBytes <= 0 || prepared.SHA256.String() == "" {
		return ports.ModelImage{}, ErrInvalidInput
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAttachmentModelImagePrepared,
		Message: "attachment model image prepared",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"asset_id":      input.AssetID.String(),
			"attachment_id": attachment.ID.String(),
			"principal_id":  input.Principal.ID.String(),
		},
	})
	return prepared, nil
}

func (a App) readAttachmentBlobForImageWork(ctx context.Context, principal identity.Principal, source audit.Source, requestID string, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, attachmentID media.ID) (media.Attachment, []byte, error) {
	attachment, err := a.authorizeImageAttachmentRead(ctx, principal, source, requestID, tenantID, inventoryID, assetID, attachmentID)
	if err != nil {
		return media.Attachment{}, nil, err
	}
	content, err := a.blobs.GetBlob(ctx, attachment.StorageKey)
	if err != nil {
		a.observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "blob storage failed"})
		return media.Attachment{}, nil, err
	}
	return attachment, content, nil
}

func (a App) authorizeImageAttachmentRead(ctx context.Context, principal identity.Principal, source audit.Source, requestID string, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, attachmentID media.ID) (media.Attachment, error) {
	if err := a.ensureActiveInventoryAccess(ctx, principal, tenantID, inventoryID, ports.InventoryPermissionView); err != nil {
		return media.Attachment{}, err
	}
	if err := a.ensureActiveAssetForAttachment(ctx, tenantID, inventoryID, assetID); err != nil {
		return media.Attachment{}, err
	}
	attachment, found, err := a.attachments.AttachmentByID(ctx, tenantID, inventoryID, assetID, attachmentID)
	if err != nil {
		return media.Attachment{}, err
	}
	if !found {
		return media.Attachment{}, ErrNotFound
	}
	if !attachment.ContentType.IsImage() {
		return media.Attachment{}, ErrInvalidInput
	}
	if err := a.saveReadAuditRecord(ctx, auditRecordInput{
		Principal:   principal,
		TenantID:    tenantID,
		InventoryID: inventoryID,
		Source:      source,
		RequestID:   requestID,
		Action:      audit.ActionAttachmentContentDownloaded,
		TargetType:  audit.TargetAttachment,
		TargetID:    attachment.ID.String(),
		Metadata: map[string]string{
			"asset_id": assetID.String(),
		},
	}); err != nil {
		return media.Attachment{}, err
	}
	return attachment, nil
}

func (a App) recordAttachmentThumbnailServed(ctx context.Context, input DownloadAttachmentThumbnailInput, attachment media.Attachment, variant media.ThumbnailVariant, source string) {
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAttachmentThumbnailGenerated,
		Message: "attachment thumbnail generated",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"asset_id":      input.AssetID.String(),
			"attachment_id": attachment.ID.String(),
			"variant":       variant.String(),
			"principal_id":  input.Principal.ID.String(),
			"source":        source,
		},
	})
}

func (a App) warmPrimarySmallThumbnails(ctx context.Context, attachments []media.Attachment) {
	if a.blobs == nil || a.imageProcessor == nil || a.thumbnailWarmState == nil || len(attachments) == 0 {
		return
	}
	limit := len(attachments)
	if limit > a.primaryThumbnailWarmLimit {
		limit = a.primaryThumbnailWarmLimit
	}
	for _, attachment := range attachments[:limit] {
		if !attachment.ContentType.IsImage() {
			continue
		}
		if _, ok := thumbnailStorageKey(attachment, media.ThumbnailVariantSmall); !ok {
			continue
		}
		if !a.thumbnailWarmState.tryAcquire() {
			continue
		}
		warmCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), a.primaryThumbnailWarmTimeout)
		go func(attachment media.Attachment) {
			defer func() {
				cancel()
				a.thumbnailWarmState.release()
			}()
			a.warmPrimarySmallThumbnail(warmCtx, attachment)
		}(attachment)
	}
}

func (a App) warmPrimarySmallThumbnail(ctx context.Context, attachment media.Attachment) {
	_, _ = a.getOrGenerateThumbnail(ctx, attachment, media.ThumbnailVariantSmall)
}

func (a App) getOrGenerateThumbnail(ctx context.Context, attachment media.Attachment, variant media.ThumbnailVariant) (thumbnailGenerationResult, error) {
	cacheKey, ok := thumbnailStorageKey(attachment, variant)
	if !ok {
		return thumbnailGenerationResult{}, ErrInvalidInput
	}
	if a.thumbnailGenerationState == nil {
		return a.generateThumbnail(ctx, attachment, variant, cacheKey)
	}
	result, err := a.generateThumbnailSingleflight(ctx, attachment, variant, cacheKey)
	if err == nil || ctx.Err() != nil {
		return result, err
	}
	return a.generateThumbnailSingleflight(ctx, attachment, variant, cacheKey)
}

func (a App) generateThumbnailSingleflight(ctx context.Context, attachment media.Attachment, variant media.ThumbnailVariant, cacheKey media.StorageKey) (thumbnailGenerationResult, error) {
	value, err, _ := a.thumbnailGenerationState.group.Do(cacheKey.String(), func() (any, error) {
		if cached, cachedContentType, err := a.cachedThumbnail(ctx, cacheKey); err == nil {
			return thumbnailGenerationResult{contentType: cachedContentType, content: cached, source: "cache"}, nil
		} else if !errors.Is(err, ports.ErrBlobNotFound) {
			a.observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "thumbnail cache read failed"})
			return thumbnailGenerationResult{}, err
		}
		return a.generateThumbnail(ctx, attachment, variant, cacheKey)
	})
	if err != nil {
		return thumbnailGenerationResult{}, err
	}
	result, ok := value.(thumbnailGenerationResult)
	if !ok {
		return thumbnailGenerationResult{}, ErrInvalidInput
	}
	return result, nil
}

func (a App) generateThumbnail(ctx context.Context, attachment media.Attachment, variant media.ThumbnailVariant, cacheKey media.StorageKey) (thumbnailGenerationResult, error) {
	content, err := a.blobs.GetBlob(ctx, attachment.StorageKey)
	if err != nil {
		a.observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "blob storage failed"})
		return thumbnailGenerationResult{}, err
	}
	thumbnail, err := a.imageProcessor.CreateThumbnail(ctx, ports.ImageDerivativeRequest{
		Attachment:  attachment,
		Variant:     variant,
		ContentType: attachment.ContentType,
		Content:     content,
	})
	if err != nil {
		a.observer.Record(ctx, ports.Event{
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
		return thumbnailGenerationResult{}, ErrInvalidInput
	}
	a.cacheThumbnailBestEffort(ctx, cacheKey, thumbnail)
	return thumbnailGenerationResult{contentType: thumbnail.ContentType, content: thumbnail.Content, source: "generated"}, nil
}

func (a App) cachedThumbnail(ctx context.Context, cacheKey media.StorageKey) ([]byte, media.ContentType, error) {
	metadataKey, ok := thumbnailMetadataStorageKey(cacheKey)
	if !ok {
		return nil, "", ErrInvalidInput
	}
	metadata, err := a.blobs.GetBlob(ctx, metadataKey)
	if err != nil {
		return nil, "", err
	}
	contentType, ok := media.NewContentType(string(metadata))
	if !ok || !contentType.IsImage() {
		return nil, "", ports.ErrBlobNotFound
	}
	content, err := a.blobs.GetBlob(ctx, cacheKey)
	if err != nil {
		return nil, "", err
	}
	return content, contentType, nil
}

func (a App) cacheThumbnailBestEffort(ctx context.Context, cacheKey media.StorageKey, thumbnail ports.ImageDerivative) {
	if err := a.blobs.PutBlob(ctx, cacheKey, thumbnail.ContentType, thumbnail.Content); err != nil {
		a.observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "thumbnail cache write failed"})
		return
	}
	metadataKey, ok := thumbnailMetadataStorageKey(cacheKey)
	if !ok {
		a.observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "thumbnail cache metadata key invalid"})
		return
	}
	if err := a.blobs.PutBlob(ctx, metadataKey, media.ContentType("text/plain"), []byte(thumbnail.ContentType.String())); err != nil {
		a.observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "thumbnail cache metadata write failed"})
	}
}

func thumbnailStorageKey(attachment media.Attachment, variant media.ThumbnailVariant) (media.StorageKey, bool) {
	return thumbnailStorageKeyForBlob(attachment.StorageKey, variant)
}

func thumbnailStorageKeyForBlob(storageKey media.StorageKey, variant media.ThumbnailVariant) (media.StorageKey, bool) {
	return media.NewStorageKey(storageKey.String() + ".thumb/" + variant.String())
}

func thumbnailMetadataStorageKey(cacheKey media.StorageKey) (media.StorageKey, bool) {
	return media.NewStorageKey(cacheKey.String() + ".meta")
}

func thumbnailStorageKeysForBlob(storageKey media.StorageKey) []media.StorageKey {
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
			if metadataKey, ok := thumbnailMetadataStorageKey(key); ok {
				keys = append(keys, metadataKey)
			}
		}
	}
	return keys
}
