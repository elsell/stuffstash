import { describe, expect, it } from 'vitest';
import type { Inventory, Page, Tenant } from '@stuff-stash/api-client';
import { ApiSettingsScopeRepository } from './ApiSettingsScopeRepository';

const tenants: Page<Tenant> = {
  items: [
    {
      id: 'tenant-home',
      name: 'Home',
      access: { relationship: 'owner', permissions: ['view', 'configure'] }
    },
    {
      id: 'tenant-cabin',
      name: 'Cabin',
      access: { relationship: 'viewer', permissions: ['view'] }
    }
  ],
  pagination: {
    limit: 2,
    nextCursor: null,
    hasMore: false
  }
};

const inventories: Record<string, Inventory> = {
  'inventory-home': {
    id: 'inventory-home',
    tenantId: 'tenant-home',
    name: 'Household',
    access: { relationship: 'owner', permissions: ['view', 'share'] }
  },
  'inventory-cabin': {
    id: 'inventory-cabin',
    tenantId: 'tenant-cabin',
    name: 'Cabin inventory',
    access: { relationship: 'viewer', permissions: ['view'] }
  }
};

describe('ApiSettingsScopeRepository', () => {
  it('resolves the current tenant for every query and preserves its permissions', async () => {
    let currentScope = {
      tenantId: 'tenant-home',
      inventory: { id: 'inventory-home', name: 'Household', permissions: ['view', 'share'] }
    };
    const repository = new ApiSettingsScopeRepository(
      { listMyTenants: async () => tenants },
      { getCurrentSettingsScope: async () => currentScope }
    );

    await expect(repository.getSelectedScope()).resolves.toEqual({
      tenant: { id: 'tenant-home', name: 'Home', permissions: ['view', 'configure'] },
      inventory: { id: 'inventory-home', name: 'Household', permissions: ['view', 'share'] }
    });

    currentScope = {
      tenantId: 'tenant-cabin',
      inventory: { id: 'inventory-cabin', name: 'Cabin inventory', permissions: ['view'] }
    };

    await expect(repository.getSelectedScope()).resolves.toEqual({
      tenant: { id: 'tenant-cabin', name: 'Cabin', permissions: ['view'] },
      inventory: { id: 'inventory-cabin', name: 'Cabin inventory', permissions: ['view'] }
    });
  });

  it('fails safely when the selected tenant is no longer accessible', async () => {
    const repository = new ApiSettingsScopeRepository(
      { listMyTenants: async () => tenants },
      { getCurrentSettingsScope: async () => ({
        tenantId: 'tenant-removed',
        inventory: { id: 'inventory-home', name: 'Household', permissions: ['view', 'share'] }
      }) }
    );

    await expect(repository.getSelectedScope()).rejects.toThrow(
      'The selected Stuff Stash tenant is no longer available.'
    );
  });
});
