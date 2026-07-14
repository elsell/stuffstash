import { assetId } from '../../domain/assets/AssetSummary';
import type { AssetSummary } from '../../domain/assets/AssetSummary';
import type { InventoryId, TenantId } from '../../domain/inventories/InventorySummary';
import type { AssetDetailViewModel } from './AssetViewModels';
import { toAssetDetailViewModel } from './AssetViewModels';
import type { InventorySummaryRepository } from '../home/InventorySummaryRepository';
import type { InventoryMapAssetRepository } from './InventoryMapQuery';

type AssetDetailSource = {
  readonly tenantId: TenantId;
  readonly inventoryId: InventoryId;
  readonly permissions: readonly string[];
  readonly asset: AssetSummary;
  readonly allAssets: readonly AssetSummary[];
};

export type AssetDetailQueryOptions = {
  readonly source?: 'default' | 'map';
};

export class AssetDetailQuery {
  constructor(
    private readonly inventories: InventorySummaryRepository,
    private readonly mapAssets?: InventoryMapAssetRepository
  ) {}

  async execute(assetIdValue: string, options: AssetDetailQueryOptions = {}): Promise<AssetDetailViewModel> {
    const selectedAssetId = assetId(assetIdValue);
    const inventory = await this.inventories.getDefaultInventorySummary();

    if (options.source === 'map') {
      const mapSource = await this.mapDetailSource(selectedAssetId, inventory.tenantId, inventory.id);
      if (mapSource) {
        return this.buildDetailView(mapSource);
      }
      throw new Error('Asset is not available in the selected inventory.');
    }

    const summaryAsset = inventory.assets.find((candidate) => candidate.id === selectedAssetId);
    if (summaryAsset) {
      if (summaryAsset.kind === 'location') {
        const mapSource = await this.mapDetailSource(selectedAssetId, inventory.tenantId, inventory.id);
        if (mapSource) {
          return this.buildDetailView(mapSource);
        }
      }
      return this.buildDetailView({
        tenantId: inventory.tenantId,
        inventoryId: inventory.id,
        permissions: inventory.permissions,
        asset: summaryAsset,
        allAssets: inventory.assets
      });
    }

    const mapSource = await this.mapDetailSource(selectedAssetId, inventory.tenantId, inventory.id);
    if (mapSource) {
      return this.buildDetailView(mapSource);
    }

    throw new Error('Asset is not available in the selected inventory.');
  }

  private async mapDetailSource(
    selectedAssetId: AssetSummary['id'],
    tenantId: TenantId,
    inventoryId: InventoryId
  ): Promise<AssetDetailSource | undefined> {
    if (!this.mapAssets) {
      return undefined;
    }

    const mapInventory = await this.mapAssets.listActiveInventoryMapAssets();
    if (
      mapInventory.tenantId !== tenantId
      || mapInventory.inventoryId !== inventoryId
    ) {
      return undefined;
    }

    const mapAsset = mapInventory.assets.find((candidate) => candidate.id === selectedAssetId);
    if (!mapAsset) {
      return undefined;
    }

    return {
      tenantId: mapInventory.tenantId,
      inventoryId: mapInventory.inventoryId,
      permissions: mapInventory.permissions,
      asset: mapAsset,
      allAssets: mapInventory.assets
    };
  }

  private async buildDetailView(source: AssetDetailSource): Promise<AssetDetailViewModel> {
    const detailAsset = this.inventories.getAssetDetail
      ? await this.inventories.getAssetDetail({
        tenantId: source.tenantId,
        inventoryId: source.inventoryId,
        asset: source.asset
      })
      : source.asset;

    return toAssetDetailViewModel(detailAsset, {
      canManageLifecycle: source.permissions.includes('edit_asset'),
      canEditAsset: source.permissions.includes('edit_asset'),
      canCreateAsset: source.permissions.includes('create_asset'),
      allAssets: source.allAssets
    });
  }
}
