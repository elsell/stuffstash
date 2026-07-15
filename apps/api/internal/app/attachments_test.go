package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"image"
	"image/color"
	"image/png"
	"sync"
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

func TestCreateAttachmentDeletesBlobWhenMetadataSaveFails(t *testing.T) {
	blobStore := &recordingBlobStorage{}
	application := New(Dependencies{
		Observer:             noopObserver{},
		Authorizer:           allowInventoryAuthorizer{},
		Tenants:              attachmentTenantRepository{},
		TenantUnitOfWork:     attachmentTenantRepository{},
		Inventories:          attachmentInventoryRepository{},
		InventoryUnitOfWork:  attachmentInventoryRepository{},
		Assets:               attachmentAssetRepository{},
		Attachments:          failingAttachmentRepository{},
		AttachmentUnitOfWork: failingAttachmentRepository{},
		Blobs:                blobStore,
		IDs:                  &attachmentIDGenerator{ids: []string{"attachment-one", "audit-one"}},
		MaxAttachmentBytes:   1024,
	})

	_, err := application.CreateAttachment(context.Background(), CreateAttachmentInput{
		Principal:   identity.Principal{ID: "owner"},
		Source:      audit.SourceAPI,
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("asset-one"),
		FileName:    "receipt.png",
		ContentType: "image/png",
		Content:     pngAttachmentBytes(),
	})
	if !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("expected repository conflict, got %v", err)
	}
	if !blobStore.put || !blobStore.deleted {
		t.Fatalf("expected blob put and compensating delete, got put=%t deleted=%t", blobStore.put, blobStore.deleted)
	}
}

func TestCreateAttachmentRejectsContentTypeMismatch(t *testing.T) {
	application := New(Dependencies{
		Observer:             noopObserver{},
		Authorizer:           allowInventoryAuthorizer{},
		Tenants:              attachmentTenantRepository{},
		TenantUnitOfWork:     attachmentTenantRepository{},
		Inventories:          attachmentInventoryRepository{},
		InventoryUnitOfWork:  attachmentInventoryRepository{},
		Assets:               attachmentAssetRepository{},
		Attachments:          failingAttachmentRepository{},
		AttachmentUnitOfWork: failingAttachmentRepository{},
		Blobs:                &recordingBlobStorage{},
		IDs:                  &attachmentIDGenerator{ids: []string{"attachment-one", "audit-one"}},
		MaxAttachmentBytes:   32,
	})

	_, err := application.CreateAttachment(context.Background(), CreateAttachmentInput{
		Principal:   identity.Principal{ID: "owner"},
		Source:      audit.SourceAPI,
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("asset-one"),
		FileName:    "receipt.png",
		ContentType: "image/png",
		Content:     []byte("not a png"),
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestCreateAttachmentRejectsUndecodableImageContent(t *testing.T) {
	application := New(Dependencies{
		Observer:             noopObserver{},
		Authorizer:           allowInventoryAuthorizer{},
		Tenants:              attachmentTenantRepository{},
		TenantUnitOfWork:     attachmentTenantRepository{},
		Inventories:          attachmentInventoryRepository{},
		InventoryUnitOfWork:  attachmentInventoryRepository{},
		Assets:               attachmentAssetRepository{},
		Attachments:          failingAttachmentRepository{},
		AttachmentUnitOfWork: failingAttachmentRepository{},
		Blobs:                &recordingBlobStorage{},
		IDs:                  &attachmentIDGenerator{ids: []string{"attachment-one", "audit-one"}},
		MaxAttachmentBytes:   1024,
	})

	_, err := application.CreateAttachment(context.Background(), CreateAttachmentInput{
		Principal:   identity.Principal{ID: "owner"},
		Source:      audit.SourceAPI,
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("asset-one"),
		FileName:    "receipt.png",
		ContentType: "image/png",
		Content:     truncatedPNGAttachmentBytes(),
	})
	if !errors.Is(err, ErrAttachmentContentMismatch) {
		t.Fatalf("expected content mismatch for undecodable image, got %v", err)
	}
}

func TestCreateAttachmentAcceptsWebP(t *testing.T) {
	repository := &recordingAttachmentRepository{}
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
		Blobs:                &recordingBlobStorage{},
		Audit:                &fakeAuditRepository{},
		IDs:                  &attachmentIDGenerator{ids: []string{"attachment-one", "audit-one"}},
		MaxAttachmentBytes:   1024,
	})

	attachment, err := application.CreateAttachment(context.Background(), CreateAttachmentInput{
		Principal:   identity.Principal{ID: "owner"},
		Source:      audit.SourceAPI,
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("asset-one"),
		FileName:    "photo.webp",
		ContentType: "image/webp",
		Content:     webPAttachmentBytes(),
	})
	if err != nil {
		t.Fatalf("create webp attachment: %v", err)
	}
	if !repository.saved || attachment.ContentType != media.ContentTypeWEBP {
		t.Fatalf("expected webp attachment to be persisted, got %+v saved=%t", attachment, repository.saved)
	}
}

func TestAttachmentThumbnailAndModelImageUseImageProcessorPort(t *testing.T) {
	content := pngAttachmentBytes()
	attachment := attachmentFixture(t, media.ContentTypePNG, content)
	repository := &recordingAttachmentRepository{attachment: attachment, found: true}
	blobStore := &recordingBlobStorage{blobs: map[media.StorageKey][]byte{attachment.StorageKey: content}}
	processor := &recordingImageProcessor{thumbnailContent: []byte("thumb"), modelContent: []byte("model")}
	application := New(Dependencies{
		Observer:             noopObserver{},
		Authorizer:           allowInventoryAuthorizer{},
		Tenants:              attachmentTenantRepository{},
		TenantUnitOfWork:     attachmentTenantRepository{},
		Inventories:          attachmentInventoryRepository{},
		InventoryUnitOfWork:  attachmentInventoryRepository{},
		Assets:               archivedAttachmentAssetRepository{},
		Attachments:          repository,
		AttachmentUnitOfWork: repository,
		Blobs:                blobStore,
		ImageProcessor:       processor,
		Audit:                &fakeAuditRepository{},
		IDs:                  &attachmentIDGenerator{ids: []string{"audit-thumb", "audit-model"}},
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
	if !processor.thumbnailCalled || string(thumbnail.Content) != "thumb" {
		t.Fatalf("expected thumbnail processor output, got %+v called=%t", thumbnail, processor.thumbnailCalled)
	}

	modelImage, err := application.PrepareAttachmentForModelUse(context.Background(), PrepareAttachmentForModelUseInput{
		Principal:    identity.Principal{ID: "viewer"},
		Source:       audit.SourceAPI,
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		AssetID:      asset.ID("asset-one"),
		AttachmentID: attachment.ID,
	})
	if err != nil {
		t.Fatalf("prepare model image: %v", err)
	}
	if !processor.modelCalled || string(modelImage.Content) != "model" || modelImage.SizeBytes != int64(len("model")) {
		t.Fatalf("expected model image processor output, got %+v called=%t", modelImage, processor.modelCalled)
	}
}

func TestAttachmentThumbnailRejectsNonImageAttachments(t *testing.T) {
	content := []byte("%PDF-1")
	attachment := attachmentFixture(t, media.ContentTypePDF, content)
	application := New(Dependencies{
		Observer:            noopObserver{},
		Authorizer:          allowInventoryAuthorizer{},
		Tenants:             attachmentTenantRepository{},
		TenantUnitOfWork:    attachmentTenantRepository{},
		Inventories:         attachmentInventoryRepository{},
		InventoryUnitOfWork: attachmentInventoryRepository{},
		Assets:              attachmentAssetRepository{},
		Attachments:         &recordingAttachmentRepository{attachment: attachment, found: true},
		Blobs:               &recordingBlobStorage{content: content},
		ImageProcessor:      &recordingImageProcessor{thumbnailContent: []byte("thumb")},
		Audit:               &fakeAuditRepository{},
		IDs:                 &attachmentIDGenerator{ids: []string{"audit-thumb"}},
	})

	_, err := application.DownloadAttachmentThumbnail(context.Background(), DownloadAttachmentThumbnailInput{
		Principal:    identity.Principal{ID: "viewer"},
		Source:       audit.SourceAPI,
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		AssetID:      asset.ID("asset-one"),
		AttachmentID: attachment.ID,
		Variant:      "small",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestDrainBlobDeletionOutboxDeadLettersAfterMaxAttempts(t *testing.T) {
	observer := &fakeObserver{}
	storageKey, ok := media.NewStorageKey("tenant/inventory/blob")
	if !ok {
		t.Fatalf("expected storage key")
	}
	outbox := &singleBlobDeletionOutbox{
		event: ports.BlobDeletionEvent{
			ID:         "blob-event-one",
			StorageKey: storageKey,
			Attempts:   1,
		},
	}
	application := New(Dependencies{
		Observer:                      observer,
		Blobs:                         failingBlobStorage{},
		BlobDeletionOutbox:            outbox,
		IDs:                           &attachmentIDGenerator{ids: []string{"claim-one"}},
		Clock:                         fakeClock{now: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)},
		BlobDeletionOutboxMaxAttempts: 2,
	})

	if err := application.DrainBlobDeletionOutbox(context.Background(), 1); err != nil {
		t.Fatalf("drain blob deletion outbox: %v", err)
	}

	if !outbox.deadLettered {
		t.Fatalf("expected blob deletion event to be dead-lettered")
	}
	if !observer.hasEvent(ports.EventBlobDeletionOutboxFailed) || !observer.hasEvent(ports.EventBlobDeletionOutboxDeadLettered) {
		t.Fatalf("expected failed and dead-lettered events, got %+v", observer.events)
	}
}

type noopObserver struct{}

func (noopObserver) Record(context.Context, ports.Event) {}

type allowInventoryAuthorizer struct{}

func (allowInventoryAuthorizer) CheckTenant(context.Context, identity.Principal, ports.TenantPermission, tenant.ID) error {
	return nil
}

func (allowInventoryAuthorizer) CheckInventory(context.Context, identity.Principal, ports.InventoryPermission, inventory.InventoryID) error {
	return nil
}

func (allowInventoryAuthorizer) ListViewableInventoryIDs(_ context.Context, _ identity.Principal, _ tenant.ID, candidates []inventory.InventoryID) ([]inventory.InventoryID, error) {
	return append([]inventory.InventoryID{}, candidates...), nil
}

func (allowInventoryAuthorizer) GrantTenantOwner(context.Context, identity.Principal, tenant.ID) error {
	return nil
}

func (allowInventoryAuthorizer) GrantInventoryOwner(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (allowInventoryAuthorizer) GrantInventoryViewer(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (allowInventoryAuthorizer) GrantInventoryEditor(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (allowInventoryAuthorizer) RevokeInventoryViewer(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (allowInventoryAuthorizer) RevokeInventoryEditor(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

type attachmentAssetRepository struct{}

type archivedAttachmentAssetRepository struct{ attachmentAssetRepository }

type attachmentTenantRepository struct{}

func (attachmentTenantRepository) SaveTenant(context.Context, tenant.Tenant) error {
	return nil
}

func (attachmentTenantRepository) TenantExists(context.Context, tenant.ID) (bool, error) {
	return true, nil
}

func (attachmentTenantRepository) TenantByID(_ context.Context, tenantID tenant.ID) (tenant.Tenant, bool, error) {
	name, _ := tenant.NewName("Tenant")
	return tenant.Tenant{ID: tenantID, Name: name, LifecycleState: tenant.LifecycleStateActive}, true, nil
}

func (attachmentTenantRepository) ListTenants(context.Context, ports.TenantListPageRequest) ([]tenant.Tenant, error) {
	return nil, nil
}

func (attachmentTenantRepository) UpdateTenant(context.Context, tenant.Tenant, audit.Record) error {
	return nil
}

func (attachmentTenantRepository) UpdateTenantLifecycle(context.Context, tenant.Tenant, audit.Record) error {
	return nil
}

func (attachmentTenantRepository) DeleteTenant(context.Context, tenant.ID, audit.Record) error {
	return nil
}

type attachmentInventoryRepository struct{}

func (attachmentInventoryRepository) SaveInventory(context.Context, inventory.Inventory) error {
	return nil
}

func (attachmentInventoryRepository) InventoryByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (inventory.Inventory, bool, error) {
	return inventory.Inventory{
		ID:       inventoryID,
		TenantID: inventory.TenantID(tenantID.String()),
	}, true, nil
}

func (attachmentInventoryRepository) UpdateInventory(context.Context, inventory.Inventory, audit.Record) error {
	return nil
}

func (attachmentInventoryRepository) UpdateInventoryLifecycle(context.Context, inventory.Inventory, audit.Record) error {
	return nil
}

func (attachmentInventoryRepository) DeleteInventory(context.Context, tenant.ID, inventory.InventoryID, audit.Record) error {
	return nil
}

func (attachmentInventoryRepository) InventoryHasActiveAssets(context.Context, tenant.ID, inventory.InventoryID) (bool, error) {
	return false, nil
}

func (attachmentInventoryRepository) ListInventoriesByTenant(context.Context, inventory.TenantID, ports.InventoryListPageRequest) ([]inventory.Inventory, error) {
	return nil, nil
}

func (attachmentInventoryRepository) SaveInventoryAccessGrantAndEnqueue(context.Context, string, ports.InventoryAccessGrant, audit.Record) error {
	return nil
}

func (attachmentInventoryRepository) DeleteInventoryAccessGrantAndClaimRevoke(context.Context, string, string, time.Time, ports.InventoryAccessGrant, audit.Record) (ports.AuthorizationOutboxEvent, bool, error) {
	return ports.AuthorizationOutboxEvent{}, false, nil
}

func (attachmentInventoryRepository) InventoryAccessGrantByID(context.Context, tenant.ID, inventory.InventoryID, identity.PrincipalID, ports.InventoryAccessRelationship) (ports.InventoryAccessGrant, bool, error) {
	return ports.InventoryAccessGrant{}, false, nil
}

func (attachmentInventoryRepository) SaveInventoryAccessInvitation(context.Context, ports.InventoryAccessInvitation, audit.Record) (ports.InventoryAccessInvitation, error) {
	return ports.InventoryAccessInvitation{}, nil
}

func (attachmentInventoryRepository) AcceptInventoryAccessInvitationAndEnqueue(context.Context, tenant.ID, inventory.InventoryID, string, string, identity.Principal, string, audit.Record) (ports.InventoryAccessInvitation, ports.InventoryAccessGrant, error) {
	return ports.InventoryAccessInvitation{}, ports.InventoryAccessGrant{}, nil
}

func (attachmentInventoryRepository) RevokeInventoryAccessInvitation(context.Context, tenant.ID, inventory.InventoryID, string, audit.Record) (bool, error) {
	return false, nil
}

func (attachmentInventoryRepository) InventoryAccessInvitationByID(context.Context, tenant.ID, inventory.InventoryID, string) (ports.InventoryAccessInvitation, bool, error) {
	return ports.InventoryAccessInvitation{}, false, nil
}

func (attachmentInventoryRepository) CancelInventoryAccessInvitation(context.Context, tenant.ID, inventory.InventoryID, string, audit.Record) (bool, error) {
	return false, nil
}

func (attachmentInventoryRepository) DeleteInventoryAccessInvitation(context.Context, tenant.ID, inventory.InventoryID, string, audit.Record) (bool, error) {
	return false, nil
}

func (attachmentInventoryRepository) ListInventoryAccessGrants(context.Context, tenant.ID, inventory.InventoryID, ports.InventoryAccessGrantPageRequest) ([]ports.InventoryAccessGrant, error) {
	return nil, nil
}

func (attachmentAssetRepository) CreateAsset(context.Context, asset.Asset, audit.Record, *ports.UndoableOperation) error {
	return nil
}

func (attachmentAssetRepository) UpdateAsset(context.Context, asset.Asset, []audit.Record, *ports.UndoableOperation) error {
	return nil
}

func (attachmentAssetRepository) UpdateAssetLifecycle(context.Context, asset.Asset, audit.Record, *ports.UndoableOperation) error {
	return nil
}

func (attachmentAssetRepository) DeleteAsset(context.Context, tenant.ID, inventory.InventoryID, asset.ID, audit.Record) error {
	return nil
}

func (attachmentAssetRepository) AssetByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (asset.Asset, bool, error) {
	return asset.Asset{
		ID:             assetID,
		TenantID:       asset.TenantID(tenantID.String()),
		InventoryID:    asset.InventoryID(inventoryID.String()),
		LifecycleState: asset.LifecycleStateActive,
	}, true, nil
}

func (archivedAttachmentAssetRepository) AssetByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (asset.Asset, bool, error) {
	return asset.Asset{
		ID:             assetID,
		TenantID:       asset.TenantID(tenantID.String()),
		InventoryID:    asset.InventoryID(inventoryID.String()),
		LifecycleState: asset.LifecycleStateArchived,
	}, true, nil
}

func (attachmentAssetRepository) AssetHasActiveChildren(context.Context, tenant.ID, inventory.InventoryID, asset.ID) (bool, error) {
	return false, nil
}

func (attachmentAssetRepository) ListAssetsByInventory(context.Context, tenant.ID, inventory.InventoryID, ports.AssetListPageRequest) ([]asset.Asset, error) {
	return nil, nil
}

type failingAttachmentRepository struct{}

func (failingAttachmentRepository) SaveAttachment(context.Context, media.Attachment, audit.Record) error {
	return ports.ErrConflict
}

func (failingAttachmentRepository) UpdateAttachmentLifecycle(context.Context, media.Attachment, audit.Record) error {
	return ports.ErrConflict
}

func (failingAttachmentRepository) DeleteAttachmentAndEnqueueBlobDeletion(context.Context, string, tenant.ID, inventory.InventoryID, asset.ID, media.ID, audit.Record) (media.Attachment, bool, error) {
	return media.Attachment{}, false, ports.ErrConflict
}

func (failingAttachmentRepository) AttachmentByID(context.Context, tenant.ID, inventory.InventoryID, asset.ID, media.ID) (media.Attachment, bool, error) {
	return media.Attachment{}, false, nil
}

func (failingAttachmentRepository) ListAttachmentsByAsset(context.Context, tenant.ID, inventory.InventoryID, asset.ID, ports.AttachmentListPageRequest) ([]media.Attachment, error) {
	return nil, nil
}

func (failingAttachmentRepository) FirstImageAttachmentsByAssets(context.Context, tenant.ID, []ports.AttachmentAssetReference) (map[ports.AttachmentAssetReference]media.Attachment, error) {
	return nil, nil
}

type recordingAttachmentRepository struct {
	attachment media.Attachment
	found      bool
	saved      bool
}

func (r *recordingAttachmentRepository) SaveAttachment(_ context.Context, attachment media.Attachment, _ audit.Record) error {
	r.attachment = attachment
	r.found = true
	r.saved = true
	return nil
}

func (r *recordingAttachmentRepository) UpdateAttachmentLifecycle(context.Context, media.Attachment, audit.Record) error {
	return nil
}

func (r *recordingAttachmentRepository) DeleteAttachmentAndEnqueueBlobDeletion(context.Context, string, tenant.ID, inventory.InventoryID, asset.ID, media.ID, audit.Record) (media.Attachment, bool, error) {
	return media.Attachment{}, false, nil
}

func (r *recordingAttachmentRepository) AttachmentByID(context.Context, tenant.ID, inventory.InventoryID, asset.ID, media.ID) (media.Attachment, bool, error) {
	return r.attachment, r.found, nil
}

func (r *recordingAttachmentRepository) ListAttachmentsByAsset(context.Context, tenant.ID, inventory.InventoryID, asset.ID, ports.AttachmentListPageRequest) ([]media.Attachment, error) {
	if !r.found {
		return nil, nil
	}
	return []media.Attachment{r.attachment}, nil
}

func (r *recordingAttachmentRepository) FirstImageAttachmentsByAssets(context.Context, tenant.ID, []ports.AttachmentAssetReference) (map[ports.AttachmentAssetReference]media.Attachment, error) {
	if !r.found {
		return nil, nil
	}
	return map[ports.AttachmentAssetReference]media.Attachment{{
		InventoryID: inventory.InventoryID(r.attachment.InventoryID.String()),
		AssetID:     asset.ID(r.attachment.AssetID.String()),
	}: r.attachment}, nil
}

type recordingBlobStorage struct {
	mu         sync.Mutex
	put        bool
	deleted    bool
	content    []byte
	blobs      map[media.StorageKey][]byte
	putKeys    []media.StorageKey
	getKeys    []media.StorageKey
	deleteKeys []media.StorageKey
	putErr     error
	deleteErrs map[media.StorageKey]error
}

type failingBlobStorage struct{}

func (failingBlobStorage) PutBlob(context.Context, media.StorageKey, media.ContentType, []byte) error {
	return errors.New("storage unavailable")
}

func (failingBlobStorage) GetBlob(context.Context, media.StorageKey) ([]byte, error) {
	return nil, errors.New("storage unavailable")
}

func (failingBlobStorage) DeleteBlob(context.Context, media.StorageKey) error {
	return errors.New("storage unavailable")
}

type singleBlobDeletionOutbox struct {
	event        ports.BlobDeletionEvent
	processed    bool
	deadLettered bool
	failed       bool
}

func (s *singleBlobDeletionOutbox) ClaimPendingBlobDeletionEvents(_ context.Context, claimID string, _ int, _ time.Time, leaseUntil time.Time) ([]ports.BlobDeletionEvent, error) {
	s.event.ClaimID = claimID
	s.event.ClaimedUntil = leaseUntil
	return []ports.BlobDeletionEvent{s.event}, nil
}

func (s *singleBlobDeletionOutbox) MarkBlobDeletionEventProcessed(context.Context, string, string) error {
	s.processed = true
	return nil
}

func (s *singleBlobDeletionOutbox) MarkBlobDeletionEventFailed(_ context.Context, _ string, _ string, reason string) error {
	s.failed = true
	s.event.LastError = reason
	return nil
}

func (s *singleBlobDeletionOutbox) MarkBlobDeletionEventDeadLettered(_ context.Context, _ string, _ string, reason string) error {
	s.deadLettered = true
	s.event.DeadLetterReason = reason
	return nil
}

func (r *recordingBlobStorage) PutBlob(_ context.Context, key media.StorageKey, _ media.ContentType, data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.put = true
	r.putKeys = append(r.putKeys, key)
	if r.putErr != nil {
		return r.putErr
	}
	if r.blobs == nil {
		r.blobs = map[media.StorageKey][]byte{}
	}
	r.blobs[key] = append([]byte(nil), data...)
	return nil
}

func (r *recordingBlobStorage) GetBlob(_ context.Context, key media.StorageKey) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.getKeys = append(r.getKeys, key)
	if r.blobs != nil {
		if data, ok := r.blobs[key]; ok {
			return append([]byte(nil), data...), nil
		}
		return nil, ports.ErrBlobNotFound
	}
	if r.content != nil {
		return append([]byte(nil), r.content...), nil
	}
	return nil, ports.ErrBlobNotFound
}

func (r *recordingBlobStorage) DeleteBlob(_ context.Context, key media.StorageKey) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deleted = true
	r.deleteKeys = append(r.deleteKeys, key)
	if err := r.deleteErrs[key]; err != nil {
		return err
	}
	if r.blobs != nil {
		delete(r.blobs, key)
	}
	return nil
}

func (r *recordingBlobStorage) getCount(key media.StorageKey) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, candidate := range r.getKeys {
		if candidate == key {
			count++
		}
	}
	return count
}

func (r *recordingBlobStorage) hasBlob(key media.StorageKey) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.blobs == nil {
		return false
	}
	_, ok := r.blobs[key]
	return ok
}

func (r *recordingBlobStorage) deletedKey(key media.StorageKey) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, candidate := range r.deleteKeys {
		if candidate == key {
			return true
		}
	}
	return false
}

type attachmentIDGenerator struct {
	ids []string
}

func (g *attachmentIDGenerator) NewID() string {
	id := g.ids[0]
	g.ids = g.ids[1:]
	return id
}

type fakeDirectAttachmentUploader struct {
	request   ports.DirectAttachmentUploadRequest
	completed ports.CompletedDirectAttachmentUpload
	err       error
}

func (f *fakeDirectAttachmentUploader) CreateDirectAttachmentUpload(_ context.Context, request ports.DirectAttachmentUploadRequest) (ports.DirectAttachmentUpload, error) {
	f.request = request
	return ports.DirectAttachmentUpload{
		UploadID:     request.UploadID,
		AttachmentID: request.AttachmentID,
		Method:       "PUT",
		URL:          "https://uploads.example.test/" + request.UploadID,
		Headers:      map[string]string{"content-type": request.ContentType.String()},
		ExpiresAt:    request.ExpiresAt,
	}, nil
}

func (f *fakeDirectAttachmentUploader) CompleteDirectAttachmentUpload(context.Context, string) (ports.CompletedDirectAttachmentUpload, error) {
	if f.err != nil {
		return ports.CompletedDirectAttachmentUpload{}, f.err
	}
	return f.completed, nil
}

type recordingImageProcessor struct {
	thumbnailCalled  bool
	thumbnailCalls   int
	modelCalled      bool
	thumbnailContent []byte
	modelContent     []byte
}

func (r *recordingImageProcessor) CreateThumbnail(_ context.Context, request ports.ImageDerivativeRequest) (ports.ImageDerivative, error) {
	r.thumbnailCalled = true
	r.thumbnailCalls++
	return ports.ImageDerivative{ContentType: request.ContentType, Content: append([]byte(nil), r.thumbnailContent...)}, nil
}

func (r *recordingImageProcessor) PrepareImageForModelUse(_ context.Context, request ports.ModelImageRequest) (ports.ModelImage, error) {
	r.modelCalled = true
	hash := sha256Of(r.modelContent)
	return ports.ModelImage{
		ContentType: request.ContentType,
		Content:     append([]byte(nil), r.modelContent...),
		SizeBytes:   int64(len(r.modelContent)),
		SHA256:      hash,
		Width:       1,
		Height:      1,
	}, nil
}

func pngAttachmentBytes() []byte {
	source := image.NewRGBA(image.Rect(0, 0, 1, 1))
	source.Set(0, 0, color.RGBA{R: 0x2e, G: 0x7d, B: 0x32, A: 0xff})
	var output bytes.Buffer
	if err := png.Encode(&output, source); err != nil {
		panic("encode png test fixture")
	}
	return output.Bytes()
}

func truncatedPNGAttachmentBytes() []byte {
	content := pngAttachmentBytes()
	return content[:len(content)-8]
}

func webPAttachmentBytes() []byte {
	content, err := base64.StdEncoding.DecodeString("UklGRrIBAABXRUJQVlA4TKUBAAAvSsAYAA8w//M///MfeJAkbXvaSG7m8Q3GfYSBJekwQztm/IcZlgwnmWImn2BK7aFmBtnVir6q//8VOkFE/xm4baTIu8c48ArEo6+B3zFKYln3pqClSCKX0begFTAXFOLXHSyF8cCNcZEG4OywuA4KVVfJCiArU7GAgJI8+lJP/OKMT/fBAjevg1cYB7YVkFuWga2lyPi5I0HFy5YTpWIHg0RZpkniRVW9odHAKOwosWuOGdxIyn2OvaCDvhg/we6TwadPBPbqBV58MsLmMJ8yZnOWk8SRz4N+QoyPL+MnamzMvcE1rHNEr91F9GKZPVUcS9w7PhhH36suB9qPeYb/oLk6cuTiJ0wOK3m5h1cKjW6EVZCYMK7dxcKCBdgP9HkKr9gkAO2P8GKZGWVdIAatQa+1IDpt6qyorVwdy01xdW8Jkfk6xjEXmVQQ+HQdFr6OKhIN34dXWq0+0qr6EJSCeeVLH9+gvGTLyqM65PQ44ihzlTXxQKjKbAvshXgir7Lil9w4L2bvMycmjQcqXaMCO6BlY28i+FOLzbfI1vEqxAhotocAAA==")
	if err != nil {
		panic("invalid webp test fixture")
	}
	return content
}

func sha256Of(content []byte) media.SHA256 {
	hashBytes := sha256.Sum256(content)
	hash, ok := media.NewSHA256(hex.EncodeToString(hashBytes[:]))
	if !ok {
		panic("invalid test hash")
	}
	return hash
}

func attachmentFixture(t *testing.T, contentType media.ContentType, content []byte) media.Attachment {
	t.Helper()
	attachment, ok := media.NewAttachment(
		media.ID("attachment-one"),
		media.TenantID("tenant-one"),
		media.InventoryID("inventory-one"),
		media.AssetID("asset-one"),
		media.StorageKey("tenant-one/inventory-one/asset-one/attachment-one"),
		media.FileName("receipt.png"),
		contentType,
		int64(len(content)),
		sha256Of(content),
		time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	)
	if !ok {
		t.Fatalf("expected attachment fixture")
	}
	return attachment
}
