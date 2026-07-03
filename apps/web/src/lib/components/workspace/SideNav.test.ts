import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import SideNav from './SideNav.svelte';
import type { Inventory, Tenant, WorkspaceMode } from '$lib/domain/inventory';

let component: ReturnType<typeof mount> | null = null;

type SideNavProps = {
  tenants: Tenant[];
  inventories: Inventory[];
  selectedTenantId: string;
  selectedInventoryId: string;
  mode: WorkspaceMode;
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

    const currentDestinations = document.body.querySelectorAll<HTMLButtonElement>('button[aria-current="page"]');
    expect(currentDestinations).toHaveLength(1);
    const current = currentDestinations[0];
    expect(current?.textContent).toContain('Settings');
    expect(current?.textContent).toContain('Access, fields, and audit');
  });

  it('marks home as the current primary destination', () => {
    component = mount(SideNav, {
      target: document.body,
      props: sideNavProps({ mode: 'home' })
    });

    const currentDestinations = document.body.querySelectorAll<HTMLButtonElement>('button[aria-current="page"]');
    expect(currentDestinations).toHaveLength(1);
    expect(currentDestinations[0]?.textContent).toContain('Home');
    expect(currentDestinations[0]?.textContent).toContain('Recent assets and places');
  });

  it('marks focused location routes under the locations destination', () => {
    component = mount(SideNav, {
      target: document.body,
      props: sideNavProps({ mode: 'location' })
    });

    const currentDestinations = document.body.querySelectorAll<HTMLButtonElement>('button[aria-current="page"]');
    expect(currentDestinations).toHaveLength(1);
    expect(currentDestinations[0]?.textContent).toContain('Locations');
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

    buttonContaining('Locations').click();
    expect(selectedMode).toBe('locations');

    buttonContaining('Import').click();
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
    userLabel: 'owner@example.com',
    onSelectTenant: () => {},
    onSelectInventory: () => {},
    onModeChange: () => {},
    onSignOut: () => {},
    ...overrides
  };
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
