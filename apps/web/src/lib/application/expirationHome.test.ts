import { expect, it } from 'vitest';
import { ExpirationHomeQuery } from './expirationHome';
import type { ExpirationRepository, ExpirationPage } from '$lib/ports/expirationRepository';
const page:ExpirationPage={items:[],counts:{all:300,expired:200,soon:100},timezone:'UTC',hasMore:false,nextCursor:null};
it('clears the prior scope immediately and does not restore it after a rejected new scope load',async()=>{
 let reject!:(error:Error)=>void;
 const query=new ExpirationHomeQuery({list:async(_t,i)=>i==='first'?page:new Promise<ExpirationPage>((_resolve,r)=>{reject=r;})},()=>{});
 await query.load('tenant','first');const pending=query.load('tenant','second');expect(query.state.page).toBeUndefined();reject(new Error('offline'));await pending;expect(query.state.page).toBeUndefined();expect(query.state.error).toBeTruthy();
});
it('reports the real failed refresh result and retains same-scope data',async()=>{
 let fail=false;const query=new ExpirationHomeQuery({list:async()=>{if(fail)throw new Error('offline');return page;}},()=>{});
 expect(await query.load('t','i')).toBe(true);fail=true;expect(await query.load('t','i')).toBe(false);expect(query.state.page?.counts.all).toBe(300);
});
