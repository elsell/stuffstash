import { assetId } from '../../domain/assets/AssetSummary';
import { inventoryId } from '../../domain/inventories/InventorySummary';
import { describe, expect, it } from 'vitest';
import { LocationsQuery, type LocationsRepository } from './LocationsQuery';

class FakeLocationsRepository implements LocationsRepository {
  constructor(private readonly canAdd: boolean) {}

  async getLocationsSnapshot() {
    return {
      canAdd: this.canAdd,
      tenantName: 'Household',
      inventoryName: 'Home',
      locations: []
    };
  }
}

describe('LocationsQuery', () => {
  it('keeps photo readiness separate from localized labels and image availability', async () => {
    const query = new LocationsQuery({ async getLocationsSnapshot() {
      return { canAdd: false, tenantName: 'Household', inventoryName: 'Home', locations: [{
        id: assetId('garage'), inventoryId: inventoryId('home'), title: 'Garage', description: '',
        containedAssetCount: 0, recentAssetTitles: [], hasPhoto: true
      }] };
    } });
    const result = await query.execute();
    expect(result.locations[0]).toMatchObject({ hasPhoto: true, photo: undefined });
  });
  it.each([
    { permissions: ['view', 'create_asset'] as const, canAdd: true },
    { permissions: ['view'] as const, canAdd: false }
  ])('maps create permission to canAdd=$canAdd', async ({ canAdd }) => {
    const query = new LocationsQuery(new FakeLocationsRepository(canAdd));

    await expect(query.execute()).resolves.toMatchObject({ canAdd });
  });
});
