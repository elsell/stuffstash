import { describe, expect, it } from 'vitest';
import { __expoPhotoSelectionProviderTestHooks } from './ExpoPhotoSelectionProvider';

describe('ExpoPhotoSelectionProvider', () => {
  it('preserves original selected image metadata for attachment upload', async () => {
    const photos = await __expoPhotoSelectionProviderTestHooks.mapImagePickerResult({
      canceled: false,
      assets: [{
        assetId: 'photo-1',
        uri: 'file:///original.png',
        fileName: 'original.png',
        mimeType: 'image/png',
        fileSize: 8
      }]
    }, 0);

    expect(photos[0]).toMatchObject({
      uri: 'file:///original.png',
      fileName: 'original.png',
      contentType: 'image/png',
      sizeBytes: 8
    });
  });

  it('derives byte size from base64 when fallback content is already available', async () => {
    const photos = await __expoPhotoSelectionProviderTestHooks.mapImagePickerResult({
      canceled: false,
      assets: [{
        uri: 'file:///photo.jpg',
        fileName: 'photo.jpg',
        mimeType: 'image/jpeg',
        base64: 'ZmFrZQ=='
      }]
    }, 0);

    expect(photos[0]?.sizeBytes).toBe(4);
    expect(photos[0]?.contentBase64).toBe('ZmFrZQ==');
  });

  it('drops unsupported selected file types', async () => {
    const photos = await __expoPhotoSelectionProviderTestHooks.mapImagePickerResult({
      canceled: false,
      assets: [{
        uri: 'file:///document.gif',
        fileName: 'document.gif',
        mimeType: 'image/gif',
        fileSize: 3
      }]
    }, 0);

    expect(photos).toEqual([]);
  });
});
