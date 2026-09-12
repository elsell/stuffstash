import { afterEach, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import { flushSync } from 'svelte';
import ExpirationWorkspace from './ExpirationWorkspace.svelte';
import { expirationWorkspaceContext, type ExpirationRepository } from '$lib/ports/expirationRepository';
let component: ReturnType<typeof mount> | undefined;
afterEach(async () => { if (component) await unmount(component); document.body.innerHTML = ''; });
async function settle() { for (let i=0;i<5;i++) { await Promise.resolve(); flushSync(); } }
it('keeps filters available for an empty inventory and uses authoritative totals', async () => {
 const repository: ExpirationRepository = { list: async () => ({items:[],counts:{all:0,soon:0,expired:0},timezone:'UTC',hasMore:false,nextCursor:null}), choices: async () => ({types:[],tags:[],locations:[]}) };
 component = mount(ExpirationWorkspace, {target:document.body,context:new Map([[expirationWorkspaceContext,{repository}]]),props:{tenantId:'t',inventoryId:'i',filter:{mode:'all'},onNavigate:()=>{},onOpenAsset:()=>{}}});
 await settle(); expect(document.body.textContent).toContain('No expiration dates'); expect([...document.querySelectorAll('button')].some(button => button.textContent?.includes('Filters'))).toBe(true);
});
it('shows retry rather than an empty state when the initial query fails', async () => {
 const repository = {choices:async()=>({types:[],tags:[],locations:[]}),list:async()=>{throw new Error('offline');}} as ExpirationRepository;
 component = mount(ExpirationWorkspace,{target:document.body,context:new Map([[expirationWorkspaceContext,{repository}]]),props:{tenantId:'t',inventoryId:'i',filter:{mode:'all'},onNavigate:()=>{},onOpenAsset:()=>{}}});
 await settle(); expect(document.querySelector('[role=alert]')?.textContent).toContain('could not'); expect(document.body.textContent).not.toContain('No expiration dates');
});
