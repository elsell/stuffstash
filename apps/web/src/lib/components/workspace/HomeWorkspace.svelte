<script lang="ts">
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import type { Asset, AssetLifecycleFilter, AssetViewModel, LocationSummary } from '$lib/domain/inventory';
  import { assetKindLabel } from '$lib/domain/inventory';
  import AssetThumb from './AssetThumb.svelte';

  let {
    lifecycleState,
    locations,
    recentAssets,
    archivedAssets,
    onOpenLocation,
    onOpenAsset,
    onOpenAdd,
    onSelectLifecycle
  }: {
    lifecycleState: AssetLifecycleFilter;
    locations: LocationSummary[];
    recentAssets: AssetViewModel[];
    archivedAssets: Asset[];
    onOpenLocation: (asset: Asset) => void;
    onOpenAsset: (asset: Asset) => void;
    onOpenAdd: () => void;
    onSelectLifecycle: (lifecycleState: AssetLifecycleFilter) => void;
  } = $props();
</script>

<section class="workspace-main" aria-labelledby="home-title">
  <div class="section-heading">
    <div>
      <h1 id="home-title">{lifecycleState === 'active' ? 'Home' : 'Archived assets'}</h1>
      <p>{lifecycleState === 'active' ? 'Recently added and the places where your things live.' : 'Assets removed from active browsing.'}</p>
    </div>
    <div class="heading-actions">
      <div class="filter-control" role="group" aria-label="Asset lifecycle">
        <Button.Root
          variant="ghost"
          aria-pressed={lifecycleState === 'active'}
          data-selected={lifecycleState === 'active'}
          onclick={() => onSelectLifecycle('active')}
        >
          Active
        </Button.Root>
        <Button.Root
          variant="ghost"
          aria-pressed={lifecycleState === 'archived'}
          data-selected={lifecycleState === 'archived'}
          onclick={() => onSelectLifecycle('archived')}
        >
          Archived
        </Button.Root>
      </div>
      {#if lifecycleState === 'active'}
        <Button.Root variant="outline" onclick={onOpenAdd}>Add location</Button.Root>
      {/if}
    </div>
  </div>

  {#if lifecycleState === 'active'}
    <section class="recent-section" aria-labelledby="recent-title">
      <div class="section-heading compact">
        <h2 id="recent-title">Recently added</h2>
      </div>
      {#if recentAssets.length === 0}
        <div class="empty-state">
          <p>No items or containers yet.</p>
        </div>
      {:else}
        <div class="recent-rail" aria-label="Recently added assets">
          {#each recentAssets as asset}
            <Button.Root variant="ghost" class="recent-card" onclick={() => onOpenAsset(asset)}>
              <AssetThumb {asset} size="lg" />
              <span>
                <strong>{asset.title}</strong>
                <small>{asset.customAssetTypeLabel ?? assetKindLabel(asset.kind)}</small>
                <small>{asset.containmentTrail}</small>
              </span>
            </Button.Root>
          {/each}
        </div>
      {/if}
    </section>
  {/if}

  {#if lifecycleState === 'archived'}
    {#if archivedAssets.length === 0}
      <div class="empty-state spacious">
        <h2>No archived assets</h2>
      </div>
    {:else}
      <div class="asset-list">
        {#each archivedAssets as asset}
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
  {:else if locations.length === 0}
    <div class="empty-state spacious">
      <h2>No locations yet</h2>
      <p>Add a location before adding things into it.</p>
      <Button.Root onclick={onOpenAdd}>Add first location</Button.Root>
    </div>
  {:else}
    <div class="section-heading compact">
      <h2>Locations</h2>
    </div>
    <div class="location-grid">
      {#each locations as summary}
        <Button.Root
          variant="ghost"
          class="location-tile"
          aria-label={`Open location ${summary.location.title}`}
          onclick={() => onOpenLocation(summary.location)}
        >
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

</section>
