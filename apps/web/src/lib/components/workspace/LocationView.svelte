<script lang="ts">
  import ArrowLeft from '@lucide/svelte/icons/arrow-left';
  import Pencil from '@lucide/svelte/icons/pencil';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import type { Asset, AssetViewModel } from '$lib/domain/inventory';
  import { assetKindLabel } from '$lib/domain/inventory';
  import AssetThumb from './AssetThumb.svelte';

  let {
    location,
    assets,
    canEdit,
    onBack,
    onOpenLocation,
    onEditLocation,
    onOpenAsset
  }: {
    location: Asset;
    assets: AssetViewModel[];
    canEdit: boolean;
    onBack: () => void;
    onOpenLocation: (asset: Asset) => void;
    onEditLocation: (asset: Asset) => void;
    onOpenAsset: (asset: Asset) => void;
  } = $props();
</script>

<section class="workspace-main" aria-labelledby="location-title">
  <Button.Root variant="ghost" class="back-button" onclick={onBack}><ArrowLeft /> Back</Button.Root>
  <div class="location-hero">
    <AssetThumb asset={location} size="lg" />
    <div>
      <h1 id="location-title">{location.title}</h1>
      <p>{location.description}</p>
      <Badge variant="secondary">{assets.length} visible assets</Badge>
    </div>
    {#if canEdit}
      <Button.Root variant="outline" onclick={() => onEditLocation(location)}><Pencil /> Edit location</Button.Root>
    {/if}
  </div>

  {#if assets.length === 0}
    <div class="empty-state spacious">
      <h2>No stuff here yet</h2>
      <p>Add an item or move existing stuff into this location.</p>
    </div>
  {:else}
    <div class="asset-list">
      {#each assets as asset}
        <Button.Root variant="ghost" class="asset-row" onclick={() => asset.kind === 'location' ? onOpenLocation(asset) : onOpenAsset(asset)}>
          <AssetThumb {asset} />
          <span class="asset-row-main">
            <strong>{asset.title}</strong>
            {#if asset.description}
              <small>{asset.description}</small>
            {/if}
          </span>
          <span class="asset-row-meta">
            <Badge variant="outline">{asset.customAssetTypeLabel ?? assetKindLabel(asset.kind)}</Badge>
            <small>{asset.containmentTrail}</small>
          </span>
        </Button.Root>
      {/each}
    </div>
  {/if}
</section>
