import { describe, expect, it } from 'vitest';
import type { Inventory, Tenant } from '$lib/domain/inventory';
import {
  contextSwitcherPresentation,
  inventoryContextOptions,
  inventoryCountLabel,
  relationshipLabel,
  tenantContextOptions
} from './workspaceContextSwitching';

const household: Tenant = {
  id: 'tenant-one',
  name: 'Household',
  access: { relationship: 'owner', permissions: ['view'] }
};

const workshop: Tenant = {
  id: 'tenant-two',
  name: 'Workshop',
  access: { relationship: 'member', permissions: ['view'] }
};

const inventories: Inventory[] = [
  inventory('garage', household.id, 'Garage', 'owner'),
  inventory('pantry', household.id, 'Pantry', 'shared_editor'),
  inventory('loft', workshop.id, 'Loft', undefined)
];

describe('workspace context switching helpers', () => {
  it('builds tenant options with count labels and selected state', () => {
    expect(tenantContextOptions({ tenants: [household, workshop], inventories, selectedTenantId: workshop.id })).toEqual([
      { id: household.id, name: 'Household', inventoryCountLabel: '2 inventories', selected: false },
      { id: workshop.id, name: 'Workshop', inventoryCountLabel: '1 inventory', selected: true }
    ]);
  });

  it('builds selected-tenant inventory options with canonical hrefs and relationship labels', () => {
    expect(
      inventoryContextOptions({
        tenants: [household, workshop],
        inventories,
        selectedTenantId: household.id,
        selectedInventoryId: 'pantry'
      })
    ).toEqual([
      {
        id: 'garage',
        tenantId: household.id,
        name: 'Garage',
        tenantName: 'Household',
        relationshipLabel: 'Owner',
        href: '/tenants/tenant-one/inventories/garage',
        selected: false
      },
      {
        id: 'pantry',
        tenantId: household.id,
        name: 'Pantry',
        tenantName: 'Household',
        relationshipLabel: 'Shared Editor',
        href: '/tenants/tenant-one/inventories/pantry',
        selected: true
      }
    ]);
  });

  it('formats count and relationship fallback labels', () => {
    expect(inventoryCountLabel(inventories, 'missing')).toBe('0 inventories');
    expect(relationshipLabel(undefined)).toBe('Member');
    expect(relationshipLabel('tenant-admin')).toBe('Tenant Admin');
  });

  it('builds context switcher fallback presentation', () => {
    expect(contextSwitcherPresentation({ selectedTenant: household, selectedInventory: inventories[0] })).toEqual({
      triggerInventoryLabel: 'Garage',
      triggerTenantLabel: 'Household',
      activeTenantLabel: 'Household',
      emptyInventoryMessage: 'No inventories in this tenant.'
    });
    expect(contextSwitcherPresentation({ selectedTenant: null, selectedInventory: null })).toEqual({
      triggerInventoryLabel: 'No inventory',
      triggerTenantLabel: 'No tenant',
      activeTenantLabel: 'No tenant',
      emptyInventoryMessage: 'No inventories in this tenant.'
    });
  });
});

function inventory(id: string, tenantId: string, name: string, relationship: string | undefined): Inventory {
  return {
    id,
    tenantId,
    name,
    access: { relationship: relationship ?? '', permissions: ['view'] }
  };
}
