<script lang="ts">
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import type { Asset, AssetLifecycleFilter, LocationSummary } from '$lib/domain/inventory';
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
    recentAssets: Asset[];
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
      <h1 id="home-title">{lifecycleState === 'active' ? 'Locations' : 'Archived assets'}</h1>
      <p>{lifecycleState === 'active' ? 'Browse the places where your things live.' : 'Assets removed from active browsing.'}</p>
    </div>
    <div class="heading-actions">
      <div class="lifecycle-switcher" role="group" aria-label="Asset lifecycle">
        <Button.Root
          variant={lifecycleState === 'active' ? 'secondary' : 'outline'}
          aria-pressed={lifecycleState === 'active'}
          onclick={() => onSelectLifecycle('active')}
        >
          Active
        </Button.Root>
        <Button.Root
          variant={lifecycleState === 'archived' ? 'secondary' : 'outline'}
          aria-pressed={lifecycleState === 'archived'}
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
  {/if}
</section>
