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

func AssetToResponse(item asset.Asset, primaryPhoto *media.Attachment, currentCheckout *asset.Checkout) dto.AssetResponse {
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
	if currentCheckout != nil {
		response.CurrentCheckout = CurrentCheckoutToResponse(*currentCheckout)
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

func AssetsToResponse(items []asset.Asset, primaryPhotos map[ports.AttachmentAssetReference]media.Attachment, currentCheckouts map[asset.ID]asset.Checkout) []dto.AssetResponse {
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
		var currentCheckout *asset.Checkout
		if checkout, ok := currentCheckouts[item.ID]; ok {
			currentCheckout = &checkout
		}
		data = append(data, AssetToResponse(item, primaryPhoto, currentCheckout))
	}
	return data
}

func CurrentCheckoutToResponse(checkout asset.Checkout) *dto.CurrentCheckout {
	return &dto.CurrentCheckout{
		ID:                      checkout.ID.String(),
		CheckedOutAt:            checkout.CheckedOutAt.UTC().Format(time.RFC3339Nano),
		CheckedOutByPrincipalID: checkout.CheckedOutByPrincipal,
	}
}

func CheckoutToResponse(checkout asset.Checkout) dto.AssetCheckoutResponse {
	response := dto.AssetCheckoutResponse{
		ID:                      checkout.ID.String(),
		TenantID:                checkout.TenantID.String(),
		InventoryID:             checkout.InventoryID.String(),
		AssetID:                 checkout.AssetID.String(),
		State:                   checkout.State.String(),
		CheckedOutAt:            checkout.CheckedOutAt.UTC().Format(time.RFC3339Nano),
		CheckedOutByPrincipalID: checkout.CheckedOutByPrincipal,
		CheckoutDetails:         checkout.CheckoutDetails.String(),
		CreatedAt:               checkout.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:               checkout.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if !checkout.ReturnedAt.IsZero() {
		response.ReturnedAt = checkout.ReturnedAt.UTC().Format(time.RFC3339Nano)
	}
	response.ReturnedByPrincipalID = checkout.ReturnedByPrincipal
	response.ReturnDetails = checkout.ReturnDetails.String()
	return response
}

func CheckoutsToResponse(checkouts []asset.Checkout) []dto.AssetCheckoutResponse {
	data := make([]dto.AssetCheckoutResponse, 0, len(checkouts))
	for _, checkout := range checkouts {
		data = append(data, CheckoutToResponse(checkout))
	}
	return data
}

func CheckedOutAssetsToResponse(items []ports.CheckedOutAsset) []dto.CheckedOutAssetResponse {
	data := make([]dto.CheckedOutAssetResponse, 0, len(items))
	for _, item := range items {
		data = append(data, dto.CheckedOutAssetResponse{
			Asset:    AssetToResponse(item.Asset, nil, &item.Checkout),
			Checkout: *CurrentCheckoutToResponse(item.Checkout),
		})
	}
	return data
}
