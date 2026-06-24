<script lang="ts">
  import Camera from '@lucide/svelte/icons/camera';
  import Upload from '@lucide/svelte/icons/upload';
  import X from '@lucide/svelte/icons/x';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { Textarea } from '$lib/components/ui/textarea/index.js';
  import type {
    AddAssetSubmission,
    AddAssetSaveResult,
    AssetKind,
    AssetViewModel,
    CustomAssetType,
    CustomFieldDefinition,
    MediaUploadPolicy,
    SelectedPhoto
  } from '$lib/domain/inventory';
  import { applicableCustomFieldDefinitions, assetKindLabel, assetKinds } from '$lib/domain/inventory';

  let {
    open,
    parentTargets,
    mediaPolicy,
    customAssetTypes,
    customFieldDefinitions,
    saving,
    onClose,
    onSave
  }: {
    open: boolean;
    parentTargets: AssetViewModel[];
    mediaPolicy: MediaUploadPolicy;
    customAssetTypes: CustomAssetType[];
    customFieldDefinitions: CustomFieldDefinition[];
    saving: boolean;
    onClose: () => void;
    onSave: (draft: AddAssetSubmission) => Promise<AddAssetSaveResult>;
  } = $props();

  let kind = $state<AssetKind>('item');
  let title = $state('');
  let description = $state('');
  let parentAssetId = $state('');
  let quickParentTitle = $state('');
  let quickParentKind = $state<'location' | 'container'>('location');
  let customAssetTypeId = $state('');
  let customFieldValues = $state<Record<string, string>>({});
  let selectedPhotos = $state<SelectedPhoto[]>([]);
  let photoError = $state('');
  let fileInputKey = $state(0);

  let activeCustomAssetTypes = $derived(customAssetTypes.filter((assetType) => assetType.lifecycleState === 'active'));
  let applicableFields = $derived(applicableCustomFieldDefinitions(customFieldDefinitions, customAssetTypeId || undefined));

  async function save(): Promise<void> {
    if (!title.trim() || photoError) {
      return;
    }
    const result = await onSave({
      kind,
      title: title.trim(),
      description: description.trim(),
      parentAssetId: parentAssetId || null,
      parentQuickCreate: quickParentTitle.trim()
        ? { kind: quickParentKind, title: quickParentTitle.trim() }
        : undefined,
      customAssetTypeId: customAssetTypeId || undefined,
      customFields: buildCustomFields(),
      photos: selectedPhotos
    });
    if (!result.saved) {
      if (result.createdParentId) {
        parentAssetId = result.createdParentId;
        quickParentTitle = '';
        quickParentKind = 'location';
      }
      return;
    }
    title = '';
    description = '';
    parentAssetId = '';
    quickParentTitle = '';
    quickParentKind = 'location';
    customAssetTypeId = '';
    customFieldValues = {};
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

  function setCustomAssetType(nextId: string): void {
    customAssetTypeId = nextId;
    customFieldValues = {};
  }

  function setCustomFieldValue(key: string, value: string): void {
    customFieldValues = { ...customFieldValues, [key]: value };
  }

  function buildCustomFields(): Record<string, unknown> {
    const values: Record<string, unknown> = {};
    for (const field of applicableFields) {
      const value = customFieldValues[field.key] ?? '';
      if (!value) {
        continue;
      }
      values[field.key] = field.type === 'number' ? Number(value) : field.type === 'boolean' ? value === 'true' : value;
    }
    return values;
  }

  function inputType(field: CustomFieldDefinition): string {
    if (field.type === 'number') return 'number';
    if (field.type === 'date') return 'date';
    if (field.type === 'url') return 'url';
    return 'text';
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
      <Label>Place in existing parent</Label>
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
      <Label for="quick-parent-title">Create a new parent inside that place</Label>
      <Input id="quick-parent-title" bind:value={quickParentTitle} placeholder="Laundry shelf" />
      <div class="kind-segment" role="group" aria-label="New parent kind">
        <Button.Root variant={quickParentKind === 'location' ? 'secondary' : 'outline'} onclick={() => { quickParentKind = 'location'; }}>
          Location
        </Button.Root>
        <Button.Root variant={quickParentKind === 'container' ? 'secondary' : 'outline'} onclick={() => { quickParentKind = 'container'; }}>
          Container
        </Button.Root>
      </div>
    </div>

    <div class="field-stack">
      <Label for="asset-description">Description</Label>
      <Textarea id="asset-description" bind:value={description} placeholder="Optional notes" />
    </div>

    {#if activeCustomAssetTypes.length > 0}
      <div class="field-stack">
        <Label>Custom type</Label>
        <div class="parent-picker" role="group" aria-label="Custom asset type">
          <Button.Root variant={customAssetTypeId === '' ? 'secondary' : 'outline'} onclick={() => setCustomAssetType('')}>
            Base asset
          </Button.Root>
          {#each activeCustomAssetTypes as assetType}
            <Button.Root
              variant={customAssetTypeId === assetType.id ? 'secondary' : 'outline'}
              onclick={() => setCustomAssetType(assetType.id)}
            >
              {assetType.displayName}
            </Button.Root>
          {/each}
        </div>
      </div>
    {/if}

    {#if applicableFields.length > 0}
      <div class="custom-field-grid" aria-label="Custom fields">
        {#each applicableFields as field}
          <div class="field-stack">
            <Label for={`custom-field-${field.key}`}>{field.displayName}</Label>
            {#if field.type === 'boolean'}
              <div class="kind-segment" role="group" aria-label={field.displayName}>
                <Button.Root variant={(customFieldValues[field.key] ?? '') === '' ? 'secondary' : 'outline'} onclick={() => setCustomFieldValue(field.key, '')}>
                  Unset
                </Button.Root>
                <Button.Root variant={customFieldValues[field.key] === 'true' ? 'secondary' : 'outline'} onclick={() => setCustomFieldValue(field.key, 'true')}>
                  Yes
                </Button.Root>
                <Button.Root variant={customFieldValues[field.key] === 'false' ? 'secondary' : 'outline'} onclick={() => setCustomFieldValue(field.key, 'false')}>
                  No
                </Button.Root>
              </div>
            {:else if field.type === 'enum'}
              <div class="parent-picker" role="group" aria-label={field.displayName}>
                <Button.Root variant={(customFieldValues[field.key] ?? '') === '' ? 'secondary' : 'outline'} onclick={() => setCustomFieldValue(field.key, '')}>
                  Unset
                </Button.Root>
                {#each field.enumOptions as option}
                  <Button.Root
                    variant={customFieldValues[field.key] === option ? 'secondary' : 'outline'}
                    onclick={() => setCustomFieldValue(field.key, option)}
                  >
                    {option}
                  </Button.Root>
                {/each}
              </div>
            {:else}
              <Input
                id={`custom-field-${field.key}`}
                type={inputType(field)}
                value={customFieldValues[field.key] ?? ''}
                oninput={(event) => setCustomFieldValue(field.key, event.currentTarget.value)}
              />
            {/if}
          </div>
        {/each}
      </div>
    {/if}

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
