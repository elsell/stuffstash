import { localization, t } from '../../presentation/localization';
import type { LocationSummary } from '../../domain/locations/LocationSummary';
import type { ReadRequest } from '../shared/ReadRequest';

export type LocationBrowserItemViewModel = {
  readonly id: string;
  readonly title: string;
  readonly pathLabel?: string;
  readonly description: string;
  readonly containedAssetCountLabel: string;
  readonly recentAssetLabel: string;
  readonly photoLabel: string;
  readonly hasPhoto: boolean;
  readonly photo?: {
    readonly variant?: 'small' | 'medium' | 'large' | 'original';
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
    pathLabel: location.parentLocationTrailIncomplete ? t('locations.partialPath', { path: [...(location.parentLocationTrail ?? []).map(parent => parent.title), location.title].join(' / ') }) : [...(location.parentLocationTrail ?? []).map(parent => parent.title), location.title].join(' / '),
    description: location.description,
    containedAssetCountLabel:
      t('locations.assetCount', { count: location.containedAssetCount }),
    recentAssetLabel:
      location.recentAssetTitles.length > 0
        ? localization.list(location.recentAssetTitles)
        : t('locations.noRecentAssets'),
    photoLabel: location.hasPhoto ? t('mobile.AssetViewModels.photoReady') : t('mobile.AssetViewModels.needsPhoto'),
    hasPhoto: location.hasPhoto,
    photo: location.photo
  };
}
