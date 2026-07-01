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
  private readonly inventory: InventorySummary;

  constructor(permissions: InventorySummary['permissions'] = ['view', 'create_asset', 'edit_asset']) {
    this.inventory = {
      id: inventoryId('inventory-home'),
      tenantId: tenantId('tenant-home'),
      name: 'Home',
      role: 'editor',
      permissions,
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
        },
        {
          id: assetId('asset-birth-certificate'),
          title: 'Birth certificate',
          kind: 'item',
          lifecycleState: 'active',
          parentAssetId: assetId('asset-passport'),
          locationLabel: 'Passport folder',
          locationTrail: ['Home', 'Office closet', 'Passport folder', 'Birth certificate'],
          description: '',
          updatedAtLabel: 'Updated yesterday',
          hasPhoto: false
        }
      ]
    };
  }

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

}

describe('AssetDetailQuery', () => {
  it('builds a reusable asset workspace view model', async () => {
    const query = new AssetDetailQuery(new FakeInventorySummaryRepository());

    await expect(query.execute('asset-passport')).resolves.toEqual({
      id: 'asset-passport',
      title: 'Passport folder',
      kind: 'container',
      kindLabel: 'Container',
      customTypeLabel: 'Documents',
      description: 'Travel documents and copies.',
      parentAssetId: undefined,
      locationTrailLabel: 'Office closet / Passport folder',
      parentLocationTrailLabel: 'Inventory root',
      lifecycleLabel: 'Active',
      isActive: true,
      canEdit: true,
      canMove: true,
      canAddPhotos: true,
      canArchive: true,
      canRestore: false,
      canDeletePermanently: false,
      canContainAssets: true,
      containedAssetsLabel: '1 thing inside',
      containedAssets: [{
        id: 'asset-birth-certificate',
        title: 'Birth certificate',
        kindLabel: 'Item',
        customTypeLabel: undefined,
        description: '',
        locationTrailLabel: 'Office closet / Passport folder / Birth certificate',
        updatedAtLabel: 'Updated yesterday',
        photoLabel: 'Needs photo',
        photo: undefined,
        imagePlaceholderLabel: 'Item'
      }],
      updatedAtLabel: 'Updated today',
      photoLabel: 'Needs photo',
      photos: [],
      photo: undefined,
      imagePlaceholderLabel: 'Box'
    });
  });

  it('rejects an unknown asset in the selected inventory', async () => {
    const query = new AssetDetailQuery(new FakeInventorySummaryRepository());

    await expect(query.execute('asset-missing')).rejects.toThrow(
      'Asset is not available in the selected inventory.'
    );
  });

  it('does not expose lifecycle controls without edit permission', async () => {
    const query = new AssetDetailQuery(new FakeInventorySummaryRepository(['view']));

    await expect(query.execute('asset-passport')).resolves.toMatchObject({
      canArchive: false,
      canRestore: false,
      canDeletePermanently: false
    });
  });
});
