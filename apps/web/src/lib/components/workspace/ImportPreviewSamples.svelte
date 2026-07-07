<script lang="ts">
  import CheckCircle2 from '@lucide/svelte/icons/check-circle-2';
  import type { ImportJobPreview } from '$lib/domain/inventory';
  import { fileSizeLabel, previewAssetContext, previewLocationContext } from './importWorkspacePresentation';

  type Props = {
    preview: ImportJobPreview;
  };

  let { preview }: Props = $props();
</script>

<div class="preview-samples">
  <section>
    <div class="sample-heading">
      <h3>Fields</h3>
      {#if preview.fieldsTruncated}<small>Showing a sample</small>{/if}
    </div>
    <div class="sample-list">
      {#each preview.fields as field}
        <div class="sample-row">
          <span>{field.displayName || field.key}</span>
          <small>{field.key} · {field.type}</small>
        </div>
      {/each}
      {#if preview.fields.length === 0}
        <div class="quiet-row"><CheckCircle2 size={16} aria-hidden="true" /> No custom fields planned.</div>
      {/if}
    </div>
  </section>
  <section>
    <div class="sample-heading">
      <h3>Locations</h3>
      {#if preview.locationsTruncated}<small>Showing a sample</small>{/if}
    </div>
    <div class="sample-list">
      {#each preview.locations as item}
        <div class="sample-row">
          <span>{item.title}</span>
          <small>{previewLocationContext(item)}</small>
        </div>
      {/each}
      {#if preview.locations.length === 0}
        <div class="quiet-row"><CheckCircle2 size={16} aria-hidden="true" /> No locations planned.</div>
      {/if}
    </div>
  </section>
  <section>
    <div class="sample-heading">
      <h3>Assets</h3>
      {#if preview.assetsTruncated}<small>Showing a sample</small>{/if}
    </div>
    <div class="sample-list">
      {#each preview.assets as item}
        <div class="sample-row">
          <span>{item.title}</span>
          <small>{previewAssetContext(item)}</small>
        </div>
      {/each}
      {#if preview.assets.length === 0}
        <div class="quiet-row"><CheckCircle2 size={16} aria-hidden="true" /> No asset records planned.</div>
      {/if}
    </div>
  </section>
  <section>
    <div class="sample-heading">
      <h3>Photos/files</h3>
      {#if preview.attachmentsTruncated}<small>Showing a sample</small>{/if}
    </div>
    <div class="sample-list">
      {#each preview.attachments as attachment}
        <div class="sample-row">
          <span>{attachment.fileName || 'Unnamed attachment'}</span>
          <small>{attachment.contentType || 'unknown type'} · {fileSizeLabel(attachment.sizeBytes)}{attachment.primary ? ' · primary' : ''}</small>
        </div>
      {/each}
      {#if preview.attachments.length === 0}
        <div class="quiet-row"><CheckCircle2 size={16} aria-hidden="true" /> No photos or files planned.</div>
      {/if}
    </div>
  </section>
</div>

<style>
  .preview-samples {
    display: grid;
    gap: 0.75rem;
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .preview-samples section {
    border-top: 1px solid var(--border);
    display: grid;
    gap: 0.6rem;
    min-width: 0;
    padding-top: 0.75rem;
  }

  .preview-samples section:nth-child(-n + 3) {
    border-top: 0;
    padding-top: 0;
  }

  .sample-heading {
    align-items: baseline;
    display: flex;
    gap: 0.5rem;
    justify-content: space-between;
  }

  .sample-heading small,
  .sample-row small {
    color: var(--muted-foreground);
    font-size: 0.78rem;
  }

  .sample-list {
    display: grid;
    gap: 0.45rem;
  }

  .sample-row {
    min-width: 0;
  }

  .sample-row span,
  .sample-row small {
    display: block;
    overflow-wrap: anywhere;
  }

  .quiet-row {
    align-items: center;
    display: flex;
    gap: 0.75rem;
  }

  h3 {
    font-size: 1rem;
    margin: 0;
  }

  @media (max-width: 860px) {
    .preview-samples {
      grid-template-columns: 1fr;
    }

    .preview-samples section,
    .preview-samples section:nth-child(-n + 3) {
      border-top: 1px solid var(--border);
      padding-top: 0.75rem;
    }

    .preview-samples section:first-child {
      border-top: 0;
      padding-top: 0;
    }
  }
</style>
