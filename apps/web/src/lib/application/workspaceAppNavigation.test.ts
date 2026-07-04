import { describe, expect, it } from 'vitest';
import type { WorkspaceContext } from '$lib/domain/inventory';
import { defaultWorkspaceRoute, type WorkspaceRouteState } from './workspaceRoute';
import {
  assetDetailBackHref,
  assetDetailBackRoute,
  inventoryHomeNormalizationHref,
  inventoryHomeNormalizationRoute,
  settingsOverviewHref,
  settingsOverviewRoute,
  workspaceAddCloseHref,
  workspaceAddCloseRoute,
  workspaceHomeHref,
  workspaceHomeRoute
} from './workspaceAppNavigation';

const context: WorkspaceContext = {
  principal: { id: 'principal-one', email: 'owner@example.test' },
  tenants: [{ id: 'tenant-home', name: 'Home', access: { relationship: 'owner', permissions: ['view'] } }],
  inventories: [
    {
      id: 'inventory-household',
      tenantId: 'tenant-home',
      name: 'Household',
      access: { relationship: 'owner', permissions: ['view'] }
    }
  ],
  selectedTenantId: 'tenant-home',
  selectedInventoryId: 'inventory-household',
  assetLifecycleState: 'archived',
  mediaUploadPolicy: { supportedContentTypes: ['image/jpeg'], maxBytes: 1024 },
  customAssetTypes: [],
  customFieldDefinitions: [],
  capability: 'editor'
};

describe('workspace app navigation helpers', () => {
  it('derives the selected inventory home route and href', () => {
    expect(workspaceHomeRoute(context)).toEqual({
      mode: 'home',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      lifecycleState: 'archived'
    });
    expect(workspaceHomeHref(context)).toBe('/tenants/tenant-home/inventories/inventory-household?lifecycle=archived');
  });

  it('derives add close destinations from the current mode', () => {
    expect(workspaceAddCloseRoute(context, { mode: 'search', selectedLocationId: null, selectedAssetId: null })).toEqual({
      mode: 'search',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      lifecycleState: 'archived'
    });
    expect(workspaceAddCloseHref(context, { mode: 'search', selectedLocationId: null, selectedAssetId: null })).toBe(
      '/tenants/tenant-home/inventories/inventory-household/search'
    );
    expect(workspaceAddCloseRoute(context, { mode: 'home', selectedLocationId: 'location-garage', selectedAssetId: null })).toEqual({
      mode: 'location',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      locationId: 'location-garage'
    });
    expect(workspaceAddCloseHref(context, { mode: 'home', selectedLocationId: 'location-garage', selectedAssetId: null })).toBe(
      '/tenants/tenant-home/inventories/inventory-household/locations/location-garage'
    );
    expect(workspaceAddCloseRoute(context, { mode: 'home', selectedLocationId: null, selectedAssetId: 'asset-passport' })).toEqual({
      mode: 'asset',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      assetId: 'asset-passport'
    });
    expect(workspaceAddCloseHref(context, { mode: 'home', selectedLocationId: null, selectedAssetId: 'asset-passport' })).toBe(
      '/tenants/tenant-home/inventories/inventory-household/assets/asset-passport'
    );
  });

  it('derives canonical settings overview and home normalization destinations', () => {
    const activeHomeRoute: WorkspaceRouteState = {
      ...defaultWorkspaceRoute,
      mode: 'home',
      lifecycleState: 'active',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household'
    };

    expect(settingsOverviewRoute(context)).toEqual({
      mode: 'settings',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      settingsSection: 'overview'
    });
    expect(settingsOverviewHref(context)).toBe('/tenants/tenant-home/inventories/inventory-household/settings');
    expect(inventoryHomeNormalizationRoute(context, activeHomeRoute)).toEqual({
      mode: 'home',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      lifecycleState: 'active'
    });
    expect(inventoryHomeNormalizationHref(context, activeHomeRoute)).toBe('/tenants/tenant-home/inventories/inventory-household');
  });

  it('derives asset detail back destinations from the previous location context', () => {
    expect(assetDetailBackRoute(context, 'location-garage')).toEqual({
      mode: 'location',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      locationId: 'location-garage'
    });
    expect(assetDetailBackHref(context, 'location-garage')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/locations/location-garage'
    );
    expect(assetDetailBackHref(context, null)).toBe('/tenants/tenant-home/inventories/inventory-household?lifecycle=archived');
  });
});
