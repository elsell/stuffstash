import { ReadPageGuard } from '../shared/ReadPageGuard';
import type {
  Asset,
  AssetPhotoReference
} from '@stuff-stash/api-client';
import { AssetSummary } from '../../domain/assets/AssetSummary';

import type { InventoryApiClient } from './InventoryApiClient';
import { mapAsset, MapAssetOptions, mapWithConcurrency } from './InventoryAssetMapping';
export class ApiInventoryAssetPhotos {
  constructor(private readonly client: Pick<InventoryApiClient, 'listAssetAttachments' | 'assetAttachmentThumbnailReference'>) {}
  async mapAssetWithPhoto(
    inventoryName: string,
    asset: Asset,
    assets: readonly Asset[],
    options: MapAssetOptions = {}
  ): Promise<AssetSummary> {
    const photos = await this.photosForAsset(asset);
    return mapAsset(inventoryName, asset, assets, photos, options);
  }

  async mapAssetWithPrimaryPhoto(
    inventoryName: string,
    asset: Asset,
    assets: readonly Asset[],
    options: MapAssetOptions = {}
  ): Promise<AssetSummary> {
    const photo = await this.primaryPhotoForAsset(asset);
    return mapAsset(inventoryName, asset, assets, photo ? [photo] : [], options);
  }

  async mapAssetsWithMapPhotos(
    inventoryName: string,
    assets: readonly Asset[],
    knownAssets: readonly Asset[] = assets,
    photoAssetIds?: ReadonlySet<string>
  ): Promise<readonly AssetSummary[]> {
    return mapWithConcurrency(assets, 6, async (asset) => {
      const photo = !photoAssetIds || photoAssetIds.has(asset.id)
        ? await this.primaryMapPhotoForAsset(asset)
        : undefined;
      return mapAsset(inventoryName, asset, knownAssets, photo ? [photo] : []);
    });
  }

  async primaryMapPhotoForAsset(asset: Asset): Promise<NonNullable<AssetSummary['photo']> | undefined> {
    if (!asset.primaryPhoto) {
      return undefined;
    }
    let smallReference: AssetPhotoReference;
    try {
      smallReference = await this.client.assetAttachmentThumbnailReference(
        asset.tenantId,
        asset.inventoryId,
        asset.id,
        asset.primaryPhoto.id,
        'small'
      );
    } catch {
      return undefined;
    }
    return {
      id: asset.primaryPhoto.id,
      fileName: asset.primaryPhoto.fileName,
      contentType: asset.primaryPhoto.contentType,
      sizeBytes: asset.primaryPhoto.sizeBytes,
      uri: smallReference.uri,
      headers: smallReference.headers
    };
  }

  async primaryPhotoForAsset(asset: Asset): Promise<NonNullable<AssetSummary['photo']> | undefined> {
    if (!asset.primaryPhoto) {
      return undefined;
    }
    const smallReference = await this.client.assetAttachmentThumbnailReference(
      asset.tenantId,
      asset.inventoryId,
      asset.id,
      asset.primaryPhoto.id,
      'small'
    );
    const mediumReference = await this.client.assetAttachmentThumbnailReference(
      asset.tenantId,
      asset.inventoryId,
      asset.id,
      asset.primaryPhoto.id,
      'medium'
    );
    const largeReference = await this.client.assetAttachmentThumbnailReference(
      asset.tenantId,
      asset.inventoryId,
      asset.id,
      asset.primaryPhoto.id,
      'large'
    );
    return {
      id: asset.primaryPhoto.id,
      fileName: asset.primaryPhoto.fileName,
      contentType: asset.primaryPhoto.contentType,
      sizeBytes: asset.primaryPhoto.sizeBytes,
      uri: smallReference.uri,
      heroUri: mediumReference.uri,
      heroHeaders: mediumReference.headers,
      viewerUri: largeReference.uri,
      viewerHeaders: largeReference.headers,
      headers: smallReference.headers
    };
  }

  async photosForAsset(
    asset: Asset,
    options: { readonly allowAttachmentListFailure?: boolean; readonly signal?: AbortSignal } = {}
  ): Promise<readonly NonNullable<AssetSummary['photo']>[]> {
    const attachments = [];
    let cursor: string | undefined;
    const guard = new ReadPageGuard();

    try {
      do {
        options.signal?.throwIfAborted();
        const page = await this.client.listAssetAttachments(
          asset.tenantId,
          asset.inventoryId,
          asset.id,
          50,
          cursor,
          options.signal
        );
        options.signal?.throwIfAborted();
        attachments.push(...page.items);
        cursor = guard.accept(page.pagination.nextCursor, page.pagination.hasMore);
      } while (cursor);
    } catch (error) {
      if (options.signal?.aborted) {
        throw error;
      }
      if (options.allowAttachmentListFailure === false) {
        throw new Error('Asset attachments could not be loaded.');
      }
      return [];
    }

    return mapWithConcurrency(
      attachments.filter((item) => item.lifecycleState === 'active' && item.contentType.startsWith('image/')),
      4,
      async (attachment) => {
        const smallReference = await this.client.assetAttachmentThumbnailReference(
          asset.tenantId,
          asset.inventoryId,
          asset.id,
          attachment.id,
          'small'
        );
        const mediumReference = await this.client.assetAttachmentThumbnailReference(
          asset.tenantId,
          asset.inventoryId,
          asset.id,
          attachment.id,
          'medium'
        );
        const largeReference = await this.client.assetAttachmentThumbnailReference(
          asset.tenantId,
          asset.inventoryId,
          asset.id,
          attachment.id,
          'large'
        );
        return {
          id: attachment.id,
          fileName: attachment.fileName,
          contentType: attachment.contentType,
          sizeBytes: attachment.sizeBytes,
          uri: smallReference.uri,
          heroUri: mediumReference.uri,
          heroHeaders: mediumReference.headers,
          viewerUri: largeReference.uri,
          viewerHeaders: largeReference.headers,
          headers: smallReference.headers
        };
      }
    );
  }

}
