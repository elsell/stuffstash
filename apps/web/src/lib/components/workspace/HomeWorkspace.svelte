<script lang="ts">
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import {
    browseAssetHref,
    browseLocationHref,
    homeAddItemHref,
    homeAddLocationHref,
    homeArchivedEmptyState,
    homeCreateLocationDenied,
    homeHeadingPresentation,
    homeLifecycleOptions,
    homeLocationsEmptyState,
    homeRecentEmptyState
  } from '$lib/application/workspaceBrowseNavigation';
  import type { Asset, AssetLifecycleFilter, AssetViewModel, LocationSummary } from '$lib/domain/inventory';
  import { assetKindLabel } from '$lib/domain/inventory';
  import AssetTagChips from './AssetTagChips.svelte';
  import AssetThumb from './AssetThumb.svelte';
  import CheckoutBadge from './CheckoutBadge.svelte';
  import SegmentedControl from './SegmentedControl.svelte';

  let {
    tenantId,
    inventoryId,
    lifecycleState,
    locations,
    recentAssets,
    archivedAssets,
    checkedOutAssets = [],
    browseMode = 'home',
    canCreateAsset = true,
    onOpenLocation,
    onOpenAsset,
    onOpenAdd,
    onSelectLifecycle
  }: {
    tenantId: string;
    inventoryId: string;
    lifecycleState: AssetLifecycleFilter;
    locations: LocationSummary[];
    recentAssets: AssetViewModel[];
    archivedAssets: Asset[];
    checkedOutAssets?: Asset[];
    browseMode?: 'home' | 'locations';
    canCreateAsset?: boolean;
    onOpenLocation: (asset: Asset) => void;
    onOpenAsset: (asset: Asset) => void;
    onOpenAdd: (kind?: 'item' | 'location') => void;
    onSelectLifecycle: (lifecycleState: AssetLifecycleFilter) => void;
  } = $props();

  let routeTenantId = $derived(
    tenantId || locations[0]?.location.tenantId || recentAssets[0]?.tenantId || archivedAssets[0]?.tenantId || null
  );
  let routeInventoryId = $derived(
    inventoryId || locations[0]?.location.inventoryId || recentAssets[0]?.inventoryId || archivedAssets[0]?.inventoryId || null
  );
  let headingPresentation = $derived(homeHeadingPresentation(lifecycleState, browseMode));
  let lifecycleOptions = $derived(homeLifecycleOptions(routeTenantId, routeInventoryId));
  const addDenied = homeCreateLocationDenied();
  const recentEmpty = homeRecentEmptyState();
  const archivedEmpty = homeArchivedEmptyState();
  let locationsEmpty = $derived(homeLocationsEmptyState(browseMode));

  function addLocationHref(): string {
    return homeAddLocationHref(routeTenantId, routeInventoryId);
  }

  function addItemHref(): string {
    return homeAddItemHref(routeTenantId, routeInventoryId);
  }

  function openAdd(event: MouseEvent, kind: 'item' | 'location' = 'location'): void {
    if (!canCreateAsset) {
      return;
    }
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    onOpenAdd(kind);
  }

  function openAsset(event: MouseEvent, asset: Asset): void {
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    onOpenAsset(asset);
  }

  function openLocation(event: MouseEvent, asset: Asset): void {
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    onOpenLocation(asset);
  }
</script>

<section class="workspace-main" aria-labelledby="home-title">
  <div class="section-heading">
    <div>
      <h1 id="home-title">{headingPresentation.title}</h1>
      <p>{headingPresentation.description}</p>
    </div>
    <div class="heading-actions">
      {#if browseMode === 'home'}
        <SegmentedControl
          label="Asset lifecycle"
          value={lifecycleState}
          options={lifecycleOptions}
          onSelect={(value) => onSelectLifecycle(value as AssetLifecycleFilter)}
        />
      {/if}
      {#if lifecycleState === 'active'}
        <Button.Root
          href={addLocationHref()}
          variant="outline"
          disabled={!canCreateAsset}
          aria-describedby={!canCreateAsset ? addDenied.id : undefined}
          onclick={(event) => openAdd(event, 'location')}
        >Add location</Button.Root>
        {#if !canCreateAsset && locations.length > 0}
          <p id={addDenied.id} class="denied-note" role="note">{addDenied.message}</p>
        {/if}
      {/if}
    </div>
  </div>

  {#if lifecycleState === 'active' && browseMode === 'home'}
    <section class="recent-section" aria-labelledby="recent-title">
      <div class="section-heading compact">
        <h2 id="recent-title">Recently added</h2>
      </div>
      {#if recentAssets.length === 0}
        <div class="empty-state">
          <p>{recentEmpty.message}</p>
        </div>
      {:else}
        <div class="recent-rail" aria-label="Recently added assets">
          {#each recentAssets as asset}
            <Button.Root href={browseAssetHref(asset)} variant="ghost" class="recent-card" onclick={(event) => openAsset(event, asset)}>
              <AssetThumb {asset} size="lg" />
              <span>
                <strong>{asset.title}</strong>
                <small>{asset.customAssetTypeLabel ?? assetKindLabel(asset.kind)}</small>
                <small>{asset.containmentTrail}</small>
                <AssetTagChips tags={asset.tags ?? []} compact overflowLimit={2} />
                {#if asset.currentCheckout}
                  <CheckoutBadge checkout={asset.currentCheckout} compact />
                {/if}
              </span>
            </Button.Root>
          {/each}
        </div>
      {/if}
    </section>
    {#if checkedOutAssets.length > 0}
      <section class="recent-section" aria-labelledby="checked-out-title">
        <div class="section-heading compact">
          <h2 id="checked-out-title">Checked out</h2>
        </div>
        <div class="asset-list compact-list">
          {#each checkedOutAssets as asset}
            <Button.Root href={browseAssetHref(asset)} variant="ghost" class="asset-row" onclick={(event) => openAsset(event, asset)}>
              <AssetThumb {asset} />
              <span class="asset-row-main">
                <strong>{asset.title}</strong>
                <small>{asset.description || assetKindLabel(asset.kind)}</small>
                <AssetTagChips tags={asset.tags ?? []} compact overflowLimit={2} />
                {#if asset.currentCheckout}
                  <CheckoutBadge checkout={asset.currentCheckout} compact />
                {/if}
              </span>
              <Badge variant="outline">{asset.lifecycleState}</Badge>
            </Button.Root>
          {/each}
        </div>
      </section>
    {/if}
  {/if}

  {#if lifecycleState === 'archived'}
    {#if archivedAssets.length === 0}
      <div class="empty-state spacious">
        <h2>{archivedEmpty.title}</h2>
      </div>
    {:else}
      <div class="asset-list">
        {#each archivedAssets as asset}
          <Button.Root href={browseAssetHref(asset)} variant="ghost" class="asset-row" onclick={(event) => openAsset(event, asset)}>
            <AssetThumb {asset} />
            <span class="asset-row-main">
              <strong>{asset.title}</strong>
              <small>{asset.description || assetKindLabel(asset.kind)}</small>
              <AssetTagChips tags={asset.tags ?? []} compact overflowLimit={2} />
              {#if asset.currentCheckout}
                <CheckoutBadge checkout={asset.currentCheckout} compact />
              {/if}
            </span>
            <Badge variant="outline">{assetKindLabel(asset.kind)}</Badge>
          </Button.Root>
        {/each}
      </div>
    {/if}
  {:else if locations.length === 0}
    <div class="empty-state spacious">
      <h2>{locationsEmpty.title}</h2>
      <p>{locationsEmpty.message}</p>
      <Button.Root
        href={addLocationHref()}
        disabled={!canCreateAsset}
        aria-describedby={!canCreateAsset ? addDenied.id : undefined}
        onclick={(event) => openAdd(event, 'location')}
      >{locationsEmpty.actionLabel}</Button.Root>
      {#if browseMode === 'home' && canCreateAsset && locationsEmpty.secondaryActionLabel}
        <Button.Root
          href={addItemHref()}
          variant="outline"
          onclick={(event) => openAdd(event, 'item')}
        >{locationsEmpty.secondaryActionLabel}</Button.Root>
      {/if}
      {#if !canCreateAsset}
        <p id={addDenied.id} class="denied-note" role="note">{addDenied.message}</p>
      {/if}
    </div>
  {:else}
    {#if browseMode === 'home'}
      <div class="section-heading compact">
        <h2>Locations</h2>
      </div>
    {/if}
    <div class="location-grid">
      {#each locations as summary}
        <Button.Root
          href={browseLocationHref(summary.location)}
          variant="ghost"
          class="location-tile"
          aria-label={`Open location ${summary.location.title}`}
          onclick={(event) => openLocation(event, summary.location)}
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
