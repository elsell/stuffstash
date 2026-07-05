import { describe, expect, it } from 'vitest';
import type { Asset, LocationAsset } from '$lib/domain/inventory';
import {
  browseAssetHref,
  browseLocationHref,
  homeAddItemHref,
  homeArchivedEmptyState,
  homeAddLocationHref,
  homeCreateLocationDenied,
  homeHeadingPresentation,
  homeLifecycleHref,
  homeLifecycleOptions,
  homeLocationsEmptyState,
  homeRecentEmptyState,
  locationAddItemHref,
  locationBackHref,
  locationEditHref,
  locationEmptyState,
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
    expect(homeAddItemHref('tenant-home', 'inventory-household')).toBe('/tenants/tenant-home/inventories/inventory-household/add/item');
    expect(homeLifecycleHref('tenant-home', 'inventory-household', 'active')).toBe('/tenants/tenant-home/inventories/inventory-household');
    expect(homeLifecycleHref('tenant-home', 'inventory-household', 'archived')).toBe(
      '/tenants/tenant-home/inventories/inventory-household?lifecycle=archived'
    );
    expect(homeLifecycleOptions('tenant-home', 'inventory-household')).toEqual([
      { value: 'active', label: 'Active', href: '/tenants/tenant-home/inventories/inventory-household' },
      {
        value: 'archived',
        label: 'Archived',
        href: '/tenants/tenant-home/inventories/inventory-household?lifecycle=archived'
      }
    ]);
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

  it('builds home heading, empty, and denied presentation', () => {
    expect(homeHeadingPresentation('active', 'home')).toEqual({
      title: 'Home',
      description: 'Recently added and the places where your things live.'
    });
    expect(homeHeadingPresentation('active', 'locations')).toEqual({
      title: 'Locations',
      description: 'The places where your things live.'
    });
    expect(homeHeadingPresentation('archived', 'home')).toEqual({
      title: 'Archived assets',
      description: 'Assets removed from active browsing.'
    });
    expect(homeRecentEmptyState()).toEqual({ message: 'No items or containers yet.' });
    expect(homeArchivedEmptyState()).toEqual({ title: 'No archived assets' });
    expect(homeLocationsEmptyState()).toEqual({
      title: 'No locations yet',
      message: 'Locations make browsing easier, but you can capture an item now.',
      actionLabel: 'Add first location',
      secondaryActionLabel: 'Add item'
    });
    expect(homeLocationsEmptyState('locations')).toEqual({
      title: 'No locations yet',
      message: 'Add a location to start browsing by place.',
      actionLabel: 'Add first location'
    });
    expect(homeCreateLocationDenied()).toEqual({
      id: 'home-add-location-denied',
      message: 'Creating locations is unavailable for this inventory.'
    });
  });

  it('builds focused location empty and denied presentation', () => {
    expect(locationEmptyState(true)).toEqual({
      title: 'No stuff here yet',
      message: 'Add an item or move existing stuff into this location.',
      actionLabel: 'Add item here',
      deniedMessage: 'Adding items is unavailable for this inventory.'
    });
    expect(locationEmptyState(false)).toEqual({
      title: 'No stuff here yet',
      message: 'This location is empty.',
      actionLabel: 'Add item here',
      deniedMessage: 'Adding items is unavailable for this inventory.'
    });
  });
});
