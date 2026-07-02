import { describe, expect, it } from 'vitest';
import { parseWorkspaceRoute, workspaceRouteHref } from './workspaceRoute';

describe('workspace route state', () => {
  it('parses a tenant inventory location deep link', () => {
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/locations/loc_1'))).toMatchObject({
      mode: 'location',
      tenantId: 'tenant_1',
      inventoryId: 'inv_1',
      locationId: 'loc_1'
    });
  });

  it('parses asset edit and search filters', () => {
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/assets/asset_1/edit'))).toMatchObject({
      mode: 'asset',
      tenantId: 'tenant_1',
      inventoryId: 'inv_1',
      assetId: 'asset_1',
      action: 'edit'
    });

    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/inv_1/search?q=drill&lifecycle=all&mode=exact'))).toMatchObject({
      mode: 'search',
      tenantId: 'tenant_1',
      searchQuery: 'drill',
      searchLifecycleState: 'all',
      searchMode: 'exact'
    });
  });

  it('formats stable workspace hrefs', () => {
    expect(workspaceRouteHref({ mode: 'asset', assetId: 'asset 1', action: 'edit' }, 'tenant 1', 'inv 1')).toBe(
      '/tenants/tenant%201/inventories/inv%201/assets/asset%201/edit'
    );
    expect(workspaceRouteHref({ mode: 'search', searchQuery: 'garage shelf', searchLifecycleState: 'archived' }, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/search?q=garage+shelf&lifecycle=archived'
    );
    expect(workspaceRouteHref({ action: 'add', addKind: 'location' }, 'tenant_1', 'inv_1')).toBe(
      '/tenants/tenant_1/inventories/inv_1/add/location'
    );
  });

  it('accepts inventory-only compatibility aliases', () => {
    expect(parseWorkspaceRoute(new URL('https://app.test/inventories/inv_1/assets/asset_1/edit'))).toMatchObject({
      mode: 'asset',
      tenantId: null,
      inventoryId: 'inv_1',
      assetId: 'asset_1',
      action: 'edit'
    });

    expect(workspaceRouteHref({ mode: 'search', inventoryId: 'inv_1', searchQuery: 'drill' }, null, null)).toBe(
      '/inventories/inv_1/search?q=drill'
    );
  });

  it('falls back for malformed encoded paths', () => {
    expect(parseWorkspaceRoute(new URL('https://app.test/tenants/tenant_1/inventories/%E0%A4%A'))).toMatchObject({
      mode: 'home',
      tenantId: null,
      inventoryId: null
    });
  });
});
