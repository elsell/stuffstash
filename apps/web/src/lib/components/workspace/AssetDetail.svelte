<script lang="ts">
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import ArrowLeft from '@lucide/svelte/icons/arrow-left';
  import Archive from '@lucide/svelte/icons/archive';
  import Image from '@lucide/svelte/icons/image';
  import MoveRight from '@lucide/svelte/icons/move-right';
  import Pencil from '@lucide/svelte/icons/pencil';
  import RotateCcw from '@lucide/svelte/icons/rotate-ccw';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import {
    assetActionHref,
    assetActionIsAvailable,
    assetDetailHref,
    attachmentDeleteHref as assetAttachmentDeleteHref
  } from '$lib/application/workspaceAssetActions';
  import { assetDescriptionText, assetEditUnavailableStatus } from '$lib/application/workspaceAssetDetail';
  import {
    buildDetailPhotos,
    photoUploadUnavailableReason,
    supportedAttachmentContentType,
    supportedImageContentType,
    unsupportedAttachmentTypeMessage,
    unsupportedImageTypeMessage
  } from '$lib/application/workspaceAssetMedia';
  import type { AssetRouteAction, AttachmentRouteAction } from '$lib/application/workspaceRoute';
  import type {
    AssetAttachment,
    AssetViewModel,
    CustomFieldDefinition,
    MediaUploadPolicy,
    ParentTargetViewModel,
    SelectedAttachment,
    UpdateAssetDraft
  } from '$lib/domain/inventory';
  import { applicableCustomFieldDefinitions, assetKindLabel } from '$lib/domain/inventory';
  import AssetDetailActionPanel, { type AssetDetailPanel } from './AssetDetailActionPanel.svelte';
  import AssetDetailHero, { PHOTO_UPLOAD_DISABLED_REASON_ID, PHOTO_UPLOAD_ERROR_ID } from './AssetDetailHero.svelte';
  import AssetFilesSection from './AssetFilesSection.svelte';
  import { formatBytes } from './formatBytes';

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
    parentTargets: ParentTargetViewModel[];
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

  let panel = $state<AssetDetailPanel>('none');
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
  let heroPhoto = $derived(detailPhotos.find((photo) => photo.id === selectedPhotoId) ?? detailPhotos.find((photo) => photo.isPrimary) ?? detailPhotos[0]);
  let canAddPhoto = $derived(canEdit && asset.lifecycleState === 'active' && !saving && imageContentTypes.length > 0);
  let editUnavailableStatus = $derived(assetEditUnavailableStatus(canEdit));
  let descriptionText = $derived(assetDescriptionText(asset.description));
  let photoUploadDisabledReason = $derived(
    photoUploadUnavailableReason({
      canEditAsset: canEdit,
      lifecycleState: asset.lifecycleState,
      isSaving: saving,
      supportedImageTypeCount: imageContentTypes.length
    })
  );
  let photoUploadDescribedBy = $derived(
    [
      photoUploadDisabledReason ? PHOTO_UPLOAD_DISABLED_REASON_ID : '',
      uploadError ? PHOTO_UPLOAD_ERROR_ID : ''
    ].filter(Boolean).join(' ')
  );
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
    if (!shouldHandleWorkspaceLinkClick(event)) {
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
    return assetActionHref(asset, nextAction);
  }

  function detailHref(): string {
    return assetDetailHref(asset);
  }

  function attachmentDeleteHref(attachment: AssetAttachment): string {
    return assetAttachmentDeleteHref(asset, attachment);
  }

  function actionIsAvailable(nextAction: Exclude<AssetRouteAction, null>): boolean {
    return assetActionIsAvailable(asset, nextAction, { canEdit, saving });
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
    if (!shouldHandleWorkspaceLinkClick(event)) {
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
    if (!supportedAttachmentContentType(mediaPolicy.supportedContentTypes, contentType)) {
      uploadError = unsupportedAttachmentTypeMessage();
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

  async function uploadPhotoAttachment(event: Event): Promise<void> {
    uploadError = '';
    const input = event.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) {
      return;
    }
    const contentType = file.type;
    if (!supportedImageContentType(imageContentTypes, contentType)) {
      uploadError = unsupportedImageTypeMessage();
      input.value = '';
      return;
    }
    await uploadAttachment(event, 'Unable to upload photo.');
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
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    selectedAttachment = attachment;
    panel = 'attachment-delete';
    onAttachmentDeleteOpen(attachment.id);
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

  function createClientAttachmentId(): string {
    return typeof crypto !== 'undefined' && 'randomUUID' in crypto ? crypto.randomUUID() : `attachment-${Date.now()}`;
  }

  function openBack(event: MouseEvent): void {
    if (!shouldHandleWorkspaceLinkClick(event)) {
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
    disabled={!canAddPhoto}
    onchange={(event) => { void uploadPhotoAttachment(event); }}
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
  <AssetDetailHero
      kind={asset.kind}
      {heroPhoto}
      photos={detailPhotos}
      {canAddPhoto}
      uploadDisabledReason={photoUploadDisabledReason}
      {uploadError}
      onChoosePhoto={() => photoInput?.click()}
      onSelectPhoto={(photoId) => { selectedPhotoId = photoId; }}
    >
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
            disabled={!canAddPhoto}
            aria-describedby={photoUploadDescribedBy || undefined}
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
        {#if editUnavailableStatus}
          <p class="denied-note">{editUnavailableStatus.message}</p>
        {/if}
      </div>
    </AssetDetailHero>
  <div class="asset-detail-sections">
      <AssetDetailActionPanel
        {panel}
        bind:panelElement={actionPanelElement}
        {asset}
        {parentTargets}
        {selectedAttachment}
        {saving}
        {saveError}
        detailHref={detailHref()}
        {applicableFields}
        bind:title
        bind:description
        bind:parentAssetId
        bind:moveParentSearch
        {customFieldValues}
        onClose={closeAction}
        onSave={save}
        onArchive={archive}
        onRestore={restore}
        onDelete={remove}
        onDeleteAttachment={removeAttachment}
        onParentSelect={selectMoveParent}
        onCustomFieldValueChange={setCustomFieldValue}
      />
    <section class="detail-section" aria-labelledby="asset-description-title">
      <h2 id="asset-description-title">Details</h2>
      <p>{descriptionText}</p>
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
    <AssetFilesSection
      attachments={fileAttachments}
      {canEdit}
      {saving}
      active={asset.lifecycleState === 'active'}
      onChooseFile={() => fileInput?.click()}
      onArchiveAttachment={(attachment) => { void archiveAttachment(attachment); }}
      onOpenAttachmentDelete={openAttachmentDelete}
      {attachmentDeleteHref}
    />
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
