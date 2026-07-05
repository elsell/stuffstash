<script lang="ts">
  import * as Button from '$lib/components/ui/button/index.js';
  import type { Asset } from '$lib/domain/inventory';
  import { assetKindLabel } from '$lib/domain/inventory';
  import AssetThumb from './AssetThumb.svelte';

  let {
    id,
    idPrefix,
    suggestions,
    activeIndex,
    label = 'Search suggestions',
    assetHref,
    onFocusIndex,
    onSuggestionKeydown,
    onOpen
  }: {
    id: string;
    idPrefix: string;
    suggestions: Asset[];
    activeIndex: number;
    label?: string;
    assetHref: (asset: Asset) => string;
    onFocusIndex: (index: number) => void;
    onSuggestionKeydown: (event: KeyboardEvent, index: number) => void;
    onOpen: (event: MouseEvent, asset: Asset) => void;
  } = $props();

  function suggestionId(index: number): string {
    return `${idPrefix}-${index}`;
  }

  function photoUnavailableId(index: number): string {
    return `${suggestionId(index)}-photo-unavailable`;
  }
</script>

{#if suggestions.length > 0}
  <ul {id} class="search-suggestions" aria-label={label}>
    {#each suggestions as suggestion, index}
      <li>
        <Button.Root
          id={suggestionId(index)}
          href={assetHref(suggestion)}
          variant="ghost"
          class="suggestion-row"
          data-active={activeIndex === index}
          aria-label={`Open ${suggestion.title}`}
          aria-describedby={suggestion.photoUnavailable ? photoUnavailableId(index) : undefined}
          onfocus={() => onFocusIndex(index)}
          onkeydown={(event) => onSuggestionKeydown(event, index)}
          onpointerenter={() => onFocusIndex(index)}
          onclick={(event) => onOpen(event, suggestion)}
        >
          <AssetThumb asset={suggestion} size="sm" />
          <span>
            <strong>{suggestion.title}</strong>
            <small>{suggestion.customAssetTypeLabel ?? assetKindLabel(suggestion.kind)}</small>
            {#if suggestion.photoUnavailable}
              <small id={photoUnavailableId(index)} class="visually-hidden">Photo unavailable</small>
            {/if}
          </span>
        </Button.Root>
      </li>
    {/each}
  </ul>
{/if}
