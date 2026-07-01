import { describe, expect, it, vi } from 'vitest';
import { __expoPhotoSelectionProviderTestHooks } from './ExpoPhotoSelectionProvider';

vi.mock('expo-image-manipulator', () => ({
  SaveFormat: { JPEG: 'jpeg' },
  manipulateAsync: vi.fn(async () => ({
    uri: 'file:///normalized.jpg',
    base64: 'bm9ybWFsaXplZA=='
  }))
}));

describe('ExpoPhotoSelectionProvider', () => {
  it('normalizes selected images to jpeg when the native manipulator is available', async () => {
    const photos = await __expoPhotoSelectionProviderTestHooks.mapImagePickerResult({
      canceled: false,
      assets: [{
        assetId: 'photo-1',
        uri: 'file:///original.png',
        fileName: 'original.png',
        mimeType: 'image/png',
        base64: 'b3JpZ2luYWw=',
        width: 3024,
        height: 4032
      }]
    }, 0);

    expect(photos[0]).toMatchObject({
      uri: 'file:///normalized.jpg',
      fileName: 'original.jpg',
      contentType: 'image/jpeg',
      contentBase64: 'bm9ybWFsaXplZA=='
    });
  });

  it('preserves the original mime type when normalization cannot run', async () => {
    const ImageManipulator = await import('expo-image-manipulator');
    vi.mocked(ImageManipulator.manipulateAsync).mockRejectedValueOnce(new Error('native module unavailable'));

    const photos = await __expoPhotoSelectionProviderTestHooks.mapImagePickerResult({
      canceled: false,
      assets: [{
        assetId: 'photo-1',
        uri: 'file:///original.png',
        fileName: 'original.png',
        mimeType: 'image/png',
        base64: 'b3JpZ2luYWw=',
        width: 3024,
        height: 4032
      }]
    }, 0);

    expect(photos[0]).toMatchObject({
      uri: 'file:///original.png',
      fileName: 'original.png',
      contentType: 'image/png',
      contentBase64: 'b3JpZ2luYWw='
    });
  });

  it('drops unsupported selected file types', async () => {
    const photos = await __expoPhotoSelectionProviderTestHooks.mapImagePickerResult({
      canceled: false,
      assets: [{
        uri: 'file:///document.gif',
        fileName: 'document.gif',
        mimeType: 'image/gif',
        base64: 'Z2lm'
      }]
    }, 0);

    expect(photos).toEqual([]);
  });
});
