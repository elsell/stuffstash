import { expect, it } from 'vitest';
import { SelectedInventoryUnavailableError } from '../../application/shared/SelectedInventoryUnavailableError';
import { ApiInventoryDirectory } from './ApiInventoryDirectory';
const tenant = { id: 'tenant', name: 'Home', access: { relationship: 'owner' as const, permissions: ['view'] } };
const inventory = { id: 'inventory', tenantId: tenant.id, name: 'Garage', access: tenant.access };
const page = <T,>(items: T[], nextCursor: string | null = null) => ({ items, pagination: { limit: 100, hasMore: Boolean(nextCursor), nextCursor } });

it('shares discovery without allowing one cancelled consumer to abort another', async () => {
  let resolve!: (value: ReturnType<typeof page<typeof tenant>>) => void; let transportSignal: AbortSignal | undefined; let calls = 0;
  const directory = new ApiInventoryDirectory({ listMyTenants: async (_limit, _cursor, signal) => { calls++; transportSignal = signal; return new Promise(r => { resolve = r; }); }, listInventories: async () => page([inventory]) }, 'tenant');
  const a = new AbortController(); const b = new AbortController();
  const first = directory.selected(a.signal); const second = directory.selected(b.signal);
  a.abort(); await expect(first).rejects.toThrow(); expect(transportSignal?.aborted).toBe(false);
  resolve(page([tenant])); expect((await second).inventory.id).toBe('inventory'); expect(calls).toBe(1);
});

it('cancels the final consumer and does not retain a late cancelled result', async () => {
  let resolve!: (value: ReturnType<typeof page<typeof tenant>>) => void; const signals: AbortSignal[] = []; let calls = 0;
  const directory = new ApiInventoryDirectory({ listMyTenants: async (_limit, _cursor, signal) => { calls++; signals.push(signal!); return calls === 1 ? new Promise(r => { resolve = r; }) : page([tenant]); }, listInventories: async () => page([inventory]) }, 'tenant');
  const a = new AbortController(); const first = directory.selected(a.signal); a.abort(); await expect(first).rejects.toThrow(); expect(signals[0]?.aborted).toBe(true);
  const second = directory.selected(); resolve(page([{ ...tenant, id: 'obsolete' }]));
  expect((await second).tenant.id).toBe('tenant'); expect(calls).toBe(2);
});

it('follows directory pages and refuses to silently retarget a removed selection', async () => {
  let now = 0; let removed = false; const cursors: (string | undefined)[] = [];
  const directory = new ApiInventoryDirectory({ listMyTenants: async (_limit, cursor) => { cursors.push(cursor); return cursor ? page([tenant]) : page([], 'next'); }, listInventories: async () => page(removed ? [{ ...inventory, id: 'other' }] : [inventory]) }, 'tenant', () => now);
  await directory.select('inventory'); expect((await directory.selected()).inventory.id).toBe('inventory'); expect(cursors).toEqual([undefined, 'next']);
  removed = true; now = 300_001;
  await expect(directory.selected()).rejects.toBeInstanceOf(SelectedInventoryUnavailableError);
});

it('keeps warmed command identity off the discovery-refresh critical path', async () => {
  let now = 0; let reads = 0;
  const directory = new ApiInventoryDirectory({ listMyTenants: async () => { reads++; return page([tenant]); }, listInventories: async () => page([inventory]) }, 'tenant', () => now);
  await directory.selected(); now = 300_001;
  expect((await directory.selectedForCommand()).inventory.id).toBe('inventory'); expect(reads).toBe(1);
  await directory.load(); expect(reads).toBe(2);
});
