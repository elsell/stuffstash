<script lang="ts" module>
  import type { MediaUploadPolicy, SelectedPhoto } from '$lib/domain/inventory';

  export type AddAssetPhotosSectionProps = {
    photos: SelectedPhoto[];
    summary: string;
    mediaPolicy: MediaUploadPolicy;
    inputKey: number;
    error: string;
    onFiles: (files: FileList | undefined) => void;
    onRemove: (photoId: string) => void;
  };
</script>

<script lang="ts">
  import Camera from '@lucide/svelte/icons/camera';
  import Upload from '@lucide/svelte/icons/upload';
  import X from '@lucide/svelte/icons/x';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { formatBytes } from './formatBytes';

  let {
    photos,
    summary,
    mediaPolicy,
    inputKey,
    error,
    onFiles,
    onRemove
  }: AddAssetPhotosSectionProps = $props();

  let fileInput = $state<HTMLInputElement | null>(null);
  let cameraInput = $state<HTMLInputElement | null>(null);
  let supportedImageTypes = $derived(mediaPolicy.supportedContentTypes.filter((type) => type.startsWith('image/')));
  let acceptTypes = $derived(supportedImageTypes.join(','));
  let supportedTypeLabel = $derived(formatSupportedTypes(supportedImageTypes));

  function openPhotoPicker(): void {
    fileInput?.click();
  }

  function openCameraPicker(): void {
    cameraInput?.click();
  }

  function formatSupportedTypes(types: string[]): string {
    if (types.length === 0) {
      return 'No image formats';
    }
    const labels = types.map(formatContentType);
    if (labels.length === 1) {
      return labels[0] ?? '';
    }
    if (labels.length === 2) {
      return `${labels[0]} or ${labels[1]}`;
    }
    return `${labels.slice(0, -1).join(', ')}, or ${labels[labels.length - 1]}`;
  }

  function formatContentType(type: string): string {
    if (type === 'image/jpeg') return 'JPEG';
    if (type === 'image/png') return 'PNG';
    if (type === 'image/webp') return 'WebP';
    return type.replace(/^image\//, '').toUpperCase();
  }
</script>

<fieldset class="selection-field attachment-section" aria-describedby="photo-help">
  <legend>Photos</legend>
  <p id="photo-help" class="selection-summary">Optional {supportedTypeLabel} up to {formatBytes(mediaPolicy.maxBytes)}.</p>
  <div class="photo-actions" role="group" aria-label="Photo actions">
    <Button.Root type="button" variant="outline" class="photo-label" onclick={openPhotoPicker}><Upload /> Upload</Button.Root>
    <Button.Root type="button" variant="outline" class="photo-label" onclick={openCameraPicker}><Camera /> Camera</Button.Root>
    <span class="photo-status" aria-live="polite">{summary}</span>
  </div>
  {#key inputKey}
    <Input
      id="asset-photos"
      bind:ref={fileInput}
      class="visually-hidden"
      type="file"
      tabindex={-1}
      accept={acceptTypes}
      multiple
      onchange={(event) => onFiles(event.currentTarget.files ?? undefined)}
    />
    <Input
      id="asset-camera"
      bind:ref={cameraInput}
      class="visually-hidden"
      type="file"
      tabindex={-1}
      accept={acceptTypes}
      capture="environment"
      onchange={(event) => onFiles(event.currentTarget.files ?? undefined)}
    />
  {/key}
</fieldset>

{#if photos.length > 0}
  <div class="photo-preview-list">
    {#each photos as photo}
      <div class="photo-preview">
        <img src={photo.previewUrl} alt={photo.name} />
        <span>{photo.name}</span>
        <Button.Root variant="ghost" size="icon-xs" aria-label={`Remove ${photo.name}`} onclick={() => onRemove(photo.id)}><X /></Button.Root>
      </div>
    {/each}
  </div>
{/if}
{#if error}
  <p class="denied-note" role="alert">{error}</p>
{/if}
