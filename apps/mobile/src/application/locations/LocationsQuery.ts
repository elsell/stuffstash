import type { LocationSummary } from '../../domain/locations/LocationSummary';
import type { InventorySummaryRepository } from '../home/InventorySummaryRepository';

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
  readonly tenantName: string;
  readonly inventoryName: string;
  readonly locations: readonly LocationBrowserItemViewModel[];
};

export class LocationsQuery {
  constructor(private readonly inventories: InventorySummaryRepository) {}

  async execute(): Promise<LocationsViewModel> {
    const workspace = await this.inventories.getInventoryWorkspace();
    const inventory =
      workspace.inventories.find((item) => item.id === workspace.defaultInventoryId) ??
      workspace.inventories[0];

    if (!inventory) {
      throw new Error('Inventory workspace must include at least one inventory.');
    }

    const tenant = workspace.tenants.find((item) => item.id === inventory.tenantId);

    if (!tenant) {
      throw new Error('Selected inventory must belong to a tenant.');
    }

    return {
      tenantName: tenant.name,
      inventoryName: inventory.name,
      locations: inventory.locations.map(toLocationViewModel)
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
