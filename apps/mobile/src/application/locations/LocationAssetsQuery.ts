import { assetId } from '../../domain/assets/AssetSummary';
import type {
  AssetCardViewModel,
  AssetDetailViewModel
} from '../assets/AssetViewModels';
import {
  toAssetCardViewModel,
  toAssetDetailViewModel
} from '../assets/AssetViewModels';
import type { InventorySummaryRepository } from '../home/InventorySummaryRepository';

export type LocationAssetsViewModel = {
  readonly locationId: string;
  readonly locationTitle: string;
  readonly inventoryName: string;
  readonly assets: readonly AssetCardViewModel[];
  readonly assetDetails: readonly AssetDetailViewModel[];
};

export class LocationAssetsQuery {
  constructor(private readonly inventories: InventorySummaryRepository) {}

  async execute(locationIdValue: string): Promise<LocationAssetsViewModel> {
    const selectedLocationId = assetId(locationIdValue);
    const inventory = await this.inventories.getDefaultInventorySummary();
    const location = inventory.locations.find((candidate) => candidate.id === selectedLocationId);

    if (!location) {
      throw new Error('Location is not available in the selected inventory.');
    }

    const assets = inventory.assets.filter((asset) => {
      if (asset.id === selectedLocationId) {
        return false;
      }

      return asset.locationTrail.includes(location.title) || asset.locationLabel === location.title;
    });

    return {
      locationId: location.id,
      locationTitle: location.title,
      inventoryName: inventory.name,
      assets: assets.map(toAssetCardViewModel),
      assetDetails: assets.map((asset) => toAssetDetailViewModel(asset))
    };
  }
}
