import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { createMobileQueryClient, mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { ExpirationHomeContent } from './ExpirationHomeContent';
it('gates navigation and cached counts through scope loading, failure and recovery', async () => {
 const h = new MobileRenderHarness(); const client = createMobileQueryClient();
 let resolve!: (scope: { tenantId: string; inventoryId: string }) => void;
 let fail = false;
 const opened: string[] = [];
 const key = mobileQueryKeys.inventoryScope('scope');
 const load = async () => { if (fail) throw new Error('Scope unavailable'); return new Promise<{tenantId:string;inventoryId:string}>(done => { resolve = done; }); };
 const settle = () => h.run(() => new Promise(done => setTimeout(done, 15)));
 try {
  await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={load}>
   <ExpirationHomeContent query={{ home: async () => ({ items: [], counts: { soon: 2, expired: 1, all: 3 }, timezone: 'UTC' }) }} onOpen={(tenant,inventory,mode)=>opened.push(`${tenant}/${inventory}/${mode}`)} onOpenAsset={()=>{}} />
  </MobileServerStateProvider>);
  await h.press(h.byLabel('View all expiration dates')); expect(opened).toEqual([]);
  await h.run(()=>resolve({tenantId:'tenant',inventoryId:'inventory'})); await settle(); await settle();
  await h.press(h.byLabel('View 1 expired items')); expect(opened).toEqual(['tenant/inventory/expired']);
  const retainedCount = h.byLabel('View 1 expired items');
  fail = true; await h.run(()=>client.refetchQueries({queryKey:key,exact:true})); await settle();
  expect(h.byLabel('View 1 expired items')).toBeUndefined();
  await h.press(retainedCount); expect(opened).toHaveLength(1);
  await h.press(h.byLabel('View all expiration dates')); expect(opened).toHaveLength(1);
  fail = false; await h.press(h.byLabel('Retry expiration')); await settle();
  expect(h.byLabel('View 1 expired items')).toBeUndefined();
  await h.run(()=>resolve({tenantId:'tenant',inventoryId:'inventory'})); await settle(); await settle();
  await h.press(h.byLabel('View all expiration dates')); expect(opened.at(-1)).toBe('tenant/inventory/all');
 } finally { await h.unmount(); }
});
