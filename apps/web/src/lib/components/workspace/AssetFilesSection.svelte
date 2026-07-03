<script lang="ts">
  import FileText from '@lucide/svelte/icons/file-text';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import Upload from '@lucide/svelte/icons/upload';
  import * as Button from '$lib/components/ui/button/index.js';
  import type { AssetAttachment } from '$lib/domain/inventory';
  import { formatBytes } from './formatBytes';

  let {
    attachments,
    titleId = 'asset-files-title',
    canEdit,
    saving,
    active,
    onChooseFile,
    onArchiveAttachment,
    onOpenAttachmentDelete,
    attachmentDeleteHref
  }: {
    attachments: AssetAttachment[];
    titleId?: string;
    canEdit: boolean;
    saving: boolean;
    active: boolean;
    onChooseFile: () => void;
    onArchiveAttachment: (attachment: AssetAttachment) => void;
    onOpenAttachmentDelete: (event: MouseEvent, attachment: AssetAttachment) => void;
    attachmentDeleteHref: (attachment: AssetAttachment) => string;
  } = $props();

</script>

<section class="attachment-section" aria-labelledby={titleId}>
  <div class="section-heading compact">
    <h2 id={titleId}>Files</h2>
    <div class="attachment-upload">
      <Button.Root
        variant="outline"
        disabled={!canEdit || !active || saving}
        onclick={onChooseFile}
      >
        <Upload /> Upload file
      </Button.Root>
    </div>
  </div>
  {#if attachments.length === 0}
    <div class="empty-state">
      <p>No active files.</p>
    </div>
  {:else}
    <div class="asset-list">
      {#each attachments as attachment}
        <div class="attachment-row">
          <div class="asset-thumb asset-thumb-sm"><FileText aria-hidden="true" /></div>
          <span class="asset-row-main">
            <strong>{attachment.fileName}</strong>
            <small>{attachment.contentType} / {formatBytes(attachment.sizeBytes)}</small>
          </span>
          <div class="attachment-actions">
            <Button.Root variant="outline" disabled={!canEdit || saving} onclick={() => onArchiveAttachment(attachment)}>Archive</Button.Root>
            <Button.Root
              href={attachmentDeleteHref(attachment)}
              variant="destructive"
              disabled={!canEdit || saving}
              onclick={(event) => onOpenAttachmentDelete(event, attachment)}
            >
              <Trash2 /> Delete
            </Button.Root>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</section>
