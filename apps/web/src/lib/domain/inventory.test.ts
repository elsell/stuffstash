import { describe, expect, it } from 'vitest';
import { canCreateAsset, canEditInventory } from './inventory';
import type { Inventory } from './inventory';

describe('inventory permissions', () => {
  it('distinguishes asset creation from broader edit capability', () => {
    const editOnlyInventory: Inventory = {
      id: 'inventory-one',
      tenantId: 'tenant-one',
      name: 'Household',
      access: { relationship: 'editor', permissions: ['view', 'edit_asset'] }
    };

    expect(canEditInventory(editOnlyInventory)).toBe(true);
    expect(canCreateAsset(editOnlyInventory)).toBe(false);
  });
});
