import type { AssetCardViewModel } from './AssetViewModels';
import { toAssetCardViewModel } from './AssetViewModels';
import type { AssetSummary } from '../../domain/assets/AssetSummary';
import type { ReadRequest } from '../shared/ReadRequest';

export type InventoryAssetsViewModel = {
  readonly inventoryName: string;
  readonly assets: readonly AssetCardViewModel[];
};

export type InventoryAssetsSnapshot = {
  readonly inventoryName: string;
  readonly assets: readonly AssetSummary[];
};

export type InventoryAssetsRepository = {
  getInventoryAssetsSnapshot(request?: ReadRequest): Promise<InventoryAssetsSnapshot>;
};

export class InventoryAssetsQuery {
  constructor(private readonly inventories: InventoryAssetsRepository) {}

  async execute(request: ReadRequest = {}): Promise<InventoryAssetsViewModel> {
    const inventory = await this.inventories.getInventoryAssetsSnapshot(request);

    return {
      inventoryName: inventory.inventoryName,
      assets: inventory.assets.map(toAssetCardViewModel)
    };
  }
}
