<script lang="ts" module>
  import type { AssetAttachment, AssetTag, AssetTagDraft, AssetViewModel, CustomFieldDefinition, ParentTargetViewModel } from '$lib/domain/inventory';

  export type AssetDetailPanel = 'none' | 'edit' | 'move' | 'archive' | 'restore' | 'delete' | 'checkout' | 'return' | 'attachment-delete';

  export type AssetDetailActionPanelProps = {
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
  import AssetTagSelector from './AssetTagSelector.svelte';
  import CustomFieldControls from './CustomFieldControls.svelte';
  import ParentTargetPicker from './ParentTargetPicker.svelte';
  import WorkspaceConfirmationDialog from './action-surface/WorkspaceConfirmationDialog.svelte';
  import WorkspaceTaskSheet from './action-surface/WorkspaceTaskSheet.svelte';

  let {
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
      return title !== asset.title || description !== asset.description ||
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
  <WorkspaceTaskSheet open title="Edit asset" description="Update the name, details, fields, and tags." busy={saving} dismissible={!taskDirty} closeHref={detailHref} closeLabel="Close edit" initialFocusSelector="#edit-asset-title" onCloseLink={onClose} onOpenChange={(open) => { if (!open) onDismiss(); }} {onCloseAutoFocus}>
    <div class="field-stack">
      <Label for="edit-asset-title">Name</Label>
      <Input id="edit-asset-title" bind:value={title} />
    </div>
    <div class="field-stack">
      <Label for="edit-asset-description">Description</Label>
      <Textarea id="edit-asset-description" bind:value={description} />
    </div>
    {#if populatedFields.length > 0}
      <CustomFieldControls
        fields={populatedFields}
        values={customFieldValues}
        idPrefix="edit-custom-field"
        label="Details"
        onValueChange={onCustomFieldValueChange}
      />
    {/if}
    {#if emptyFields.length > 0}
      <details class="edit-empty-fields">
        <summary>Show {emptyFields.length} empty {emptyFields.length === 1 ? 'field' : 'fields'}</summary>
        <CustomFieldControls
          fields={emptyFields}
          values={customFieldValues}
          idPrefix="edit-custom-field"
          label="Empty details"
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
      <Button.Root href={detailHref} variant="outline" disabled={saving} onclick={onClose}>Cancel</Button.Root>
      <Button.Root disabled={saving || title.trim().length === 0 || !taskDirty} onclick={() => { void onSave(); }}>Save</Button.Root>
    {/snippet}
  </WorkspaceTaskSheet>
{:else if panel === 'move'}
  <WorkspaceTaskSheet open title="Move asset" description={`Choose a new place for ${asset.title}.`} busy={saving} dismissible={!taskDirty} closeHref={detailHref} closeLabel="Close move" initialFocusSelector="#move-parent-search" onCloseLink={onClose} onOpenChange={(open) => { if (!open) onDismiss(); }} {onCloseAutoFocus}>
    <ParentTargetPicker
      legend="Parent"
      searchId="move-parent-search"
      groupLabel="Move target"
      bind:search={moveParentSearch}
      selectedId={parentAssetId}
      targets={parentTargets}
      onSelect={onParentSelect}
    />
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
    {#snippet footer()}
      <Button.Root href={detailHref} variant="outline" disabled={saving} onclick={onClose}>Cancel</Button.Root>
      <Button.Root disabled={saving || !taskDirty} onclick={() => { void onSave(); }}>Move</Button.Root>
    {/snippet}
  </WorkspaceTaskSheet>
{:else if panel === 'archive'}
  <WorkspaceConfirmationDialog open title="Archive asset" description={`Move ${asset.title} out of active browsing?`} busy={saving} onOpenChange={(open) => { if (!open) onDismiss(); }} {onCloseAutoFocus}>
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
    {#snippet cancel()}<Button.Root href={detailHref} variant="outline" disabled={saving} onclick={onClose}>Cancel</Button.Root>{/snippet}
    {#snippet action()}<Button.Root variant="outline" disabled={saving} onclick={() => { void onArchive(); }}>Archive</Button.Root>{/snippet}
  </WorkspaceConfirmationDialog>
{:else if panel === 'restore'}
  <WorkspaceConfirmationDialog open title="Restore asset" description={`Return ${asset.title} to active browsing?`} busy={saving} onOpenChange={(open) => { if (!open) onDismiss(); }} {onCloseAutoFocus}>
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
    {#snippet cancel()}<Button.Root href={detailHref} variant="outline" disabled={saving} onclick={onClose}>Cancel</Button.Root>{/snippet}
    {#snippet action()}<Button.Root disabled={saving} onclick={() => { void onRestore(); }}>Restore</Button.Root>{/snippet}
  </WorkspaceConfirmationDialog>
{:else if panel === 'delete'}
  <WorkspaceConfirmationDialog open title="Delete asset" description={`Delete ${asset.title} permanently?`} busy={saving} onOpenChange={(open) => { if (!open) onDismiss(); }} {onCloseAutoFocus}>
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
    {#snippet cancel()}<Button.Root href={detailHref} variant="outline" disabled={saving} onclick={onClose}>Cancel</Button.Root>{/snippet}
    {#snippet action()}<Button.Root variant="destructive" disabled={saving} onclick={() => { void onDelete(); }}>Delete</Button.Root>{/snippet}
  </WorkspaceConfirmationDialog>
{:else if panel === 'checkout'}
  <WorkspaceTaskSheet open title="Check out asset" description={`${asset.title} will stay in its home location and be marked as checked out.`} busy={saving} dismissible={!taskDirty} closeHref={detailHref} closeLabel="Close check out" initialFocusSelector="#checkout-asset-details" onCloseLink={onClose} onOpenChange={(open) => { if (!open) onDismiss(); }} {onCloseAutoFocus}>
    <div class="field-stack">
      <Label for="checkout-asset-details">Details</Label>
      <Textarea id="checkout-asset-details" bind:value={checkoutDetails} placeholder="Optional: using at desk, loaned to Sam" />
    </div>
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
    {#snippet footer()}
      <Button.Root href={detailHref} variant="outline" disabled={saving} onclick={onClose}>Cancel</Button.Root>
      <Button.Root disabled={saving} onclick={() => { void onCheckout(); }}>Check out</Button.Root>
    {/snippet}
  </WorkspaceTaskSheet>
{:else if panel === 'return'}
  <WorkspaceTaskSheet open title="Return asset" description={`Mark ${asset.title} as returned.`} busy={saving} dismissible={!taskDirty} closeHref={detailHref} closeLabel="Close return" initialFocusSelector="#return-asset-details" onCloseLink={onClose} onOpenChange={(open) => { if (!open) onDismiss(); }} {onCloseAutoFocus}>
    <div class="field-stack">
      <Label for="return-asset-details">Details</Label>
      <Textarea id="return-asset-details" bind:value={checkoutDetails} placeholder="Optional: back in bin, returned by Alex" />
    </div>
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
    {#snippet footer()}
      <Button.Root href={detailHref} variant="outline" disabled={saving} onclick={onClose}>Cancel</Button.Root>
      <Button.Root disabled={saving} onclick={() => { void onReturn(); }}>Return</Button.Root>
    {/snippet}
  </WorkspaceTaskSheet>
{:else if panel === 'attachment-delete' && selectedAttachment}
  <WorkspaceConfirmationDialog open title="Delete attachment" description={`Delete ${selectedAttachment.fileName} permanently?`} busy={saving} onOpenChange={(open) => { if (!open) onDismiss(); }} {onCloseAutoFocus}>
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
    {#snippet cancel()}<Button.Root href={detailHref} variant="outline" disabled={saving} onclick={onClose}>Cancel</Button.Root>{/snippet}
    {#snippet action()}<Button.Root variant="destructive" disabled={saving} onclick={() => { void onDeleteAttachment(); }}>Delete</Button.Root>{/snippet}
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
