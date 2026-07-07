<script lang="ts">
  import AlertCircle from '@lucide/svelte/icons/alert-circle';
  import CheckCircle2 from '@lucide/svelte/icons/check-circle-2';
  import RefreshCw from '@lucide/svelte/icons/refresh-cw';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import type { ImportJob } from '$lib/domain/inventory';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import {
    actorSummary,
    canRemoveJobFromHistory,
    isTerminal,
    phaseLabel,
    progressBarLabel,
    progressBarStyle,
    progressKnown,
    progressPercent,
    progressTimeline,
    resourceDiagnosticLabel,
    resourceLabel,
    resultCountCells,
    sourceDescription,
    sourceOptionsSummary,
    statusLabel,
    statusSentence,
    visibleCountCells,
    visiblePreviewCountCells
  } from './importWorkspacePresentation';
  import ImportCountGrid from './ImportCountGrid.svelte';
  import ImportMessagesList from './ImportMessagesList.svelte';
  import ImportPreviewSamples from './ImportPreviewSamples.svelte';

  type ImportResource = ImportJob['resources'][number];

  type Props = {
    job: ImportJob;
    canRequestCancellation: boolean;
    detailLoading: boolean;
    canCreateImports: boolean;
    busy: boolean;
    auditHistoryHref: string;
    resourceCanOpen: (job: ImportJob, resource: ImportResource) => boolean;
    resourceHref: (resource: ImportResource) => string;
    onCancel: () => void;
    onContinue: () => void;
    onRemove: () => void;
  };

  let {
    job,
    canRequestCancellation,
    detailLoading,
    canCreateImports,
    busy,
    auditHistoryHref,
    resourceCanOpen,
    resourceHref,
    onCancel,
    onContinue,
    onRemove
  }: Props = $props();

  function hasPreviewPlan(job: ImportJob): boolean {
    if (job.status === 'previewed') {
      return true;
    }
    return (
      job.preview.fields.length > 0 ||
      job.preview.locations.length > 0 ||
      job.preview.assets.length > 0 ||
      job.preview.attachments.length > 0
    );
  }

  function detailMessages(job: ImportJob): ImportJob['messages'] {
    return job.messages.length > 0 ? job.messages : job.preview.messages;
  }

  function cancellationSummary(job: ImportJob): string {
    if (job.status === 'cancel_requested' && job.cancellationMode === 'discard_partial_progress') {
      return 'Cancellation requested. Partial progress will be discarded. Audit history remains.';
    }
    if (job.status === 'cancel_requested') {
      return 'Cancellation requested. Partial progress will be kept.';
    }
    if (job.cancellationMode === 'discard_partial_progress') {
      return 'Partial progress discard was requested. Audit history remains.';
    }
    return 'Partial progress was kept.';
  }
</script>

<Card.Root>
  <Card.Header>
    <Card.Title>{job.source.name}</Card.Title>
    <Card.Description>{statusLabel(job)} · {sourceDescription(job)}{actorSummary(job) ? ` · ${actorSummary(job)}` : ''}</Card.Description>
  </Card.Header>
  <Card.Content>
    <div class="import-detail-content">
    <div class="detail-topline">
      <div>
        <strong>{phaseLabel(job)}</strong>
        <span>{statusSentence(job)}</span>
      </div>
      <Badge variant={job.status === 'failed' || job.status === 'discard_failed' ? 'destructive' : 'secondary'}>
        {statusLabel(job)}
      </Badge>
    </div>
    {#if detailLoading}
      <div class="quiet-row">
        <RefreshCw size={16} aria-hidden="true" />
        Refreshing import details.
      </div>
    {/if}
    {#if progressKnown(job) || !isTerminal(job)}
      <div
        class="progress-track large"
        class:indeterminate={!progressKnown(job)}
        role="progressbar"
        aria-label={progressBarLabel(job)}
        aria-valuemin={progressKnown(job) ? 0 : undefined}
        aria-valuemax={progressKnown(job) ? 100 : undefined}
        aria-valuenow={progressKnown(job) ? progressPercent(job) : undefined}
      >
        <span style={progressBarStyle(job)}></span>
      </div>
    {/if}
    <section class="timeline-section">
      <div class="sample-heading">
        <h3>Progress timeline</h3>
        <small>{progressTimeline(job).length} phases</small>
      </div>
      <div class="timeline-list">
        {#each progressTimeline(job) as progress}
          <div class="timeline-row">
            <span>{phaseLabel({ ...job, progress })}</span>
            <small>
              {progress.total > 0 ? `${Math.min(progress.done, progress.total)} / ${progress.total}` : ''}
              {progress.total > 0 && progress.updatedAt ? ' · ' : ''}
              {progress.updatedAt ? new Date(progress.updatedAt).toLocaleString() : ''}
            </small>
          </div>
        {/each}
      </div>
    </section>
    <section class="source-options-section" aria-label="Import source options">
      <div class="sample-heading">
        <h3>Source</h3>
        <small>{job.source.type === 'legacy_homebox_csv' ? 'File snapshot' : 'Live connection'}</small>
      </div>
      <ul class="source-option-list" aria-label="Selected source options">
        {#each sourceOptionsSummary(job) as option}
          <li>{option}</li>
        {/each}
      </ul>
    </section>
    <ImportCountGrid cells={job.status === 'previewed' ? visiblePreviewCountCells(job) : visibleCountCells(resultCountCells(job))} />
    {#if hasPreviewPlan(job)}
      <section class="preview-plan-section" aria-label="Import preview plan">
        <div class="sample-heading">
          <h3>Preview plan</h3>
          <small>{job.status === 'previewed' ? 'Before import' : 'Original plan'}</small>
        </div>
        <ImportPreviewSamples preview={job.preview} />
      </section>
    {/if}
    {#if job.cancellationMode}
      <div class="quiet-row">
        <AlertCircle size={16} aria-hidden="true" />
        {cancellationSummary(job)}
      </div>
    {/if}
    {#if job.status === 'cancelled_discarded' && job.resources.length > 0}
      <div class="quiet-row">
        <CheckCircle2 size={16} aria-hidden="true" />
        Records created by this job were discarded. Audit history remains.
      </div>
    {:else if job.resources.length > 0}
      <section class="resource-section">
        <div class="sample-heading">
          <h3>Imported records</h3>
          {#if job.resources.length >= 50}<small>Showing a sample</small>{/if}
        </div>
        <div class="sample-list">
          {#each job.resources as resource}
            <div class="sample-row resource-row">
              <span>{resourceLabel(resource)}</span>
              <small>{resourceDiagnosticLabel(resource)} · Imported {new Date(resource.createdAt).toLocaleString()}</small>
              {#if resourceCanOpen(job, resource)}
                <a class="resource-link" href={resourceHref(resource)}>Open</a>
              {/if}
            </div>
          {/each}
        </div>
      </section>
    {/if}
    <ImportMessagesList messages={detailMessages(job)} emptyText="No import messages." truncated={job.messages.length === 0 && job.preview.messagesTruncated} />
    <div class="action-row">
      <a class="detail-link" href={auditHistoryHref}>
        View audit history
      </a>
      {#if canCreateImports && canRequestCancellation}
        <Button.Root variant="outline" onclick={onCancel} disabled={busy}>
          Cancel
        </Button.Root>
      {/if}
      {#if canCreateImports && job.status === 'previewed'}
        <Button.Root onclick={onContinue} disabled={busy}>
          Continue import
        </Button.Root>
      {/if}
      {#if canCreateImports && canRemoveJobFromHistory(job)}
        <Button.Root variant="ghost" onclick={onRemove} disabled={busy}>
          <Trash2 size={16} aria-hidden="true" />
        Remove from history
      </Button.Root>
      {/if}
    </div>
    </div>
  </Card.Content>
</Card.Root>

<style>
  .import-detail-content {
    display: grid;
    gap: 1rem;
    min-width: 0;
  }

  .action-row,
  .quiet-row,
  .detail-topline {
    align-items: center;
    display: flex;
    gap: 0.75rem;
  }

  .action-row {
    flex-wrap: wrap;
  }

  .detail-topline {
    justify-content: space-between;
    min-width: 0;
  }

  .detail-topline span {
    color: hsl(var(--muted-foreground));
    display: block;
    font-size: 0.85rem;
    margin-top: 0.15rem;
    overflow-wrap: anywhere;
  }

  .progress-track {
    background: hsl(var(--muted));
    border-radius: 999px;
    height: 0.45rem;
    margin-top: 0.45rem;
    overflow: hidden;
    width: min(22rem, 100%);
  }

  .progress-track.large {
    height: 0.65rem;
    width: 100%;
  }

  .progress-track span {
    background: hsl(var(--primary));
    display: block;
    height: 100%;
    transition: width 180ms ease;
  }

  .progress-track.indeterminate span {
    animation: import-progress-indeterminate 1.4s ease-in-out infinite;
    width: 35%;
  }

  @keyframes import-progress-indeterminate {
    0% {
      transform: translateX(-110%);
    }

    50% {
      transform: translateX(95%);
    }

    100% {
      transform: translateX(310%);
    }
  }

  .sample-list,
  .source-option-list,
  .timeline-list {
    display: grid;
  }

  .sample-list {
    gap: 0.45rem;
  }

  .timeline-list {
    gap: 0.35rem;
  }

  .preview-plan-section,
  .source-options-section {
    border: 1px solid hsl(var(--border));
    border-radius: 8px;
    display: grid;
    gap: 0.65rem;
    padding: 0.75rem;
  }

  .source-option-list {
    gap: 0.45rem;
    grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .source-option-list li {
    border-left: 2px solid hsl(var(--border));
    color: hsl(var(--foreground));
    font-size: 0.82rem;
    min-width: 0;
    overflow-wrap: anywhere;
    padding: 0.2rem 0 0.2rem 0.55rem;
  }

  .timeline-section {
    border: 1px solid hsl(var(--border));
    border-radius: 8px;
    display: grid;
    gap: 0.6rem;
    min-width: 0;
    padding: 0.75rem;
  }

  .timeline-section {
    gap: 0.65rem;
  }

  .sample-heading {
    align-items: baseline;
    display: flex;
    gap: 0.5rem;
    justify-content: space-between;
    min-width: 0;
  }

  .sample-heading h3,
  .detail-topline strong {
    min-width: 0;
    overflow-wrap: anywhere;
  }

  .sample-heading small,
  .sample-row small {
    color: hsl(var(--muted-foreground));
    font-size: 0.78rem;
  }

  .timeline-row {
    border-left: 2px solid hsl(var(--border));
    display: grid;
    gap: 0.15rem;
    min-width: 0;
    padding-left: 0.6rem;
  }

  .timeline-row span {
    font-weight: 600;
  }

  .timeline-row small {
    color: hsl(var(--muted-foreground));
    overflow-wrap: anywhere;
  }

  .sample-row {
    min-width: 0;
  }

  .resource-row {
    align-items: center;
    display: grid;
    gap: 0.5rem;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) auto;
  }

  .sample-row span,
  .sample-row small {
    display: block;
    overflow-wrap: anywhere;
  }

  .resource-link,
  .detail-link {
    color: hsl(var(--primary));
    font-size: 0.85rem;
    font-weight: 600;
    text-decoration: none;
  }

  .resource-link:hover,
  .resource-link:focus-visible,
  .detail-link:hover,
  .detail-link:focus-visible {
    text-decoration: underline;
  }

  h3 {
    font-size: 1rem;
    margin: 0;
  }

  @media (prefers-reduced-motion: reduce) {
    .progress-track span,
    .progress-track.indeterminate span {
      animation: none;
      transition: none;
    }

    .progress-track.indeterminate span {
      transform: none;
      width: 42%;
    }
  }

  @media (max-width: 860px) {
    .action-row,
    .detail-topline {
      align-items: flex-start;
      flex-direction: column;
    }

    .resource-row {
      grid-template-columns: 1fr;
    }
  }
</style>
