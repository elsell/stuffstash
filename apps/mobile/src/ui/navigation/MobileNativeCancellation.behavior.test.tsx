import React from 'react';
import { Text } from 'react-native';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { ApiInventoryDirectory } from '../../adapters/inventories/ApiInventoryDirectory';
import { createMobileQueryClient, mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { MobileServerStateProvider } from './MobileServerStateProvider';
import { useMobileInventoryServerQuery } from '../serverState/useMobileInventoryServerQuery';

const settle = (h: MobileRenderHarness) => h.run(() => new Promise(resolve => setTimeout(resolve, 10)));
const page = <T,>(items: T[]) => ({ items, pagination: { limit: 100, hasMore: false, nextCursor: null } });
const tenant = { id: 'tenant', name: 'Home', access: { relationship: 'owner' as const, permissions: ['view'] } };
const inventory = { id: 'inventory', tenantId: tenant.id, name: 'Garage', access: tenant.access };

it('loads shared inventory reads with the AbortController shipped by React Native', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); client.setDefaultOptions({ queries: { retry: false } }); let reads = 0;
  const directory = new ApiInventoryDirectory({ listMyTenants: async () => page([tenant]), listInventories: async () => page([inventory]) }, tenant.id);
  function Surface() {
    const result = useMobileInventoryServerQuery({ key: mobileQueryKeys.home, query: async signal => { reads++; return (await directory.selected(signal)).inventory.name; } });
    return <Text>{result.data ?? result.error?.message ?? 'Loading'}</Text>;
  }
  try {
    expect(typeof new AbortController().signal.throwIfAborted).toBe('undefined');
    await h.render(<MobileServerStateProvider client={client} scopeId="native" loadInventoryScope={async request => { const selected = await directory.selected(request?.signal); return { tenantId: selected.tenant.id, inventoryId: selected.inventory.id }; }}><Surface /></MobileServerStateProvider>);
    await settle(h); await settle(h);
    expect(h.allText()).toEqual(['Garage']); expect(reads).toBe(1);
  } finally { await h.unmount(); }
});

it('stops native discovery before network work when cancelled without a reason', async () => {
  const controller = new AbortController(); controller.abort(); let reads = 0;
  const directory = new ApiInventoryDirectory({ listMyTenants: async () => { reads++; return page([tenant]); }, listInventories: async () => page([inventory]) }, tenant.id);
  await expect(directory.selected(controller.signal)).rejects.toMatchObject({ name: 'AbortError' });
  expect(reads).toBe(0);
});
