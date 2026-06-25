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
import { LocationAssetsQuery } from './LocationAssetsQuery';

class FakeInventorySummaryRepository implements InventorySummaryRepository {
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

  async archiveAsset(): Promise<void> {}

  async restoreAsset(): Promise<void> {}

  async deleteAsset(): Promise<void> {}

  async browseAssets() {
    return { assets: [], hasMore: false };
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
    locationCount: 1,
    locations: [
      {
        id: assetId('asset-garage'),
        inventoryId: inventoryId('inventory-home'),
        title: 'Garage',
        description: 'Shelves and bins.',
        containedAssetCount: 2,
        recentAssetTitles: ['Furnace filters'],
        hasPhoto: false
      }
    ],
    assets: [
      {
        id: assetId('asset-garage'),
        title: 'Garage',
        kind: 'location',
        lifecycleState: 'active',
        locationLabel: 'Home',
        locationTrail: ['Home', 'Garage'],
        description: 'Shelves and bins.',
        updatedAtLabel: 'Updated today',
        hasPhoto: false
      },
      {
        id: assetId('asset-filters'),
        title: 'Furnace filters',
        kind: 'item',
        lifecycleState: 'active',
        locationLabel: 'Garage',
        locationTrail: ['Home', 'Garage', 'Furnace filters'],
        description: 'MERV 11 three-pack.',
        updatedAtLabel: 'Updated today',
        hasPhoto: false
      }
    ]
  };
}

describe('LocationAssetsQuery', () => {
  it('builds image-card and detail view models for assets inside a location', async () => {
    const query = new LocationAssetsQuery(new FakeInventorySummaryRepository());

    await expect(query.execute('asset-garage')).resolves.toEqual({
      locationId: 'asset-garage',
      locationTitle: 'Garage',
      inventoryName: 'Home',
      assets: [
        {
          id: 'asset-filters',
          title: 'Furnace filters',
          kindLabel: 'Item',
          customTypeLabel: undefined,
          description: 'MERV 11 three-pack.',
          locationTrailLabel: 'Garage / Furnace filters',
          updatedAtLabel: 'Updated today',
          photoLabel: 'Needs photo',
          imagePlaceholderLabel: 'Item'
        }
      ],
      assetDetails: [
        {
          id: 'asset-filters',
          title: 'Furnace filters',
          kindLabel: 'Item',
          customTypeLabel: undefined,
          description: 'MERV 11 three-pack.',
          locationTrailLabel: 'Garage / Furnace filters',
          lifecycleLabel: 'Active',
          canArchive: true,
          canRestore: false,
          canDeletePermanently: false,
          updatedAtLabel: 'Updated today',
          photoLabel: 'Needs photo',
          imagePlaceholderLabel: 'Item'
        }
      ]
    });
  });

  it('rejects an unknown location in the selected inventory', async () => {
    const query = new LocationAssetsQuery(new FakeInventorySummaryRepository());

    await expect(query.execute('asset-attic')).rejects.toThrow(
      'Location is not available in the selected inventory.'
    );
  });
});
