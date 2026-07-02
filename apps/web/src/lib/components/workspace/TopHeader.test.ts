import { tick } from 'svelte';
import { mount, unmount } from 'svelte';
import { afterEach, describe, expect, it } from 'vitest';
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
});

describe('TopHeader', () => {
  it('opens search suggestions from the keyboard', async () => {
    const { selectedAssets } = mountHeader();
    const input = document.body.querySelector<HTMLInputElement>('input[aria-label="Search this inventory"]');

    input?.dispatchEvent(new FocusEvent('focus', { bubbles: true }));
    await tick();
    input?.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true }));
    await tick();

    expect(input?.getAttribute('aria-activedescendant')).toBe('global-search-suggestion-0');

    input?.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));
    await tick();

    expect(selectedAssets.map((selected) => selected.id)).toEqual(['tape']);
    expect(input?.value).toBe('Tape measure');
  });

  it('closes search suggestions with Escape', async () => {
    mountHeader();
    const input = document.body.querySelector<HTMLInputElement>('input[aria-label="Search this inventory"]');

    input?.dispatchEvent(new FocusEvent('focus', { bubbles: true }));
    await tick();

    expect(document.body.querySelector('[role="listbox"]')).not.toBeNull();

    input?.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    await tick();

    expect(document.body.querySelector('[role="listbox"]')).toBeNull();
    expect(input?.getAttribute('aria-expanded')).toBe('false');
  });
});
