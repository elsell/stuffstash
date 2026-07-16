<script lang="ts">
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import ChevronLeft from '@lucide/svelte/icons/chevron-left';
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import { onMount } from 'svelte';
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
    homeLocationsHref,
    homeRecentEmptyState,
    visibleAssetCountLabel
  } from '$lib/application/workspaceBrowseNavigation';
  import type { Asset, AssetLifecycleFilter, AssetTag, AssetViewModel, LocationAsset, LocationSummary } from '$lib/domain/inventory';
  import { assetKindLabel } from '$lib/domain/inventory';
  import { homeLocationPreview } from '$lib/application/workspace';
  import AssetTagChips from './AssetTagChips.svelte';
  import AssetThumb from './AssetThumb.svelte';
  import CheckoutBadge from './CheckoutBadge.svelte';
  import SegmentedControl from './SegmentedControl.svelte';

  let recentRail = $state<HTMLElement | null>(null);
  let canScrollRecentBackward = $state(false);
  let canScrollRecentForward = $state(false);
  let hasRecentOverflow = $state(false);

  function updateRecentRailControls(): void {
    if (!recentRail) {
      canScrollRecentBackward = false;
      canScrollRecentForward = false;
      hasRecentOverflow = false;
      return;
    }
    const edgeTolerance = 2;
    hasRecentOverflow = recentRail.scrollWidth > recentRail.clientWidth + edgeTolerance;
    canScrollRecentBackward = recentRail.scrollLeft > edgeTolerance;
    canScrollRecentForward = recentRail.scrollLeft + recentRail.clientWidth < recentRail.scrollWidth - edgeTolerance;
  }

  function scrollRecent(direction: -1 | 1): void {
    if (!recentRail) return;
    const reducedMotion = typeof window.matchMedia === 'function' && window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    recentRail.scrollBy({
      left: direction * Math.round(recentRail.clientWidth * 0.85),
      behavior: reducedMotion ? 'auto' : 'smooth'
    });
  }

  $effect(() => {
    recentAssets.length;
    recentRail;
    queueMicrotask(updateRecentRailControls);
  });

  onMount(() => {
    const observer = typeof ResizeObserver === 'undefined' ? null : new ResizeObserver(updateRecentRailControls);
    if (recentRail) observer?.observe(recentRail);
    window.addEventListener('resize', updateRecentRailControls);
    return () => {
      observer?.disconnect();
      window.removeEventListener('resize', updateRecentRailControls);
    };
  });

  let {
    tenantId,
    inventoryId,
    lifecycleState,
    locations,
    recentAssets,
    archivedAssets,
    checkedOutAssets = [],
    canCreateAsset = true,
    canEditAsset = false,
    onOpenLocation,
    onOpenLocations = () => {},
    onOpenAsset,
    onReturnAsset = async () => {},
    onOpenAdd,
    onSelectLifecycle,
    onTagSearch
  }: {
    tenantId: string;
    inventoryId: string;
    lifecycleState: AssetLifecycleFilter;
    locations: LocationSummary[];
    recentAssets: AssetViewModel[];
    archivedAssets: Asset[];
    checkedOutAssets?: Asset[];
    canCreateAsset?: boolean;
    canEditAsset?: boolean;
    onOpenLocation: (asset: Asset) => void;
    onOpenLocations?: () => void;
    onOpenAsset: (asset: Asset) => void;
    onReturnAsset?: (asset: Asset) => Promise<void>;
    onOpenAdd: (kind?: 'item' | 'location', parentAssetId?: string | null, opener?: HTMLElement | null) => void;
    onSelectLifecycle: (lifecycleState: AssetLifecycleFilter) => void;
    onTagSearch?: (tag: AssetTag) => void;
  } = $props();

  let routeTenantId = $derived(
    tenantId || locations[0]?.location.tenantId || recentAssets[0]?.tenantId || archivedAssets[0]?.tenantId || null
  );
  let routeInventoryId = $derived(
    inventoryId || locations[0]?.location.inventoryId || recentAssets[0]?.inventoryId || archivedAssets[0]?.inventoryId || null
  );
  let headingPresentation = $derived(homeHeadingPresentation(lifecycleState));
  let lifecycleOptions = $derived(homeLifecycleOptions(routeTenantId, routeInventoryId));
  const addDenied = homeCreateLocationDenied();
  const recentEmpty = homeRecentEmptyState();
  const archivedEmpty = homeArchivedEmptyState();
  let locationsEmpty = $derived(homeLocationsEmptyState());
  let displayedLocations = $derived(homeLocationPreview(locations));
  let returningAssetId = $state<string | null>(null);

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
    onOpenAdd(kind, null, event.currentTarget as HTMLElement);
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

  function recentAssetHref(asset: Asset): string {
    return asset.kind === 'location' ? browseLocationHref(asset as LocationAsset) : browseAssetHref(asset);
  }

  function openRecentAsset(event: MouseEvent, asset: Asset): void {
    if (asset.kind === 'location') {
      openLocation(event, asset);
      return;
    }
    openAsset(event, asset);
  }

  function openLocations(event: MouseEvent): void {
    if (!shouldHandleWorkspaceLinkClick(event)) return;
    event.preventDefault();
    onOpenLocations();
  }

  async function returnAsset(asset: Asset): Promise<void> {
    if (!canEditAsset || returningAssetId) return;
    returningAssetId = asset.id;
    try {
      await onReturnAsset(asset);
    } finally {
      returningAssetId = null;
    }
  }
</script>

<section class="workspace-main" aria-labelledby="home-title">
  <div class="section-heading home-heading">
    <div>
      <h1 id="home-title">{headingPresentation.title}</h1>
      <p>{headingPresentation.description}</p>
    </div>
    <div class="heading-actions">
      <SegmentedControl
        label="Asset lifecycle"
        value={lifecycleState}
        options={lifecycleOptions}
        onSelect={(value) => onSelectLifecycle(value as AssetLifecycleFilter)}
      />
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

  {#if lifecycleState === 'active'}
    <section class="recent-section" aria-labelledby="recent-title">
      <div class="section-heading compact recent-heading">
        <h2 id="recent-title">Recently changed</h2>
        {#if hasRecentOverflow}
          <div class="recent-rail-controls" aria-label="Recently changed navigation">
            <Button.Root
              variant="outline"
              size="icon"
              aria-label="Previous recently changed assets"
              disabled={!canScrollRecentBackward}
              onclick={() => scrollRecent(-1)}
            ><ChevronLeft aria-hidden="true" /></Button.Root>
            <Button.Root
              variant="outline"
              size="icon"
              aria-label="Next recently changed assets"
              disabled={!canScrollRecentForward}
              onclick={() => scrollRecent(1)}
            ><ChevronRight aria-hidden="true" /></Button.Root>
          </div>
        {/if}
      </div>
      {#if recentAssets.length === 0}
        <div class="empty-state">
          <p>{recentEmpty.message}</p>
        </div>
      {:else}
        <div bind:this={recentRail} class="recent-rail" aria-label="Recently changed assets" onscroll={updateRecentRailControls}>
          {#each recentAssets as asset}
            <article class="recent-card" data-recent-card={asset.id}>
              <Button.Root href={recentAssetHref(asset)} variant="ghost" class="recent-card-open" data-recent-card-link onclick={(event) => openRecentAsset(event, asset)}>
                <div data-recent-card-media><AssetThumb {asset} size="lg" /></div>
                <span class="recent-card-copy" data-recent-card-copy>
                  <strong data-recent-card-title>{asset.title}</strong>
                  <small>{asset.customAssetTypeLabel ?? assetKindLabel(asset.kind)}</small>
                  <small>{asset.containmentTrail}</small>
                  {#if asset.currentCheckout}
                    <CheckoutBadge checkout={asset.currentCheckout} compact />
                  {/if}
                </span>
              </Button.Root>
              <div class="recent-card-tags" data-recent-card-tags>
                <AssetTagChips tags={asset.tags ?? []} compact overflowLimit={2} onTagSelect={onTagSearch} />
              </div>
            </article>
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
            <div class="asset-row">
              <Button.Root href={browseAssetHref(asset)} variant="ghost" class="asset-row-open" onclick={(event) => openAsset(event, asset)}>
                <AssetThumb {asset} />
                <span class="asset-row-main">
                  <strong>{asset.title}</strong>
                  <small>{asset.description || assetKindLabel(asset.kind)}</small>
                  {#if asset.currentCheckout}
                    <CheckoutBadge checkout={asset.currentCheckout} compact />
                  {/if}
                </span>
              </Button.Root>
              <div class="asset-row-actions">
                <AssetTagChips tags={asset.tags ?? []} compact overflowLimit={2} onTagSelect={onTagSearch} />
                {#if asset.lifecycleState === 'archived'}
                  <Badge variant="outline">Archived</Badge>
                {/if}
                {#if canEditAsset}
                  <Button.Root
                    variant="outline"
                    aria-label={returningAssetId === asset.id ? `Returning ${asset.title}` : `Return ${asset.title}`}
                    aria-busy={returningAssetId === asset.id}
                    disabled={returningAssetId !== null}
                    onclick={() => returnAsset(asset)}
                  >{returningAssetId === asset.id ? 'Returning…' : 'Return'}</Button.Root>
                {/if}
              </div>
            </div>
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
          <div class="asset-row">
            <Button.Root href={browseAssetHref(asset)} variant="ghost" class="asset-row-open" onclick={(event) => openAsset(event, asset)}>
              <AssetThumb {asset} />
              <span class="asset-row-main">
                <strong>{asset.title}</strong>
                <small>{asset.description || assetKindLabel(asset.kind)}</small>
                {#if asset.currentCheckout}
                  <CheckoutBadge checkout={asset.currentCheckout} compact />
                {/if}
              </span>
            </Button.Root>
            <div class="asset-row-actions">
              <AssetTagChips tags={asset.tags ?? []} compact overflowLimit={2} onTagSelect={onTagSearch} />
              <Badge variant="outline">Archived</Badge>
            </div>
          </div>
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
      {#if canCreateAsset && locationsEmpty.secondaryActionLabel}
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
    <div class="section-heading compact locations-heading">
      <h2>Places</h2>
      {#if locations.length > displayedLocations.length}
        <Button.Root href={homeLocationsHref(routeTenantId, routeInventoryId)} variant="ghost" onclick={openLocations}>
          View all places
        </Button.Root>
      {/if}
    </div>
    <div class="location-grid">
      {#each displayedLocations as summary}
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
            <small>{visibleAssetCountLabel(summary.assetCount)}</small>
          </span>
          <ChevronRight aria-hidden="true" />
        </Button.Root>
      {/each}
    </div>
  {/if}

</section>
