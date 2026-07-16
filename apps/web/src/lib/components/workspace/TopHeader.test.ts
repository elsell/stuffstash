import { tick } from 'svelte';
import { mount, unmount } from 'svelte';
import type { ComponentProps } from 'svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { Asset, AssetKind, Inventory, Tenant } from '$lib/domain/inventory';
import TopHeader from './TopHeader.svelte';

let component: ReturnType<typeof mount> | null = null;

const tenant: Tenant = {
  id: 'tenant-home',
  name: 'Home',
  access: { relationship: 'owner', permissions: [] }
};

const inventory: Inventory = {
  id: 'inventory-household',
  tenantId: tenant.id,
  name: 'Household',
  access: { relationship: 'owner', permissions: [] }
};

function asset(id: string, title: string, photoUrl?: string, kind: Asset['kind'] = 'item'): Asset {
  return {
    id,
    tenantId: tenant.id,
    inventoryId: inventory.id,
    kind,
    title,
    description: '',
    parentAssetId: null,
    lifecycleState: 'active',
    ...(photoUrl ? { photo: { id: `${id}-photo`, assetId: id, url: photoUrl, alt: title } } : {})
  };
}

function mountHeader(props: Partial<ComponentProps<typeof TopHeader>> = {}) {
  const selectedAssets: Asset[] = [];

  component = mount(TopHeader, {
    target: document.body,
    props: {
      tenants: [tenant],
      inventories: [inventory],
      selectedTenantId: tenant.id,
      inventory,
      suggestions: [asset('tape', 'Tape measure'), asset('tags', 'Gift tags')],
      query: 'ta',
      canCreateAsset: true,
      disablePortal: true,
      onSelectTenant: () => {},
      onSelectInventory: () => {},
      onSearch: () => {},
      onOpenAsset: (selected) => {
        selectedAssets.push(selected);
      },
      onOpenAdd: () => {},
      userLabel: 'owner@example.com',
      onOpenSettings: () => {},
      onSignOut: () => {},
      ...props
    }
  });

  return { selectedAssets };
}

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
  vi.useRealTimers();
});

describe('TopHeader', () => {
  it('provides the mobile shell with an account control', () => {
    mountHeader();

    const account = document.body.querySelector<HTMLButtonElement>('[aria-label="Open account menu"]');
    expect(account).not.toBeNull();
    expect(getComputedStyle(account!).minHeight).toBe('44px');
  });

  it('uses a compact inventory toolbar when the page owns search', () => {
    mountHeader({ showSearch: false });

    const header = document.querySelector<HTMLElement>('.workspace-header');
    expect(header?.classList.contains('contextual-toolbar')).toBe(true);
    expect(header?.querySelector('.desktop-header-context')?.textContent).toContain('Household');
    expect(header?.querySelector('.global-search-wrap')).toBeNull();
    expect(header?.textContent).toContain('Add');
  });

  it('opens search suggestions from the keyboard', async () => {
    const { selectedAssets } = mountHeader({
      suggestions: [asset('tape', 'Tape measure', 'blob:tape-photo'), asset('tags', 'Gift tags')]
    });
    const input = document.body.querySelector<HTMLInputElement>('input[aria-label="Search this inventory"]');

    input?.focus();
    await flush();
    expect(document.body.querySelector<HTMLImageElement>('#global-search-suggestions img')?.src).toBe('blob:tape-photo');
    expect(document.body.querySelector<HTMLImageElement>('#global-search-suggestions img')?.alt).toBe('Tape measure');
    expect(controlWithLabel('Open Tape measure').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/assets/tape'
    );
    input?.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true }));
    await flush();

    expect(input?.getAttribute('role')).toBe('combobox');
    expect(input?.getAttribute('aria-activedescendant')).toBe('global-search-suggestion-0');
    expect(document.body.querySelector('[role="listbox"]')).not.toBeNull();
    expect(document.activeElement).toBe(input);

    input?.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true }));
    await flush();
    expect(input?.getAttribute('aria-activedescendant')).toBe('global-search-suggestion-1');
    expect(document.activeElement).toBe(input);

    input?.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowUp', bubbles: true }));
    await flush();
    expect(input?.getAttribute('aria-activedescendant')).toBe('global-search-suggestion-0');

    input?.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }));
    await flush();

    expect(selectedAssets.map((selected) => selected.id)).toEqual(['tape']);
    expect(input?.value).toBe('Tape measure');
  });

  it('uses kind fallbacks for global suggestions without their own photo', async () => {
    mountHeader({
      suggestions: [
        {
          ...asset('box', 'Holiday box'),
          kind: 'container',
          photo: { id: 'wrong-photo', assetId: 'different-asset', url: 'blob:wrong-photo', alt: 'Wrong photo' }
        },
        asset('tags', 'Gift tags')
      ]
    });
    const input = document.body.querySelector<HTMLInputElement>('input[aria-label="Search this inventory"]');

    input?.focus();
    await flush();

    expect(document.body.querySelector('#global-search-suggestions img')).toBeNull();
    expect(document.body.querySelectorAll('#global-search-suggestions .asset-thumb svg')).toHaveLength(2);
  });

  it('shows calm no-suggestion feedback for focused global search queries', async () => {
    mountHeader({ query: 'box', suggestions: [] });
    const input = document.body.querySelector<HTMLInputElement>('input[aria-label="Search this inventory"]');

    input?.focus();
    await flush();

    const noSuggestions = document.body.querySelector<HTMLElement>('.search-suggestions-empty');
    expect(noSuggestions?.getAttribute('role')).toBe('status');
    expect(noSuggestions?.textContent).toBe('No suggestions for "box". Press Search to run a full search.');
    expect(document.body.querySelector('#global-search-suggestions')).toBeNull();
  });

  it('closes global no-suggestion feedback when submitting search', async () => {
    const searches: string[] = [];
    mountHeader({
      query: 'box',
      suggestions: [],
      onSearch: () => {
        searches.push('search');
      }
    });
    const input = document.body.querySelector<HTMLInputElement>('input[aria-label="Search this inventory"]');

    input?.focus();
    await flush();
    expect(document.body.querySelector('.search-suggestions-empty')).not.toBeNull();

    document.body.querySelector('form.global-search')?.dispatchEvent(new SubmitEvent('submit', { bubbles: true, cancelable: true }));
    await flush();

    expect(searches).toEqual(['search']);
    expect(document.body.querySelector('.search-suggestions-empty')).toBeNull();
  });

  it('marks global suggestions when a primary photo cannot render', async () => {
    mountHeader({
      suggestions: [{ ...asset('tape', 'Tape measure'), photoUnavailable: true }]
    });
    const input = document.body.querySelector<HTMLInputElement>('input[aria-label="Search this inventory"]');

    input?.focus();
    await flush();

    expect(document.body.querySelector('#global-search-suggestions img')).toBeNull();
    const suggestion = controlWithLabel('Open Tape measure');
    expect(suggestion.getAttribute('aria-describedby')).toBe('global-search-suggestion-0-photo-unavailable');
    expect(document.getElementById('global-search-suggestion-0-photo-unavailable')?.textContent).toBe('Photo unavailable');
    expect(document.body.querySelector('.photo-unavailable-mark')).not.toBeNull();
  });

  it('routes global location suggestions to the focused location surface', async () => {
    const { selectedAssets } = mountHeader({
      suggestions: [asset('garage', 'Garage', undefined, 'location')]
    });
    const input = document.body.querySelector<HTMLInputElement>('input[aria-label="Search this inventory"]');

    input?.focus();
    await flush();

    expect(controlWithLabel('Open Garage').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/locations/garage'
    );

    controlWithLabel('Open Garage').click();
    await flush();

    expect(selectedAssets.map((selected) => selected.id)).toEqual(['garage']);
  });

  it('preserves modified clicks on global suggestion links', async () => {
    const { selectedAssets } = mountHeader();
    const input = document.body.querySelector<HTMLInputElement>('input[aria-label="Search this inventory"]');

    input?.focus();
    await flush();

    let componentPreventedModifiedClick = false;
    const target = controlWithLabel('Open Tape measure');
    target.addEventListener('click', (event) => {
      componentPreventedModifiedClick = event.defaultPrevented;
      event.preventDefault();
    });
    target.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true, metaKey: true }));

    expect(selectedAssets).toEqual([]);
    expect(componentPreventedModifiedClick).toBe(false);
  });

  it('keeps keyboard focus on the combobox while traversing suggestions', async () => {
    vi.useFakeTimers();
    mountHeader();
    const input = document.body.querySelector<HTMLInputElement>('input[aria-label="Search this inventory"]');

    input?.focus();
    await flush();
    input?.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true }));
    await flush();
    vi.advanceTimersByTime(160);
    await flush();

    expect(document.activeElement).toBe(input);
    expect(input?.getAttribute('aria-activedescendant')).toBe('global-search-suggestion-0');
    expect(document.body.querySelector('#global-search-suggestions')).not.toBeNull();
  });

  it('closes suggestions with Escape while focus remains on the search field', async () => {
    mountHeader();
    const input = document.body.querySelector<HTMLInputElement>('input[aria-label="Search this inventory"]');

    input?.focus();
    await flush();
    input?.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true }));
    await flush();
    input?.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    await flush();

    expect(document.activeElement).toBe(input);
    expect(document.body.querySelector('#global-search-suggestions')).toBeNull();
  });

  it('closes search suggestions with Escape', async () => {
    mountHeader();
    const input = document.body.querySelector<HTMLInputElement>('input[aria-label="Search this inventory"]');

    input?.focus();
    await flush();

    expect(document.body.querySelector('#global-search-suggestions')).not.toBeNull();

    input?.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    await flush();

    expect(document.body.querySelector('#global-search-suggestions')).toBeNull();
    expect(input?.getAttribute('aria-expanded')).toBe('false');
  });

  it('exposes durable Add links through a real keyboard menu', async () => {
    const addedKinds: AssetKind[] = [];
    mountHeader({
      onOpenAdd: (kind) => {
        addedKinds.push(kind);
      }
    });

    const trigger = addTrigger();
    expect(trigger.getAttribute('aria-haspopup')).toBe('menu');
    expect(trigger.getAttribute('aria-expanded')).toBe('false');
    expect(trigger.classList.contains('min-h-11')).toBe(true);

    trigger.click();
    await waitForAddMenu();

    expect(trigger.getAttribute('aria-expanded')).toBe('true');
    expect(document.body.querySelector('[role="menu"]')).not.toBeNull();
    expect(menuItemContaining('Item').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/add/item');
    expect(menuItemContaining('Container').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/add/container');
    expect(menuItemContaining('Location').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/add/location');
    const menu = document.body.querySelector<HTMLElement>('[role="menu"]');
    menu?.focus();
    menu?.dispatchEvent(new KeyboardEvent('keydown', { key: 'End', bubbles: true, cancelable: true }));
    await flush();
    expect(document.activeElement).toBe(menuItemContaining('Location'));

    menuItemContaining('Location').click();
    await flush();

    expect(addedKinds).toEqual(['location']);
    expect(trigger.getAttribute('aria-expanded')).toBe('false');
  });

  it('exposes a perceivable disabled reason when header add is unavailable', () => {
    mountHeader({ canCreateAsset: false });

    const trigger = addTrigger();
    expect(trigger.disabled).toBe(true);
    expect(trigger.getAttribute('aria-describedby')).toBe('header-add-denied');
    expect(document.body.querySelector('#header-add-denied')?.textContent).toBe('Adding assets is unavailable for this inventory.');
  });

  it('preserves modified clicks on durable Add menu links', async () => {
    const addedKinds: AssetKind[] = [];
    mountHeader({ onOpenAdd: (kind) => addedKinds.push(kind) });

    addTrigger().click();
    await waitForAddMenu();

    const target = menuItemContaining('Container');
    let componentPreventedModifiedClick = false;
    target.addEventListener('click', (event) => {
      componentPreventedModifiedClick = event.defaultPrevented;
      event.preventDefault();
    });
    target.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true, metaKey: true }));

    expect(addedKinds).toEqual([]);
    expect(componentPreventedModifiedClick).toBe(false);
  });

  it('does not open the add menu without a selected inventory', () => {
    mountHeader({ inventory: null, canCreateAsset: true });

    const trigger = addTrigger();
    trigger.click();

    expect(trigger.disabled).toBe(true);
    expect(trigger.getAttribute('aria-describedby')).toBe('header-add-denied');
    expect(document.body.querySelector('#header-add-denied')?.textContent).toBe('Select an inventory before adding assets.');
    expect(trigger.getAttribute('aria-expanded')).toBe('false');
  });

  it('closes the Add menu with Escape and restores focus to its trigger', async () => {
    mountHeader();

    const trigger = addTrigger();
    trigger.focus();
    trigger.click();
    await waitForAddMenu();

    const menu = document.body.querySelector<HTMLElement>('[role="menu"]');
    expect(menu).not.toBeNull();
    menu?.focus();
    menu?.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
    await flush();

    expect(trigger.getAttribute('aria-expanded')).toBe('false');
    expect(document.activeElement).toBe(trigger);
  });
});

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
}

function addTrigger(): HTMLButtonElement {
  const trigger = document.body.querySelector<HTMLButtonElement>('[data-workspace-add-trigger="desktop"]');
  if (!trigger) throw new Error('Missing desktop Add trigger');
  return trigger;
}

function menuItemContaining(text: string): HTMLAnchorElement {
  const item = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('[role="menuitem"]')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!item) throw new Error(`Missing menu item containing ${text}`);
  return item;
}

async function waitForAddMenu(): Promise<void> {
  const deadline = Date.now() + 1_000;
  while (!document.body.querySelector('[role="menu"]')) {
    if (Date.now() >= deadline) throw new Error('Timed out waiting for Add menu');
    await new Promise<void>((resolve) => window.setTimeout(resolve, 10));
    await flush();
  }
}

function buttonContaining(text: string): HTMLButtonElement {
  const button = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!button) {
    throw new Error(`Missing button containing ${text}`);
  }
  return button;
}

function linkContaining(text: string): HTMLAnchorElement {
  const link = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!link) {
    throw new Error(`Missing link containing ${text}`);
  }
  return link;
}

function controlWithLabel(label: string): HTMLElement {
  const control = document.body.querySelector<HTMLElement>(`button[aria-label="${label}"], a[aria-label="${label}"]`);
  if (!control) {
    throw new Error(`Missing control labelled ${label}`);
  }
  return control;
}
