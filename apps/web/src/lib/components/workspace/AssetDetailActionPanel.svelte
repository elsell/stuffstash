<script lang="ts" module>
  import { t } from '$lib/presentation/localization';
  import type { CustomAssetType, AssetExpiration, AssetAttachment, AssetTag, AssetTagDraft, AssetViewModel, CustomFieldDefinition, ParentTargetViewModel } from '$lib/domain/inventory';

  export type AssetDetailPanel = 'none' | 'edit' | 'move' | 'archive' | 'restore' | 'delete' | 'checkout' | 'return' | 'attachment-delete';

  export type AssetDetailActionPanelProps = {
    customAssetTypes?: CustomAssetType[];
    customAssetTypeId?: string;
    onCustomTypeSelect?: (id: string) => void;
    expiration?: AssetExpiration;
    expirationValid?: boolean;
    expirationEnabled?: boolean;
    onExpirationChange?: (value: AssetExpiration | undefined, valid: boolean) => void;
    panel: AssetDetailPanel;
    asset: AssetViewModel;
    parentTargets: ParentTargetViewModel[];
    selectedAttachment: AssetAttachment | null;
    saving: boolean;
    saveError: string;
    detailHref: string;
    applicableFields: CustomFieldDefinition[];
    assetTags?: AssetTag[];
    selectedTagIds?: string[];
    newTags?: AssetTagDraft[];
    title: string;
    description: string;
    parentAssetId: string | null;
    moveParentSearch: string;
    checkoutDetails: string;
    customFieldValues: Record<string, string>;
    onClose: (event: MouseEvent) => void;
    onDismiss: () => void;
    onCloseAutoFocus: (event: Event) => void;
    onSave: () => Promise<void>;
    onArchive: () => Promise<void>;
    onRestore: () => Promise<void>;
    onDelete: () => Promise<void>;
    onCheckout: () => Promise<void>;
    onReturn: () => Promise<void>;
    onDeleteAttachment: () => Promise<void>;
    onParentSelect: (id: string | null) => void;
    onCustomFieldValueChange: (key: string, value: string) => void;
    onSelectedTagIdsChange?: (ids: string[]) => void;
    onNewTagsChange?: (tags: AssetTagDraft[]) => void;
  };
</script>

<script lang="ts">
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { Textarea } from '$lib/components/ui/textarea/index.js';
  import ChoiceGrid from './ChoiceGrid.svelte';
  import ExpirationField from './ExpirationField.svelte';
  import AssetTagSelector from './AssetTagSelector.svelte';
  import CustomFieldControls from './CustomFieldControls.svelte';
  import ParentTargetPicker from './ParentTargetPicker.svelte';
  import WorkspaceConfirmationDialog from './action-surface/WorkspaceConfirmationDialog.svelte';
  import WorkspaceTaskSheet from './action-surface/WorkspaceTaskSheet.svelte';

  let {
    customAssetTypes = [],
    customAssetTypeId,
    onCustomTypeSelect = () => {},
    expiration,
    expirationValid = true,
    expirationEnabled = false,
    onExpirationChange = () => {},
    panel,
    asset,
    parentTargets,
    selectedAttachment,
    saving,
    saveError,
    detailHref,
    applicableFields,
    assetTags = [],
    selectedTagIds = [],
    newTags = [],
    title = $bindable(),
    description = $bindable(),
    parentAssetId = $bindable(),
    moveParentSearch = $bindable(),
    checkoutDetails = $bindable(),
    customFieldValues,
    onClose,
    onDismiss,
    onCloseAutoFocus,
    onSave,
    onArchive,
    onRestore,
    onDelete,
    onCheckout,
    onReturn,
    onDeleteAttachment,
    onParentSelect,
    onCustomFieldValueChange,
    onSelectedTagIdsChange = () => {},
    onNewTagsChange = () => {}
  }: AssetDetailActionPanelProps = $props();

  let taskDirty = $derived.by(() => {
    if (panel === 'edit') {
      const currentTagIds = (asset.tags ?? []).map((tag) => tag.id).sort().join(',');
      const nextTagIds = [...selectedTagIds].sort().join(',');
      return (!asset.customAssetTypeId && !!customAssetTypeId) || expiration?.date !== asset.expiration?.date || expiration?.precision !== asset.expiration?.precision ||
        title !== asset.title || description !== asset.description ||
        applicableFields.some((field) => String(asset.customFields?.[field.key] ?? '') !== (customFieldValues[field.key] ?? '')) ||
        currentTagIds !== nextTagIds || newTags.length > 0;
    }
    if (panel === 'move') return parentAssetId !== asset.parentAssetId;
    if (panel === 'checkout' || panel === 'return') return checkoutDetails.trim().length > 0;
    return false;
  });
  let populatedFields = $derived(
    applicableFields.filter((field) => String(asset.customFields?.[field.key] ?? '').trim().length > 0)
  );
  let emptyFields = $derived(
    applicableFields.filter((field) => String(asset.customFields?.[field.key] ?? '').trim().length === 0)
  );
</script>

{#if panel === 'edit'}
  <WorkspaceTaskSheet open title={t('web.AssetDetailActionPanel.editAsset')} description={t('web.AssetDetailActionPanel.updateTheNameDetailsFieldsAndTags')} busy={saving} dismissible={!taskDirty} closeHref={detailHref} closeLabel={t('web.AssetDetailActionPanel.closeEdit')} initialFocusSelector="#edit-asset-title" onCloseLink={onClose} onOpenChange={(open) => { if (!open) onDismiss(); }} {onCloseAutoFocus}>
    <div class="field-stack">
      <Label for="edit-asset-title">{t('web.AssetDetailActionPanel.name')}</Label>
      <Input id="edit-asset-title" bind:value={title} />
    </div>
    <div class="field-stack">
      <Label for="edit-asset-description">{t('web.AssetDetailActionPanel.description')}</Label>
      <Textarea id="edit-asset-description" bind:value={description} />
    </div>
    {#if !asset.customAssetTypeId && customAssetTypes.some((type) => type.lifecycleState === 'active')}
      <fieldset>
        <legend>{t('web.AssetDetailActionPanel.customType')}</legend>
        <ChoiceGrid label={t('web.AssetDetailActionPanel.customAssetType')} options={[{ value: '', label: 'Base asset' }, ...customAssetTypes.filter((type) => type.lifecycleState === 'active').map((type) => ({ value: type.id, label: type.displayName }))]} selectedValues={[customAssetTypeId ?? '']} onSelect={onCustomTypeSelect} />
      </fieldset>
    {/if}
    {#if expirationEnabled}
      {#key `${asset.id}:${customAssetTypeId ?? asset.customAssetTypeId ?? ''}`}
        <ExpirationField id="edit-asset-expiration" initialValue={asset.expiration} onChange={onExpirationChange} />
      {/key}
    {:else if asset.expiration}
      <p>{t('web.AssetDetailActionPanel.expirationTrackingIsDisabledForThisTypeFull', { date: asset.expiration.date })}</p>
      {#if expiration}<Button.Root type="button" variant="ghost" onclick={() => onExpirationChange(undefined, true)}>{t('web.AssetDetailActionPanel.clearExpiration')}</Button.Root>{/if}
    {/if}
    {#if populatedFields.length > 0}
      <CustomFieldControls
        fields={populatedFields}
        values={customFieldValues}
        idPrefix="edit-custom-field"
        label={t('web.AssetDetailActionPanel.details')}
        onValueChange={onCustomFieldValueChange}
      />
    {/if}
    {#if emptyFields.length > 0}
      <details class="edit-empty-fields">
        <summary>{t('fields.showEmpty', { count: emptyFields.length })}</summary>
        <CustomFieldControls
          fields={emptyFields}
          values={customFieldValues}
          idPrefix="edit-custom-field"
          label={t('web.AssetDetailActionPanel.emptyDetails')}
          onValueChange={onCustomFieldValueChange}
        />
      </details>
    {/if}
    <AssetTagSelector
      tags={assetTags}
      selectedIds={selectedTagIds}
      {newTags}
      onSelectedIdsChange={onSelectedTagIdsChange}
      onNewTagsChange={onNewTagsChange}
    />
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
    {#snippet footer()}
      <Button.Root href={detailHref} variant="outline" disabled={saving} onclick={onClose}>{t('web.AssetDetailActionPanel.cancel')}</Button.Root>
      <Button.Root disabled={saving || !expirationValid || title.trim().length === 0 || !taskDirty} onclick={() => { void onSave(); }}>{t('web.AssetDetailActionPanel.save')}</Button.Root>
    {/snippet}
  </WorkspaceTaskSheet>
{:else if panel === 'move'}
  <WorkspaceTaskSheet open title={asset.kind === 'location' ? t('web.AssetDetailActionPanel.movePlace') : t('web.AssetDetailActionPanel.moveAsset')} description={t('web.AssetDetailActionPanel.chooseANewPlaceFor', { title: String(asset.title) })} busy={saving} dismissible={!taskDirty} closeHref={detailHref} closeLabel={t('web.AssetDetailActionPanel.closeMove')} initialFocusSelector="#move-parent-search" onCloseLink={onClose} onOpenChange={(open) => { if (!open) onDismiss(); }} {onCloseAutoFocus}>
    <ParentTargetPicker
      legend="Parent"
      searchId="move-parent-search"
      groupLabel={t('web.AssetDetailActionPanel.moveTarget')}
      bind:search={moveParentSearch}
      selectedId={parentAssetId}
      targets={parentTargets}
      onSelect={onParentSelect}
    />
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
    {#snippet footer()}
      <Button.Root href={detailHref} variant="outline" disabled={saving} onclick={onClose}>{t('web.AssetDetailActionPanel.cancel')}</Button.Root>
      <Button.Root disabled={saving || !taskDirty} onclick={() => { void onSave(); }}>{t('web.AssetDetailActionPanel.move')}</Button.Root>
    {/snippet}
  </WorkspaceTaskSheet>
{:else if panel === 'archive'}
  <WorkspaceConfirmationDialog open title={t('web.AssetDetailActionPanel.archiveAsset')} description={t('web.AssetDetailActionPanel.moveOutOfActiveBrowsing', { title: String(asset.title) })} busy={saving} onOpenChange={(open) => { if (!open) onDismiss(); }} {onCloseAutoFocus}>
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
    {#snippet cancel()}<Button.Root href={detailHref} variant="outline" disabled={saving} onclick={onClose}>{t('web.AssetDetailActionPanel.cancel')}</Button.Root>{/snippet}
    {#snippet action()}<Button.Root variant="outline" disabled={saving} onclick={() => { void onArchive(); }}>{t('web.AssetDetailActionPanel.archive')}</Button.Root>{/snippet}
  </WorkspaceConfirmationDialog>
{:else if panel === 'restore'}
  <WorkspaceConfirmationDialog open title={t('web.AssetDetailActionPanel.restoreAsset')} description={t('web.AssetDetailActionPanel.returnToActiveBrowsing', { title: String(asset.title) })} busy={saving} onOpenChange={(open) => { if (!open) onDismiss(); }} {onCloseAutoFocus}>
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
    {#snippet cancel()}<Button.Root href={detailHref} variant="outline" disabled={saving} onclick={onClose}>{t('web.AssetDetailActionPanel.cancel')}</Button.Root>{/snippet}
    {#snippet action()}<Button.Root disabled={saving} onclick={() => { void onRestore(); }}>{t('web.AssetDetailActionPanel.restore')}</Button.Root>{/snippet}
  </WorkspaceConfirmationDialog>
{:else if panel === 'delete'}
  <WorkspaceConfirmationDialog open title={t('web.AssetDetailActionPanel.deleteAsset')} description={t('web.AssetDetailActionPanel.deletePermanently', { title: String(asset.title) })} busy={saving} onOpenChange={(open) => { if (!open) onDismiss(); }} {onCloseAutoFocus}>
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
    {#snippet cancel()}<Button.Root href={detailHref} variant="outline" disabled={saving} onclick={onClose}>{t('web.AssetDetailActionPanel.cancel')}</Button.Root>{/snippet}
    {#snippet action()}<Button.Root variant="destructive" disabled={saving} onclick={() => { void onDelete(); }}>{t('web.AssetDetailActionPanel.delete')}</Button.Root>{/snippet}
  </WorkspaceConfirmationDialog>
{:else if panel === 'checkout'}
  <WorkspaceTaskSheet open title={t('web.AssetDetailActionPanel.checkOutAsset')} description={t('web.AssetDetailActionPanel.willStayInItsHomeLocationAndBeMarked', { title: String(asset.title) })} busy={saving} dismissible={!taskDirty} closeHref={detailHref} closeLabel={t('web.AssetDetailActionPanel.closeCheckOut')} initialFocusSelector="#checkout-asset-details" onCloseLink={onClose} onOpenChange={(open) => { if (!open) onDismiss(); }} {onCloseAutoFocus}>
    <div class="field-stack">
      <Label for="checkout-asset-details">{t('web.AssetDetailActionPanel.details')}</Label>
      <Textarea id="checkout-asset-details" bind:value={checkoutDetails} placeholder={t('web.AssetDetailActionPanel.optionalUsingAtDeskLoanedToSam')} />
    </div>
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
    {#snippet footer()}
      <Button.Root href={detailHref} variant="outline" disabled={saving} onclick={onClose}>{t('web.AssetDetailActionPanel.cancel')}</Button.Root>
      <Button.Root disabled={saving} onclick={() => { void onCheckout(); }}>{t('web.AssetDetailActionPanel.checkOut')}</Button.Root>
    {/snippet}
  </WorkspaceTaskSheet>
{:else if panel === 'return'}
  <WorkspaceTaskSheet open title={t('web.AssetDetailActionPanel.returnAsset')} description={t('web.AssetDetailActionPanel.markAsReturned', { title: String(asset.title) })} busy={saving} dismissible={!taskDirty} closeHref={detailHref} closeLabel={t('web.AssetDetailActionPanel.closeReturn')} initialFocusSelector="#return-asset-details" onCloseLink={onClose} onOpenChange={(open) => { if (!open) onDismiss(); }} {onCloseAutoFocus}>
    <div class="field-stack">
      <Label for="return-asset-details">{t('web.AssetDetailActionPanel.details')}</Label>
      <Textarea id="return-asset-details" bind:value={checkoutDetails} placeholder={t('web.AssetDetailActionPanel.optionalBackInBinReturnedByAlex')} />
    </div>
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
    {#snippet footer()}
      <Button.Root href={detailHref} variant="outline" disabled={saving} onclick={onClose}>{t('web.AssetDetailActionPanel.cancel')}</Button.Root>
      <Button.Root disabled={saving} onclick={() => { void onReturn(); }}>{t('web.AssetDetailActionPanel.return')}</Button.Root>
    {/snippet}
  </WorkspaceTaskSheet>
{:else if panel === 'attachment-delete' && selectedAttachment}
  <WorkspaceConfirmationDialog open title={t('web.AssetDetailActionPanel.deleteAttachment')} description={t('web.AssetDetailActionPanel.deletePermanently2', { fileName: String(selectedAttachment.fileName) })} busy={saving} onOpenChange={(open) => { if (!open) onDismiss(); }} {onCloseAutoFocus}>
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
    {#snippet cancel()}<Button.Root href={detailHref} variant="outline" disabled={saving} onclick={onClose}>{t('web.AssetDetailActionPanel.cancel')}</Button.Root>{/snippet}
    {#snippet action()}<Button.Root variant="destructive" disabled={saving} onclick={() => { void onDeleteAttachment(); }}>{t('web.AssetDetailActionPanel.delete')}</Button.Root>{/snippet}
  </WorkspaceConfirmationDialog>
{/if}

<style>
  .edit-empty-fields {
    border-top: 1px solid var(--border);
    padding-top: 16px;
  }

  .edit-empty-fields summary {
    display: flex;
    min-height: 44px;
    width: fit-content;
    align-items: center;
    color: var(--muted-foreground);
    cursor: pointer;
    font-weight: 600;
  }

  .edit-empty-fields[open] summary {
    margin-bottom: 16px;
  }
</style>
