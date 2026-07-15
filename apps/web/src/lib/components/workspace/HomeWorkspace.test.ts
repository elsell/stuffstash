import { afterEach, describe, expect, it, vi } from 'vitest';
import { mount, tick, unmount } from 'svelte';
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
  it('provides accessible controls for traversing the recently changed rail', async () => {
    const scrollBy = vi.fn();
    component = mount(HomeWorkspace, {
      target: document.body,
      props: {
        lifecycleState: 'active',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-household',
        locations: [],
        recentAssets: [
          recentAsset('tape', 'Tape measure'),
          recentAsset('drill', 'Cordless drill')
        ],
        archivedAssets: [],
        onOpenLocation: () => {},
        onOpenAsset: () => {},
        onOpenAdd: () => {},
        onSelectLifecycle: () => {}
      }
    });

    expect(document.body.querySelector('.section-heading.home-heading')).not.toBeNull();
    await tick();
    const rail = document.body.querySelector<HTMLElement>('[aria-label="Recently changed assets"]');
    if (!rail) throw new Error('Missing recent rail');
    Object.defineProperties(rail, {
      clientWidth: { configurable: true, value: 600 },
      scrollWidth: { configurable: true, value: 1200 },
      scrollBy: { configurable: true, value: scrollBy }
    });
    window.dispatchEvent(new Event('resize'));
    await tick();

    controlWithLabel('Next recently changed assets').click();
    expect(scrollBy).toHaveBeenCalledWith({ left: 510, behavior: 'smooth' });
    expect(controlWithLabel('Previous recently changed assets')).not.toBeNull();
  });

  it('does not show dead rail controls when recent cards fit', async () => {
    component = mount(HomeWorkspace, {
      target: document.body,
      props: {
        lifecycleState: 'active', tenantId: 'tenant-home', inventoryId: 'inventory-household', locations: [],
        recentAssets: [recentAsset('tape', 'Tape measure'), recentAsset('drill', 'Cordless drill')], archivedAssets: [],
        onOpenLocation: () => {}, onOpenAsset: () => {}, onOpenAdd: () => {}, onSelectLifecycle: () => {}
      }
    });
    await tick();
    const rail = document.body.querySelector<HTMLElement>('[aria-label="Recently changed assets"]')!;
    Object.defineProperties(rail, {
      clientWidth: { configurable: true, value: 800 },
      scrollWidth: { configurable: true, value: 500 }
    });
    window.dispatchEvent(new Event('resize'));
    await tick();

    expect(document.body.querySelector('[aria-label="Recently changed navigation"]')).toBeNull();
  });

  it('avoids smooth rail motion when reduced motion is requested', async () => {
    const originalMatchMedia = window.matchMedia;
    const scrollBy = vi.fn();
    Object.defineProperty(window, 'matchMedia', {
      configurable: true,
      value: vi.fn(() => ({ matches: true, addEventListener: vi.fn(), removeEventListener: vi.fn() }))
    });
    try {
      component = mount(HomeWorkspace, {
        target: document.body,
        props: {
          lifecycleState: 'active', tenantId: 'tenant-home', inventoryId: 'inventory-household', locations: [],
          recentAssets: [recentAsset('tape', 'Tape measure'), recentAsset('drill', 'Cordless drill')], archivedAssets: [],
          onOpenLocation: () => {}, onOpenAsset: () => {}, onOpenAdd: () => {}, onSelectLifecycle: () => {}
        }
      });
      await tick();
      const rail = document.body.querySelector<HTMLElement>('[aria-label="Recently changed assets"]')!;
      Object.defineProperties(rail, {
        clientWidth: { configurable: true, value: 600 },
        scrollWidth: { configurable: true, value: 1200 },
        scrollBy: { configurable: true, value: scrollBy }
      });
      window.dispatchEvent(new Event('resize'));
      await tick();

      controlWithLabel('Next recently changed assets').click();
      expect(scrollBy).toHaveBeenCalledWith({ left: 510, behavior: 'auto' });
    } finally {
      Object.defineProperty(window, 'matchMedia', { configurable: true, value: originalMatchMedia });
    }
  });

  it('opens a recently changed place through its focused location workspace', () => {
    let openedLocation = '';
    let openedAsset = '';
    const place: AssetViewModel = {
      ...recentAsset('garage', 'Garage'),
      kind: 'location',
      containmentTrail: 'Inventory root'
    };
    component = mount(HomeWorkspace, {
      target: document.body,
      props: {
        lifecycleState: 'active',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-household',
        locations: [],
        recentAssets: [place],
        archivedAssets: [],
        onOpenLocation: (asset) => { openedLocation = asset.id; },
        onOpenAsset: (asset) => { openedAsset = asset.id; },
        onOpenAdd: () => {},
        onSelectLifecycle: () => {}
      }
    });

    expect(link('Garage').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/locations/garage');
    link('Garage').click();
    expect(openedLocation).toBe('garage');
    expect(openedAsset).toBe('');
  });

  it('labels the recency rail as recently changed', () => {
    component = mount(HomeWorkspace, {
      target: document.body,
      props: {
        lifecycleState: 'active',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-household',
        locations: [],
        recentAssets: [{
          id: 'tape',
          tenantId: 'tenant-home',
          inventoryId: 'inventory-household',
          kind: 'item',
          title: 'Tape measure',
          description: '',
          parentAssetId: null,
          lifecycleState: 'active',
          containmentTrail: 'Inventory root'
        }],
        archivedAssets: [],
        onOpenLocation: () => {},
        onOpenAsset: () => {},
        onOpenAdd: () => {},
        onSelectLifecycle: () => {}
      }
    });

    expect(document.body.textContent).toContain('Recently changed');
    expect(document.body.querySelector('[aria-label="Recently changed assets"]')).not.toBeNull();
    expect(document.body.querySelector('[data-recent-card="tape"] [data-recent-card-media]')).not.toBeNull();
    expect(document.body.querySelector('[data-recent-card="tape"] [data-recent-card-title]')?.textContent).toBe('Tape measure');
    expect(document.body.querySelector('[data-recent-card="tape"] [data-recent-card-tags]')).not.toBeNull();
    expect(document.body.textContent).not.toContain('Recently added');
  });

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
    expect(document.body.querySelector('.locations-heading h2')?.textContent).toBe('Places');
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
    const archivedRow = link('Old drill').closest('.asset-row');
    expect(archivedRow?.querySelector('.asset-row-actions')).not.toBeNull();
    expect(archivedRow?.querySelector('.asset-row-actions')?.textContent).toContain('Archived');
  });

  it('keeps long archived identity and secondary copy inside the flexible row link', () => {
    const archived = {
      ...recentAsset('long-archived', 'Walgreens Daytime Severe Cold & Flu Liquid Maximum Strength - 12.0 fl oz'),
      lifecycleState: 'archived' as const,
      description: 'Daytime cold and flu medicine stored on the upper shelf for the winter season.'
    };
    component = mount(HomeWorkspace, {
      target: document.body,
      props: {
        lifecycleState: 'archived', tenantId: 'tenant-home', inventoryId: 'inventory-household', locations: [],
        recentAssets: [], archivedAssets: [archived], onOpenLocation: () => {}, onOpenAsset: () => {},
        onOpenAdd: () => {}, onSelectLifecycle: () => {}
      }
    });

    const rowLink = link('Walgreens Daytime Severe Cold & Flu');
    expect(rowLink.classList.contains('asset-row-open')).toBe(true);
    expect(rowLink.textContent).toContain('Daytime cold and flu medicine stored on the upper shelf for the winter season.');
    expect(rowLink.closest('.asset-row')?.querySelector('.asset-row-actions')?.textContent).toContain('Archived');
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

  it('shows checked-out photos and provides editors a trailing one-click Return action', async () => {
    const returned: string[] = [];
    const checkedOut: AssetViewModel = {
      id: 'ratchet',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'item',
      title: 'Ratchet straps',
      description: '',
      parentAssetId: 'garage',
      lifecycleState: 'active',
      containmentTrail: 'Garage',
      photo: { id: 'photo-one', assetId: 'ratchet', url: 'blob:ratchet', alt: 'Ratchet straps' },
      currentCheckout: {
        id: 'checkout-one',
        state: 'open',
        checkedOutAt: '2026-07-14T12:00:00Z',
        checkedOutByPrincipalId: 'principal-one'
      }
    };

    component = mount(HomeWorkspace, {
      target: document.body,
      props: {
        lifecycleState: 'active',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-household',
        locations: [],
        recentAssets: [],
        archivedAssets: [],
        checkedOutAssets: [checkedOut],
        canEditAsset: true,
        onOpenLocation: () => {},
        onOpenAsset: () => {},
        onReturnAsset: async (asset) => { returned.push(asset.id); },
        onOpenAdd: () => {},
        onSelectLifecycle: () => {}
      }
    });

    expect(link('Ratchet straps').textContent).toContain('Checked out');
    expect(document.body.querySelector<HTMLImageElement>('img[src="blob:ratchet"]')?.alt).toBe('Ratchet straps');
    expect(document.body.textContent).not.toContain('active');
    controlWithLabel('Return Ratchet straps').click();
    await tick();
    expect(returned).toEqual(['ratchet']);
  });

  it('does not expose the Home return action to viewers', () => {
    const checkedOut = { ...recentAsset('drill', 'Cordless drill'), currentCheckout: {
      id: 'checkout-drill', state: 'open' as const, checkedOutAt: '2026-07-14T12:00:00Z', checkedOutByPrincipalId: 'principal-one'
    } };
    component = mount(HomeWorkspace, { target: document.body, props: {
      lifecycleState: 'active', tenantId: 'tenant-home', inventoryId: 'inventory-household', locations: [],
      recentAssets: [], archivedAssets: [], checkedOutAssets: [checkedOut], canEditAsset: false,
      onOpenLocation: () => {}, onOpenAsset: () => {}, onOpenAdd: () => {}, onSelectLifecycle: () => {},
      onReturnAsset: async () => {}
    } });

    expect(document.body.querySelector('[aria-label="Return Cordless drill"]')).toBeNull();
  });

  it('announces a pending return and prevents duplicate activation', async () => {
    let finishReturn: (() => void) | undefined;
    const onReturnAsset = vi.fn(() => new Promise<void>((resolve) => { finishReturn = resolve; }));
    const checkedOut = { ...recentAsset('drill', 'Cordless drill'), currentCheckout: {
      id: 'checkout-drill', state: 'open' as const, checkedOutAt: '2026-07-14T12:00:00Z', checkedOutByPrincipalId: 'principal-one'
    } };
    component = mount(HomeWorkspace, { target: document.body, props: {
      lifecycleState: 'active', tenantId: 'tenant-home', inventoryId: 'inventory-household', locations: [],
      recentAssets: [], archivedAssets: [], checkedOutAssets: [checkedOut], canEditAsset: true,
      onOpenLocation: () => {}, onOpenAsset: () => {}, onOpenAdd: () => {}, onSelectLifecycle: () => {}, onReturnAsset
    } });

    controlWithLabel('Return Cordless drill').click();
    await tick();
    const pending = controlWithLabel('Returning Cordless drill');
    expect(pending.getAttribute('aria-busy')).toBe('true');
    pending.click();
    expect(onReturnAsset).toHaveBeenCalledTimes(1);
    finishReturn?.();
    await tick();
  });
});

function recentAsset(id: string, title: string): AssetViewModel {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    kind: 'item',
    title,
    description: '',
    parentAssetId: null,
    lifecycleState: 'active',
    containmentTrail: 'Inventory root'
  };
}

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
