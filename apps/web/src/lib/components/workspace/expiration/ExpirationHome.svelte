<script lang="ts">
 import { getContext, untrack, onDestroy } from 'svelte';
 import { ExpirationHomeQuery, type ExpirationHomeState } from '$lib/application/expirationHome';
 import type { Asset } from '$lib/domain/inventory';
 import { expirationWorkspaceContext, type ExpirationWorkspace, type ExpirationPage, type ExpirationMode } from '$lib/ports/expirationRepository';
 import { workspaceRouteHref, type WorkspaceRouteState } from '$lib/application/workspaceRoute';
 import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
 import * as Button from '$lib/components/ui/button/index.js';
 import ExpirationRows from './ExpirationRows.svelte';
 import ExpirationRefresh from '../ExpirationRefresh.svelte';
 let {tenantId,inventoryId,onNavigate,onOpenAsset}:{tenantId:string;inventoryId:string;onNavigate?:(route:Partial<WorkspaceRouteState>)=>void;onOpenAsset:(asset:Asset)=>void}=$props();
 const workspace=getContext<ExpirationWorkspace|undefined>(expirationWorkspaceContext);
 let home=$state<ExpirationHomeState>({error:''});
 const query=workspace?new ExpirationHomeQuery(workspace.repository,next=>{home=next;},workspace.observer):undefined;
 let page=$derived(home.page);let error=$derived(home.error);
 $effect(()=>{const tenant=tenantId,inventory=inventoryId;workspace?.revision?.();untrack(()=>{void query?.load(tenant,inventory);});});
 onDestroy(()=>query?.dispose());
 async function refresh(){return await query?.load(tenantId,inventoryId) ?? false;}
 function route(mode:ExpirationMode):Partial<WorkspaceRouteState>{return{mode:'expiration',tenantId,inventoryId,expirationFilter:{mode}};}
 function open(event:MouseEvent,mode:ExpirationMode){if(onNavigate&&shouldHandleWorkspaceLinkClick(event)){event.preventDefault();onNavigate(route(mode));}}
</script>
{#if workspace && (error || !page || page.counts.all>0)}
 <section class="expiration-home" aria-label="Expiration"><header><h2>Expiration</h2><Button.Root variant="ghost" href={workspaceRouteHref(route('all'),tenantId,inventoryId)} onclick={event=>open(event,'all')}>See all</Button.Root></header>
 {#if error}<p role="alert">{error}</p><Button.Root variant="outline" onclick={()=>{void refresh();}}>Retry expiration</Button.Root>{/if}
 {#if !page&&!error}<p role="status">Loading expiration…</p>{/if}
 {#if page}
 <ExpirationRefresh assets={page.items} timezone={page.timezone} scope={`${tenantId}/${inventoryId}`} onRefresh={refresh} />
 {#if page.counts.expired+page.counts.soon>0}<div class="counts"><Button.Root variant="ghost" href={workspaceRouteHref(route('expired'),tenantId,inventoryId)} onclick={event=>open(event,'expired')}>Expired · {page.counts.expired}</Button.Root><Button.Root variant="ghost" href={workspaceRouteHref(route('soon'),tenantId,inventoryId)} onclick={event=>open(event,'soon')}>Expiring soon · {page.counts.soon}</Button.Root></div><ExpirationRows items={page.items} grouped={false} {onOpenAsset} />{:else}<p class="quiet">None expiring soon</p>{/if}
 {/if}</section>
{/if}
<style>.expiration-home{margin-bottom:2rem;}header{display:flex;align-items:center;justify-content:space-between;gap:1rem;}h2{font-size:var(--text-section-size);font-weight:600;}.counts{display:flex;gap:.75rem;flex-wrap:wrap;}.quiet{color:var(--muted-foreground);padding:.5rem 0;}</style>
