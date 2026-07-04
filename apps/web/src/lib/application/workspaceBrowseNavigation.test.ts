import { describe, expect, it } from 'vitest';
import type { Asset, LocationAsset } from '$lib/domain/inventory';
import {
  browseAssetHref,
  browseLocationHref,
  homeAddLocationHref,
  homeLifecycleHref,
  locationAddItemHref,
  locationBackHref,
  locationEditHref,
  locationRowHref
} from './workspaceBrowseNavigation';

function asset(id: string, kind: Asset['kind'] = 'item'): Asset {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    kind,
    title: id,
    description: '',
    parentAssetId: null,
    lifecycleState: 'active'
  };
}

function locationAsset(id: string): LocationAsset {
  return { ...asset(id, 'location'), kind: 'location' };
}

describe('workspace browse navigation helpers', () => {
  it('derives home action and lifecycle hrefs', () => {
    expect(homeAddLocationHref('tenant-home', 'inventory-household')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/add/location'
    );
    expect(homeLifecycleHref('tenant-home', 'inventory-household', 'active')).toBe('/tenants/tenant-home/inventories/inventory-household');
    expect(homeLifecycleHref('tenant-home', 'inventory-household', 'archived')).toBe(
      '/tenants/tenant-home/inventories/inventory-household?lifecycle=archived'
    );
  });

  it('derives home asset and location row hrefs', () => {
    expect(browseAssetHref(asset('tape'))).toBe('/tenants/tenant-home/inventories/inventory-household/assets/tape');
    expect(browseLocationHref(locationAsset('garage'))).toBe('/tenants/tenant-home/inventories/inventory-household/locations/garage');
  });

  it('derives focused location action and row hrefs', () => {
    const garage = locationAsset('garage');

    expect(locationBackHref(garage)).toBe('/tenants/tenant-home/inventories/inventory-household/locations');
    expect(locationEditHref(garage)).toBe('/tenants/tenant-home/inventories/inventory-household/locations/garage/edit');
    expect(locationAddItemHref(garage)).toBe('/tenants/tenant-home/inventories/inventory-household/add/item?parent=garage');
    expect(locationRowHref(locationAsset('shelf'))).toBe('/tenants/tenant-home/inventories/inventory-household/locations/shelf');
    expect(locationRowHref(asset('tape'))).toBe('/tenants/tenant-home/inventories/inventory-household/assets/tape');
  });
});
