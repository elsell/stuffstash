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

export function isAssetPhotoId(
  photos: readonly AssetPhotoViewModel[],
  photoId: string
): boolean {
  return photos.some((photo) => photo.id === photoId);
}

export function selectedAssetPhotoViewerIndex(
  photos: readonly AssetPhotoViewModel[],
  model: AssetPhotoViewerModel | undefined
): number | undefined {
  if (!model) {
    return undefined;
  }

  const selectedPhotoIndex = photos.findIndex((photo) => photo.id === model.photo.id);
  return selectedPhotoIndex >= 0 ? selectedPhotoIndex : undefined;
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

export function assetPhotoStatusLabel({
  index,
  label
}: {
  readonly index: number;
  readonly label: string;
}): string {
  if (index !== 0) {
    return label;
  }
  return 'First photo';
}

export function assetPhotoMetadataLabel(photo: AssetPhotoViewModel | undefined): string | undefined {
  if (!photo) {
    return undefined;
  }

  return photoMetadataLabel(photo);
}

export function photoMetadataLabel(photo: {
  readonly contentType?: string;
  readonly sizeBytes?: number;
}): string | undefined {
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
