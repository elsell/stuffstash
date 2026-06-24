<script lang="ts">
  import Search from '@lucide/svelte/icons/search';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import type { SearchResult } from '$lib/domain/inventory';
  import AssetThumb from './AssetThumb.svelte';

  let {
    query = $bindable(''),
    results,
    busy,
    onSearch,
    onOpenAsset
  }: {
    query: string;
    results: SearchResult[];
    busy: boolean;
    onSearch: () => void;
    onOpenAsset: (assetId: string) => void;
  } = $props();
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

  {#if results.length === 0}
    <div class="empty-state spacious">
      <h2>No results</h2>
      <p>Search by item, container, location, or description.</p>
    </div>
  {:else}
    <div class="asset-list">
      {#each results as result}
        <Button.Root variant="ghost" class="asset-row" onclick={() => onOpenAsset(result.asset.id)}>
          <AssetThumb asset={result.asset} />
          <span class="asset-row-main">
            <strong>{result.asset.title}</strong>
            <small>{result.inventory.name}</small>
          </span>
          <span class="asset-row-meta">
            <small>{result.matches[0]?.field ?? 'match'}</small>
          </span>
        </Button.Root>
      {/each}
    </div>
  {/if}
</section>
