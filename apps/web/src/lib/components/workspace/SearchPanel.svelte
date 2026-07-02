<script lang="ts">
  import Search from '@lucide/svelte/icons/search';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import type { SearchLifecycleFilter, SearchMode, SearchResult } from '$lib/domain/inventory';
  import AssetThumb from './AssetThumb.svelte';
  import SegmentedControl from './SegmentedControl.svelte';

  let {
    query = $bindable(''),
    lifecycleState = $bindable<SearchLifecycleFilter>('active'),
    searchMode = $bindable<SearchMode>('fuzzy'),
    results,
    submitted,
    error,
    busy,
    onSearch,
    onOpenAsset
  }: {
    query: string;
    lifecycleState: SearchLifecycleFilter;
    searchMode: SearchMode;
    results: SearchResult[];
    submitted: boolean;
    error: string;
    busy: boolean;
    onSearch: () => void;
    onOpenAsset: (assetId: string) => void;
  } = $props();

  const lifecycleOptions: SearchLifecycleFilter[] = ['active', 'archived', 'all'];
  const modeOptions: SearchMode[] = ['fuzzy', 'exact'];
  const lifecycleControlOptions = lifecycleOptions.map((option) => ({
    value: option,
    label: option === 'active' ? 'Active' : option === 'archived' ? 'Archived' : 'All'
  }));
  const modeControlOptions = modeOptions.map((option) => ({
    value: option,
    label: option === 'fuzzy' ? 'Contains' : 'Exact'
  }));
</script>

<section class="workspace-main" aria-labelledby="search-title">
  <div class="section-heading">
    <div>
      <h1 id="search-title">Search</h1>
      <p>Find authorized assets in this inventory.</p>
    </div>
  </div>

  <form class="search-panel" onsubmit={(event) => { event.preventDefault(); onSearch(); }}>
    <Search aria-hidden="true" />
    <Input bind:value={query} placeholder="Asset, location, container, or field" aria-label="Search query" />
    <Button.Root disabled={busy || query.trim().length === 0}>Search</Button.Root>
  </form>
  <div class="search-filters" aria-label="Search filters">
    <SegmentedControl
      label="Result lifecycle"
      value={lifecycleState}
      options={lifecycleControlOptions}
      onSelect={(value) => { lifecycleState = value as SearchLifecycleFilter; if (query.trim()) onSearch(); }}
    />
    <SegmentedControl
      label="Search mode"
      value={searchMode}
      options={modeControlOptions}
      onSelect={(value) => { searchMode = value as SearchMode; if (query.trim()) onSearch(); }}
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
        <Button.Root variant="ghost" class="asset-row" onclick={() => onOpenAsset(result.asset.id)}>
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
