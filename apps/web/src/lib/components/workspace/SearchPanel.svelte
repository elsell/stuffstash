<script lang="ts">
  import { tick } from 'svelte';
  import Search from '@lucide/svelte/icons/search';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { workspaceRouteHref } from '$lib/application/workspaceRoute';
  import type { Asset, SearchLifecycleFilter, SearchMode, SearchResult } from '$lib/domain/inventory';
  import { assetKindLabel } from '$lib/domain/inventory';
  import AssetThumb from './AssetThumb.svelte';
  import SegmentedControl from './SegmentedControl.svelte';

  let {
    tenantId,
    inventoryId,
    query = $bindable(''),
    lifecycleState = $bindable<SearchLifecycleFilter>('active'),
    searchMode = $bindable<SearchMode>('fuzzy'),
    results,
    suggestions,
    submitted,
    error,
    busy,
    onSearch,
    onOpenAsset
  }: {
    tenantId: string;
    inventoryId: string;
    query: string;
    lifecycleState: SearchLifecycleFilter;
    searchMode: SearchMode;
    results: SearchResult[];
    suggestions: Asset[];
    submitted: boolean;
    error: string;
    busy: boolean;
    onSearch: () => void;
    onOpenAsset: (assetId: string) => void;
  } = $props();

  const lifecycleOptions: SearchLifecycleFilter[] = ['active', 'archived', 'all'];
  const modeOptions: SearchMode[] = ['fuzzy', 'exact'];
  let lifecycleControlOptions = $derived(
    lifecycleOptions.map((option) => ({
      value: option,
      label: option === 'active' ? 'Active' : option === 'archived' ? 'Archived' : 'All',
      href: searchFilterHref(option, searchMode)
    }))
  );
  let modeControlOptions = $derived(
    modeOptions.map((option) => ({
      value: option,
      label: option === 'fuzzy' ? 'Contains' : 'Exact',
      href: searchFilterHref(lifecycleState, option)
    }))
  );
  let searchFocused = $state(false);
  let activeSuggestionIndex = $state(-1);
  let searchRegion = $state<HTMLElement | null>(null);
  let visibleSuggestions = $derived(searchFocused && query.trim().length > 0 ? suggestions.slice(0, 6) : []);

  $effect(() => {
    if (activeSuggestionIndex >= visibleSuggestions.length) {
      activeSuggestionIndex = visibleSuggestions.length - 1;
    }
    if (visibleSuggestions.length === 0) {
      activeSuggestionIndex = -1;
    }
  });

  function assetHref(asset: Asset): string {
    return workspaceRouteHref({ mode: 'asset', tenantId: asset.tenantId, inventoryId: asset.inventoryId, assetId: asset.id }, asset.tenantId, asset.inventoryId);
  }

  function searchFilterHref(nextLifecycleState: SearchLifecycleFilter, nextSearchMode: SearchMode): string {
    return workspaceRouteHref(
      {
        mode: 'search',
        tenantId,
        inventoryId,
        searchQuery: query,
        searchLifecycleState: nextLifecycleState,
        searchMode: nextSearchMode
      },
      tenantId,
      inventoryId
    );
  }

  function openSuggestion(event: MouseEvent, asset: Asset): void {
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    query = asset.title;
    closeSuggestions();
    onOpenAsset(asset.id);
  }

  function openAsset(event: MouseEvent, asset: Asset): void {
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    onOpenAsset(asset.id);
  }

  function shouldHandleInApp(event: MouseEvent): boolean {
    return event.button === 0 && !event.metaKey && !event.ctrlKey && !event.shiftKey && !event.altKey;
  }

  function suggestionId(index: number): string {
    return `search-page-suggestion-${index}`;
  }

  function suggestionElement(index: number): HTMLElement | null {
    if (typeof document === 'undefined') {
      return null;
    }
    return document.getElementById(suggestionId(index));
  }

  async function focusSuggestion(index: number): Promise<void> {
    activeSuggestionIndex = index;
    await tick();
    suggestionElement(index)?.focus();
  }

  function focusSearchInput(): void {
    if (typeof document === 'undefined') {
      return;
    }
    document.getElementById('search-page-query')?.focus();
  }

  function handleSearchKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      if (visibleSuggestions.length > 0) {
        event.preventDefault();
      }
      closeSuggestions();
      return;
    }

    if (visibleSuggestions.length === 0) {
      return;
    }

    if (event.key === 'ArrowDown') {
      event.preventDefault();
      searchFocused = true;
      void focusSuggestion(0);
    }
  }

  function handleSuggestionKeydown(event: KeyboardEvent, index: number): void {
    if (event.key === 'Escape') {
      event.preventDefault();
      focusSearchInput();
      closeSuggestions();
      return;
    }

    if (event.key === 'ArrowDown') {
      event.preventDefault();
      void focusSuggestion((index + 1) % visibleSuggestions.length);
      return;
    }

    if (event.key === 'ArrowUp') {
      event.preventDefault();
      if (index === 0) {
        activeSuggestionIndex = -1;
        focusSearchInput();
      } else {
        void focusSuggestion(index - 1);
      }
    }
  }

  function closeSuggestions(): void {
    searchFocused = false;
    activeSuggestionIndex = -1;
  }

  function handleSearchFocusout(event: FocusEvent): void {
    const nextTarget = event.relatedTarget instanceof Node ? event.relatedTarget : null;
    if (nextTarget && searchRegion?.contains(nextTarget)) {
      return;
    }
    window.setTimeout(() => {
      const activeElement = document.activeElement;
      if (activeElement && searchRegion?.contains(activeElement)) {
        return;
      }
      closeSuggestions();
    }, 120);
  }
</script>

<section class="workspace-main" aria-labelledby="search-title">
  <div class="section-heading">
    <div>
      <h1 id="search-title">Search</h1>
      <p>Find authorized assets in this inventory.</p>
    </div>
  </div>

  <div bind:this={searchRegion} class="search-page-wrap" onfocusout={handleSearchFocusout}>
    <form class="search-panel" onsubmit={(event) => { event.preventDefault(); closeSuggestions(); onSearch(); }}>
      <Search aria-hidden="true" />
      <Input
        id="search-page-query"
        bind:value={query}
        placeholder="Asset, location, container, or field"
        aria-label="Search query"
        onfocus={() => { searchFocused = true; }}
        onkeydown={handleSearchKeydown}
      />
      <Button.Root disabled={busy || query.trim().length === 0}>Search</Button.Root>
    </form>
    {#if visibleSuggestions.length > 0}
      <ul id="search-page-suggestions" class="search-suggestions" aria-label="Search suggestions">
        {#each visibleSuggestions as suggestion, index}
          <li>
            <Button.Root
              id={suggestionId(index)}
              href={assetHref(suggestion)}
              variant="ghost"
              class="suggestion-row"
              data-active={activeSuggestionIndex === index}
              aria-label={`Open ${suggestion.title}`}
              onfocus={() => { activeSuggestionIndex = index; }}
              onkeydown={(event) => handleSuggestionKeydown(event, index)}
              onpointerenter={() => { activeSuggestionIndex = index; }}
              onclick={(event) => openSuggestion(event, suggestion)}
            >
              <AssetThumb asset={suggestion} size="sm" />
              <span>
                <strong>{suggestion.title}</strong>
                <small>{suggestion.customAssetTypeLabel ?? assetKindLabel(suggestion.kind)}</small>
              </span>
            </Button.Root>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
  <div class="search-filters" aria-label="Search filters">
    <SegmentedControl
      label="Result lifecycle"
      value={lifecycleState}
      options={lifecycleControlOptions}
      onSelect={(value) => { lifecycleState = value as SearchLifecycleFilter; onSearch(); }}
    />
    <SegmentedControl
      label="Search mode"
      value={searchMode}
      options={modeControlOptions}
      onSelect={(value) => { searchMode = value as SearchMode; onSearch(); }}
    />
  </div>

  {#if error}
    <div class="empty-state spacious" role="alert">
      <h2>Search failed</h2>
      <p>{error}</p>
    </div>
  {:else if busy}
    <div class="empty-state spacious">
      <h2>Searching</h2>
    </div>
  {:else if !submitted}
    <div class="empty-state spacious">
      <h2>Search this inventory</h2>
      <p>Use asset, location, container, custom field, or attachment terms.</p>
    </div>
  {:else if results.length === 0}
    <div class="empty-state spacious">
      <h2>No results</h2>
      <p>{lifecycleState === 'all' ? 'No authorized assets matched this query.' : `No authorized ${lifecycleState} assets matched this query.`}</p>
    </div>
  {:else}
    <div class="asset-list">
      {#each results as result}
        <Button.Root href={assetHref(result.asset)} variant="ghost" class="asset-row" onclick={(event) => openAsset(event, result.asset)}>
          <AssetThumb asset={result.asset} />
          <span class="asset-row-main">
            <strong>{result.asset.title}</strong>
            <small>{result.inventory.name} / {result.asset.lifecycleState}</small>
          </span>
          <span class="asset-row-meta">
            <small>{result.matches[0]?.field ?? 'match'}</small>
            <small>{result.matches[0]?.value ?? ''}</small>
          </span>
        </Button.Root>
      {/each}
    </div>
  {/if}
</section>
