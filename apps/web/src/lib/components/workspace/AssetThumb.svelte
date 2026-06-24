<script lang="ts">
  import type { Asset } from '$lib/domain/inventory';
  import KindIcon from './KindIcon.svelte';

  let { asset, size = 'md' }: { asset: Asset; size?: 'sm' | 'md' | 'lg' } = $props();
</script>

<div class="asset-thumb asset-thumb-{size}">
  {#if asset.photo}
    <img src={asset.photo.url} alt={asset.photo.alt} />
  {:else}
    <KindIcon kind={asset.kind} />
  {/if}
</div>
