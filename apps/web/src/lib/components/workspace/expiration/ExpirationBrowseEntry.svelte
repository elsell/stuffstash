<script lang="ts">
 import { getContext } from 'svelte';
 import { expirationWorkspaceContext, type ExpirationWorkspace, type ExpirationMode, type ExpirationFilter } from '$lib/ports/expirationRepository';
 import { workspaceRouteHref, type WorkspaceRouteState } from '$lib/application/workspaceRoute';
 import type { BrowseScope, SearchCheckoutFilter } from '$lib/domain/inventory';
 import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
 import * as Button from '$lib/components/ui/button/index.js';
 let {tenantId,inventoryId,query,tagIds,scope,checkoutState,onNavigate}:{tenantId:string;inventoryId:string;query:string;tagIds:string[];scope:BrowseScope;checkoutState:SearchCheckoutFilter;onNavigate?:(route:Partial<WorkspaceRouteState>)=>void}=$props();
 const workspace=getContext<ExpirationWorkspace|undefined>(expirationWorkspaceContext);
 function route(mode:ExpirationMode):Partial<WorkspaceRouteState>{const filter:ExpirationFilter={mode,query,tagIds,kind:scope==='items'?'item':scope==='containers'?'container':scope==='places'?'location':undefined,checkoutState};return{mode:'expiration',tenantId,inventoryId,expirationFilter:filter};}
 function open(event:MouseEvent,mode:ExpirationMode){if(onNavigate&&shouldHandleWorkspaceLinkClick(event)){event.preventDefault();onNavigate(route(mode));}}
</script>
{#if workspace}<nav aria-label="Browse expiration dates" class="expiry-refinement"><span>Expiration · Active items</span>{#each [['soon','Expiring soon'],['expired','Expired'],['all','All dates']] as [mode,label]}<Button.Root variant="ghost" href={workspaceRouteHref(route(mode as ExpirationMode),tenantId,inventoryId)} onclick={event=>open(event,mode as ExpirationMode)}>{label}</Button.Root>{/each}</nav>{/if}
<style>.expiry-refinement{display:flex;align-items:center;gap:.5rem;flex-wrap:wrap;padding-bottom:1rem;}.expiry-refinement>span{color:var(--muted-foreground);font-size:var(--text-metadata-size);margin-right:.5rem;}</style>
