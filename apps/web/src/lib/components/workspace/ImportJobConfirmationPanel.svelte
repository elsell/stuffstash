<script lang="ts">
  import type { ImportJob, ImportJobCancellationMode } from '$lib/domain/inventory';
  import * as Button from '$lib/components/ui/button/index.js';
  import { historyCountSummary, sourceDescription, statusLabel } from './importWorkspacePresentation';
  import WorkspaceConfirmationDialog from './action-surface/WorkspaceConfirmationDialog.svelte';

  type Props = {
    cancelJob: ImportJob | null;
    removeJob: ImportJob | null;
    busy: boolean;
    error?: string;
    onCancelJob: (job: ImportJob, mode: ImportJobCancellationMode) => void;
    onDismissCancel: () => void;
    onRemoveJob: (job: ImportJob) => void;
    onDismissRemove: () => void;
  };

  let {
    cancelJob,
    removeJob,
    busy,
    error = '',
    onCancelJob,
    onDismissCancel,
    onRemoveJob,
    onDismissRemove
  }: Props = $props();
  let pendingAction = $state<'keep' | 'discard' | 'remove' | null>(null);

  $effect(() => {
    if (!busy) pendingAction = null;
  });
</script>

{#if cancelJob}
  <WorkspaceConfirmationDialog
    open
    title={`Cancel ${cancelJob.source.name}?`}
    description={`${sourceDescription(cancelJob)} · ${cancelJob.progress.message || statusLabel(cancelJob)}`}
    {busy}
    onOpenChange={(open) => { if (!open) onDismissCancel(); }}
  >
    {#snippet children()}
      {#if error}<p class="denied-note" role="alert">{error}</p>{/if}
    {/snippet}
    {#snippet cancel()}
      <Button.Root variant="outline" class="min-h-11" onclick={onDismissCancel} disabled={busy} autofocus>Keep running</Button.Root>
    {/snippet}
    {#snippet action()}
      <div class="confirmation-choice-grid">
        <Button.Root
          variant="outline"
          class="confirmation-choice"
          type="button"
          disabled={busy}
          onclick={() => {
            pendingAction = 'keep';
            onCancelJob(cancelJob, 'keep_partial_progress');
          }}
        >
          <strong><Button.BusyContent busy={pendingAction === 'keep' && busy} label="Keep imported items" busyLabel="Cancelling import" /></strong>
          <span>Stop future work and leave anything already imported in the inventory.</span>
        </Button.Root>
        <Button.Root
          variant="destructive"
          class="confirmation-choice danger"
          type="button"
          disabled={busy}
          onclick={() => {
            pendingAction = 'discard';
            onCancelJob(cancelJob, 'discard_partial_progress');
          }}
        >
          <strong><Button.BusyContent busy={pendingAction === 'discard' && busy} label="Discard imported items" busyLabel="Cancelling import" /></strong>
          <span>Stop future work and remove records created by this job. Audit history remains.</span>
        </Button.Root>
      </div>
    {/snippet}
  </WorkspaceConfirmationDialog>
{:else if removeJob}
  <WorkspaceConfirmationDialog
    open
    title={`Remove ${removeJob.source.name} from history?`}
    description="Imported records and audit history will remain. This only removes the run from the import history list."
    {busy}
    onOpenChange={(open) => { if (!open) onDismissRemove(); }}
  >
    {#snippet children()}
      <div class="confirmation-topline">
        <div>
          <strong>{statusLabel(removeJob)}</strong>
          <span>{historyCountSummary(removeJob)} · {sourceDescription(removeJob)}</span>
        </div>
      </div>
      {#if error}<p class="denied-note" role="alert">{error}</p>{/if}
    {/snippet}
    {#snippet cancel()}
      <Button.Root variant="outline" class="min-h-11" onclick={onDismissRemove} disabled={busy} autofocus>Keep in history</Button.Root>
    {/snippet}
    {#snippet action()}
      <Button.Root
        variant="destructive"
        class="min-h-11"
        onclick={() => {
          pendingAction = 'remove';
          onRemoveJob(removeJob);
        }}
        disabled={busy}
      >
        <Button.BusyContent busy={pendingAction === 'remove' && busy} label="Remove from history" busyLabel="Removing from history" />
      </Button.Root>
    {/snippet}
  </WorkspaceConfirmationDialog>
{/if}

<style>
  .confirmation-choice-grid {
    display: grid;
    gap: 0.75rem;
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

  :global(.confirmation-choice strong),
  :global(.confirmation-choice span) {
    display: block;
  }

  :global(.confirmation-choice span),
  .confirmation-topline span {
    color: var(--muted-foreground);
    font-size: 0.85rem;
    font-weight: 400;
    line-height: 1.35;
  }

  .confirmation-topline strong,
  .confirmation-topline span {
    display: block;
  }

  @media (max-width: 40rem) {
    .confirmation-choice-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
