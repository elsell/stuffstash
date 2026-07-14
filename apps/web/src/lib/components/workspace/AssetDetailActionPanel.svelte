<script lang="ts" module>
  import type { AssetAttachment, AssetTag, AssetTagDraft, AssetViewModel, CustomFieldDefinition, ParentTargetViewModel } from '$lib/domain/inventory';

  export type AssetDetailPanel = 'none' | 'edit' | 'move' | 'archive' | 'restore' | 'delete' | 'checkout' | 'return' | 'attachment-delete';

  export type AssetDetailActionPanelProps = {
    panel: AssetDetailPanel;
    panelElement: HTMLElement | null;
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

  let {
    panel,
    panelElement = $bindable(),
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
</script>

{#if panel === 'edit'}
  <section
    bind:this={panelElement}
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
    <CustomFieldControls
      fields={applicableFields}
      values={customFieldValues}
      idPrefix="edit-custom-field"
      label="Edit custom fields"
      onValueChange={onCustomFieldValueChange}
    />
    <AssetTagSelector
      tags={assetTags}
      selectedIds={selectedTagIds}
      {newTags}
      onSelectedIdsChange={onSelectedTagIdsChange}
      onNewTagsChange={onNewTagsChange}
    />
    <div class="tray-actions">
      <Button.Root href={detailHref} variant="outline" onclick={onClose}>Cancel</Button.Root>
      <Button.Root disabled={saving || title.trim().length === 0} onclick={() => { void onSave(); }}>Save</Button.Root>
    </div>
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
  </section>
{:else if panel === 'move'}
  <section
    bind:this={panelElement}
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
      onSelect={onParentSelect}
    />
    <div class="tray-actions">
      <Button.Root href={detailHref} variant="outline" onclick={onClose}>Cancel</Button.Root>
      <Button.Root disabled={saving} onclick={() => { void onSave(); }}>Move</Button.Root>
    </div>
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
  </section>
{:else if panel === 'archive'}
  <section
    bind:this={panelElement}
    class="detail-action-panel"
    aria-labelledby="archive-asset-panel-title"
    tabindex="-1"
  >
    <h2 id="archive-asset-panel-title">Archive asset</h2>
    <p>Move {asset.title} out of active browsing?</p>
    <div class="tray-actions">
      <Button.Root href={detailHref} variant="outline" onclick={onClose}>Cancel</Button.Root>
      <Button.Root variant="outline" disabled={saving} onclick={() => { void onArchive(); }}>Archive</Button.Root>
    </div>
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
  </section>
{:else if panel === 'restore'}
  <section
    bind:this={panelElement}
    class="detail-action-panel"
    aria-labelledby="restore-asset-panel-title"
    tabindex="-1"
  >
    <h2 id="restore-asset-panel-title">Restore asset</h2>
    <p>Return {asset.title} to active browsing?</p>
    <div class="tray-actions">
      <Button.Root href={detailHref} variant="outline" onclick={onClose}>Cancel</Button.Root>
      <Button.Root disabled={saving} onclick={() => { void onRestore(); }}>Restore</Button.Root>
    </div>
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
  </section>
{:else if panel === 'delete'}
  <section
    bind:this={panelElement}
    class="detail-action-panel"
    aria-labelledby="delete-asset-panel-title"
    tabindex="-1"
  >
    <h2 id="delete-asset-panel-title">Delete asset</h2>
    <p>Delete {asset.title} permanently?</p>
    <div class="tray-actions">
      <Button.Root href={detailHref} variant="outline" onclick={onClose}>Cancel</Button.Root>
      <Button.Root variant="destructive" disabled={saving} onclick={() => { void onDelete(); }}>Delete</Button.Root>
    </div>
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
  </section>
{:else if panel === 'checkout'}
  <section
    bind:this={panelElement}
    class="detail-action-panel"
    aria-labelledby="checkout-asset-panel-title"
    tabindex="-1"
  >
    <h2 id="checkout-asset-panel-title">Check out asset</h2>
    <p>{asset.title} will stay in its home location and be marked as checked out.</p>
    <div class="field-stack">
      <Label for="checkout-asset-details">Details</Label>
      <Textarea id="checkout-asset-details" bind:value={checkoutDetails} placeholder="Optional: using at desk, loaned to Sam" />
    </div>
    <div class="tray-actions">
      <Button.Root href={detailHref} variant="outline" onclick={onClose}>Cancel</Button.Root>
      <Button.Root disabled={saving} onclick={() => { void onCheckout(); }}>Check out</Button.Root>
    </div>
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
  </section>
{:else if panel === 'return'}
  <section
    bind:this={panelElement}
    class="detail-action-panel"
    aria-labelledby="return-asset-panel-title"
    tabindex="-1"
  >
    <h2 id="return-asset-panel-title">Return asset</h2>
    <p>Mark {asset.title} as returned.</p>
    <div class="field-stack">
      <Label for="return-asset-details">Details</Label>
      <Textarea id="return-asset-details" bind:value={checkoutDetails} placeholder="Optional: back in bin, returned by Alex" />
    </div>
    <div class="tray-actions">
      <Button.Root href={detailHref} variant="outline" onclick={onClose}>Cancel</Button.Root>
      <Button.Root disabled={saving} onclick={() => { void onReturn(); }}>Return</Button.Root>
    </div>
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
  </section>
{:else if panel === 'attachment-delete' && selectedAttachment}
  <section
    bind:this={panelElement}
    class="detail-action-panel"
    aria-labelledby="delete-attachment-panel-title"
    tabindex="-1"
  >
    <h2 id="delete-attachment-panel-title">Delete attachment</h2>
    <p>Delete {selectedAttachment.fileName} permanently?</p>
    <div class="tray-actions">
      <Button.Root href={detailHref} variant="outline" onclick={onClose}>Cancel</Button.Root>
      <Button.Root variant="destructive" disabled={saving} onclick={() => { void onDeleteAttachment(); }}>Delete</Button.Root>
    </div>
    {#if saveError}
      <p class="denied-note" role="alert">{saveError}</p>
    {/if}
  </section>
{/if}
