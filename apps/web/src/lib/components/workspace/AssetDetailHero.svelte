<script lang="ts" module>
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
    uploadError,
    children,
    onChoosePhoto,
    onSelectPhoto
  }: {
    kind: AssetKind;
    heroPhoto: DetailPhoto | undefined;
    photos: DetailPhoto[];
    canAddPhoto: boolean;
    uploadError: string;
    children?: Snippet;
    onChoosePhoto: () => void;
    onSelectPhoto: (photoId: string) => void;
  } = $props();
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
      <Button.Root variant="outline" disabled={!canAddPhoto} onclick={onChoosePhoto}>
        <Image /> Add photo
      </Button.Root>
    </div>
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
      <p class="denied-note" role="alert">{uploadError}</p>
    {/if}
  </div>
</div>
