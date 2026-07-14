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
import { InventoryAssetsQuery } from './InventoryAssetsQuery';

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
      parentLocationTrail: [],
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
    locations: [],
    assets: [
      {
        id: assetId('asset-garage'),
        title: 'Garage',
        kind: 'location',
        lifecycleState: 'active',
        locationLabel: 'Home',
        locationTrail: ['Home', 'Garage'],
        parentLocationTrail: [],
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
        parentLocationTrail: [{ id: assetId('asset-garage'), title: 'Garage' }],
        description: 'MERV 11 three-pack.',
        updatedAtLabel: 'Updated today',
        hasPhoto: false,
        tags: [{ id: 'tag-workshop', key: 'workshop', displayName: 'Workshop', color: '#2F80ED' }]
      }
    ]
  };
}

describe('InventoryAssetsQuery', () => {
  it('builds image-card and detail view models for selected-inventory assets', async () => {
    const query = new InventoryAssetsQuery(new FakeInventorySummaryRepository());

    await expect(query.execute()).resolves.toEqual({
      inventoryName: 'Home',
      assets: [
        {
          id: 'asset-garage',
          title: 'Garage',
          kindLabel: 'Location',
          customTypeLabel: undefined,
          description: 'Shelves and bins.',
          locationTrailLabel: 'Garage',
          parentLocationTrail: [],
          updatedAtLabel: 'Updated today',
          photoLabel: 'Needs photo',
          imagePlaceholderLabel: 'Place'
        },
        {
          id: 'asset-filters',
          title: 'Furnace filters',
          kindLabel: 'Item',
          customTypeLabel: undefined,
          description: 'MERV 11 three-pack.',
          locationTrailLabel: 'Garage / Furnace filters',
          parentLocationTrail: [{ id: 'asset-garage', title: 'Garage', isImmediateParent: true }],
          updatedAtLabel: 'Updated today',
          photoLabel: 'Needs photo',
          tags: [{ id: 'tag-workshop', label: 'Workshop', color: '#2F80ED' }],
          imagePlaceholderLabel: 'Item'
        }
      ]
    });
  });
});
