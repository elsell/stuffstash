import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import LocationView from './LocationView.svelte';
import type { Asset, AssetViewModel } from '$lib/domain/inventory';

let component: ReturnType<typeof mount> | null = null;

const location: Asset = {
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

    click('Shelf');
    click('Tape measure');

    expect(openedLocationId).toBe('garage-shelf');
    expect(openedAssetId).toBe('tape');
  });

  it('opens the current location for editing', () => {
    let editedLocationId = '';
    component = mount(LocationView, {
      target: document.body,
      props: {
        location,
        assets: [],
        canEdit: true,
        onBack: () => {},
        onOpenLocation: () => {},
        onEditLocation: (asset) => {
          editedLocationId = asset.id;
        },
        onOpenAsset: () => {}
      }
    });

    click('Edit location');

    expect(editedLocationId).toBe('garage');
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
});

function click(text: string): void {
  const button = Array.from(document.body.querySelectorAll('button')).find((candidate) => candidate.textContent?.includes(text));
  if (!button) throw new Error(`Missing button ${text}`);
  button.click();
}
