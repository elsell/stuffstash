import { afterEach, describe, expect, it, vi } from 'vitest';
import { mount, tick, unmount, type ComponentProps } from 'svelte';
import type { Inventory, Tenant } from '$lib/domain/inventory';
import WorkspaceContextSwitcher from './WorkspaceContextSwitcher.svelte';
import WorkspaceContextSwitcherHarness from './WorkspaceContextSwitcherHarness.test.svelte';

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

const secondTenant: Tenant = {
  id: 'tenant-two',
  name: 'Workshop',
  access: { relationship: 'editor', permissions: ['view'] }
};

const secondInventory: Inventory = {
  id: 'inventory-two',
  tenantId: secondTenant.id,
  name: 'Loft',
  access: { relationship: 'editor', permissions: ['view'] }
};

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
  vi.useRealTimers();
});

describe('WorkspaceContextSwitcher', () => {
  it('keeps the desktop switcher collapsed to one row until opened', async () => {
    component = mount(WorkspaceContextSwitcher, {
      target: document.body,
      props: contextProps({
        tenants: [tenant, secondTenant],
        inventories: [inventory, secondInventory]
      })
    });

    expect(buttonContaining('Garage').textContent).toContain('Household');
    expect(document.body.textContent).not.toContain('Owner');
    expect(document.body.textContent).not.toContain('Loft');
    expect(document.body.textContent).not.toContain('Inventory settings');

    buttonContaining('Garage').click();
    await tick();

    expect(document.body.textContent).toContain('Inventories');
    expect(pressedButtonContaining('Garage').getAttribute('aria-pressed')).toBe('true');
    expect(document.body.textContent).toContain('Owner');
    expect(document.body.textContent).not.toContain('Loft');
  });

  it('keeps the popover open and replaces inventories after switching tenants', async () => {
    component = mount(WorkspaceContextSwitcherHarness, {
      target: document.body,
      props: {
        tenants: [tenant, secondTenant],
        inventories: [inventory, secondInventory],
        initialTenantId: tenant.id,
        initialInventoryId: inventory.id
      }
    });

    buttonContaining('Garage').click();
    await tick();
    buttonContaining('Switch tenant').click();
    await tick();
    buttonContaining('Workshop').click();
    await tick();

    expect(document.body.textContent).toContain('Inventories');
    expect(document.body.textContent).toContain('Loft');
    expect(document.body.textContent).not.toContain('GarageHousehold');
    expect(pressedButtonContaining('Loft').getAttribute('aria-pressed')).toBe('true');
  });

  it('focuses the replacement inventory list after an async tenant switch', async () => {
    vi.useFakeTimers();
    component = mount(WorkspaceContextSwitcherHarness, {
      target: document.body,
      props: {
        tenants: [tenant, secondTenant],
        inventories: [inventory, secondInventory],
        initialTenantId: tenant.id,
        initialInventoryId: inventory.id,
        asyncTenantUpdate: true
      }
    });

    buttonContaining('Garage').click();
    await tick();
    buttonContaining('Switch tenant').click();
    await tick();
    buttonContaining('Workshop').click();
    await tick();

    vi.runAllTimers();
    await tick();
    await tick();

    expect(document.body.textContent).toContain('Loft');
    expect(document.activeElement?.textContent).toContain('Loft');
  });

  it('closes the desktop popover when focus leaves the switcher', async () => {
    vi.useFakeTimers();
    const outsideButton = document.createElement('button');
    outsideButton.textContent = 'Outside';
    document.body.append(outsideButton);
    component = mount(WorkspaceContextSwitcher, {
      target: document.body,
      props: contextProps({
        tenants: [tenant, secondTenant],
        inventories: [inventory, secondInventory]
      })
    });

    buttonContaining('Garage').click();
    await tick();
    expect(document.body.textContent).toContain('Inventories');

    pressedButtonContaining('Garage').dispatchEvent(
      new FocusEvent('focusout', { bubbles: true, relatedTarget: outsideButton })
    );
    outsideButton.focus();
    vi.runAllTimers();
    await tick();

    expect(document.body.textContent).not.toContain('Inventories');
  });

  it('uses the same one-row trigger and tenant-first sheet structure on mobile', async () => {
    component = mount(WorkspaceContextSwitcher, {
      target: document.body,
      props: contextProps({
        mobile: true,
        tenants: [tenant, secondTenant],
        inventories: [inventory, secondInventory]
      })
    });

    expect(document.body.textContent).not.toContain('Inventory settings');

    buttonContaining('Garage').click();
    await tick();

    expect(document.body.textContent).toContain('Inventories');
    expect(pressedButtonContaining('Garage').getAttribute('aria-pressed')).toBe('true');
    expect(document.activeElement?.textContent).toContain('Garage');
    expect(document.body.textContent).not.toContain('Loft');

    buttonContaining('Switch tenant').click();
    await tick();

    expect(document.body.textContent).toContain('Tenants');
    expect(document.body.textContent).toContain('Workshop');
    expect(buttonContaining('Back')).toBeTruthy();
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

function pressedButtonContaining(text: string): HTMLButtonElement {
  const button = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button[aria-pressed]')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!button) {
    throw new Error(`Missing pressed button containing ${text}`);
  }
  return button;
}
