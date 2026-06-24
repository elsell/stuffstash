import { describe, expect, it } from 'vitest';
import { mapAsset, mapCapability, mapInventory, mapSearchResult, mapTenant } from './inventoryMapper';

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

  it('maps access metadata and derives workspace capabilities from inventory permissions', () => {
    expect(
      mapTenant({
        id: 'tenant-one',
        name: 'Home',
        access: { relationship: 'owner', permissions: ['view', 'create_inventory'] }
      }).access
    ).toEqual({ relationship: 'owner', permissions: ['view', 'create_inventory'] });

    const editableInventory = mapInventory({
      id: 'inventory-one',
      tenantId: 'tenant-one',
      name: 'Household',
      access: { relationship: 'editor', permissions: ['view', 'create_asset'] }
    });
    const viewerInventory = mapInventory({
      id: 'inventory-two',
      tenantId: 'tenant-one',
      name: 'Archive',
      access: { relationship: 'viewer', permissions: ['view'] }
    });

    expect(mapCapability(editableInventory)).toBe('editor');
    expect(mapCapability(viewerInventory)).toBe('viewer');
  });
});
