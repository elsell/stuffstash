import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import SideNav from './SideNav.svelte';
import type { SettingsSection } from '$lib/application/workspaceRoute';
import type { Inventory, Tenant, WorkspaceMode } from '$lib/domain/inventory';

let component: ReturnType<typeof mount> | null = null;

type SideNavProps = {
  tenants: Tenant[];
  inventories: Inventory[];
  selectedTenantId: string;
  selectedInventoryId: string;
  mode: WorkspaceMode;
  settingsSection: SettingsSection;
  userLabel: string;
  onSelectTenant: (tenantId: string) => void;
  onSelectInventory: (tenantId: string, inventoryId: string) => void;
  onModeChange: (mode: WorkspaceMode) => void;
  onSignOut: () => void;
};

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('SideNav', () => {
  it('groups destinations and exposes the current destination', () => {
    component = mount(SideNav, {
      target: document.body,
      props: sideNavProps({ mode: 'settings' })
    });

    expect(document.body.querySelector('[aria-labelledby="primary-nav-label"]')?.textContent).toContain('Home');
    expect(document.body.querySelector('[aria-labelledby="primary-nav-label"]')?.textContent).toContain('Locations');
    expect(document.body.querySelector('[aria-labelledby="utility-nav-label"]')?.textContent).toContain('Import');
    expect(document.body.querySelector('[aria-labelledby="utility-nav-label"]')?.textContent).toContain('Settings');
    expect(linkContaining('Import').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/import');

    const currentDestinations = document.body.querySelectorAll<HTMLAnchorElement>('a[aria-current="page"]');
    expect(currentDestinations).toHaveLength(1);
    const current = currentDestinations[0];
    expect(current?.textContent).toContain('Settings');
    expect(current?.textContent).toContain('Access, fields, and audit');
    expect(current?.getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/settings');
  });

  it('marks home as the current primary destination', () => {
    component = mount(SideNav, {
      target: document.body,
      props: sideNavProps({ mode: 'home' })
    });

    const currentDestinations = document.body.querySelectorAll<HTMLAnchorElement>('a[aria-current="page"]');
    expect(currentDestinations).toHaveLength(1);
    expect(currentDestinations[0]?.textContent).toContain('Home');
    expect(currentDestinations[0]?.textContent).toContain('Recent assets and places');
    expect(currentDestinations[0]?.getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one');
  });

  it('marks focused location routes under the locations destination', () => {
    component = mount(SideNav, {
      target: document.body,
      props: sideNavProps({ mode: 'location' })
    });

    const currentDestinations = document.body.querySelectorAll<HTMLAnchorElement>('a[aria-current="page"]');
    expect(currentDestinations).toHaveLength(1);
    expect(currentDestinations[0]?.textContent).toContain('Locations');
    expect(currentDestinations[0]?.getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/locations');
  });

  it('routes destination clicks through the workspace mode callback', () => {
    let selectedMode: WorkspaceMode | null = null;
    component = mount(SideNav, {
      target: document.body,
      props: sideNavProps({
        mode: 'home',
        onModeChange: (mode) => {
          selectedMode = mode;
        }
      })
    });

    linkContaining('Locations').click();
    expect(selectedMode).toBe('locations');

    linkContaining('Import').click();
    expect(selectedMode).toBe('import');
  });

});

function sideNavProps(overrides: Partial<SideNavProps> = {}): SideNavProps {
  return {
    tenants: [{ id: 'tenant-one', name: 'Household', access: { relationship: 'owner', permissions: ['view'] } }],
    inventories: [
      {
        id: 'inventory-one',
        tenantId: 'tenant-one',
        name: 'Garage',
        access: { relationship: 'owner', permissions: ['view'] }
      }
    ],
    selectedTenantId: 'tenant-one',
    selectedInventoryId: 'inventory-one',
    mode: 'home',
    settingsSection: 'overview',
    userLabel: 'owner@example.com',
    onSelectTenant: () => {},
    onSelectInventory: () => {},
    onModeChange: () => {},
    onSignOut: () => {},
    ...overrides
  };
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
