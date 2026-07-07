<script lang="ts">
  import AlertTriangle from '@lucide/svelte/icons/alert-triangle';
  import CheckCircle2 from '@lucide/svelte/icons/check-circle-2';
  import Clock3 from '@lucide/svelte/icons/clock-3';
  import Database from '@lucide/svelte/icons/database';
  import Eye from '@lucide/svelte/icons/eye';
  import LoaderCircle from '@lucide/svelte/icons/loader-circle';
  import Plus from '@lucide/svelte/icons/plus';
  import ArrowUpDown from '@lucide/svelte/icons/arrow-up-down';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import XCircle from '@lucide/svelte/icons/x-circle';
  import type { ImportJob, Principal } from '$lib/domain/inventory';
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
    issueCountSummary,
    isTerminal,
    jobTimeLabel,
    phaseLabel,
    progressBarLabel,
    progressBarStyle,
    progressKnown,
    progressPercent,
    progressSummary,
    changedRecordSummary,
    sourceDescription,
    statusLabel,
    statusSentence,
    statusVariant
  } from './importWorkspacePresentation';

  type HistoryFilter = 'all' | 'current' | 'drafts' | 'attention' | 'warnings' | 'completed';
  type HistorySortKey = 'priority' | 'source' | 'status' | 'changed' | 'finished';
  type HistorySortDirection = 'asc' | 'desc';

  type Props = {
    jobs: ImportJob[];
    activeJobs: ImportJob[];
    draftJobs: ImportJob[];
    terminalJobs: ImportJob[];
    completedJobs: ImportJob[];
    attentionJobs: ImportJob[];
    currentWorkJobs: ImportJob[];
    summaryDescription: string;
    currentPrincipal?: Principal;
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
    currentPrincipal,
    canCreateImports,
    busy,
    onBeginImport,
    onOpenJob,
    onResumePreviewedJob,
    onRequestCancellation,
    onRequestRemove
  }: Props = $props();

  let historyFilter = $state<HistoryFilter>('all');
  let historySortKey = $state<HistorySortKey>('priority');
  let historySortDirection = $state<HistorySortDirection>('desc');
  let warningJobs = $derived(terminalJobs.filter(jobHasReviewWarnings));
  let sortedTerminalJobs = $derived(sortTerminalJobs(terminalJobs));
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

  function sortTerminalJobs(jobs: ImportJob[]): ImportJob[] {
    return [...jobs].sort(compareTerminalJobs);
  }

  function compareTerminalJobs(left: ImportJob, right: ImportJob): number {
    if (historySortKey !== 'priority') {
      const sorted = compareBySelectedSort(left, right);
      if (sorted !== 0) return historySortDirection === 'asc' ? sorted : -sorted;
    }
    const severityDelta = severityRank(right) - severityRank(left);
    if (severityDelta !== 0) return severityDelta;
    return jobSortTime(right) - jobSortTime(left);
  }

  function compareBySelectedSort(left: ImportJob, right: ImportJob): number {
    if (historySortKey === 'source') return sourceSortValue(left).localeCompare(sourceSortValue(right));
    if (historySortKey === 'status') return statusSortValue(left).localeCompare(statusSortValue(right));
    if (historySortKey === 'changed') return changedRecordCount(left) - changedRecordCount(right);
    if (historySortKey === 'finished') return jobSortTime(left) - jobSortTime(right);
    return 0;
  }

  function sourceSortValue(job: ImportJob): string {
    return `${job.source.name} ${sourceDescription(job)}`.toLocaleLowerCase();
  }

  function statusSortValue(job: ImportJob): string {
    return `${severityRank(job)} ${statusLabel(job)} ${statusDetail(job)}`.toLocaleLowerCase();
  }

  function changedRecordCount(job: ImportJob): number {
    if (job.status === 'cancelled_discarded') {
      return job.counts.recordsDiscarded + job.counts.sourceLinksDiscarded;
    }
    return (
      job.counts.fieldsCreated +
      job.counts.locationsCreated +
      job.counts.assetsCreated +
      job.counts.attachmentsCreated +
      job.counts.assetsSkipped +
      job.counts.attachmentsSkipped
    );
  }

  function setHistorySort(key: Exclude<HistorySortKey, 'priority'>): void {
    if (historySortKey === key) {
      historySortDirection = historySortDirection === 'asc' ? 'desc' : 'asc';
      return;
    }
    historySortKey = key;
    historySortDirection = key === 'finished' || key === 'changed' ? 'desc' : 'asc';
  }

  function sortButtonLabel(key: Exclude<HistorySortKey, 'priority'>, label: string): string {
    if (historySortKey !== key) return `Sort by ${label}`;
    return `Sort by ${label}, ${historySortDirection === 'asc' ? 'ascending' : 'descending'}`;
  }

  function sortIndicator(key: Exclude<HistorySortKey, 'priority'>): string {
    if (historySortKey !== key) return '';
    return historySortDirection === 'asc' ? 'Ascending' : 'Descending';
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
    if (historyFilter === 'warnings') return 'Warning-only imports.';
    if (historyFilter === 'completed') return 'Completed imports.';
    return 'Needs-action and recent runs first.';
  }

  function allRunCount(): number {
    return jobs.length;
  }

  function jobRequiresAction(job: ImportJob): boolean {
    return importIssueTone(job) === 'action';
  }

  function jobHasReviewWarnings(job: ImportJob): boolean {
    return importIssueTone(job) === 'warning';
  }

  function statusDetail(job: ImportJob): string {
    if (jobRequiresAction(job)) return `${statusLabel(job)} · ${attentionSummary(job)}`;
    if (job.status === 'cancelled_kept' || job.status === 'cancelled_discarded') return statusSentence(job);
    if (jobHasReviewWarnings(job)) return statusLabel(job);
    if (job.status === 'succeeded') return 'No action needed';
    return statusSentence(job);
  }

  function ledgerChangeSummary(job: ImportJob): string {
    if (!isTerminal(job) || job.status === 'cancelled_discarded') return historyCountSummary(job);
    const skipped = job.counts.assetsSkipped + job.counts.attachmentsSkipped;
    const saved = changedRecordSummary(job);
    if (skipped === 0) return saved;
    if (saved === 'No records changed') return `${skipped} skipped`;
    return `${saved} · ${skipped} skipped`;
  }

  function openHistoryRow(event: MouseEvent, job: ImportJob): void {
    const target = event.target instanceof HTMLElement ? event.target : null;
    if (target?.closest('button, a')) return;
    onOpenJob(job);
  }

  function openHistoryRowFromKeyboard(event: KeyboardEvent, job: ImportJob): void {
    if (event.key !== 'Enter' && event.key !== ' ') return;
    event.preventDefault();
    onOpenJob(job);
  }
</script>

<div class={jobs.length > 0 && currentWorkJobs.length === 0 ? 'history-header compact' : 'history-header'}>
  {#if jobs.length === 0 || currentWorkJobs.length > 0}
    <p>{summaryDescription}</p>
  {/if}
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
      class={historyFilter === 'all' ? 'status-chip selected' : 'status-chip'}
      variant="ghost"
      onclick={() => (historyFilter = 'all')}
      aria-pressed={historyFilter === 'all'}
    >
      <Database size={14} aria-hidden="true" />
      <span>All runs</span>
      <strong>{allRunCount()}</strong>
    </Button.Root>
    {#if activeJobs.length > 0 || historyFilter === 'current'}
      <Button.Root
        class={historyFilter === 'current' ? 'status-chip active selected' : 'status-chip active'}
        variant="ghost"
        onclick={() => (historyFilter = historyFilter === 'current' ? 'all' : 'current')}
        aria-pressed={historyFilter === 'current'}
      >
        <LoaderCircle size={14} aria-hidden="true" />
        <span>Running</span>
        <strong>{activeJobs.length}</strong>
      </Button.Root>
    {/if}
    {#if draftJobs.length > 0 || historyFilter === 'drafts'}
      <Button.Root
        class={historyFilter === 'drafts' ? 'status-chip active selected' : 'status-chip active'}
        variant="ghost"
        onclick={() => (historyFilter = historyFilter === 'drafts' ? 'all' : 'drafts')}
        aria-pressed={historyFilter === 'drafts'}
      >
        <Clock3 size={14} aria-hidden="true" />
        <span>Ready to review</span>
        <strong>{draftJobs.length}</strong>
      </Button.Root>
    {/if}
    {#if attentionJobs.length > 0 || historyFilter === 'attention'}
      <Button.Root
        class={historyFilter === 'attention' ? 'status-chip danger selected' : 'status-chip danger'}
        variant="ghost"
        onclick={() => (historyFilter = historyFilter === 'attention' ? 'all' : 'attention')}
        aria-pressed={historyFilter === 'attention'}
      >
        <XCircle size={14} aria-hidden="true" />
        <span>Action required</span>
        <strong>{attentionJobs.length}</strong>
      </Button.Root>
    {/if}
    {#if warningJobs.length > 0 || historyFilter === 'warnings'}
      <Button.Root
        class={historyFilter === 'warnings' ? 'status-chip warning selected' : 'status-chip warning'}
        variant="ghost"
        onclick={() => (historyFilter = historyFilter === 'warnings' ? 'all' : 'warnings')}
        aria-pressed={historyFilter === 'warnings'}
      >
        <AlertTriangle size={14} aria-hidden="true" />
        <span>Warnings</span>
        <strong>{warningJobs.length}</strong>
      </Button.Root>
    {/if}
    {#if completedJobs.length > 0 || historyFilter === 'completed'}
      <Button.Root
        class={historyFilter === 'completed' ? 'status-chip active selected' : 'status-chip active'}
        variant="ghost"
        onclick={() => (historyFilter = historyFilter === 'completed' ? 'all' : 'completed')}
        aria-pressed={historyFilter === 'completed'}
      >
        <CheckCircle2 size={14} aria-hidden="true" />
        <span>Completed</span>
        <strong>{completedJobs.length}</strong>
      </Button.Root>
    {/if}
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
          <Card.Content
            class="import-job-card current-work-row clickable-row"
            role="button"
            tabindex={0}
            aria-label={jobActionLabel('View details for', job)}
            onclick={(event) => openHistoryRow(event, job)}
            onkeydown={(event) => openHistoryRowFromKeyboard(event, job)}
          >
            <div class="job-main">
              <span class="active-status-icon"><LoaderCircle class="import-history-spin" size={18} aria-hidden="true" /></span>
              <div class="active-job-body">
                <div class="history-title">
                  <strong>{job.source.name}</strong>
                  <Badge variant={statusVariant(job)}>{statusLabel(job)}</Badge>
                </div>
                <span>{job.status === 'cancel_requested' ? statusSentence(job) : actorSummary(job, currentPrincipal) || 'Background job'}</span>
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
        <div
          class="history-row draft-row clickable-row"
          role="button"
          tabindex={0}
          aria-label={jobActionLabel('View details for', job)}
          onclick={(event) => openHistoryRow(event, job)}
          onkeydown={(event) => openHistoryRowFromKeyboard(event, job)}
        >
          <span class="status-icon"><Clock3 size={18} aria-hidden="true" /></span>
          <div>
            <div class="history-title">
              <strong>{job.source.name}</strong>
              <Badge variant="secondary">{statusLabel(job)}</Badge>
            </div>
            <div class="history-meta">
              <span>{statusSentence(job)}</span>
              <span>{historyCountSummary(job)}</span>
              {#if actorSummary(job, currentPrincipal)}<span>{actorSummary(job, currentPrincipal)}</span>{/if}
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
          <span role="columnheader" aria-sort={historySortKey === 'source' ? (historySortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>
            <Button.Root variant="ghost" size="xs" class="ledger-sort-button" onclick={() => setHistorySort('source')} aria-label={sortButtonLabel('source', 'source')}>
              Source
              <ArrowUpDown size={12} aria-hidden="true" />
              {#if sortIndicator('source')}<small>{sortIndicator('source')}</small>{/if}
            </Button.Root>
          </span>
          <span role="columnheader" aria-sort={historySortKey === 'status' ? (historySortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>
            <Button.Root variant="ghost" size="xs" class="ledger-sort-button" onclick={() => setHistorySort('status')} aria-label={sortButtonLabel('status', 'status')}>
              Status
              <ArrowUpDown size={12} aria-hidden="true" />
              {#if sortIndicator('status')}<small>{sortIndicator('status')}</small>{/if}
            </Button.Root>
          </span>
          <span role="columnheader" aria-sort={historySortKey === 'changed' ? (historySortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>
            <Button.Root variant="ghost" size="xs" class="ledger-sort-button" onclick={() => setHistorySort('changed')} aria-label={sortButtonLabel('changed', 'changed records')}>
              Changed
              <ArrowUpDown size={12} aria-hidden="true" />
              {#if sortIndicator('changed')}<small>{sortIndicator('changed')}</small>{/if}
            </Button.Root>
          </span>
          <span role="columnheader" aria-sort={historySortKey === 'finished' ? (historySortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>
            <Button.Root variant="ghost" size="xs" class="ledger-sort-button" onclick={() => setHistorySort('finished')} aria-label={sortButtonLabel('finished', 'finished time')}>
              Finished
              <ArrowUpDown size={12} aria-hidden="true" />
              {#if sortIndicator('finished')}<small>{sortIndicator('finished')}</small>{/if}
            </Button.Root>
          </span>
          <span role="columnheader">Actions</span>
        </div>
        {#each filteredTerminalJobs as job}
          <div
            class={jobRequiresAction(job) ? 'history-row attention-row clickable-row' : jobHasReviewWarnings(job) ? 'history-row warning-row clickable-row' : 'history-row clickable-row'}
            role="row"
            tabindex="0"
            aria-label={jobActionLabel(jobRequiresAction(job) || jobHasReviewWarnings(job) ? 'Review Details for' : 'View details for', job)}
            onclick={(event) => openHistoryRow(event, job)}
            onkeydown={(event) => openHistoryRowFromKeyboard(event, job)}
          >
            <div class="status-cell" role="cell" data-cell-label="Source">
              <span class="status-icon">
                {#if jobRequiresAction(job)}
                  <XCircle size={18} aria-hidden="true" />
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
                <span>{sourceDescription(job)}</span>
              </div>
            </div>
            <div class={jobRequiresAction(job) ? 'issue-cell action' : jobHasReviewWarnings(job) ? 'issue-cell warning' : 'issue-cell'} role="cell" data-cell-label="Status">
              {#if jobRequiresAction(job)}
                <Badge variant="destructive">Action required</Badge>
              {:else if jobHasReviewWarnings(job)}
                <Badge variant="secondary" class="warning-badge">Warnings</Badge>
              {:else}
                <Badge variant={statusVariant(job)}>{statusLabel(job)}</Badge>
              {/if}
              <span>{statusDetail(job)}</span>
            </div>
            <div class="result-cell" role="cell" data-cell-label="Records">
              <span>
                {ledgerChangeSummary(job)}
                {#if job.cancellationMode === 'keep_partial_progress'} · partial progress kept{/if}
                {#if job.cancellationMode === 'discard_partial_progress'} · partial progress discarded{/if}
              </span>
            </div>
            <div class="time-cell" role="cell" data-cell-label="Completed">
              {#if job.completedAt}
                <span>{jobTimeLabel('', job.completedAt).trim()}</span>
              {:else if job.startedAt}
                <span>{jobTimeLabel('', job.startedAt).trim()}</span>
              {:else}
                <span>{jobTimeLabel('', job.createdAt).trim()}</span>
              {/if}
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

  .history-header.compact {
    justify-content: flex-end;
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
    padding: 0.6rem 0.7rem;
  }

  .attention-marker {
    color: var(--destructive);
    display: grid;
    place-items: center;
  }

  .attention-alert > div:nth-child(2) {
    display: grid;
    gap: 0.08rem;
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
    display: flex;
    flex-wrap: wrap;
    gap: 0.45rem;
    justify-content: start;
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
    min-height: 2.3rem;
    padding: 0.35rem 0.5rem;
    text-align: left;
  }

  :global(.status-chip:not(:disabled)) {
    cursor: pointer;
  }

  :global(.status-chip.active) {
    background: transparent;
    border-color: var(--border);
    color: var(--foreground);
  }

  :global(.status-chip.warning) {
    background: color-mix(in oklab, var(--color-warning) 7%, transparent);
    border-color: color-mix(in oklab, var(--color-warning) 30%, transparent);
    color: var(--color-warning-foreground);
  }

  :global(.status-chip.danger) {
    background: color-mix(in oklab, var(--destructive) 6%, transparent);
    border-color: color-mix(in oklab, var(--destructive) 30%, transparent);
    color: var(--destructive);
  }

  :global(.status-chip.selected) {
    background: color-mix(in oklab, var(--primary) 7%, transparent);
    border-color: color-mix(in oklab, var(--ring) 34%, var(--border));
    box-shadow: 0 0 0 1px color-mix(in oklab, var(--ring) 16%, transparent);
    color: var(--foreground);
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

  .history-title > span {
    color: var(--muted-foreground);
    flex-basis: 100%;
    font-size: 0.78rem;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
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
    grid-template-columns: minmax(12rem, 1.2fr) minmax(10rem, 0.82fr) minmax(15rem, 1.25fr) minmax(8rem, 0.66fr) auto;
  }

  .history-ledger-head {
    color: var(--muted-foreground);
    font-size: 0.75rem;
    font-weight: 700;
    padding: 0 0.4rem;
    text-transform: uppercase;
  }

  :global(.ledger-sort-button) {
    color: inherit;
    font-size: inherit;
    font-weight: inherit;
    gap: 0.25rem;
    height: 1.65rem;
    justify-content: flex-start;
    letter-spacing: inherit;
    padding: 0 0.32rem;
    text-transform: inherit;
  }

  :global(.ledger-sort-button small) {
    color: var(--foreground);
    font-size: 0.68rem;
    font-weight: 700;
    letter-spacing: 0;
    text-transform: none;
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

  .issue-cell,
  .result-cell,
  .time-cell {
    color: var(--muted-foreground);
    display: grid;
    font-size: 0.82rem;
    gap: 0.18rem;
    min-width: 0;
  }

  .result-cell span:first-child,
  .issue-cell span:first-child {
    align-items: center;
    color: var(--foreground);
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem;
  }

  .result-cell span,
  .issue-cell span,
  .time-cell span {
    overflow-wrap: anywhere;
  }

  .issue-cell {
    align-content: start;
    justify-items: start;
  }

  .result-cell span {
    color: var(--foreground);
  }

  .issue-cell.warning span {
    color: var(--color-warning-foreground);
  }

  :global(.warning-badge) {
    background: color-mix(in oklab, var(--color-warning) 16%, transparent);
    color: var(--color-warning-foreground);
  }

  .issue-cell.action span {
    color: var(--destructive);
  }

  .history-row:hover {
    background: color-mix(in oklab, var(--muted) 25%, transparent);
  }

  :global(.current-work-row.clickable-row:hover) {
    background: color-mix(in oklab, var(--muted) 20%, transparent);
  }

  .clickable-row {
    cursor: pointer;
  }

  .clickable-row:focus-visible,
  :global(.current-work-row.clickable-row:focus-visible) {
    outline: 2px solid var(--ring);
    outline-offset: 2px;
  }

  .history-row.attention-row {
    background: color-mix(in oklab, var(--destructive) 2.6%, transparent);
    border-color: color-mix(in oklab, var(--destructive) 30%, transparent);
  }

  .attention-row .status-icon {
    color: var(--destructive);
  }

  .history-row.warning-row {
    background: color-mix(in oklab, var(--color-warning) 3.5%, transparent);
    border-color: color-mix(in oklab, var(--color-warning) 26%, var(--border));
  }

  .warning-row .status-icon {
    color: var(--color-warning-foreground);
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

    .current-work-section :global(.import-job-card) {
      gap: 0.7rem;
    }

    .current-work-section .job-main {
      align-items: flex-start;
      flex-direction: row;
      width: 100%;
    }

    .current-work-section .action-row {
      flex-direction: row;
      width: 100%;
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

    .status-cell {
      display: block;
    }

    .result-cell,
    .issue-cell,
    .time-cell {
      gap: 0.12rem;
    }

    .history-status-strip {
      flex-wrap: nowrap;
      margin-inline: -0.15rem;
      overflow-x: auto;
      overscroll-behavior-x: contain;
      padding: 0 0.15rem 0.2rem;
      scrollbar-width: none;
    }

    .history-status-strip::-webkit-scrollbar {
      display: none;
    }

    :global(.status-chip) {
      flex: 0 0 auto;
      justify-content: flex-start;
      min-width: 8.25rem;
      min-height: 2.2rem;
      padding: 0.42rem 0.55rem;
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

    .job-section:last-child {
      padding-bottom: var(--mobile-scroll-clearance, 7rem);
    }
  }
</style>
