import { expect, it } from 'vitest';
import { groupExpirationItems } from './ExpirationSections';
import { ExpirationWorkspaceQuery } from './ExpirationWorkspaceQuery';
import { assetId, type AssetSummary } from '../../domain/assets/AssetSummary';
const item = (id:string,state:'upcoming'|'expired',date:string):AssetSummary => ({id:assetId(id),title:id,kind:'item',lifecycleState:'active',description:'',locationLabel:'Cabinet',locationTrail:[],parentLocationTrail:[],updatedAtLabel:'',hasPhoto:false,expiration:{date,precision:date.length===7?'month':'day'},expirationContext:{state,trackingEnabled:true,advanceDays:7,timezone:'UTC'}});
it('Home includes both attention groups and never truncates counts to preview rows',async()=>{
 const query=new ExpirationWorkspaceQuery({async list(_t,_i,filter){return {items:filter.mode==='soon'?[item('next','upcoming','2026-09')]:[item('old-a','expired','2026-08'),item('old-b','expired','2026-07')],counts:{soon:1,expired:200,all:350},timezone:'UTC',nextCursor:'next',hasMore:true};}}, {record(){}});
 const home=await query.home('tenant','inventory');
 expect(home.items.map(item=>item.id)).toEqual(['old-a','next','old-b']);expect(home.counts.all).toBe(350);
 expect(home.items[1].expiration).toEqual({date:'2026-09',precision:'month'});
});
it('does not publish results after a canceled inventory read',async()=>{
 const abort=new AbortController();const query=new ExpirationWorkspaceQuery({async list(){abort.abort();return {items:[],counts:{soon:0,expired:0,all:0},timezone:'UTC',nextCursor:null,hasMore:false};}},{record(){}});
 await expect(query.list('tenant','inventory',{mode:'all'},abort.signal)).rejects.toThrow();
});
it('groups expired and future items in separate month sections without inventing dates',async()=>{
 const query=new ExpirationWorkspaceQuery({async list(){return {items:[item('old','expired','2026-08'),item('next','upcoming','2026-09')],counts:{soon:1,expired:1,all:2},timezone:'UTC',nextCursor:null,hasMore:false};}},{record(){}});
 const page=await query.list('tenant','inventory',{mode:'all'});
 expect(groupExpirationItems(page.items).map(group=>group.key)).toEqual(['expired:2026-08','future:2026-09']);
 expect(page.items[0].expiration?.date).toBe('2026-08');expect(page.items[1].expiration?.date).toBe('2026-09');
});
