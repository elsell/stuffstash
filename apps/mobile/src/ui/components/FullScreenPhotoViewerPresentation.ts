import { t } from '../../presentation/localization';
export type FullScreenPhotoViewerPhoto = {
  readonly id?: string;
  readonly label: string;
  readonly uri: string;
  readonly headers?: Readonly<Record<string, string>>;
  readonly metadataLabel?: string;
};

export type FullScreenPhotoViewerActionState = {
  readonly canGoPrevious: boolean;
  readonly canGoNext: boolean;
  readonly canRemove: boolean;
  readonly positionLabel: string;
  readonly fileLabel: string;
  readonly metadataLabel?: string;
};

export function fullScreenPhotoViewerActionState(
  photos: readonly FullScreenPhotoViewerPhoto[],
  imageIndex: number,
  canRemove: boolean
): FullScreenPhotoViewerActionState {
  if (imageIndex < 0 || imageIndex >= photos.length) {
    return {
      canGoPrevious: false,
      canGoNext: false,
      canRemove: false,
      positionLabel: t('mobile.FullScreenPhotoViewerPresentation.0Of0'),
      fileLabel: t('mobile.FullScreenPhotoViewerPresentation.photo'),
      metadataLabel: undefined
    };
  }

  const photo = photos[imageIndex];
  return {
    canGoPrevious: imageIndex > 0,
    canGoNext: imageIndex >= 0 && imageIndex < photos.length - 1,
    canRemove: canRemove && photo?.id !== undefined,
    positionLabel: photos.length > 0 ? t('mobile.FullScreenPhotoViewerPresentation.of', { value: String((imageIndex + 1).toString()), value2: String(photos.length.toString()) }) : t('mobile.FullScreenPhotoViewerPresentation.0Of0'),
    fileLabel: photo?.label ?? t('mobile.FullScreenPhotoViewerPresentation.photo'),
    metadataLabel: photo?.metadataLabel
  };
}
