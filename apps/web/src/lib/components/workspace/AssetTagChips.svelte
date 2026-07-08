<script lang="ts">
  import * as Button from '$lib/components/ui/button/index.js';
  import type { AssetTag } from '$lib/domain/inventory';

  let {
    tags = [],
    compact = false,
    overflowLimit,
    onTagSelect,
    actionMode = 'button'
  }: {
    tags?: AssetTag[];
    compact?: boolean;
    overflowLimit?: number;
    onTagSelect?: (tag: AssetTag) => void;
    actionMode?: 'button' | 'inline';
  } = $props();
  let visibleLimit = $derived(overflowLimit ?? tags.length);
  let visibleTags = $derived(tags.slice(0, visibleLimit));
  let hiddenCount = $derived(Math.max(0, tags.length - visibleTags.length));

  function selectTag(event: MouseEvent, tag: AssetTag): void {
    event.preventDefault();
    event.stopPropagation();
    onTagSelect?.(tag);
  }

  function selectTagWithKeyboard(event: KeyboardEvent, tag: AssetTag): void {
    if (event.key !== 'Enter' && event.key !== ' ') {
      return;
    }
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
      {#if onTagSelect}
        {#if actionMode === 'inline'}
          <span
            role="button"
            tabindex="0"
            class={`tag-chip tag-chip-action${tag.color ? ' tag-chip-colored' : ''}`}
            style={tag.color ? `--tag-color: ${tag.color}` : undefined}
            aria-label={`Search for tag ${tag.displayName}`}
            onclick={(event) => selectTag(event, tag)}
            onkeydown={(event) => selectTagWithKeyboard(event, tag)}
          >
            <span>{tag.displayName}</span>
          </span>
        {:else}
          <Button.Root
            type="button"
            variant="ghost"
            class={`tag-chip tag-chip-action${tag.color ? ' tag-chip-colored' : ''}`}
            style={tag.color ? `--tag-color: ${tag.color}` : undefined}
            aria-label={`Search for tag ${tag.displayName}`}
            onclick={(event) => selectTag(event, tag)}
          >
            <span>{tag.displayName}</span>
          </Button.Root>
        {/if}
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
