import { describe, expect, it } from 'vitest';
import { SeededInventoryRepository } from './seededInventoryRepository';
import type { WorkspaceSeed } from '$lib/ports/inventoryRepository';

const seed: WorkspaceSeed = {
  principal: { id: 'person-one', email: 'person@example.test' },
  tenants: [
    { id: 'tenant-home', name: 'Home', access: { relationship: 'owner', permissions: ['view'] } },
    { id: 'tenant-cabin', name: 'Cabin', access: { relationship: 'editor', permissions: ['view'] } },
    { id: 'tenant-empty', name: 'Empty', access: { relationship: 'owner', permissions: ['view', 'create_inventory'] } },
    { id: 'tenant-viewer-empty', name: 'Viewer Empty', access: { relationship: 'viewer', permissions: ['view'] } }
  ],
  inventories: [
    {
      id: 'inventory-household',
      tenantId: 'tenant-home',
      name: 'Household',
      access: { relationship: 'owner', permissions: ['view', 'create_asset'] }
    },
    {
      id: 'inventory-cabin',
      tenantId: 'tenant-cabin',
      name: 'Cabin Gear',
      access: { relationship: 'viewer', permissions: ['view'] }
    }
  ],
  assets: [
    {
      id: 'asset-home',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'item',
      title: 'Passport',
      description: 'Blue folder',
      parentAssetId: null,
      lifecycleState: 'active'
    },
    {
      id: 'asset-cabin',
      tenantId: 'tenant-cabin',
      inventoryId: 'inventory-cabin',
      kind: 'item',
      title: 'Lantern',
      description: 'Shelf',
      parentAssetId: null,
      lifecycleState: 'active'
    }
  ]
};

describe('SeededInventoryRepository tenant selection', () => {
  it('loads the selected tenant inventories and scopes assets to its first inventory', async () => {
    const repository = new SeededInventoryRepository(seed);

    const data = await repository.selectTenant('tenant-cabin');

    expect(data.context.selectedTenantId).toBe('tenant-cabin');
    expect(data.context.inventories.map((inventory) => inventory.id)).toEqual(['inventory-cabin']);
    expect(data.context.selectedInventoryId).toBe('inventory-cabin');
    expect(data.context.capability).toBe('viewer');
    expect(data.assets.map((asset) => asset.id)).toEqual(['asset-cabin']);
  });

  it('keeps an empty tenant selected without leaking another tenant inventory', async () => {
    const repository = new SeededInventoryRepository(seed);

    const data = await repository.selectTenant('tenant-empty');

    expect(data.context.selectedTenantId).toBe('tenant-empty');
    expect(data.context.inventories).toEqual([]);
    expect(data.context.selectedInventoryId).toBe('');
    expect(data.assets).toEqual([]);
  });

  it('creates a starter inventory inside the selected tenant', async () => {
    const repository = new SeededInventoryRepository(seed);
    await repository.selectTenant('tenant-empty');

    const data = await repository.createInventory('tenant-empty', 'Household');

    expect(data.context.selectedTenantId).toBe('tenant-empty');
    expect(data.context.inventories).toMatchObject([{ tenantId: 'tenant-empty', name: 'Household' }]);
    expect(data.context.selectedInventoryId).toBe(data.context.inventories[0]?.id);
  });

  it('rejects starter inventory creation when the selected tenant lacks permission', async () => {
    const repository = new SeededInventoryRepository(seed);
    await repository.selectTenant('tenant-viewer-empty');

    await expect(repository.createInventory('tenant-viewer-empty', 'Household')).rejects.toThrow(
      'You do not have permission'
    );
  });

  it('does not leak assets when inventory selection is mismatched across tenants', async () => {
    const repository = new SeededInventoryRepository(seed);

    const data = await repository.selectInventory('tenant-home', 'inventory-cabin');

    expect(data.context.selectedTenantId).toBe('tenant-home');
    expect(data.context.selectedInventoryId).toBe('inventory-household');
    expect(data.assets.map((asset) => asset.id)).toEqual(['asset-home']);
    await expect(repository.searchAssets('tenant-home', 'Lantern')).resolves.toEqual([]);
  });

  it('loads and updates asset detail inside the selected inventory', async () => {
    const repository = new SeededInventoryRepository(seed);

    await expect(repository.getAsset('tenant-home', 'inventory-household', 'asset-home')).resolves.toMatchObject({
      id: 'asset-home',
      title: 'Passport'
    });

    const updated = await repository.updateAsset('tenant-home', 'inventory-household', 'asset-home', {
      title: 'Updated Passport',
      description: 'Fire safe',
      parentAssetId: null
    });

    expect(updated).toMatchObject({
      id: 'asset-home',
      title: 'Updated Passport',
      description: 'Fire safe',
      parentAssetId: null
    });
    await expect(repository.getAsset('tenant-home', 'inventory-household', 'asset-home')).resolves.toMatchObject({
      title: 'Updated Passport'
    });
  });
});
