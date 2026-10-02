import { expect, it } from 'vitest';
const tenant = { id: 'tenant', name: 'Home', access: { relationship: 'owner' as const, permissions: ['view'] } };
const inventory = { id: 'inventory', tenantId: tenant.id, name: 'Garage', access: tenant.access };
const page = <T,>(items: T[], nextCursor: string | null = null) => ({ items, pagination: { limit: 100, hasMore: Boolean(nextCursor), nextCursor } });

it('localizes rejected selection and malformed discovery without changing selection', async () => {
  const previous = process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE;
  process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE = 'en-XA';
  try {
    const { ApiInventoryDirectory } = await import('./ApiInventoryDirectory');
    const { SelectedInventoryUnavailableError } = await import('../../application/shared/SelectedInventoryUnavailableError');
    const directory = new ApiInventoryDirectory({
      listMyTenants: async () => page([tenant]), listInventories: async () => page([inventory])
    }, tenant.id);
    await directory.select(inventory.id);
    const unavailable = await directory.select('missing').catch(error => error);
    expect(unavailable).toBeInstanceOf(SelectedInventoryUnavailableError);
    expect(unavailable.message).toMatch(/^\[/);
    expect((await directory.selected()).inventory.id).toBe(inventory.id);
    for (const repeated of [true, false]) {
      let reads = 0;
      const broken = new ApiInventoryDirectory({
        listMyTenants: async () => page([], repeated ? 'repeated' : String(++reads)),
        listInventories: async () => page([])
      }, tenant.id);
      await expect(broken.load()).rejects.toThrow(/^\[/);
      if (!repeated) expect(reads).toBe(1000);
    }
  } finally {
    if (previous === undefined) delete process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE;
    else process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE = previous;
  }
});
