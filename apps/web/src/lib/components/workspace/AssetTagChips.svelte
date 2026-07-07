<script lang="ts">
  import type { AssetTag } from '$lib/domain/inventory';

  let { tags = [], compact = false, overflowLimit }: { tags?: AssetTag[]; compact?: boolean; overflowLimit?: number } = $props();
  let visibleLimit = $derived(overflowLimit ?? tags.length);
  let visibleTags = $derived(tags.slice(0, visibleLimit));
  let hiddenCount = $derived(Math.max(0, tags.length - visibleTags.length));
</script>

{#if tags.length > 0}
  <span
    class="tag-chip-list"
    data-compact={compact ? 'true' : undefined}
    data-overflow={overflowLimit !== undefined ? 'true' : undefined}
    aria-label="Asset tags"
  >
    {#each visibleTags as tag}
      <span class="tag-chip">
        {#if tag.color}
          <span class="tag-swatch" style={`--tag-color: ${tag.color}`} aria-hidden="true"></span>
        {/if}
        <span>{tag.displayName}</span>
      </span>
    {/each}
    {#if hiddenCount > 0}
      <span class="tag-chip tag-chip-overflow" aria-label={`${hiddenCount} more tags`}>+{hiddenCount}</span>
    {/if}
  </span>
{/if}
