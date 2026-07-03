<script lang="ts">
  import ArrowLeft from '@lucide/svelte/icons/arrow-left';
  import Plus from '@lucide/svelte/icons/plus';
  import Pencil from '@lucide/svelte/icons/pencil';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { workspaceRouteHref } from '$lib/application/workspaceRoute';
  import type { Asset, AssetViewModel } from '$lib/domain/inventory';
  import { assetKindLabel } from '$lib/domain/inventory';
  import AssetThumb from './AssetThumb.svelte';

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
    location: Asset;
    assets: AssetViewModel[];
    canEdit: boolean;
    canCreateAsset?: boolean;
    onBack: () => void;
    onOpenLocation: (asset: Asset) => void;
    onEditLocation: (asset: Asset) => void;
    onOpenAsset: (asset: Asset) => void;
    onOpenAdd?: (kind: 'item', parentAssetId: string) => void;
  } = $props();

  function backHref(): string {
    return workspaceRouteHref({ mode: 'locations' }, location.tenantId, location.inventoryId);
  }

  function editLocationHref(): string {
    return workspaceRouteHref(
      { mode: 'asset', locationId: location.id, assetId: location.id, action: 'edit', assetAction: 'edit' },
      location.tenantId,
      location.inventoryId
    );
  }

  function addItemHereHref(): string {
    return workspaceRouteHref(
      { action: 'add', addKind: 'item', addParentAssetId: location.id },
      location.tenantId,
      location.inventoryId
    );
  }

  function rowHref(asset: Asset): string {
    if (asset.kind === 'location') {
      return workspaceRouteHref({ mode: 'location', locationId: asset.id }, asset.tenantId, asset.inventoryId);
    }
    return workspaceRouteHref({ mode: 'asset', assetId: asset.id }, asset.tenantId, asset.inventoryId);
  }

  function openBack(event: MouseEvent): void {
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    onBack();
  }

  function openEditLocation(event: MouseEvent): void {
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    onEditLocation(location);
  }

  function openAddItemHere(event: MouseEvent): void {
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    onOpenAdd('item', location.id);
  }

  function openRow(event: MouseEvent, asset: Asset): void {
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    if (asset.kind === 'location') {
      onOpenLocation(asset);
    } else {
      onOpenAsset(asset);
    }
  }

  function shouldHandleInApp(event: MouseEvent): boolean {
    return event.button === 0 && !event.metaKey && !event.ctrlKey && !event.shiftKey && !event.altKey;
  }
</script>

<section class="workspace-main" aria-labelledby="location-title">
  <Button.Root href={backHref()} variant="ghost" class="back-button" onclick={openBack}><ArrowLeft /> Back</Button.Root>
  <div class="location-hero">
    <AssetThumb asset={location} size="lg" />
    <div>
      <h1 id="location-title">{location.title}</h1>
      <p>{location.description}</p>
      <Badge variant="secondary">{assets.length} visible assets</Badge>
    </div>
    {#if canEdit}
      <Button.Root href={editLocationHref()} variant="outline" onclick={openEditLocation}><Pencil /> Edit location</Button.Root>
    {/if}
    {#if canCreateAsset && assets.length > 0}
      <Button.Root href={addItemHereHref()} onclick={openAddItemHere}><Plus /> Add item here</Button.Root>
    {/if}
  </div>

  {#if assets.length === 0}
    <div class="empty-state spacious">
      <h2>No stuff here yet</h2>
      <p>{canCreateAsset ? 'Add an item or move existing stuff into this location.' : 'This location is empty.'}</p>
      {#if canCreateAsset}
        <Button.Root href={addItemHereHref()} onclick={openAddItemHere}>Add item here</Button.Root>
      {:else}
        <p class="denied-note" role="note">Adding items is unavailable for this inventory.</p>
      {/if}
    </div>
  {:else}
    <div class="asset-list">
      {#each assets as asset}
        <Button.Root href={rowHref(asset)} variant="ghost" class="asset-row" onclick={(event) => openRow(event, asset)}>
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
