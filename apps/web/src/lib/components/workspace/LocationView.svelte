<script lang="ts">
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import ArrowLeft from '@lucide/svelte/icons/arrow-left';
  import Pencil from '@lucide/svelte/icons/pencil';
  import MoveRight from '@lucide/svelte/icons/move-right';
  import Archive from '@lucide/svelte/icons/archive';
  import * as Button from '$lib/components/ui/button/index.js';
  import {
    locationBackHref,
    locationEditHref
  } from '$lib/application/workspaceBrowseNavigation';
  import type { Asset, AssetTag, AssetViewModel, LocationAsset } from '$lib/domain/inventory';
  import { assetActionHref } from '$lib/application/workspaceAssetActions';
  import AssetThumb from './AssetThumb.svelte';
  import ContainedAssetWorkspace from './ContainedAssetWorkspace.svelte';

  let {
    location,
    assets = [],
    workspaceAssets = assets,
    canEdit,
    canCreateAsset = false,
    saving = false,
    moveHereOpen = false,
    onBack,
    onOpenLocation,
    onEditLocation,
    onOpenAsset,
    onOpenAdd = () => {},
    onOpenMoveHere = () => {},
    onCloseMoveHere = () => {},
    onMoveHere = async () => {},
    onTagSearch
  }: {
    location: LocationAsset;
    assets?: AssetViewModel[];
    workspaceAssets?: Asset[];
    canEdit: boolean;
    canCreateAsset?: boolean;
    saving?: boolean;
    moveHereOpen?: boolean;
    onBack: () => void;
    onOpenLocation: (asset: Asset) => void;
    onEditLocation: (asset: Asset) => void;
    onOpenAsset: (asset: Asset) => void;
    onOpenAdd?: (kind: 'item', parentAssetId: string) => void;
    onOpenMoveHere?: () => void;
    onCloseMoveHere?: () => void;
    onMoveHere?: (asset: Asset) => Promise<void>;
    onTagSearch?: (tag: AssetTag) => void;
  } = $props();

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

</script>

<section class="workspace-main location-view" aria-labelledby="location-title">
  <Button.Root href={locationBackHref(location)} variant="ghost" class="back-button" onclick={openBack}><ArrowLeft /> Back</Button.Root>
  <header class="location-detail-hero">
    <AssetThumb asset={location} size="lg" />
    <div class="location-identity">
      <span class="location-kind-label">Location</span>
      <h1 id="location-title">{location.title}</h1>
      {#if location.description}<p>{location.description}</p>{/if}
    </div>
    {#if canEdit}
      <div class="location-maintenance-actions" aria-label="Place maintenance">
        <Button.Root href={locationEditHref(location)} variant="outline" onclick={openEditLocation}><Pencil /> Edit location</Button.Root>
        <Button.Root href={assetActionHref(location, 'move')} variant="outline"><MoveRight /> Move place</Button.Root>
        <Button.Root href={assetActionHref(location, 'archive')} variant="ghost"><Archive /> Archive</Button.Root>
      </div>
    {/if}
  </header>
  <ContainedAssetWorkspace
    target={{ ...location, containmentTrail: '' }}
    assets={workspaceAssets}
    canCreate={canCreateAsset}
    {canEdit}
    {saving}
    {moveHereOpen}
    onOpenAsset={(asset) => asset.kind === 'location' ? onOpenLocation(asset) : onOpenAsset(asset)}
    {onOpenAdd}
    {onOpenMoveHere}
    {onCloseMoveHere}
    {onMoveHere}
    {onTagSearch}
  />
</section>
