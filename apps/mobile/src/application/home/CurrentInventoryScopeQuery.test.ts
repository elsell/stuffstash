import { describe, expect, it } from 'vitest';
import { CurrentInventoryScopeQuery } from './CurrentInventoryScopeQuery';

describe('CurrentInventoryScopeQuery', () => {
  it('returns the explicit tenant and inventory identity from its port', async () => {
    const query = new CurrentInventoryScopeQuery({
      getCurrentInventoryScope: async () => ({
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home'
      })
    });

    await expect(query.execute()).resolves.toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home'
    });
  });
});
