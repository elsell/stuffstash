import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import HomeWorkspace from './HomeWorkspace.svelte';
import type { AssetViewModel } from '$lib/domain/inventory';

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
    const location: AssetViewModel = {
      id: 'garage',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'location',
      title: 'Garage',
      description: '',
      parentAssetId: null,
      lifecycleState: 'active',
      containmentTrail: 'Garage'
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
            parentAssetId: 'garage'
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
    const location: AssetViewModel = {
      id: 'garage',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'location',
      title: 'Garage',
      description: '',
      parentAssetId: null,
      lifecycleState: 'active',
      containmentTrail: 'Garage'
    };
    const archived: AssetViewModel = {
      ...location,
      id: 'old-drill',
      kind: 'item',
      title: 'Old drill',
      lifecycleState: 'archived'
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

  it('uses selected context for add-location links when the inventory is empty', () => {
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
        onSelectLifecycle: () => {}
      }
    });

    expect(link('Add first location').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/add/location');
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
