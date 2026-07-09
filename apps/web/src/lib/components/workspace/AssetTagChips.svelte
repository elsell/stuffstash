<script lang="ts">
  import * as Button from '$lib/components/ui/button/index.js';
  import type { AssetTag } from '$lib/domain/inventory';

  let {
    tags = [],
    compact = false,
    overflowLimit,
    selectedTagIds = [],
    onTagSelect
  }: {
    tags?: AssetTag[];
    compact?: boolean;
    overflowLimit?: number;
    selectedTagIds?: string[];
    onTagSelect?: (tag: AssetTag) => void;
  } = $props();
  let visibleLimit = $derived(overflowLimit ?? tags.length);
  let visibleTags = $derived(tags.slice(0, visibleLimit));
  let hiddenCount = $derived(Math.max(0, tags.length - visibleTags.length));
  let selectedTagIdSet = $derived(new Set(selectedTagIds));

  function selectTag(event: MouseEvent, tag: AssetTag): void {
    event.preventDefault();
    event.stopPropagation();
    onTagSelect?.(tag);
  }

</script>

{#if tags.length > 0}
  <span
    class="tag-chip-list"
    data-compact={compact ? 'true' : undefined}
    data-overflow={overflowLimit !== undefined ? 'true' : undefined}
    aria-label="Asset tags"
  >
    {#each visibleTags as tag}
      {@const selected = selectedTagIdSet.has(tag.id)}
      {#if onTagSelect}
        <Button.Root
          type="button"
          variant="ghost"
          class={`tag-chip tag-chip-action${tag.color ? ' tag-chip-colored' : ''}`}
          style={tag.color ? `--tag-color: ${tag.color}` : undefined}
          aria-label={`Search for tag ${tag.displayName}`}
          aria-pressed={selected}
          data-selected={selected ? 'true' : undefined}
          onclick={(event) => selectTag(event, tag)}
        >
          <span>{tag.displayName}</span>
        </Button.Root>
      {:else}
        <span class={`tag-chip${tag.color ? ' tag-chip-colored' : ''}`} style={tag.color ? `--tag-color: ${tag.color}` : undefined}>
          <span>{tag.displayName}</span>
        </span>
      {/if}
    {/each}
    {#if hiddenCount > 0}
      <span class="tag-chip tag-chip-overflow" aria-label={`${hiddenCount} more tags`}>+{hiddenCount}</span>
    {/if}
  </span>
{/if}
