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

function asset(id: string, title: string, photoUrl?: string): Asset {
  return {
    id,
    tenantId: tenant.id,
    inventoryId: inventory.id,
    kind: 'item',
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
      onSelectTenant: () => {},
      onSelectInventory: () => {},
      onSearch: () => {},
      onOpenAsset: (selected) => {
        selectedAssets.push(selected);
      },
      onOpenAdd: () => {},
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
  it('opens search suggestions from the keyboard', async () => {
    const { selectedAssets } = mountHeader({
      suggestions: [asset('tape', 'Tape measure', 'blob:tape-photo'), asset('tags', 'Gift tags')]
    });
    const input = document.body.querySelector<HTMLInputElement>('input[aria-label="Search this inventory"]');

    input?.focus();
    await flush();
    expect(document.body.querySelector<HTMLImageElement>('#global-search-suggestions img')?.src).toBe('blob:tape-photo');
    expect(document.body.querySelector<HTMLImageElement>('#global-search-suggestions img')?.alt).toBe('Tape measure');
    input?.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true }));
    await flush();

    expect(input?.getAttribute('role')).toBeNull();
    expect(input?.getAttribute('aria-activedescendant')).toBeNull();
    expect(document.body.querySelector('[role="listbox"]')).toBeNull();
    expect(document.activeElement?.id).toBe('global-search-suggestion-0');

    document.activeElement?.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true }));
    await flush();
    expect(document.activeElement?.id).toBe('global-search-suggestion-1');

    document.activeElement?.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowUp', bubbles: true }));
    await flush();
    expect(document.activeElement?.id).toBe('global-search-suggestion-0');

    (document.activeElement as HTMLButtonElement | null)?.click();
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

  it('keeps suggestions open when keyboard focus moves into the suggestion list', async () => {
    vi.useFakeTimers();
    mountHeader();
    const input = document.body.querySelector<HTMLInputElement>('input[aria-label="Search this inventory"]');

    input?.focus();
    await flush();
    input?.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true }));
    await flush();
    vi.advanceTimersByTime(160);
    await flush();

    expect(document.activeElement?.id).toBe('global-search-suggestion-0');
    expect(document.body.querySelector('#global-search-suggestions')).not.toBeNull();
  });

  it('closes suggestions with Escape from a focused suggestion and returns to the search field', async () => {
    mountHeader();
    const input = document.body.querySelector<HTMLInputElement>('input[aria-label="Search this inventory"]');

    input?.focus();
    await flush();
    input?.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true }));
    await flush();
    document.activeElement?.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
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
    expect(input?.getAttribute('aria-expanded')).toBeNull();
  });

  it('exposes durable add action links from the create menu', async () => {
    const addedKinds: AssetKind[] = [];
    mountHeader({
      onOpenAdd: (kind) => {
        addedKinds.push(kind);
      }
    });

    buttonContaining('Add').click();
    await flush();

    expect(document.activeElement?.textContent).toContain('Item');
    expect(document.body.querySelector('#header-add-menu')?.getAttribute('role')).toBeNull();
    expect(buttonContaining('Add').getAttribute('aria-haspopup')).toBeNull();
    expect(linkContaining('Item').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/add/item');
    expect(linkContaining('Container').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/add/container');
    expect(linkContaining('Location').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/add/location');

    linkContaining('Location').click();
    await flush();

    expect(addedKinds).toEqual(['location']);
    expect(document.body.querySelector('#header-add-menu')).toBeNull();
  });

  it('closes the add menu with Escape and restores focus to the trigger', async () => {
    mountHeader();

    const trigger = buttonContaining('Add');
    trigger.focus();
    trigger.click();
    await flush();

    expect(document.body.querySelector('#header-add-menu')).not.toBeNull();
    linkContaining('Item').dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    await flush();

    expect(document.body.querySelector('#header-add-menu')).toBeNull();
    expect(document.activeElement).toBe(trigger);
  });
});

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
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
