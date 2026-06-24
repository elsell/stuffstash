import type { LocationSummary } from '../../domain/locations/LocationSummary';
import type { InventorySummaryRepository } from '../home/InventorySummaryRepository';

export type LocationLookupResult = {
  readonly id: string;
  readonly title: string;
  readonly description: string;
  readonly containedAssetCountLabel: string;
};

export class LocationLookupQuery {
  constructor(private readonly inventories: InventorySummaryRepository) {}

  async execute(query: string): Promise<readonly LocationLookupResult[]> {
    const trimmed = query.trim();
    if (trimmed.length === 0) {
      return [];
    }

    const locations = await this.inventories.searchLocations(trimmed);
    return locations.slice(0, 4).map(toLocationLookupResult);
  }
}

function toLocationLookupResult(location: LocationSummary): LocationLookupResult {
  return {
    id: location.id,
    title: location.title,
    description: location.description,
    containedAssetCountLabel:
      location.containedAssetCount === 1
        ? '1 asset'
        : `${location.containedAssetCount.toString()} assets`
  };
}
