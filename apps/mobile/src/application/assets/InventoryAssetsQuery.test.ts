import { describe, expect, it } from 'vitest';
import { assetId } from '../../domain/assets/AssetSummary';
import { InventoryAssetsQuery, type InventoryAssetsRepository, type InventoryAssetsSnapshot } from './InventoryAssetsQuery';

class FakeInventoryAssetsRepository implements InventoryAssetsRepository {
  async getInventoryAssetsSnapshot() {
    return this.inventory;
  }

  private readonly inventory: InventoryAssetsSnapshot = {
    inventoryName: 'Home',
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
    const query = new InventoryAssetsQuery(new FakeInventoryAssetsRepository());

    await expect(query.execute()).resolves.toEqual({
      inventoryName: 'Home',
      assets: [
        {
          id: 'asset-garage',
          title: 'Garage',
          kindLabel: 'Place',
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
