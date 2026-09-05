import type {
  AssetDetailViewModel,
  AssetPhotoViewModel
} from '../../application/assets/AssetViewModels';

export function mergeProgressiveAssetDetail(
  core: AssetDetailViewModel,
  contents: AssetDetailViewModel | undefined,
  photos: readonly AssetPhotoViewModel[] | undefined
): AssetDetailViewModel {
  const resolvedPhotos = photos ?? core.photos;
  const primaryPhoto = resolvedPhotos[0];
  return {
    ...core,
    ...(contents ? {
      locationTrailLabel: contents.locationTrailLabel,
      parentLocationTrailLabel: contents.parentLocationTrailLabel,
      parentLocationTrail: contents.parentLocationTrail,
      isPlacementLoading: false,
      containedAssets: contents.containedAssets,
      containedAssetsLabel: contents.containedAssetsLabel,
      containedSpaces: contents.containedSpaces,
      containedSpacesLabel: contents.containedSpacesLabel,
      containedItems: contents.containedItems,
      containedItemsLabel: contents.containedItemsLabel
    } : {}),
    photos: resolvedPhotos,
    photoLabel: resolvedPhotos.length > 0 ? 'Photo ready' : 'Needs photo',
    ...(primaryPhoto ? {
      photo: { uri: primaryPhoto.uri, headers: primaryPhoto.headers }
    } : { photo: undefined })
  };
}
