import { describe, expect, it } from 'vitest';
import { assetId, AssetSummary } from '../../domain/assets/AssetSummary';
import {
  inventoryId,
  InventorySummary,
  tenantId
} from '../../domain/inventories/InventorySummary';
import type {
  CreateInventoryAssetInput,
  InventorySummaryRepository,
  InventoryWorkspace
} from '../home/InventorySummaryRepository';
import { ParentLookupQuery } from './ParentLookupQuery';

class FakeInventorySummaryRepository implements InventorySummaryRepository {
  searchedQuery: string | undefined;

  async getInventoryWorkspace(): Promise<InventoryWorkspace> {
    return {
      tenants: [{ id: tenantId('tenant-home'), name: 'Home tenant' }],
      defaultInventoryId: inventoryId('inventory-home'),
      inventories: [this.inventory]
    };
  }

  async getDefaultInventorySummary(): Promise<InventorySummary> {
    return this.inventory;
  }

  async selectInventory(): Promise<void> {}

  async createAsset(_input: CreateInventoryAssetInput): Promise<AssetSummary> {
    throw new Error('Not used in parent lookup tests.');
  }

  async addAssetPhoto(): Promise<void> {}

  async archiveAsset(): Promise<void> {}

  async restoreAsset(): Promise<void> {}

  async deleteAsset(): Promise<void> {}

  async browseAssets() {
    return { assets: [], hasMore: false };
  }

  async searchAssets(query: string): Promise<readonly AssetSummary[]> {
    this.searchedQuery = query;
    if (query === 'medicine') {
      return [
        asset('asset-medicine-ish-1', 'Medicine cabinet backup', 'container', 'Bathroom'),
        asset('asset-medicine-ish-2', 'Travel medicine bag', 'container', 'Hall closet'),
        asset('asset-medicine-ish-3', 'Cold medicine', 'item', 'Bathroom'),
        asset('asset-medicine-ish-4', 'Medicine labels', 'item', 'Office'),
        asset('asset-medicine-ish-5', 'Medicine bin old', 'container', 'Garage'),
        asset('asset-medicine-ish-6', 'Pet medicine', 'item', 'Kitchen'),
        asset('asset-medicine-exact', 'Medicine', 'location', 'No parent')
      ];
    }

    return this.inventory.assets.filter((asset) =>
      asset.title.toLocaleLowerCase().includes(query.toLocaleLowerCase())
    );
  }

  async searchLocations() {
    return [];
  }

  private readonly inventory: InventorySummary = {
    id: inventoryId('inventory-home'),
    tenantId: tenantId('tenant-home'),
    name: 'Home',
    role: 'editor',
    permissions: ['view', 'create_asset', 'edit_asset'],
    description: 'Home inventory.',
    updatedAtLabel: 'Updated today',
    locationCount: 1,
    locations: [],
    assets: [
      asset('asset-garage', 'Garage', 'location', 'No parent'),
      asset('asset-bin', 'Blue bin', 'container', 'Garage'),
      asset('asset-drill', 'Cordless drill', 'item', 'Blue bin')
    ]
  };
}

describe('ParentLookupQuery', () => {
  it('uses recent inventory assets as bounded empty-query parent candidates', async () => {
    const repository = new FakeInventorySummaryRepository();
    const query = new ParentLookupQuery(repository);

    await expect(query.execute('   ')).resolves.toMatchObject([
      { id: 'asset-garage', title: 'Garage', kind: 'location', willPromoteToContainer: false },
      { id: 'asset-bin', title: 'Blue bin', kind: 'container', willPromoteToContainer: false },
      {
        id: 'asset-drill',
        title: 'Cordless drill',
        kind: 'item',
        selectionHint: 'Will become a container for this item',
        willPromoteToContainer: true
      }
    ]);
    expect(repository.searchedQuery).toBeUndefined();
  });

  it('searches every asset kind for parent candidates', async () => {
    const repository = new FakeInventorySummaryRepository();
    const query = new ParentLookupQuery(repository);

    await expect(query.execute('drill')).resolves.toMatchObject([
      {
        id: 'asset-drill',
        title: 'Cordless drill',
        kind: 'item',
        selectionHint: 'Will become a container for this item',
        willPromoteToContainer: true
      }
    ]);
    expect(repository.searchedQuery).toBe('drill');
  });

  it('keeps exact parent title matches before trimming compact search results', async () => {
    const repository = new FakeInventorySummaryRepository();
    const query = new ParentLookupQuery(repository);

    const results = await query.execute('medicine');

    expect(results[0]).toMatchObject({
      id: 'asset-medicine-exact',
      title: 'Medicine',
      kind: 'location',
      selectionHint: 'Location',
      willPromoteToContainer: false
    });
    expect(results).toHaveLength(6);
  });
});

function asset(
  id: string,
  title: string,
  kind: AssetSummary['kind'],
  locationLabel: string
): AssetSummary {
  return {
    id: assetId(id),
    title,
    kind,
    lifecycleState: 'active',
    locationLabel,
    locationTrail: ['Home', locationLabel, title].filter((value) => value !== 'No parent'),
    description: '',
    updatedAtLabel: 'Updated today',
    hasPhoto: false
  };
}
