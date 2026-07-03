<script lang="ts">
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import type { AssetViewModel } from '$lib/domain/inventory';

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
  let visibleTargets = $derived(matchingTargets.slice(0, visibleLimit));
  let selectedTarget = $derived(targets.find((target) => target.id === selectedId) ?? null);
  let hasSearch = $derived(normalizedSearch.length > 0);
</script>

<fieldset class="selection-field parent-selection">
  <legend>{legend}</legend>
  <div class="field-stack">
    <Label for={searchId}>{searchLabel}</Label>
    <Input id={searchId} bind:value={search} placeholder={searchPlaceholder} />
  </div>
  <p class="selection-summary">
    {selectedTarget ? `Selected ${selectedTarget.title}` : `Selected ${rootSummaryLabel}`}
  </p>
  <div class="parent-picker parent-current" role="group" aria-label={`${groupLabel} current destination`}>
    <Button.Root
      type="button"
      variant={selectedId === null ? 'secondary' : 'outline'}
      aria-pressed={selectedId === null}
      onclick={() => onSelect(null)}
    >
      {rootLabel}
    </Button.Root>
    {#if selectedTarget}
      <Button.Root
        type="button"
        variant="secondary"
        class="parent-target-button"
        aria-pressed="true"
        onclick={() => onSelect(selectedTarget.id)}
      >
        <span>{selectedTarget.title}</span>
        <small>{selectedTarget.containmentTrail}</small>
      </Button.Root>
    {/if}
  </div>
  {#if hasSearch}
    <div class="parent-picker parent-picker-results option-grid" role="group" aria-label={`${groupLabel} search results`}>
      {#each visibleTargets as target}
        <Button.Root
          type="button"
          variant={selectedId === target.id ? 'secondary' : 'outline'}
          class="parent-target-button"
          aria-pressed={selectedId === target.id}
          onclick={() => onSelect(target.id)}
        >
          <span>{target.title}</span>
          <small>{target.containmentTrail}</small>
        </Button.Root>
      {/each}
    </div>
    {#if visibleTargets.length === 0}
      <p class="muted-note">No matching locations or containers.</p>
    {:else if matchingTargets.length > visibleTargets.length}
      <p class="muted-note">Showing the first {visibleTargets.length} of {matchingTargets.length} matches.</p>
    {/if}
  {:else if targets.length > 0}
    <p class="muted-note">Search {targets.length} available locations and containers.</p>
  {:else}
    <p class="muted-note">No locations or containers yet.</p>
  {/if}
</fieldset>
