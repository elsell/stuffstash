package mapper

import (
	"net/url"
	"time"

	assetmapper "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/search/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func AssetSearchResultsToResponse(results []ports.AssetSearchResult, primaryPhotos map[ports.AttachmentAssetReference]media.Attachment, checkoutPrincipals map[identity.PrincipalID]identity.User) []dto.AssetSearchResultResponse {
	data := make([]dto.AssetSearchResultResponse, 0, len(results))
	for _, result := range results {
		var primaryPhoto *media.Attachment
		ref := ports.AttachmentAssetReference{
			InventoryID: inventory.InventoryID(result.Asset.InventoryID.String()),
			AssetID:     result.Asset.ID,
		}
		if photo, ok := primaryPhotos[ref]; ok {
			primaryPhoto = &photo
		}
		data = append(data, AssetSearchResultToResponse(result, primaryPhoto, checkoutPrincipals))
	}
	return data
}

func AssetSearchResultToResponse(result ports.AssetSearchResult, primaryPhoto *media.Attachment, checkoutPrincipals map[identity.PrincipalID]identity.User) dto.AssetSearchResultResponse {
	assetSummary := dto.AssetSummary{
		ID:                result.Asset.ID.String(),
		InventoryID:       result.Asset.InventoryID.String(),
		ParentAssetID:     result.Asset.ParentAssetID.String(),
		CustomAssetTypeID: result.Asset.CustomAssetTypeID.String(),
		Kind:              result.Asset.Kind.String(),
		Title:             result.Asset.Title.String(),
		Description:       result.Asset.Description.String(),
		CustomFields:      result.Asset.CustomFields.Values(),
		Tags:              assetmapper.TagsToResponse(result.AssignedTags),
		LifecycleState:    result.Asset.LifecycleState.String(),
		CreatedAt:         result.Asset.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:         result.Asset.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if primaryPhoto != nil {
		assetSummary.PrimaryPhoto = assetPrimaryPhotoToResponse(*primaryPhoto)
	}
	if result.CurrentCheckout != nil {
		assetSummary.CurrentCheckout = &dto.SearchCurrentCheckout{
			ID:                      result.CurrentCheckout.ID.String(),
			State:                   result.CurrentCheckout.State.String(),
			CheckedOutAt:            result.CurrentCheckout.CheckedOutAt.UTC().Format(time.RFC3339Nano),
			CheckedOutByPrincipalID: result.CurrentCheckout.CheckedOutByPrincipal,
		}
		if user, ok := checkoutPrincipals[identity.PrincipalID(result.CurrentCheckout.CheckedOutByPrincipal)]; ok {
			assetSummary.CurrentCheckout.CheckedOutByPrincipal = &dto.SearchCheckoutPrincipalResponse{
				ID:    user.ID.String(),
				Email: user.Email.String(),
			}
		}
	}
	return dto.AssetSearchResultResponse{
		Type:     result.Type.String(),
		TenantID: result.TenantID.String(),
		Inventory: dto.InventorySummary{
			ID:   result.Inventory.ID.String(),
			Name: result.Inventory.Name.String(),
		},
		Asset:   assetSummary,
		Matches: searchMatchesToResponse(result.Matches),
	}
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

func searchMatchesToResponse(matches []search.Match) []dto.SearchMatch {
	data := make([]dto.SearchMatch, 0, len(matches))
	for _, match := range matches {
		data = append(data, dto.SearchMatch{
			Field: match.Field.String(),
			Value: match.Value,
		})
	}
	return data
}
