import type { AssetPhotoViewModel } from '../../application/assets/AssetViewModels';

export type AssetPhotoViewerModel = {
  readonly photo: AssetPhotoViewModel;
  readonly positionLabel: string;
  readonly previousPhotoId?: string;
  readonly nextPhotoId?: string;
};

export const localAssetPhotoOrderNotice = 'Preview order only. Resets after refresh.';

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

export function orderedAssetPhotos(
  photos: readonly AssetPhotoViewModel[],
  photoOrder: readonly string[]
): readonly AssetPhotoViewModel[] {
  if (photoOrder.length === 0) {
    return photos;
  }

  const photosById = new Map(photos.flatMap((photo) => photo.id ? [[photo.id, photo] as const] : []));
  const ordered = photoOrder
    .map((photoId) => photosById.get(photoId))
    .filter((photo): photo is AssetPhotoViewModel => photo !== undefined);
  const orderedIds = new Set(ordered.map((photo) => photo.id).filter((photoId): photoId is string => photoId !== undefined));
  const unordered = photos.filter((photo) => !photo.id || !orderedIds.has(photo.id));

  return [...ordered, ...unordered];
}

export function moveAssetPhotoOrder({
  direction,
  photoId,
  photoOrder,
  photos
}: {
  readonly direction: -1 | 1;
  readonly photoId: string;
  readonly photoOrder: readonly string[];
  readonly photos: readonly AssetPhotoViewModel[];
}): readonly string[] {
  const orderedPhotos = orderedAssetPhotos(photos, photoOrder);
  const photoIndex = orderedPhotos.findIndex((photo) => photo.id === photoId);
  const nextIndex = photoIndex + direction;
  if (photoIndex < 0 || nextIndex < 0 || nextIndex >= orderedPhotos.length) {
    return photoOrder;
  }

  const nextPhotos = [...orderedPhotos];
  const [movingPhoto] = nextPhotos.splice(photoIndex, 1);
  if (!movingPhoto) {
    return photoOrder;
  }
  nextPhotos.splice(nextIndex, 0, movingPhoto);

  return nextPhotos
    .map((photo) => photo.id)
    .filter((nextPhotoId): nextPhotoId is string => nextPhotoId !== undefined);
}

export function assetPhotoStatusLabel({
  hasLocalOrder,
  index,
  label
}: {
  readonly hasLocalOrder: boolean;
  readonly index: number;
  readonly label: string;
}): string {
  if (index !== 0) {
    return label;
  }
  return hasLocalOrder ? 'Preview first' : 'First photo';
}

export function resetLocalAssetPhotoOrder(): readonly string[] {
  return [];
}
