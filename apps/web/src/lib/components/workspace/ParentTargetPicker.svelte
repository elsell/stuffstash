<script lang="ts">
  import X from '@lucide/svelte/icons/x';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import type { AssetViewModel } from '$lib/domain/inventory';
  import { assetKindLabel } from '$lib/domain/inventory';
  import AssetThumb from './AssetThumb.svelte';
  import ParentTargetButton from './ParentTargetButton.svelte';

  let {
    legend,
    searchId,
    searchLabel = 'Find parent',
    groupLabel,
    rootLabel = 'Inventory root',
    rootSummaryLabel = 'inventory root',
    searchPlaceholder = 'Search locations or containers',
    search = $bindable(''),
    selectedId,
    targets,
    visibleLimit = 8,
    onSelect
  }: {
    legend: string;
    searchId: string;
    searchLabel?: string;
    groupLabel: string;
    rootLabel?: string;
    rootSummaryLabel?: string;
    searchPlaceholder?: string;
    search: string;
    selectedId: string | null;
    targets: AssetViewModel[];
    visibleLimit?: number;
    onSelect: (id: string | null) => void;
  } = $props();

  let normalizedSearch = $derived(search.trim().toLowerCase());
  let matchingTargets = $derived(
    targets.filter((target) => {
      if (!normalizedSearch) {
        return true;
      }
      return target.title.toLowerCase().includes(normalizedSearch) || target.containmentTrail.toLowerCase().includes(normalizedSearch);
    })
  );
  let sortedMatchingTargets = $derived([...matchingTargets].sort((left, right) => compareTargetsForSearch(left, right, normalizedSearch)));
  let visibleTargets = $derived(sortedMatchingTargets.slice(0, visibleLimit));
  let suggestedTargets = $derived(suggestionTargets(targets, selectedId, visibleLimit));
  let locationResults = $derived(visibleTargets.filter((target) => target.kind === 'location'));
  let containerResults = $derived(visibleTargets.filter((target) => target.kind === 'container'));
  let selectedTarget = $derived(targets.find((target) => target.id === selectedId) ?? null);
  let hasSearch = $derived(normalizedSearch.length > 0);
  let resultCountLabel = $derived(`${matchingTargets.length} ${matchingTargets.length === 1 ? 'match' : 'matches'}`);
  let destinationCountLabel = $derived(`${targets.length} possible ${targets.length === 1 ? 'destination' : 'destinations'}`);
  let suggestedCountLabel = $derived(
    `Showing ${suggestedTargets.length} suggested ${suggestedTargets.length === 1 ? 'destination' : 'destinations'}.`
  );

  function clearSelection(): void {
    onSelect(null);
  }

  function suggestionTargets(items: AssetViewModel[], currentId: string | null, limit: number): AssetViewModel[] {
    const sorted = [...items].sort(compareTargets);
    const chosen: AssetViewModel[] = [];
    for (const target of sorted) {
      if (chosen.length >= limit) {
        break;
      }
      if (target.id === currentId || chosen.some((candidate) => candidate.id === target.id)) {
        continue;
      }
      chosen.push(target);
    }
    return chosen;
  }

  function compareTargets(left: AssetViewModel, right: AssetViewModel): number {
    const kindRank = kindSortRank(left.kind) - kindSortRank(right.kind);
    if (kindRank !== 0) {
      return kindRank;
    }
    return left.title.localeCompare(right.title);
  }

  function compareTargetsForSearch(left: AssetViewModel, right: AssetViewModel, query: string): number {
    if (!query) {
      return compareTargets(left, right);
    }
    const kindRank = kindSortRank(left.kind) - kindSortRank(right.kind);
    if (kindRank !== 0) {
      return kindRank;
    }
    const relevanceRank = searchRank(left, query) - searchRank(right, query);
    if (relevanceRank !== 0) {
      return relevanceRank;
    }
    return left.title.localeCompare(right.title);
  }

  function searchRank(target: AssetViewModel, query: string): number {
    const title = target.title.toLowerCase();
    const trail = target.containmentTrail.toLowerCase();
    const trailSegments = trail.split('/').map((segment) => segment.trim());
    if (title === query) {
      return 0;
    }
    if (title.startsWith(query)) {
      return 1;
    }
    if (title.includes(query)) {
      return 2;
    }
    if (trailSegments.includes(query)) {
      return 3;
    }
    if (trail.includes(query)) {
      return 4;
    }
    return 5;
  }

  function kindSortRank(kind: AssetViewModel['kind']): number {
    if (kind === 'location') {
      return 0;
    }
    if (kind === 'container') {
      return 1;
    }
    return 2;
  }
</script>

<fieldset class="selection-field parent-selection">
  <legend>{legend}</legend>
  <div class="field-stack">
    <Label for={searchId}>{searchLabel}</Label>
    <Input id={searchId} bind:value={search} placeholder={searchPlaceholder} />
  </div>
  <div class="parent-current-shell">
    <div>
      <p class="selection-summary">Current destination</p>
      <div class="parent-current-card" data-selected={selectedTarget ? 'target' : 'root'}>
        <div class="parent-kind-mark" data-kind={selectedTarget ? 'target' : 'root'} aria-hidden="true">
          {#if selectedTarget}
            <AssetThumb asset={selectedTarget} size="sm" />
          {:else}
            <span></span>
          {/if}
        </div>
        <span>
          <strong>{selectedTarget?.title ?? rootLabel}</strong>
          <small>{selectedTarget ? `${assetKindLabel(selectedTarget.kind)} / ${selectedTarget.containmentTrail}` : rootSummaryLabel}</small>
        </span>
      </div>
    </div>
    {#if selectedTarget}
      <Button.Root
        type="button"
        variant="outline"
        size="sm"
        aria-label="Clear parent selection"
        onclick={clearSelection}
      >
        <X /> Clear parent
      </Button.Root>
    {/if}
  </div>
  <div class="parent-picker parent-current" role="group" aria-label={`${groupLabel} current destination`}>
    <Button.Root
      type="button"
      variant={selectedId === null ? 'secondary' : 'outline'}
      aria-pressed={selectedId === null}
      onclick={() => onSelect(null)}
    >
      {rootLabel}
    </Button.Root>
  </div>
  <p class="selection-summary" aria-live="polite" aria-atomic="true">{hasSearch ? resultCountLabel : ''}</p>
  {#if hasSearch}
    <div class="parent-picker parent-picker-results option-grid" role="group" aria-label={`${groupLabel} search results`}>
      {#if locationResults.length > 0}
        <div class="parent-result-group" role="group" aria-label="Locations" aria-labelledby={`${searchId}-location-results-label`}>
          <p id={`${searchId}-location-results-label`} class="parent-result-heading">Locations</p>
          {#each locationResults as target}
            <ParentTargetButton {target} selected={selectedId === target.id} onSelect={onSelect} />
          {/each}
        </div>
      {/if}
      {#if containerResults.length > 0}
        <div class="parent-result-group" role="group" aria-label="Containers" aria-labelledby={`${searchId}-container-results-label`}>
          <p id={`${searchId}-container-results-label`} class="parent-result-heading">Containers</p>
          {#each containerResults as target}
            <ParentTargetButton {target} selected={selectedId === target.id} onSelect={onSelect} />
          {/each}
        </div>
      {/if}
    </div>
    {#if visibleTargets.length === 0}
      <p class="muted-note">No matching locations or containers.</p>
    {:else if matchingTargets.length > visibleTargets.length}
      <p class="muted-note">Showing the first {visibleTargets.length} of {matchingTargets.length} matches.</p>
    {/if}
  {:else if targets.length > 0}
    <div class="parent-suggestion-header">
      <p class="selection-summary">Suggested destinations</p>
      <p class="muted-note">{destinationCountLabel}. {suggestedCountLabel}</p>
    </div>
    <div class="parent-picker parent-picker-results option-grid" role="group" aria-label={`${groupLabel} suggested destinations`}>
      {#each suggestedTargets as target}
        <ParentTargetButton {target} selected={selectedId === target.id} onSelect={onSelect} />
      {/each}
    </div>
  {:else}
    <p class="muted-note">No locations or containers yet.</p>
  {/if}
</fieldset>
