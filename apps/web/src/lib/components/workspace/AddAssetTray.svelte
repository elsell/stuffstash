<script lang="ts">
  import Camera from '@lucide/svelte/icons/camera';
  import Upload from '@lucide/svelte/icons/upload';
  import X from '@lucide/svelte/icons/x';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { Textarea } from '$lib/components/ui/textarea/index.js';
  import type { AddAssetDraft, AssetKind, AssetViewModel, MediaUploadPolicy, SelectedPhoto } from '$lib/domain/inventory';
  import { assetKindLabel, assetKinds } from '$lib/domain/inventory';

  let {
    open,
    parentTargets,
    mediaPolicy,
    saving,
    onClose,
    onSave
  }: {
    open: boolean;
    parentTargets: AssetViewModel[];
    mediaPolicy: MediaUploadPolicy;
    saving: boolean;
    onClose: () => void;
    onSave: (draft: AddAssetDraft) => Promise<boolean>;
  } = $props();

  let kind = $state<AssetKind>('item');
  let title = $state('');
  let description = $state('');
  let parentAssetId = $state('');
  let selectedPhotos = $state<SelectedPhoto[]>([]);
  let photoError = $state('');
  let fileInputKey = $state(0);

  async function save(): Promise<void> {
    if (!title.trim() || photoError) {
      return;
    }
    const saved = await onSave({
      kind,
      title: title.trim(),
      description: description.trim(),
      parentAssetId: parentAssetId || null,
      photos: selectedPhotos
    });
    if (!saved) {
      return;
    }
    title = '';
    description = '';
    parentAssetId = '';
    selectedPhotos = [];
    photoError = '';
  }

  function captureFiles(files: FileList | undefined): void {
    if (!files) {
      return;
    }
    const nextPhotos: SelectedPhoto[] = [];
    const rejected: string[] = [];
    for (const file of Array.from(files)) {
      if (!mediaPolicy.supportedContentTypes.includes(file.type as SelectedPhoto['contentType'])) {
        rejected.push(`${file.name} is not a supported image type.`);
        continue;
      }
      if (file.size <= 0 || file.size > mediaPolicy.maxBytes) {
        rejected.push(`${file.name} is larger than ${formatBytes(mediaPolicy.maxBytes)}.`);
        continue;
      }
      nextPhotos.push({
        id: `${file.name}-${file.lastModified}`,
        name: file.name,
        sizeBytes: file.size,
        contentType: file.type as SelectedPhoto['contentType'],
        previewUrl: URL.createObjectURL(file),
        file
      });
    }
    selectedPhotos = nextPhotos;
    photoError = rejected.join(' ');
    fileInputKey += 1;
  }

  function removePhoto(id: string): void {
    selectedPhotos = selectedPhotos.filter((photo) => photo.id !== id);
    if (selectedPhotos.length === 0) {
      photoError = '';
    }
  }

  function formatBytes(sizeBytes: number): string {
    if (sizeBytes < 1024 * 1024) {
      return `${Math.round(sizeBytes / 1024)} KB`;
    }
    return `${(sizeBytes / 1024 / 1024).toFixed(1)} MB`;
  }
</script>

{#if open}
  <div class="tray-backdrop" role="presentation" onclick={onClose}></div>
  <div class="add-tray" role="dialog" aria-modal="true" aria-labelledby="add-title">
    <div class="section-heading compact">
      <h2 id="add-title">Add stuff</h2>
      <Button.Root variant="ghost" size="icon-sm" aria-label="Close add tray" onclick={onClose}><X /></Button.Root>
    </div>

    <div class="kind-segment" role="group" aria-label="Asset kind">
      {#each assetKinds as option}
        <Button.Root variant={kind === option ? 'secondary' : 'outline'} onclick={() => { kind = option; }}>
          {assetKindLabel(option)}
        </Button.Root>
      {/each}
    </div>

    <div class="field-stack">
      <Label for="asset-title">Name</Label>
      <Input id="asset-title" bind:value={title} placeholder="Tomato fertilizer" />
    </div>

    <div class="field-stack">
      <Label>Parent</Label>
      <div class="parent-picker" role="listbox" aria-label="Parent target">
        <Button.Root variant={parentAssetId === '' ? 'secondary' : 'outline'} onclick={() => { parentAssetId = ''; }}>
          Inventory root
        </Button.Root>
        {#each parentTargets as target}
          <Button.Root variant={parentAssetId === target.id ? 'secondary' : 'outline'} onclick={() => { parentAssetId = target.id; }}>
            {target.title}
          </Button.Root>
        {/each}
      </div>
    </div>

    <div class="field-stack">
      <Label for="asset-description">Description</Label>
      <Textarea id="asset-description" bind:value={description} placeholder="Optional notes" />
    </div>

    <div class="photo-actions">
      <Label for="asset-photos" class="photo-label"><Upload /> Upload</Label>
      <Label for="asset-photos" class="photo-label"><Camera /> Camera</Label>
      {#key fileInputKey}
        <Input id="asset-photos" class="visually-hidden" type="file" accept="image/jpeg,image/png,image/webp" multiple onchange={(event) => captureFiles(event.currentTarget.files ?? undefined)} />
      {/key}
    </div>

    {#if selectedPhotos.length > 0}
      <div class="photo-preview-list">
        {#each selectedPhotos as photo}
          <div class="photo-preview">
            <img src={photo.previewUrl} alt={photo.name} />
            <span>{photo.name}</span>
            <Button.Root variant="ghost" size="icon-xs" aria-label={`Remove ${photo.name}`} onclick={() => removePhoto(photo.id)}><X /></Button.Root>
          </div>
        {/each}
      </div>
    {/if}
    {#if photoError}
      <p class="denied-note" role="alert">{photoError}</p>
    {/if}

    <div class="tray-actions">
      <Button.Root variant="outline" onclick={onClose}>Cancel</Button.Root>
      <Button.Root disabled={saving || title.trim().length === 0 || !!photoError} onclick={() => { void save(); }}>Save</Button.Root>
    </div>
  </div>
{/if}
