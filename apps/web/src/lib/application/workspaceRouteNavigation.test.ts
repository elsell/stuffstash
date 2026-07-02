import { afterEach, describe, expect, it } from 'vitest';
import type { Asset, WorkspaceData } from '$lib/domain/inventory';
import {
  assetRouteActionIsAvailable,
  currentWorkspaceRoute,
  findRouteInventory,
  findRouteTenant,
  pushWorkspaceRoute,
  replaceCanonicalWorkspaceAlias,
  replaceWorkspaceRoute,
  shouldCanonicalizeWorkspaceAlias
} from './workspaceRouteNavigation';
import { parseWorkspaceRoute } from './workspaceRoute';

const workspaceData: WorkspaceData = {
  context: {
    principal: { id: 'principal-1', email: 'owner@example.com' },
    tenants: [
      { id: 'tenant-home', name: 'Home', access: { relationship: 'owner', permissions: [] } },
      { id: 'tenant-shop', name: 'Shop', access: { relationship: 'owner', permissions: [] } }
    ],
    inventories: [
      {
        id: 'inventory-household',
        tenantId: 'tenant-home',
        name: 'Household',
        access: { relationship: 'editor', permissions: ['edit_asset'] }
      },
      {
        id: 'inventory-shop',
        tenantId: 'tenant-shop',
        name: 'Shop',
        access: { relationship: 'viewer', permissions: [] }
      }
    ],
    selectedTenantId: 'tenant-home',
    selectedInventoryId: 'inventory-household',
    assetLifecycleState: 'active',
    customAssetTypes: [],
    customFieldDefinitions: [],
    mediaUploadPolicy: { supportedContentTypes: ['image/jpeg'], maxBytes: 1024 },
    capability: 'editor'
  },
  assets: []
};

function asset(lifecycleState: Asset['lifecycleState']): Asset {
  return {
    id: 'asset-1',
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    kind: 'item',
    title: 'Tape measure',
    description: '',
    parentAssetId: null,
    lifecycleState
  };
}

afterEach(() => {
  window.history.replaceState({}, '', '/');
});

describe('workspace route navigation helpers', () => {
  it('pushes route state through the canonical route contract', () => {
    const route = pushWorkspaceRoute(
      { mode: 'asset', assetId: 'asset 1', assetAction: 'edit' },
      'tenant home',
      'inventory 1'
    );

    expect(window.location.pathname).toBe('/tenants/tenant%20home/inventories/inventory%201/assets/asset%201/edit');
    expect(route).toMatchObject({
      mode: 'asset',
      tenantId: 'tenant home',
      inventoryId: 'inventory 1',
      assetId: 'asset 1',
      assetAction: 'edit'
    });
  });

  it('keeps route helpers safe without a browser target', () => {
    expect(currentWorkspaceRoute(null)).toMatchObject({ mode: 'home' });

    const pushed = pushWorkspaceRoute({ mode: 'settings', settingsSection: 'access' }, 'tenant-home', 'inventory-household', null);
    replaceWorkspaceRoute({ mode: 'search', searchQuery: 'tape' }, 'tenant-home', 'inventory-household', null);

    expect(pushed).toMatchObject({ mode: 'settings', settingsSection: 'access' });
    expect(window.location.pathname).toBe('/');
  });

  it('replaces inventory aliases with tenant-scoped canonical routes', () => {
    const alias = parseWorkspaceRoute(new URL('https://app.test/inventories/inventory-household/search?q=tape'));

    expect(shouldCanonicalizeWorkspaceAlias(alias)).toBe(true);

    replaceCanonicalWorkspaceAlias(alias, 'tenant-home', 'inventory-household');

    expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/search');
    expect(window.location.search).toBe('?q=tape');
  });

  it('resolves route tenant and inventory changes against visible workspace data', () => {
    expect(findRouteTenant(workspaceData, parseWorkspaceRoute(new URL('https://app.test/tenants/tenant-shop/inventories/inventory-shop')))).toBe(
      'tenant-shop'
    );
    const routeInventory = findRouteInventory(
      workspaceData,
      parseWorkspaceRoute(new URL('https://app.test/tenants/tenant-shop/inventories/inventory-shop'))
    );
    expect(routeInventory?.id).toBe('inventory-shop');
    expect(findRouteTenant(workspaceData, parseWorkspaceRoute(new URL('https://app.test/tenants/nope/inventories/inventory-shop')))).toBeNull();
    expect(
      findRouteInventory(workspaceData, parseWorkspaceRoute(new URL('https://app.test/tenants/tenant-home/inventories/inventory-shop')))
    ).toBeNull();
  });

  it('keeps asset action route availability explicit', () => {
    expect(assetRouteActionIsAvailable('edit', workspaceData.context.inventories[0], asset('active'))).toBe(true);
    expect(assetRouteActionIsAvailable('move', workspaceData.context.inventories[0], asset('archived'))).toBe(false);
    expect(assetRouteActionIsAvailable('delete', workspaceData.context.inventories[0], asset('archived'))).toBe(true);
    expect(assetRouteActionIsAvailable('edit', workspaceData.context.inventories[1], asset('active'))).toBe(false);
  });
});
