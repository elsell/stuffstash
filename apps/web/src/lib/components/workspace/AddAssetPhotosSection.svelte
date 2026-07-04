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
  import {
    addPhotoAcceptTypes,
    addPhotoHelpText,
    addPhotoPickerPresentation,
    addPhotoRemoveLabel,
    addPhotoSupportedTypeLabel,
    addSupportedImageTypes
  } from '$lib/application/workspaceAddPresentation';
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
  let supportedImageTypes = $derived(addSupportedImageTypes(mediaPolicy));
  let acceptTypes = $derived(addPhotoAcceptTypes(supportedImageTypes));
  let supportedTypeLabel = $derived(addPhotoSupportedTypeLabel(supportedImageTypes));
  let helpText = $derived(addPhotoHelpText(supportedTypeLabel, formatBytes(mediaPolicy.maxBytes)));
  let describedBy = $derived(['photo-help', 'photo-status', error ? 'photo-error' : ''].filter(Boolean).join(' '));

  function openPhotoPicker(): void {
    fileInput?.click();
  }

  function openCameraPicker(): void {
    cameraInput?.click();
  }
</script>

<fieldset class="selection-field attachment-section" aria-describedby={describedBy}>
  <legend>Photos</legend>
  <p id="photo-help" class="selection-summary">{helpText}</p>
  <div class="photo-actions" role="group" aria-label={addPhotoPickerPresentation.actionGroupLabel} aria-describedby={describedBy}>
    <Button.Root type="button" variant="outline" class="photo-label" aria-describedby={describedBy} onclick={openPhotoPicker}>
      <Upload />
      {addPhotoPickerPresentation.uploadLabel}
    </Button.Root>
    <Button.Root type="button" variant="outline" class="photo-label" aria-describedby={describedBy} onclick={openCameraPicker}>
      <Camera />
      {addPhotoPickerPresentation.cameraLabel}
    </Button.Root>
    <span id="photo-status" class="photo-status" aria-live="polite">{summary}</span>
  </div>
  {#key inputKey}
    <Input
      id="asset-photos"
      bind:ref={fileInput}
      class="visually-hidden"
      type="file"
      tabindex={-1}
      aria-label={addPhotoPickerPresentation.uploadInputLabel}
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
      aria-label={addPhotoPickerPresentation.cameraInputLabel}
      accept={acceptTypes}
      capture="environment"
      onchange={(event) => onFiles(event.currentTarget.files ?? undefined)}
    />
  {/key}
</fieldset>

{#if photos.length > 0}
  <div class="photo-preview-list" role="list" aria-label={addPhotoPickerPresentation.selectedListLabel}>
    {#each photos as photo}
      <div class="photo-preview" role="listitem">
        <img src={photo.previewUrl} alt={photo.name} />
        <span>{photo.name}</span>
        <Button.Root variant="ghost" size="icon-xs" aria-label={addPhotoRemoveLabel(photo)} onclick={() => onRemove(photo.id)}><X /></Button.Root>
      </div>
    {/each}
  </div>
{/if}
{#if error}
  <p id="photo-error" class="denied-note" role="alert">{error}</p>
{/if}
