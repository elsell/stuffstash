package app

import (
	"context"

	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type CreateAssetInput = assetapp.CreateAssetInput
type ListAssetsInput = assetapp.ListAssetsInput
type GetAssetInput = assetapp.GetAssetInput
type AssetParentUpdate = assetapp.AssetParentUpdate
type UpdateAssetInput = assetapp.UpdateAssetInput
type UpdateAssetLifecycleInput = assetapp.UpdateAssetLifecycleInput
type GetAssetResult = assetapp.GetAssetResult
type ListAssetsResult = assetapp.ListAssetsResult
type CheckoutAssetInput = assetapp.CheckoutAssetInput
type ReturnAssetInput = assetapp.ReturnAssetInput
type UpdateReturnedCheckoutDetailsInput = assetapp.UpdateReturnedCheckoutDetailsInput
type ListAssetCheckoutHistoryInput = assetapp.ListAssetCheckoutHistoryInput
type ListCheckedOutAssetsInput = assetapp.ListCheckedOutAssetsInput
type AssetCheckoutHistoryResult = assetapp.AssetCheckoutHistoryResult
type CheckedOutAssetsResult = assetapp.CheckedOutAssetsResult
type CheckoutOperationResult = assetapp.CheckoutOperationResult
type CreateAssetTagInput = assetapp.CreateAssetTagInput
type UpdateAssetTagInput = assetapp.UpdateAssetTagInput
type ListAssetTagsInput = assetapp.ListAssetTagsInput
type AssetTagLifecycleInput = assetapp.AssetTagLifecycleInput
type ListAssetTagsResult = assetapp.ListAssetTagsResult
type GetAssetAssignedTagsInput = assetapp.GetAssetAssignedTagsInput

func (a App) CreateAsset(ctx context.Context, input CreateAssetInput) (asset.Asset, error) {
	return a.assetService.CreateAsset(ctx, input)
}

func (a App) UpdateAsset(ctx context.Context, input UpdateAssetInput) (asset.Asset, error) {
	return a.assetService.UpdateAsset(ctx, input)
}

func (a App) ArchiveAsset(ctx context.Context, input UpdateAssetLifecycleInput) (asset.Asset, error) {
	return a.assetService.ArchiveAsset(ctx, input)
}

func (a App) RestoreAsset(ctx context.Context, input UpdateAssetLifecycleInput) (asset.Asset, error) {
	return a.assetService.RestoreAsset(ctx, input)
}

func (a App) GetAsset(ctx context.Context, input GetAssetInput) (asset.Asset, error) {
	return a.assetService.GetAsset(ctx, input)
}

func (a App) GetAssetDetail(ctx context.Context, input GetAssetInput) (GetAssetResult, error) {
	result, err := a.assetService.GetAssetDetail(ctx, input)
	if err != nil {
		return GetAssetResult{}, err
	}
	if result.PrimaryPhoto != nil {
		a.warmPrimarySmallThumbnails(ctx, []media.Attachment{*result.PrimaryPhoto})
	}
	return result, nil
}

func (a App) DeleteAsset(ctx context.Context, input UpdateAssetLifecycleInput) error {
	return a.assetService.DeleteAsset(ctx, input)
}

func (a App) CheckoutAsset(ctx context.Context, input CheckoutAssetInput) (asset.Checkout, error) {
	return a.assetService.CheckoutAsset(ctx, input)
}

func (a App) CheckoutAssetWithOperation(ctx context.Context, input CheckoutAssetInput) (CheckoutOperationResult, error) {
	return a.assetService.CheckoutAssetWithOperation(ctx, input)
}

func (a App) ReturnAsset(ctx context.Context, input ReturnAssetInput) (asset.Checkout, error) {
	return a.assetService.ReturnAsset(ctx, input)
}

func (a App) ReturnAssetWithOperation(ctx context.Context, input ReturnAssetInput) (CheckoutOperationResult, error) {
	return a.assetService.ReturnAssetWithOperation(ctx, input)
}

func (a App) UpdateReturnedCheckoutDetails(ctx context.Context, input UpdateReturnedCheckoutDetailsInput) (asset.Checkout, error) {
	return a.assetService.UpdateReturnedCheckoutDetails(ctx, input)
}

func (a App) ListAssetCheckoutHistory(ctx context.Context, input ListAssetCheckoutHistoryInput) (AssetCheckoutHistoryResult, error) {
	return a.assetService.ListAssetCheckoutHistory(ctx, input)
}

func (a App) ListCheckedOutAssets(ctx context.Context, input ListCheckedOutAssetsInput) (CheckedOutAssetsResult, error) {
	return a.assetService.ListCheckedOutAssets(ctx, input)
}

func (a App) ListAssets(ctx context.Context, input ListAssetsInput) (ListAssetsResult, error) {
	result, err := a.assetService.ListAssets(ctx, input)
	if err != nil {
		return ListAssetsResult{}, err
	}
	a.warmPrimarySmallThumbnails(ctx, primaryPhotosForAssets(result.Items, result.PrimaryPhotos))
	return result, nil
}

func (a App) CreateAssetTag(ctx context.Context, input CreateAssetTagInput) (assettag.Tag, error) {
	return a.assetService.CreateAssetTag(ctx, input)
}

func (a App) UpdateAssetTag(ctx context.Context, input UpdateAssetTagInput) (assettag.Tag, error) {
	return a.assetService.UpdateAssetTag(ctx, input)
}

func (a App) ArchiveAssetTag(ctx context.Context, input AssetTagLifecycleInput) (assettag.Tag, error) {
	return a.assetService.ArchiveAssetTag(ctx, input)
}

func (a App) GetAssetAssignedTags(ctx context.Context, input GetAssetAssignedTagsInput) ([]assettag.Tag, error) {
	return a.assetService.GetAssetAssignedTags(ctx, input)
}

func (a App) ListAssetTags(ctx context.Context, input ListAssetTagsInput) (ListAssetTagsResult, error) {
	return a.assetService.ListAssetTags(ctx, input)
}

func primaryPhotosForAssets(items []asset.Asset, primaryPhotos map[ports.AttachmentAssetReference]media.Attachment) []media.Attachment {
	photos := make([]media.Attachment, 0, len(items))
	for _, item := range items {
		ref := ports.AttachmentAssetReference{
			InventoryID: inventory.InventoryID(item.InventoryID.String()),
			AssetID:     item.ID,
		}
		if photo, ok := primaryPhotos[ref]; ok {
			photos = append(photos, photo)
		}
	}
	return photos
}
