import { expect, it } from 'vitest';
import { ExpirationList } from './expirationList';
import type { ExpirationPage, ExpirationRepository } from '$lib/ports/expirationRepository';
const page: ExpirationPage = { items: [], counts: { all: 320, soon: 20, expired: 300 }, timezone: 'UTC', hasMore: false, nextCursor: null };
it('ignores a late result from a replaced filter', async () => {
 let resolve!: (page: ExpirationPage) => void;
 const repository = { choices: async () => ({ types: [], tags: [], locations: [] }), list: async (_t: string, _i: string, filter: { mode: string }) => filter.mode === 'expired' ? new Promise<ExpirationPage>(r => { resolve = r; }) : { ...page, counts: { all: 2, soon: 2, expired: 0 } } } satisfies ExpirationRepository;
 const list = new ExpirationList(repository, () => {});
 const old = list.load('tenant', 'inventory', { mode: 'expired' });
 await list.load('tenant', 'inventory', { mode: 'soon' }); resolve(page); await old;
 expect(list.state.page?.counts.all).toBe(2);
});
it('keeps prior rows on refresh failure and clears them on permission loss', async () => {
 let error: unknown;
 const list = new ExpirationList({ choices: async () => ({ types: [], tags: [], locations: [] }), list: async () => { if (error) throw error; return page; } } satisfies ExpirationRepository, () => {});
 await list.load('t', 'i', { mode: 'all' }); error = new Error('offline'); await list.load('t', 'i', { mode: 'all' });
 expect(list.state.page?.counts.all).toBe(320); expect(list.state.error).toBeTruthy();
 error = { status: 403 }; await list.load('t', 'i', { mode: 'all' }); expect(list.state.page).toBeUndefined();
});
it('never presents a partial cursor response as a complete list', async () => {
 const list = new ExpirationList({ choices: async () => ({ types: [], tags: [], locations: [] }), list: async () => ({ ...page, hasMore: true, nextCursor: null }) } satisfies ExpirationRepository, () => {});
 await list.load('t', 'i', { mode: 'all' }); expect(list.state.error).toBeTruthy();
});
it('permission loss clears cached modes for the entire inventory',async()=>{
 const cache=new Map<string,ExpirationPage>();let denied=false;
 const repository={choices:async()=>({types:[],tags:[],locations:[]}),list:async()=>{if(denied)throw {status:403};return page;}} satisfies ExpirationRepository;
 const list=new ExpirationList(repository,()=>{},cache);await list.load('t','i',{mode:'all'});await list.load('t','i',{mode:'soon'});expect(cache.size).toBe(2);
 denied=true;await list.load('t','i',{mode:'soon'});expect(cache.size).toBe(0);
});
