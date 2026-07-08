import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import HomeWorkspace from './HomeWorkspace.svelte';
import type { AssetViewModel, LocationAsset } from '$lib/domain/inventory';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('HomeWorkspace', () => {
  it('shows containment context once for descriptionless recent assets', () => {
    const asset: AssetViewModel = {
      id: 'tape',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'item',
      title: 'Tape measure',
      description: '',
      parentAssetId: 'garage',
      lifecycleState: 'active',
      containmentTrail: 'Garage'
    };

    component = mount(HomeWorkspace, {
      target: document.body,
      props: {
        lifecycleState: 'active',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-household',
        locations: [],
        recentAssets: [asset],
        archivedAssets: [],
        onOpenLocation: () => {},
        onOpenAsset: () => {},
        onOpenAdd: () => {},
        onSelectLifecycle: () => {}
      }
    });

    const row = link('Tape measure');
    expect(row?.textContent?.match(/Garage/g)).toHaveLength(1);
    expect(row.getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/assets/tape');
  });

  it('renders a locations-focused browse view without the recent rail', () => {
    const location: LocationAsset = {
      id: 'garage',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'location',
      title: 'Garage',
      description: '',
      parentAssetId: null,
      lifecycleState: 'active',
    };

    component = mount(HomeWorkspace, {
      target: document.body,
      props: {
        lifecycleState: 'active',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-household',
        browseMode: 'locations',
        locations: [{ location, assetCount: 4 }],
        recentAssets: [
          {
            ...location,
            id: 'recent-item',
            kind: 'item',
            title: 'Tape measure',
            parentAssetId: 'garage',
            containmentTrail: 'Garage'
          }
        ],
        archivedAssets: [],
        onOpenLocation: () => {},
        onOpenAsset: () => {},
        onOpenAdd: () => {},
        onSelectLifecycle: () => {}
      }
    });

    expect(document.body.textContent).toContain('Locations');
    expect(document.body.textContent).toContain('The places where your things live.');
    expect(document.body.textContent).toContain('Garage');
    expect(document.body.textContent).not.toContain('Recently added');
    expect(document.body.textContent).not.toContain('Tape measure');
    expect(document.body.querySelector('[aria-label="Asset lifecycle"]')).toBeNull();
    expect(link('Garage').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/locations/garage');
  });

  it('exposes durable hrefs for home actions, location tiles, and archived rows', () => {
    const location: LocationAsset = {
      id: 'garage',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'location',
      title: 'Garage',
      description: '',
      parentAssetId: null,
      lifecycleState: 'active',
    };
    const archived: AssetViewModel = {
      ...location,
      id: 'old-drill',
      kind: 'item',
      title: 'Old drill',
      lifecycleState: 'archived',
      containmentTrail: 'Garage'
    };

    component = mount(HomeWorkspace, {
      target: document.body,
      props: {
        lifecycleState: 'active',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-household',
        locations: [{ location, assetCount: 4 }],
        recentAssets: [],
        archivedAssets: [],
        onOpenLocation: () => {},
        onOpenAsset: () => {},
        onOpenAdd: () => {},
        onSelectLifecycle: () => {}
      }
    });

    expect(link('Add location').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/add/location');
    expect(link('Garage').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/locations/garage');

    unmount(component);
    component = null;
    document.body.innerHTML = '';

    component = mount(HomeWorkspace, {
      target: document.body,
      props: {
        lifecycleState: 'archived',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-household',
        locations: [{ location, assetCount: 4 }],
        recentAssets: [],
        archivedAssets: [archived],
        onOpenLocation: () => {},
        onOpenAsset: () => {},
        onOpenAdd: () => {},
        onSelectLifecycle: () => {}
      }
    });

    expect(link('Old drill').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/assets/old-drill');
  });

  it('exposes lifecycle filter hrefs and preserves modified clicks', () => {
    let selectedLifecycle = '';
    component = mount(HomeWorkspace, {
      target: document.body,
      props: {
        lifecycleState: 'active',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-household',
        locations: [],
        recentAssets: [],
        archivedAssets: [],
        onOpenLocation: () => {},
        onOpenAsset: () => {},
        onOpenAdd: () => {},
        onSelectLifecycle: (lifecycle) => {
          selectedLifecycle = lifecycle;
        }
      }
    });

    expect(link('Active').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household');
    expect(link('Archived').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household?lifecycle=archived');

    link('Archived').click();
    expect(selectedLifecycle).toBe('archived');

    selectedLifecycle = '';
    let componentPreventedModifiedClick = true;
    const target = link('Archived');
    target.addEventListener('click', (event) => {
      componentPreventedModifiedClick = event.defaultPrevented;
      event.preventDefault();
    });
    target.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true, metaKey: true }));

    expect(selectedLifecycle).toBe('');
    expect(componentPreventedModifiedClick).toBe(false);
  });

  it('uses selected context for empty-inventory add links', () => {
    const openedKinds: Array<'item' | 'location' | undefined> = [];
    component = mount(HomeWorkspace, {
      target: document.body,
      props: {
        lifecycleState: 'active',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-household',
        locations: [],
        recentAssets: [],
        archivedAssets: [],
        onOpenLocation: () => {},
        onOpenAsset: () => {},
        onOpenAdd: (kind) => {
          openedKinds.push(kind);
        },
        onSelectLifecycle: () => {}
      }
    });

    expect(document.body.textContent).toContain('Locations make browsing easier, but you can capture an item now.');
    expect(link('Add first location').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/add/location');
    expect(link('Add item').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/add/item');

    link('Add item').click();
    expect(openedKinds).toEqual(['item']);
  });

  it('keeps the empty locations route focused on creating a location', () => {
    component = mount(HomeWorkspace, {
      target: document.body,
      props: {
        lifecycleState: 'active',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-household',
        browseMode: 'locations',
        locations: [],
        recentAssets: [],
        archivedAssets: [],
        onOpenLocation: () => {},
        onOpenAsset: () => {},
        onOpenAdd: () => {},
        onSelectLifecycle: () => {}
      }
    });

    expect(document.body.textContent).toContain('Add a location to start browsing by place.');
    expect(link('Add first location').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/add/location');
    expect(document.body.textContent).not.toContain('Add item');
  });

  it('disables home add-location controls when creation is unavailable', () => {
    let openedAdd = false;
    component = mount(HomeWorkspace, {
      target: document.body,
      props: {
        lifecycleState: 'active',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-household',
        locations: [],
        recentAssets: [],
        archivedAssets: [],
        canCreateAsset: false,
        onOpenLocation: () => {},
        onOpenAsset: () => {},
        onOpenAdd: () => {
          openedAdd = true;
        },
        onSelectLifecycle: () => {}
      }
    });

    const headerAdd = disabledLink('Add location');
    const emptyAdd = disabledLink('Add first location');
    expect(document.body.textContent).not.toContain('Add item');
    expect(headerAdd.hasAttribute('href')).toBe(false);
    expect(headerAdd.getAttribute('aria-disabled')).toBe('true');
    expect(headerAdd.getAttribute('aria-describedby')).toBe('home-add-location-denied');
    expect(emptyAdd.hasAttribute('href')).toBe(false);
    expect(emptyAdd.getAttribute('aria-disabled')).toBe('true');
    expect(emptyAdd.getAttribute('aria-describedby')).toBe('home-add-location-denied');
    expect(document.body.textContent).toContain('Creating locations is unavailable for this inventory.');

    headerAdd.click();
    emptyAdd.click();
    expect(openedAdd).toBe(false);
  });

  it('preserves modified clicks on linked home rows', () => {
    let openedAssetId = '';
    const asset: AssetViewModel = {
      id: 'tape',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'item',
      title: 'Tape measure',
      description: '',
      parentAssetId: 'garage',
      lifecycleState: 'active',
      containmentTrail: 'Garage'
    };

    component = mount(HomeWorkspace, {
      target: document.body,
      props: {
        lifecycleState: 'active',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-household',
        locations: [],
        recentAssets: [asset],
        archivedAssets: [],
        onOpenLocation: () => {},
        onOpenAsset: (selected) => {
          openedAssetId = selected.id;
        },
        onOpenAdd: () => {},
        onSelectLifecycle: () => {}
      }
    });

    let componentPreventedModifiedClick = false;
    const target = link('Tape measure');
    target.addEventListener('click', (event) => {
      componentPreventedModifiedClick = event.defaultPrevented;
      event.preventDefault();
    });
    target.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true, metaKey: true }));

    expect(openedAssetId).toBe('');
    expect(componentPreventedModifiedClick).toBe(false);
  });

  it('searches by a recent asset tag without opening the asset row', () => {
    let openedAssetId = '';
    const searchedTags: string[] = [];
    const asset: AssetViewModel = {
      id: 'tent',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'item',
      title: 'Family tent',
      description: '',
      parentAssetId: 'garage',
      lifecycleState: 'active',
      containmentTrail: 'Garage',
      tags: [{ id: 'tag-camping', key: 'camping', displayName: 'Camping', color: '#2F80ED' }]
    };

    component = mount(HomeWorkspace, {
      target: document.body,
      props: {
        lifecycleState: 'active',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-household',
        locations: [],
        recentAssets: [asset],
        archivedAssets: [],
        onOpenLocation: () => {},
        onOpenAsset: (selected) => {
          openedAssetId = selected.id;
        },
        onOpenAdd: () => {},
        onSelectLifecycle: () => {},
        onTagSearch: (tag) => {
          searchedTags.push(tag.displayName);
        }
      }
    });

    controlWithLabel('Search for tag Camping').click();

    expect(searchedTags).toEqual(['Camping']);
    expect(openedAssetId).toBe('');
  });
});

function link(text: string): HTMLAnchorElement {
  const target = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!target) {
    throw new Error(`Missing link ${text}`);
  }
  return target;
}

function disabledLink(text: string): HTMLAnchorElement {
  const target = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!target) {
    throw new Error(`Missing disabled link ${text}`);
  }
  return target;
}

function controlWithLabel(label: string): HTMLElement {
  const target = document.body.querySelector<HTMLElement>(`[aria-label="${label}"]`);
  if (!target) {
    throw new Error(`Missing control ${label}`);
  }
  return target;
}
