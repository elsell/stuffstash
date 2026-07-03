<script lang="ts">
  import ArrowLeft from '@lucide/svelte/icons/arrow-left';
  import Archive from '@lucide/svelte/icons/archive';
  import FileText from '@lucide/svelte/icons/file-text';
  import Image from '@lucide/svelte/icons/image';
  import MoveRight from '@lucide/svelte/icons/move-right';
  import Pencil from '@lucide/svelte/icons/pencil';
  import RotateCcw from '@lucide/svelte/icons/rotate-ccw';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import Upload from '@lucide/svelte/icons/upload';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { Textarea } from '$lib/components/ui/textarea/index.js';
  import {
    workspaceRouteHref,
    type AssetRouteAction,
    type AttachmentRouteAction,
    type WorkspaceRouteState
  } from '$lib/application/workspaceRoute';
  import type {
    AssetAttachment,
    AssetViewModel,
    CustomFieldDefinition,
    MediaUploadPolicy,
    SelectedAttachment,
    UpdateAssetDraft
  } from '$lib/domain/inventory';
  import { applicableCustomFieldDefinitions, assetKindLabel } from '$lib/domain/inventory';
  import KindIcon from './KindIcon.svelte';
  import ParentTargetPicker from './ParentTargetPicker.svelte';
  import SegmentedControl from './SegmentedControl.svelte';

  type DetailPhoto = {
    id: string;
    url: string;
    alt: string;
    fileName: string;
    sizeBytes?: number;
    isPrimary: boolean;
  };

  let {
    asset,
    canEdit,
    action = null,
    attachmentId = null,
    attachmentAction = null,
    parentTargets,
    customFieldDefinitions,
    saving,
    attachments,
    mediaPolicy,
    backHref,
    onBack,
    onActionOpen,
    onActionClose,
    onSave,
    onArchive,
    onRestore,
    onDelete,
    onUploadAttachment,
    onArchiveAttachment,
    onAttachmentDeleteOpen,
    onAttachmentDeleteClose,
    onDeleteAttachment
  }: {
    asset: AssetViewModel;
    canEdit: boolean;
    action?: AssetRouteAction;
    attachmentId?: string | null;
    attachmentAction?: AttachmentRouteAction;
    parentTargets: AssetViewModel[];
    customFieldDefinitions: CustomFieldDefinition[];
    saving: boolean;
    attachments: AssetAttachment[];
    mediaPolicy: MediaUploadPolicy;
    backHref: string;
    onBack: () => void;
    onActionOpen: (action: 'edit' | 'move' | 'archive' | 'restore' | 'delete') => void;
    onActionClose: () => void;
    onSave: (draft: UpdateAssetDraft) => Promise<void>;
    onArchive: () => Promise<void>;
    onRestore: () => Promise<void>;
    onDelete: () => Promise<void>;
    onUploadAttachment: (attachment: SelectedAttachment) => Promise<void>;
    onArchiveAttachment: (attachment: AssetAttachment) => Promise<void>;
    onAttachmentDeleteOpen: (attachmentId: string) => void;
    onAttachmentDeleteClose: () => void;
    onDeleteAttachment: (attachment: AssetAttachment) => Promise<void>;
  } = $props();

  let panel = $state<'none' | 'edit' | 'move' | 'archive' | 'restore' | 'delete' | 'attachment-delete'>('none');
  let title = $state('');
  let description = $state('');
  let parentAssetId = $state<string | null>(null);
  let moveParentSearch = $state('');
  let customFieldValues = $state<Record<string, string>>({});
  let saveError = $state('');
  let uploadError = $state('');
  let photoInput = $state<HTMLInputElement | null>(null);
  let fileInput = $state<HTMLInputElement | null>(null);
  let selectedAttachment = $state<AssetAttachment | null>(null);
  let selectedPhotoId = $state<string | null>(null);
  let lastRouteActionKey = $state('');
  let actionPanelElement = $state<HTMLElement | null>(null);
  let applicableFields = $derived(applicableCustomFieldDefinitions(customFieldDefinitions, asset.customAssetTypeId));
  let imageContentTypes = $derived(mediaPolicy.supportedContentTypes.filter((contentType) => contentType.startsWith('image/')));
  let photoAttachments = $derived(attachments.filter((attachment) => attachment.contentType.startsWith('image/')));
  let fileAttachments = $derived(attachments.filter((attachment) => !attachment.contentType.startsWith('image/')));
  let detailPhotos = $derived(buildDetailPhotos(asset, photoAttachments));
  let heroPhoto = $derived(
    detailPhotos.find((photo) => photo.id === selectedPhotoId) ?? detailPhotos.find((photo) => photo.isPrimary) ?? detailPhotos[0]
  );
  const booleanOptions = [
    { value: '', label: 'Unset' },
    { value: 'true', label: 'Yes' },
    { value: 'false', label: 'No' }
  ];
  let displayFields = $derived(
    customFieldDefinitions.filter(
      (definition) =>
        definition.applicability === 'all_assets' ||
        (!!asset.customAssetTypeId && definition.customAssetTypeIds.includes(asset.customAssetTypeId))
    )
  );

  $effect(() => {
    const attachmentKey = attachmentAction === 'delete' ? attachments.map((attachment) => attachment.id).join(',') : 'none';
    const actionKey = `${asset.id}:${action ?? 'none'}:${attachmentAction ?? 'none'}:${attachmentId ?? 'none'}:${attachmentKey}`;
    if (actionKey === lastRouteActionKey) {
      return;
    }
    const initializingWithoutRouteAction = lastRouteActionKey === '' && !action && !attachmentAction;
    lastRouteActionKey = actionKey;
    if (attachmentAction === 'delete' && canEdit) {
      const routeAttachment = attachments.find((attachment) => attachment.id === attachmentId) ?? null;
      selectedAttachment = routeAttachment;
      panel = routeAttachment ? 'attachment-delete' : 'none';
    } else if (action === 'edit' && canEdit && asset.lifecycleState === 'active') {
      openEdit(false);
    } else if (action === 'move' && canEdit && asset.lifecycleState === 'active') {
      openMove(false);
    } else if (action === 'archive' && canEdit && asset.lifecycleState === 'active') {
      panel = 'archive';
    } else if (action === 'restore' && canEdit && asset.lifecycleState === 'archived') {
      panel = 'restore';
    } else if (action === 'delete' && canEdit) {
      panel = 'delete';
    } else if (!action && !attachmentAction && !initializingWithoutRouteAction) {
      panel = 'none';
      selectedAttachment = null;
    }
  });

  $effect(() => {
    if (
      (panel === 'edit' ||
        panel === 'move' ||
        panel === 'archive' ||
        panel === 'restore' ||
        panel === 'delete' ||
        panel === 'attachment-delete') &&
      actionPanelElement
    ) {
      actionPanelElement.focus();
    }
  });

  $effect(() => {
    if (selectedPhotoId && !detailPhotos.some((photo) => photo.id === selectedPhotoId)) {
      selectedPhotoId = null;
    }
  });

  function openEdit(notify = true): void {
    title = asset.title;
    description = asset.description;
    parentAssetId = asset.parentAssetId;
    customFieldValues = Object.fromEntries(
      applicableFields.map((field) => [field.key, stringifyCustomFieldValue(asset.customFields?.[field.key])])
    );
    panel = 'edit';
    if (notify) {
      onActionOpen('edit');
    }
  }

  function openMove(notify = true): void {
    title = asset.title;
    description = asset.description;
    parentAssetId = asset.parentAssetId;
    moveParentSearch = '';
    customFieldValues = Object.fromEntries(
      applicableFields.map((field) => [field.key, stringifyCustomFieldValue(asset.customFields?.[field.key])])
    );
    panel = 'move';
    if (notify) {
      onActionOpen('move');
    }
  }

  function openArchive(): void {
    panel = 'archive';
    onActionOpen('archive');
  }

  function openRestore(): void {
    panel = 'restore';
    onActionOpen('restore');
  }

  function openAction(event: MouseEvent, nextAction: Exclude<AssetRouteAction, null>): void {
    if (!actionIsAvailable(nextAction)) {
      return;
    }
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    if (nextAction === 'edit') {
      openEdit();
    } else if (nextAction === 'move') {
      openMove();
    } else if (nextAction === 'archive') {
      openArchive();
    } else if (nextAction === 'restore') {
      openRestore();
    } else {
      panel = 'delete';
      onActionOpen('delete');
    }
  }

  function actionHref(nextAction: Exclude<AssetRouteAction, null>): string {
    const route: Partial<WorkspaceRouteState> = {
      mode: 'asset',
      tenantId: asset.tenantId,
      inventoryId: asset.inventoryId,
      assetId: asset.id,
      assetAction: nextAction,
      action: nextAction === 'edit' ? 'edit' : null
    };
    if (nextAction === 'edit' && asset.kind === 'location') {
      route.locationId = asset.id;
    }
    return workspaceRouteHref(route, asset.tenantId, asset.inventoryId);
  }

  function detailHref(): string {
    return workspaceRouteHref(
      asset.kind === 'location'
        ? {
            mode: 'location',
            tenantId: asset.tenantId,
            inventoryId: asset.inventoryId,
            locationId: asset.id
          }
        : {
            mode: 'asset',
            tenantId: asset.tenantId,
            inventoryId: asset.inventoryId,
            assetId: asset.id
          },
      asset.tenantId,
      asset.inventoryId
    );
  }

  function attachmentDeleteHref(attachment: AssetAttachment): string {
    return workspaceRouteHref(
      {
        mode: 'asset',
        tenantId: asset.tenantId,
        inventoryId: asset.inventoryId,
        assetId: asset.id,
        attachmentId: attachment.id,
        attachmentAction: 'delete'
      },
      asset.tenantId,
      asset.inventoryId
    );
  }

  function actionIsAvailable(nextAction: Exclude<AssetRouteAction, null>): boolean {
    if (!canEdit || saving) {
      return false;
    }
    if (nextAction === 'delete') {
      return true;
    }
    if (nextAction === 'restore') {
      return asset.lifecycleState === 'archived';
    }
    return asset.lifecycleState === 'active';
  }

  function shouldHandleInApp(event: MouseEvent): boolean {
    return event.button === 0 && !event.metaKey && !event.ctrlKey && !event.shiftKey && !event.altKey;
  }

  async function save(): Promise<void> {
    if (!title.trim()) {
      return;
    }
    saveError = '';
    try {
      await onSave({
        title: title.trim(),
        description: description.trim(),
        parentAssetId,
        customFields: buildCustomFields()
      });
      closePanel();
    } catch (caught) {
      saveError = caught instanceof Error ? caught.message : 'Unable to save asset.';
    }
  }

  function closePanel(): void {
    const previousPanel = panel;
    panel = 'none';
    selectedAttachment = null;
    if (
      previousPanel === 'edit' ||
      previousPanel === 'move' ||
      previousPanel === 'archive' ||
      previousPanel === 'restore' ||
      previousPanel === 'delete'
    ) {
      onActionClose();
    } else if (previousPanel === 'attachment-delete') {
      onAttachmentDeleteClose();
    }
  }

  function closeAction(event: MouseEvent): void {
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    closePanel();
  }

  async function archive(): Promise<void> {
    saveError = '';
    try {
      await onArchive();
    } catch (caught) {
      saveError = caught instanceof Error ? caught.message : 'Unable to archive asset.';
    }
  }

  async function restore(): Promise<void> {
    saveError = '';
    try {
      await onRestore();
    } catch (caught) {
      saveError = caught instanceof Error ? caught.message : 'Unable to restore asset.';
    }
  }

  async function remove(): Promise<void> {
    saveError = '';
    try {
      await onDelete();
    } catch (caught) {
      saveError = caught instanceof Error ? caught.message : 'Unable to delete asset.';
    }
  }

  async function archiveAttachment(attachment: AssetAttachment): Promise<void> {
    saveError = '';
    try {
      await onArchiveAttachment(attachment);
    } catch (caught) {
      saveError = caught instanceof Error ? caught.message : 'Unable to archive attachment.';
    }
  }

  async function uploadAttachment(event: Event, fallbackMessage = 'Unable to upload attachment.'): Promise<void> {
    uploadError = '';
    const input = event.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) {
      return;
    }
    const contentType = file.type;
    if (!isSupportedAttachmentContentType(contentType)) {
      uploadError = 'Unsupported file type.';
      input.value = '';
      return;
    }
    if (file.size > mediaPolicy.maxBytes) {
      uploadError = `Attachment must be ${formatBytes(mediaPolicy.maxBytes)} or smaller.`;
      input.value = '';
      return;
    }
    try {
      await onUploadAttachment({
        id: createClientAttachmentId(),
        name: file.name,
        sizeBytes: file.size,
        contentType,
        file
      });
      input.value = '';
    } catch (caught) {
      uploadError = caught instanceof Error ? caught.message : fallbackMessage;
      input.value = '';
    }
  }

  async function removeAttachment(): Promise<void> {
    if (!selectedAttachment) {
      return;
    }
    saveError = '';
    try {
      await onDeleteAttachment(selectedAttachment);
      selectedAttachment = null;
      panel = 'none';
      onAttachmentDeleteClose();
    } catch (caught) {
      saveError = caught instanceof Error ? caught.message : 'Unable to delete attachment.';
    }
  }

  function openAttachmentDelete(event: MouseEvent, attachment: AssetAttachment): void {
    if (!canEdit || saving) {
      return;
    }
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    selectedAttachment = attachment;
    panel = 'attachment-delete';
    onAttachmentDeleteOpen(attachment.id);
  }

  function formatBytes(sizeBytes: number): string {
    if (sizeBytes < 1024) {
      return `${sizeBytes} B`;
    }
    if (sizeBytes < 1024 * 1024) {
      return `${Math.round(sizeBytes / 1024)} KB`;
    }
    return `${(sizeBytes / 1024 / 1024).toFixed(1)} MB`;
  }

  function setCustomFieldValue(key: string, value: string): void {
    customFieldValues = { ...customFieldValues, [key]: value };
  }

  function selectMoveParent(id: string | null): void {
    parentAssetId = id;
    moveParentSearch = id ? parentTargets.find((target) => target.id === id)?.title ?? moveParentSearch : '';
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

  function stringifyCustomFieldValue(value: unknown): string {
    if (value === undefined || value === null) {
      return '';
    }
    return String(value);
  }

  function inputType(field: CustomFieldDefinition): string {
    if (field.type === 'number') return 'number';
    if (field.type === 'date') return 'date';
    if (field.type === 'url') return 'url';
    return 'text';
  }

  function isSupportedAttachmentContentType(contentType: string): contentType is SelectedAttachment['contentType'] {
    return mediaPolicy.supportedContentTypes.includes(contentType as SelectedAttachment['contentType']);
  }

  function buildDetailPhotos(currentAsset: AssetViewModel, imageAttachments: AssetAttachment[]): DetailPhoto[] {
    const ownAssetPhoto = currentAsset.photo?.assetId === currentAsset.id ? currentAsset.photo : undefined;
    const photos: DetailPhoto[] = imageAttachments
      .filter((attachment) => attachment.assetId === currentAsset.id)
      .filter((attachment) => attachment.thumbnailUrl)
      .map((attachment) => ({
        id: attachment.id,
        url: attachment.id === ownAssetPhoto?.id ? ownAssetPhoto.url : (attachment.thumbnailUrl ?? ''),
        alt: attachment.id === ownAssetPhoto?.id ? ownAssetPhoto.alt : attachment.fileName,
        fileName: attachment.fileName,
        sizeBytes: attachment.sizeBytes,
        isPrimary: attachment.id === ownAssetPhoto?.id
      }));
    if (ownAssetPhoto && !photos.some((photo) => photo.id === ownAssetPhoto.id)) {
      photos.unshift({
        id: ownAssetPhoto.id,
        url: ownAssetPhoto.url,
        alt: ownAssetPhoto.alt,
        fileName: ownAssetPhoto.alt,
        isPrimary: true
      });
    }
    if (ownAssetPhoto && photos.length > 0 && !photos.some((photo) => photo.isPrimary)) {
      photos[0] = { ...photos[0], isPrimary: true };
    }
    return photos;
  }

  function createClientAttachmentId(): string {
    return typeof crypto !== 'undefined' && 'randomUUID' in crypto
      ? crypto.randomUUID()
      : `attachment-${Date.now()}`;
  }

  function openBack(event: MouseEvent): void {
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    onBack();
  }
</script>

<section class="workspace-main detail-view" aria-labelledby="asset-title">
  <Button.Root href={backHref} variant="ghost" class="back-button" onclick={openBack}><ArrowLeft /> Back</Button.Root>
  <Input
    bind:ref={photoInput}
    aria-label="Choose photo"
    class="visually-hidden"
    type="file"
    accept={imageContentTypes.join(',')}
    disabled={!canEdit || asset.lifecycleState !== 'active' || saving}
    onchange={(event) => { void uploadAttachment(event, 'Unable to upload photo.'); }}
  />
  <Input
    bind:ref={fileInput}
    aria-label="Choose file"
    class="visually-hidden"
    type="file"
    accept={mediaPolicy.supportedContentTypes.join(',')}
    disabled={!canEdit || asset.lifecycleState !== 'active' || saving}
    onchange={(event) => { void uploadAttachment(event); }}
  />
  <div class="asset-detail-hero">
    <div class="asset-photo-panel" aria-label="Asset photos">
      <div class="asset-hero-photo">
        {#if heroPhoto}
          <img src={heroPhoto.url} alt={heroPhoto.alt} />
        {:else}
          <div class="asset-hero-fallback">
            <KindIcon kind={asset.kind} />
          </div>
        {/if}
      </div>
    </div>
    <div class="asset-detail-copy">
      <div class="detail-title-row">
        <div>
          <h1 id="asset-title">{asset.title}</h1>
          <p>{asset.containmentTrail}</p>
        </div>
        <Badge variant={asset.lifecycleState === 'active' ? 'secondary' : 'outline'}>{asset.lifecycleState}</Badge>
      </div>
      <dl class="detail-list">
        <div><dt>Kind</dt><dd>{assetKindLabel(asset.kind)}</dd></div>
        <div><dt>Type</dt><dd>{asset.customAssetTypeLabel ?? 'Base asset'}</dd></div>
        <div><dt>Updated</dt><dd>{asset.updatedAt ? new Date(asset.updatedAt).toLocaleString() : 'Not available'}</dd></div>
      </dl>
      <div class="detail-actions">
        <Button.Root href={actionHref('edit')} disabled={!actionIsAvailable('edit')} onclick={(event) => openAction(event, 'edit')}><Pencil /> Edit</Button.Root>
        <Button.Root
          href={actionHref('move')}
          variant="outline"
          disabled={!actionIsAvailable('move')}
          onclick={(event) => openAction(event, 'move')}
        ><MoveRight /> Move</Button.Root>
        <Button.Root
          variant="outline"
          disabled={!canEdit || asset.lifecycleState !== 'active' || saving || imageContentTypes.length === 0}
          onclick={() => photoInput?.click()}
        >
          <Image /> Add photo
        </Button.Root>
        {#if asset.lifecycleState === 'active'}
          <Button.Root
            href={actionHref('archive')}
            variant="outline"
            disabled={!actionIsAvailable('archive')}
            onclick={(event) => openAction(event, 'archive')}
          ><Archive /> Archive</Button.Root>
        {:else}
          <Button.Root
            href={actionHref('restore')}
            variant="outline"
            disabled={!actionIsAvailable('restore')}
            onclick={(event) => openAction(event, 'restore')}
          ><RotateCcw /> Restore</Button.Root>
        {/if}
      </div>
      {#if !canEdit}
        <p class="denied-note">Edit actions require asset edit access.</p>
      {/if}
    </div>
    <div class="photo-gallery-section" aria-label="Asset photo gallery">
      <div class="photo-panel-actions">
        <Button.Root
          variant="outline"
          disabled={!canEdit || asset.lifecycleState !== 'active' || saving || imageContentTypes.length === 0}
          onclick={() => photoInput?.click()}
        >
          <Image /> Add photo
        </Button.Root>
      </div>
      {#if detailPhotos.length > 0}
        <div class="photo-rail" aria-label="Photos">
          {#each detailPhotos as photo}
            <Button.Root
              variant="ghost"
              class={photo.id === heroPhoto?.id ? 'active' : ''}
              aria-label={`Show ${photo.fileName}`}
              aria-pressed={photo.id === heroPhoto?.id}
              onclick={() => { selectedPhotoId = photo.id; }}
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
  <div class="asset-detail-sections">
    <section class="detail-section" aria-labelledby="asset-description-title">
      <h2 id="asset-description-title">Details</h2>
      <p>{asset.description || 'No description.'}</p>
      {#if displayFields.length > 0}
        <dl class="detail-list custom-detail-list" aria-label="Custom field values">
          {#each displayFields as field}
            <div>
              <dt>{field.displayName}</dt>
              <dd>{stringifyCustomFieldValue(asset.customFields?.[field.key]) || 'Not set'}</dd>
            </div>
          {/each}
        </dl>
      {/if}
    </section>
    <section class="attachment-section" aria-labelledby="files-title">
        <div class="section-heading compact">
          <h2 id="files-title">Files</h2>
          <div class="attachment-upload">
            <Button.Root
              variant="outline"
              disabled={!canEdit || asset.lifecycleState !== 'active' || saving}
              onclick={() => fileInput?.click()}
            >
              <Upload /> Upload file
            </Button.Root>
          </div>
        </div>
        {#if fileAttachments.length === 0}
          <div class="empty-state">
            <p>No active files.</p>
          </div>
        {:else}
          <div class="asset-list">
            {#each fileAttachments as attachment}
              <div class="attachment-row">
                <div class="asset-thumb asset-thumb-sm"><FileText aria-hidden="true" /></div>
                <span class="asset-row-main">
                  <strong>{attachment.fileName}</strong>
                  <small>{attachment.contentType} / {formatBytes(attachment.sizeBytes)}</small>
                </span>
                <div class="attachment-actions">
                  <Button.Root variant="outline" disabled={!canEdit || saving} onclick={() => { void archiveAttachment(attachment); }}>Archive</Button.Root>
                  <Button.Root
                    href={attachmentDeleteHref(attachment)}
                    variant="destructive"
                    disabled={!canEdit || saving}
                    onclick={(event) => openAttachmentDelete(event, attachment)}
                  >
                    <Trash2 /> Delete
                  </Button.Root>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </section>
      {#if panel === 'edit'}
        <section
          bind:this={actionPanelElement}
          class="detail-action-panel"
          aria-labelledby="edit-asset-panel-title"
          tabindex="-1"
        >
          <h2 id="edit-asset-panel-title">Edit asset</h2>
          <div class="field-stack">
            <Label for="edit-asset-title">Name</Label>
            <Input id="edit-asset-title" bind:value={title} />
          </div>
          <div class="field-stack">
            <Label for="edit-asset-description">Description</Label>
            <Textarea id="edit-asset-description" bind:value={description} />
          </div>
          {#if applicableFields.length > 0}
            <div class="custom-field-grid" aria-label="Edit custom fields">
              {#each applicableFields as field}
                <div class="field-stack">
                  <Label for={`edit-custom-field-${field.key}`}>{field.displayName}</Label>
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
                      id={`edit-custom-field-${field.key}`}
                      type={inputType(field)}
                      value={customFieldValues[field.key] ?? ''}
                      oninput={(event) => setCustomFieldValue(field.key, event.currentTarget.value)}
                    />
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
          <div class="tray-actions">
            <Button.Root href={detailHref()} variant="outline" onclick={closeAction}>Cancel</Button.Root>
            <Button.Root disabled={saving || title.trim().length === 0} onclick={() => { void save(); }}>Save</Button.Root>
          </div>
          {#if saveError}
            <p class="denied-note" role="alert">{saveError}</p>
          {/if}
        </section>
      {:else if panel === 'move'}
        <section
          bind:this={actionPanelElement}
          class="detail-action-panel"
          aria-labelledby="move-asset-panel-title"
          tabindex="-1"
        >
          <h2 id="move-asset-panel-title">Move asset</h2>
          <ParentTargetPicker
            legend="Parent"
            searchId="move-parent-search"
            groupLabel="Move target"
            bind:search={moveParentSearch}
            selectedId={parentAssetId}
            targets={parentTargets}
            onSelect={selectMoveParent}
          />
          <div class="tray-actions">
            <Button.Root href={detailHref()} variant="outline" onclick={closeAction}>Cancel</Button.Root>
            <Button.Root disabled={saving} onclick={() => { void save(); }}>Move</Button.Root>
          </div>
          {#if saveError}
            <p class="denied-note" role="alert">{saveError}</p>
          {/if}
        </section>
      {:else if panel === 'archive'}
        <section
          bind:this={actionPanelElement}
          class="detail-action-panel"
          aria-labelledby="archive-asset-panel-title"
          tabindex="-1"
        >
          <h2 id="archive-asset-panel-title">Archive asset</h2>
          <p>Move {asset.title} out of active browsing?</p>
          <div class="tray-actions">
            <Button.Root href={detailHref()} variant="outline" onclick={closeAction}>Cancel</Button.Root>
            <Button.Root variant="outline" disabled={saving} onclick={() => { void archive(); }}>Archive</Button.Root>
          </div>
          {#if saveError}
            <p class="denied-note" role="alert">{saveError}</p>
          {/if}
        </section>
      {:else if panel === 'restore'}
        <section
          bind:this={actionPanelElement}
          class="detail-action-panel"
          aria-labelledby="restore-asset-panel-title"
          tabindex="-1"
        >
          <h2 id="restore-asset-panel-title">Restore asset</h2>
          <p>Return {asset.title} to active browsing?</p>
          <div class="tray-actions">
            <Button.Root href={detailHref()} variant="outline" onclick={closeAction}>Cancel</Button.Root>
            <Button.Root disabled={saving} onclick={() => { void restore(); }}>Restore</Button.Root>
          </div>
          {#if saveError}
            <p class="denied-note" role="alert">{saveError}</p>
          {/if}
        </section>
      {:else if panel === 'delete'}
        <section
          bind:this={actionPanelElement}
          class="detail-action-panel"
          aria-labelledby="delete-asset-panel-title"
          tabindex="-1"
        >
          <h2 id="delete-asset-panel-title">Delete asset</h2>
          <p>Delete {asset.title} permanently?</p>
          <div class="tray-actions">
            <Button.Root href={detailHref()} variant="outline" onclick={closeAction}>Cancel</Button.Root>
            <Button.Root variant="destructive" disabled={saving} onclick={() => { void remove(); }}>Delete</Button.Root>
          </div>
          {#if saveError}
            <p class="denied-note" role="alert">{saveError}</p>
          {/if}
        </section>
      {:else if panel === 'attachment-delete' && selectedAttachment}
        <section
          bind:this={actionPanelElement}
          class="detail-action-panel"
          aria-labelledby="delete-attachment-panel-title"
          tabindex="-1"
        >
          <h2 id="delete-attachment-panel-title">Delete attachment</h2>
          <p>Delete {selectedAttachment.fileName} permanently?</p>
          <div class="tray-actions">
            <Button.Root href={detailHref()} variant="outline" onclick={closeAction}>Cancel</Button.Root>
            <Button.Root variant="destructive" disabled={saving} onclick={() => { void removeAttachment(); }}>Delete</Button.Root>
          </div>
          {#if saveError}
            <p class="denied-note" role="alert">{saveError}</p>
          {/if}
        </section>
      {/if}
      <div class="danger-zone" aria-label="Danger area">
        <div>
          <strong>Permanent deletion</strong>
          <p>Remove this asset from the inventory permanently.</p>
        </div>
        <Button.Root
          href={actionHref('delete')}
          variant="destructive"
          disabled={!actionIsAvailable('delete')}
          onclick={(event) => openAction(event, 'delete')}
        ><Trash2 /> Delete</Button.Root>
      </div>
  </div>
</section>
