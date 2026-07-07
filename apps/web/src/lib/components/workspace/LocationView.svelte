<script lang="ts">
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import ArrowLeft from '@lucide/svelte/icons/arrow-left';
  import Plus from '@lucide/svelte/icons/plus';
  import Pencil from '@lucide/svelte/icons/pencil';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import {
    locationAddItemHref,
    locationBackHref,
    locationEditHref,
    locationEmptyState,
    locationRowHref
  } from '$lib/application/workspaceBrowseNavigation';
  import type { Asset, AssetViewModel, LocationAsset } from '$lib/domain/inventory';
  import { assetKindLabel } from '$lib/domain/inventory';
  import AssetTagChips from './AssetTagChips.svelte';
  import AssetThumb from './AssetThumb.svelte';
  import CheckoutBadge from './CheckoutBadge.svelte';

  let {
    location,
    assets,
    canEdit,
    canCreateAsset = false,
    onBack,
    onOpenLocation,
    onEditLocation,
    onOpenAsset,
    onOpenAdd = () => {}
  }: {
    location: LocationAsset;
    assets: AssetViewModel[];
    canEdit: boolean;
    canCreateAsset?: boolean;
    onBack: () => void;
    onOpenLocation: (asset: Asset) => void;
    onEditLocation: (asset: Asset) => void;
    onOpenAsset: (asset: Asset) => void;
    onOpenAdd?: (kind: 'item', parentAssetId: string) => void;
  } = $props();

  let emptyState = $derived(locationEmptyState(canCreateAsset));

  function openBack(event: MouseEvent): void {
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    onBack();
  }

  function openEditLocation(event: MouseEvent): void {
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    onEditLocation(location);
  }

  function openAddItemHere(event: MouseEvent): void {
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    onOpenAdd('item', location.id);
  }

  function openRow(event: MouseEvent, asset: Asset): void {
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    if (asset.kind === 'location') {
      onOpenLocation(asset);
    } else {
      onOpenAsset(asset);
    }
  }
</script>

<section class="workspace-main" aria-labelledby="location-title">
  <Button.Root href={locationBackHref(location)} variant="ghost" class="back-button" onclick={openBack}><ArrowLeft /> Back</Button.Root>
  <div class="location-hero">
    <AssetThumb asset={location} size="lg" />
    <div>
      <h1 id="location-title">{location.title}</h1>
      <p>{location.description}</p>
      <Badge variant="secondary">{assets.length} visible assets</Badge>
    </div>
    {#if canEdit}
      <Button.Root href={locationEditHref(location)} variant="outline" onclick={openEditLocation}><Pencil /> Edit location</Button.Root>
    {/if}
    {#if canCreateAsset && assets.length > 0}
      <Button.Root href={locationAddItemHref(location)} onclick={openAddItemHere}><Plus /> {emptyState.actionLabel}</Button.Root>
    {/if}
  </div>

  {#if assets.length === 0}
    <div class="empty-state spacious">
      <h2>{emptyState.title}</h2>
      <p>{emptyState.message}</p>
      {#if canCreateAsset}
        <Button.Root href={locationAddItemHref(location)} onclick={openAddItemHere}>{emptyState.actionLabel}</Button.Root>
      {:else}
        <p class="denied-note" role="note">{emptyState.deniedMessage}</p>
      {/if}
    </div>
  {:else}
    <div class="asset-list">
      {#each assets as asset}
        <Button.Root href={locationRowHref(asset)} variant="ghost" class="asset-row" onclick={(event) => openRow(event, asset)}>
          <AssetThumb {asset} />
          <span class="asset-row-main">
            <strong>{asset.title}</strong>
            {#if asset.description}
              <small>{asset.description}</small>
            {/if}
            <AssetTagChips tags={asset.tags ?? []} compact overflowLimit={2} />
            {#if asset.currentCheckout}
              <CheckoutBadge checkout={asset.currentCheckout} compact />
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
