package gormstore

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestStoreFindsFirstImageAttachmentsByAssetReferences(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	otherInventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAX")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")
	saveInventory(t, ctx, store, otherInventoryID.String(), tenantID, "Supplies")

	item := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	if err := createAsset(t, ctx, store, item); err != nil {
		t.Fatalf("save item asset: %v", err)
	}
	otherInventoryItem := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAZ", tenantID.String(), otherInventoryID.String(), asset.KindItem, "")
	if err := createAsset(t, ctx, store, otherInventoryItem); err != nil {
		t.Fatalf("save other inventory item asset: %v", err)
	}

	saveSearchAttachment(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FB0", item, "manual.pdf", media.ContentTypePDF)
	saveSearchAttachment(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FB1", item, "front.jpg", media.ContentTypeJPEG)
	saveSearchAttachment(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FB2", item, "back.jpg", media.ContentTypeJPEG)
	saveSearchAttachment(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FB3", otherInventoryItem, "other.jpg", media.ContentTypeJPEG)

	ref := ports.AttachmentAssetReference{InventoryID: inventoryID, AssetID: item.ID}
	otherRef := ports.AttachmentAssetReference{InventoryID: otherInventoryID, AssetID: otherInventoryItem.ID}
	result, err := store.FirstImageAttachmentsByAssets(ctx, tenantID, []ports.AttachmentAssetReference{ref, otherRef})
	if err != nil {
		t.Fatalf("find first image attachments: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected one primary image per requested asset reference, got %+v", result)
	}
	if result[ref].ID != media.ID("01ARZ3NDEKTSV4RRFFQ69G5FB1") || result[ref].FileName != media.FileName("front.jpg") {
		t.Fatalf("expected first image attachment for item, got %+v", result[ref])
	}
	if result[otherRef].ID != media.ID("01ARZ3NDEKTSV4RRFFQ69G5FB3") {
		t.Fatalf("expected scoped other-inventory image attachment, got %+v", result[otherRef])
	}
}

func TestSaveAttachmentAdvancesOwningAssetUpdatedAt(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FC0")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FC1")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	createdAt := time.Date(2026, 7, 7, 10, 0, 0, 0, time.UTC)
	attachmentCreatedAt := time.Date(2026, 7, 8, 11, 0, 0, 0, time.UTC)
	item := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FC2", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	item.CreatedAt = createdAt
	item.UpdatedAt = createdAt
	if err := createAsset(t, ctx, store, item); err != nil {
		t.Fatalf("save item asset: %v", err)
	}
	attachment := testAttachment(t, "01ARZ3NDEKTSV4RRFFQ69G5FC3", item, "photo.jpg", media.ContentTypeJPEG, attachmentCreatedAt)

	if err := store.SaveAttachment(ctx, attachment, auditRecord(t, "audit-photo-create", tenantID, inventoryID, audit.ActionAttachmentCreated)); err != nil {
		t.Fatalf("save attachment: %v", err)
	}
	updated, found, err := store.AssetByID(ctx, tenantID, inventoryID, item.ID)
	if err != nil {
		t.Fatalf("load updated asset: %v", err)
	}
	if !found {
		t.Fatalf("expected updated asset to exist")
	}
	if !updated.UpdatedAt.Equal(attachmentCreatedAt) {
		t.Fatalf("expected attachment save to update asset recency to %s, got %s", attachmentCreatedAt, updated.UpdatedAt)
	}
}

func testAttachment(t *testing.T, id string, item asset.Asset, fileNameValue string, contentType media.ContentType, createdAt time.Time) media.Attachment {
	t.Helper()

	attachmentID, ok := media.NewID(id)
	if !ok {
		t.Fatalf("expected valid attachment id")
	}
	storageKey, ok := media.NewStorageKey("test/" + id)
	if !ok {
		t.Fatalf("expected valid storage key")
	}
	fileName, ok := media.NewFileName(fileNameValue)
	if !ok {
		t.Fatalf("expected valid file name")
	}
	hash, ok := media.NewSHA256("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	if !ok {
		t.Fatalf("expected valid sha256")
	}
	attachment, ok := media.NewAttachment(
		attachmentID,
		media.TenantID(item.TenantID.String()),
		media.InventoryID(item.InventoryID.String()),
		media.AssetID(item.ID.String()),
		storageKey,
		fileName,
		contentType,
		12,
		hash,
		createdAt,
	)
	if !ok {
		t.Fatalf("expected valid attachment")
	}
	return attachment
}
