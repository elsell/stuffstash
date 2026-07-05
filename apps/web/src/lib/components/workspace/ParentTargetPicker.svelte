<script lang="ts">
  import { tick } from 'svelte';
  import X from '@lucide/svelte/icons/x';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import {
    normalizeParentTargetQuery,
    parentTargetMetadataLabel,
    parentTargetPickerPresentation,
    parentTargetSuggestions,
    searchParentTargets
  } from '$lib/application/workspaceParentTargets';
  import type { ParentTargetViewModel } from '$lib/domain/inventory';
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
    visibleLimit = 4,
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
    targets: ParentTargetViewModel[];
    visibleLimit?: number;
    onSelect: (id: string | null) => void;
  } = $props();

  let searchExpanded = $state(false);
  let previousNormalizedSearch = $state('');
  let normalizedSearch = $derived(normalizeParentTargetQuery(search));
  let effectiveVisibleLimit = $derived(normalizedSearch && searchExpanded ? Math.max(visibleLimit, targets.length) : visibleLimit);
  let searchResult = $derived(searchParentTargets(targets, search, effectiveVisibleLimit));
  let matchingTargets = $derived(searchResult.matchingTargets);
  let visibleTargets = $derived(searchResult.visibleTargets);
  let suggestedTargets = $derived(parentTargetSuggestions(targets, selectedId, visibleLimit));
  let locationResults = $derived(searchResult.locationResults);
  let containerResults = $derived(searchResult.containerResults);
  let selectedTarget = $derived(targets.find((target) => target.id === selectedId) ?? null);
  let selectedTargetMetadataLabel = $derived(selectedTarget ? parentTargetMetadataLabel(selectedTarget) : rootSummaryLabel);
  let selectedDestinationName = $derived(selectedTarget?.title ?? rootLabel);
  let selectedDestinationAnnouncement = $derived(`Current destination: ${selectedDestinationName}, ${selectedTargetMetadataLabel}`);
  let hasSearch = $derived(normalizedSearch.length > 0);
  let presentation = $derived(
    parentTargetPickerPresentation({
      hasSearch,
      matchingCount: matchingTargets.length,
      visibleCount: visibleTargets.length,
      targetCount: targets.length,
      suggestedCount: suggestedTargets.length
    })
  );
  let canExpandSearch = $derived(hasSearch && matchingTargets.length > visibleTargets.length);

  $effect(() => {
    if (normalizedSearch === previousNormalizedSearch) {
      return;
    }
    previousNormalizedSearch = normalizedSearch;
    searchExpanded = false;
  });

  function clearSelection(): void {
    onSelect(null);
  }

  function expandSearchResults(): void {
    const firstNewResultIndex = visibleTargets.length;
    searchExpanded = true;
    void tick().then(() => {
      resultButtonElements()[firstNewResultIndex]?.focus();
    });
  }

  function resultButtonElements(): HTMLButtonElement[] {
    if (typeof document === 'undefined') {
      return [];
    }
    return Array.from(document.querySelectorAll<HTMLButtonElement>(`#${searchId}-results .parent-target-button`));
  }

</script>

<fieldset class="selection-field parent-selection">
  <legend>{legend}</legend>
  <div class="field-stack">
    <Label for={searchId}>{searchLabel}</Label>
    <Input id={searchId} bind:value={search} placeholder={searchPlaceholder} />
  </div>
  <div class="parent-current-shell" role="group" aria-label={`${groupLabel} current destination`}>
    <div>
      <p class="selection-summary">Current destination</p>
      <div
        class="parent-current-card"
        data-selected={selectedTarget ? 'target' : 'root'}
        aria-label={selectedDestinationAnnouncement}
      >
        <div class="parent-kind-mark" data-kind={selectedTarget ? 'target' : 'root'} aria-hidden="true">
          {#if selectedTarget}
            <AssetThumb asset={selectedTarget} size="sm" />
          {:else}
            <span></span>
          {/if}
        </div>
        <span>
          <strong>{selectedDestinationName}</strong>
          <small>{selectedTargetMetadataLabel}</small>
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
  {#if selectedTarget}
    <div class="parent-picker parent-current" role="group" aria-label={`${groupLabel} root destination`}>
      <Button.Root
        type="button"
        variant="outline"
        aria-pressed="false"
        onclick={() => onSelect(null)}
      >
        {rootLabel}
      </Button.Root>
    </div>
  {/if}
  <p class="selection-summary" aria-live="polite" aria-atomic="true">{presentation.resultCountLabel}</p>
  {#if hasSearch}
    <div id={`${searchId}-results`} class="parent-picker parent-picker-results option-grid" role="group" aria-label={`${groupLabel} search results`}>
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
    {#if presentation.status.kind !== 'none'}
      <p class="muted-note">{presentation.status.message}</p>
      {#if canExpandSearch}
        <Button.Root
          type="button"
          variant="outline"
          size="sm"
          class="parent-show-more"
          aria-label={`Show all ${matchingTargets.length} matching parent destinations`}
          onclick={expandSearchResults}
        >
          Show all {matchingTargets.length} matches
        </Button.Root>
      {/if}
    {/if}
  {:else if targets.length > 0}
    <div class="parent-suggestion-header">
      <p class="selection-summary">Suggested destinations</p>
      <p class="muted-note">{presentation.destinationCountLabel}. {presentation.suggestedCountLabel}</p>
    </div>
    <div class="parent-picker parent-picker-results option-grid" role="group" aria-label={`${groupLabel} suggested destinations`}>
      {#each suggestedTargets as target}
        <ParentTargetButton {target} selected={selectedId === target.id} onSelect={onSelect} />
      {/each}
    </div>
  {:else}
    <p class="muted-note">{presentation.status.message}</p>
  {/if}
</fieldset>
