import { describe, expect, it } from 'vitest';
import type { Inventory } from '$lib/domain/inventory';
import {
  accountDisplayLabel,
  contextInventoryHref,
  desktopShellNavigationGroups,
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
  access: { relationship: 'owner', permissions: ['view', 'view_import_job', 'create_import_job'] }
};

describe('workspace shell navigation helpers', () => {
  it('uses verified email for account copy without leaking an opaque principal ID', () => {
    expect(accountDisplayLabel({ id: 'principal-secret', email: 'owner@example.com' })).toBe('owner@example.com');
    expect(accountDisplayLabel({ id: 'principal-secret' })).toBe('Signed-in account');
    expect(accountDisplayLabel({ id: 'principal-secret', email: '   ' })).toBe('Signed-in account');
  });

  it('derives durable shell mode hrefs', () => {
    expect(shellModeHref('home', 'tenant-one', 'inventory-one')).toBe('/tenants/tenant-one/inventories/inventory-one');
    expect(shellModeHref('browse', 'tenant-one', 'inventory-one')).toBe('/tenants/tenant-one/inventories/inventory-one/browse');
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
        inventory,
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
            mode: 'browse',
            label: 'Browse',
            description: 'Find and explore your inventory',
            icon: 'browse',
            href: '/tenants/tenant-one/inventories/inventory-one/browse',
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
            description: 'Bring in outside data',
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

  it('omits import navigation when the inventory lacks import view access', () => {
    const viewerInventory: Inventory = {
      ...inventory,
      access: { relationship: 'viewer', permissions: ['view'] }
    };

    const groups = desktopShellNavigationGroups({
      mode: 'home',
      tenantId: 'tenant-one',
      inventoryId: 'inventory-one',
      inventory: viewerInventory,
      settingsSection: 'overview'
    });

    expect(groups[1].destinations.map((destination) => destination.mode)).toEqual(['settings']);
  });

  it('builds mobile navigation items without desktop-only utility grouping', () => {
    expect(
      mobileShellNavigationItems({
        mode: 'browse',
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
        mode: 'browse',
        label: 'Browse',
        description: 'Find and explore',
        icon: 'browse',
        href: '/tenants/tenant-one/inventories/inventory-one/browse',
        current: true
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

  it('treats focused location routes as current for Browse', () => {
    expect(shellModeIsCurrent('location', 'browse')).toBe(true);
    expect(shellModeIsCurrent('asset', 'browse')).toBe(false);
    expect(shellModeIsCurrent('settings', 'settings')).toBe(true);
  });

  it('derives inventory hrefs', () => {
    expect(contextInventoryHref(inventory)).toBe('/tenants/tenant-one/inventories/inventory-one');
  });
});
