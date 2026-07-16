<script lang="ts" module>
  import type { DetailPhoto } from '$lib/application/workspaceAssetMedia';

  export const PHOTO_UPLOAD_DISABLED_REASON_ID = 'asset-photo-upload-disabled';
  export const PHOTO_UPLOAD_ERROR_ID = 'asset-photo-upload-error';
</script>

<script lang="ts">
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
    removePhotoButton = $bindable(null),
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
    removePhotoButton?: HTMLElement | null;
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
  <div class="asset-photo-panel" aria-label="Asset photos">
    <div class="asset-hero-photo">
      {#if heroPhoto}
        <img src={heroPhoto.url} alt={heroPhoto.alt} />
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
        <Image /> Add photo
      </Button.Root>
      {#if heroPhoto && removePhotoHref}
        <Button.Root
          bind:ref={removePhotoButton}
          href={removePhotoHref}
          variant="outline"
          aria-label={`Remove photo ${heroPhoto.fileName}`}
          title={`Remove ${heroPhoto.fileName}`}
          onclick={onRemovePhoto}
        ><Trash2 /> Remove photo</Button.Root>
      {/if}
    </div>
  </div>
  {@render children?.()}
  {#if photos.length > 0 || uploadDisabledReason || uploadError || uploadBusy || retryPhotoName}
  <div class="photo-gallery-section" aria-label="Asset photo gallery">
    {#if uploadDisabledReason}
      <p id={PHOTO_UPLOAD_DISABLED_REASON_ID} class="denied-note" role="note">{uploadDisabledReason}</p>
    {/if}
    {#if photos.length > 0}
      <div class="photo-rail" aria-label="Photos">
        {#each photos as photo}
          <Button.Root
            variant="ghost"
            class={photo.id === heroPhoto?.id ? 'active' : ''}
            aria-label={`Show ${photo.fileName}`}
            aria-pressed={photo.id === heroPhoto?.id}
            onclick={() => onSelectPhoto(photo.id)}
          >
            <img src={photo.url} alt="" />
            {#if photo.isPrimary}
              <span>Primary</span>
            {/if}
          </Button.Root>
        {/each}
      </div>
    {/if}
    {#if uploadError}
      <p id={PHOTO_UPLOAD_ERROR_ID} class="denied-note" role="alert">{uploadError}</p>
    {/if}
    {#if uploadBusy}
      <p class="photo-upload-status" role="status">Uploading photo…</p>
    {:else if retryPhotoName}
      <Button.Root variant="outline" aria-label={`Retry ${retryPhotoName}`} onclick={onRetryPhoto}>Retry {retryPhotoName}</Button.Root>
    {/if}
  </div>
  {/if}
</div>
