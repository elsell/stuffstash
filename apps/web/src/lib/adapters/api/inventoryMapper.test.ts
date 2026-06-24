import { describe, expect, it } from 'vitest';
import { mapAsset, mapSearchResult } from './inventoryMapper';

describe('inventory API mapper', () => {
  it('maps generated asset DTOs into frontend domain assets', () => {
    expect(
      mapAsset({
        id: 'asset-one',
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        kind: 'location',
        title: 'Garage',
        description: 'Main storage',
        parentAssetId: null,
        lifecycleState: 'active'
      })
    ).toEqual({
      id: 'asset-one',
      tenantId: 'tenant-one',
      inventoryId: 'inventory-one',
      kind: 'location',
      title: 'Garage',
      description: 'Main storage',
      parentAssetId: null,
      lifecycleState: 'active',
      updatedAt: undefined
    });
  });

  it('maps search results through the same asset boundary', () => {
    const result = mapSearchResult({
      type: 'asset',
      tenantId: 'tenant-one',
      inventory: { id: 'inventory-one', name: 'Household' },
      asset: {
        id: 'asset-one',
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        kind: 'item',
        title: 'Passport',
        description: 'Blue folder',
        lifecycleState: 'active',
        parentAssetId: 'hall-closet'
      },
      matches: [{ field: 'title', value: 'Passport' }]
    });

    expect(result.asset.parentAssetId).toBe('hall-closet');
    expect(result.inventory.name).toBe('Household');
  });
});
