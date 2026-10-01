package search

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func PrimaryPhotosForSearchResults(items []ports.AssetSearchResult, primaryPhotos map[ports.AttachmentAssetReference]media.Attachment) []media.Attachment {
	photos := make([]media.Attachment, 0, len(items))
	for _, item := range items {
		ref := ports.AttachmentAssetReference{
			InventoryID: inventory.InventoryID(item.Asset.InventoryID.String()),
			AssetID:     item.Asset.ID,
		}
		if photo, ok := primaryPhotos[ref]; ok {
			photos = append(photos, photo)
		}
	}
	return photos
}

func PrimaryImageAttachmentsForSearchResults(ctx context.Context, attachments ports.AttachmentRepository, tenantID tenant.ID, items []ports.AssetSearchResult) (map[ports.AttachmentAssetReference]media.Attachment, error) {
	if attachments == nil || len(items) == 0 {
		return nil, nil
	}
	assetRefs := make([]ports.AttachmentAssetReference, 0, len(items))
	for _, item := range items {
		assetRefs = append(assetRefs, ports.AttachmentAssetReference{
			InventoryID: inventory.InventoryID(item.Asset.InventoryID.String()),
			AssetID:     item.Asset.ID,
		})
	}
	return attachments.FirstImageAttachmentsByAssets(ctx, tenantID, assetRefs)
}
