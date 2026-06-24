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
  import type { AssetViewModel, UpdateAssetDraft } from '$lib/domain/inventory';
  import { assetKindLabel } from '$lib/domain/inventory';
  import AssetThumb from './AssetThumb.svelte';

  let {
    asset,
    canEdit,
    parentTargets,
    saving,
    onBack,
    onSave,
    onArchive,
    onRestore,
    onDelete
  }: {
    asset: AssetViewModel;
    canEdit: boolean;
    parentTargets: AssetViewModel[];
    saving: boolean;
    onBack: () => void;
    onSave: (draft: UpdateAssetDraft) => Promise<void>;
    onArchive: () => Promise<void>;
    onRestore: () => Promise<void>;
    onDelete: () => Promise<void>;
  } = $props();

  let panel = $state<'none' | 'edit' | 'move' | 'delete'>('none');
  let title = $state('');
  let description = $state('');
  let parentAssetId = $state<string | null>(null);
  let saveError = $state('');

  function openEdit(): void {
    title = asset.title;
    description = asset.description;
    parentAssetId = asset.parentAssetId;
    panel = 'edit';
  }

  function openMove(): void {
    title = asset.title;
    description = asset.description;
    parentAssetId = asset.parentAssetId;
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
        parentAssetId
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
