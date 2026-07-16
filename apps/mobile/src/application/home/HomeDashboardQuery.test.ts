import { describe, expect, it } from 'vitest';
import { assetId } from '../../domain/assets/AssetSummary';
import {
  inventoryId,
  tenantId
} from '../../domain/inventories/InventorySummary';
import { HomeDashboardQuery } from './HomeDashboardQuery';
import {
  HomeDashboardSnapshot,
  HomeDashboardSnapshotRepository,
  InventoryWorkspace
} from './InventorySummaryRepository';

class FakeInventorySummaryRepository implements HomeDashboardSnapshotRepository {
  constructor(private readonly workspace: InventoryWorkspace) {}

  async getHomeDashboardSnapshot(): Promise<HomeDashboardSnapshot> {
    const inventory =
      this.workspace.inventories.find(
        (inventory) => inventory.id === this.workspace.defaultInventoryId
      ) ?? this.workspace.inventories[0];

    if (!inventory) {
      throw new Error('Fake workspace must include at least one inventory.');
    }

    return {
      workspace: this.workspace,
      checkedOutAssets: inventory.assets.filter((asset) => asset.currentCheckout)
    };
  }
}

describe('HomeDashboardQuery', () => {
  it('builds a dashboard view model from the default inventory summary', async () => {
    const query = new HomeDashboardQuery(
      new FakeInventorySummaryRepository({
        tenants: [
          { id: tenantId('tenant-home'), name: 'Ksell Household' },
          { id: tenantId('tenant-cabin'), name: 'Ksell Cabin' }
        ],
        defaultInventoryId: inventoryId('inventory-home'),
        inventories: [
          {
            id: inventoryId('inventory-home'),
            tenantId: tenantId('tenant-home'),
            name: 'Home',
            role: 'owner',
            permissions: ['view', 'create_asset', 'edit_asset', 'share', 'configure'],
            description: 'Everyday household items.',
            updatedAtLabel: 'Updated 12 min ago',
            locationCount: 2,
            assetTags: [
              {
                id: 'tag-workshop',
                key: 'workshop',
                displayName: 'Workshop',
                color: '#2F80ED'
              }
            ],
            locations: [
              {
                id: assetId('asset-kitchen'),
                inventoryId: inventoryId('inventory-home'),
                title: 'Kitchen',
                description: 'Pantry, utility drawer, and cleaning supplies.',
                containedAssetCount: 12,
                recentAssetTitles: ['AA batteries', 'LED bulbs'],
                hasPhoto: true
              }
            ],
            assets: [
              {
                id: assetId('asset-fresh-batteries'),
                title: 'Fresh batteries',
                kind: 'item',
                lifecycleState: 'active',
                locationLabel: 'Utility drawer',
                locationTrail: ['Home', 'Kitchen', 'Utility drawer'],
                parentLocationTrail: [{ id: assetId('asset-kitchen'), title: 'Kitchen' }],
                description: 'AA batteries just added from the Add sheet.',
                updatedAtLabel: 'Updated just now',
                hasPhoto: false
              },
              {
                id: assetId('asset-kitchen'),
                title: 'Kitchen',
                kind: 'location',
                lifecycleState: 'active',
                locationLabel: 'Home',
                locationTrail: ['Home'],
                parentLocationTrail: [],
                description: 'Main household location.',
                updatedAtLabel: 'Updated today',
                hasPhoto: true
              },
              {
                id: assetId('asset-camera-bag'),
                title: 'Camera bag',
                kind: 'container',
                lifecycleState: 'active',
                locationLabel: 'Hall closet',
                locationTrail: ['Home', 'Hall closet'],
                parentLocationTrail: [],
                description: 'Camera kit and accessories.',
                updatedAtLabel: 'Updated yesterday',
                hasPhoto: false,
                currentCheckout: {
                  id: 'checkout-camera-bag',
                  state: 'open',
                  checkedOutAt: '2026-06-24T11:00:00Z',
                  checkedOutByPrincipalId: 'user-one'
                }
              },
              {
                id: assetId('asset-old-router'),
                title: 'Old router',
                kind: 'item',
                lifecycleState: 'archived',
                locationLabel: 'Office bin',
                locationTrail: ['Home', 'Office', 'Office bin'],
                parentLocationTrail: [{ id: assetId('asset-office'), title: 'Office' }],
                description: 'Retired network hardware.',
                updatedAtLabel: 'Updated last week',
                hasPhoto: true
              }
            ]
          },
          {
            id: inventoryId('inventory-cabin'),
            tenantId: tenantId('tenant-cabin'),
            name: 'Cabin',
            role: 'viewer',
            permissions: ['view'],
            description: 'Seasonal supplies.',
            updatedAtLabel: 'Updated yesterday',
            locationCount: 0,
            locations: [],
            assets: []
          }
        ]
      })
    );

    const dashboard = await query.execute();

    expect(dashboard.tenantName).toBe('Ksell Household');
    expect(dashboard.tenantId).toBe('tenant-home');
    expect(dashboard.inventoryId).toBe('inventory-home');
    expect(dashboard.inventoryName).toBe('Home');
    expect(dashboard.canAdd).toBe(true);
    expect(dashboard.tenants).toEqual([
      { id: 'tenant-home', name: 'Ksell Household' },
      { id: 'tenant-cabin', name: 'Ksell Cabin' }
    ]);
    expect(dashboard.inventories).toEqual([
      {
        id: 'inventory-home',
        tenantId: 'tenant-home',
        tenantName: 'Ksell Household',
        name: 'Home',
        roleLabel: 'Owner',
        updatedAtLabel: 'Updated 12 min ago'
      },
      {
        id: 'inventory-cabin',
        tenantId: 'tenant-cabin',
        tenantName: 'Ksell Cabin',
        name: 'Cabin',
        roleLabel: 'Viewer',
        updatedAtLabel: 'Updated yesterday'
      }
    ]);
    expect(dashboard).not.toHaveProperty('topLocations');
    expect(dashboard).not.toHaveProperty('locations');
    expect(dashboard.assetTags).toEqual([
      {
        id: 'tag-workshop',
        key: 'workshop',
        displayName: 'Workshop',
        color: '#2F80ED'
      }
    ]);
    expect(dashboard.recentAssets).toEqual([
      {
        id: 'asset-fresh-batteries',
        title: 'Fresh batteries',
        kindLabel: 'Item',
        customTypeLabel: undefined,
        description: 'AA batteries just added from the Add sheet.',
        locationTrailLabel: 'Kitchen / Utility drawer',
        parentLocationTrail: [{ id: 'asset-kitchen', title: 'Kitchen', isImmediateParent: true }],
        updatedAtLabel: 'Updated just now',
        photoLabel: 'Needs photo',
        imagePlaceholderLabel: 'Item'
      },
      {
        id: 'asset-kitchen',
        title: 'Kitchen',
        kindLabel: 'Place',
        customTypeLabel: undefined,
        description: 'Main household location.',
        locationTrailLabel: 'Home',
        parentLocationTrail: [],
        updatedAtLabel: 'Updated today',
        photoLabel: 'Photo ready',
        imagePlaceholderLabel: 'Place'
      },
      {
        id: 'asset-camera-bag',
        title: 'Camera bag',
        kindLabel: 'Container',
        customTypeLabel: undefined,
        description: 'Camera kit and accessories.',
        locationTrailLabel: 'Hall closet',
        parentLocationTrail: [],
        updatedAtLabel: 'Updated yesterday',
        photoLabel: 'Needs photo',
        imagePlaceholderLabel: 'Box',
        checkedOutLabel: 'Checked out'
      },
      {
        id: 'asset-old-router',
        title: 'Old router',
        kindLabel: 'Item',
        customTypeLabel: undefined,
        description: 'Retired network hardware.',
        locationTrailLabel: 'Office / Office bin',
        parentLocationTrail: [{ id: 'asset-office', title: 'Office', isImmediateParent: true }],
        updatedAtLabel: 'Updated last week',
        photoLabel: 'Photo ready',
        imagePlaceholderLabel: 'Item'
      }
    ]);
    expect(dashboard.checkedOutAssets).toEqual([
      {
        id: 'asset-camera-bag',
        title: 'Camera bag',
        kindLabel: 'Container',
        customTypeLabel: undefined,
        description: 'Camera kit and accessories.',
        locationTrailLabel: 'Hall closet',
        parentLocationTrail: [],
        updatedAtLabel: 'Updated yesterday',
        photoLabel: 'Needs photo',
        imagePlaceholderLabel: 'Box',
        checkedOutLabel: 'Checked out'
      }
    ]);
  });
});
