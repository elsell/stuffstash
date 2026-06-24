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
import { SearchAssetsQuery } from './SearchAssetsQuery';

class FakeInventorySummaryRepository implements InventorySummaryRepository {
  searchedQuery: string | undefined;

  constructor(private readonly results: readonly AssetSummary[]) {}

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

  async createAsset(input: CreateInventoryAssetInput): Promise<AssetSummary> {
    return {
      id: assetId('created-asset'),
      title: input.title,
      kind: input.kind,
      lifecycleState: 'active',
      locationLabel: 'Inventory root',
      locationTrail: ['Home', input.title],
      description: input.description,
      updatedAtLabel: 'Updated now',
      hasPhoto: false
    };
  }

  async addAssetPhoto(): Promise<void> {}

  async searchAssets(query: string): Promise<readonly AssetSummary[]> {
    this.searchedQuery = query;
    return this.results;
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
    locationCount: 0,
    locations: [],
    assets: []
  };
}

describe('SearchAssetsQuery', () => {
  it('searches assets through the repository and maps result rows', async () => {
    const repository = new FakeInventorySummaryRepository([
      {
        id: assetId('asset-passport'),
        title: 'Passport folder',
        kind: 'container',
        lifecycleState: 'active',
        locationLabel: 'Office closet',
        locationTrail: ['Home', 'Office closet', 'Passport folder'],
        customType: 'Documents',
        description: 'Travel documents and copies.',
        updatedAtLabel: 'Updated today',
        hasPhoto: false
      }
    ]);
    const query = new SearchAssetsQuery(repository);

    await expect(query.execute(' passport ')).resolves.toEqual({
      query: 'passport',
      assets: [
        {
          id: 'asset-passport',
          title: 'Passport folder',
          kindLabel: 'Container',
          customTypeLabel: 'Documents',
          description: 'Travel documents and copies.',
          locationTrailLabel: 'Office closet / Passport folder',
          updatedAtLabel: 'Updated today',
          photoLabel: 'Needs photo',
          imagePlaceholderLabel: 'Box'
        }
      ],
      assetDetails: [
        {
          id: 'asset-passport',
          title: 'Passport folder',
          kindLabel: 'Container',
          customTypeLabel: 'Documents',
          description: 'Travel documents and copies.',
          locationTrailLabel: 'Office closet / Passport folder',
          lifecycleLabel: 'Active',
          updatedAtLabel: 'Updated today',
          photoLabel: 'Needs photo',
          imagePlaceholderLabel: 'Box'
        }
      ]
    });
    expect(repository.searchedQuery).toBe('passport');
  });

  it('returns no rows without calling the repository for an empty query', async () => {
    const repository = new FakeInventorySummaryRepository([]);
    const query = new SearchAssetsQuery(repository);

    await expect(query.execute('   ')).resolves.toEqual({
      query: '',
      assets: [],
      assetDetails: []
    });
    expect(repository.searchedQuery).toBeUndefined();
  });
});
