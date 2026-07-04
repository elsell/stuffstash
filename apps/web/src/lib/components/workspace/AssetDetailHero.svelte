<script lang="ts" module>
  export const PHOTO_UPLOAD_DISABLED_REASON_ID = 'asset-photo-upload-disabled';
  export const PHOTO_UPLOAD_ERROR_ID = 'asset-photo-upload-error';

  export type DetailPhoto = {
    id: string;
    url: string;
    alt: string;
    fileName: string;
    sizeBytes?: number;
    isPrimary: boolean;
  };
</script>

<script lang="ts">
  import Image from '@lucide/svelte/icons/image';
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
    children,
    onChoosePhoto,
    onSelectPhoto
  }: {
    kind: AssetKind;
    heroPhoto: DetailPhoto | undefined;
    photos: DetailPhoto[];
    canAddPhoto: boolean;
    uploadDisabledReason: string;
    uploadError: string;
    children?: Snippet;
    onChoosePhoto: () => void;
    onSelectPhoto: (photoId: string) => void;
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
  </div>
  {@render children?.()}
  <div class="photo-gallery-section" aria-label="Asset photo gallery">
    <div class="photo-panel-actions">
      <Button.Root
        variant="outline"
        disabled={!canAddPhoto}
        aria-describedby={uploadDescribedBy || undefined}
        onclick={onChoosePhoto}
      >
        <Image /> Add photo
      </Button.Root>
    </div>
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
    {:else}
      <div class="empty-state compact-empty">
        <p>No photos yet.</p>
      </div>
    {/if}
    {#if uploadError}
      <p id={PHOTO_UPLOAD_ERROR_ID} class="denied-note" role="alert">{uploadError}</p>
    {/if}
  </div>
</div>
