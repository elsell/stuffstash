<script lang="ts">
  import {visibleImage} from '$lib/presentation/observability/visibleImage';
  import {imageObserverContext} from '$lib/presentation/observability/imageObserverContext';
  const imageObserver = imageObserverContext();

  import { t } from '$lib/presentation/localization';
  import ImageOff from '@lucide/svelte/icons/image-off';
  import { getContext, hasContext } from 'svelte';
  import type { Asset } from '$lib/domain/inventory';
  import { assetThumbnailLoaderContext, type AssetThumbnailLoader } from '$lib/ports/assetThumbnailLoader';
  import KindIcon from './KindIcon.svelte';

  let { asset, size = 'md', surface = 'list' }: { asset: Asset; size?: 'sm' | 'md' | 'lg'; surface?: 'home' | 'list' | 'detail' } = $props();
  const thumbnailLoader = hasContext(assetThumbnailLoaderContext)
    ? getContext<AssetThumbnailLoader>(assetThumbnailLoaderContext)
    : null;
  let ownPhoto = $derived(asset.photo?.assetId === asset.id ? asset.photo : undefined);
  let thumbnailRequest = $derived(
    !ownPhoto && asset.primaryPhotoId && thumbnailLoader
      ? thumbnailLoader.loadAssetThumbnail(asset)
      : null
  );
</script>

{#snippet fallback(unavailable: boolean)}
  <KindIcon kind={asset.kind} />
  {#if unavailable}
    <span class="photo-unavailable-mark" aria-hidden="true" title={t('web.AssetThumb.photoUnavailable')}>
      <ImageOff aria-hidden="true" />
    </span>
    <span class="visually-hidden">{t('web.AssetThumb.photoUnavailable')}</span>
  {/if}
{/snippet}

<div class="asset-thumb asset-thumb-{size}">
  {#if ownPhoto}
    <img use:visibleImage={{source: ownPhoto.url, observer: imageObserver, surface, variant: ownPhoto.variant ?? 'none'}} src={ownPhoto.url} alt={ownPhoto.alt} />
  {:else if thumbnailRequest}
    {#await thumbnailRequest}
      {@render fallback(false)}
    {:then loadedPhoto}
      {#if loadedPhoto?.assetId === asset.id}
        <img use:visibleImage={{source: loadedPhoto.url, observer: imageObserver, surface, variant: loadedPhoto.variant ?? 'none'}} src={loadedPhoto.url} alt={loadedPhoto.alt} />
      {:else}
        {@render fallback(true)}
      {/if}
    {:catch}
      {@render fallback(true)}
    {/await}
  {:else}
    {@render fallback(Boolean(asset.photoUnavailable))}
  {/if}
</div>
