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
import { AssetDetailQuery } from './AssetDetailQuery';

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
    assets: [
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
    ]
  };
}

describe('AssetDetailQuery', () => {
  it('builds a reusable read-only asset detail view model', async () => {
    const query = new AssetDetailQuery(new FakeInventorySummaryRepository());

    await expect(query.execute('asset-passport')).resolves.toEqual({
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
    });
  });

  it('rejects an unknown asset in the selected inventory', async () => {
    const query = new AssetDetailQuery(new FakeInventorySummaryRepository());

    await expect(query.execute('asset-missing')).rejects.toThrow(
      'Asset is not available in the selected inventory.'
    );
  });
});
