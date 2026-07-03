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
  it('routes Places to the durable locations destination', () => {
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

    linkContaining('Places').click();

    expect(selectedMode).toBe('locations');
    expect(linkContaining('Places').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/locations');
  });

  it('marks focused locations as the current Places section', () => {
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

    expect(linkContaining('Places').getAttribute('aria-current')).toBe('page');
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
    expect(linkContaining('Search').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/search');
    expect(document.body.querySelector<HTMLAnchorElement>('a[aria-label="Add asset"]')?.getAttribute('href')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/add/item'
    );
    expect(linkContaining('Settings').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/settings/activity');
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
