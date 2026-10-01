<script lang="ts" module>
  import { t } from '$lib/presentation/localization';
  import type { DetailPhoto } from '$lib/application/workspaceAssetMedia';

  export const PHOTO_UPLOAD_DISABLED_REASON_ID = 'asset-photo-upload-disabled';
  export const PHOTO_UPLOAD_ERROR_ID = 'asset-photo-upload-error';
</script>

<script lang="ts">
  import {visibleImage} from '$lib/presentation/observability/visibleImage';
  import {imageObserverContext} from '$lib/presentation/observability/imageObserverContext';
  const imageObserver = imageObserverContext();

  import Image from '@lucide/svelte/icons/image';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import type { Snippet } from 'svelte';
  import * as Button from '$lib/components/ui/button/index.js';
  import type { AssetKind } from '$lib/domain/inventory';
  import KindIcon from './KindIcon.svelte';

  let {
    kind,
    heroPhoto,
    photos,
    canAddPhoto,
    uploadDisabledReason,
    uploadError,
    uploadBusy = false,
    retryPhotoName = '',
    removePhotoHref = '',
    children,
    onChoosePhoto,
    onSelectPhoto,
    onRetryPhoto = () => {},
    onRemovePhoto = () => {}
  }: {
    kind: AssetKind;
    heroPhoto: DetailPhoto | undefined;
    photos: DetailPhoto[];
    canAddPhoto: boolean;
    uploadDisabledReason: string;
    uploadError: string;
    uploadBusy?: boolean;
    retryPhotoName?: string;
    removePhotoHref?: string;
    children?: Snippet;
    onChoosePhoto: () => void;
    onSelectPhoto: (photoId: string) => void;
    onRetryPhoto?: () => void;
    onRemovePhoto?: (event: MouseEvent) => void;
  } = $props();
  let uploadDescribedBy = $derived(
    [uploadDisabledReason ? PHOTO_UPLOAD_DISABLED_REASON_ID : '', uploadError ? PHOTO_UPLOAD_ERROR_ID : ''].filter(Boolean).join(' ')
  );
</script>

<div class="asset-detail-hero">
  <div class="asset-photo-panel" aria-label={t('web.AssetDetailHero.assetPhotos')}>
    <div class="asset-hero-photo">
      {#if heroPhoto}
        <img use:visibleImage={{source: heroPhoto.url, observer: imageObserver, surface: 'detail', variant: heroPhoto.variant ?? 'none'}} src={heroPhoto.url} alt={heroPhoto.alt} />
      {:else}
        <div class="asset-hero-fallback">
          <KindIcon kind={kind} />
        </div>
      {/if}
    </div>
    <div class="photo-panel-actions">
      <Button.Root
        variant="outline"
        disabled={!canAddPhoto}
        aria-describedby={uploadDescribedBy || undefined}
        onclick={onChoosePhoto}
      >
        <Image /> {t('web.AssetDetailHero.addPhoto')} </Button.Root>
      {#if heroPhoto && removePhotoHref}
        <Button.Root
          href={removePhotoHref}
          variant="outline"
          aria-label={t('web.AssetDetailHero.removePhoto2', { fileName: String(heroPhoto.fileName) })}
          title={t('web.AssetDetailHero.remove', { fileName: String(heroPhoto.fileName) })}
          onclick={onRemovePhoto}
        ><Trash2 /> {t('web.AssetDetailHero.removePhoto')}</Button.Root>
      {/if}
    </div>
  </div>
  {@render children?.()}
  {#if photos.length > 0 || uploadDisabledReason || uploadError || uploadBusy || retryPhotoName}
  <div class="photo-gallery-section" aria-label={t('web.AssetDetailHero.assetPhotoGallery')}>
    {#if uploadDisabledReason}
      <p id={PHOTO_UPLOAD_DISABLED_REASON_ID} class="denied-note" role="note">{uploadDisabledReason}</p>
    {/if}
    {#if photos.length > 0}
      <div class="photo-rail" aria-label={t('web.AssetDetailHero.photos')}>
        {#each photos as photo}
          <Button.Root
            variant="ghost"
            class={photo.id === heroPhoto?.id ? 'active' : ''}
            aria-label={t('web.AssetDetailHero.show', { fileName: String(photo.fileName) })}
            aria-pressed={photo.id === heroPhoto?.id}
            onclick={() => onSelectPhoto(photo.id)}
          >
            <img use:visibleImage={{source: photo.url, observer: imageObserver, surface: 'gallery', variant: photo.variant ?? 'none'}} src={photo.url} alt="" />
            {#if photo.isPrimary}
              <span>{t('web.AssetDetailHero.primary')}</span>
            {/if}
          </Button.Root>
        {/each}
      </div>
    {/if}
    {#if uploadError}
      <p id={PHOTO_UPLOAD_ERROR_ID} class="denied-note" role="alert">{uploadError}</p>
    {/if}
    {#if uploadBusy}
      <p class="photo-upload-status" role="status">{t('web.AssetDetailHero.uploadingPhoto')}</p>
    {:else if retryPhotoName}
      <Button.Root variant="outline" aria-label={t('web.AssetDetailHero.retry2', { retryPhotoName: String(retryPhotoName) })} onclick={onRetryPhoto}>{t('web.AssetDetailHero.retryFull', { retryPhotoName: retryPhotoName })}</Button.Root>
    {/if}
  </div>
  {/if}
</div>
