<script lang="ts">
  import { tick } from 'svelte';
  import type { ImportJob, ImportJobCancellationMode } from '$lib/domain/inventory';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import { historyCountSummary, sourceDescription, statusLabel } from './importWorkspacePresentation';

  type Props = {
    cancelJob: ImportJob | null;
    removeJob: ImportJob | null;
    busy: boolean;
    onCancelJob: (job: ImportJob, mode: ImportJobCancellationMode) => void;
    onDismissCancel: () => void;
    onRemoveJob: (job: ImportJob) => void;
    onDismissRemove: () => void;
  };

  let { cancelJob, removeJob, busy, onCancelJob, onDismissCancel, onRemoveJob, onDismissRemove }: Props = $props();
  let cancelKeepChoiceElement = $state<HTMLButtonElement | null>(null);

  $effect(() => {
    if (!cancelJob) return;
    void tick().then(() => {
      if (cancelJob) {
        cancelKeepChoiceElement?.focus();
      }
    });
  });
</script>

{#if cancelJob}
  <Card.Root role="group" aria-labelledby="cancel-import-heading">
    <Card.Header>
      <Card.Title id="cancel-import-heading">Cancel {cancelJob.source.name}?</Card.Title>
      <Card.Description>
        {sourceDescription(cancelJob)} · {cancelJob.progress.message || statusLabel(cancelJob)}
      </Card.Description>
    </Card.Header>
    <Card.Content class="confirmation-choice-grid">
      <Button.Root
        bind:ref={cancelKeepChoiceElement}
        variant="outline"
        class="confirmation-choice"
        type="button"
        disabled={busy}
        onclick={() => onCancelJob(cancelJob, 'keep_partial_progress')}
      >
        <strong>Keep imported items</strong>
        <span>Stop future work and leave anything already imported in the inventory.</span>
      </Button.Root>
      <Button.Root
        variant="outline"
        class="confirmation-choice danger"
        type="button"
        disabled={busy}
        onclick={() => onCancelJob(cancelJob, 'discard_partial_progress')}
      >
        <strong>Discard imported items</strong>
        <span>Stop future work and remove records created by this job. Audit history remains.</span>
      </Button.Root>
      <div class="action-row">
        <Button.Root variant="outline" onclick={onDismissCancel} disabled={busy}>Keep running</Button.Root>
      </div>
    </Card.Content>
  </Card.Root>
{:else if removeJob}
  <Card.Root role="group" aria-labelledby="remove-import-heading">
    <Card.Header>
      <Card.Title id="remove-import-heading">Remove {removeJob.source.name} from history?</Card.Title>
      <Card.Description>
        Imported records and audit history will remain. This only removes the run from the import history list.
      </Card.Description>
    </Card.Header>
    <Card.Content class="confirmation-content">
      <div class="confirmation-topline">
        <div>
          <strong>{statusLabel(removeJob)}</strong>
          <span>{historyCountSummary(removeJob)} · {sourceDescription(removeJob)}</span>
        </div>
      </div>
      <div class="action-row">
        <Button.Root variant="destructive" onclick={() => onRemoveJob(removeJob)} disabled={busy}>
          Remove from history
        </Button.Root>
        <Button.Root variant="outline" onclick={onDismissRemove} disabled={busy}>Keep in history</Button.Root>
      </div>
    </Card.Content>
  </Card.Root>
{/if}

<style>
  .action-row {
    align-items: center;
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem;
  }

  :global(.confirmation-choice-grid),
  :global(.confirmation-content) {
    display: grid;
    gap: 1rem;
  }

  :global(.confirmation-choice-grid) {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  :global(.confirmation-choice) {
    align-content: flex-start;
    display: grid;
    gap: 0.75rem;
    height: auto;
    min-height: 6.5rem;
    padding: 1rem;
    text-align: left;
    white-space: normal;
  }

  :global(.confirmation-choice.danger) {
    border-color: color-mix(in oklab, var(--destructive) 45%, transparent);
  }

  :global(.confirmation-choice.danger:hover),
  :global(.confirmation-choice.danger:focus-visible) {
    background: color-mix(in oklab, var(--destructive) 8%, transparent);
    color: var(--foreground);
  }

  :global(.confirmation-choice strong),
  :global(.confirmation-choice span) {
    display: block;
  }

  :global(.confirmation-choice span) {
    color: var(--muted-foreground);
    font-size: 0.85rem;
    font-weight: 400;
    line-height: 1.35;
  }

  .confirmation-topline {
    align-items: center;
    display: flex;
    gap: 0.75rem;
    justify-content: space-between;
  }

  .confirmation-topline span {
    color: var(--muted-foreground);
    display: block;
    font-size: 0.85rem;
    margin-top: 0.15rem;
  }

  @media (max-width: 860px) {
    :global(.confirmation-choice-grid) {
      grid-template-columns: 1fr;
    }

    .confirmation-topline {
      align-items: flex-start;
      flex-direction: column;
    }
  }
</style>
