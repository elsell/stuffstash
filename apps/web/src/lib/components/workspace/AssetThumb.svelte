<script lang="ts">
  import type { Asset } from '$lib/domain/inventory';
  import KindIcon from './KindIcon.svelte';

  let { asset, size = 'md' }: { asset: Asset; size?: 'sm' | 'md' | 'lg' } = $props();
  let ownPhoto = $derived(asset.photo?.assetId === asset.id ? asset.photo : undefined);
</script>

<div class="asset-thumb asset-thumb-{size}">
  {#if ownPhoto}
    <img src={ownPhoto.url} alt={ownPhoto.alt} />
  {:else}
    <KindIcon kind={asset.kind} />
  {/if}
</div>
