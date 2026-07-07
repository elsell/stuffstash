import { describe, expect, it } from 'vitest';
import { assetId, type AssetSummary } from '../../domain/assets/AssetSummary';
import {
  inventoryId,
  type InventorySummary,
  tenantId
} from '../../domain/inventories/InventorySummary';
import type {
  CreateInventoryAssetInput,
  InventorySummaryRepository,
  InventoryWorkspace
} from '../home/InventorySummaryRepository';
import { InventoryAssetTagsQuery } from './InventoryAssetTagsQuery';

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
      id: assetId('asset-created'),
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
    locationCount: 0,
    locations: [],
    assets: [],
    assetTags: [
      {
        id: 'tag-workshop',
        key: 'workshop',
        displayName: 'Workshop',
        color: '#2F80ED'
      }
    ]
  };
}

describe('InventoryAssetTagsQuery', () => {
  it('maps active inventory tags into mobile edit options', async () => {
    const query = new InventoryAssetTagsQuery(new FakeInventorySummaryRepository());

    await expect(query.execute()).resolves.toEqual([
      {
        id: 'tag-workshop',
        key: 'workshop',
        label: 'Workshop',
        color: '#2F80ED'
      }
    ]);
  });
});
