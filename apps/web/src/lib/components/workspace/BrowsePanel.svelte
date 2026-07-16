<script lang="ts">
  import Search from '@lucide/svelte/icons/search';
  import SlidersHorizontal from '@lucide/svelte/icons/sliders-horizontal';
  import ArrowUpDown from '@lucide/svelte/icons/arrow-up-down';
  import { compareNaturalText } from '$lib/application/textCollation';
  import {
    browseEmptyPresentation,
    browseFilterOptions,
    browseFiltersAreDirty,
    browseFilterCount,
    browseSearchHref,
    buildAppliedBrowseFilters,
    filterBrowseAssets,
    mergeBrowseSearchState,
    searchAssetHref
  } from '$lib/application/workspaceSearch';
  import { workspaceRouteHref } from '$lib/application/workspaceRoute';
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import { buildPlaceBrowseSummaries } from '$lib/application/workspaceBrowsePresentation';
  import type {
    Asset, AssetTag, BrowseScope, BrowseSort, BrowseSurface, SearchCheckoutFilter,
    SearchLifecycleFilter, SearchMode, SearchResult
  } from '$lib/domain/inventory';
  import { assetKindLabel } from '$lib/domain/inventory';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import AssetTagChips from './AssetTagChips.svelte';
  import AssetThumb from './AssetThumb.svelte';
  import CheckoutBadge from './CheckoutBadge.svelte';
  import SearchSuggestions from './SearchSuggestions.svelte';
  import WorkspaceTaskSheet from './action-surface/WorkspaceTaskSheet.svelte';
  import KindIcon from './KindIcon.svelte';

  type BrowseState = {
    surface?: BrowseSurface;
    scope?: BrowseScope;
    lifecycleState?: SearchLifecycleFilter;
    checkoutState?: SearchCheckoutFilter;
    sort?: BrowseSort;
    selectedTagIds?: string[];
  };

  let {
    tenantId, inventoryId, inventoryName, assets, placementAssets, results, suggestions, assetTags, query = $bindable(''), submitted,
    error, busy, surface, scope, lifecycleState, searchMode: _searchMode, checkoutState, sort,
    selectedTagIds, canCreateAsset, inventoryEmpty, hasMore, loadingMore, errorPhase, onStateChange, onLoadMore, onRetry, onSearch, onOpenAsset, onOpenAdd
  }: {
    tenantId: string | null;
    inventoryId: string | null;
    inventoryName: string;
    assets: Asset[];
    placementAssets: Asset[];
    results: SearchResult[];
    suggestions: Asset[];
    assetTags: AssetTag[];
    query: string;
    submitted: boolean;
    error: string;
    busy: boolean;
    surface: BrowseSurface;
    scope: BrowseScope;
    lifecycleState: SearchLifecycleFilter;
    searchMode: SearchMode;
    checkoutState: SearchCheckoutFilter;
    sort: BrowseSort;
    selectedTagIds: string[];
    canCreateAsset: boolean;
    inventoryEmpty: boolean;
    hasMore: boolean;
    loadingMore: boolean;
    errorPhase: 'initial' | 'replacement' | 'append' | 'map' | null;
    onStateChange: (state: BrowseState) => void;
    onLoadMore: () => void;
    onRetry: () => void;
    onSearch: () => void;
    onOpenAsset: (asset: Asset) => void;
    onOpenAdd: (kind: 'item' | 'location', parentAssetId?: string | null, opener?: HTMLElement | null) => void;
  } = $props();

  const scopes = browseFilterOptions.scopes;
  const lifecycleOptions = browseFilterOptions.lifecycle;
  const availabilityOptions = browseFilterOptions.availability;

  let mapPathIds = $state<string[]>([]);
  let mapQuery = $state('');
  let mapActiveIndex = $state(-1);
  let mapColumnLimits = $state<Record<string, number>>({});
  let filterOpen = $state(false);
  let searchFocused = $state(false);
  let activeSuggestionIndex = $state(-1);
  let draftLifecycleState = $state<SearchLifecycleFilter>('active');
  let draftCheckoutState = $state<SearchCheckoutFilter>('any');
  let draftTagIds = $state<string[]>([]);
  let sourceAssets = $derived(submitted ? results.map((result) => result.asset) : assets);
  let filteredAssets = $derived(filterBrowseAssets(sourceAssets, { scope, lifecycleState, checkoutState, selectedTagIds }));
  let mapAssets = $derived(assets.filter((asset) => asset.lifecycleState === 'active'));
  let mapColumns = $derived([null, ...mapPathIds].map((parentId) => ({
    parentId,
    title: parentId ? mapAssets.find((asset) => asset.id === parentId)?.title ?? 'Contents' : 'Inventory root',
    assets: mapAssets.filter((asset) => asset.parentAssetId === parentId).sort((a, b) => compareNaturalText(a.title, b.title))
  })));
  let mapMatches = $derived(mapQuery.trim() ? mapAssets.filter((asset) => asset.kind !== 'item' && asset.title.toLocaleLowerCase().includes(mapQuery.trim().toLocaleLowerCase())).slice(0, 8) : []);
  let selectedMapAsset = $derived(mapPathIds.length ? mapAssets.find((asset) => asset.id === mapPathIds.at(-1)) ?? null : null);
  let filterCount = $derived(browseFilterCount(lifecycleState, checkoutState, selectedTagIds));
  let filterDirty = $derived(browseFiltersAreDirty(
    draftLifecycleState, lifecycleState, draftCheckoutState, checkoutState, draftTagIds, selectedTagIds
  ));
  let appliedFilters = $derived(buildAppliedBrowseFilters(lifecycleState, checkoutState, selectedTagIds, assetTags));
  let placeSummaries = $derived(new Map(buildPlaceBrowseSummaries(filteredAssets.filter((asset) => asset.kind === 'location'), placementAssets).map((summary) => [summary.asset.id, summary])));
  let visibleSuggestions = $derived(searchFocused && query.trim() ? suggestions.slice(0, 6) : []);
  let emptyPresentation = $derived(browseEmptyPresentation(
    inventoryEmpty, query, scope, lifecycleState, checkoutState, selectedTagIds, canCreateAsset
  ));

  function browseHref(next: BrowseState = {}): string {
    return browseSearchHref(tenantId, inventoryId, mergeBrowseSearchState({
      query,
      lifecycleState,
      mode: _searchMode,
      checkoutState,
      surface,
      scope,
      sort,
      selectedTagIds
    }, next));
  }

  function changeBrowseState(event: MouseEvent, next: BrowseState): void {
    if (!shouldHandleWorkspaceLinkClick(event)) return;
    event.preventDefault();
    onStateChange(next);
  }

  function openAdd(event: MouseEvent, kind: 'item' | 'location'): void {
    if (!shouldHandleWorkspaceLinkClick(event)) return;
    event.preventDefault();
    onOpenAdd(kind, null, event.currentTarget as HTMLElement);
  }

  function open(event: MouseEvent, asset: Asset): void {
    if (!shouldHandleWorkspaceLinkClick(event)) return;
    event.preventDefault();
    onOpenAsset(asset);
  }

  function toggleTag(tagId: string): void {
    onStateChange({ selectedTagIds: selectedTagIds.includes(tagId) ? selectedTagIds.filter((id) => id !== tagId) : [...selectedTagIds, tagId] });
  }

  function removeFilter(key: string): void {
    if (key === 'lifecycle') onStateChange({ lifecycleState: 'active' });
    else if (key === 'availability') onStateChange({ checkoutState: 'any' });
    else if (key.startsWith('tag:')) toggleTag(key.slice(4));
  }

  function tabKeydown<T extends string>(event: KeyboardEvent, values: readonly T[], current: T, select: (value: T) => void): void {
    const currentIndex = values.indexOf(current);
    let nextIndex = currentIndex;
    if (event.key === 'ArrowRight') nextIndex = (currentIndex + 1) % values.length;
    else if (event.key === 'ArrowLeft') nextIndex = (currentIndex - 1 + values.length) % values.length;
    else if (event.key === 'Home') nextIndex = 0;
    else if (event.key === 'End') nextIndex = values.length - 1;
    else return;
    event.preventDefault();
    const tablist = (event.currentTarget as HTMLElement).parentElement;
    select(values[nextIndex]!);
    queueMicrotask(() => tablist?.querySelector<HTMLElement>('[tabindex="0"]')?.focus());
  }

  function setFiltersOpen(nextOpen: boolean): void {
    if (!nextOpen) {
      filterOpen = false;
      return;
    }
    draftLifecycleState = lifecycleState;
    draftCheckoutState = checkoutState;
    draftTagIds = [...selectedTagIds];
    filterOpen = true;
  }

  function closeFilters(): void {
    filterOpen = false;
  }

  function toggleDraftTag(tagId: string): void {
    draftTagIds = draftTagIds.includes(tagId) ? draftTagIds.filter((id) => id !== tagId) : [...draftTagIds, tagId];
  }

  function applyFilters(): void {
    onStateChange({ lifecycleState: draftLifecycleState, checkoutState: draftCheckoutState, selectedTagIds: draftTagIds });
    closeFilters();
  }

  function openMapNode(asset: Asset, columnIndex: number): void {
    if (asset.kind === 'item') {
      onOpenAsset(asset);
      return;
    }
    mapPathIds = [...mapPathIds.slice(0, columnIndex), asset.id];
  }

  function revealMapMatch(asset: Asset): void {
    const ancestors: string[] = [];
    let parentId = asset.parentAssetId;
    while (parentId) {
      const parent = mapAssets.find((candidate) => candidate.id === parentId);
      if (!parent) break;
      ancestors.unshift(parent.id);
      parentId = parent.parentAssetId;
    }
    mapPathIds = asset.kind === 'item' ? ancestors : [...ancestors, asset.id];
    mapQuery = '';
    mapActiveIndex = -1;
    queueMicrotask(() => document.querySelector<HTMLInputElement>('input[aria-label="Jump to a place or container"]')?.focus());
  }

  function mapSearchKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      mapQuery = '';
      mapActiveIndex = -1;
      return;
    }
    if (event.key !== 'ArrowDown' || mapMatches.length === 0) return;
    event.preventDefault();
    mapActiveIndex = 0;
    queueMicrotask(() => document.getElementById('map-jump-option-0')?.focus());
  }

  function mapMatchKeydown(event: KeyboardEvent, index: number): void {
    if (event.key === 'Escape') {
      event.preventDefault();
      mapActiveIndex = -1;
      document.querySelector<HTMLInputElement>('input[aria-label="Jump to a place or container"]')?.focus();
      return;
    }
    if (event.key !== 'ArrowDown' && event.key !== 'ArrowUp') return;
    event.preventDefault();
    const next = event.key === 'ArrowDown'
      ? (index + 1) % mapMatches.length
      : (index - 1 + mapMatches.length) % mapMatches.length;
    mapActiveIndex = next;
    document.getElementById(`map-jump-option-${next}`)?.focus();
  }

  function mapColumnKey(parentId: string | null): string {
    return parentId ?? 'root';
  }

  function mapColumnLimit(parentId: string | null): number {
    return mapColumnLimits[mapColumnKey(parentId)] ?? 100;
  }

  function showMoreMapNodes(parentId: string | null): void {
    const key = mapColumnKey(parentId);
    mapColumnLimits = { ...mapColumnLimits, [key]: mapColumnLimit(parentId) + 100 };
  }

  function searchKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      searchFocused = false;
      activeSuggestionIndex = -1;
      return;
    }
    if (visibleSuggestions.length === 0) return;
    if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
      event.preventDefault();
      activeSuggestionIndex = event.key === 'ArrowDown'
        ? (activeSuggestionIndex + 1) % visibleSuggestions.length
        : (activeSuggestionIndex <= 0 ? visibleSuggestions.length - 1 : activeSuggestionIndex - 1);
      return;
    }
    if (event.key === 'Enter' && activeSuggestionIndex >= 0) {
      event.preventDefault();
      const selectedSuggestion = visibleSuggestions[activeSuggestionIndex];
      searchFocused = false;
      if (selectedSuggestion) onOpenAsset(selectedSuggestion);
    }
  }

  function closeSearchOnBlur(event: FocusEvent): void {
    const region = event.currentTarget as HTMLElement;
    window.setTimeout(() => {
      if (!region.contains(document.activeElement)) searchFocused = false;
    }, 0);
  }

  function clearSearch(): void {
    query = '';
    searchFocused = false;
    activeSuggestionIndex = -1;
    onSearch();
  }
</script>

<section class="workspace-main browse-workspace" aria-labelledby="browse-title">
  <header class="browse-heading">
    <div><h1 id="browse-title">Browse</h1><p>{inventoryName}</p></div>
    <div class="browse-surface-tabs" role="tablist" aria-label="Browse surface">
      {#each [{ value: 'list', label: 'List' }, { value: 'map', label: 'Map' }] as option}
        <Button.Root href={browseHref({ surface: option.value as BrowseSurface })} id={`browse-surface-${option.value}-tab`} role="tab" tabindex={surface === option.value ? 0 : -1} aria-controls={`browse-${option.value}-panel`} aria-selected={surface === option.value} variant={surface === option.value ? 'secondary' : 'ghost'} onkeydown={(event) => tabKeydown(event, ['list', 'map'], surface, (value) => onStateChange({ surface: value }))} onclick={(event) => changeBrowseState(event, { surface: option.value as BrowseSurface })}>{option.label}</Button.Root>
      {/each}
    </div>
  </header>

  {#if surface === 'list'}
    <div id="browse-list-panel" role="tabpanel" aria-labelledby="browse-surface-list-tab" class="browse-panel-content">
    <div class="browse-search-wrap" onfocusout={closeSearchOnBlur}>
      <form class="browse-search" onsubmit={(event) => { event.preventDefault(); searchFocused = false; onSearch(); }}>
        <Search aria-hidden="true" />
        <Input bind:value={query} aria-label="Search Browse" role="combobox" aria-autocomplete="list" aria-expanded={visibleSuggestions.length > 0} aria-controls={visibleSuggestions.length > 0 ? 'browse-suggestions' : undefined} aria-activedescendant={activeSuggestionIndex >= 0 ? `browse-suggestion-${activeSuggestionIndex}` : undefined} placeholder="Search this inventory" onfocus={() => { searchFocused = true; }} oninput={() => { activeSuggestionIndex = -1; }} onkeydown={searchKeydown} />
        <Button.Root type="submit" disabled={busy || query.trim().length === 0}>Search</Button.Root>
      </form>
      <SearchSuggestions id="browse-suggestions" idPrefix="browse-suggestion" suggestions={visibleSuggestions} activeIndex={activeSuggestionIndex} {query} showEmpty={searchFocused && query.trim().length > 0 && visibleSuggestions.length === 0} assetHref={searchAssetHref} onFocusIndex={(index) => { activeSuggestionIndex = index; }} onSuggestionKeydown={() => {}} onOpen={(event, asset) => { searchFocused = false; open(event, asset); }} />
    </div>

    <div class="browse-scope-tabs" role="tablist" aria-label="Browse scope">
      {#each scopes as option}
        <Button.Root href={browseHref({ scope: option.value })} id={`browse-scope-${option.value}-tab`} role="tab" tabindex={scope === option.value ? 0 : -1} aria-controls="browse-results" aria-selected={scope === option.value} variant={scope === option.value ? 'secondary' : 'ghost'} onkeydown={(event) => tabKeydown(event, scopes.map((item) => item.value), scope, (value) => onStateChange({ scope: value }))} onclick={(event) => changeBrowseState(event, { scope: option.value })}>{option.label}</Button.Root>
      {/each}
    </div>

    <div class="browse-tools">
      <Button.Root variant="outline" class="browse-filter-trigger" onclick={() => setFiltersOpen(true)}><SlidersHorizontal aria-hidden="true" /> Filters{filterCount ? ` (${filterCount})` : ''}</Button.Root>
      {#if filterOpen}
        <WorkspaceTaskSheet open title="Filters" description="Limit the assets shown in Browse." dismissible={!filterDirty} onOpenChange={(open) => { if (!open) closeFilters(); }}>
          <div class="browse-filter-popover">
          <fieldset><legend>Status</legend>{#each lifecycleOptions as option}<Button.Root autofocus={option.value === lifecycleState} variant={draftLifecycleState === option.value ? 'secondary' : 'ghost'} aria-pressed={draftLifecycleState === option.value} onclick={() => { draftLifecycleState = option.value; }}>{option.label}</Button.Root>{/each}</fieldset>
          <fieldset><legend>Availability</legend>{#each availabilityOptions as option}<Button.Root variant={draftCheckoutState === option.value ? 'secondary' : 'ghost'} aria-pressed={draftCheckoutState === option.value} onclick={() => { draftCheckoutState = option.value; }}>{option.label}</Button.Root>{/each}</fieldset>
          {#if assetTags.length}<fieldset class="browse-filter-tags"><legend>Tags</legend>{#each [...assetTags].sort((a,b) => compareNaturalText(a.displayName,b.displayName)) as tag}<Button.Root class="browse-filter-chip" variant={draftTagIds.includes(tag.id) ? 'secondary' : 'outline'} aria-pressed={draftTagIds.includes(tag.id)} onclick={() => toggleDraftTag(tag.id)}>{tag.displayName}</Button.Root>{/each}</fieldset>{/if}
          </div>
          {#snippet footer()}<Button.Root variant="ghost" onclick={() => { draftLifecycleState = 'active'; draftCheckoutState = 'any'; draftTagIds = []; }}>Reset</Button.Root><Button.Root variant="outline" onclick={closeFilters}>Cancel</Button.Root><Button.Root onclick={applyFilters}>Apply filters</Button.Root>{/snippet}
        </WorkspaceTaskSheet>
      {/if}
      <div class="browse-sort" role="group" aria-label="Sort Browse"><ArrowUpDown aria-hidden="true" /> <span>Sort</span><Button.Root href={query.trim() ? undefined : browseHref({ sort: 'updated_desc' })} variant={sort === 'updated_desc' ? 'secondary' : 'ghost'} aria-pressed={sort === 'updated_desc'} disabled={query.trim().length > 0} onclick={(event) => changeBrowseState(event, { sort: 'updated_desc' })}>Recently changed</Button.Root><Button.Root href={query.trim() ? undefined : browseHref({ sort: 'id_asc' })} variant={sort === 'id_asc' ? 'secondary' : 'ghost'} aria-pressed={sort === 'id_asc'} disabled={query.trim().length > 0} onclick={(event) => changeBrowseState(event, { sort: 'id_asc' })}>Default order</Button.Root></div>
      {#if query.trim()}<small id="browse-sort-note" class="browse-sort-note">Results are ordered by search relevance.</small>{/if}
      <span class="browse-shown-count">{filteredAssets.length} shown</span>
      {#if busy && filteredAssets.length > 0}<small class="browse-updating" role="status">Updating…</small>{/if}
    </div>

    {#if appliedFilters.length}
      <div class="browse-applied-filters" aria-label="Applied filters">{#each appliedFilters as filter}<Button.Root variant="secondary" aria-label={`Remove ${filter.label}`} onclick={() => removeFilter(filter.key)}>{filter.label}<span aria-hidden="true">×</span></Button.Root>{/each}{#if filterCount > 1}<Button.Root variant="ghost" onclick={() => onStateChange({ lifecycleState: 'active', checkoutState: 'any', selectedTagIds: [] })}>Clear all</Button.Root>{/if}</div>
    {/if}
    {#if error && filteredAssets.length === 0}<div class="empty-state spacious" role="alert"><h2>Browse failed</h2><p>{error}</p><Button.Root onclick={onRetry}>Try again</Button.Root></div>
    {:else if busy && filteredAssets.length === 0}<div class="empty-state spacious" role="status"><h2>Loading inventory…</h2></div>
    {:else if filteredAssets.length === 0}
      <div class="empty-state spacious">
        <h2>{emptyPresentation.title}</h2>
        <p>{emptyPresentation.description}</p>
        {#if emptyPresentation.kind === 'inventory'}
          {#if emptyPresentation.showCreateActions}
            <div class="empty-state-actions">
              <Button.Root href={workspaceRouteHref({ action: 'add', addKind: 'item' }, tenantId, inventoryId)} onclick={(event) => openAdd(event, 'item')}>Add item</Button.Root>
              <Button.Root href={workspaceRouteHref({ action: 'add', addKind: 'location' }, tenantId, inventoryId)} variant="outline" onclick={(event) => openAdd(event, 'location')}>Add location</Button.Root>
            </div>
          {/if}
        {:else if emptyPresentation.showClearSearch}
          <Button.Root variant="outline" onclick={clearSearch}>Clear search</Button.Root>
        {/if}
      </div>
    {:else}
      <div id="browse-results" role="tabpanel" aria-labelledby={`browse-scope-${scope}-tab`} class:place-results={scope === 'places'} class="browse-result-grid">
        {#each filteredAssets as asset}
          <article class="browse-card">
            <Button.Root href={searchAssetHref(asset)} variant="ghost" class="browse-card-open" onclick={(event) => open(event, asset)}>
              <AssetThumb {asset} size="lg" />
              <span class="browse-card-copy"><strong>{asset.title}</strong><small>{asset.kind === 'location' ? `${placeSummaries.get(asset.id)?.containedCount ?? 0} contained` : assetKindLabel(asset.kind)}</small>{#if asset.kind === 'location' && placeSummaries.get(asset.id)?.recentContainedNames.length}<small>{placeSummaries.get(asset.id)?.recentContainedNames.join(' · ')}</small>{/if}{#if asset.currentCheckout}<CheckoutBadge checkout={asset.currentCheckout} compact />{/if}</span>
            </Button.Root>
            <AssetTagChips tags={asset.tags ?? []} compact overflowLimit={2} />
          </article>
        {/each}
      </div>
      {#if hasMore}<div class="browse-load-more"><Button.Root disabled={loadingMore} onclick={onLoadMore}>{loadingMore ? 'Loading…' : 'Load more'}</Button.Root></div>{/if}
      {#if error && filteredAssets.length > 0}<div class="browse-inline-error" role="alert"><span>{error}</span><Button.Root variant="outline" onclick={onRetry}>Try again</Button.Root></div>{/if}
    {/if}
    </div>
  {:else}
    <div id="browse-map-panel" role="tabpanel" aria-labelledby="browse-surface-map-tab" class="containment-map">
      {#if busy && mapAssets.length === 0}
        <div class="empty-state spacious" role="status"><h2>Loading map…</h2></div>
      {:else}
      {#if busy}<small class="browse-updating" role="status">Updating map…</small>{/if}
      {#if error}<div class="browse-inline-error" role="alert"><span>{error}</span><Button.Root variant="outline" onclick={onRetry}>Try map again</Button.Root></div>{/if}
      <div class="containment-map-toolbar">
        <nav class="containment-breadcrumb" aria-label="Containment path"><Button.Root variant="ghost" onclick={() => { mapPathIds = []; }}>Inventory root</Button.Root>{#each mapPathIds as assetId, index}<span>/</span><Button.Root variant="ghost" onclick={() => { mapPathIds = mapPathIds.slice(0, index + 1); }}>{mapAssets.find((asset) => asset.id === assetId)?.title}</Button.Root>{/each}</nav>
        <div class="containment-jump"><Search aria-hidden="true" /><Input bind:value={mapQuery} aria-label="Jump to a place or container" role="combobox" aria-autocomplete="list" aria-expanded={mapMatches.length > 0} aria-controls={mapMatches.length > 0 ? 'map-jump-results' : undefined} aria-activedescendant={mapActiveIndex >= 0 ? `map-jump-option-${mapActiveIndex}` : undefined} placeholder="Jump to…" onkeydown={mapSearchKeydown} />{#if mapMatches.length}<div id="map-jump-results" class="containment-jump-results" role="listbox" aria-label="Map jump results">{#each mapMatches as asset, index}<Button.Root id={`map-jump-option-${index}`} role="option" aria-selected={mapActiveIndex === index} variant="ghost" onfocus={() => { mapActiveIndex = index; }} onkeydown={(event) => mapMatchKeydown(event, index)} onclick={() => revealMapMatch(asset)}>{asset.title}<small>{assetKindLabel(asset.kind)}</small></Button.Root>{/each}</div>{/if}</div>
      </div>
      {#if selectedMapAsset}<aside class="containment-inspector" aria-label="Selected map asset"><AssetThumb asset={selectedMapAsset} size="md" /><div><strong>{selectedMapAsset.title}</strong><small>{assetKindLabel(selectedMapAsset.kind)} · {selectedMapAsset.lifecycleState}</small></div><Button.Root href={searchAssetHref(selectedMapAsset)} onclick={(event) => open(event, selectedMapAsset)}>Open</Button.Root></aside>{/if}
      <div class:root-only={mapColumns.length === 1} class="containment-columns">{#each mapColumns as column, columnIndex}<section aria-label={column.title}><h2>{column.title}</h2>{#each column.assets.slice(0, mapColumnLimit(column.parentId)) as asset}<Button.Root variant={mapPathIds[columnIndex] === asset.id ? 'secondary' : 'ghost'} class="containment-node" onclick={() => openMapNode(asset, columnIndex)}><span class="containment-node-kind" aria-hidden="true"><KindIcon kind={asset.kind} /></span><span><strong>{asset.title}</strong><small>{assetKindLabel(asset.kind)}</small></span></Button.Root>{/each}{#if column.assets.length === 0}<p class="muted">Nothing is contained here.</p>{:else if column.assets.length > mapColumnLimit(column.parentId)}<Button.Root variant="outline" class="containment-show-more" onclick={() => showMoreMapNodes(column.parentId)}>Show next {Math.min(100, column.assets.length - mapColumnLimit(column.parentId))}</Button.Root><p class="muted" role="status">{mapColumnLimit(column.parentId)} of {column.assets.length} shown</p>{/if}</section>{/each}</div>
      {/if}
    </div>
  {/if}
</section>
