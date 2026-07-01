import type { AssetPhotoViewModel } from '../../application/assets/AssetViewModels';

export type AssetPhotoViewerModel = {
  readonly photo: AssetPhotoViewModel;
  readonly positionLabel: string;
  readonly previousPhotoId?: string;
  readonly nextPhotoId?: string;
};

export function assetPhotoViewerModel(
  photos: readonly AssetPhotoViewModel[],
  selectedPhotoId: string | undefined
): AssetPhotoViewerModel | undefined {
  if (!selectedPhotoId || photos.length === 0) {
    return undefined;
  }

  const selectedIndex = photos.findIndex((photo) => photo.id === selectedPhotoId);
  if (selectedIndex < 0) {
    return undefined;
  }

  return {
    photo: photos[selectedIndex] as AssetPhotoViewModel,
    positionLabel: `${(selectedIndex + 1).toString()} of ${photos.length.toString()}`,
    previousPhotoId: photos[selectedIndex - 1]?.id,
    nextPhotoId: photos[selectedIndex + 1]?.id
  };
}
