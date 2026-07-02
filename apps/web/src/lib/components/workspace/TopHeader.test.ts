import { tick } from 'svelte';
import { mount, unmount } from 'svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { Asset, Inventory, Tenant } from '$lib/domain/inventory';
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

function asset(id: string, title: string): Asset {
  return {
    id,
    tenantId: tenant.id,
    inventoryId: inventory.id,
    kind: 'item',
    title,
    description: '',
    parentAssetId: null,
    lifecycleState: 'active'
  };
}

function mountHeader(props: Partial<Parameters<typeof TopHeader>[0]> = {}) {
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
      onOpenSettings: () => {},
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
    const { selectedAssets } = mountHeader();
    const input = document.body.querySelector<HTMLInputElement>('input[aria-label="Search this inventory"]');

    input?.focus();
    await flush();
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
});

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
}
