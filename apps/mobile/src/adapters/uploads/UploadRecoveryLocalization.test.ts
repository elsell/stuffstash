import { expect, it } from 'vitest';

it('localizes upload rejection and missing content without unsafe fallback', async () => {
  const previous = process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE;
  process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE = 'en-XA';
  try {
    const { ExpoDirectUploadTransport, attachmentContentBase64 } = await import('./ExpoDirectUploadTransport');
    const { FakeInventoryApiClient, FakeDirectUploadTransport } = await import('../inventories/testing/InventoryApiClient');
    const { ApiInventorySummaryRepository } = await import('../inventories/ApiInventorySummaryRepository');
    const { t } = await import('../../presentation/localization');
    const configurationError = t('recovery.uploadConfiguration');
    expect(configurationError).toMatch(/^\[/);
    const { assetId } = await import('../../domain/assets/AssetSummary');
    const input = {
      upload: { uploadId: 'upload', attachmentId: 'photo', expiresAt: '2026-10-03T00:00:00Z', url: 'https://uploads.example.test/photo', method: 'PUT', headers: {}, formFields: {} },
      fileUri: 'file:///photo.jpg', fileName: 'photo.jpg', contentType: 'image/jpeg' as const
    };
    let uploads = 0;
    const transport = new ExpoDirectUploadTransport({}, async () => { uploads++; return { status: 503 }; });
    await expect(transport.upload({ ...input, upload: { ...input.upload, url: 'http://public.example.test/photo' } })).rejects.toThrow(configurationError);
    await expect(transport.upload({ ...input, upload: { ...input.upload, method: 'DELETE' } })).rejects.toThrow(configurationError);
    expect(uploads).toBe(0);
    await expect(transport.upload(input)).rejects.toThrow(t('recovery.uploadFailed'));
    expect(uploads).toBe(1);
    const successful = new ExpoDirectUploadTransport({}, async (received, method) => {
      expect(received).toEqual(input);
      expect(method).toBe('PUT');
      return { status: 204 };
    });
    await expect(successful.upload(input)).resolves.toBe(true);
    await expect(attachmentContentBase64({ fileName: input.fileName, contentType: input.contentType })).rejects.toThrow(t('recovery.uploadContentUnavailable'));
    await expect(attachmentContentBase64({ fileName: input.fileName, contentType: input.contentType, contentBase64: 'aGVsbG8=' })).resolves.toBe('aGVsbG8=');
    const client = new FakeInventoryApiClient();
    client.directUploadURL = 'http://public.example.test/photo';
    const direct = new FakeDirectUploadTransport();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home', direct);
    await expect(repository.addAssetPhoto(assetId('asset-created'), { fileName: input.fileName, contentType: input.contentType, uri: input.fileUri, sizeBytes: 5 })).rejects.toThrow(configurationError);
    expect(direct.uploads).toEqual([]);
    expect(client.createdAttachmentInput).toBeUndefined();
    expect(client.completedDirectUploadInput).toBeUndefined();
  } finally {
    if (previous === undefined) delete process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE;
    else process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE = previous;
  }
});
