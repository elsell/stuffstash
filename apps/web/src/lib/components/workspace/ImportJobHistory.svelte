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
    attentionSummary,
    canRequestCancellation,
    canRemoveJobFromHistory,
    historyCountSummary,
    importIssueTone,
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

  type HistoryFilter = 'all' | 'current' | 'drafts' | 'attention' | 'warnings' | 'completed';

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

  let historyFilter = $state<HistoryFilter>('all');
  let warningJobs = $derived(terminalJobs.filter(jobHasReviewWarnings));
  let sortedTerminalJobs = $derived([...terminalJobs].sort(compareTerminalJobs));
  let filteredTerminalJobs = $derived(
    historyFilter === 'attention'
      ? sortedTerminalJobs.filter(jobRequiresAction)
      : historyFilter === 'warnings'
        ? sortedTerminalJobs.filter(jobHasReviewWarnings)
      : historyFilter === 'completed'
        ? sortedTerminalJobs.filter((job) => job.status === 'succeeded' && !jobRequiresAction(job))
        : historyFilter === 'current' || historyFilter === 'drafts'
          ? []
          : sortedTerminalJobs
  );

  function compareTerminalJobs(left: ImportJob, right: ImportJob): number {
    const severityDelta = severityRank(right) - severityRank(left);
    if (severityDelta !== 0) return severityDelta;
    return jobSortTime(right) - jobSortTime(left);
  }

  function severityRank(job: ImportJob): number {
    if (jobRequiresAction(job)) return 2;
    if (jobHasReviewWarnings(job)) return 1;
    return 0;
  }

  function jobSortTime(job: ImportJob): number {
    const value = job.completedAt ?? job.startedAt ?? job.createdAt;
    return value ? Date.parse(value) || 0 : 0;
  }

  function jobActionLabel(action: string, job: ImportJob): string {
    const status = statusLabel(job);
    const time = job.completedAt ?? job.startedAt ?? job.createdAt;
    const timeLabel = time ? `, ${new Date(time).toLocaleString()}` : '';
    return `${action} ${job.source.name} import, ${status}${timeLabel}`;
  }

  function ledgerDescription(): string {
    if (historyFilter === 'attention') return 'Imports that need action.';
    if (historyFilter === 'warnings') return 'Completed imports with warnings.';
    if (historyFilter === 'completed') return 'Completed imports.';
    return 'Import runs, newest and most important first.';
  }

  function jobRequiresAction(job: ImportJob): boolean {
    return importIssueTone(job) === 'action';
  }

  function jobHasReviewWarnings(job: ImportJob): boolean {
    return importIssueTone(job) === 'warning';
  }
</script>

<div class="history-header">
  <p>{summaryDescription}</p>
  {#if jobs.length > 0}
    <Button.Root onclick={onBeginImport} disabled={!canCreateImports} variant={currentWorkJobs.length > 0 ? 'outline' : 'default'}>
      <Plus size={16} aria-hidden="true" />
      New import
    </Button.Root>
  {/if}
</div>

{#if jobs.length > 0}
  <div class="history-status-strip" aria-label="Import history filters">
    <Button.Root
      class={historyFilter === 'current' ? 'status-chip active selected' : activeJobs.length > 0 ? 'status-chip active' : 'status-chip'}
      variant="ghost"
      onclick={() => (historyFilter = historyFilter === 'current' ? 'all' : 'current')}
      disabled={activeJobs.length === 0}
      aria-pressed={historyFilter === 'current'}
    >
      <LoaderCircle size={14} aria-hidden="true" />
      <span>Running</span>
      <strong>{activeJobs.length}</strong>
    </Button.Root>
    <Button.Root
      class={historyFilter === 'drafts' ? 'status-chip active selected' : draftJobs.length > 0 ? 'status-chip active' : 'status-chip'}
      variant="ghost"
      onclick={() => (historyFilter = historyFilter === 'drafts' ? 'all' : 'drafts')}
      disabled={draftJobs.length === 0}
      aria-pressed={historyFilter === 'drafts'}
    >
      <Clock3 size={14} aria-hidden="true" />
      <span>Ready to review</span>
      <strong>{draftJobs.length}</strong>
    </Button.Root>
    <Button.Root
      class={historyFilter === 'completed' ? 'status-chip active selected' : completedJobs.length > 0 ? 'status-chip active' : 'status-chip'}
      variant="ghost"
      onclick={() => (historyFilter = historyFilter === 'completed' ? 'all' : 'completed')}
      disabled={completedJobs.length === 0}
      aria-pressed={historyFilter === 'completed'}
    >
      <CheckCircle2 size={14} aria-hidden="true" />
      <span>Completed</span>
      <strong>{completedJobs.length}</strong>
    </Button.Root>
    <Button.Root
      class={historyFilter === 'warnings' ? 'status-chip warning selected' : warningJobs.length > 0 ? 'status-chip warning' : 'status-chip'}
      variant="ghost"
      onclick={() => (historyFilter = historyFilter === 'warnings' ? 'all' : 'warnings')}
      disabled={warningJobs.length === 0}
      aria-pressed={historyFilter === 'warnings'}
    >
      <AlertTriangle size={14} aria-hidden="true" />
      <span>Warnings</span>
      <strong>{warningJobs.length}</strong>
    </Button.Root>
    <Button.Root
      class={historyFilter === 'attention' ? 'status-chip danger selected' : attentionJobs.length > 0 ? 'status-chip danger' : 'status-chip'}
      variant="ghost"
      onclick={() => (historyFilter = historyFilter === 'attention' ? 'all' : 'attention')}
      disabled={attentionJobs.length === 0}
      aria-pressed={historyFilter === 'attention'}
    >
      <XCircle size={14} aria-hidden="true" />
      <span>Action required</span>
      <strong>{attentionJobs.length}</strong>
    </Button.Root>
  </div>
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
  {#if currentWorkJobs.length > 0 && (historyFilter === 'all' || historyFilter === 'current' || historyFilter === 'drafts')}
    <div class="job-section current-work-section">
      <div class="section-heading">
        <div>
          <h3>Current work</h3>
          <p>Resume drafts, watch progress, or cancel running imports.</p>
        </div>
      </div>
      {#if activeJobs.length > 0 && historyFilter !== 'drafts'}
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
      {/if}
      {#if draftJobs.length > 0 && historyFilter !== 'current'}
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
      {/if}
    </div>
  {/if}
  {#if attentionJobs.length > 0 && historyFilter !== 'attention'}
    <div class="attention-alert" role="status">
      <div class="attention-marker" aria-hidden="true">
        <AlertTriangle size={18} />
      </div>
      <div>
        <strong>{attentionJobs.length === 1 ? '1 import requires action' : `${attentionJobs.length} imports require action`}</strong>
        <span>{attentionJobs.length === 1 ? 'A blocking issue or cleanup failure is flagged below.' : 'Blocking issues or cleanup failures are flagged below.'}</span>
      </div>
      <Button.Root variant="outline" size="sm" onclick={() => (historyFilter = 'attention')}>Show only those</Button.Root>
    </div>
  {/if}
  {#if terminalJobs.length > 0 && historyFilter !== 'current' && historyFilter !== 'drafts'}
    <div class="job-section">
      <div class="ledger-heading">
        <div>
          <h3>Runs</h3>
          <p>{ledgerDescription()}</p>
        </div>
        {#if historyFilter !== 'all'}
          <Button.Root variant="ghost" size="sm" onclick={() => (historyFilter = 'all')}>Show all</Button.Root>
        {/if}
      </div>
      <div class="history-ledger" role="table" aria-label="Import history">
        <div class="history-ledger-head" role="row">
          <span role="columnheader">Status</span>
          <span role="columnheader">Source</span>
          <span role="columnheader">Result</span>
          <span role="columnheader">Time</span>
          <span role="columnheader">Actions</span>
        </div>
        {#each filteredTerminalJobs as job}
          <div class={jobRequiresAction(job) ? 'history-row attention-row' : jobHasReviewWarnings(job) ? 'history-row warning-row' : 'history-row'} role="row">
            <div class="status-cell" role="cell" data-cell-label="Status">
              <span class="status-icon">
                {#if jobRequiresAction(job)}
                  <AlertTriangle size={18} aria-hidden="true" />
                {:else if jobHasReviewWarnings(job)}
                  <AlertTriangle size={18} aria-hidden="true" />
                {:else if isTerminal(job) && job.status !== 'succeeded'}
                  <XCircle size={18} aria-hidden="true" />
                {:else}
                  <CheckCircle2 size={18} aria-hidden="true" />
                {/if}
              </span>
              <div class="history-title">
                <strong>{job.source.name}</strong>
                <Badge variant={statusVariant(job)}>{statusLabel(job)}</Badge>
                {#if jobRequiresAction(job)}
                  <Badge variant="destructive">Action required</Badge>
                {:else if jobHasReviewWarnings(job)}
                  <Badge variant="secondary">Warnings</Badge>
                {/if}
              </div>
            </div>
            <div class="source-cell" role="cell" data-cell-label="Source">
              <span>{sourceDescription(job)}</span>
              {#if actorSummary(job)}<span>{actorSummary(job)}</span>{/if}
            </div>
            <div class="result-cell" role="cell" data-cell-label="Result">
              <span>{statusSentence(job)}</span>
              <span>
                {historyCountSummary(job)}
                {#if jobHasReviewWarnings(job)} · {attentionSummary(job)}{/if}
                {#if job.cancellationMode === 'keep_partial_progress'} · Partial progress kept{/if}
                {#if job.cancellationMode === 'discard_partial_progress'} · Partial progress discarded{/if}
              </span>
            </div>
            <div class="time-cell" role="cell" data-cell-label="Time">
              {#if job.startedAt}<span>{jobTimeLabel('Started', job.startedAt)}</span>{/if}
              {#if job.completedAt}<span>{jobTimeLabel('Completed', job.completedAt)}</span>{/if}
            </div>
            <div class="row-actions" role="cell" data-cell-label="Actions">
              <Button.Root
                variant={jobRequiresAction(job) ? 'outline' : 'ghost'}
                size="sm"
                onclick={() => onOpenJob(job)}
                aria-label={jobActionLabel(jobRequiresAction(job) || jobHasReviewWarnings(job) ? 'Review Details for' : 'View details for', job)}
              >
                <Eye size={16} aria-hidden="true" />
                {jobRequiresAction(job) || jobHasReviewWarnings(job) ? 'Review Details' : 'Details'}
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
      {#if filteredTerminalJobs.length === 0}
        <div class="quiet-row">
          <CheckCircle2 size={16} aria-hidden="true" />
          {historyFilter === 'all' && attentionJobs.length > 0 ? 'No other import runs to show.' : 'No imports match this filter.'}
        </div>
      {/if}
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

  .history-header > p {
    max-width: 44rem;
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
    color: var(--muted-foreground);
    margin: 0.25rem 0 0;
  }

  .job-section {
    display: grid;
    gap: 0.65rem;
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
    border: 1px solid color-mix(in oklab, var(--primary) 18%, transparent);
    border-radius: 8px;
    padding: 0.75rem;
  }

  .attention-alert {
    align-items: center;
    background: color-mix(in oklab, var(--destructive) 3%, transparent);
    border: 1px solid color-mix(in oklab, var(--destructive) 24%, transparent);
    border-radius: 8px;
    display: grid;
    gap: 0.75rem;
    grid-template-columns: auto minmax(0, 1fr) auto;
    padding: 0.65rem 0.75rem;
  }

  .attention-marker {
    color: var(--destructive);
    display: grid;
    place-items: center;
  }

  .attention-alert > div:nth-child(2) {
    display: grid;
    gap: 0.1rem;
    min-width: 0;
  }

  .attention-alert strong {
    color: var(--foreground);
    font-size: 0.9rem;
  }

  .attention-alert span {
    color: var(--muted-foreground);
    font-size: 0.82rem;
    overflow-wrap: anywhere;
  }

  .history-status-strip {
    align-items: center;
    display: grid;
    gap: 0.45rem;
    grid-template-columns: repeat(5, minmax(0, 1fr));
    margin: 0;
  }

  :global(.status-chip) {
    align-items: center;
    border: 1px solid var(--border);
    border-radius: 8px;
    color: var(--muted-foreground);
    display: grid;
    font-size: 0.82rem;
    gap: 0.12rem 0.4rem;
    grid-template-columns: auto 1fr auto;
    min-width: 0;
    min-height: 2.65rem;
    padding: 0.42rem 0.55rem;
    text-align: left;
  }

  :global(.status-chip:not(:disabled)) {
    cursor: pointer;
  }

  :global(.status-chip.active) {
    background: color-mix(in oklab, var(--primary) 6%, transparent);
    border-color: color-mix(in oklab, var(--primary) 28%, transparent);
    color: var(--foreground);
  }

  :global(.status-chip.warning) {
    background: color-mix(in oklab, var(--color-warning, #a15c00) 7%, transparent);
    border-color: color-mix(in oklab, var(--color-warning, #a15c00) 30%, transparent);
    color: var(--color-warning-foreground, #6b3a00);
  }

  :global(.status-chip.danger) {
    background: color-mix(in oklab, var(--destructive) 6%, transparent);
    border-color: color-mix(in oklab, var(--destructive) 30%, transparent);
    color: var(--destructive);
  }

  :global(.status-chip.selected) {
    box-shadow: 0 0 0 2px color-mix(in oklab, var(--ring) 18%, transparent);
  }

  :global(.status-chip span) {
    font-weight: 500;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  :global(.status-chip strong) {
    color: var(--foreground);
    font-weight: 700;
    font-size: 1rem;
    line-height: 1;
  }

  .job-main span {
    color: var(--muted-foreground);
    display: block;
    font-size: 0.82rem;
  }

  .active-status-icon {
    color: var(--primary);
    flex: 0 0 auto;
  }

  .active-job-body {
    min-width: 0;
    width: min(32rem, 100%);
  }

  .progress-track {
    background: var(--muted);
    border-radius: 999px;
    height: 0.45rem;
    margin-top: 0.45rem;
    overflow: hidden;
    width: min(22rem, 100%);
  }

  .progress-track span {
    background: var(--primary);
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
    color: var(--muted-foreground);
    font-size: 0.76rem;
    line-height: 1.2;
  }

  .progress-header strong {
    color: var(--foreground);
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
    color: var(--muted-foreground);
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
    color: var(--border);
    content: "·";
    margin-left: 0.65rem;
  }

  .history-row {
    align-items: center;
    border: 1px solid var(--border);
    border-radius: 8px;
    display: grid;
    gap: 0.55rem;
    grid-template-columns: auto minmax(0, 1fr) auto;
    padding: 0.58rem 0.7rem;
  }

  .history-ledger {
    display: grid;
    gap: 0.35rem;
  }

  .ledger-heading {
    align-items: end;
    display: flex;
    gap: 0.75rem;
    justify-content: space-between;
  }

  .ledger-heading p {
    font-size: 0.86rem;
  }

  .history-ledger-head,
  .history-ledger .history-row {
    display: grid;
    gap: 0.65rem;
    grid-template-columns: minmax(10rem, 1fr) minmax(11rem, 1fr) minmax(12rem, 1.25fr) minmax(8rem, 0.72fr) auto;
  }

  .history-ledger-head {
    color: var(--muted-foreground);
    font-size: 0.75rem;
    font-weight: 700;
    padding: 0 0.8rem;
    text-transform: uppercase;
  }

  .history-row > div {
    min-width: 0;
  }

  .status-cell {
    align-items: flex-start;
    display: flex;
    gap: 0.6rem;
    min-width: 0;
  }

  .result-cell,
  .source-cell,
  .time-cell {
    color: var(--muted-foreground);
    display: grid;
    font-size: 0.82rem;
    gap: 0.18rem;
    min-width: 0;
  }

  .result-cell span:first-child {
    color: var(--foreground);
  }

  .result-cell span,
  .source-cell span,
  .time-cell span {
    overflow-wrap: anywhere;
  }

  .history-row:hover {
    background: color-mix(in oklab, var(--muted) 25%, transparent);
  }

  .history-row.attention-row {
    background: color-mix(in oklab, var(--destructive) 2.6%, transparent);
    border-color: color-mix(in oklab, var(--destructive) 30%, transparent);
  }

  .attention-row .status-icon {
    color: var(--destructive);
  }

  .history-row.warning-row {
    border-color: color-mix(in oklab, var(--color-warning, #a15c00) 26%, var(--border));
  }

  .warning-row .status-icon {
    color: var(--color-warning-foreground, #6b3a00);
  }

  .row-actions {
    align-items: center;
    display: flex;
    gap: 0.35rem;
    justify-content: flex-end;
  }

  .status-icon {
    color: var(--muted-foreground);
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
    color: var(--muted-foreground);
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
    :global(.import-history-empty-state),
    .attention-alert {
      grid-template-columns: 1fr;
    }

    .history-ledger-head {
      display: none;
    }

    .history-ledger .history-row {
      grid-template-columns: 1fr;
    }

    .history-ledger .history-row > [role="cell"]::before {
      color: var(--muted-foreground);
      content: attr(data-cell-label);
      display: block;
      font-size: 0.68rem;
      font-weight: 700;
      margin-bottom: 0.22rem;
      text-transform: uppercase;
    }

    .status-cell {
      display: block;
    }

    .history-status-strip {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    :global(.status-chip) {
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

    .status-icon,
    .attention-marker {
      display: none;
    }

    .row-actions {
      justify-content: flex-start;
    }

    .ledger-heading {
      align-items: flex-start;
      flex-direction: column;
    }

    .current-work-section .section-heading p {
      display: none;
    }
  }
</style>
