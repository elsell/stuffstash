import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import LocationView from './LocationView.svelte';
import type { AssetViewModel, LocationAsset } from '$lib/domain/inventory';

let component: ReturnType<typeof mount> | null = null;

const location: LocationAsset = {
  id: 'garage',
  tenantId: 'tenant-home',
  inventoryId: 'inventory-household',
  kind: 'location',
  title: 'Garage',
  description: '',
  parentAssetId: null,
  lifecycleState: 'active'
};

const nestedLocation: AssetViewModel = {
  id: 'garage-shelf',
  tenantId: 'tenant-home',
  inventoryId: 'inventory-household',
  kind: 'location',
  title: 'Shelf',
  description: '',
  parentAssetId: 'garage',
  lifecycleState: 'active',
  containmentTrail: 'Garage'
};

const item: AssetViewModel = {
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

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('LocationView', () => {
  it('opens nested locations as location navigation and items as asset detail', () => {
    let openedLocationId = '';
    let openedAssetId = '';
    component = mount(LocationView, {
      target: document.body,
      props: {
        location,
        assets: [nestedLocation, item],
        canEdit: true,
        onBack: () => {},
        onOpenLocation: (asset) => {
          openedLocationId = asset.id;
        },
        onEditLocation: () => {},
        onOpenAsset: (asset) => {
          openedAssetId = asset.id;
        }
      }
    });

    clickLink('Shelf');
    clickLink('Tape measure');

    expect(openedLocationId).toBe('garage-shelf');
    expect(openedAssetId).toBe('tape');
  });

  it('exposes canonical hrefs for back, edit, nested location, and asset rows', () => {
    component = mount(LocationView, {
      target: document.body,
      props: {
        location,
        assets: [nestedLocation, item],
        canEdit: true,
        onBack: () => {},
        onOpenLocation: () => {},
        onEditLocation: () => {},
        onOpenAsset: () => {}
      }
    });

    expect(link('Back').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/browse?scope=places');
    expect(link('Edit location').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/locations/garage/edit'
    );
    expect(link('Shelf').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/locations/garage-shelf'
    );
    expect(link('Tape measure').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/assets/tape'
    );
  });

  it('does not intercept modified location-view link clicks', () => {
    let openedAssetId = '';
    component = mount(LocationView, {
      target: document.body,
      props: {
        location,
        assets: [item],
        canEdit: true,
        onBack: () => {},
        onOpenLocation: () => {},
        onEditLocation: () => {},
        onOpenAsset: (asset) => {
          openedAssetId = asset.id;
        }
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

  it('opens the current location for editing', () => {
    let editedLocationId = '';
    component = mount(LocationView, {
      target: document.body,
      props: {
        location,
        assets: [],
        canEdit: true,
        canCreateAsset: true,
        onBack: () => {},
        onOpenLocation: () => {},
        onEditLocation: (asset) => {
          editedLocationId = asset.id;
        },
        onOpenAsset: () => {}
      }
    });

    clickLink('Edit location');

    expect(editedLocationId).toBe('garage');
  });

  it('exposes a route-backed add-item action preselected to the current location', () => {
    let addKind = '';
    let addParentId = '';
    component = mount(LocationView, {
      target: document.body,
      props: {
        location,
        assets: [],
        canEdit: true,
        canCreateAsset: true,
        onBack: () => {},
        onOpenLocation: () => {},
        onEditLocation: () => {},
        onOpenAsset: () => {},
        onOpenAdd: (kind, parentId) => {
          addKind = kind;
          addParentId = parentId;
        }
      }
    });

    expect(link('Add item here').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/add/item?parent=garage'
    );
    clickLink('Add item here');

    expect(addKind).toBe('item');
    expect(addParentId).toBe('garage');
  });

  it('uses the helper-backed add label for non-empty location actions', () => {
    component = mount(LocationView, {
      target: document.body,
      props: {
        location,
        assets: [item],
        canEdit: true,
        canCreateAsset: true,
        onBack: () => {},
        onOpenLocation: () => {},
        onEditLocation: () => {},
        onOpenAsset: () => {}
      }
    });

    expect(link('Add item here').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/add/item?parent=garage'
    );
  });

  it('hides location editing when edit access is missing', () => {
    component = mount(LocationView, {
      target: document.body,
      props: {
        location,
        assets: [],
        canEdit: false,
        onBack: () => {},
        onOpenLocation: () => {},
        onEditLocation: () => {
          throw new Error('Edit should not be available.');
        },
        onOpenAsset: () => {}
      }
    });

    expect(document.body.textContent).not.toContain('Edit location');
  });

  it('uses a denied empty-state note when location creation access is missing', () => {
    component = mount(LocationView, {
      target: document.body,
      props: {
        location,
        assets: [],
        canEdit: false,
        canCreateAsset: false,
        onBack: () => {},
        onOpenLocation: () => {},
        onEditLocation: () => {},
        onOpenAsset: () => {},
        onOpenAdd: () => {
          throw new Error('Add should not be available.');
        }
      }
    });

    expect(document.body.textContent).toContain('This location is empty.');
    expect(document.body.textContent).toContain('Adding items is unavailable for this inventory.');
    expect(document.body.textContent).not.toContain('Add item here');
  });

  it('searches by contained asset tag without opening the contained asset row', () => {
    let openedAssetId = '';
    const searchedTags: string[] = [];
    component = mount(LocationView, {
      target: document.body,
      props: {
        location,
        assets: [
          {
            ...item,
            tags: [{ id: 'tag-tools', key: 'tools', displayName: 'Tools', color: '#2F80ED' }]
          }
        ],
        canEdit: true,
        onBack: () => {},
        onOpenLocation: () => {},
        onEditLocation: () => {},
        onOpenAsset: (asset) => {
          openedAssetId = asset.id;
        },
        onTagSearch: (tag) => {
          searchedTags.push(tag.displayName);
        }
      }
    });

    controlWithLabel('Search for tag Tools').click();

    expect(searchedTags).toEqual(['Tools']);
    expect(openedAssetId).toBe('');
  });
});

function clickLink(text: string): void {
  link(text).click();
}

function link(text: string): HTMLAnchorElement {
  const target = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!target) throw new Error(`Missing link ${text}`);
  return target;
}

function controlWithLabel(label: string): HTMLElement {
  const target = document.body.querySelector<HTMLElement>(`[aria-label="${label}"]`);
  if (!target) throw new Error(`Missing control ${label}`);
  return target;
}
