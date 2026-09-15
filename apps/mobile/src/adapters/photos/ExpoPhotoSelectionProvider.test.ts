import { beforeEach, describe, expect, it } from 'vitest';
import { ExpoPhotoSelectionProvider, __expoPhotoSelectionProviderTestHooks } from './ExpoPhotoSelectionProvider';

import { photoPickerFake } from '../../test-support/expo-image-picker';

describe('ExpoPhotoSelectionProvider', () => {
  beforeEach(() => photoPickerFake.reset());

  it('returns selected photos without requiring broad library permission', async () => {
    photoPickerFake.result = { canceled: false, assets: [{
      uri: 'file:///selected.png', width: 10, height: 10,
      mimeType: 'image/png', fileName: 'selected.png', base64: 'ZmFrZQ=='
    }] };
    const photos = await new ExpoPhotoSelectionProvider().selectFromLibrary(0);
    expect(photos).toHaveLength(1);
    expect(photos[0]).toMatchObject({ uri: 'file:///selected.png', contentBase64: 'ZmFrZQ==' });
    expect(photoPickerFake.libraryLaunches).toBe(1);
    expect(photoPickerFake.libraryPermissionRequests).toBe(0);
    expect(photoPickerFake.libraryOptions?.mediaTypes).toEqual(['images']);
  });

  it('treats system picker cancellation as no selection', async () => {
    await expect(new ExpoPhotoSelectionProvider().selectFromLibrary(2)).resolves.toEqual([]);
    expect(photoPickerFake.libraryLaunches).toBe(1);
  });

  it('does not launch the camera without permission', async () => {
    await expect(new ExpoPhotoSelectionProvider().captureFromCamera(0)).rejects.toThrow('Camera access');
    expect(photoPickerFake.cameraLaunches).toBe(0);
  });

  it('opens the camera after permission and preserves cancellation', async () => {
    photoPickerFake.cameraGranted = true;
    await expect(new ExpoPhotoSelectionProvider().captureFromCamera(0)).resolves.toEqual([]);
    expect(photoPickerFake.cameraLaunches).toBe(1);
  });

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

  it.each(['library', 'camera'] as const)('reports unsupported %s images instead of treating selection as cancellation', async source => {
    photoPickerFake.cameraGranted = true;
    photoPickerFake.result = { canceled: false, assets: [{
      uri: 'file:///document.gif', fileName: 'document.gif', mimeType: 'image/gif', fileSize: 3, width: 10, height: 10
    }] };
    const provider = new ExpoPhotoSelectionProvider();
    await expect(source === 'library' ? provider.selectFromLibrary(2) : provider.captureFromCamera(2))
      .rejects.toThrow('Choose JPEG, PNG, or WebP photos. This selection includes an unsupported image format.');
  });

  it('rejects a mixed selection rather than silently accepting only supported images', async () => {
    photoPickerFake.result = { canceled: false, assets: [
      { uri: 'file:///photo.jpg', fileName: 'photo.jpg', mimeType: 'image/jpeg', fileSize: 4, width: 10, height: 10 },
      { uri: 'file:///animation.gif', fileName: 'animation.gif', mimeType: 'image/gif', fileSize: 3, width: 10, height: 10 }
    ] };
    await expect(new ExpoPhotoSelectionProvider().selectFromLibrary(0)).rejects.toThrow('unsupported image format');
  });
});
