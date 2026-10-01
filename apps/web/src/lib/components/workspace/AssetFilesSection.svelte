<script lang="ts" module>
  import { t } from '$lib/presentation/localization';
  export type AssetFilesError =
    | { operation: 'upload'; message: string }
    | { operation: 'archive'; attachmentId: string; message: string };
</script>

<script lang="ts">
  import FileText from '@lucide/svelte/icons/file-text';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import Upload from '@lucide/svelte/icons/upload';
  import * as Button from '$lib/components/ui/button/index.js';
  import { assetFilesStatus } from '$lib/application/workspaceAssetDetail';
  import type { AssetAttachment } from '$lib/domain/inventory';
  import { formatBytes } from './formatBytes';

  let {
    attachments,
    titleId = 'asset-files-title',
    canEdit,
    saving,
    active,
    error = null,
    onChooseFile,
    onArchiveAttachment,
    onOpenAttachmentDelete,
    attachmentDeleteHref,
    loading = false
  }: {
    attachments: AssetAttachment[];
    titleId?: string;
    canEdit: boolean;
    saving: boolean;
    active: boolean;
    error?: AssetFilesError | null;
    onChooseFile: () => void;
    onArchiveAttachment: (attachment: AssetAttachment) => void;
    onOpenAttachmentDelete: (event: MouseEvent, attachment: AssetAttachment) => void;
    attachmentDeleteHref: (attachment: AssetAttachment) => string;
    loading?: boolean;
  } = $props();

  let status = $derived(assetFilesStatus(attachments.length));
  let errorId = $derived(`${titleId}-error`);
</script>

<section class="attachment-section" aria-labelledby={titleId}>
  <div class="section-heading compact">
    <h2 id={titleId}>{t('web.AssetFilesSection.files')}</h2>
    <div class="attachment-upload">
      <Button.Root
        variant="outline"
        disabled={!canEdit || !active || saving}
        aria-describedby={error?.operation === 'upload' ? errorId : undefined}
        onclick={onChooseFile}
      >
        <Upload /> {t('web.AssetFilesSection.uploadFile')} </Button.Root>
    </div>
  </div>
  {#if error}<p id={errorId} class="denied-note" role="alert">{error.message}</p>{/if}
  {#if loading}
    <p role="status">{t('web.AssetFilesSection.loadingFiles')}</p>
  {:else if status}
    <div class="empty-state">
      <p>{status.message}</p>
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
            <Button.Root
              variant="outline"
              disabled={!canEdit || saving}
              aria-describedby={error?.operation === 'archive' && error.attachmentId === attachment.id ? errorId : undefined}
              onclick={() => onArchiveAttachment(attachment)}
            >{t('web.AssetFilesSection.archive')}</Button.Root>
            <Button.Root
              href={attachmentDeleteHref(attachment)}
              variant="destructive"
              disabled={!canEdit || saving}
              onclick={(event) => onOpenAttachmentDelete(event, attachment)}
            >
              <Trash2 /> {t('web.AssetFilesSection.delete')} </Button.Root>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</section>
