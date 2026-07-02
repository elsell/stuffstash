package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestAttachmentThumbnailCachesDerivativeByVariant(t *testing.T) {
	content := pngAttachmentBytes()
	attachment := attachmentFixture(t, media.ContentTypePNG, content)
	repository := &recordingAttachmentRepository{attachment: attachment, found: true}
	blobStore := &recordingBlobStorage{blobs: map[media.StorageKey][]byte{attachment.StorageKey: content}}
	processor := &recordingImageProcessor{thumbnailContent: []byte("small-thumb")}
	application := New(Dependencies{
		Observer:             noopObserver{},
		Authorizer:           allowInventoryAuthorizer{},
		Tenants:              attachmentTenantRepository{},
		TenantUnitOfWork:     attachmentTenantRepository{},
		Inventories:          attachmentInventoryRepository{},
		InventoryUnitOfWork:  attachmentInventoryRepository{},
		Assets:               attachmentAssetRepository{},
		Attachments:          repository,
		AttachmentUnitOfWork: repository,
		Blobs:                blobStore,
		ImageProcessor:       processor,
		Audit:                &fakeAuditRepository{},
		IDs:                  &attachmentIDGenerator{ids: []string{"audit-first", "audit-second"}},
		MaxAttachmentBytes:   32,
	})

	first, err := application.DownloadAttachmentThumbnail(context.Background(), DownloadAttachmentThumbnailInput{
		Principal:    identity.Principal{ID: "viewer"},
		Source:       audit.SourceAPI,
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		AssetID:      asset.ID("asset-one"),
		AttachmentID: attachment.ID,
		Variant:      "small",
	})
	if err != nil {
		t.Fatalf("download first thumbnail: %v", err)
	}
	second, err := application.DownloadAttachmentThumbnail(context.Background(), DownloadAttachmentThumbnailInput{
		Principal:    identity.Principal{ID: "viewer"},
		Source:       audit.SourceAPI,
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		AssetID:      asset.ID("asset-one"),
		AttachmentID: attachment.ID,
		Variant:      "small",
	})
	if err != nil {
		t.Fatalf("download cached thumbnail: %v", err)
	}

	if string(first.Content) != "small-thumb" || string(second.Content) != "small-thumb" {
		t.Fatalf("expected cached thumbnail content, got first=%q second=%q", first.Content, second.Content)
	}
	if first.ContentType != media.ContentTypePNG || second.ContentType != media.ContentTypePNG {
		t.Fatalf("expected cached thumbnail content type to come from processor metadata, got first=%q second=%q", first.ContentType, second.ContentType)
	}
	if processor.thumbnailCalls != 1 {
		t.Fatalf("expected thumbnail processor once, got %d", processor.thumbnailCalls)
	}
	if blobStore.getCount(attachment.StorageKey) != 1 {
		t.Fatalf("expected original blob read once, got %d", blobStore.getCount(attachment.StorageKey))
	}
	cacheKey, ok := thumbnailStorageKey(attachment, media.ThumbnailVariantSmall)
	if !ok {
		t.Fatalf("expected cache key")
	}
	if blobStore.getCount(cacheKey) != 1 || !blobStore.hasBlob(cacheKey) {
		t.Fatalf("expected cache key to be read once and stored, gets=%d stored=%t", blobStore.getCount(cacheKey), blobStore.hasBlob(cacheKey))
	}
	metadataKey, ok := thumbnailMetadataStorageKey(cacheKey)
	if !ok {
		t.Fatalf("expected cache metadata key")
	}
	if blobStore.getCount(metadataKey) != 2 || !blobStore.hasBlob(metadataKey) {
		t.Fatalf("expected metadata key to be read twice and stored, gets=%d stored=%t", blobStore.getCount(metadataKey), blobStore.hasBlob(metadataKey))
	}
}

func TestAttachmentThumbnailServesGeneratedDerivativeWhenCacheWriteFails(t *testing.T) {
	content := pngAttachmentBytes()
	attachment := attachmentFixture(t, media.ContentTypePNG, content)
	blobStore := &recordingBlobStorage{
		blobs:  map[media.StorageKey][]byte{attachment.StorageKey: content},
		putErr: errors.New("cache unavailable"),
	}
	processor := &recordingImageProcessor{thumbnailContent: []byte("uncached-thumb")}
	application := New(Dependencies{
		Observer:             noopObserver{},
		Authorizer:           allowInventoryAuthorizer{},
		Tenants:              attachmentTenantRepository{},
		TenantUnitOfWork:     attachmentTenantRepository{},
		Inventories:          attachmentInventoryRepository{},
		InventoryUnitOfWork:  attachmentInventoryRepository{},
		Assets:               attachmentAssetRepository{},
		Attachments:          &recordingAttachmentRepository{attachment: attachment, found: true},
		AttachmentUnitOfWork: &recordingAttachmentRepository{},
		Blobs:                blobStore,
		ImageProcessor:       processor,
		Audit:                &fakeAuditRepository{},
		IDs:                  &attachmentIDGenerator{ids: []string{"audit-thumb"}},
		MaxAttachmentBytes:   32,
	})

	thumbnail, err := application.DownloadAttachmentThumbnail(context.Background(), DownloadAttachmentThumbnailInput{
		Principal:    identity.Principal{ID: "viewer"},
		Source:       audit.SourceAPI,
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		AssetID:      asset.ID("asset-one"),
		AttachmentID: attachment.ID,
		Variant:      "small",
	})
	if err != nil {
		t.Fatalf("download thumbnail: %v", err)
	}
	if string(thumbnail.Content) != "uncached-thumb" || processor.thumbnailCalls != 1 {
		t.Fatalf("expected generated thumbnail despite cache failure, got %+v calls=%d", thumbnail, processor.thumbnailCalls)
	}
}

func TestDrainBlobDeletionOutboxDeletesCachedThumbnailDerivatives(t *testing.T) {
	observer := &fakeObserver{}
	storageKey, ok := media.NewStorageKey("tenant/inventory/asset/attachment-one")
	if !ok {
		t.Fatalf("expected storage key")
	}
	outbox := &singleBlobDeletionOutbox{
		event: ports.BlobDeletionEvent{
			ID:         "blob-event-one",
			StorageKey: storageKey,
		},
	}
	blobStore := &recordingBlobStorage{blobs: map[media.StorageKey][]byte{
		storageKey: []byte("original"),
	}}
	for _, key := range thumbnailStorageKeysForBlob(storageKey) {
		blobStore.blobs[key] = []byte("thumb")
	}
	application := New(Dependencies{
		Observer:                      observer,
		Blobs:                         blobStore,
		BlobDeletionOutbox:            outbox,
		IDs:                           &attachmentIDGenerator{ids: []string{"claim-one"}},
		Clock:                         fakeClock{now: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)},
		BlobDeletionOutboxMaxAttempts: 2,
	})

	if err := application.DrainBlobDeletionOutbox(context.Background(), 1); err != nil {
		t.Fatalf("drain blob deletion outbox: %v", err)
	}

	for _, key := range append([]media.StorageKey{storageKey}, thumbnailStorageKeysForBlob(storageKey)...) {
		if blobStore.hasBlob(key) {
			t.Fatalf("expected blob %q to be deleted", key.String())
		}
		if !blobStore.deletedKey(key) {
			t.Fatalf("expected delete attempt for %q", key.String())
		}
	}
	if !outbox.processed {
		t.Fatalf("expected outbox event to be processed")
	}
}

func TestDrainBlobDeletionOutboxAttemptsAllThumbnailDeletesBeforeFailing(t *testing.T) {
	storageKey, ok := media.NewStorageKey("tenant/inventory/asset/attachment-one")
	if !ok {
		t.Fatalf("expected storage key")
	}
	outbox := &singleBlobDeletionOutbox{
		event: ports.BlobDeletionEvent{
			ID:         "blob-event-one",
			StorageKey: storageKey,
		},
	}
	thumbnailKeys := thumbnailStorageKeysForBlob(storageKey)
	blobStore := &recordingBlobStorage{
		deleteErrs: map[media.StorageKey]error{
			storageKey: errors.New("original delete failed"),
		},
	}
	application := New(Dependencies{
		Observer:                      noopObserver{},
		Blobs:                         blobStore,
		BlobDeletionOutbox:            outbox,
		IDs:                           &attachmentIDGenerator{ids: []string{"claim-one"}},
		Clock:                         fakeClock{now: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)},
		BlobDeletionOutboxMaxAttempts: 2,
	})

	if err := application.DrainBlobDeletionOutbox(context.Background(), 1); err != nil {
		t.Fatalf("drain should mark failure without returning worker error: %v", err)
	}

	for _, key := range append([]media.StorageKey{storageKey}, thumbnailKeys...) {
		if !blobStore.deletedKey(key) {
			t.Fatalf("expected delete attempt for %q", key.String())
		}
	}
	if !outbox.failed || outbox.processed {
		t.Fatalf("expected event failure after all delete attempts, processed=%t failed=%t", outbox.processed, outbox.failed)
	}
}
