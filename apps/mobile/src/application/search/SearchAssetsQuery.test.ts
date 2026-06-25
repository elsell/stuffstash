import { describe, expect, it } from 'vitest';
import { assetId, AssetSummary } from '../../domain/assets/AssetSummary';
import {
  inventoryId,
  InventorySummary,
  tenantId
} from '../../domain/inventories/InventorySummary';
import type {
  AssetBrowsePage,
  AssetBrowsePageInput,
  CreateInventoryAssetInput,
  InventorySummaryRepository,
  InventoryWorkspace
} from '../home/InventorySummaryRepository';
import { SearchAssetsQuery } from './SearchAssetsQuery';

class FakeInventorySummaryRepository implements InventorySummaryRepository {
  browseInputs: AssetBrowsePageInput[] = [];

  constructor(private readonly pages: readonly AssetBrowsePage[]) {}

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
    return asset('created-asset', input.title, input.kind, 'Inventory root');
  }

  async addAssetPhoto(): Promise<void> {}

  async archiveAsset(): Promise<void> {}

  async restoreAsset(): Promise<void> {}

  async deleteAsset(): Promise<void> {}

  async browseAssets(input: AssetBrowsePageInput): Promise<AssetBrowsePage> {
    this.browseInputs.push(input);
    return this.pages[this.browseInputs.length - 1] ?? { assets: [], hasMore: false };
  }

  async searchAssets(): Promise<readonly AssetSummary[]> {
    return [];
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
  it('loads a filtered browse page and maps asset cards', async () => {
    const repository = new FakeInventorySummaryRepository([
      {
        assets: [asset('asset-passport', 'Passport folder', 'container', 'Office closet')],
        nextCursor: 'next-page',
        hasMore: true
      }
    ]);
    const query = new SearchAssetsQuery(repository);

    await expect(
      query.execute({
        query: '',
        lifecycleState: 'active',
        kind: 'container',
        sort: 'updated_desc',
        limit: 20
      })
    ).resolves.toMatchObject({
      query: '',
      mode: 'browse',
      nextCursor: 'next-page',
      hasMore: true,
      assets: [
        {
          id: 'asset-passport',
          title: 'Passport folder',
          kindLabel: 'Container',
          locationTrailLabel: 'Office closet / Passport folder'
        }
      ]
    });
    expect(repository.browseInputs).toEqual([
      {
        query: '',
        lifecycleState: 'active',
        kind: 'container',
        sort: 'updated_desc',
        limit: 20
      }
    ]);
  });

  it('passes query, cursor, lifecycle, and kind filters for paged search', async () => {
    const repository = new FakeInventorySummaryRepository([
      {
        assets: [asset('asset-ibuprofen', 'Ibuprofen', 'item', 'Medicine Bin')],
        hasMore: false
      }
    ]);
    const query = new SearchAssetsQuery(repository);

    await expect(
      query.execute({
        query: ' ibu ',
        cursor: 'cursor-1',
        lifecycleState: 'all',
        kind: 'item',
        sort: 'id_asc',
        limit: 10
      })
    ).resolves.toMatchObject({
      query: 'ibu',
      mode: 'search',
      hasMore: false,
      assets: [{ id: 'asset-ibuprofen', title: 'Ibuprofen', kindLabel: 'Item' }]
    });
    expect(repository.browseInputs).toEqual([
      {
        query: 'ibu',
        cursor: 'cursor-1',
        lifecycleState: 'all',
        kind: 'item',
        sort: 'id_asc',
        limit: 10
      }
    ]);
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
    locationTrail: ['Home', locationLabel, title],
    customType: kind === 'container' ? 'Documents' : undefined,
    description: `${title} description.`,
    updatedAtLabel: 'Updated today',
    hasPhoto: false
  };
}
