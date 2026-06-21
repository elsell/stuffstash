package app

import (
	"context"

	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
)

type CreateAssetInput = assetapp.CreateAssetInput
type ListAssetsInput = assetapp.ListAssetsInput
type GetAssetInput = assetapp.GetAssetInput
type AssetParentUpdate = assetapp.AssetParentUpdate
type UpdateAssetInput = assetapp.UpdateAssetInput
type UpdateAssetLifecycleInput = assetapp.UpdateAssetLifecycleInput
type ListAssetsResult = assetapp.ListAssetsResult

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

func (a App) DeleteAsset(ctx context.Context, input UpdateAssetLifecycleInput) error {
	return a.assetService.DeleteAsset(ctx, input)
}

func (a App) ListAssets(ctx context.Context, input ListAssetsInput) (ListAssetsResult, error) {
	return a.assetService.ListAssets(ctx, input)
}
