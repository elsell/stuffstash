package gormstore

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
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
