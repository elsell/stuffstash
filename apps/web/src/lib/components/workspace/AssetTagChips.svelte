<script lang="ts">
  import type { AssetTag } from '$lib/domain/inventory';

  let { tags = [], compact = false }: { tags?: AssetTag[]; compact?: boolean } = $props();
</script>

{#if tags.length > 0}
  <span class="tag-chip-list" data-compact={compact ? 'true' : undefined} aria-label="Asset tags">
    {#each tags as tag}
      <span class="tag-chip">
        {#if tag.color}
          <span class="tag-swatch" style={`--tag-color: ${tag.color}`} aria-hidden="true"></span>
        {/if}
        <span>{tag.displayName}</span>
      </span>
    {/each}
  </span>
{/if}
