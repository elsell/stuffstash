import { expect, it } from 'vitest';
import { assetId } from '../../domain/assets/AssetSummary';
import { ApiInventorySummaryRepository } from './ApiInventorySummaryRepository';
import { FakeDirectUploadTransport, FakeInventoryApiClient } from './testing/InventoryApiClient';

it('prefers direct upload targets when adding asset photos', async () => {
    const client = new FakeInventoryApiClient();
    const directUploads = new FakeDirectUploadTransport();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home', directUploads);

    await repository.addAssetPhoto(assetId('asset-created'), {
      fileName: 'created.jpg',
      contentType: 'image/jpeg',
      uri: 'file:///created.jpg',
      sizeBytes: 4
    });

    expect(client.initiatedDirectUploadInput).toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      assetId: 'asset-created',
      fileName: 'created.jpg',
      sizeBytes: 4
    });
    expect(directUploads.uploads).toEqual([{
      url: 'https://uploads.example.test/object-one',
      fileUri: 'file:///created.jpg',
      fileName: 'created.jpg',
      contentType: 'image/jpeg'
    }]);
    expect(client.completedDirectUploadInput).toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      assetId: 'asset-created',
      uploadId: 'upload-one'
    });
    expect(client.createdAttachmentInput).toBeUndefined();
  });

it('allows private-network HTTP direct upload targets for local Garage development', async () => {
    const client = new FakeInventoryApiClient();
    client.directUploadURL = 'http://192.168.2.52:3900/stuffstash/object-one';
    const directUploads = new FakeDirectUploadTransport();
    const repository = new ApiInventorySummaryRepository(
      client,
      'tenant-home',
      directUploads,
      'test-scope',
      { allowLocalDevelopmentTargets: true }
    );

    await repository.addAssetPhoto(assetId('asset-created'), {
      fileName: 'created.jpg',
      contentType: 'image/jpeg',
      uri: 'file:///created.jpg',
      sizeBytes: 4
    });

    expect(directUploads.uploads).toEqual([{
      url: 'http://192.168.2.52:3900/stuffstash/object-one',
      fileUri: 'file:///created.jpg',
      fileName: 'created.jpg',
      contentType: 'image/jpeg'
    }]);
    expect(client.completedDirectUploadInput).toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      assetId: 'asset-created',
      uploadId: 'upload-one'
    });
    expect(client.createdAttachmentInput).toBeUndefined();
  });

it('deletes asset photos through the generated client wrapper', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await repository.deleteAssetPhoto(assetId('asset-filters'), 'attachment-filters-photo');

    expect(client.deletedAttachmentInput).toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      assetId: 'asset-filters',
      attachmentId: 'attachment-filters-photo'
    });
  });

it('falls back to JSON attachment upload for local-only direct upload targets', async () => {
    const client = new FakeInventoryApiClient();
    client.directUploadURL = 'stuffstash-local://direct-uploads/upload-one';
    const directUploads = new FakeDirectUploadTransport(false);
    const repository = new ApiInventorySummaryRepository(
      client,
      'tenant-home',
      directUploads,
      'test-scope',
      { allowLocalDevelopmentTargets: true }
    );

    await repository.addAssetPhoto(assetId('asset-created'), {
      fileName: 'created.jpg',
      contentType: 'image/jpeg',
      contentBase64: 'ZmFrZQ==',
      uri: 'file:///created.jpg',
      sizeBytes: 4
    });

    expect(client.completedDirectUploadInput).toBeUndefined();
    expect(client.createdAttachmentInput).toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      assetId: 'asset-created',
      fileName: 'created.jpg'
    });
  });

it('rejects local-only direct upload targets when local development targets are not enabled', async () => {
    const client = new FakeInventoryApiClient();
    client.directUploadURL = 'stuffstash-local://direct-uploads/upload-one';
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home', new FakeDirectUploadTransport(false));

    await expect(
      repository.addAssetPhoto(assetId('asset-created'), {
        fileName: 'created.jpg',
        contentType: 'image/jpeg',
        contentBase64: 'ZmFrZQ==',
        uri: 'file:///created.jpg',
        sizeBytes: 4
      })
    ).rejects.toThrow('Unsupported direct attachment upload target.');

    expect(client.createdAttachmentInput).toBeUndefined();
    expect(client.completedDirectUploadInput).toBeUndefined();
  });

it('rejects unexpected direct upload target schemes instead of silently falling back', async () => {
    const client = new FakeInventoryApiClient();
    client.directUploadURL = 'ftp://uploads.example.test/object-one';
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home', new FakeDirectUploadTransport());

    await expect(
      repository.addAssetPhoto(assetId('asset-created'), {
        fileName: 'created.jpg',
        contentType: 'image/jpeg',
        contentBase64: 'ZmFrZQ==',
        uri: 'file:///created.jpg',
        sizeBytes: 4
      })
    ).rejects.toThrow('Unsupported direct attachment upload target.');

    expect(client.createdAttachmentInput).toBeUndefined();
    expect(client.completedDirectUploadInput).toBeUndefined();
  });

it('rejects public cleartext direct upload targets', async () => {
    const client = new FakeInventoryApiClient();
    client.directUploadURL = 'http://uploads.example.test/object-one';
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home', new FakeDirectUploadTransport());

    await expect(
      repository.addAssetPhoto(assetId('asset-created'), {
        fileName: 'created.jpg',
        contentType: 'image/jpeg',
        contentBase64: 'ZmFrZQ==',
        uri: 'file:///created.jpg',
        sizeBytes: 4
      })
    ).rejects.toThrow('Unsupported direct attachment upload target.');

    expect(client.createdAttachmentInput).toBeUndefined();
    expect(client.completedDirectUploadInput).toBeUndefined();
  });
