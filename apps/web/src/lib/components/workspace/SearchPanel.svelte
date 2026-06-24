<script lang="ts">
  import Search from '@lucide/svelte/icons/search';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import type { SearchLifecycleFilter, SearchMode, SearchResult } from '$lib/domain/inventory';
  import AssetThumb from './AssetThumb.svelte';

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
    <div class="lifecycle-switcher" role="group" aria-label="Result lifecycle">
      {#each lifecycleOptions as option}
        <Button.Root
          variant={lifecycleState === option ? 'secondary' : 'outline'}
          aria-pressed={lifecycleState === option}
          onclick={() => { lifecycleState = option; if (query.trim()) onSearch(); }}
        >
          {option === 'active' ? 'Active' : option === 'archived' ? 'Archived' : 'All'}
        </Button.Root>
      {/each}
    </div>
    <div class="lifecycle-switcher" role="group" aria-label="Search mode">
      {#each modeOptions as option}
        <Button.Root
          variant={searchMode === option ? 'secondary' : 'outline'}
          aria-pressed={searchMode === option}
          onclick={() => { searchMode = option; if (query.trim()) onSearch(); }}
        >
          {option === 'fuzzy' ? 'Contains' : 'Exact'}
        </Button.Root>
      {/each}
    </div>
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
