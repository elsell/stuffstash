import { assetId } from '../../domain/assets/AssetSummary';
import type { AssetDetailViewModel } from './AssetViewModels';
import { toAssetDetailViewModel } from './AssetViewModels';
import type { InventorySummaryRepository } from '../home/InventorySummaryRepository';
import type { InventoryMapAssetRepository } from './InventoryMapQuery';

export class AssetDetailQuery {
  constructor(
    private readonly inventories: InventorySummaryRepository,
    private readonly mapAssets?: InventoryMapAssetRepository
  ) {}

  async execute(assetIdValue: string): Promise<AssetDetailViewModel> {
    const selectedAssetId = assetId(assetIdValue);
    const inventory = await this.inventories.getDefaultInventorySummary();
    const summaryAsset = inventory.assets.find((candidate) => candidate.id === selectedAssetId);

    if (summaryAsset) {
      return toAssetDetailViewModel(summaryAsset, {
        canManageLifecycle: inventory.permissions.includes('edit_asset'),
        canEditAsset: inventory.permissions.includes('edit_asset'),
        canCreateAsset: inventory.permissions.includes('create_asset'),
        allAssets: inventory.assets
      });
    }

    if (this.mapAssets) {
      const mapInventory = await this.mapAssets.listActiveInventoryMapAssets();
      if (
        mapInventory.tenantId === inventory.tenantId
        && mapInventory.inventoryId === inventory.id
      ) {
        const mapAsset = mapInventory.assets.find((candidate) => candidate.id === selectedAssetId);
        if (mapAsset) {
          return toAssetDetailViewModel(mapAsset, {
            canManageLifecycle: mapInventory.permissions.includes('edit_asset'),
            canEditAsset: mapInventory.permissions.includes('edit_asset'),
            canCreateAsset: mapInventory.permissions.includes('create_asset'),
            allAssets: mapInventory.assets
          });
        }
      }
    }

    throw new Error('Asset is not available in the selected inventory.');
  }
}
