<script lang="ts">
  import AlertTriangle from '@lucide/svelte/icons/alert-triangle';
  import CheckCircle2 from '@lucide/svelte/icons/check-circle-2';
  import Clock3 from '@lucide/svelte/icons/clock-3';
  import Database from '@lucide/svelte/icons/database';
  import Eye from '@lucide/svelte/icons/eye';
  import LoaderCircle from '@lucide/svelte/icons/loader-circle';
  import Plus from '@lucide/svelte/icons/plus';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import XCircle from '@lucide/svelte/icons/x-circle';
  import type { ImportJob } from '$lib/domain/inventory';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import {
    actorSummary,
    canRequestCancellation,
    canRemoveJobFromHistory,
    historyCountSummary,
    isTerminal,
    jobTimeLabel,
    phaseLabel,
    progressBarLabel,
    progressBarStyle,
    progressKnown,
    progressPercent,
    progressSummary,
    sourceDescription,
    statusLabel,
    statusSentence,
    statusVariant
  } from './importWorkspacePresentation';

  type Props = {
    jobs: ImportJob[];
    activeJobs: ImportJob[];
    draftJobs: ImportJob[];
    terminalJobs: ImportJob[];
    completedJobs: ImportJob[];
    attentionJobs: ImportJob[];
    currentWorkJobs: ImportJob[];
    summaryDescription: string;
    canCreateImports: boolean;
    busy: boolean;
    onBeginImport: () => void;
    onOpenJob: (job: ImportJob) => void;
    onResumePreviewedJob: (job: ImportJob) => void;
    onRequestCancellation: (job: ImportJob) => void;
    onRequestRemove: (job: ImportJob) => void;
  };

  let {
    jobs,
    activeJobs,
    draftJobs,
    terminalJobs,
    completedJobs,
    attentionJobs,
    currentWorkJobs,
    summaryDescription,
    canCreateImports,
    busy,
    onBeginImport,
    onOpenJob,
    onResumePreviewedJob,
    onRequestCancellation,
    onRequestRemove
  }: Props = $props();

  function jobActionLabel(action: string, job: ImportJob): string {
    const status = statusLabel(job);
    const time = job.completedAt ?? job.startedAt ?? job.createdAt;
    const timeLabel = time ? `, ${new Date(time).toLocaleString()}` : '';
    return `${action} ${job.source.name} import, ${status}${timeLabel}`;
  }
</script>

<div class="history-header">
  <div>
    <h2>Import history</h2>
    <p>{summaryDescription}</p>
  </div>
  {#if jobs.length > 0}
    <Button.Root onclick={onBeginImport} disabled={!canCreateImports} variant={currentWorkJobs.length > 0 ? 'outline' : 'default'}>
      <Plus size={16} aria-hidden="true" />
      New import
    </Button.Root>
  {/if}
</div>

{#if jobs.length > 0}
  <dl class="history-status-strip" aria-label="Import history summary">
    <div class={activeJobs.length > 0 ? 'status-chip active' : 'status-chip'}>
      <LoaderCircle size={14} aria-hidden="true" />
      <dt>Running</dt>
      <dd>{activeJobs.length}</dd>
    </div>
    <div class={draftJobs.length > 0 ? 'status-chip active' : 'status-chip'}>
      <Clock3 size={14} aria-hidden="true" />
      <dt>Ready to review</dt>
      <dd>{draftJobs.length}</dd>
    </div>
    <div class={completedJobs.length > 0 ? 'status-chip active' : 'status-chip'}>
      <CheckCircle2 size={14} aria-hidden="true" />
      <dt>Completed</dt>
      <dd>{completedJobs.length}</dd>
    </div>
    <div class={attentionJobs.length > 0 ? 'status-chip warning' : 'status-chip'}>
      <AlertTriangle size={14} aria-hidden="true" />
      <dt>Needs attention</dt>
      <dd>{attentionJobs.length}</dd>
    </div>
  </dl>
{/if}

{#if jobs.length === 0}
  <Card.Root>
    <Card.Content class="import-history-empty-state">
      <Database size={28} aria-hidden="true" />
      <div>
        <h2>No import runs yet</h2>
        <p>Start with Homebox, preview what Stuff Stash will create, then run it in the background.</p>
      </div>
      {#if canCreateImports}
        <Button.Root onclick={onBeginImport}>
          <Plus size={16} aria-hidden="true" />
          New import
        </Button.Root>
      {:else}
        <p class="empty-state-access-note">Creating imports requires import job create access.</p>
      {/if}
    </Card.Content>
  </Card.Root>
{:else}
  {#if currentWorkJobs.length > 0}
    <div class="job-section current-work-section">
      <div class="section-heading">
        <div>
          <h3>Current work</h3>
          <p>Resume drafts, watch progress, or cancel running imports.</p>
        </div>
      </div>
      {#each activeJobs as job}
        <Card.Root>
          <Card.Content class="import-job-card">
            <div class="job-main">
              <span class="active-status-icon"><LoaderCircle class="import-history-spin" size={18} aria-hidden="true" /></span>
              <div class="active-job-body">
                <div class="history-title">
                  <strong>{job.source.name}</strong>
                  <Badge variant={statusVariant(job)}>{statusLabel(job)}</Badge>
                </div>
                <span>{job.status === 'cancel_requested' ? statusSentence(job) : actorSummary(job) || 'Background job'}</span>
                <div class="progress-header">
                  <span>{phaseLabel(job)}</span>
                  <strong>{progressSummary(job)}</strong>
                </div>
                <div
                  class="progress-track"
                  class:indeterminate={!progressKnown(job) && !isTerminal(job)}
                  role="progressbar"
                  aria-label={progressBarLabel(job)}
                  aria-valuemin={progressKnown(job) ? 0 : undefined}
                  aria-valuemax={progressKnown(job) ? 100 : undefined}
                  aria-valuenow={progressKnown(job) ? progressPercent(job) : undefined}
                >
                  <span style={progressBarStyle(job)}></span>
                </div>
              </div>
            </div>
            <div class="action-row">
              <Button.Root variant="ghost" size="sm" onclick={() => onOpenJob(job)} disabled={busy} aria-label={jobActionLabel('View details for', job)}>
                <Eye size={16} aria-hidden="true" />
                Details
              </Button.Root>
              {#if canRequestCancellation(job)}
                <Button.Root
                  variant="outline"
                  size="sm"
                  onclick={() => onRequestCancellation(job)}
                  disabled={busy || !canCreateImports}
                  aria-label={jobActionLabel('Cancel', job)}
                >
                  Cancel
                </Button.Root>
              {/if}
            </div>
          </Card.Content>
        </Card.Root>
      {/each}
      {#each draftJobs as job}
        <div class="history-row draft-row">
          <span class="status-icon"><Clock3 size={18} aria-hidden="true" /></span>
          <div>
            <div class="history-title">
              <strong>{job.source.name}</strong>
              <Badge variant="secondary">{statusLabel(job)}</Badge>
            </div>
            <div class="history-meta">
              <span>{statusSentence(job)}</span>
              <span>{historyCountSummary(job)}</span>
              {#if actorSummary(job)}<span>{actorSummary(job)}</span>{/if}
              <span>{jobTimeLabel('Previewed', job.createdAt)}</span>
            </div>
          </div>
          <Button.Root variant="outline" size="sm" onclick={() => onResumePreviewedJob(job)} aria-label={jobActionLabel('Continue', job)}>Continue</Button.Root>
          <Button.Root variant="ghost" size="sm" onclick={() => onOpenJob(job)} aria-label={jobActionLabel('View details for', job)}>Details</Button.Root>
        </div>
      {/each}
    </div>
  {/if}
  {#if terminalJobs.length > 0}
    <div class="job-section">
      <div class="section-heading">
        <div>
          <h3>History</h3>
          <p>Completed, failed, and cancelled runs stay here until you remove them.</p>
        </div>
      </div>
      {#each terminalJobs as job}
        <div class="history-row">
          <span class="status-icon">
            {#if isTerminal(job) && job.status !== 'succeeded'}
              <XCircle size={18} aria-hidden="true" />
            {:else}
              <CheckCircle2 size={18} aria-hidden="true" />
            {/if}
          </span>
          <div class="history-row-body">
            <div class="history-title">
              <strong>{job.source.name}</strong>
              <Badge variant={statusVariant(job)}>{statusLabel(job)}</Badge>
            </div>
            <div class="history-meta">
              <span>{statusSentence(job)}</span>
              <span>{historyCountSummary(job)}</span>
              <span>{sourceDescription(job)}</span>
              {#if actorSummary(job)}<span>{actorSummary(job)}</span>{/if}
              {#if job.startedAt}<span>{jobTimeLabel('Started', job.startedAt)}</span>{/if}
              {#if job.completedAt}<span>{jobTimeLabel('Completed', job.completedAt)}</span>{/if}
              {#if job.cancellationMode === 'keep_partial_progress'}<span>Partial progress kept</span>{/if}
              {#if job.cancellationMode === 'discard_partial_progress'}<span>Partial progress discarded</span>{/if}
            </div>
          </div>
          <div class="row-actions">
            <Button.Root variant="ghost" size="sm" onclick={() => onOpenJob(job)} aria-label={jobActionLabel('View details for', job)}>
              <Eye size={16} aria-hidden="true" />
              Details
            </Button.Root>
            {#if canRemoveJobFromHistory(job)}
              <Button.Root variant="ghost" size="icon" onclick={() => onRequestRemove(job)} aria-label={jobActionLabel('Remove from history', job)}>
                <Trash2 size={16} aria-hidden="true" />
              </Button.Root>
            {/if}
          </div>
        </div>
      {/each}
    </div>
  {:else if currentWorkJobs.length > 0}
    <div class="quiet-row">
      <CheckCircle2 size={16} aria-hidden="true" />
      No completed import runs yet.
    </div>
  {/if}
{/if}

<style>
  .history-header,
  :global(.import-job-card),
  .job-main,
  .quiet-row,
  .action-row {
    align-items: center;
    display: flex;
    gap: 0.75rem;
  }

  .history-header,
  :global(.import-job-card) {
    justify-content: space-between;
  }

  h2,
  h3 {
    margin: 0;
  }

  h2 {
    font-size: 1.25rem;
  }

  h3 {
    font-size: 1rem;
  }

  p {
    color: hsl(var(--muted-foreground));
    margin: 0.25rem 0 0;
  }

  .job-section {
    display: grid;
    gap: 1rem;
  }

  .section-heading {
    align-items: flex-end;
    display: flex;
    gap: 0.75rem;
    justify-content: space-between;
  }

  .section-heading p {
    font-size: 0.86rem;
  }

  .current-work-section {
    border: 1px solid hsl(var(--primary) / 0.18);
    border-radius: 8px;
    padding: 0.75rem;
  }

  .history-status-strip {
    align-items: center;
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin: 0;
  }

  .status-chip {
    align-items: baseline;
    border: 1px solid hsl(var(--border));
    border-radius: 999px;
    color: hsl(var(--muted-foreground));
    display: flex;
    font-size: 0.82rem;
    gap: 0.35rem;
    min-width: 0;
    padding: 0.35rem 0.6rem;
  }

  .status-chip.active {
    background: hsl(var(--primary) / 0.06);
    border-color: hsl(var(--primary) / 0.28);
    color: hsl(var(--foreground));
  }

  .status-chip.warning {
    background: hsl(var(--destructive) / 0.06);
    border-color: hsl(var(--destructive) / 0.28);
    color: hsl(var(--destructive));
  }

  .status-chip dt {
    font-weight: 500;
  }

  .status-chip dd {
    color: hsl(var(--foreground));
    font-weight: 700;
    line-height: 1;
    margin: 0;
    order: -1;
  }

  .job-main span {
    color: hsl(var(--muted-foreground));
    display: block;
    font-size: 0.82rem;
  }

  .active-status-icon {
    color: hsl(var(--primary));
    flex: 0 0 auto;
  }

  .active-job-body {
    min-width: 0;
    width: min(32rem, 100%);
  }

  .progress-track {
    background: hsl(var(--muted));
    border-radius: 999px;
    height: 0.45rem;
    margin-top: 0.45rem;
    overflow: hidden;
    width: min(22rem, 100%);
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

  .progress-header {
    align-items: baseline;
    display: flex;
    gap: 0.75rem;
    justify-content: space-between;
    margin-top: 0.5rem;
    max-width: 22rem;
    width: 100%;
  }

  .progress-header span,
  .progress-header strong {
    color: hsl(var(--muted-foreground));
    font-size: 0.76rem;
    line-height: 1.2;
  }

  .progress-header strong {
    color: hsl(var(--foreground));
    font-weight: 650;
    white-space: nowrap;
  }

  .history-title {
    align-items: center;
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .history-meta {
    color: hsl(var(--muted-foreground));
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem 0.65rem;
    margin-top: 0.25rem;
  }

  .history-meta span {
    font-size: 0.82rem;
    min-width: 0;
    overflow-wrap: anywhere;
  }

  .history-meta span:not(:last-child)::after {
    color: hsl(var(--border));
    content: "·";
    margin-left: 0.65rem;
  }

  .history-row {
    align-items: center;
    border: 1px solid hsl(var(--border));
    border-radius: 8px;
    display: grid;
    gap: 0.75rem;
    grid-template-columns: auto minmax(0, 1fr) auto;
    padding: 0.8rem;
  }

  .history-row > div {
    min-width: 0;
  }

  .history-row:hover {
    background: hsl(var(--muted) / 0.25);
  }

  .row-actions {
    align-items: center;
    display: flex;
    gap: 0.35rem;
    justify-content: flex-end;
  }

  .status-icon {
    color: hsl(var(--muted-foreground));
    display: grid;
    place-items: center;
  }

  :global(.import-history-empty-state) {
    align-items: center;
    display: grid;
    gap: 1rem;
    grid-template-columns: auto minmax(0, 1fr) auto;
  }

  .empty-state-access-note {
    color: hsl(var(--muted-foreground));
    font-size: 0.9rem;
    margin: 0;
  }

  :global(.import-history-spin) {
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
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

  @media (prefers-reduced-motion: reduce) {
    .progress-track span,
    .progress-track.indeterminate span,
    :global(.import-history-spin) {
      animation: none;
      transition: none;
    }

    .progress-track.indeterminate span {
      transform: none;
      width: 42%;
    }
  }

  @media (max-width: 860px) {
    .history-header,
    :global(.import-job-card),
    .job-main {
      align-items: flex-start;
      flex-direction: column;
    }

    .history-row,
    :global(.import-history-empty-state) {
      grid-template-columns: 1fr;
    }

    .history-status-strip {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .status-chip {
      border-radius: 8px;
      justify-content: flex-start;
      padding: 0.5rem 0.6rem;
    }

    .history-meta {
      display: grid;
      gap: 0.2rem;
    }

    .history-meta span:not(:last-child)::after {
      content: "";
      margin-left: 0;
    }

    .status-icon {
      display: none;
    }

    .row-actions {
      justify-content: flex-start;
    }
  }
</style>
