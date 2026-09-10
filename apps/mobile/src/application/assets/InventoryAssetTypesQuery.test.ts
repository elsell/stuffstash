import { expect, it } from 'vitest';
import { InventoryAssetTypesQuery } from './InventoryAssetTypesQuery';
import type { CustomAssetTypeDefinition } from '../../domain/customization/Customization';
import type { CustomizationContext } from '../customization/CustomizationRepository';

const context: CustomizationContext = { tenantId: 'tenant', inventoryId: 'inventory', tenantName: '', inventoryName: '', tenantPermissions: ['view'], inventoryPermissions: ['view'] };
function type(id: string, scope: 'tenant' | 'inventory'): CustomAssetTypeDefinition {
  return { id, scope, kind: 'asset-type', tenantId: 'tenant', inventoryId: scope === 'inventory' ? 'inventory' : undefined,
    key: id, displayName: id, description: '', lifecycle: 'active', expirationEnabled: true };
}
it('combines authorized tenant and inventory types while filtering foreign or inactive records', async () => {
  const query = new InventoryAssetTypesQuery({ execute: async () => context }, {
    assetTypes: async (_context, scope) => {
      if (scope === 'tenant') throw new Error('Tenant configuration permission denied');
      return { complete: true, items: [type('inventory', 'inventory'), type('tenant', 'tenant'), { ...type('foreign', 'inventory'), tenantId: 'other' }, { ...type('archived', 'inventory'), lifecycle: 'archived' }] };
    }
  });
  const result = await query.execute('tenant', 'inventory');
  expect(result.map((item) => item.id)).toEqual(['inventory', 'tenant']);
  expect(result.every((item) => item.expirationEnabled)).toBe(true);
});
it('refuses a changed scope and incomplete type lists', async () => {
  let calls = 0;
  const query = new InventoryAssetTypesQuery({ execute: async () => context }, {
    assetTypes: async () => { calls++; return { complete: false, items: [] }; }
  });
  await expect(query.execute('other', 'inventory')).rejects.toThrow('Inventory changed');
  expect(calls).toBe(0);
  await expect(query.execute('tenant', 'inventory')).rejects.toThrow('complete');
});

it('does not return a late collection after cancellation', async () => {
  const controller = new AbortController();
  const query = new InventoryAssetTypesQuery({ execute: async () => context }, {
    assetTypes: async () => { controller.abort(); return { complete: true, items: [] }; }
  });
  await expect(query.execute('tenant', 'inventory', { signal: controller.signal })).rejects.toThrow();
});
