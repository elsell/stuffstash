import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import type { WorkspaceMode } from '$lib/domain/inventory';
import MobileNav from './MobileNav.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('MobileNav', () => {
  it('routes Browse to the durable unified destination', () => {
    let selectedMode: WorkspaceMode | null = null;
    component = mount(MobileNav, {
      target: document.body,
      props: {
        mode: 'home',
        selectedTenantId: 'tenant-one',
        selectedInventoryId: 'inventory-one',
        settingsSection: 'overview',
        canCreateAsset: true,
        onModeChange: (mode) => {
          selectedMode = mode;
        },
        onOpenAdd: () => {}
      }
    });

    linkContaining('Browse').click();

    expect(selectedMode).toBe('browse');
    expect(linkContaining('Browse').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/browse');
  });

  it('marks focused locations as the current Browse section', () => {
    component = mount(MobileNav, {
      target: document.body,
      props: {
        mode: 'location',
        selectedTenantId: 'tenant-one',
        selectedInventoryId: 'inventory-one',
        settingsSection: 'overview',
        canCreateAsset: true,
        onModeChange: () => {},
        onOpenAdd: () => {}
      }
    });

    expect(linkContaining('Browse').getAttribute('aria-current')).toBe('page');
  });

  it('exposes hrefs for durable mobile destinations and add action', () => {
    component = mount(MobileNav, {
      target: document.body,
      props: {
        mode: 'settings',
        selectedTenantId: 'tenant-one',
        selectedInventoryId: 'inventory-one',
        settingsSection: 'activity',
        canCreateAsset: true,
        onModeChange: () => {},
        onOpenAdd: () => {}
      }
    });

    expect(linkContaining('Home').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one');
    expect(linkContaining('Browse').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/browse');
    expect(document.body.querySelector<HTMLAnchorElement>('a[aria-label="Add asset"]')?.getAttribute('href')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/add/item'
    );
    expect(document.body.textContent).not.toContain('Settings');
  });

  it('does not open add when creation is unavailable', () => {
    let addOpened = false;
    component = mount(MobileNav, {
      target: document.body,
      props: {
        mode: 'home',
        selectedTenantId: 'tenant-one',
        selectedInventoryId: 'inventory-one',
        settingsSection: 'overview',
        canCreateAsset: false,
        onModeChange: () => {},
        onOpenAdd: () => {
          addOpened = true;
        }
      }
    });

    document.body.querySelector<HTMLAnchorElement>('a[aria-label="Add asset"]')?.click();

    expect(addOpened).toBe(false);
    expect(document.body.querySelector<HTMLAnchorElement>('a[aria-label="Add asset"]')?.getAttribute('href')).toBeNull();
    expect(document.body.querySelector<HTMLAnchorElement>('a[aria-label="Add asset"]')?.getAttribute('aria-disabled')).toBe('true');
    expect(document.body.querySelector<HTMLAnchorElement>('a[aria-label="Add asset"]')?.getAttribute('aria-describedby')).toBe('mobile-add-denied');
    expect(document.body.querySelector('#mobile-add-denied')?.textContent).toBe('Adding assets is unavailable for this inventory.');
  });

  it('explains disabled mobile add when no inventory is selected even with create permission', () => {
    let addOpened = false;
    component = mount(MobileNav, {
      target: document.body,
      props: {
        mode: 'home',
        selectedTenantId: 'tenant-one',
        selectedInventoryId: '',
        settingsSection: 'overview',
        canCreateAsset: true,
        onModeChange: () => {},
        onOpenAdd: () => {
          addOpened = true;
        }
      }
    });

    document.body.querySelector<HTMLAnchorElement>('a[aria-label="Add asset"]')?.click();

    expect(addOpened).toBe(false);
    expect(document.body.querySelector<HTMLAnchorElement>('a[aria-label="Add asset"]')?.getAttribute('href')).toBeNull();
    expect(document.body.querySelector<HTMLAnchorElement>('a[aria-label="Add asset"]')?.getAttribute('aria-disabled')).toBe('true');
    expect(document.body.querySelector<HTMLAnchorElement>('a[aria-label="Add asset"]')?.getAttribute('aria-describedby')).toBe('mobile-add-denied');
    expect(document.body.querySelector('#mobile-add-denied')?.textContent).toBe('Select an inventory before adding assets.');
  });
});

function linkContaining(text: string): HTMLAnchorElement {
  const link = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!link) {
    throw new Error(`Missing link containing ${text}`);
  }
  return link;
}
