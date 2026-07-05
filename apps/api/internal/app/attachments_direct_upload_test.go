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

func TestCompleteAttachmentDirectUploadPersistsVerifiedMetadata(t *testing.T) {
	repository := &recordingAttachmentRepository{}
	directUploads := &fakeDirectAttachmentUploader{
		completed: ports.CompletedDirectAttachmentUpload{
			UploadID:     "upload-one",
			AttachmentID: media.ID("attachment-one"),
			TenantID:     tenant.ID("tenant-one"),
			InventoryID:  inventory.InventoryID("inventory-one"),
			AssetID:      asset.ID("asset-one"),
			StorageKey:   media.StorageKey("tenant-one/inventory-one/asset-one/attachment-one"),
			FileName:     media.FileName("receipt.png"),
			ContentType:  media.ContentTypePNG,
			SizeBytes:    int64(len(pngAttachmentBytes())),
			SHA256:       sha256Of(pngAttachmentBytes()),
			ExpiresAt:    time.Now().Add(time.Hour),
		},
	}
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
		Blobs:                &recordingBlobStorage{content: pngAttachmentBytes()},
		DirectUploads:        directUploads,
		Audit:                &fakeAuditRepository{},
		IDs:                  &attachmentIDGenerator{ids: []string{"audit-one"}},
		MaxAttachmentBytes:   1024,
	})

	attachment, err := application.CompleteAttachmentDirectUpload(context.Background(), CompleteAttachmentDirectUploadInput{
		Principal:   identity.Principal{ID: "owner"},
		Source:      audit.SourceAPI,
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("asset-one"),
		UploadID:    "upload-one",
	})
	if err != nil {
		t.Fatalf("complete direct upload: %v", err)
	}
	if attachment.ID != "attachment-one" || !repository.saved {
		t.Fatalf("expected verified attachment to be persisted, got %+v saved=%t", attachment, repository.saved)
	}
}

func TestCompleteAttachmentDirectUploadRejectsContentTypeMismatch(t *testing.T) {
	repository := &recordingAttachmentRepository{}
	directUploads := &fakeDirectAttachmentUploader{
		completed: ports.CompletedDirectAttachmentUpload{
			UploadID:     "upload-one",
			AttachmentID: media.ID("attachment-one"),
			TenantID:     tenant.ID("tenant-one"),
			InventoryID:  inventory.InventoryID("inventory-one"),
			AssetID:      asset.ID("asset-one"),
			StorageKey:   media.StorageKey("tenant-one/inventory-one/asset-one/attachment-one"),
			FileName:     media.FileName("receipt.jpg"),
			ContentType:  media.ContentTypeJPEG,
			SizeBytes:    int64(len(pngAttachmentBytes())),
			SHA256:       sha256Of(pngAttachmentBytes()),
			ExpiresAt:    time.Now().Add(time.Hour),
		},
	}
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
		Blobs:                &recordingBlobStorage{content: pngAttachmentBytes()},
		DirectUploads:        directUploads,
		Audit:                &fakeAuditRepository{},
		IDs:                  &attachmentIDGenerator{ids: []string{"audit-one"}},
		MaxAttachmentBytes:   1024,
	})

	_, err := application.CompleteAttachmentDirectUpload(context.Background(), CompleteAttachmentDirectUploadInput{
		Principal:   identity.Principal{ID: "owner"},
		Source:      audit.SourceAPI,
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("asset-one"),
		UploadID:    "upload-one",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input for mismatched direct upload content, got %v", err)
	}
	if repository.saved {
		t.Fatalf("expected mismatched direct upload to avoid metadata persistence")
	}
}

func TestCompleteAttachmentDirectUploadRejectsUndecodableImageContent(t *testing.T) {
	content := truncatedPNGAttachmentBytes()
	repository := &recordingAttachmentRepository{}
	directUploads := &fakeDirectAttachmentUploader{
		completed: ports.CompletedDirectAttachmentUpload{
			UploadID:     "upload-one",
			AttachmentID: media.ID("attachment-one"),
			TenantID:     tenant.ID("tenant-one"),
			InventoryID:  inventory.InventoryID("inventory-one"),
			AssetID:      asset.ID("asset-one"),
			StorageKey:   media.StorageKey("tenant-one/inventory-one/asset-one/attachment-one"),
			FileName:     media.FileName("receipt.png"),
			ContentType:  media.ContentTypePNG,
			SizeBytes:    int64(len(content)),
			SHA256:       sha256Of(content),
			ExpiresAt:    time.Now().Add(time.Hour),
		},
	}
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
		Blobs:                &recordingBlobStorage{content: content},
		DirectUploads:        directUploads,
		Audit:                &fakeAuditRepository{},
		IDs:                  &attachmentIDGenerator{ids: []string{"audit-one"}},
		MaxAttachmentBytes:   1024,
	})

	_, err := application.CompleteAttachmentDirectUpload(context.Background(), CompleteAttachmentDirectUploadInput{
		Principal:   identity.Principal{ID: "owner"},
		Source:      audit.SourceAPI,
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("asset-one"),
		UploadID:    "upload-one",
	})
	if !errors.Is(err, ErrAttachmentContentMismatch) {
		t.Fatalf("expected content mismatch for undecodable image, got %v", err)
	}
	if repository.saved {
		t.Fatalf("expected undecodable direct upload to avoid metadata persistence")
	}
}

func TestCompleteAttachmentDirectUploadRejectsMismatchedScope(t *testing.T) {
	directUploads := &fakeDirectAttachmentUploader{
		completed: ports.CompletedDirectAttachmentUpload{
			UploadID:     "upload-one",
			AttachmentID: media.ID("attachment-one"),
			TenantID:     tenant.ID("other-tenant"),
			InventoryID:  inventory.InventoryID("inventory-one"),
			AssetID:      asset.ID("asset-one"),
			StorageKey:   media.StorageKey("other-tenant/inventory-one/asset-one/attachment-one"),
			FileName:     media.FileName("receipt.png"),
			ContentType:  media.ContentTypePNG,
			SizeBytes:    int64(len(pngAttachmentBytes())),
			SHA256:       sha256Of(pngAttachmentBytes()),
			ExpiresAt:    time.Now().Add(time.Hour),
		},
	}
	application := New(Dependencies{
		Observer:             noopObserver{},
		Authorizer:           allowInventoryAuthorizer{},
		Tenants:              attachmentTenantRepository{},
		TenantUnitOfWork:     attachmentTenantRepository{},
		Inventories:          attachmentInventoryRepository{},
		InventoryUnitOfWork:  attachmentInventoryRepository{},
		Assets:               attachmentAssetRepository{},
		Attachments:          &recordingAttachmentRepository{},
		AttachmentUnitOfWork: &recordingAttachmentRepository{},
		DirectUploads:        directUploads,
		MaxAttachmentBytes:   32,
	})

	_, err := application.CompleteAttachmentDirectUpload(context.Background(), CompleteAttachmentDirectUploadInput{
		Principal:   identity.Principal{ID: "owner"},
		Source:      audit.SourceAPI,
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("asset-one"),
		UploadID:    "upload-one",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestCompleteAttachmentDirectUploadMapsExpectedAdapterFailures(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want error
	}{
		{name: "incomplete", err: ports.ErrDirectUploadIncomplete, want: ErrNotFound},
		{name: "expired", err: ports.ErrDirectUploadExpired, want: ErrInvalidInput},
		{name: "mismatch", err: ports.ErrDirectUploadMismatch, want: ErrInvalidInput},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
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
				DirectUploads:        &fakeDirectAttachmentUploader{err: tc.err},
				MaxAttachmentBytes:   32,
			})

			_, err := application.CompleteAttachmentDirectUpload(context.Background(), CompleteAttachmentDirectUploadInput{
				Principal:   identity.Principal{ID: "owner"},
				Source:      audit.SourceAPI,
				TenantID:    tenant.ID("tenant-one"),
				InventoryID: inventory.InventoryID("inventory-one"),
				AssetID:     asset.ID("asset-one"),
				UploadID:    "upload-one",
			})
			if !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
			if repository.saved {
				t.Fatalf("expected failed direct upload to avoid metadata persistence")
			}
		})
	}
}
