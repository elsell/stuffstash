import type { AssetPhotoViewModel } from '../../application/assets/AssetViewModels';

export type AssetPhotoViewerModel = {
  readonly photo: AssetPhotoViewModel;
  readonly positionLabel: string;
  readonly previousPhotoId?: string;
  readonly nextPhotoId?: string;
};

export type AssetPhotoViewerControls = {
  readonly canGoPrevious: boolean;
  readonly canGoNext: boolean;
  readonly canRemove: boolean;
  readonly fileLabel: string;
  readonly metadataLabel?: string;
  readonly positionLabel: string;
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

export function assetPhotoViewerControls(
  model: AssetPhotoViewerModel | undefined,
  canRemoveAssetPhoto: boolean
): AssetPhotoViewerControls {
  return {
    canGoPrevious: model?.previousPhotoId !== undefined,
    canGoNext: model?.nextPhotoId !== undefined,
    canRemove: canRemoveAssetPhoto && model?.photo.id !== undefined,
    fileLabel: model?.photo.fileName ?? model?.photo.label ?? 'Photo',
    metadataLabel: assetPhotoMetadataLabel(model?.photo),
    positionLabel: model?.positionLabel ?? '0 of 0'
  };
}

export function assetPhotoViewerModelAtIndex(
  photos: readonly AssetPhotoViewModel[],
  imageIndex: number
): AssetPhotoViewerModel | undefined {
  const photo = photos[imageIndex];
  if (!photo) {
    return undefined;
  }
  return {
    photo,
    positionLabel: `${(imageIndex + 1).toString()} of ${photos.length.toString()}`,
    previousPhotoId: photos[imageIndex - 1]?.id,
    nextPhotoId: photos[imageIndex + 1]?.id
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

function assetPhotoMetadataLabel(photo: AssetPhotoViewModel | undefined): string | undefined {
  if (!photo) {
    return undefined;
  }

  const parts = [
    safeImageContentTypeLabel(photo.contentType),
    formatByteSize(photo.sizeBytes)
  ].filter((value): value is string => value !== undefined && value.length > 0);

  return parts.length > 0 ? parts.join(' · ') : undefined;
}

function safeImageContentTypeLabel(contentType: string | undefined): string | undefined {
  switch (contentType?.trim().toLocaleLowerCase()) {
    case 'image/jpeg':
      return 'JPEG image';
    case 'image/png':
      return 'PNG image';
    case 'image/webp':
      return 'WebP image';
    default:
      return undefined;
  }
}

function formatByteSize(sizeBytes: number | undefined): string | undefined {
  if (!sizeBytes || sizeBytes <= 0) {
    return undefined;
  }

  if (sizeBytes < 1024) {
    return `${sizeBytes.toString()} B`;
  }

  const units = ['KB', 'MB', 'GB'] as const;
  let value = sizeBytes / 1024;
  let unitIndex = 0;

  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }

  const rounded = value >= 10 ? Math.round(value).toString() : value.toFixed(1);
  return `${rounded} ${units[unitIndex]}`;
}
