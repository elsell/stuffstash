import type { AssetCardViewModel } from './AssetViewModels';
import { toAssetCardViewModel } from './AssetViewModels';
import type { InventorySummaryRepository } from '../home/InventorySummaryRepository';

export type InventoryAssetsViewModel = {
  readonly inventoryName: string;
  readonly assets: readonly AssetCardViewModel[];
};

export class InventoryAssetsQuery {
  constructor(private readonly inventories: InventorySummaryRepository) {}

  async execute(): Promise<InventoryAssetsViewModel> {
    const inventory = await this.inventories.getDefaultInventorySummary();

    return {
      inventoryName: inventory.name,
      assets: inventory.assets.map(toAssetCardViewModel)
    };
  }
}
