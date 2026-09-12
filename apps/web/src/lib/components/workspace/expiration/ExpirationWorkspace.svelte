<script lang="ts">
 import { getContext, onDestroy, untrack, tick } from 'svelte';
 import type { Asset } from '$lib/domain/inventory';
 import { expirationWorkspaceContext, type ExpirationWorkspace, type ExpirationFilter, type ExpirationChoices } from '$lib/ports/expirationRepository';
 import { ExpirationList, type ExpirationListState } from '$lib/application/expirationList';
 import { workspaceRouteHref, type WorkspaceRouteState } from '$lib/application/workspaceRoute';
 import * as Button from '$lib/components/ui/button/index.js';
 import { Input } from '$lib/components/ui/input/index.js';
 import SegmentedControl from '../SegmentedControl.svelte';
 import ExpirationRefresh from '../ExpirationRefresh.svelte';
 import ExpirationRows from './ExpirationRows.svelte';
 import ExpirationFilters from './ExpirationFilters.svelte';
 let {tenantId,inventoryId,filter,onNavigate,onOpenAsset}: {tenantId:string;inventoryId:string;filter:ExpirationFilter;onNavigate:(route:Partial<WorkspaceRouteState>)=>void;onOpenAsset:(asset:Asset)=>void}=$props();
 const workspace = getContext<ExpirationWorkspace | undefined>(expirationWorkspaceContext);
 let listState = $state<ExpirationListState>({loading:true,appending:false,error:''});
 const list = workspace ? new ExpirationList(workspace.repository, next => { listState = next; }, workspace.cache, workspace.observer) : undefined;
 let query = $state(''); let filtersOpen = $state(false); let choices = $state<ExpirationChoices>(); let choicesLoading = $state(false); let choicesError = $state(''); let choicesController: AbortController | undefined;
 $effect(() => { const nextFilter = filter; const tenant = tenantId; const inventory = inventoryId; workspace?.revision?.(); query = nextFilter.query ?? ''; untrack(() => { void list?.load(tenant,inventory,nextFilter); }); });
 $effect(() => { tenantId; inventoryId; choicesController?.abort(); choices = undefined; choicesError = ''; filtersOpen = false; });
 let disposed=false;
 let viewKey=$derived(JSON.stringify([tenantId,inventoryId,filter]));
 $effect(()=>{if(listState.page&&!listState.loading){const key=viewKey;const position=workspace?.positions?.get(key);if(position){workspace?.positions?.delete(key);void tick().then(()=>{if(disposed)return;document.getElementById(`expiration-${position.assetId}`)?.focus({preventScroll:true});window.scrollTo({top:position.scrollY,behavior:'instant'});});}}});
 function openAsset(asset:Asset){workspace?.positions?.set(viewKey,{scrollY:window.scrollY,assetId:asset.id});onOpenAsset(asset);}
 onDestroy(() => {disposed=true;});
 onDestroy(() => { list?.dispose(); choicesController?.abort(); });
 function route(next:ExpirationFilter):Partial<WorkspaceRouteState> {return {mode:'expiration',tenantId,inventoryId,expirationFilter:next};}
 function navigate(next:ExpirationFilter) {onNavigate(route(next));}
 async function refresh() {await list?.load(tenantId,inventoryId,filter);return !list?.state.error;}
 async function loadChoices() {
  if (!workspace) return; choicesController?.abort(); const controller = new AbortController(); choicesController = controller; choicesLoading = true; choicesError = '';
  try { const next = await workspace.repository.choices(tenantId,inventoryId,controller.signal); if (!controller.signal.aborted) choices = next; }
  catch { if (!controller.signal.aborted) choicesError = 'Filter choices could not be loaded.'; }
  finally { if (!controller.signal.aborted) choicesLoading = false; }
 }
 let filtered = $derived(!!(filter.kind || filter.checkoutState || filter.query || filter.typeId || filter.locationId || filter.tagIds?.length || filter.fromDate || filter.throughDate));
 let options = $derived(([['soon','Expiring soon'],['expired','Expired'],['all','All dates']] as const).map(([value,label]) => ({value,label,description:listState.page ? String(listState.page.counts[value]) : undefined,href:workspaceRouteHref(route({...filter,mode:value}),tenantId,inventoryId)})));
</script>
<section class="workspace-main expiration-workspace" aria-label="Expiration">
 <header><div><h1>Expiration</h1><p>Review dates for active items across your inventory.</p></div><Button.Root variant="outline" onclick={() => { void refresh(); }} disabled={listState.loading}>Refresh</Button.Root></header>
 {#if workspace}
 <SegmentedControl label="Expiration status" value={filter.mode} {options} onSelect={mode => navigate({...filter,mode:mode as ExpirationFilter['mode']})} />
 <form class="search" onsubmit={event => { event.preventDefault();navigate({...filter,query:query.trim()}); }}><Input aria-label="Search item names and descriptions" placeholder="Search names and descriptions" bind:value={query} /><Button.Root type="submit" variant="outline">Search</Button.Root><Button.Root variant="outline" onclick={() => {filtersOpen=true;if(!choices)void loadChoices();}}>Filters{filtered?' · Active':''}</Button.Root></form>
 {#if filtered}<div class="filter-summary"><span>Filters applied{filter.fromDate ? ` · From ${filter.fromDate}` : ''}{filter.throughDate ? ` · Through ${filter.throughDate}` : ''}</span><Button.Root variant="ghost" onclick={() => navigate({mode:filter.mode})}>Clear filters</Button.Root></div>{/if}
 {#if listState.error}<div role="alert"><p>{listState.error}</p><Button.Root variant="outline" onclick={() => {void refresh();}}>Retry</Button.Root></div>{/if}
 {#if listState.loading && !listState.page}<p role="status" class="empty">Loading expiration dates…</p>{/if}
 {#if listState.page}
  <ExpirationRefresh assets={listState.page.items} timezone={listState.page.timezone} scope={`${tenantId}/${inventoryId}`} onRefresh={refresh} />
  <ExpirationRows items={listState.page.items} onOpenAsset={openAsset} />
  {#if !listState.page.items.length}<div class="empty"><h2>{filtered?'No matching items':filter.mode==='all'?'No expiration dates':filter.mode==='soon'?'None expiring soon':'No expired items'}</h2><p>{filtered?'Try changing or clearing your filters.':filter.mode==='all'?'Add expiration dates from an item’s details.':'You can review all recorded dates in All dates.'}</p></div>{/if}
  {#if listState.page.hasMore}<div class="more"><Button.Root variant="outline" disabled={listState.appending || listState.loading} onclick={() => {void list?.load(tenantId,inventoryId,filter,true);}}>{listState.appending?'Loading…':'Load more'}</Button.Root></div>{/if}
 {/if}
 <ExpirationFilters bind:open={filtersOpen} {filter} {choices} loading={choicesLoading} error={choicesError} onApply={navigate} onRetry={() => {void loadChoices();}} />
 {:else}<p role="alert">Expiration is unavailable in this session.</p>{/if}
</section>
<style>
 .expiration-workspace{max-width:64rem;margin:0 auto;min-height:65vh;padding:clamp(1rem,3vw,2.5rem);}header{display:flex;align-items:start;justify-content:space-between;gap:1rem;margin-bottom:1.5rem;}h1{font-size:var(--text-title-size);font-weight:650;}header p,.empty p,.filter-summary{color:var(--muted-foreground);} .search{display:flex;gap:.5rem;margin:1.25rem 0 .5rem;flex-wrap:wrap;}.search :global(input){flex:1;min-width:12rem;}.filter-summary{display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;}.empty{text-align:center;padding:4rem 1rem;} .empty h2{font-weight:600;margin-bottom:.5rem;}.more{display:flex;justify-content:center;padding:2rem;}[role=alert]{padding:1rem 0;}
</style>
