package mapper

import (
	"net/url"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func AssetToResponse(item asset.Asset, primaryPhoto *media.Attachment, currentCheckout *asset.Checkout, checkoutPrincipals map[identity.PrincipalID]identity.User) dto.AssetResponse {
	return AssetToResponseWithTags(item, nil, primaryPhoto, currentCheckout, checkoutPrincipals)
}

func AssetToResponseWithTags(item asset.Asset, tags []assettag.Tag, primaryPhoto *media.Attachment, currentCheckout *asset.Checkout, checkoutPrincipals map[identity.PrincipalID]identity.User) dto.AssetResponse {
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
		Tags:              TagsToResponse(tags),
		LifecycleState:    item.LifecycleState.String(),
		CreatedAt:         item.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:         item.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if primaryPhoto != nil {
		response.PrimaryPhoto = assetPrimaryPhotoToResponse(*primaryPhoto)
	}
	if currentCheckout != nil {
		response.CurrentCheckout = CurrentCheckoutToResponse(*currentCheckout, checkoutPrincipals)
	}
	return response
}

func TagsToResponse(tags []assettag.Tag) []dto.CompactTag {
	if len(tags) == 0 {
		return []dto.CompactTag{}
	}
	data := make([]dto.CompactTag, 0, len(tags))
	for _, tag := range tags {
		data = append(data, dto.CompactTag{
			ID:          tag.ID.String(),
			Key:         tag.Key.String(),
			DisplayName: tag.DisplayName.String(),
			Color:       tag.Color.String(),
		})
	}
	return data
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

func AssetsToResponse(items []asset.Asset, primaryPhotos map[ports.AttachmentAssetReference]media.Attachment, currentCheckouts map[asset.ID]asset.Checkout, checkoutPrincipals map[identity.PrincipalID]identity.User) []dto.AssetResponse {
	return AssetsToResponseWithTags(items, nil, primaryPhotos, currentCheckouts, checkoutPrincipals)
}

func AssetsToResponseWithTags(items []asset.Asset, tags map[asset.ID][]assettag.Tag, primaryPhotos map[ports.AttachmentAssetReference]media.Attachment, currentCheckouts map[asset.ID]asset.Checkout, checkoutPrincipals map[identity.PrincipalID]identity.User) []dto.AssetResponse {
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
		data = append(data, AssetToResponseWithTags(item, tags[item.ID], primaryPhoto, currentCheckout, checkoutPrincipals))
	}
	return data
}

func CurrentCheckoutToResponse(checkout asset.Checkout, checkoutPrincipals map[identity.PrincipalID]identity.User) *dto.CurrentCheckout {
	response := &dto.CurrentCheckout{
		ID:                      checkout.ID.String(),
		State:                   checkout.State.String(),
		CheckedOutAt:            checkout.CheckedOutAt.UTC().Format(time.RFC3339Nano),
		CheckedOutByPrincipalID: checkout.CheckedOutByPrincipal,
	}
	if user, ok := checkoutPrincipals[identity.PrincipalID(checkout.CheckedOutByPrincipal)]; ok {
		response.CheckedOutByPrincipal = &dto.AssetCheckoutPrincipalResponse{
			ID:    user.ID.String(),
			Email: user.Email.String(),
		}
	}
	return response
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

func CheckedOutAssetsToResponse(items []ports.CheckedOutAsset, checkoutPrincipals map[identity.PrincipalID]identity.User) []dto.CheckedOutAssetResponse {
	data := make([]dto.CheckedOutAssetResponse, 0, len(items))
	for _, item := range items {
		data = append(data, dto.CheckedOutAssetResponse{
			Asset:    AssetToResponse(item.Asset, nil, &item.Checkout, checkoutPrincipals),
			Checkout: *CurrentCheckoutToResponse(item.Checkout, checkoutPrincipals),
		})
	}
	return data
}
