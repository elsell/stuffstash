import { describe, expect, it } from 'vitest';
import { mapAsset, mapAssetCheckout, mapAssetTag, mapCapability, mapCheckedOutAsset, mapInventory, mapSearchResult, mapTenant } from './inventoryMapper';

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
        lifecycleState: 'active',
        customFields: {},
        tags: [{ id: 'tag-one', key: 'workshop', displayName: 'Workshop', color: '#2F80ED' }],
        currentCheckout: {
          id: 'checkout-one',
          state: 'open',
          checkedOutAt: '2026-06-24T11:00:00Z',
          checkedOutByPrincipalId: 'principal-one'
        }
      } as any)
    ).toEqual({
      id: 'asset-one',
      tenantId: 'tenant-one',
      inventoryId: 'inventory-one',
      kind: 'location',
      title: 'Garage',
      description: 'Main storage',
      parentAssetId: null,
      lifecycleState: 'active',
      customFields: {},
      tags: [{ id: 'tag-one', key: 'workshop', displayName: 'Workshop', color: '#2F80ED' }],
      currentCheckout: {
        id: 'checkout-one',
        state: 'open',
        checkedOutAt: '2026-06-24T11:00:00Z',
        checkedOutByPrincipalId: 'principal-one'
      },
      updatedAt: undefined
    });
  });

  it('maps generated tag DTOs into frontend domain tags', () => {
    expect(
      mapAssetTag({
        id: 'tag-one',
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        key: 'workshop',
        displayName: 'Workshop',
        color: '#2F80ED',
        lifecycleState: 'active',
        createdAt: '2026-07-07T12:00:00Z',
        updatedAt: '2026-07-07T12:00:00Z'
      })
    ).toEqual({ id: 'tag-one', key: 'workshop', displayName: 'Workshop', color: '#2F80ED' });
  });

  it('maps checkout history and checked-out asset DTOs', () => {
    const checkout = mapAssetCheckout({
      id: 'checkout-one',
      tenantId: 'tenant-one',
      inventoryId: 'inventory-one',
      assetId: 'asset-one',
      state: 'returned',
      checkedOutAt: '2026-06-24T11:00:00Z',
      checkedOutByPrincipalId: 'principal-one',
      checkoutDetails: 'using at desk',
      returnedAt: '2026-06-24T12:00:00Z',
      returnedByPrincipalId: 'principal-two',
      returnDetails: 'back in bin',
      createdAt: '2026-06-24T11:00:00Z',
      updatedAt: '2026-06-24T12:00:00Z'
    } as any);

    expect(checkout).toMatchObject({
      id: 'checkout-one',
      state: 'returned',
      checkoutDetails: 'using at desk',
      returnDetails: 'back in bin'
    });
    expect(
      mapCheckedOutAsset({
        asset: {
          id: 'asset-one',
          tenantId: 'tenant-one',
          inventoryId: 'inventory-one',
          kind: 'item',
          title: 'Socket set',
          description: '',
          parentAssetId: null,
          lifecycleState: 'archived',
          customFields: {},
          tags: []
        },
        checkout: {
          id: 'checkout-open',
          state: 'open',
          checkedOutAt: '2026-06-24T11:00:00Z',
          checkedOutByPrincipalId: 'principal-one'
        }
      } as any)
    ).toMatchObject({
      asset: { id: 'asset-one', lifecycleState: 'archived' },
      checkout: { id: 'checkout-open', state: 'open' }
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
        customFields: {},
        tags: [],
        parentAssetId: 'hall-closet'
      } as any,
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
