import { assetId } from '../../domain/assets/AssetSummary';
import type { AssetDetailViewModel } from './AssetViewModels';
import { toAssetDetailViewModel } from './AssetViewModels';
import type { InventorySummaryRepository } from '../home/InventorySummaryRepository';

export class AssetDetailQuery {
  constructor(private readonly inventories: InventorySummaryRepository) {}

  async execute(assetIdValue: string): Promise<AssetDetailViewModel> {
    const selectedAssetId = assetId(assetIdValue);
    const inventory = await this.inventories.getDefaultInventorySummary();
    const asset = inventory.assets.find((candidate) => candidate.id === selectedAssetId);

    if (!asset) {
      throw new Error('Asset is not available in the selected inventory.');
    }

    return toAssetDetailViewModel(asset, {
      canManageLifecycle: inventory.permissions.includes('edit_asset')
    });
  }
}
