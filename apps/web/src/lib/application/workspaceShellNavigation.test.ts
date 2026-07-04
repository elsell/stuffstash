import { describe, expect, it } from 'vitest';
import type { Inventory } from '$lib/domain/inventory';
import {
  contextInventoryHref,
  desktopShellNavigationGroups,
  importSourceHref,
  mobileShellNavigationItems,
  shellAddOptions,
  shellAddHref,
  shellModeHref,
  shellModeIsCurrent
} from './workspaceShellNavigation';

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

  it('builds shell add menu options with kind labels and durable hrefs', () => {
    expect(shellAddOptions('tenant-one', 'inventory-one')).toEqual([
      { kind: 'item', label: 'Item', href: '/tenants/tenant-one/inventories/inventory-one/add/item' },
      { kind: 'container', label: 'Container', href: '/tenants/tenant-one/inventories/inventory-one/add/container' },
      { kind: 'location', label: 'Location', href: '/tenants/tenant-one/inventories/inventory-one/add/location' }
    ]);
  });

  it('builds grouped desktop navigation destinations with current-state rules', () => {
    expect(
      desktopShellNavigationGroups({
        mode: 'location',
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        settingsSection: 'activity'
      })
    ).toEqual([
      {
        id: 'primary',
        label: 'Inventory',
        destinations: [
          {
            mode: 'home',
            label: 'Home',
            description: 'Recent assets and places',
            icon: 'home',
            href: '/tenants/tenant-one/inventories/inventory-one',
            current: false
          },
          {
            mode: 'locations',
            label: 'Locations',
            description: 'Browse rooms, shelves, and places',
            icon: 'locations',
            href: '/tenants/tenant-one/inventories/inventory-one/locations',
            current: true
          }
        ]
      },
      {
        id: 'utility',
        label: 'Tools',
        destinations: [
          {
            mode: 'import',
            label: 'Import',
            description: 'Bring in legacy data',
            icon: 'import',
            href: '/tenants/tenant-one/inventories/inventory-one/import',
            current: false
          },
          {
            mode: 'settings',
            label: 'Settings',
            description: 'Access, fields, and audit',
            icon: 'settings',
            href: '/tenants/tenant-one/inventories/inventory-one/settings/activity',
            current: false
          }
        ]
      }
    ]);
  });

  it('builds mobile navigation items without desktop-only utility grouping', () => {
    expect(
      mobileShellNavigationItems({
        mode: 'search',
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        settingsSection: 'overview'
      })
    ).toEqual([
      {
        mode: 'home',
        label: 'Home',
        description: 'Inventory home',
        icon: 'home',
        href: '/tenants/tenant-one/inventories/inventory-one',
        current: false
      },
      {
        mode: 'search',
        label: 'Search',
        description: 'Find assets',
        icon: 'search',
        href: '/tenants/tenant-one/inventories/inventory-one/search',
        current: true
      },
      {
        mode: 'locations',
        label: 'Places',
        description: 'Browse places',
        icon: 'locations',
        href: '/tenants/tenant-one/inventories/inventory-one/locations',
        current: false
      },
      {
        mode: 'settings',
        label: 'Settings',
        description: 'Inventory settings',
        icon: 'settings',
        href: '/tenants/tenant-one/inventories/inventory-one/settings',
        current: false
      }
    ]);
  });

  it('treats focused location routes as current for the locations destination', () => {
    expect(shellModeIsCurrent('location', 'locations')).toBe(true);
    expect(shellModeIsCurrent('asset', 'locations')).toBe(false);
    expect(shellModeIsCurrent('settings', 'settings')).toBe(true);
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
