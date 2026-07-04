import { describe, expect, it } from 'vitest';
import type { AssetAttachment, AssetViewModel } from '$lib/domain/inventory';
import {
  buildDetailPhotos,
  photoGalleryEmptyMessage,
  photoUploadUnavailableReason,
  supportedAttachmentContentType,
  supportedImageContentType,
  unsupportedAttachmentTypeMessage,
  unsupportedImageTypeMessage
} from './workspaceAssetMedia';

describe('workspace asset media helpers', () => {
  it('promotes the exact asset primary photo and ignores media owned by another asset', () => {
    const photos = buildDetailPhotos(
      {
        ...asset(),
        photo: { id: 'photo-one', assetId: 'asset-one', url: 'blob:primary', alt: 'Bottle front' }
      },
      [
        attachment('photo-one', 'front.jpg', 'asset-one', 'blob:front-thumb'),
        attachment('photo-two', 'side.jpg', 'asset-one', 'blob:side-thumb'),
        attachment('other-photo', 'wrong.jpg', 'other-asset', 'blob:wrong-thumb')
      ]
    );

    expect(photos).toEqual([
      {
        id: 'photo-one',
        url: 'blob:primary',
        alt: 'Bottle front',
        fileName: 'front.jpg',
        sizeBytes: 12,
        isPrimary: true
      },
      {
        id: 'photo-two',
        url: 'blob:side-thumb',
        alt: 'side.jpg',
        fileName: 'side.jpg',
        sizeBytes: 12,
        isPrimary: false
      }
    ]);
  });

  it('uses an own primary photo even when it is not in the attachment rail', () => {
    expect(
      buildDetailPhotos(
        {
          ...asset(),
          photo: { id: 'photo-primary', assetId: 'asset-one', url: 'blob:primary', alt: 'Primary photo' }
        },
        []
      )
    ).toEqual([
      {
        id: 'photo-primary',
        url: 'blob:primary',
        alt: 'Primary photo',
        fileName: 'Primary photo',
        isPrimary: true
      }
    ]);
  });

  it('keeps mismatched primary photos out of the gallery', () => {
    const photos = buildDetailPhotos(
      {
        ...asset(),
        photo: { id: 'wrong-photo', assetId: 'other-asset', url: 'blob:wrong', alt: 'Wrong photo' }
      },
      [attachment('wrong-photo', 'wrong.jpg', 'other-asset', 'blob:wrong-thumb')]
    );

    expect(photos).toEqual([]);
  });

  it('checks supported attachment and image media types', () => {
    expect(supportedAttachmentContentType(['image/jpeg', 'application/pdf'], 'application/pdf')).toBe(true);
    expect(supportedAttachmentContentType(['image/jpeg'], 'application/pdf')).toBe(false);
    expect(supportedImageContentType(['image/jpeg'], 'image/jpeg')).toBe(true);
    expect(supportedImageContentType(['image/jpeg'], 'application/pdf')).toBe(false);
  });

  it('explains why photo upload is unavailable', () => {
    expect(photoUploadUnavailableReason({ canEditAsset: false, lifecycleState: 'active', isSaving: false, supportedImageTypeCount: 1 })).toBe(
      'Photo upload requires asset edit access.'
    );
    expect(photoUploadUnavailableReason({ canEditAsset: true, lifecycleState: 'archived', isSaving: false, supportedImageTypeCount: 1 })).toBe(
      'Restore this asset before adding photos.'
    );
    expect(photoUploadUnavailableReason({ canEditAsset: true, lifecycleState: 'active', isSaving: true, supportedImageTypeCount: 1 })).toBe(
      'Finish the current change before adding photos.'
    );
    expect(photoUploadUnavailableReason({ canEditAsset: true, lifecycleState: 'active', isSaving: false, supportedImageTypeCount: 0 })).toBe(
      'Photo uploads are unavailable for this media policy.'
    );
    expect(photoUploadUnavailableReason({ canEditAsset: true, lifecycleState: 'active', isSaving: false, supportedImageTypeCount: 1 })).toBe('');
  });

  it('builds asset media empty and unsupported type presentation', () => {
    expect(photoGalleryEmptyMessage()).toBe('No photos yet.');
    expect(unsupportedAttachmentTypeMessage()).toBe('Unsupported file type.');
    expect(unsupportedImageTypeMessage()).toBe('Unsupported image type.');
  });
});

function asset(): AssetViewModel {
  return {
    id: 'asset-one',
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    kind: 'item',
    title: 'Ibuprofen',
    description: '',
    parentAssetId: null,
    lifecycleState: 'active',
    containmentTrail: 'Hall closet'
  };
}

function attachment(id: string, fileName: string, assetId: string, thumbnailUrl: string): AssetAttachment {
  return {
    id,
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    assetId,
    fileName,
    contentType: 'image/jpeg',
    sizeBytes: 12,
    lifecycleState: 'active',
    thumbnailUrl
  };
}
