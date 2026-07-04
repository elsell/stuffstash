import type { AssetAttachment, AssetLifecycleState, AssetViewModel, SelectedAttachment } from '$lib/domain/inventory';

export type DetailPhoto = {
  id: string;
  url: string;
  alt: string;
  fileName: string;
  sizeBytes?: number;
  isPrimary: boolean;
};

export function supportedAttachmentContentType(
  supportedContentTypes: string[],
  contentType: string
): contentType is SelectedAttachment['contentType'] {
  return supportedContentTypes.includes(contentType as SelectedAttachment['contentType']);
}

export function supportedImageContentType(imageContentTypes: string[], contentType: string): contentType is SelectedAttachment['contentType'] {
  return imageContentTypes.includes(contentType as SelectedAttachment['contentType']);
}

export function buildDetailPhotos(currentAsset: AssetViewModel, imageAttachments: AssetAttachment[]): DetailPhoto[] {
  const ownAssetPhoto = currentAsset.photo?.assetId === currentAsset.id ? currentAsset.photo : undefined;
  const photos: DetailPhoto[] = imageAttachments
    .filter((attachment) => attachment.assetId === currentAsset.id)
    .filter((attachment) => attachment.thumbnailUrl)
    .map((attachment) => ({
      id: attachment.id,
      url: attachment.id === ownAssetPhoto?.id ? ownAssetPhoto.url : (attachment.thumbnailUrl ?? ''),
      alt: attachment.id === ownAssetPhoto?.id ? ownAssetPhoto.alt : attachment.fileName,
      fileName: attachment.fileName,
      sizeBytes: attachment.sizeBytes,
      isPrimary: attachment.id === ownAssetPhoto?.id
    }));

  if (ownAssetPhoto && !photos.some((photo) => photo.id === ownAssetPhoto.id)) {
    photos.unshift({
      id: ownAssetPhoto.id,
      url: ownAssetPhoto.url,
      alt: ownAssetPhoto.alt,
      fileName: ownAssetPhoto.alt,
      isPrimary: true
    });
  }

  if (ownAssetPhoto && photos.length > 0 && !photos.some((photo) => photo.isPrimary)) {
    photos[0] = { ...photos[0], isPrimary: true };
  }

  return photos;
}

export function photoUploadUnavailableReason(input: {
  canEditAsset: boolean;
  lifecycleState: AssetLifecycleState;
  isSaving: boolean;
  supportedImageTypeCount: number;
}): string {
  if (!input.canEditAsset) {
    return 'Photo upload requires asset edit access.';
  }
  if (input.lifecycleState !== 'active') {
    return 'Restore this asset before adding photos.';
  }
  if (input.isSaving) {
    return 'Finish the current change before adding photos.';
  }
  if (input.supportedImageTypeCount === 0) {
    return 'Photo uploads are unavailable for this media policy.';
  }
  return '';
}
