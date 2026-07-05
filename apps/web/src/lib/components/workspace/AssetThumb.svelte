<script lang="ts">
  import ImageOff from '@lucide/svelte/icons/image-off';
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
    {#if asset.photoUnavailable}
      <span class="photo-unavailable-mark" aria-hidden="true" title="Photo unavailable">
        <ImageOff aria-hidden="true" />
      </span>
      <span class="visually-hidden">Photo unavailable</span>
    {/if}
  {/if}
</div>
