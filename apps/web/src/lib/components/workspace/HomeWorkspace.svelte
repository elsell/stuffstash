<script lang="ts">
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { workspaceRouteHref } from '$lib/application/workspaceRoute';
  import type { Asset, AssetLifecycleFilter, AssetViewModel, LocationSummary } from '$lib/domain/inventory';
  import { assetKindLabel } from '$lib/domain/inventory';
  import AssetThumb from './AssetThumb.svelte';
  import SegmentedControl from './SegmentedControl.svelte';

  let {
    tenantId,
    inventoryId,
    lifecycleState,
    locations,
    recentAssets,
    archivedAssets,
    browseMode = 'home',
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
    browseMode?: 'home' | 'locations';
    onOpenLocation: (asset: Asset) => void;
    onOpenAsset: (asset: Asset) => void;
    onOpenAdd: () => void;
    onSelectLifecycle: (lifecycleState: AssetLifecycleFilter) => void;
  } = $props();

  let routeTenantId = $derived(
    tenantId || locations[0]?.location.tenantId || recentAssets[0]?.tenantId || archivedAssets[0]?.tenantId || null
  );
  let routeInventoryId = $derived(
    inventoryId || locations[0]?.location.inventoryId || recentAssets[0]?.inventoryId || archivedAssets[0]?.inventoryId || null
  );

  function addLocationHref(): string {
    return workspaceRouteHref({ action: 'add', addKind: 'location' }, routeTenantId, routeInventoryId);
  }

  function assetHref(asset: Asset): string {
    return workspaceRouteHref({ mode: 'asset', tenantId: asset.tenantId, inventoryId: asset.inventoryId, assetId: asset.id }, asset.tenantId, asset.inventoryId);
  }

  function locationHref(asset: Asset): string {
    return workspaceRouteHref(
      { mode: 'location', tenantId: asset.tenantId, inventoryId: asset.inventoryId, locationId: asset.id },
      asset.tenantId,
      asset.inventoryId
    );
  }

  function openAdd(event: MouseEvent): void {
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    onOpenAdd();
  }

  function openAsset(event: MouseEvent, asset: Asset): void {
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    onOpenAsset(asset);
  }

  function openLocation(event: MouseEvent, asset: Asset): void {
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    onOpenLocation(asset);
  }

  function shouldHandleInApp(event: MouseEvent): boolean {
    return event.button === 0 && !event.metaKey && !event.ctrlKey && !event.shiftKey && !event.altKey;
  }
</script>

<section class="workspace-main" aria-labelledby="home-title">
  <div class="section-heading">
    <div>
      <h1 id="home-title">{lifecycleState === 'active' ? (browseMode === 'locations' ? 'Locations' : 'Home') : 'Archived assets'}</h1>
      <p>
        {lifecycleState === 'active'
          ? browseMode === 'locations'
            ? 'The places where your things live.'
            : 'Recently added and the places where your things live.'
          : 'Assets removed from active browsing.'}
      </p>
    </div>
    <div class="heading-actions">
      {#if browseMode === 'home'}
        <SegmentedControl
          label="Asset lifecycle"
          value={lifecycleState}
          options={[
            { value: 'active', label: 'Active' },
            { value: 'archived', label: 'Archived' }
          ]}
          onSelect={(value) => onSelectLifecycle(value as AssetLifecycleFilter)}
        />
      {/if}
      {#if lifecycleState === 'active'}
        <Button.Root href={addLocationHref()} variant="outline" onclick={openAdd}>Add location</Button.Root>
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
          <p>No items or containers yet.</p>
        </div>
      {:else}
        <div class="recent-rail" aria-label="Recently added assets">
          {#each recentAssets as asset}
            <Button.Root href={assetHref(asset)} variant="ghost" class="recent-card" onclick={(event) => openAsset(event, asset)}>
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
          <Button.Root href={assetHref(asset)} variant="ghost" class="asset-row" onclick={(event) => openAsset(event, asset)}>
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
      <Button.Root href={addLocationHref()} onclick={openAdd}>Add first location</Button.Root>
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
          href={locationHref(summary.location)}
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
