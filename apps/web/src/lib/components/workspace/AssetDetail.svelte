<script lang="ts">
  import ArrowLeft from '@lucide/svelte/icons/arrow-left';
  import Archive from '@lucide/svelte/icons/archive';
  import MoveRight from '@lucide/svelte/icons/move-right';
  import Pencil from '@lucide/svelte/icons/pencil';
  import RotateCcw from '@lucide/svelte/icons/rotate-ccw';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { Textarea } from '$lib/components/ui/textarea/index.js';
  import type { AssetAttachment, AssetViewModel, CustomFieldDefinition, UpdateAssetDraft } from '$lib/domain/inventory';
  import { applicableCustomFieldDefinitions, assetKindLabel } from '$lib/domain/inventory';
  import AssetThumb from './AssetThumb.svelte';

  let {
    asset,
    canEdit,
    parentTargets,
    customFieldDefinitions,
    saving,
    attachments,
    onBack,
    onSave,
    onArchive,
    onRestore,
    onDelete,
    onArchiveAttachment,
    onDeleteAttachment
  }: {
    asset: AssetViewModel;
    canEdit: boolean;
    parentTargets: AssetViewModel[];
    customFieldDefinitions: CustomFieldDefinition[];
    saving: boolean;
    attachments: AssetAttachment[];
    onBack: () => void;
    onSave: (draft: UpdateAssetDraft) => Promise<void>;
    onArchive: () => Promise<void>;
    onRestore: () => Promise<void>;
    onDelete: () => Promise<void>;
    onArchiveAttachment: (attachment: AssetAttachment) => Promise<void>;
    onDeleteAttachment: (attachment: AssetAttachment) => Promise<void>;
  } = $props();

  let panel = $state<'none' | 'edit' | 'move' | 'delete' | 'attachment-delete'>('none');
  let title = $state('');
  let description = $state('');
  let parentAssetId = $state<string | null>(null);
  let customFieldValues = $state<Record<string, string>>({});
  let saveError = $state('');
  let selectedAttachment = $state<AssetAttachment | null>(null);
  let applicableFields = $derived(applicableCustomFieldDefinitions(customFieldDefinitions, asset.customAssetTypeId));
  let displayFields = $derived(
    customFieldDefinitions.filter(
      (definition) =>
        definition.applicability === 'all_assets' ||
        (!!asset.customAssetTypeId && definition.customAssetTypeIds.includes(asset.customAssetTypeId))
    )
  );

  function openEdit(): void {
    title = asset.title;
    description = asset.description;
    parentAssetId = asset.parentAssetId;
    customFieldValues = Object.fromEntries(
      applicableFields.map((field) => [field.key, stringifyCustomFieldValue(asset.customFields?.[field.key])])
    );
    panel = 'edit';
  }

  function openMove(): void {
    title = asset.title;
    description = asset.description;
    parentAssetId = asset.parentAssetId;
    customFieldValues = Object.fromEntries(
      applicableFields.map((field) => [field.key, stringifyCustomFieldValue(asset.customFields?.[field.key])])
    );
    panel = 'move';
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
      panel = 'none';
    } catch (caught) {
      saveError = caught instanceof Error ? caught.message : 'Unable to save asset.';
    }
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

  async function removeAttachment(): Promise<void> {
    if (!selectedAttachment) {
      return;
    }
    saveError = '';
    try {
      await onDeleteAttachment(selectedAttachment);
      selectedAttachment = null;
      panel = 'none';
    } catch (caught) {
      saveError = caught instanceof Error ? caught.message : 'Unable to delete attachment.';
    }
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
</script>

<section class="workspace-main detail-view" aria-labelledby="asset-title">
  <Button.Root variant="ghost" class="back-button" onclick={onBack}><ArrowLeft /> Back</Button.Root>
  <div class="asset-detail-grid">
    <AssetThumb asset={asset} size="lg" />
    <div class="asset-detail-copy">
      <div class="detail-title-row">
        <div>
          <h1 id="asset-title">{asset.title}</h1>
          <p>{asset.containmentTrail}</p>
        </div>
        <Badge variant={asset.lifecycleState === 'active' ? 'secondary' : 'outline'}>{asset.lifecycleState}</Badge>
      </div>
      <p>{asset.description || 'No description.'}</p>
      <dl class="detail-list">
        <div><dt>Kind</dt><dd>{assetKindLabel(asset.kind)}</dd></div>
        <div><dt>Type</dt><dd>{asset.customAssetTypeLabel ?? 'Base asset'}</dd></div>
        <div><dt>Updated</dt><dd>{asset.updatedAt ? new Date(asset.updatedAt).toLocaleString() : 'Not available'}</dd></div>
      </dl>
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
      <div class="detail-actions">
        <Button.Root disabled={!canEdit || asset.lifecycleState !== 'active'} onclick={openEdit}><Pencil /> Edit</Button.Root>
        <Button.Root variant="outline" disabled={!canEdit || asset.lifecycleState !== 'active'} onclick={openMove}><MoveRight /> Move</Button.Root>
        {#if asset.lifecycleState === 'active'}
          <Button.Root variant="outline" disabled={!canEdit || saving} onclick={() => { void archive(); }}><Archive /> Archive</Button.Root>
        {:else}
          <Button.Root variant="outline" disabled={!canEdit || saving} onclick={() => { void restore(); }}><RotateCcw /> Restore</Button.Root>
        {/if}
      </div>
      {#if !canEdit}
        <p class="denied-note">Edit actions require asset edit access.</p>
      {/if}
      <section class="attachment-section" aria-labelledby="attachments-title">
        <div class="section-heading compact">
          <h2 id="attachments-title">Attachments</h2>
        </div>
        {#if attachments.length === 0}
          <div class="empty-state">
            <p>No active attachments.</p>
          </div>
        {:else}
          <div class="asset-list">
            {#each attachments as attachment}
              <div class="attachment-row">
                {#if attachment.thumbnailUrl}
                  <img src={attachment.thumbnailUrl} alt={attachment.fileName} />
                {:else}
                  <div class="asset-thumb asset-thumb-sm"><Archive aria-hidden="true" /></div>
                {/if}
                <span class="asset-row-main">
                  <strong>{attachment.fileName}</strong>
                  <small>{attachment.contentType} / {formatBytes(attachment.sizeBytes)}</small>
                </span>
                <div class="attachment-actions">
                  <Button.Root variant="outline" disabled={!canEdit || saving} onclick={() => { void archiveAttachment(attachment); }}>Archive</Button.Root>
                  <Button.Root
                    variant="destructive"
                    disabled={!canEdit || saving}
                    onclick={() => { selectedAttachment = attachment; panel = 'attachment-delete'; }}
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
        <div class="detail-action-panel" aria-label="Edit asset">
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
            <Button.Root variant="outline" onclick={() => { panel = 'none'; }}>Cancel</Button.Root>
            <Button.Root disabled={saving || title.trim().length === 0} onclick={() => { void save(); }}>Save</Button.Root>
          </div>
          {#if saveError}
            <p class="denied-note" role="alert">{saveError}</p>
          {/if}
        </div>
      {:else if panel === 'move'}
        <div class="detail-action-panel" aria-label="Move asset">
          <div class="field-stack">
            <Label>Parent</Label>
            <div class="parent-picker" role="group" aria-label="Move target">
              <Button.Root
                variant={parentAssetId === null ? 'secondary' : 'outline'}
                aria-pressed={parentAssetId === null}
                onclick={() => { parentAssetId = null; }}
              >
                Inventory root
              </Button.Root>
              {#each parentTargets as target}
                <Button.Root
                  variant={parentAssetId === target.id ? 'secondary' : 'outline'}
                  aria-pressed={parentAssetId === target.id}
                  onclick={() => { parentAssetId = target.id; }}
                >
                  {target.title}
                </Button.Root>
              {/each}
            </div>
          </div>
          <div class="tray-actions">
            <Button.Root variant="outline" onclick={() => { panel = 'none'; }}>Cancel</Button.Root>
            <Button.Root disabled={saving} onclick={() => { void save(); }}>Move</Button.Root>
          </div>
          {#if saveError}
            <p class="denied-note" role="alert">{saveError}</p>
          {/if}
        </div>
      {:else if panel === 'delete'}
        <div class="detail-action-panel" aria-label="Delete asset">
          <p>Delete {asset.title} permanently?</p>
          <div class="tray-actions">
            <Button.Root variant="outline" onclick={() => { panel = 'none'; }}>Cancel</Button.Root>
            <Button.Root variant="destructive" disabled={saving} onclick={() => { void remove(); }}>Delete</Button.Root>
          </div>
          {#if saveError}
            <p class="denied-note" role="alert">{saveError}</p>
          {/if}
        </div>
      {:else if panel === 'attachment-delete' && selectedAttachment}
        <div class="detail-action-panel" aria-label="Delete attachment">
          <p>Delete {selectedAttachment.fileName} permanently?</p>
          <div class="tray-actions">
            <Button.Root variant="outline" onclick={() => { panel = 'none'; selectedAttachment = null; }}>Cancel</Button.Root>
            <Button.Root variant="destructive" disabled={saving} onclick={() => { void removeAttachment(); }}>Delete</Button.Root>
          </div>
          {#if saveError}
            <p class="denied-note" role="alert">{saveError}</p>
          {/if}
        </div>
      {/if}
      <div class="danger-zone" aria-label="Danger area">
        <div>
          <strong>Permanent deletion</strong>
          <p>Remove this asset from the inventory permanently.</p>
        </div>
        <Button.Root variant="destructive" disabled={!canEdit || saving} onclick={() => { panel = 'delete'; }}><Trash2 /> Delete</Button.Root>
      </div>
    </div>
  </div>
</section>
