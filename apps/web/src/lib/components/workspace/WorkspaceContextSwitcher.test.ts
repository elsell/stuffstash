import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount, type ComponentProps } from 'svelte';
import type { Inventory, Tenant } from '$lib/domain/inventory';
import WorkspaceContextSwitcher from './WorkspaceContextSwitcher.svelte';

let component: ReturnType<typeof mount> | null = null;

const tenant: Tenant = {
  id: 'tenant-one',
  name: 'Household',
  access: { relationship: 'owner', permissions: ['view'] }
};

const inventory: Inventory = {
  id: 'inventory-one',
  tenantId: tenant.id,
  name: 'Garage',
  access: { relationship: 'owner', permissions: ['view'] }
};

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('WorkspaceContextSwitcher', () => {
  it('exposes a durable settings link from the desktop context switcher', () => {
    let openedSettings = false;
    component = mount(WorkspaceContextSwitcher, {
      target: document.body,
      props: contextProps({
        onOpenSettings: () => {
          openedSettings = true;
        }
      })
    });

    const settingsLink = linkContaining('Inventory settings');
    expect(settingsLink.getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/settings');

    settingsLink.click();

    expect(openedSettings).toBe(true);
  });

  it('exposes a durable settings link from the mobile context sheet', async () => {
    component = mount(WorkspaceContextSwitcher, {
      target: document.body,
      props: contextProps({ mobile: true })
    });

    buttonContaining('Garage').click();
    await tick();

    expect(linkContaining('Inventory settings').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/settings');
  });
});

function contextProps(
  overrides: Partial<ComponentProps<typeof WorkspaceContextSwitcher>> = {}
): ComponentProps<typeof WorkspaceContextSwitcher> {
  return {
    tenants: [tenant],
    inventories: [inventory],
    selectedTenantId: tenant.id,
    selectedInventoryId: inventory.id,
    onSelectTenant: () => {},
    onSelectInventory: () => {},
    onOpenSettings: () => {},
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

function linkContaining(text: string): HTMLAnchorElement {
  const link = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!link) {
    throw new Error(`Missing link containing ${text}`);
  }
  return link;
}
