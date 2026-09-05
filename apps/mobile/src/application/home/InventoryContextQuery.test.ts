import { describe, expect, it } from 'vitest';
import { InventoryContextQuery } from './InventoryContextQuery';

describe('InventoryContextQuery', () => {
  it('reads selected identity and create permission without workspace hydration', async () => {
    const controller = new AbortController();
    let signal: AbortSignal | undefined;
    const query = new InventoryContextQuery({ getCurrentSettingsScope: async (request) => {
      signal = request?.signal;
      return { tenantId: 'tenant', inventory: { id: 'inventory', name: 'Home', permissions: ['view', 'create_asset'] } };
    } });
    await expect(query.execute({ signal: controller.signal })).resolves.toEqual({ inventoryName: 'Home', canAdd: true });
    expect(signal).toBe(controller.signal);
  });
});
