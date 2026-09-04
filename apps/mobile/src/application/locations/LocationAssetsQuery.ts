import type { AssetSummary } from '../../domain/assets/AssetSummary';
import type { ReadRequest } from '../shared/ReadRequest';
import type {
  AssetCardViewModel,
  AssetDetailViewModel
} from '../assets/AssetViewModels';
import {
  toAssetCardViewModel,
  toAssetDetailViewModel
} from '../assets/AssetViewModels';

export type LocationAssetsViewModel = {
  readonly locationId: string;
  readonly locationTitle: string;
  readonly inventoryName: string;
  readonly assets: readonly AssetCardViewModel[];
  readonly assetDetails: readonly AssetDetailViewModel[];
};

export type LocationAssetsSnapshot = {
  readonly locationId: string;
  readonly locationTitle: string;
  readonly inventoryName: string;
  readonly assets: readonly AssetSummary[];
};

export type LocationAssetsRepository = {
  getLocationAssetsSnapshot(locationId: string, request?: ReadRequest): Promise<LocationAssetsSnapshot>;
};

export class LocationAssetsQuery {
  constructor(private readonly locations: LocationAssetsRepository) {}

  async execute(locationIdValue: string, request: ReadRequest = {}): Promise<LocationAssetsViewModel> {
    const snapshot = await this.locations.getLocationAssetsSnapshot(locationIdValue, request);

    return {
      locationId: snapshot.locationId,
      locationTitle: snapshot.locationTitle,
      inventoryName: snapshot.inventoryName,
      assets: snapshot.assets.map(toAssetCardViewModel),
      assetDetails: snapshot.assets.map((asset) => toAssetDetailViewModel(asset))
    };
  }
}
