<script lang="ts">
  import { t } from '$lib/presentation/localization';
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
  import * as Table from '$lib/components/ui/table/index.js';
  import {
    actorSummary,
    attentionSummary,
    canRequestCancellation,
    canRemoveJobFromHistory,
    historyCountSummary,
    importIssueTone,
    issueCountSummary,
    isTerminal,
    shortDateTime,
    phaseLabel,
    progressBarLabel,
    progressBarStyle,
    progressKnown,
    progressPercent,
    progressSummary,
    ledgerChangeSummary,
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

  function sortButtonLabel(key: Exclude<HistorySortKey, 'priority'>): string {
    return t(`web.ImportJobHistory.sort.${key}.${historySortKey === key ? historySortDirection : 'none'}`);
  }

  function sortIndicator(key: Exclude<HistorySortKey, 'priority'>): string {
    if (historySortKey !== key) return '';
    return historySortDirection === 'asc' ? t('web.ImportJobHistory.ascending') : t('web.ImportJobHistory.descending');
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

  function jobActionLabel(action: 'view' | 'review' | 'continue' | 'cancel' | 'remove', job: ImportJob): string {
    const time = job.completedAt ?? job.startedAt ?? job.createdAt;
    return t(`web.ImportJobHistory.action.${action}.${time ? 'timed' : 'plain'}`, {
      name: job.source.name, status: statusLabel(job), time: time ? shortDateTime(time) : ''
    });
  }

  function ledgerDescription(): string {
    if (historyFilter === 'attention') return t('web.ImportJobHistory.importsThatNeedAction');
    if (historyFilter === 'warnings') return t('web.ImportJobHistory.warningOnlyImports');
    if (historyFilter === 'completed') return t('web.ImportJobHistory.completedImports');
    return t('web.ImportJobHistory.needsActionAndRecentRunsFirst');
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
    if (job.status === 'succeeded') return t('web.ImportJobHistory.noActionNeeded');
    return statusSentence(job);
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
      <Plus size={16} aria-hidden="true" /> {t('web.ImportJobHistory.newImport')} </Button.Root>
  {/if}
</div>

{#if jobs.length > 0}
  <div class="history-status-strip" aria-label={t('web.ImportJobHistory.importHistoryFilters')}>
    <Button.Root
      class={historyFilter === 'all' ? 'status-chip selected' : 'status-chip'}
      variant="ghost"
      onclick={() => (historyFilter = 'all')}
      aria-pressed={historyFilter === 'all'}
    >
      <Database size={14} aria-hidden="true" />
      <span>{t('web.ImportJobHistory.allRuns')}</span>
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
        <span>{t('web.ImportJobHistory.running')}</span>
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
        <span>{t('web.ImportJobHistory.readyToReview')}</span>
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
        <span>{t('web.ImportJobHistory.actionRequired')}</span>
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
        <span>{t('web.ImportJobHistory.warnings')}</span>
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
        <span>{t('web.ImportJobHistory.completed')}</span>
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
        <h2>{t('web.ImportJobHistory.noImportRunsYet')}</h2>
        <p>{t('web.ImportJobHistory.startWithHomeboxPreviewWhatStuffStashWillCreate')}</p>
      </div>
      {#if canCreateImports}
        <Button.Root onclick={onBeginImport}>
          <Plus size={16} aria-hidden="true" /> {t('web.ImportJobHistory.newImport')} </Button.Root>
      {:else}
        <p class="empty-state-access-note">{t('web.ImportJobHistory.creatingImportsRequiresImportJobCreateAccess')}</p>
      {/if}
    </Card.Content>
  </Card.Root>
{:else}
  {#if currentWorkJobs.length > 0 && (historyFilter === 'all' || historyFilter === 'current' || historyFilter === 'drafts')}
    <div class="job-section current-work-section">
      <div class="section-heading">
        <div>
          <h3>{t('web.ImportJobHistory.currentWork')}</h3>
          <p>{t('web.ImportJobHistory.resumeDraftsWatchProgressOrCancelRunningImports')}</p>
        </div>
      </div>
      {#if activeJobs.length > 0 && historyFilter !== 'drafts'}
      {#each activeJobs as job}
        <Card.Root>
          <Card.Content
            class="import-job-card current-work-row clickable-row"
            role="button"
            tabindex={0}
            aria-label={jobActionLabel('view', job)}
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
                <span>{job.status === 'cancel_requested' ? statusSentence(job) : actorSummary(job, currentPrincipal) || t('web.ImportJobHistory.backgroundJob')}</span>
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
              <Button.Root variant="ghost" size="sm" onclick={() => onOpenJob(job)} disabled={busy} aria-label={jobActionLabel('view', job)}>
                <Eye size={16} aria-hidden="true" /> {t('web.ImportJobHistory.details')} </Button.Root>
              {#if canRequestCancellation(job)}
                <Button.Root
                  variant="outline"
                  size="sm"
                  onclick={() => onRequestCancellation(job)}
                  disabled={busy || !canCreateImports}
                  aria-label={jobActionLabel('cancel', job)}
                > {t('web.ImportJobHistory.cancel')} </Button.Root>
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
          aria-label={jobActionLabel('view', job)}
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
              <span>{t('import.history.previewed', { time: shortDateTime(job.createdAt) })}</span>
            </div>
          </div>
          <Button.Root variant="outline" size="sm" onclick={() => onResumePreviewedJob(job)} aria-label={jobActionLabel('continue', job)}>{t('web.ImportJobHistory.continue')}</Button.Root>
          <Button.Root variant="ghost" size="sm" onclick={() => onOpenJob(job)} aria-label={jobActionLabel('view', job)}>{t('web.ImportJobHistory.details')}</Button.Root>
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
        <strong>{t('import.requiresAction', { count: attentionJobs.length })}</strong>
        <span>{t('import.blockingIssues', { count: attentionJobs.length })}</span>
      </div>
      <Button.Root variant="outline" size="sm" onclick={() => (historyFilter = 'attention')}>{t('web.ImportJobHistory.showOnlyThose')}</Button.Root>
    </div>
  {/if}
  {#if terminalJobs.length > 0 && historyFilter !== 'current' && historyFilter !== 'drafts'}
    <div class="job-section">
      <div class="ledger-heading">
        <div>
          <h3>{t('web.ImportJobHistory.runs')}</h3>
          <p>{ledgerDescription()}</p>
        </div>
        {#if historyFilter !== 'all'}
          <Button.Root variant="ghost" size="sm" onclick={() => (historyFilter = 'all')}>{t('web.ImportJobHistory.showAll')}</Button.Root>
        {/if}
      </div>
      <Table.Root class="history-ledger" aria-label={t('web.ImportJobHistory.importHistory')}>
        <Table.Header class="history-ledger-head">
          <Table.Row>
            <Table.Head scope="col" aria-sort={historySortKey === 'source' ? (historySortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>
              <Button.Root variant="ghost" size="xs" class="ledger-sort-button" onclick={() => setHistorySort('source')} aria-label={sortButtonLabel('source')}> {t('web.ImportJobHistory.source')} <ArrowUpDown size={12} aria-hidden="true" />
                {#if sortIndicator('source')}<small>{sortIndicator('source')}</small>{/if}
              </Button.Root>
            </Table.Head>
            <Table.Head scope="col" aria-sort={historySortKey === 'status' ? (historySortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>
              <Button.Root variant="ghost" size="xs" class="ledger-sort-button" onclick={() => setHistorySort('status')} aria-label={sortButtonLabel('status')}> {t('web.ImportJobHistory.status')} <ArrowUpDown size={12} aria-hidden="true" />
                {#if sortIndicator('status')}<small>{sortIndicator('status')}</small>{/if}
              </Button.Root>
            </Table.Head>
            <Table.Head scope="col" aria-sort={historySortKey === 'changed' ? (historySortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>
              <Button.Root variant="ghost" size="xs" class="ledger-sort-button" onclick={() => setHistorySort('changed')} aria-label={sortButtonLabel('changed')}> {t('web.ImportJobHistory.changed')} <ArrowUpDown size={12} aria-hidden="true" />
                {#if sortIndicator('changed')}<small>{sortIndicator('changed')}</small>{/if}
              </Button.Root>
            </Table.Head>
            <Table.Head scope="col" aria-sort={historySortKey === 'finished' ? (historySortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}>
              <Button.Root variant="ghost" size="xs" class="ledger-sort-button" onclick={() => setHistorySort('finished')} aria-label={sortButtonLabel('finished')}> {t('web.ImportJobHistory.finished')} <ArrowUpDown size={12} aria-hidden="true" />
                {#if sortIndicator('finished')}<small>{sortIndicator('finished')}</small>{/if}
              </Button.Root>
            </Table.Head>
            <Table.Head scope="col" class="w-px whitespace-nowrap">{t('web.ImportJobHistory.actions')}</Table.Head>
          </Table.Row>
        </Table.Header>
        <Table.Body>
        {#each filteredTerminalJobs as job}
          <Table.Row
            class={jobRequiresAction(job) ? 'history-row attention-row clickable-row' : jobHasReviewWarnings(job) ? 'history-row warning-row clickable-row' : 'history-row clickable-row'}
            tabindex={0}
            aria-label={jobActionLabel(jobRequiresAction(job) || jobHasReviewWarnings(job) ? 'review' : 'view', job)}
            onclick={(event) => openHistoryRow(event, job)}
            onkeydown={(event) => openHistoryRowFromKeyboard(event, job)}
          >
            <Table.Cell class="status-cell" data-cell-label={t('web.ImportJobHistory.source')}>
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
            </Table.Cell>
            <Table.Cell class={jobRequiresAction(job) ? 'issue-cell action' : jobHasReviewWarnings(job) ? 'issue-cell warning' : 'issue-cell'} data-cell-label={t('web.ImportJobHistory.status')}>
              {#if jobRequiresAction(job)}
                <Badge variant="destructive">{t('web.ImportJobHistory.actionRequired')}</Badge>
              {:else if jobHasReviewWarnings(job)}
                <Badge variant="secondary" class="warning-badge">{t('web.ImportJobHistory.warnings')}</Badge>
              {:else}
                <Badge variant={statusVariant(job)}>{statusLabel(job)}</Badge>
              {/if}
              <span>{statusDetail(job)}</span>
            </Table.Cell>
            <Table.Cell class="result-cell" data-cell-label={t('web.ImportJobHistory.changed')}>
              <span>
                {ledgerChangeSummary(job)}
                {#if job.cancellationMode === 'keep_partial_progress'} {t('web.ImportJobHistory.partialProgressKept')}{/if}
                {#if job.cancellationMode === 'discard_partial_progress'} {t('web.ImportJobHistory.partialProgressDiscarded')}{/if}
              </span>
            </Table.Cell>
            <Table.Cell class="time-cell" data-cell-label={t('web.ImportJobHistory.finished')}>
              {#if job.completedAt}
                <span>{shortDateTime(job.completedAt)}</span>
              {:else if job.startedAt}
                <span>{shortDateTime(job.startedAt)}</span>
              {:else}
                <span>{shortDateTime(job.createdAt)}</span>
              {/if}
            </Table.Cell>
            <Table.Cell class="row-actions" data-cell-label={t('web.ImportJobHistory.actions')}>
              <Button.Root
                variant={jobRequiresAction(job) ? 'outline' : 'ghost'}
                size="sm"
                onclick={() => onOpenJob(job)}
                aria-label={jobActionLabel(jobRequiresAction(job) || jobHasReviewWarnings(job) ? 'review' : 'view', job)}
              >
                <Eye size={16} aria-hidden="true" />
                {jobRequiresAction(job) || jobHasReviewWarnings(job) ? t('web.ImportJobHistory.reviewDetails') : t('web.ImportJobHistory.details')}
              </Button.Root>
              {#if canRemoveJobFromHistory(job)}
                <Button.Root variant="ghost" size="icon" onclick={() => onRequestRemove(job)} aria-label={jobActionLabel('remove', job)}>
                  <Trash2 size={16} aria-hidden="true" />
                </Button.Root>
              {/if}
            </Table.Cell>
          </Table.Row>
        {/each}
        </Table.Body>
      </Table.Root>
      {#if filteredTerminalJobs.length === 0}
        <div class="quiet-row">
          <CheckCircle2 size={16} aria-hidden="true" />
          {historyFilter === 'all' && attentionJobs.length > 0 ? t('web.ImportJobHistory.noOtherImportRunsToShow') : t('web.ImportJobHistory.noImportsMatchThisFilter')}
        </div>
      {/if}
    </div>
  {:else if currentWorkJobs.length > 0}
    <div class="quiet-row">
      <CheckCircle2 size={16} aria-hidden="true" /> {t('web.ImportJobHistory.noCompletedImportRunsYet')} </div>
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

  .draft-row.history-row {
    align-items: center;
    border: 1px solid var(--border);
    border-radius: 8px;
    display: grid;
    gap: 0.55rem;
    grid-template-columns: auto minmax(0, 1fr) auto;
    padding: 0.58rem 0.7rem;
  }

  :global(.history-ledger) {
    table-layout: auto;
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

  :global(.history-ledger th:nth-child(1)),
  :global(.history-ledger td:nth-child(1)) {
    width: 31%;
  }

  :global(.history-ledger th:nth-child(2)),
  :global(.history-ledger td:nth-child(2)) {
    width: 21%;
  }

  :global(.history-ledger th:nth-child(3)),
  :global(.history-ledger td:nth-child(3)) {
    width: 25%;
  }

  :global(.history-ledger th:nth-child(4)),
  :global(.history-ledger td:nth-child(4)) {
    width: 16%;
  }

  :global(.history-ledger-head) {
    color: var(--muted-foreground);
    font-size: 0.75rem;
    font-weight: 700;
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

  .draft-row.history-row > div,
  :global(.history-ledger .history-row > td) {
    min-width: 0;
  }

  :global(.history-ledger .status-cell) {
    align-items: flex-start;
    display: flex;
    gap: 0.6rem;
    min-width: 0;
  }

  :global(.history-ledger .issue-cell),
  :global(.history-ledger .result-cell),
  :global(.history-ledger .time-cell) {
    color: var(--muted-foreground);
    font-size: 0.82rem;
    gap: 0.18rem;
    min-width: 0;
  }

  :global(.history-ledger .issue-cell) {
    align-content: start;
    justify-items: start;
  }

  :global(.history-ledger .issue-cell > span) {
    display: block;
    margin-top: 0.25rem;
  }

  :global(.history-ledger .result-cell span:first-child),
  :global(.history-ledger .issue-cell span:first-child) {
    align-items: center;
    color: var(--foreground);
    flex-wrap: wrap;
    gap: 0.35rem;
  }

  :global(.history-ledger .result-cell span),
  :global(.history-ledger .issue-cell span),
  :global(.history-ledger .time-cell span) {
    overflow-wrap: anywhere;
  }

  :global(.history-ledger .result-cell span) {
    color: var(--foreground);
  }

  :global(.history-ledger .issue-cell.warning span) {
    color: var(--color-warning-foreground);
  }

  :global(.warning-badge) {
    background: color-mix(in oklab, var(--color-warning) 16%, transparent);
    color: var(--color-warning-foreground);
  }

  :global(.history-ledger .issue-cell.action span) {
    color: var(--destructive);
  }

  .draft-row.history-row:hover,
  :global(.history-ledger .history-row:hover) {
    background: color-mix(in oklab, var(--muted) 25%, transparent);
  }

  :global(.current-work-row.clickable-row:hover) {
    background: color-mix(in oklab, var(--muted) 20%, transparent);
  }

  .clickable-row,
  :global(.history-ledger .clickable-row) {
    cursor: pointer;
  }

  .clickable-row:focus-visible,
  :global(.history-ledger .clickable-row:focus-visible),
  :global(.current-work-row.clickable-row:focus-visible) {
    outline: 2px solid var(--ring);
    outline-offset: 2px;
  }

  :global(.history-ledger .history-row.attention-row) {
    background: color-mix(in oklab, var(--destructive) 2.6%, transparent);
  }

  :global(.history-ledger .attention-row .status-icon) {
    color: var(--destructive);
  }

  :global(.history-ledger .history-row.warning-row) {
    background: color-mix(in oklab, var(--color-warning) 3.5%, transparent);
  }

  :global(.history-ledger .warning-row .status-icon) {
    color: var(--color-warning-foreground);
  }

  :global(.history-ledger .row-actions) {
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
    .history-header :global([data-slot='button']),
    .history-status-strip :global([data-slot='button']),
    .action-row :global([data-slot='button']),
    .attention-alert :global([data-slot='button']),
    :global(.history-ledger .row-actions [data-slot='button']),
    :global(.import-history-empty-state [data-slot='button']) {
      min-height: 44px;
      min-width: 44px;
    }

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

    :global(.history-ledger thead) {
      display: none;
    }

    :global(.history-ledger),
    :global(.history-ledger tbody),
    :global(.history-ledger tr),
    :global(.history-ledger td) {
      display: block;
      width: 100% !important;
    }

    :global(.history-ledger tbody) {
      display: grid;
      gap: 0.55rem;
    }

    :global(.history-ledger .history-row) {
      border-top: 1px solid var(--border);
      display: grid;
      gap: 0.45rem;
      padding: 0.65rem;
    }

    :global(.history-ledger .history-row:first-child) {
      border-top: 0;
    }

    :global(.history-ledger td) {
      border-top: 0;
      display: grid;
      gap: 0.18rem;
      grid-template-columns: minmax(5.5rem, 0.34fr) minmax(0, 1fr);
      padding: 0;
    }

    :global(.history-ledger td::before) {
      color: var(--muted-foreground);
      content: attr(data-cell-label);
      font-size: 0.7rem;
      font-weight: 700;
      letter-spacing: 0;
      text-transform: uppercase;
    }

    :global(.history-ledger .status-cell) {
      display: block;
    }

    :global(.history-ledger .result-cell),
    :global(.history-ledger .issue-cell),
    :global(.history-ledger .time-cell) {
      gap: 0.12rem;
    }

    :global(.history-ledger .row-actions) {
      justify-content: flex-start;
    }

    .history-status-strip {
      flex-wrap: nowrap;
      margin-inline: -0.15rem;
      overflow-x: auto;
      overscroll-behavior-x: contain;
      padding: 0 1.5rem 0.2rem 0.15rem;
      scroll-padding-inline: 0.15rem 1.5rem;
      scrollbar-width: none;
      mask-image: linear-gradient(to right, #000 0, #000 calc(100% - 1.25rem), transparent 100%);
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
