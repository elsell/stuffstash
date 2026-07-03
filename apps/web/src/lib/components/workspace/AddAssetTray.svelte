<script lang="ts">
  import { onDestroy, tick } from 'svelte';
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
  import BinaryOption from './BinaryOption.svelte';
  import ParentTargetPicker from './ParentTargetPicker.svelte';
  import SegmentedControl from './SegmentedControl.svelte';

  let {
    open,
    initialKind = 'item',
    initialParentAssetId = null,
    closeHref,
    parentTargets,
    mediaPolicy,
    customAssetTypes,
    customFieldDefinitions,
    saving,
    onClose,
    onSave
  }: {
    open: boolean;
    initialKind?: AssetKind;
    initialParentAssetId?: string | null;
    closeHref: string;
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
  let parentSearch = $state('');
  let quickParentEnabled = $state(false);
  let quickParentTitle = $state('');
  let quickParentKind = $state<'location' | 'container'>('location');
  let customAssetTypeId = $state('');
  let customFieldValues = $state<Record<string, string>>({});
  let selectedPhotos = $state<SelectedPhoto[]>([]);
  let photoError = $state('');
  let fileInputKey = $state(0);
  let lastInitialKind = $state<AssetKind>('item');
  let lastInitialParentAssetId = $state<string | null>(null);
  let wasOpen = $state(false);
  let dialogElement = $state<HTMLElement | null>(null);
  let titleInput = $state<HTMLInputElement | null>(null);
  let fileInput = $state<HTMLInputElement | null>(null);
  let cameraInput = $state<HTMLInputElement | null>(null);
  let returnFocusElement: HTMLElement | null = null;
  const assetKindOptions = assetKinds.map((option) => ({ value: option, label: assetKindLabel(option) }));
  const parentKindOptions = [
    { value: 'location', label: 'Location' },
    { value: 'container', label: 'Container' }
  ];
  const booleanOptions = [
    { value: '', label: 'Unset' },
    { value: 'true', label: 'Yes' },
    { value: 'false', label: 'No' }
  ];

  let activeCustomAssetTypes = $derived(customAssetTypes.filter((assetType) => assetType.lifecycleState === 'active'));
  let applicableFields = $derived(applicableCustomFieldDefinitions(customFieldDefinitions, customAssetTypeId || undefined));
  let quickParentMissingName = $derived(quickParentEnabled && quickParentTitle.trim().length === 0);
  let selectedParent = $derived(parentTargets.find((target) => target.id === parentAssetId) ?? null);
  let parentSummary = $derived(destinationSummary());
  let photoSummary = $derived(photoCountLabel());

  $effect(() => {
    if (open && !wasOpen) {
      returnFocusElement = document.activeElement instanceof HTMLElement ? document.activeElement : null;
      resetDraft(initialKind, validInitialParentId(initialParentAssetId));
      wasOpen = true;
      void tick().then(() => titleInput?.focus());
    } else if (!open && wasOpen) {
      wasOpen = false;
      revokePhotoPreviews(selectedPhotos);
      selectedPhotos = [];
      photoError = '';
      returnFocusElement?.focus();
      returnFocusElement = null;
    } else if (open && initialKind !== lastInitialKind) {
      kind = initialKind;
      lastInitialKind = initialKind;
    } else if (open && initialParentAssetId !== lastInitialParentAssetId) {
      parentAssetId = validInitialParentId(initialParentAssetId) ?? '';
      parentSearch = parentAssetId ? parentTargets.find((target) => target.id === parentAssetId)?.title ?? '' : '';
      lastInitialParentAssetId = initialParentAssetId;
    }
  });

  onDestroy(() => {
    revokePhotoPreviews(selectedPhotos);
  });

  async function save(): Promise<void> {
    if (!title.trim() || photoError) {
      return;
    }
    const result = await onSave({
      kind,
      title: title.trim(),
      description: description.trim(),
      parentAssetId: parentAssetId || null,
      parentQuickCreate: quickParentEnabled && quickParentTitle.trim()
        ? { kind: quickParentKind, title: quickParentTitle.trim() }
        : undefined,
      customAssetTypeId: customAssetTypeId || undefined,
      customFields: buildCustomFields(),
      photos: selectedPhotos
    });
    if (!result.saved) {
      if (result.createdParentId) {
        parentAssetId = result.createdParentId;
        quickParentEnabled = false;
        quickParentTitle = '';
        quickParentKind = 'location';
      }
      return;
    }
    title = '';
    description = '';
    parentAssetId = '';
    parentSearch = '';
    quickParentEnabled = false;
    quickParentTitle = '';
    quickParentKind = 'location';
    customAssetTypeId = '';
    customFieldValues = {};
    revokePhotoPreviews(selectedPhotos);
    selectedPhotos = [];
    photoError = '';
    lastInitialKind = kind;
  }

  function resetDraft(nextKind: AssetKind, nextParentAssetId: string | null = null): void {
    revokePhotoPreviews(selectedPhotos);
    kind = nextKind;
    title = '';
    description = '';
    parentAssetId = nextParentAssetId ?? '';
    parentSearch = nextParentAssetId ? parentTargets.find((target) => target.id === nextParentAssetId)?.title ?? '' : '';
    quickParentEnabled = false;
    quickParentTitle = '';
    quickParentKind = 'location';
    customAssetTypeId = '';
    customFieldValues = {};
    selectedPhotos = [];
    photoError = '';
    fileInputKey += 1;
    lastInitialKind = nextKind;
    lastInitialParentAssetId = initialParentAssetId;
  }

  function handleDialogKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      event.preventDefault();
      onClose();
      return;
    }
    if (event.key !== 'Tab' || !dialogElement) {
      return;
    }
    const focusable = Array.from(
      dialogElement.querySelectorAll<HTMLElement>(
        'a[href], button:not([disabled]), input:not([disabled]), textarea:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])'
      )
    ).filter((element) => !element.hasAttribute('disabled') && element.getAttribute('aria-hidden') !== 'true');
    if (focusable.length === 0) {
      event.preventDefault();
      dialogElement.focus();
      return;
    }
    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  }

  function closeFromLink(event: MouseEvent): void {
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    onClose();
  }

  function shouldHandleInApp(event: MouseEvent): boolean {
    return event.button === 0 && !event.metaKey && !event.ctrlKey && !event.shiftKey && !event.altKey && !event.defaultPrevented;
  }

  function captureFiles(files: FileList | undefined): void {
    if (!files) {
      return;
    }
    revokePhotoPreviews(selectedPhotos);
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

  function openPhotoPicker(): void {
    fileInput?.click();
  }

  function openCameraPicker(): void {
    cameraInput?.click();
  }

  function removePhoto(id: string): void {
    const removed = selectedPhotos.find((photo) => photo.id === id);
    if (removed) {
      URL.revokeObjectURL(removed.previewUrl);
    }
    selectedPhotos = selectedPhotos.filter((photo) => photo.id !== id);
    if (selectedPhotos.length === 0) {
      photoError = '';
    }
  }

  function selectParentTarget(id: string | null): void {
    parentAssetId = id ?? '';
    parentSearch = id ? parentTargets.find((target) => target.id === id)?.title ?? parentSearch : '';
  }

  function validInitialParentId(id: string | null | undefined): string | null {
    return id && parentTargets.some((target) => target.id === id) ? id : null;
  }

  function toggleQuickParent(): void {
    quickParentEnabled = !quickParentEnabled;
    if (!quickParentEnabled) {
      quickParentTitle = '';
      quickParentKind = 'location';
    }
  }

  function setCustomAssetType(nextId: string): void {
    customAssetTypeId = nextId;
    customFieldValues = {};
  }

  function setCustomFieldValue(key: string, value: string): void {
    customFieldValues = { ...customFieldValues, [key]: value };
  }

  function destinationSummary(): string {
    if (quickParentEnabled) {
      const parentKindLabel = assetKindLabel(quickParentKind);
      return quickParentTitle.trim() ? `New ${parentKindLabel}: ${quickParentTitle.trim()}` : `New ${parentKindLabel}`;
    }
    return selectedParent?.title ?? 'Inventory root';
  }

  function photoCountLabel(): string {
    if (selectedPhotos.length === 0) {
      return 'No photos';
    }
    return `${selectedPhotos.length} ${selectedPhotos.length === 1 ? 'photo' : 'photos'}`;
  }

  function selectedKindLabel(): string {
    return assetKindLabel(kind).toLowerCase();
  }

  function addHeading(): string {
    return `Add ${selectedKindLabel()}`;
  }

  function nameLabel(): string {
    return `${assetKindLabel(kind)} name`;
  }

  function namePlaceholder(): string {
    if (kind === 'location') {
      return 'Garage shelf';
    }
    if (kind === 'container') {
      return 'Clear storage bin';
    }
    return 'Tomato fertilizer';
  }

  function saveLabel(): string {
    return `Save ${selectedKindLabel()}`;
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

  function revokePhotoPreviews(photos: SelectedPhoto[]): void {
    for (const photo of photos) {
      URL.revokeObjectURL(photo.previewUrl);
    }
  }
</script>

{#if open}
  <div class="tray-backdrop" role="presentation" onclick={onClose}></div>
  <div
    bind:this={dialogElement}
    class="add-tray"
    role="dialog"
    aria-modal="true"
    aria-labelledby="add-title"
    tabindex="-1"
    onkeydown={handleDialogKeydown}
  >
    <div class="section-heading compact">
      <h2 id="add-title">{addHeading()}</h2>
      <Button.Root href={closeHref} variant="ghost" size="icon-sm" aria-label="Close add tray" onclick={closeFromLink}><X /></Button.Root>
    </div>

    <div class="add-summary" aria-live="polite">
      <div>
        <small>Type</small>
        <strong>{assetKindLabel(kind)}</strong>
      </div>
      <div>
        <small>Parent</small>
        <strong>{parentSummary}</strong>
      </div>
      <div>
        <small>Photos</small>
        <strong>{photoSummary}</strong>
      </div>
    </div>

    <fieldset class="selection-field">
      <legend>Asset kind</legend>
      <SegmentedControl label="Asset kind" value={kind} options={assetKindOptions} onSelect={(value) => { kind = value as AssetKind; }} />
    </fieldset>

    <div class="field-stack">
      <Label for="asset-title">{nameLabel()}</Label>
      <Input id="asset-title" bind:ref={titleInput} bind:value={title} placeholder={namePlaceholder()} required aria-required="true" />
    </div>

    <ParentTargetPicker
      legend="Place in existing parent"
      searchId="parent-search"
      groupLabel="Parent target"
      bind:search={parentSearch}
      selectedId={parentAssetId || null}
      targets={parentTargets}
      onSelect={selectParentTarget}
    />

    <fieldset class="selection-field quick-parent-section">
      <legend>Create missing parent</legend>
      <BinaryOption
        label="Create a parent first"
        description="Use this when the shelf, box, or location does not exist yet."
        checked={quickParentEnabled}
        onToggle={toggleQuickParent}
      />
      {#if quickParentEnabled}
        <div class="quick-parent-fields">
          <div class="field-stack">
            <Label for="quick-parent-title">Parent name</Label>
            <Input
              id="quick-parent-title"
              bind:value={quickParentTitle}
              placeholder="Laundry shelf"
              required={quickParentEnabled}
              aria-required={quickParentEnabled}
              aria-invalid={quickParentMissingName}
              aria-describedby={quickParentMissingName ? 'quick-parent-error' : undefined}
            />
            {#if quickParentMissingName}
              <p id="quick-parent-error" class="denied-note" role="alert">Enter a parent name or turn this option off.</p>
            {/if}
          </div>
          <SegmentedControl
            label="New parent kind"
            value={quickParentKind}
            options={parentKindOptions}
            onSelect={(value) => { quickParentKind = value as 'location' | 'container'; }}
          />
        </div>
      {/if}
    </fieldset>

    <div class="field-stack">
      <Label for="asset-description">Description</Label>
      <Textarea id="asset-description" bind:value={description} placeholder="Optional notes" />
    </div>

    {#if activeCustomAssetTypes.length > 0}
      <div class="field-stack">
        <fieldset class="selection-field">
          <legend>Custom type</legend>
          <div class="parent-picker option-grid" role="group" aria-label="Custom asset type">
            <Button.Root
              variant={customAssetTypeId === '' ? 'secondary' : 'outline'}
              aria-pressed={customAssetTypeId === ''}
              onclick={() => setCustomAssetType('')}
            >
              Base asset
            </Button.Root>
            {#each activeCustomAssetTypes as assetType}
              <Button.Root
                variant={customAssetTypeId === assetType.id ? 'secondary' : 'outline'}
                aria-pressed={customAssetTypeId === assetType.id}
                onclick={() => setCustomAssetType(assetType.id)}
              >
                {assetType.displayName}
              </Button.Root>
            {/each}
          </div>
        </fieldset>
      </div>
    {/if}

    {#if applicableFields.length > 0}
      <div class="custom-field-grid" aria-label="Custom fields">
        {#each applicableFields as field}
          <div class="field-stack">
            <Label for={`custom-field-${field.key}`}>{field.displayName}</Label>
            {#if field.type === 'boolean'}
              <SegmentedControl
                label={field.displayName}
                value={customFieldValues[field.key] ?? ''}
                options={booleanOptions}
                onSelect={(value) => setCustomFieldValue(field.key, value)}
              />
            {:else if field.type === 'enum'}
              <div class="parent-picker option-grid" role="group" aria-label={field.displayName}>
                <Button.Root
                  variant={(customFieldValues[field.key] ?? '') === '' ? 'secondary' : 'outline'}
                  aria-pressed={(customFieldValues[field.key] ?? '') === ''}
                  onclick={() => setCustomFieldValue(field.key, '')}
                >
                  Unset
                </Button.Root>
                {#each field.enumOptions as option}
                  <Button.Root
                    variant={customFieldValues[field.key] === option ? 'secondary' : 'outline'}
                    aria-pressed={customFieldValues[field.key] === option}
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

    <fieldset class="selection-field attachment-section" aria-describedby="photo-help">
      <legend>Photos</legend>
      <p id="photo-help" class="selection-summary">Optional JPEG, PNG, or WebP up to {formatBytes(mediaPolicy.maxBytes)}.</p>
      <div class="photo-actions" role="group" aria-label="Photo actions">
        <Button.Root type="button" variant="outline" class="photo-label" onclick={openPhotoPicker}><Upload /> Upload</Button.Root>
        <Button.Root type="button" variant="outline" class="photo-label" onclick={openCameraPicker}><Camera /> Camera</Button.Root>
        <span class="photo-status" aria-live="polite">{photoSummary}</span>
      </div>
      {#key fileInputKey}
        <Input
          id="asset-photos"
          bind:ref={fileInput}
          class="visually-hidden"
          type="file"
          tabindex={-1}
          accept="image/jpeg,image/png,image/webp"
          multiple
          onchange={(event) => captureFiles(event.currentTarget.files ?? undefined)}
        />
        <Input
          id="asset-camera"
          bind:ref={cameraInput}
          class="visually-hidden"
          type="file"
          tabindex={-1}
          accept="image/jpeg,image/png,image/webp"
          capture="environment"
          onchange={(event) => captureFiles(event.currentTarget.files ?? undefined)}
        />
      {/key}
    </fieldset>

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
      <Button.Root href={closeHref} variant="outline" onclick={closeFromLink}>Cancel</Button.Root>
      <Button.Root disabled={saving || title.trim().length === 0 || !!photoError || quickParentMissingName} onclick={() => { void save(); }}>{saveLabel()}</Button.Root>
    </div>
  </div>
{/if}
