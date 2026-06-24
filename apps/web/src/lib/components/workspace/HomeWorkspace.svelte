<script lang="ts">
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import type { Asset, LocationSummary } from '$lib/domain/inventory';
  import { assetKindLabel } from '$lib/domain/inventory';
  import AssetThumb from './AssetThumb.svelte';

  let {
    locations,
    recentAssets,
    onOpenLocation,
    onOpenAsset,
    onOpenAdd
  }: {
    locations: LocationSummary[];
    recentAssets: Asset[];
    onOpenLocation: (asset: Asset) => void;
    onOpenAsset: (asset: Asset) => void;
    onOpenAdd: () => void;
  } = $props();
</script>

<section class="workspace-main" aria-labelledby="home-title">
  <div class="section-heading">
    <div>
      <h1 id="home-title">Locations</h1>
      <p>Browse the places where your things live.</p>
    </div>
    <Button.Root variant="outline" onclick={onOpenAdd}>Add location</Button.Root>
  </div>

  {#if locations.length === 0}
    <div class="empty-state spacious">
      <h2>No locations yet</h2>
      <p>Add a location before adding things into it.</p>
      <Button.Root onclick={onOpenAdd}>Add first location</Button.Root>
    </div>
  {:else}
    <div class="location-grid">
      {#each locations as summary}
        <Button.Root variant="ghost" class="location-tile" onclick={() => onOpenLocation(summary.location)}>
          <AssetThumb asset={summary.location} size="lg" />
          <span>
            <strong>{summary.location.title}</strong>
            <small>{summary.assetCount} visible assets</small>
          </span>
          <ChevronRight aria-hidden="true" />
        </Button.Root>
      {/each}
    </div>
  {/if}

  <section class="recent-section" aria-labelledby="recent-title">
    <div class="section-heading compact">
      <h2 id="recent-title">Recently added</h2>
    </div>
    {#if recentAssets.length === 0}
      <div class="empty-state">
        <p>No items or containers yet.</p>
      </div>
    {:else}
      <div class="asset-list">
        {#each recentAssets as asset}
          <Button.Root variant="ghost" class="asset-row" onclick={() => onOpenAsset(asset)}>
            <AssetThumb {asset} />
            <span class="asset-row-main">
              <strong>{asset.title}</strong>
              <small>{asset.description || assetKindLabel(asset.kind)}</small>
            </span>
            <Badge variant="outline">{assetKindLabel(asset.kind)}</Badge>
          </Button.Root>
        {/each}
      </div>
    {/if}
  </section>
</section>
