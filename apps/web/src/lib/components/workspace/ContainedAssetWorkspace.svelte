<script lang="ts" module>
  import type { Asset, AssetTag, AssetViewModel } from '$lib/domain/inventory';

  export type ContainedAssetWorkspaceProps = {
    target: AssetViewModel;
    assets: Asset[];
    canCreate: boolean;
    canEdit: boolean;
    saving: boolean;
    moveHereOpen: boolean;
    onOpenAsset: (asset: Asset) => void;
    onOpenAdd: (kind: 'item', parentAssetId: string) => void;
    onOpenMoveHere: () => void;
    onCloseMoveHere: () => void;
    onMoveHere: (asset: Asset) => Promise<void>;
    onTagSearch?: (tag: AssetTag) => void;
  };
</script>

<script lang="ts">
  import { tick } from 'svelte';
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import MoveRight from '@lucide/svelte/icons/move-right';
  import Plus from '@lucide/svelte/icons/plus';
  import Search from '@lucide/svelte/icons/search';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import { assetDetailHref } from '$lib/application/workspaceAssetActions';
  import { assetKindLabel } from '$lib/domain/inventory';
  import {
    addItemHereHref,
    containableWorkspaceSections,
    moveHereCandidatePage,
    moveHereHref
  } from '$lib/application/workspaceContainedAssets';
  import AssetThumb from './AssetThumb.svelte';
  import AssetTagChips from './AssetTagChips.svelte';
  import CheckoutBadge from './CheckoutBadge.svelte';
  import WorkspaceTaskSheet from './action-surface/WorkspaceTaskSheet.svelte';

  let {
    target,
    assets,
    canCreate,
    canEdit,
    saving,
    moveHereOpen,
    onOpenAsset,
    onOpenAdd,
    onOpenMoveHere,
    onCloseMoveHere,
    onMoveHere,
    onTagSearch
  }: ContainedAssetWorkspaceProps = $props();

  let contentsQuery = $state('');
  let moveQuery = $state('');
  let selectedCandidateId = $state<string | null>(null);
  let showAllCandidates = $state(false);
  let saveError = $state('');
  let wasMoveHereOpen = $state(false);
  let moveTrigger = $state<HTMLElement | null>(null);

  let sections = $derived(containableWorkspaceSections(target, assets));
  let totalContainedRows = $derived(sections.reduce((count, section) => count + section.assets.length, 0));
  let showContentsSearch = $derived(target.kind === 'location' && totalContainedRows >= 20);
  let visibleSections = $derived(sections.map((section) => ({
    ...section,
    assets: filterContents(section.assets, contentsQuery)
  })));
  let candidatePage = $derived(moveHereCandidatePage(target, assets, moveQuery, showAllCandidates ? Number.MAX_SAFE_INTEGER : 8));
  let selectedCandidate = $derived(assets.find((candidate) => candidate.id === selectedCandidateId) ?? null);
  let canAddHere = $derived(canCreate && canEdit && target.lifecycleState === 'active' && !saving);
  let canMoveHere = $derived(canEdit && target.lifecycleState === 'active' && !saving);

  $effect(() => {
    if (moveHereOpen && !wasMoveHereOpen) {
      moveQuery = '';
      selectedCandidateId = null;
      showAllCandidates = false;
      saveError = '';
    } else if (!moveHereOpen && wasMoveHereOpen) {
      void tick().then(() => moveTrigger?.focus());
    }
    wasMoveHereOpen = moveHereOpen;
  });

  function openAsset(event: MouseEvent, asset: Asset): void {
    if (!shouldHandleWorkspaceLinkClick(event)) return;
    event.preventDefault();
    onOpenAsset(asset);
  }

  function openAdd(event: MouseEvent): void {
    if (!canAddHere || !shouldHandleWorkspaceLinkClick(event)) return;
    event.preventDefault();
    onOpenAdd('item', target.id);
  }

  function openMove(event: MouseEvent): void {
    if (!canMoveHere || !shouldHandleWorkspaceLinkClick(event)) return;
    event.preventDefault();
    onOpenMoveHere();
  }

  function closeMove(event: MouseEvent): void {
    if (!shouldHandleWorkspaceLinkClick(event)) return;
    event.preventDefault();
    onCloseMoveHere();
  }

  async function confirmMove(): Promise<void> {
    if (!selectedCandidate || saving) return;
    saveError = '';
    try {
      await onMoveHere(selectedCandidate);
    } catch (caught) {
      saveError = caught instanceof Error ? caught.message : `Move not saved. ${selectedCandidate.title} stayed where it was.`;
    }
  }

  function filterContents<T extends { title: string; relativePath: string }>(items: T[], query: string): T[] {
    const normalized = query.trim().toLocaleLowerCase();
    if (!normalized) return items;
    return items.filter((item) => `${item.title} ${item.relativePath}`.toLocaleLowerCase().includes(normalized));
  }
</script>

<section
  class:location-contained-workspace={target.kind === 'location'}
  class="contained-workspace"
  aria-label={`${target.title} contents`}
>
  {#if canAddHere || canMoveHere}
    <div class="contained-spatial-actions" aria-label="Spatial actions">
      {#if canAddHere}
        <Button.Root
          href={addItemHereHref(target)}
          data-workspace-add-return-focus={target.kind === 'location' ? 'location-item' : `container-item-${target.id}`}
          onclick={openAdd}
        ><Plus /> Add item here</Button.Root>
      {/if}
      {#if canMoveHere}
        <Button.Root
          bind:ref={moveTrigger}
          href={moveHereHref(target)}
          data-workspace-move-here-trigger
          variant="outline"
          onclick={openMove}
        ><MoveRight /> Move items here</Button.Root>
      {/if}
    </div>
  {/if}
  {#if target.lifecycleState === 'active' && !canAddHere && !saving}
    <p class="denied-note" role="note">Adding items is unavailable for this inventory.{canEdit ? '' : ' Moving items is also unavailable.'}</p>
  {/if}

  {#if moveHereOpen}
    <WorkspaceTaskSheet open title="Move items here" description={`Choose one item, container, or place to move into ${target.title}.`} busy={saving} dismissible={!selectedCandidate} closeHref={assetDetailHref(target)} closeLabel="Close move items here" initialFocusSelector="#move-here-search" onCloseLink={closeMove} onOpenChange={(open) => { if (!open) onCloseMoveHere(); }} onCloseAutoFocus={(event) => { if (moveTrigger?.isConnected) { event.preventDefault(); moveTrigger.focus(); } }}>
      <div class="field-stack">
        <Label for="move-here-search">Find an asset</Label>
        <div class="relative">
          <Search class="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
          <Input class="pl-9" id="move-here-search" bind:value={moveQuery} placeholder="Search by name or current place" />
        </div>
      </div>
      <p class="visually-hidden" aria-live="polite">{candidatePage.totalCount} eligible {candidatePage.totalCount === 1 ? 'asset' : 'assets'}</p>
      {#if candidatePage.candidates.length > 0}
        <div class="grid gap-2" data-move-here-candidates role="group" aria-label="Eligible assets">
          {#each candidatePage.candidates as candidate (candidate.id)}
            <Button.Root
              variant={candidate.id === selectedCandidateId ? 'secondary' : 'ghost'}
              class="h-auto min-h-14 w-full justify-start gap-3 whitespace-normal rounded-xl border border-transparent px-3 py-2 text-left"
              aria-label={`Select ${candidate.title}`}
              aria-pressed={candidate.id === selectedCandidateId}
              onclick={() => { selectedCandidateId = candidate.id; saveError = ''; }}
            >
              <AssetThumb asset={candidate} />
              <span class="grid min-w-0 flex-1 gap-0.5" data-move-here-candidate-copy>
                <strong class="break-words text-sm leading-tight">{candidate.title}</strong>
                <small class="break-words text-xs leading-snug text-muted-foreground">
                  {assetKindLabel(candidate.kind)} · {candidate.containmentTrail || 'Inventory root'}
                </small>
              </span>
            </Button.Root>
          {/each}
        </div>
        {#if candidatePage.hasMore}
          <Button.Root variant="ghost" onclick={() => { showAllCandidates = true; }}>Show all {candidatePage.totalCount}</Button.Root>
        {/if}
      {:else}
        <div class="empty-state compact">
          <h3>{moveQuery.trim() ? 'No matching movable assets' : 'Everything eligible is already here'}</h3>
          <p>{moveQuery.trim() ? 'Try another name or current place.' : 'Search after adding or moving something elsewhere.'}</p>
          {#if moveQuery.trim()}
            <Button.Root variant="outline" onclick={() => { moveQuery = ''; }}>Clear search</Button.Root>
          {/if}
        </div>
      {/if}
      {#if selectedCandidate}
        <p class="move-here-preview">Move <strong>{selectedCandidate.title}</strong> into <strong>{target.title}</strong>.</p>
      {/if}
      {#if saveError}<p class="denied-note" role="alert">{saveError}</p>{/if}
      {#snippet footer()}
        <Button.Root href={assetDetailHref(target)} variant="outline" disabled={saving} onclick={closeMove}>Cancel</Button.Root>
        <Button.Root disabled={!selectedCandidate || saving} onclick={() => { void confirmMove(); }}>
          {selectedCandidate ? `Move ${selectedCandidate.title} here` : 'Choose an asset'}
        </Button.Root>
      {/snippet}
    </WorkspaceTaskSheet>
  {/if}

  {#if showContentsSearch}
    <div class="field-stack contained-contents-search">
      <Label for={`contents-search-${target.id}`}>Search contents</Label>
      <Input id={`contents-search-${target.id}`} bind:value={contentsQuery} placeholder="Search this place" />
    </div>
  {/if}

  {#each visibleSections as section (section.key)}
    <section class="contained-section" aria-labelledby={`contained-${target.id}-${section.key}`}>
      <div class="contained-section-heading">
        <h2 id={`contained-${target.id}-${section.key}`}>{section.heading}</h2>
        <span>{section.assets.length} {section.countNoun}{section.assets.length === 1 ? '' : 's'}</span>
      </div>
      {#if section.assets.length > 0}
        <div class="contained-asset-list">
          {#each section.assets as asset (asset.id)}
            <article class="contained-asset-item">
              <Button.Root
                href={assetDetailHref(asset)}
                variant="ghost"
                class="contained-asset-row"
                onclick={(event) => openAsset(event, asset)}
              >
                <AssetThumb {asset} />
                <span class="contained-asset-copy">
                  <strong>{asset.title}</strong>
                  <small>{assetKindLabel(asset.kind)}{asset.relativePath ? ` · ${asset.relativePath}` : ''}</small>
                  {#if asset.currentCheckout}<CheckoutBadge checkout={asset.currentCheckout} compact />{/if}
                </span>
                <ChevronRight class="contained-asset-row-indicator" aria-hidden="true" />
              </Button.Root>
              {#if asset.tags?.length}
                <div class="contained-asset-tags">
                  <AssetTagChips tags={asset.tags} compact overflowLimit={2} onTagSelect={onTagSearch} />
                </div>
              {/if}
            </article>
          {/each}
        </div>
      {:else}
        <div class="empty-state compact">
          <h3>{contentsQuery.trim() ? `No matching ${section.key}` : section.emptyTitle}</h3>
          <p>{contentsQuery.trim() ? 'Try another name or path.' : section.emptyMessage}</p>
          {#if contentsQuery.trim()}<Button.Root variant="outline" onclick={() => { contentsQuery = ''; }}>Clear search</Button.Root>{/if}
        </div>
      {/if}
    </section>
  {/each}
</section>
