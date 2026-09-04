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
  it.each([
    { permissions: ['view', 'create_asset'] as const, canAdd: true },
    { permissions: ['view'] as const, canAdd: false }
  ])('maps create permission to canAdd=$canAdd', async ({ canAdd }) => {
    const query = new LocationsQuery(new FakeLocationsRepository(canAdd));

    await expect(query.execute()).resolves.toMatchObject({ canAdd });
  });
});
