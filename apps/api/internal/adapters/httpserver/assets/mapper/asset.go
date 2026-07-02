package mapper

import (
	"net/url"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func AssetToResponse(item asset.Asset, primaryPhoto *media.Attachment) dto.AssetResponse {
	response := dto.AssetResponse{
		ID:                item.ID.String(),
		TenantID:          item.TenantID.String(),
		InventoryID:       item.InventoryID.String(),
		ParentAssetID:     item.ParentAssetID.String(),
		CustomAssetTypeID: item.CustomAssetTypeID.String(),
		Kind:              item.Kind.String(),
		Title:             item.Title.String(),
		Description:       item.Description.String(),
		CustomFields:      item.CustomFields.Values(),
		LifecycleState:    item.LifecycleState.String(),
		CreatedAt:         item.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:         item.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if primaryPhoto != nil {
		response.PrimaryPhoto = assetPrimaryPhotoToResponse(*primaryPhoto)
	}
	return response
}

func assetPrimaryPhotoToResponse(attachment media.Attachment) *dto.AssetPrimaryPhoto {
	return &dto.AssetPrimaryPhoto{
		ID:          attachment.ID.String(),
		FileName:    attachment.FileName.String(),
		ContentType: attachment.ContentType.String(),
		SizeBytes:   attachment.SizeBytes,
		Thumbnails: dto.AssetPhotoThumbnails{
			Small:  assetAttachmentThumbnailPath(attachment, media.ThumbnailVariantSmall),
			Medium: assetAttachmentThumbnailPath(attachment, media.ThumbnailVariantMedium),
			Large:  assetAttachmentThumbnailPath(attachment, media.ThumbnailVariantLarge),
		},
	}
}

func assetAttachmentThumbnailPath(attachment media.Attachment, variant media.ThumbnailVariant) string {
	path := "/tenants/" + url.PathEscape(attachment.TenantID.String()) +
		"/inventories/" + url.PathEscape(attachment.InventoryID.String()) +
		"/assets/" + url.PathEscape(attachment.AssetID.String()) +
		"/attachments/" + url.PathEscape(attachment.ID.String()) +
		"/thumbnail"
	query := url.Values{}
	query.Set("variant", variant.String())
	return path + "?" + query.Encode()
}

func AssetsToResponse(items []asset.Asset, primaryPhotos map[ports.AttachmentAssetReference]media.Attachment) []dto.AssetResponse {
	data := make([]dto.AssetResponse, 0, len(items))
	for _, item := range items {
		var primaryPhoto *media.Attachment
		ref := ports.AttachmentAssetReference{
			InventoryID: inventory.InventoryID(item.InventoryID.String()),
			AssetID:     item.ID,
		}
		if photo, ok := primaryPhotos[ref]; ok {
			primaryPhoto = &photo
		}
		data = append(data, AssetToResponse(item, primaryPhoto))
	}
	return data
}
