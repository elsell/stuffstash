import type { LocationSummary } from '../../domain/locations/LocationSummary';
import type { ReadRequest } from '../shared/ReadRequest';

export type LocationBrowserItemViewModel = {
  readonly id: string;
  readonly title: string;
  readonly description: string;
  readonly containedAssetCountLabel: string;
  readonly recentAssetLabel: string;
  readonly photoLabel: string;
  readonly photo?: {
    readonly uri: string;
    readonly headers?: Readonly<Record<string, string>>;
  };
};

export type LocationsViewModel = {
  readonly canAdd: boolean;
  readonly tenantName: string;
  readonly inventoryName: string;
  readonly locations: readonly LocationBrowserItemViewModel[];
};

export type LocationsSnapshot = {
  readonly canAdd: boolean;
  readonly tenantName: string;
  readonly inventoryName: string;
  readonly locations: readonly LocationSummary[];
};

export type LocationsRepository = {
  getLocationsSnapshot(request?: ReadRequest): Promise<LocationsSnapshot>;
};

export class LocationsQuery {
  constructor(private readonly locations: LocationsRepository) {}

  async execute(request: ReadRequest = {}): Promise<LocationsViewModel> {
    const snapshot = await this.locations.getLocationsSnapshot(request);

    return {
      canAdd: snapshot.canAdd,
      tenantName: snapshot.tenantName,
      inventoryName: snapshot.inventoryName,
      locations: snapshot.locations.map(toLocationViewModel)
    };
  }
}

function toLocationViewModel(location: LocationSummary): LocationBrowserItemViewModel {
  return {
    id: location.id,
    title: location.title,
    description: location.description,
    containedAssetCountLabel:
      location.containedAssetCount === 1
        ? '1 asset'
        : `${location.containedAssetCount.toString()} assets`,
    recentAssetLabel:
      location.recentAssetTitles.length > 0
        ? location.recentAssetTitles.join(', ')
        : 'No recent assets',
    photoLabel: location.hasPhoto ? 'Photo ready' : 'Needs photo',
    photo: location.photo
  };
}
