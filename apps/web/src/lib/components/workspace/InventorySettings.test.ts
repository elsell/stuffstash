import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import InventorySettings from './InventorySettings.svelte';
import type { Inventory, Tenant } from '$lib/domain/inventory';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('InventorySettings', () => {
  it('uses edit_asset access for asset edit status and disables unsupported entry points', () => {
    const tenant: Tenant = {
      id: 'tenant-one',
      name: 'Household',
      access: { relationship: 'owner', permissions: ['view', 'configure'] }
    };
    const inventory: Inventory = {
      id: 'inventory-one',
      tenantId: 'tenant-one',
      name: 'Garage',
      access: {
        relationship: 'editor',
        permissions: ['view', 'share', 'configure', 'edit_asset']
      }
    };

    component = mount(InventorySettings, {
      target: document.body,
      props: { tenant, inventory, inventoryCount: 2 }
    });

    expect(document.body.textContent).toContain('Asset editsAllowed');
    expect(document.body.textContent).toContain('Direct grants and invitations are planned');
    expect(
      Array.from(document.body.querySelectorAll('button')).map((button) => ({
        text: button.textContent,
        disabled: button.disabled
      }))
    ).toEqual([
      { text: 'Manage sharing unavailable', disabled: true },
      { text: 'View activity unavailable', disabled: true },
      { text: 'Manage fields unavailable', disabled: true },
      { text: 'Tenant administration unavailable', disabled: true }
    ]);
  });

  it('does not treat create-only access as asset edit access', () => {
    const inventory: Inventory = {
      id: 'inventory-one',
      tenantId: 'tenant-one',
      name: 'Garage',
      access: {
        relationship: 'editor',
        permissions: ['view', 'create_asset']
      }
    };

    component = mount(InventorySettings, {
      target: document.body,
      props: { tenant: null, inventory, inventoryCount: 1 }
    });

    expect(document.body.textContent).toContain('Asset editsView only');
  });
});
