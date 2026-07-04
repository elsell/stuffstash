import { describe, expect, it } from 'vitest';
import type { Inventory } from '$lib/domain/inventory';
import { contextInventoryHref, importSourceHref, shellAddHref, shellModeHref } from './workspaceShellNavigation';

const inventory: Inventory = {
  id: 'inventory-one',
  tenantId: 'tenant-one',
  name: 'Garage',
  access: { relationship: 'owner', permissions: ['view'] }
};

describe('workspace shell navigation helpers', () => {
  it('derives durable shell mode hrefs', () => {
    expect(shellModeHref('home', 'tenant-one', 'inventory-one')).toBe('/tenants/tenant-one/inventories/inventory-one');
    expect(shellModeHref('locations', 'tenant-one', 'inventory-one')).toBe('/tenants/tenant-one/inventories/inventory-one/locations');
    expect(shellModeHref('search', 'tenant-one', 'inventory-one')).toBe('/tenants/tenant-one/inventories/inventory-one/search');
    expect(shellModeHref('import', 'tenant-one', 'inventory-one')).toBe('/tenants/tenant-one/inventories/inventory-one/import');
    expect(shellModeHref('settings', 'tenant-one', 'inventory-one', 'activity')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings/activity'
    );
  });

  it('derives add hrefs for shell add controls', () => {
    expect(shellAddHref('item', 'tenant-one', 'inventory-one')).toBe('/tenants/tenant-one/inventories/inventory-one/add/item');
    expect(shellAddHref('container', 'tenant-one', 'inventory-one')).toBe('/tenants/tenant-one/inventories/inventory-one/add/container');
    expect(shellAddHref('location', 'tenant-one', 'inventory-one')).toBe('/tenants/tenant-one/inventories/inventory-one/add/location');
  });

  it('derives inventory and import source hrefs', () => {
    expect(contextInventoryHref(inventory)).toBe('/tenants/tenant-one/inventories/inventory-one');
    expect(importSourceHref('tenant-one', 'inventory-one', 'legacy_homebox')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/import/legacy-homebox'
    );
    expect(importSourceHref('tenant-one', 'inventory-one', 'legacy_homebox_csv')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/import/legacy-homebox-csv'
    );
  });
});
