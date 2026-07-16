<script lang="ts">
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import { tick } from 'svelte';
  import ArrowLeft from '@lucide/svelte/icons/arrow-left';
  import Archive from '@lucide/svelte/icons/archive';
  import LogOut from '@lucide/svelte/icons/log-out';
  import MoveRight from '@lucide/svelte/icons/move-right';
  import Pencil from '@lucide/svelte/icons/pencil';
  import RotateCcw from '@lucide/svelte/icons/rotate-ccw';
  import Undo2 from '@lucide/svelte/icons/undo-2';
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
  import { assetDescriptionText, assetEditUnavailableStatus, partitionAssetDetailFields } from '$lib/application/workspaceAssetDetail';
  import {
    buildDetailPhotos,
    photoUploadUnavailableReason,
    supportedAttachmentContentType,
    supportedImageContentType,
    unsupportedAttachmentTypeMessage,
    unsupportedImageTypeMessage,
    userSafeMediaErrorMessage
  } from '$lib/application/workspaceAssetMedia';
  import type { AssetRouteAction, AttachmentRouteAction } from '$lib/application/workspaceRoute';
	  import type {
	    Asset,
	    AssetAttachment,
	    AssetCheckout,
	    AssetViewModel,
    AssetTag,
    AssetTagDraft,
    CustomFieldDefinition,
    MediaUploadPolicy,
    ParentTargetViewModel,
    SelectedAttachment,
    UpdateAssetDraft
  } from '$lib/domain/inventory';
  import { applicableCustomFieldDefinitions, assetKindLabel } from '$lib/domain/inventory';
  import AssetDetailActionPanel, { type AssetDetailPanel } from './AssetDetailActionPanel.svelte';
  import AssetDetailHero, { PHOTO_UPLOAD_DISABLED_REASON_ID, PHOTO_UPLOAD_ERROR_ID } from './AssetDetailHero.svelte';
  import AssetTagChips from './AssetTagChips.svelte';
  import AssetFilesSection, { type AssetFilesError } from './AssetFilesSection.svelte';
  import CheckoutBadge from './CheckoutBadge.svelte';
  import ContainedAssetWorkspace from './ContainedAssetWorkspace.svelte';
  import { formatBytes } from './formatBytes';

  let {
    asset,
    canEdit,
    canCreate = false,
    workspaceAssets = [],
    action = null,
    attachmentId = null,
    attachmentAction = null,
    parentTargets,
    customFieldDefinitions,
    assetTags = [],
    saving,
    attachments,
    checkoutHistory,
    mediaPolicy,
    backHref,
    onBack,
    onActionOpen,
    onActionClose,
    onOpenAsset = () => {},
    onOpenAdd = () => {},
    onMoveHere = async () => {},
    onSave,
    onArchive,
    onRestore,
    onDelete,
    onCheckout,
    onReturn,
    onUploadAttachment,
    onArchiveAttachment,
    onAttachmentDeleteOpen,
    onAttachmentDeleteClose,
    onDeleteAttachment,
    onTagSearch
  }: {
    asset: AssetViewModel;
    canEdit: boolean;
    canCreate?: boolean;
    workspaceAssets?: Asset[];
    action?: AssetRouteAction;
    attachmentId?: string | null;
    attachmentAction?: AttachmentRouteAction;
    parentTargets: ParentTargetViewModel[];
    customFieldDefinitions: CustomFieldDefinition[];
    assetTags?: AssetTag[];
    saving: boolean;
    attachments: AssetAttachment[];
    checkoutHistory: AssetCheckout[];
    mediaPolicy: MediaUploadPolicy;
    backHref: string;
    onBack: () => void;
    onActionOpen: (action: Exclude<AssetRouteAction, null>) => void;
    onActionClose: () => void;
    onOpenAsset?: (asset: Asset) => void;
    onOpenAdd?: (kind: 'item', parentAssetId: string) => void;
    onMoveHere?: (asset: Asset) => Promise<void>;
    onSave: (draft: UpdateAssetDraft) => Promise<void>;
    onArchive: () => Promise<void>;
    onRestore: () => Promise<void>;
    onDelete: () => Promise<void>;
    onCheckout: (details: string) => Promise<void>;
    onReturn: (details: string) => Promise<void>;
    onUploadAttachment: (attachment: SelectedAttachment) => Promise<void>;
    onArchiveAttachment: (attachment: AssetAttachment) => Promise<void>;
    onAttachmentDeleteOpen: (attachmentId: string) => void;
    onAttachmentDeleteClose: () => void;
    onDeleteAttachment: (attachment: AssetAttachment) => Promise<void>;
    onTagSearch?: (tag: AssetTag) => Promise<void>;
  } = $props();

  let panel = $state<AssetDetailPanel>('none');
  let title = $state('');
  let description = $state('');
  let parentAssetId = $state<string | null>(null);
  let moveParentSearch = $state('');
  let checkoutDetails = $state('');
  let customFieldValues = $state<Record<string, string>>({});
  let selectedTagIds = $state<string[]>([]);
  let newTags = $state<AssetTagDraft[]>([]);
  let saveError = $state('');
  let photoUploadError = $state('');
  let fileError = $state<AssetFilesError | null>(null);
  let failedPhotoUpload = $state<SelectedAttachment | null>(null);
  let photoUploading = $state(false);
  let photoInput = $state<HTMLInputElement | null>(null);
  let fileInput = $state<HTMLInputElement | null>(null);
  let selectedAttachment = $state<AssetAttachment | null>(null);
  let selectedPhotoId = $state<string | null>(null);
  let lastRouteActionKey = $state('');
  let actionReturnFocus = $state<HTMLElement | null>(null);
  let actionReturnHref = $state('');
  let removePhotoButton = $state<HTMLElement | null>(null);
  let applicableFields = $derived(applicableCustomFieldDefinitions(customFieldDefinitions, asset.customAssetTypeId));
  let imageContentTypes = $derived(mediaPolicy.supportedContentTypes.filter((contentType) => contentType.startsWith('image/')));
  let photoAttachments = $derived(attachments.filter((attachment) => attachment.contentType.startsWith('image/')));
  let fileAttachments = $derived(attachments.filter((attachment) => !attachment.contentType.startsWith('image/')));
  let detailPhotos = $derived(buildDetailPhotos(asset, photoAttachments));
  let heroPhoto = $derived(detailPhotos.find((photo) => photo.id === selectedPhotoId) ?? detailPhotos.find((photo) => photo.isPrimary) ?? detailPhotos[0]);
  let heroPhotoAttachment = $derived(heroPhoto ? photoAttachments.find((attachment) => attachment.id === heroPhoto.id) ?? null : null);
  let canAddPhoto = $derived(canEdit && asset.lifecycleState === 'active' && !saving && !photoUploading && imageContentTypes.length > 0);
  let editUnavailableStatus = $derived(assetEditUnavailableStatus(canEdit));
  let descriptionText = $derived(assetDescriptionText(asset.description));
  let photoUploadDisabledReason = $derived(
    photoUploading
      ? 'Photo upload is already in progress.'
      : photoUploadUnavailableReason({
          canEditAsset: canEdit,
          lifecycleState: asset.lifecycleState,
          isSaving: saving,
          supportedImageTypeCount: imageContentTypes.length
        })
  );
  let photoUploadDescribedBy = $derived(
    [
      photoUploadDisabledReason ? PHOTO_UPLOAD_DISABLED_REASON_ID : '',
      photoUploadError ? PHOTO_UPLOAD_ERROR_ID : ''
    ].filter(Boolean).join(' ')
  );
  let displayFields = $derived(
    customFieldDefinitions.filter(
      (definition) =>
        definition.applicability === 'all_assets' ||
        (!!asset.customAssetTypeId && definition.customAssetTypeIds.includes(asset.customAssetTypeId))
    )
  );
  let detailFieldGroups = $derived(partitionAssetDetailFields(displayFields, asset.customFields));

  $effect(() => {
    const attachmentKey = attachmentAction === 'delete' ? attachments.map((attachment) => attachment.id).join(',') : 'none';
    const actionKey = `${asset.id}:${action ?? 'none'}:${attachmentAction ?? 'none'}:${attachmentId ?? 'none'}:${attachmentKey}`;
    if (actionKey === lastRouteActionKey) {
      return;
    }
    const initializingWithoutRouteAction = lastRouteActionKey === '' && !action && !attachmentAction;
    lastRouteActionKey = actionKey;
    saveError = '';
    if (attachmentAction === 'delete' && canEdit) {
      const routeAttachment = attachments.find((attachment) => attachment.id === attachmentId) ?? null;
      selectedAttachment = routeAttachment;
      actionReturnHref = routeAttachment ? attachmentDeleteHref(routeAttachment) : '';
      panel = routeAttachment ? 'attachment-delete' : 'none';
    } else if (action === 'edit' && canEdit && asset.lifecycleState === 'active') {
      actionReturnHref = actionHref('edit');
      openEdit(false);
    } else if (action === 'move' && canEdit && asset.lifecycleState === 'active') {
      actionReturnHref = actionHref('move');
      openMove(false);
    } else if (action === 'move-here' && canEdit && asset.kind === 'container' && asset.lifecycleState === 'active') {
      panel = 'none';
    } else if (action === 'archive' && canEdit && asset.lifecycleState === 'active') {
      actionReturnHref = actionHref('archive');
      panel = 'archive';
    } else if (action === 'restore' && canEdit && asset.lifecycleState === 'archived') {
      actionReturnHref = actionHref('restore');
      panel = 'restore';
    } else if (action === 'delete' && canEdit) {
      actionReturnHref = actionHref('delete');
      panel = 'delete';
    } else if (action === 'checkout' && actionIsAvailable('checkout')) {
      actionReturnHref = actionHref('checkout');
      checkoutDetails = '';
      panel = 'checkout';
    } else if (action === 'return' && actionIsAvailable('return')) {
      actionReturnHref = actionHref('return');
      checkoutDetails = '';
      panel = 'return';
    } else if (!action && !attachmentAction && !initializingWithoutRouteAction) {
      panel = 'none';
      selectedAttachment = null;
      actionReturnHref = '';
    }
  });

  $effect(() => {
    if (selectedPhotoId && !detailPhotos.some((photo) => photo.id === selectedPhotoId)) {
      selectedPhotoId = null;
    }
  });

  function openEdit(notify = true): void {
    saveError = '';
    title = asset.title;
    description = asset.description;
    parentAssetId = asset.parentAssetId;
    customFieldValues = Object.fromEntries(
      applicableFields.map((field) => [field.key, stringifyCustomFieldValue(asset.customFields?.[field.key])])
    );
    selectedTagIds = asset.tags?.map((tag) => tag.id) ?? [];
    newTags = [];
    panel = 'edit';
    if (notify) {
      onActionOpen('edit');
    }
  }

  function openMove(notify = true): void {
    saveError = '';
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
    saveError = '';
    panel = 'archive';
    onActionOpen('archive');
  }

  function openRestore(): void {
    saveError = '';
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
    saveError = '';
    actionReturnFocus = event.currentTarget instanceof HTMLElement ? event.currentTarget : null;
    actionReturnHref = actionHref(nextAction);
    if (nextAction === 'edit') {
      openEdit();
    } else if (nextAction === 'move') {
      openMove();
    } else if (nextAction === 'archive') {
      openArchive();
    } else if (nextAction === 'restore') {
      openRestore();
    } else if (nextAction === 'checkout') {
      checkoutDetails = '';
      panel = 'checkout';
      onActionOpen('checkout');
    } else if (nextAction === 'return') {
      checkoutDetails = '';
      panel = 'return';
      onActionOpen('return');
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
        customFields: buildCustomFields(),
        tagIds: selectedTagIds,
        newTags
      });
      closePanel();
    } catch (caught) {
      saveError = caught instanceof Error ? caught.message : 'Unable to save asset.';
    }
  }

  function closePanel(): void {
    const previousPanel = panel;
    const restorePhotoFocus = previousPanel === 'attachment-delete' && selectedAttachment?.contentType.startsWith('image/');
    panel = 'none';
    saveError = '';
    selectedAttachment = null;
    if (
      previousPanel === 'edit' ||
      previousPanel === 'move' ||
      previousPanel === 'archive' ||
      previousPanel === 'restore' ||
      previousPanel === 'delete' ||
      previousPanel === 'checkout' ||
      previousPanel === 'return'
    ) {
      onActionClose();
    } else if (previousPanel === 'attachment-delete') {
      onAttachmentDeleteClose();
      if (restorePhotoFocus) void tick().then(() => removePhotoButton?.focus());
    }
  }

  function closeAction(event: MouseEvent): void {
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    saveError = '';
    closePanel();
  }

  function restoreActionFocus(event: Event): void {
    event.preventDefault();
    const target = actionReturnFocus;
    const href = actionReturnHref;
    actionReturnFocus = null;
    actionReturnHref = '';
    void tick().then(() => {
      const hrefFallback = href
        ? Array.from(document.querySelectorAll<HTMLElement>('[href]')).find(
            (candidate) => candidate.getAttribute('href') === href
          ) ?? null
        : null;
      const pageFallback = document.querySelector<HTMLElement>('#asset-title, main h1');
      const focusTarget = target?.isConnected ? target : hrefFallback ?? pageFallback;
      if (focusTarget && focusTarget === pageFallback && !focusTarget.hasAttribute('tabindex')) {
        focusTarget.setAttribute('tabindex', '-1');
      }
      focusTarget?.focus();
    });
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

  async function checkout(): Promise<void> {
    saveError = '';
    try {
      await onCheckout(checkoutDetails.trim());
      closePanel();
    } catch (caught) {
      saveError = caught instanceof Error ? caught.message : 'Unable to checkout asset.';
    }
  }

  async function returnAsset(): Promise<void> {
    saveError = '';
    try {
      await onReturn(checkoutDetails.trim());
      closePanel();
    } catch (caught) {
      saveError = caught instanceof Error ? caught.message : 'Unable to return asset.';
    }
  }

  async function archiveAttachment(attachment: AssetAttachment): Promise<void> {
    fileError = null;
    try {
      await onArchiveAttachment(attachment);
    } catch (caught) {
      fileError = {
        operation: 'archive',
        attachmentId: attachment.id,
        message: userSafeMediaErrorMessage(caught, 'Unable to archive file.')
      };
    }
  }

  async function uploadAttachment(event: Event): Promise<void> {
    fileError = null;
    const input = event.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) {
      return;
    }
    const contentType = file.type;
    if (!supportedAttachmentContentType(mediaPolicy.supportedContentTypes, contentType)) {
      fileError = { operation: 'upload', message: unsupportedAttachmentTypeMessage() };
      input.value = '';
      return;
    }
    if (file.size > mediaPolicy.maxBytes) {
      fileError = { operation: 'upload', message: `Attachment must be ${formatBytes(mediaPolicy.maxBytes)} or smaller.` };
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
      fileError = { operation: 'upload', message: userSafeMediaErrorMessage(caught, 'Unable to upload file.') };
      input.value = '';
    }
  }

  async function uploadPhotoAttachment(event: Event): Promise<void> {
    photoUploadError = '';
    failedPhotoUpload = null;
    const input = event.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) {
      return;
    }
    const contentType = file.type;
    if (!supportedImageContentType(imageContentTypes, contentType)) {
      photoUploadError = unsupportedImageTypeMessage();
      input.value = '';
      return;
    }
    if (file.size > mediaPolicy.maxBytes) {
      photoUploadError = `Attachment must be ${formatBytes(mediaPolicy.maxBytes)} or smaller.`;
      input.value = '';
      return;
    }
    const attachment: SelectedAttachment = {
      id: createClientAttachmentId(),
      name: file.name,
      sizeBytes: file.size,
      contentType,
      file
    };
    await uploadPhoto(attachment);
    input.value = '';
  }

  async function uploadPhoto(attachment: SelectedAttachment): Promise<void> {
    if (photoUploading) return;
    photoUploading = true;
    photoUploadError = '';
    try {
      await onUploadAttachment(attachment);
      failedPhotoUpload = null;
    } catch (caught) {
      failedPhotoUpload = attachment;
      photoUploadError = userSafeMediaErrorMessage(caught, 'Unable to upload photo.');
    } finally {
      photoUploading = false;
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
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    actionReturnFocus = event.currentTarget instanceof HTMLElement ? event.currentTarget : null;
    actionReturnHref = attachmentDeleteHref(attachment);
    selectedAttachment = attachment;
    panel = 'attachment-delete';
    onAttachmentDeleteOpen(attachment.id);
  }

  function setCustomFieldValue(key: string, value: string): void {
    customFieldValues = { ...customFieldValues, [key]: value };
  }

  function setSelectedTagIds(ids: string[]): void {
    selectedTagIds = ids;
  }

  function setNewTags(tags: AssetTagDraft[]): void {
    newTags = tags;
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
      uploadError={photoUploadError}
      uploadBusy={photoUploading}
      retryPhotoName={failedPhotoUpload?.name ?? ''}
      removePhotoHref={heroPhotoAttachment && canEdit && asset.lifecycleState === 'active' && !saving ? attachmentDeleteHref(heroPhotoAttachment) : ''}
      bind:removePhotoButton
      onChoosePhoto={() => photoInput?.click()}
      onSelectPhoto={(photoId) => { selectedPhotoId = photoId; }}
      onRetryPhoto={() => { if (failedPhotoUpload) void uploadPhoto(failedPhotoUpload); }}
      onRemovePhoto={(event) => { if (heroPhotoAttachment) openAttachmentDelete(event, heroPhotoAttachment); }}
    >
      <div class="asset-detail-copy">
        <div class="detail-title-row">
          <div>
            <h1 id="asset-title" tabindex="-1">{asset.title}</h1>
            <p>{asset.containmentTrail}</p>
            <AssetTagChips tags={asset.tags ?? []} onTagSelect={onTagSearch} />
          </div>
	          <span class="detail-title-badges">
	            {#if asset.currentCheckout}
	              <CheckoutBadge checkout={asset.currentCheckout} />
	            {/if}
	            {#if !asset.currentCheckout || asset.lifecycleState === 'archived'}
	              <Badge variant={asset.lifecycleState === 'active' ? 'secondary' : 'outline'}>{asset.lifecycleState === 'active' ? 'Active' : 'Archived'}</Badge>
	            {/if}
	          </span>
        </div>
	        <dl class="detail-list">
	          <div><dt>Kind</dt><dd>{assetKindLabel(asset.kind)}</dd></div>
	          {#if asset.customAssetTypeLabel}<div><dt>Type</dt><dd>{asset.customAssetTypeLabel}</dd></div>{/if}
	          {#if asset.currentCheckout}
	            <div><dt>Checkout</dt><dd>{new Date(asset.currentCheckout.checkedOutAt).toLocaleString()}</dd></div>
	          {/if}
	          <div><dt>Updated</dt><dd>{asset.updatedAt ? new Date(asset.updatedAt).toLocaleString() : 'Not available'}</dd></div>
        </dl>
        <div class="detail-actions">
	          {#if asset.currentCheckout}
	            <Button.Root
	              href={actionHref('return')}
	              class="availability-action"
	              disabled={!actionIsAvailable('return')}
	              onclick={(event) => openAction(event, 'return')}
	            ><Undo2 /> Return</Button.Root>
	          {:else}
	            <Button.Root
	              href={actionHref('checkout')}
	              class="availability-action"
	              disabled={!actionIsAvailable('checkout')}
	              onclick={(event) => openAction(event, 'checkout')}
	            ><LogOut /> Check out</Button.Root>
	          {/if}
          <Button.Root class="maintenance-action" href={actionHref('edit')} variant="outline" disabled={!actionIsAvailable('edit')} onclick={(event) => openAction(event, 'edit')}><Pencil /> Edit</Button.Root>
	          <Button.Root
	            href={actionHref('move')}
	            class="maintenance-action"
	            variant="outline"
	            disabled={!actionIsAvailable('move')}
	            onclick={(event) => openAction(event, 'move')}
	          ><MoveRight /> Move</Button.Root>
          {#if asset.lifecycleState === 'active'}
            <Button.Root
              href={actionHref('archive')}
              class="lifecycle-action"
              variant="ghost"
              disabled={!actionIsAvailable('archive')}
              onclick={(event) => openAction(event, 'archive')}
            ><Archive /> Archive</Button.Root>
          {:else}
            <Button.Root
              href={actionHref('restore')}
              class="lifecycle-action"
              variant="ghost"
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
    {#if asset.kind === 'container'}
      <ContainedAssetWorkspace
        target={asset}
        assets={workspaceAssets}
        {canCreate}
        {canEdit}
        {saving}
        moveHereOpen={action === 'move-here'}
        {onOpenAsset}
        {onOpenAdd}
        onOpenMoveHere={() => onActionOpen('move-here')}
        onCloseMoveHere={onActionClose}
        {onMoveHere}
      />
    {/if}
  <div class="asset-detail-sections">
      <AssetDetailActionPanel
        {panel}
        {asset}
        {parentTargets}
        {selectedAttachment}
        {saving}
        {saveError}
        detailHref={detailHref()}
        {applicableFields}
        assetTags={assetTags}
        selectedTagIds={selectedTagIds}
        {newTags}
        bind:title
        bind:description
        bind:parentAssetId
        bind:moveParentSearch
        bind:checkoutDetails
        {customFieldValues}
        onClose={closeAction}
        onDismiss={closePanel}
        onCloseAutoFocus={restoreActionFocus}
        onSave={save}
        onArchive={archive}
        onRestore={restore}
        onDelete={remove}
        onCheckout={checkout}
        onReturn={returnAsset}
        onDeleteAttachment={removeAttachment}
        onParentSelect={selectMoveParent}
        onCustomFieldValueChange={setCustomFieldValue}
        onSelectedTagIdsChange={setSelectedTagIds}
        onNewTagsChange={setNewTags}
      />
    <section class="detail-section" aria-labelledby="asset-description-title">
      <h2 id="asset-description-title">Details</h2>
      <p>{descriptionText}</p>
      {#if detailFieldGroups.populated.length > 0}
        <dl class="detail-list custom-detail-list" aria-label="Custom field values">
          {#each detailFieldGroups.populated as field}
            <div>
              <dt>{field.displayName}</dt>
              <dd>{stringifyCustomFieldValue(asset.customFields?.[field.key])}</dd>
            </div>
          {/each}
        </dl>
      {/if}
      {#if detailFieldGroups.unset.length > 0}
        <details class="unset-field-disclosure">
          <summary>Show {detailFieldGroups.unset.length} unset {detailFieldGroups.unset.length === 1 ? 'field' : 'fields'}</summary>
          <dl class="detail-list custom-detail-list" aria-label="Unset custom fields">
            {#each detailFieldGroups.unset as field}
              <div><dt>{field.displayName}</dt><dd>Not set</dd></div>
            {/each}
          </dl>
        </details>
      {/if}
    </section>
    <AssetFilesSection
      attachments={fileAttachments}
      {canEdit}
      {saving}
      active={asset.lifecycleState === 'active'}
      error={fileError}
      onChooseFile={() => fileInput?.click()}
      onArchiveAttachment={(attachment) => { void archiveAttachment(attachment); }}
      onOpenAttachmentDelete={openAttachmentDelete}
      {attachmentDeleteHref}
    />
    <section class="detail-section" aria-labelledby="asset-checkout-history-title">
      <h2 id="asset-checkout-history-title">Checkout history</h2>
      {#if checkoutHistory.length === 0}
        <p>No checkout history.</p>
      {:else}
        <div class="asset-list compact-list checkout-history-list" aria-label="Checkout history">
          {#each checkoutHistory as checkout}
            <div class="history-row">
              <div>
                <strong>{checkout.state === 'returned' ? 'Returned' : checkout.state === 'undone' ? 'Undone' : 'Checked out'}</strong>
                <small>{new Date(checkout.checkedOutAt).toLocaleString()}</small>
                {#if checkout.checkoutDetails}
                  <small>{checkout.checkoutDetails}</small>
                {/if}
                {#if checkout.returnedAt}
                  <small>Returned {new Date(checkout.returnedAt).toLocaleString()}</small>
                {/if}
                {#if checkout.returnDetails}
                  <small>{checkout.returnDetails}</small>
                {/if}
                <details class="history-technical">
                  <summary>Technical details</summary>
                  <code>Checked out by {checkout.checkedOutByPrincipalId}{#if checkout.returnedByPrincipalId}; returned by {checkout.returnedByPrincipalId}{/if}</code>
                </details>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </section>
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

<style>
  .detail-title-badges {
    display: inline-flex;
    flex-wrap: wrap;
    gap: 0.4rem;
    justify-content: flex-end;
  }

  .history-row {
    border-bottom: 1px solid var(--border);
    padding: var(--space-3) var(--space-4);
  }

  .history-row:last-child {
    border-bottom: 0;
  }

  .history-row > div {
    display: grid;
    gap: var(--space-1);
  }

  .history-row small {
    color: var(--muted-foreground);
    font-size: var(--text-metadata-size);
    line-height: var(--text-metadata-line-height);
  }

  .history-technical {
    margin-top: var(--space-1);
    color: var(--muted-foreground);
    font-size: var(--text-metadata-size);
    line-height: var(--text-metadata-line-height);
  }

  .history-technical summary {
    width: fit-content;
    cursor: pointer;
    font-weight: 600;
  }

  .history-technical code {
    display: block;
    margin-top: var(--space-2);
    overflow-wrap: anywhere;
  }

  .checkout-history-list {
    border-radius: var(--radius-surface);
  }

  .unset-field-disclosure {
    margin-top: 0.75rem;
  }

  .unset-field-disclosure summary {
    width: fit-content;
    min-height: 44px;
    display: flex;
    align-items: center;
    color: var(--color-muted-foreground, #64748b);
    cursor: pointer;
    font-weight: 600;
  }
</style>
