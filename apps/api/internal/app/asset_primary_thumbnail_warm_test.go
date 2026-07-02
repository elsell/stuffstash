package app

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestListAssetsWarmsPrimarySmallThumbnails(t *testing.T) {
	content := pngAttachmentBytes()
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	item := assetItem("asset-one", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	attachment := attachmentForAsset(t, item, "attachment-one", media.ContentTypePNG, content)
	attachments := &recordingAttachmentRepository{attachment: attachment, found: true}
	blobStore := &recordingBlobStorage{blobs: map[media.StorageKey][]byte{attachment.StorageKey: content}}
	processor := &warmImageProcessor{thumbnailContent: []byte("small-card-thumb")}
	application := New(Dependencies{
		Observer:   noopObserver{},
		Authorizer: &fakeAuthorizer{},
		Tenants:    &fakeTenantRepository{exists: true},
		Inventories: &fakeInventoryRepository{items: []inventory.Inventory{
			inventoryItem(inventoryID.String(), tenantID.String(), "Home"),
		}},
		Assets:             &fakeAssetRepository{items: map[asset.ID]asset.Asset{item.ID: item}},
		Attachments:        attachments,
		Blobs:              blobStore,
		ImageProcessor:     processor,
		Audit:              &fakeAuditRepository{},
		IDs:                &fakeIDGenerator{ids: []string{"audit-list"}},
		DefaultPageLimit:   50,
		MaxPageLimit:       100,
		MaxAttachmentBytes: 32,
	})

	result, err := application.ListAssets(context.Background(), ListAssetsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-one")},
		Source:      audit.SourceAPI,
		TenantID:    tenantID,
		InventoryID: inventoryID,
	})
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}

	ref := ports.AttachmentAssetReference{InventoryID: inventoryID, AssetID: item.ID}
	if result.PrimaryPhotos[ref].ID != attachment.ID {
		t.Fatalf("expected primary photo summary, got %+v", result.PrimaryPhotos)
	}
	cacheKey, ok := thumbnailStorageKey(attachment, media.ThumbnailVariantSmall)
	if !ok {
		t.Fatalf("expected thumbnail cache key")
	}
	waitFor(t, time.Second, func() bool {
		return processor.thumbnailCallCount() == 1 && blobStore.hasBlob(cacheKey)
	})
	if processor.thumbnailCallCount() != 1 {
		t.Fatalf("expected one warmed thumbnail, got %d", processor.thumbnailCallCount())
	}
}

func TestListAssetsDoesNotWaitForPrimaryThumbnailWarm(t *testing.T) {
	content := pngAttachmentBytes()
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	item := assetItem("asset-one", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	attachment := attachmentForAsset(t, item, "attachment-one", media.ContentTypePNG, content)
	blobStore := &recordingBlobStorage{blobs: map[media.StorageKey][]byte{attachment.StorageKey: content}}
	processor := newBlockingWarmImageProcessor([]byte("small-card-thumb"))
	application := New(Dependencies{
		Observer:   noopObserver{},
		Authorizer: &fakeAuthorizer{},
		Tenants:    &fakeTenantRepository{exists: true},
		Inventories: &fakeInventoryRepository{items: []inventory.Inventory{
			inventoryItem(inventoryID.String(), tenantID.String(), "Home"),
		}},
		Assets:             &fakeAssetRepository{items: map[asset.ID]asset.Asset{item.ID: item}},
		Attachments:        &recordingAttachmentRepository{attachment: attachment, found: true},
		Blobs:              blobStore,
		ImageProcessor:     processor,
		Audit:              &fakeAuditRepository{},
		IDs:                &fakeIDGenerator{ids: []string{"audit-list"}},
		DefaultPageLimit:   50,
		MaxPageLimit:       100,
		MaxAttachmentBytes: 32,
	})

	_, err := application.ListAssets(context.Background(), ListAssetsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-one")},
		Source:      audit.SourceAPI,
		TenantID:    tenantID,
		InventoryID: inventoryID,
	})
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}

	select {
	case <-processor.started:
	case <-time.After(time.Second):
		t.Fatalf("expected thumbnail warm to start")
	}
	cacheKey, ok := thumbnailStorageKey(attachment, media.ThumbnailVariantSmall)
	if !ok {
		t.Fatalf("expected thumbnail cache key")
	}
	if blobStore.hasBlob(cacheKey) {
		t.Fatalf("list assets must not wait for thumbnail warm to finish")
	}
	close(processor.release)
	waitFor(t, time.Second, func() bool { return blobStore.hasBlob(cacheKey) })
}

func TestConcurrentListAssetsWarmsPrimaryThumbnailOnce(t *testing.T) {
	content := pngAttachmentBytes()
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	item := assetItem("asset-one", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	attachment := attachmentForAsset(t, item, "attachment-one", media.ContentTypePNG, content)
	blobStore := &recordingBlobStorage{blobs: map[media.StorageKey][]byte{attachment.StorageKey: content}}
	processor := newBlockingWarmImageProcessor([]byte("small-card-thumb"))
	application := New(Dependencies{
		Observer:       noopObserver{},
		Blobs:          blobStore,
		ImageProcessor: processor,
	})

	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			application.warmPrimarySmallThumbnails(context.Background(), []media.Attachment{attachment})
		}()
	}
	wg.Wait()
	select {
	case <-processor.started:
	case <-time.After(time.Second):
		t.Fatalf("expected thumbnail warm to start")
	}
	if processor.thumbnailCallCount() != 1 {
		t.Fatalf("expected duplicate suppression while warm is in flight, got %d calls", processor.thumbnailCallCount())
	}
	close(processor.release)
	cacheKey, ok := thumbnailStorageKey(attachment, media.ThumbnailVariantSmall)
	if !ok {
		t.Fatalf("expected thumbnail cache key")
	}
	waitFor(t, time.Second, func() bool { return blobStore.hasBlob(cacheKey) })
	if processor.thumbnailCallCount() != 1 {
		t.Fatalf("expected one warmed thumbnail, got %d", processor.thumbnailCallCount())
	}
}

func TestDownloadThumbnailJoinsPrimaryThumbnailWarm(t *testing.T) {
	content := pngAttachmentBytes()
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	item := assetItem("asset-one", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	attachment := attachmentForAsset(t, item, "attachment-one", media.ContentTypePNG, content)
	blobStore := &recordingBlobStorage{blobs: map[media.StorageKey][]byte{attachment.StorageKey: content}}
	processor := newBlockingWarmImageProcessor([]byte("small-card-thumb"))
	application := New(Dependencies{
		Observer:   noopObserver{},
		Authorizer: &fakeAuthorizer{},
		Tenants:    &fakeTenantRepository{exists: true},
		Inventories: &fakeInventoryRepository{items: []inventory.Inventory{
			inventoryItem(inventoryID.String(), tenantID.String(), "Home"),
		}},
		Assets:             &fakeAssetRepository{items: map[asset.ID]asset.Asset{item.ID: item}},
		Attachments:        &recordingAttachmentRepository{attachment: attachment, found: true},
		Blobs:              blobStore,
		ImageProcessor:     processor,
		Audit:              &fakeAuditRepository{},
		IDs:                &fakeIDGenerator{ids: []string{"audit-list", "audit-download"}},
		DefaultPageLimit:   50,
		MaxPageLimit:       100,
		MaxAttachmentBytes: 32,
	})

	_, err := application.ListAssets(context.Background(), ListAssetsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-one")},
		Source:      audit.SourceAPI,
		TenantID:    tenantID,
		InventoryID: inventoryID,
	})
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	select {
	case <-processor.started:
	case <-time.After(time.Second):
		t.Fatalf("expected thumbnail warm to start")
	}

	downloaded := make(chan AttachmentThumbnailResult, 1)
	downloadErr := make(chan error, 1)
	go func() {
		result, err := application.DownloadAttachmentThumbnail(context.Background(), DownloadAttachmentThumbnailInput{
			Principal:    identity.Principal{ID: identity.PrincipalID("user-one")},
			Source:       audit.SourceAPI,
			TenantID:     tenantID,
			InventoryID:  inventoryID,
			AssetID:      item.ID,
			AttachmentID: attachment.ID,
			Variant:      media.ThumbnailVariantSmall.String(),
		})
		if err != nil {
			downloadErr <- err
			return
		}
		downloaded <- result
	}()

	time.Sleep(25 * time.Millisecond)
	if processor.thumbnailCallCount() != 1 {
		t.Fatalf("expected immediate download to join warm generation, got %d generations", processor.thumbnailCallCount())
	}
	close(processor.release)
	select {
	case err := <-downloadErr:
		t.Fatalf("download thumbnail: %v", err)
	case result := <-downloaded:
		if string(result.Content) != "small-card-thumb" {
			t.Fatalf("expected warmed thumbnail content, got %q", string(result.Content))
		}
	case <-time.After(time.Second):
		t.Fatalf("expected thumbnail download to finish after warm generation")
	}
	if processor.thumbnailCallCount() != 1 {
		t.Fatalf("expected one thumbnail generation after download, got %d", processor.thumbnailCallCount())
	}
}

func TestDownloadThumbnailRetriesWhenPrimaryThumbnailWarmTimesOut(t *testing.T) {
	content := pngAttachmentBytes()
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	item := assetItem("asset-one", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	attachment := attachmentForAsset(t, item, "attachment-one", media.ContentTypePNG, content)
	blobStore := &recordingBlobStorage{blobs: map[media.StorageKey][]byte{attachment.StorageKey: content}}
	processor := newBlockingWarmImageProcessor([]byte("small-card-thumb"))
	application := New(Dependencies{
		Observer:   noopObserver{},
		Authorizer: &fakeAuthorizer{},
		Tenants:    &fakeTenantRepository{exists: true},
		Inventories: &fakeInventoryRepository{items: []inventory.Inventory{
			inventoryItem(inventoryID.String(), tenantID.String(), "Home"),
		}},
		Assets:                          &fakeAssetRepository{items: map[asset.ID]asset.Asset{item.ID: item}},
		Attachments:                     &recordingAttachmentRepository{attachment: attachment, found: true},
		Blobs:                           blobStore,
		ImageProcessor:                  processor,
		Audit:                           &fakeAuditRepository{},
		IDs:                             &fakeIDGenerator{ids: []string{"audit-list", "audit-download"}},
		DefaultPageLimit:                50,
		MaxPageLimit:                    100,
		MaxAttachmentBytes:              32,
		PrimaryThumbnailWarmTimeout:     10 * time.Millisecond,
		PrimaryThumbnailWarmLimit:       12,
		PrimaryThumbnailWarmConcurrency: 4,
	})

	_, err := application.ListAssets(context.Background(), ListAssetsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-one")},
		Source:      audit.SourceAPI,
		TenantID:    tenantID,
		InventoryID: inventoryID,
	})
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	select {
	case <-processor.started:
	case <-time.After(time.Second):
		t.Fatalf("expected thumbnail warm to start")
	}

	downloaded := make(chan AttachmentThumbnailResult, 1)
	downloadErr := make(chan error, 1)
	go func() {
		result, err := application.DownloadAttachmentThumbnail(context.Background(), DownloadAttachmentThumbnailInput{
			Principal:    identity.Principal{ID: identity.PrincipalID("user-one")},
			Source:       audit.SourceAPI,
			TenantID:     tenantID,
			InventoryID:  inventoryID,
			AssetID:      item.ID,
			AttachmentID: attachment.ID,
			Variant:      media.ThumbnailVariantSmall.String(),
		})
		if err != nil {
			downloadErr <- err
			return
		}
		downloaded <- result
	}()

	waitFor(t, time.Second, func() bool { return processor.thumbnailCallCount() == 2 })
	close(processor.release)
	select {
	case err := <-downloadErr:
		t.Fatalf("download thumbnail should retry after warm timeout: %v", err)
	case result := <-downloaded:
		if string(result.Content) != "small-card-thumb" {
			t.Fatalf("expected retried thumbnail content, got %q", string(result.Content))
		}
	case <-time.After(time.Second):
		t.Fatalf("expected thumbnail download to finish after retry")
	}
}

func TestSearchAssetsWarmsPrimarySmallThumbnails(t *testing.T) {
	content := pngAttachmentBytes()
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	item := assetItem("asset-one", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	attachment := attachmentForAsset(t, item, "attachment-one", media.ContentTypePNG, content)
	blobStore := &recordingBlobStorage{blobs: map[media.StorageKey][]byte{attachment.StorageKey: content}}
	processor := &warmImageProcessor{thumbnailContent: []byte("small-card-thumb")}
	searchRepo := &recordingAssetSearchRepository{items: []ports.AssetSearchResult{{
		Type:      search.ResultTypeAsset,
		TenantID:  tenantID,
		Inventory: inventoryItem(inventoryID.String(), tenantID.String(), "Home"),
		Asset:     item,
	}}}
	application := New(Dependencies{
		Observer:   noopObserver{},
		Authorizer: &visibilityAuthorizer{t: t, tenantID: tenantID, visible: []inventory.InventoryID{inventoryID}},
		Tenants:    &fakeTenantRepository{exists: true},
		Inventories: &fakeInventoryRepository{items: []inventory.Inventory{
			inventoryItem(inventoryID.String(), tenantID.String(), "Home"),
		}},
		Search:           searchRepo,
		Attachments:      &recordingAttachmentRepository{attachment: attachment, found: true},
		Blobs:            blobStore,
		ImageProcessor:   processor,
		DefaultPageLimit: 50,
		MaxPageLimit:     100,
	})

	_, err := application.SearchAssets(context.Background(), SearchAssetsInput{
		Principal: identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:  tenantID,
		Query:     "water",
		Mode:      "fuzzy",
	})
	if err != nil {
		t.Fatalf("search assets: %v", err)
	}

	cacheKey, ok := thumbnailStorageKey(attachment, media.ThumbnailVariantSmall)
	if !ok {
		t.Fatalf("expected thumbnail cache key")
	}
	waitFor(t, time.Second, func() bool {
		return processor.thumbnailCallCount() == 1 && blobStore.hasBlob(cacheKey)
	})
	if !blobStore.hasBlob(cacheKey) {
		t.Fatalf("expected small thumbnail cache to be warmed")
	}
}

func TestListAssetsIgnoresPrimaryThumbnailWarmFailures(t *testing.T) {
	content := pngAttachmentBytes()
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	item := assetItem("asset-one", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	attachment := attachmentForAsset(t, item, "attachment-one", media.ContentTypePNG, content)
	application := New(Dependencies{
		Observer:   noopObserver{},
		Authorizer: &fakeAuthorizer{},
		Tenants:    &fakeTenantRepository{exists: true},
		Inventories: &fakeInventoryRepository{items: []inventory.Inventory{
			inventoryItem(inventoryID.String(), tenantID.String(), "Home"),
		}},
		Assets:             &fakeAssetRepository{items: map[asset.ID]asset.Asset{item.ID: item}},
		Attachments:        &recordingAttachmentRepository{attachment: attachment, found: true},
		Blobs:              &recordingBlobStorage{blobs: map[media.StorageKey][]byte{attachment.StorageKey: content}, putErr: errors.New("cache unavailable")},
		ImageProcessor:     &warmImageProcessor{thumbnailContent: []byte("small-card-thumb")},
		Audit:              &fakeAuditRepository{},
		IDs:                &fakeIDGenerator{ids: []string{"audit-list"}},
		DefaultPageLimit:   50,
		MaxPageLimit:       100,
		MaxAttachmentBytes: 32,
	})

	result, err := application.ListAssets(context.Background(), ListAssetsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-one")},
		Source:      audit.SourceAPI,
		TenantID:    tenantID,
		InventoryID: inventoryID,
	})
	if err != nil {
		t.Fatalf("list assets should ignore thumbnail warm failure: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected list result despite warm failure, got %+v", result)
	}
}

func TestListAssetsIgnoresPrimaryThumbnailWarmFailuresWithNilObserver(t *testing.T) {
	content := pngAttachmentBytes()
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	item := assetItem("asset-one", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	attachment := attachmentForAsset(t, item, "attachment-one", media.ContentTypePNG, content)
	application := New(Dependencies{
		Authorizer: &fakeAuthorizer{},
		Tenants:    &fakeTenantRepository{exists: true},
		Inventories: &fakeInventoryRepository{items: []inventory.Inventory{
			inventoryItem(inventoryID.String(), tenantID.String(), "Home"),
		}},
		Assets:             &fakeAssetRepository{items: map[asset.ID]asset.Asset{item.ID: item}},
		Attachments:        &recordingAttachmentRepository{attachment: attachment, found: true},
		Blobs:              &recordingBlobStorage{blobs: map[media.StorageKey][]byte{attachment.StorageKey: content}, putErr: errors.New("cache unavailable")},
		ImageProcessor:     &warmImageProcessor{thumbnailContent: []byte("small-card-thumb")},
		Audit:              &fakeAuditRepository{},
		IDs:                &fakeIDGenerator{ids: []string{"audit-list"}},
		DefaultPageLimit:   50,
		MaxPageLimit:       100,
		MaxAttachmentBytes: 32,
	})

	result, err := application.ListAssets(context.Background(), ListAssetsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-one")},
		Source:      audit.SourceAPI,
		TenantID:    tenantID,
		InventoryID: inventoryID,
	})
	if err != nil {
		t.Fatalf("list assets should ignore thumbnail warm failure: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected list result despite warm failure, got %+v", result)
	}
}

func attachmentForAsset(t *testing.T, item asset.Asset, id string, contentType media.ContentType, content []byte) media.Attachment {
	t.Helper()

	attachmentID, ok := media.NewID(id)
	if !ok {
		t.Fatalf("expected valid attachment id")
	}
	storageKey, ok := media.NewStorageKey("test/" + id)
	if !ok {
		t.Fatalf("expected valid storage key")
	}
	fileName, ok := media.NewFileName(id + ".png")
	if !ok {
		t.Fatalf("expected valid file name")
	}
	attachment, ok := media.NewAttachment(
		attachmentID,
		media.TenantID(item.TenantID.String()),
		media.InventoryID(item.InventoryID.String()),
		media.AssetID(item.ID.String()),
		storageKey,
		fileName,
		contentType,
		int64(len(content)),
		sha256Of(content),
		time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC),
	)
	if !ok {
		t.Fatalf("expected valid attachment")
	}
	return attachment
}

type warmImageProcessor struct {
	mu               sync.Mutex
	thumbnailContent []byte
	thumbnailCalls   int
}

func (p *warmImageProcessor) CreateThumbnail(_ context.Context, request ports.ImageDerivativeRequest) (ports.ImageDerivative, error) {
	p.mu.Lock()
	p.thumbnailCalls++
	p.mu.Unlock()
	return ports.ImageDerivative{ContentType: request.ContentType, Content: append([]byte(nil), p.thumbnailContent...)}, nil
}

func (p *warmImageProcessor) PrepareImageForModelUse(_ context.Context, request ports.ModelImageRequest) (ports.ModelImage, error) {
	return ports.ModelImage{ContentType: request.ContentType, Content: request.Content, SizeBytes: int64(len(request.Content)), SHA256: sha256Of(request.Content)}, nil
}

func (p *warmImageProcessor) thumbnailCallCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.thumbnailCalls
}

type blockingWarmImageProcessor struct {
	warmImageProcessor
	startOnce sync.Once
	started   chan struct{}
	release   chan struct{}
}

func newBlockingWarmImageProcessor(content []byte) *blockingWarmImageProcessor {
	return &blockingWarmImageProcessor{
		warmImageProcessor: warmImageProcessor{thumbnailContent: content},
		started:            make(chan struct{}),
		release:            make(chan struct{}),
	}
}

func (p *blockingWarmImageProcessor) CreateThumbnail(ctx context.Context, request ports.ImageDerivativeRequest) (ports.ImageDerivative, error) {
	p.startOnce.Do(func() { close(p.started) })
	p.mu.Lock()
	p.thumbnailCalls++
	p.mu.Unlock()
	select {
	case <-p.release:
	case <-ctx.Done():
		return ports.ImageDerivative{}, ctx.Err()
	}
	return ports.ImageDerivative{ContentType: request.ContentType, Content: append([]byte(nil), p.thumbnailContent...)}, nil
}

func waitFor(t *testing.T, timeout time.Duration, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !condition() {
		t.Fatalf("condition was not met within %s", timeout)
	}
}
