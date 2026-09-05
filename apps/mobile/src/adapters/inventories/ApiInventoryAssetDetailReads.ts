import type { AssetContentsSnapshot } from '../../application/assets/AssetContentsQuery';
import type { AssetCoreSnapshot } from '../../application/assets/AssetCoreQuery';
import {
  GetInventoryAssetDetailInput
} from '../../application/home/InventorySummaryRepository';
import type { ReadRequest } from '../../application/shared/ReadRequest';
import { AssetSummary } from '../../domain/assets/AssetSummary';
import {
  inventoryId,
  tenantId
} from '../../domain/inventories/InventorySummary';
import { ApiInventoryDirectory } from './ApiInventoryDirectory';

import { ApiInventoryAssetPhotos } from './ApiInventoryAssetPhotos';
import { ApiInventoryAssetTraversal } from './ApiInventoryAssetTraversal';
import type { InventoryApiClient } from './InventoryApiClient';
import { mapAsset, selectAssetDetailWorkspace, summaryToApiAsset } from './InventoryAssetMapping';
export class ApiInventoryAssetDetailReads {
  constructor(private readonly client: Pick<InventoryApiClient, 'getAsset'>, private readonly directory: ApiInventoryDirectory, private readonly traversal: ApiInventoryAssetTraversal, private readonly photos: ApiInventoryAssetPhotos) {}
  async getAssetCore(
    selectedAssetId: AssetSummary['id'],
    request: ReadRequest = {}
  ): Promise<AssetCoreSnapshot> {
    const selected = await this.directory.selected(request.signal);
    const asset = await this.client.getAsset(
      selected.tenant.id,
      selected.inventory.id,
      selectedAssetId,
      request.signal
    );
    return {
      tenantId: tenantId(selected.tenant.id),
      inventoryId: inventoryId(selected.inventory.id),
      permissions: selected.inventory.access.permissions,
      revision: asset.updatedAt,
      asset: mapAsset(selected.inventory.name, asset, [asset])
    };
  }

  async getAssetPlacement(core: AssetCoreSnapshot, request: ReadRequest = {}): Promise<AssetSummary> {
    const selected = await this.directory.selected(request.signal);
    if (selected.tenant.id !== core.tenantId || selected.inventory.id !== core.inventoryId) {
      throw new Error('Asset scope no longer matches the selected inventory.');
    }
    const source = summaryToApiAsset(core.tenantId, core.inventoryId, core.asset);
    const ancestors = await this.traversal.loadAssetAncestors(source, request.signal);
    return mapAsset(selected.inventory.name, source, [...ancestors, source]);
  }

  async getAssetContents(
    core: AssetCoreSnapshot,
    request: ReadRequest = {}
  ): Promise<AssetContentsSnapshot> {
    const selected = await this.directory.selected(request.signal);
    if (
      tenantId(selected.tenant.id) !== core.tenantId
      || inventoryId(selected.inventory.id) !== core.inventoryId
    ) {
      throw new Error('Asset scope no longer matches the selected inventory.');
    }
    const sourceAsset = summaryToApiAsset(core.tenantId, core.inventoryId, core.asset);
    const isActiveContainableAsset = sourceAsset.lifecycleState === 'active'
      && (sourceAsset.kind === 'location' || sourceAsset.kind === 'container');
    if (!isActiveContainableAsset) {
      const ancestors = await this.traversal.loadAssetAncestors(sourceAsset, request.signal);
      return {
        asset: mapAsset(selected.inventory.name, sourceAsset, [...ancestors, sourceAsset]),
        allAssets: []
      };
    }

    const activeAssets = await this.traversal.listAllActiveInventoryAssets(
      selected.tenant.id,
      selected.inventory.id,
      request.signal
    );
    const selectedFromTraversal = activeAssets.find((asset) => asset.id === sourceAsset.id) ?? sourceAsset;
    const workspace = selectAssetDetailWorkspace(selectedFromTraversal, activeAssets);
    const assets = await this.photos.mapAssetsWithMapPhotos(
      selected.inventory.name,
      workspace.assets,
      activeAssets,
      new Set(workspace.assets.filter((asset) => asset.id !== core.asset.id).map((asset) => asset.id))
    );
    const asset = assets.find((candidate) => candidate.id === core.asset.id);
    if (!asset) {
      throw new Error('Asset is not available in the selected inventory.');
    }
    return { asset, allAssets: assets };
  }

  async getAssetPhotos(
    core: AssetCoreSnapshot,
    request: ReadRequest = {}
  ): Promise<AssetSummary['photos']> {
    const selected = await this.directory.selected(request.signal);
    if (
      tenantId(selected.tenant.id) !== core.tenantId
      || inventoryId(selected.inventory.id) !== core.inventoryId
    ) {
      throw new Error('Asset scope no longer matches the selected inventory.');
    }
    const asset = summaryToApiAsset(core.tenantId, core.inventoryId, core.asset);
    return this.photos.photosForAsset(asset, {
      allowAttachmentListFailure: false,
      signal: request.signal
    });
  }

  async getAssetDetail(input: GetInventoryAssetDetailInput, request: ReadRequest = {}): Promise<AssetSummary> {
    const asset = summaryToApiAsset(input.tenantId, input.inventoryId, input.asset);
    const photos = await this.photos.photosForAsset(asset, {
      allowAttachmentListFailure: false,
      signal: request.signal
    });

    return {
      ...input.asset,
      hasPhoto: photos.length > 0,
      photos,
      photo: photos[0]
    };
  }

}
