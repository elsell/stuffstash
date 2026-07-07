<script lang="ts">
  import Activity from '@lucide/svelte/icons/activity';
  import AlertCircle from '@lucide/svelte/icons/alert-circle';
  import CheckCircle2 from '@lucide/svelte/icons/check-circle-2';
  import MoreHorizontal from '@lucide/svelte/icons/more-horizontal';
  import RefreshCw from '@lucide/svelte/icons/refresh-cw';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import type { ImportJob, Principal } from '$lib/domain/inventory';
  import type { ImportDetailTabRoute } from '$lib/application/workspaceRoute';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import * as Tabs from '$lib/components/ui/tabs/index.js';
  import {
    actorSummary,
    canRemoveJobFromHistory,
    changedRecordSummary,
    type CountCell,
    importIssueTone,
    isTerminal,
    issueCountSummary,
    issueTotalCount,
    phaseLabel,
    progressBarLabel,
    progressBarStyle,
    progressKnown,
    progressPercent,
    progressTimeline,
    resourceDiagnosticLabel,
    resourceLabel,
    resultCountCells,
    reportedErrorCount,
    reportedWarningCount,
    sourceDescription,
    sourceOptionsSummary,
    statusLabel,
    statusSentence,
    uniqueImportMessages,
    visibleCountCells,
    visiblePreviewCountCells
  } from './importWorkspacePresentation';
  import ImportCountGrid from './ImportCountGrid.svelte';
  import ImportMessagesList from './ImportMessagesList.svelte';
  import ImportPreviewSamples from './ImportPreviewSamples.svelte';

  const RESOURCE_PAGE_SIZE = 25;

  type ImportResource = ImportJob['resources'][number];

  type Props = {
    job: ImportJob;
    selectedTab: ImportDetailTabRoute;
    canRequestCancellation: boolean;
    detailLoading: boolean;
    canCreateImports: boolean;
    busy: boolean;
    currentPrincipal?: Principal;
    auditHistoryHref: string;
    resourceCanOpen: (job: ImportJob, resource: ImportResource) => boolean;
    resourceHref: (resource: ImportResource) => string;
    onOpenAuditHistory: (event: MouseEvent) => void;
    onOpenResource: (event: MouseEvent, resource: ImportResource) => void;
    onCancel: () => void;
    onContinue: () => void;
    onRemove: () => void;
  };

  let {
    job,
    selectedTab = $bindable(),
    canRequestCancellation,
    detailLoading,
    canCreateImports,
    busy,
    currentPrincipal,
    auditHistoryHref,
    resourceCanOpen,
    resourceHref,
    onOpenAuditHistory,
    onOpenResource,
    onCancel,
    onContinue,
    onRemove
  }: Props = $props();

  let resourcePage = $state(0);
  let actionMenuOpen = $state(false);
  let issueCount = $derived(issueTotalCount(job));
  let issueTone = $derived(importIssueTone(job));
  let resourcePageCount = $derived(Math.max(1, Math.ceil(job.resources.length / RESOURCE_PAGE_SIZE)));
  let visibleResourceStart = $derived(Math.min(resourcePage * RESOURCE_PAGE_SIZE, Math.max(0, job.resources.length - 1)));
  let visibleResources = $derived(job.resources.slice(visibleResourceStart, visibleResourceStart + RESOURCE_PAGE_SIZE));
  let visibleResourceEnd = $derived(Math.min(job.resources.length, visibleResourceStart + visibleResources.length));
  let actor = $derived(actorSummary(job, currentPrincipal));
  let visibleSourceOptions = $derived(sourceOptionsSummary(job));
  let overviewCells = $derived(detailOverviewCells(job));

  $effect(() => {
    job.id;
    resourcePage = 0;
    actionMenuOpen = false;
  });

  $effect(() => {
    if (resourcePage >= resourcePageCount) {
      resourcePage = Math.max(0, resourcePageCount - 1);
    }
  });

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
    const messages = job.messages.length > 0 ? job.messages : job.preview.messages;
    return uniqueImportMessages(messages);
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

  function detailOverviewCells(job: ImportJob): CountCell[] {
    const cells = job.status === 'previewed' ? visiblePreviewCountCells(job) : visibleCountCells(resultCountCells(job));
    return cells.map((cell) => {
      const label = cell.label.toLowerCase();
      if (label.includes('blocking')) {
        return {
          ...cell,
          tone: cell.value > 0 ? 'action' : 'muted',
          actionLabel: cell.value > 0 ? `Open issues for ${cell.value} ${cell.label}` : undefined
        };
      }
      if (label.includes('warning')) {
        return {
          ...cell,
          tone: 'warning',
          actionLabel: cell.value > 0 ? `Open issues for ${cell.value} ${cell.label}` : undefined
        };
      }
      if ((label.includes('created') || label.includes('imported') || label.includes('saved')) && !label.includes('field')) {
        return {
          ...cell,
          tone: cell.value > 0 ? 'success' : 'muted',
          actionLabel: job.resources.length > 0 && cell.value > 0 ? `Open imported records for ${cell.value} ${cell.label}` : undefined
        };
      }
      return { ...cell, tone: cell.muted ? 'muted' : 'default' };
    });
  }

  function handleOverviewCellAction(cell: CountCell): void {
    const label = cell.label.toLowerCase();
    if (label.includes('warning') || label.includes('blocking')) {
      selectedTab = 'issues';
      return;
    }
    if (cell.actionLabel?.includes('imported records')) {
      selectedTab = 'records';
    }
  }
</script>

<Card.Root>
  <Card.Content>
    <div class="import-detail-content">
      <div class="detail-hero">
        <div class="job-heading">
          <div>
            <strong>{job.source.name}</strong>
            <span>{sourceDescription(job)}</span>
          </div>
          <Badge variant={job.status === 'failed' || job.status === 'discard_failed' ? 'destructive' : 'secondary'}>
            {statusLabel(job)}
          </Badge>
        </div>
        {#if !isTerminal(job)}
          <div class="detail-topline">
            <div>
              <strong>{phaseLabel(job)}</strong>
              <span>{statusSentence(job)}</span>
            </div>
          </div>
        {/if}
        {#if !isTerminal(job)}
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
        <div class="detail-summary-strip" aria-label="Import run summary">
          <div class="summary-tile">
            <span>Changed</span>
            <strong>{changedRecordSummary(job)}</strong>
          </div>
          <div class={`summary-tile ${issueTone}`}>
            <span>Issues</span>
            <strong>{issueCount === 0 ? 'No issues' : issueTone === 'action' ? 'Action required' : issueCountSummary(job)}</strong>
          </div>
        </div>
        {#if issueCount > 0 && issueTone === 'action' && selectedTab !== 'issues'}
          <div class={`detail-issue-callout ${issueTone}`}>
            <AlertCircle class="issue-callout-icon" size={18} aria-hidden="true" />
            <div>
              <strong>{issueCountSummary(job)}</strong>
              <span>Review the blocking items before relying on this run.</span>
            </div>
            <Button.Root variant="outline" size="sm" onclick={() => (selectedTab = 'issues')}>Review issues</Button.Root>
          </div>
        {/if}
      </div>

      {#if detailLoading}
        <div class="quiet-row">
          <RefreshCw size={16} aria-hidden="true" />
          Refreshing import details.
        </div>
      {/if}

      <div class="detail-grid">
        <div class="detail-main">
          <Tabs.Root bind:value={selectedTab} class="detail-tabs">
            <Tabs.List variant="line" aria-label="Import detail sections" class="detail-tab-list">
              <Tabs.Trigger value="overview">Overview</Tabs.Trigger>
              <Tabs.Trigger value="issues">Issues{issueCount > 0 ? ` (${issueCount})` : ''}</Tabs.Trigger>
              <Tabs.Trigger value="plan">Plan</Tabs.Trigger>
              <Tabs.Trigger value="records">Records</Tabs.Trigger>
              <Tabs.Trigger value="timeline" aria-label="Timeline">Log</Tabs.Trigger>
            </Tabs.List>

            <Tabs.Content value="overview">
              <section class="detail-panel" aria-label="Import result overview">
                <div class="sample-heading">
                  <h3>Result</h3>
                  <small>{statusSentence(job)}</small>
                </div>
                <ImportCountGrid cells={overviewCells} actionForCell={handleOverviewCellAction} />
              </section>
            </Tabs.Content>

            <Tabs.Content value="issues">
              <section class="detail-panel" aria-label="Import issues">
                <div class="sample-heading">
                  <h3>Issues</h3>
                  <small>{issueCount === 0 ? 'No issues' : 'Grouped by cause'}</small>
                </div>
                <ImportMessagesList
                  messages={detailMessages(job)}
                  emptyText="No import messages."
                  truncated={job.messages.length === 0 && job.preview.messagesTruncated}
                  reportedWarnings={reportedWarningCount(job)}
                  reportedErrors={reportedErrorCount(job)}
                />
              </section>
            </Tabs.Content>

            <Tabs.Content value="plan">
              {#if hasPreviewPlan(job)}
                <section class="detail-panel" aria-label="Import preview plan">
                  <div class="sample-heading">
                    <h3>Preview plan</h3>
                    <small>{job.status === 'previewed' ? 'Before import' : 'Original plan'}</small>
                  </div>
                  <ImportPreviewSamples preview={job.preview} />
                </section>
              {:else}
                <div class="quiet-row">
                  <CheckCircle2 size={16} aria-hidden="true" />
                  No preview plan is available for this run.
                </div>
              {/if}
            </Tabs.Content>

            <Tabs.Content value="records">
              {#if job.status === 'cancelled_discarded' && job.resources.length > 0}
                <div class="quiet-row">
                  <CheckCircle2 size={16} aria-hidden="true" />
                  Records created by this job were discarded. Audit history remains.
                </div>
              {:else if job.resources.length > 0}
                <section class="detail-panel">
                  <div class="sample-heading">
                    <h3>Imported records</h3>
                    {#if job.resources.length > RESOURCE_PAGE_SIZE}
                      <small>{visibleResourceStart + 1}-{visibleResourceEnd} of {job.resources.length}</small>
                    {/if}
                  </div>
                  <div class="resource-table-wrap">
                    <table class="resource-list" aria-label="Imported records">
                      <thead>
                        <tr>
                          <th scope="col">Record</th>
                          <th scope="col">Source</th>
                          <th scope="col">Open</th>
                        </tr>
                      </thead>
                      <tbody>
                        {#each visibleResources as resource}
                          <tr>
                            <td class="resource-name-cell">{resourceLabel(resource)}</td>
                            <td>{resourceDiagnosticLabel(resource)} · Imported {new Date(resource.createdAt).toLocaleString()}</td>
                            <td>
                              {#if resourceCanOpen(job, resource)}
                                <a class="resource-link" href={resourceHref(resource)} onclick={(event) => onOpenResource(event, resource)}>Open</a>
                              {:else}
                                <span class="resource-empty-action">-</span>
                              {/if}
                            </td>
                          </tr>
                        {/each}
                      </tbody>
                    </table>
                  </div>
                  {#if job.resources.length > RESOURCE_PAGE_SIZE}
                    <div class="resource-overflow-action">
                      <span>Page {resourcePage + 1} of {resourcePageCount}</span>
                      <Button.Root
                        variant="outline"
                        size="sm"
                        disabled={resourcePage === 0}
                        onclick={() => (resourcePage = Math.max(0, resourcePage - 1))}
                      >
                        Previous
                      </Button.Root>
                      <Button.Root
                        variant="outline"
                        size="sm"
                        disabled={resourcePage >= resourcePageCount - 1}
                        onclick={() => (resourcePage = Math.min(resourcePageCount - 1, resourcePage + 1))}
                      >
                        Next
                      </Button.Root>
                    </div>
                  {/if}
                </section>
              {:else}
                <div class="quiet-row">
                  <CheckCircle2 size={16} aria-hidden="true" />
                  No imported record summaries are available for this run.
                </div>
              {/if}
            </Tabs.Content>

            <Tabs.Content value="timeline">
              <section class="detail-panel">
                <div class="sample-heading">
                  <h3>Progress timeline</h3>
                  <small>{progressTimeline(job).length} phases</small>
                </div>
                <div class="timeline-list">
                  {#each progressTimeline(job) as progress}
                    <div class="timeline-row">
                      <span aria-hidden="true"></span>
                      <div>
                        <strong>{phaseLabel({ ...job, progress })}</strong>
                        <small>
                          {progress.total > 0 ? `${Math.min(progress.done, progress.total)} / ${progress.total}` : ''}
                          {progress.total > 0 && progress.updatedAt ? ' · ' : ''}
                          {progress.updatedAt ? new Date(progress.updatedAt).toLocaleString() : ''}
                        </small>
                      </div>
                    </div>
                  {/each}
                </div>
              </section>
            </Tabs.Content>
          </Tabs.Root>
        </div>

        <div class="detail-side" aria-label="Import controls and source">
          <section class="source-options-section" aria-label="Import run details">
            <div class="sample-heading">
              <h3>Method</h3>
              <small>How Stuff Stash read the source</small>
            </div>
            {#if actor}
              <p class="source-note">{actor}</p>
            {/if}
            {#if visibleSourceOptions.length > 0}
              <ul class="source-option-list" aria-label="Selected source options">
                {#each visibleSourceOptions as option}
                  <li>{option}</li>
                {/each}
              </ul>
            {/if}
          </section>
          {#if job.cancellationMode}
            <div class="quiet-row">
              <AlertCircle size={16} aria-hidden="true" />
              {cancellationSummary(job)}
            </div>
          {/if}
          <section class="detail-actions" aria-label="Import actions">
            {#if canCreateImports && canRequestCancellation}
              <Button.Root variant="outline" onclick={onCancel} disabled={busy}>
                <Button.BusyContent {busy} label="Cancel" busyLabel="Cancelling" />
              </Button.Root>
            {/if}
            {#if canCreateImports && job.status === 'previewed'}
              <Button.Root onclick={onContinue} disabled={busy}>
                <Button.BusyContent {busy} label="Continue import" busyLabel="Opening preview" />
              </Button.Root>
            {/if}
            <div class="detail-more-actions">
              <Button.Root
                variant="outline"
                size="sm"
                aria-expanded={actionMenuOpen}
                aria-controls="import-detail-secondary-actions"
                onclick={() => (actionMenuOpen = !actionMenuOpen)}
              >
                <MoreHorizontal size={16} aria-hidden="true" />
                More
              </Button.Root>
              {#if actionMenuOpen}
                <div id="import-detail-secondary-actions" class="detail-action-menu" role="menu" aria-label="More import actions">
                  <a class="detail-action-item" role="menuitem" href={auditHistoryHref} onclick={onOpenAuditHistory}>
                    <Activity size={16} aria-hidden="true" />
                    <span>
                      <strong>Open inventory activity</strong>
                      <small>Shows the full inventory activity log.</small>
                    </span>
                  </a>
                  {#if canCreateImports && canRemoveJobFromHistory(job)}
                    <Button.Root variant="ghost" class="detail-action-item danger" role="menuitem" onclick={onRemove} disabled={busy}>
                      <Button.BusyContent {busy} icon={Trash2} label="Remove from history" busyLabel="Removing from history" />
                      <small>Imported records and audit history remain.</small>
                    </Button.Root>
                  {/if}
                </div>
              {/if}
            </div>
          </section>
        </div>
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

  .quiet-row,
  .job-heading,
  .detail-topline {
    align-items: center;
    display: flex;
    gap: 0.75rem;
  }

  .detail-topline {
    justify-content: space-between;
    min-width: 0;
  }

  .job-heading {
    justify-content: space-between;
    min-width: 0;
  }

  .job-heading > div {
    min-width: 0;
  }

  .job-heading span {
    color: var(--muted-foreground);
    display: block;
    font-size: 0.85rem;
    margin-top: 0.15rem;
    overflow-wrap: anywhere;
  }

  .detail-hero {
    border-bottom: 1px solid var(--border);
    display: grid;
    gap: 0.75rem;
    padding-bottom: 1rem;
  }

  .detail-summary-strip {
    display: grid;
    gap: 0.5rem;
    grid-template-columns: minmax(10rem, 0.72fr) minmax(0, 1fr);
  }

  .summary-tile {
    background: color-mix(in oklab, var(--muted) 38%, transparent);
    border: 1px solid var(--border);
    border-radius: 8px;
    display: grid;
    gap: 0.2rem;
    min-width: 0;
    padding: 0.7rem;
  }

  .summary-tile.warning {
    background: color-mix(in oklab, var(--color-warning) 7%, transparent);
    border-color: color-mix(in oklab, var(--color-warning) 26%, transparent);
  }

  .summary-tile.action {
    background: color-mix(in oklab, var(--destructive) 5.5%, transparent);
    border-color: color-mix(in oklab, var(--destructive) 22%, transparent);
  }

  .summary-tile span {
    color: var(--muted-foreground);
    font-size: 0.72rem;
    font-weight: 700;
    letter-spacing: 0;
    text-transform: uppercase;
  }

  .summary-tile strong {
    font-size: 0.95rem;
    min-width: 0;
    overflow-wrap: anywhere;
  }

  .detail-issue-callout {
    align-items: center;
    border-radius: 8px;
    border: 1px solid var(--border);
    display: grid;
    gap: 0.75rem;
    grid-template-columns: auto minmax(0, 1fr) auto;
    padding: 0.75rem;
  }

  .detail-issue-callout.warning {
    background: color-mix(in oklab, var(--color-warning) 7%, transparent);
    border-color: color-mix(in oklab, var(--color-warning) 26%, transparent);
  }

  .detail-issue-callout.action {
    background: color-mix(in oklab, var(--destructive) 5.5%, transparent);
    border-color: color-mix(in oklab, var(--destructive) 22%, transparent);
  }

  :global(.detail-issue-callout.warning .issue-callout-icon) {
    color: var(--color-warning-foreground);
  }

  :global(.detail-issue-callout.action .issue-callout-icon) {
    color: var(--destructive);
  }

  .detail-issue-callout div {
    display: grid;
    gap: 0.15rem;
    min-width: 0;
  }

  .detail-issue-callout span {
    color: var(--muted-foreground);
    font-size: 0.82rem;
  }

  .detail-grid {
    align-items: start;
    display: grid;
    gap: 1rem;
    grid-template-columns: minmax(0, 1fr) minmax(16rem, 0.36fr);
  }

  .detail-main,
  .detail-side,
  .detail-actions {
    display: grid;
    gap: 1rem;
    min-width: 0;
  }

  .detail-main {
    overflow: hidden;
  }

  :global(.detail-tab-list) {
    max-width: 100%;
    overflow-x: auto;
    overscroll-behavior-x: contain;
    scrollbar-width: none;
    scroll-snap-type: x proximity;
    width: 100%;
  }

  :global(.detail-tab-list::-webkit-scrollbar) {
    display: none;
  }

  :global(.detail-tab-list [data-slot='tabs-trigger']) {
    flex: 0 0 auto;
    min-width: max-content;
    scroll-snap-align: start;
  }

  .detail-panel {
    display: grid;
    gap: 1rem;
    min-width: 0;
    padding-top: 0.75rem;
  }

  .detail-side {
    position: sticky;
    top: 1rem;
  }

  .detail-actions {
    border-top: 1px solid var(--border);
    justify-items: start;
    padding-top: 0.75rem;
  }

  .detail-more-actions {
    display: grid;
    gap: 0.5rem;
    justify-items: start;
    position: relative;
    width: 100%;
  }

  .detail-action-menu {
    background: var(--background);
    border: 1px solid var(--border);
    border-radius: 8px;
    box-shadow: 0 1rem 2.5rem color-mix(in oklab, var(--foreground) 12%, transparent);
    display: grid;
    gap: 0.25rem;
    min-width: min(100%, 17rem);
    padding: 0.35rem;
    z-index: 2;
  }

  .detail-action-item {
    align-items: flex-start;
    border-radius: 6px;
    color: var(--foreground);
    display: flex;
    gap: 0.55rem;
    min-width: 0;
    padding: 0.55rem 0.6rem;
    text-align: left;
    text-decoration: none;
    width: 100%;
  }

  :global(.detail-action-item[data-slot='button']) {
    height: auto;
    justify-content: flex-start;
    white-space: normal;
  }

  .detail-action-item:hover,
  .detail-action-item:focus-visible {
    background: color-mix(in oklab, var(--muted) 36%, transparent);
    text-decoration: none;
  }

  .detail-action-item span {
    display: grid;
    gap: 0.12rem;
    min-width: 0;
  }

  .detail-action-item small {
    color: var(--muted-foreground);
    display: block;
    font-size: 0.76rem;
    font-weight: 400;
    line-height: 1.3;
    margin-top: 0.14rem;
    overflow-wrap: anywhere;
  }

  .detail-action-item.danger {
    color: var(--destructive);
  }

  .detail-topline span {
    color: var(--muted-foreground);
    display: block;
    font-size: 0.85rem;
    margin-top: 0.15rem;
    overflow-wrap: anywhere;
  }

  .progress-track {
    background: var(--muted);
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
    background: var(--primary);
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

  .source-option-list,
  .timeline-list {
    display: grid;
  }

  .resource-overflow-action {
    align-items: center;
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    justify-content: flex-end;
  }

  .resource-overflow-action span {
    color: var(--muted-foreground);
    font-size: 0.82rem;
    margin-right: auto;
  }

  .timeline-list {
    gap: 0.35rem;
  }

  .source-options-section {
    border-top: 1px solid var(--border);
    display: grid;
    gap: 0.65rem;
    padding-top: 0.75rem;
  }

  .source-option-list {
    gap: 0.45rem;
    grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .source-option-list li {
    border-left: 2px solid var(--border);
    color: var(--foreground);
    font-size: 0.82rem;
    min-width: 0;
    overflow-wrap: anywhere;
    padding: 0.2rem 0 0.2rem 0.55rem;
  }

  .source-note {
    color: var(--muted-foreground);
    font-size: 0.82rem;
    margin: 0;
    overflow-wrap: anywhere;
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

  .sample-heading small {
    color: var(--muted-foreground);
    font-size: 0.78rem;
  }

  .timeline-row {
    display: grid;
    gap: 0.55rem;
    grid-template-columns: auto minmax(0, 1fr);
    min-width: 0;
  }

  .timeline-row > span {
    background: var(--primary);
    border-radius: 999px;
    height: 0.5rem;
    margin-top: 0.35rem;
    width: 0.5rem;
  }

  .timeline-row strong {
    display: block;
    font-weight: 600;
  }

  .timeline-row small {
    color: var(--muted-foreground);
    overflow-wrap: anywhere;
  }

  .sample-row,
  .resource-list {
    min-width: 0;
  }

  .resource-table-wrap {
    border: 1px solid var(--border);
    border-radius: 8px;
    min-width: 0;
    overflow-x: auto;
  }

  .resource-list {
    border-collapse: collapse;
    width: 100%;
  }

  .resource-list th,
  .resource-list td {
    border-top: 1px solid var(--border);
    font-size: 0.82rem;
    min-width: 8rem;
    padding: 0.55rem 0.65rem;
    text-align: left;
    vertical-align: top;
  }

  .resource-list th {
    background: color-mix(in oklab, var(--muted) 32%, transparent);
    border-top: 0;
    color: var(--muted-foreground);
    font-size: 0.72rem;
    font-weight: 700;
    text-transform: uppercase;
  }

  .resource-list td {
    color: var(--muted-foreground);
  }

  .resource-list th:first-child,
  .resource-list td:first-child {
    min-width: 12rem;
  }

  .resource-list th:last-child,
  .resource-list td:last-child {
    min-width: 5rem;
    white-space: nowrap;
    width: 1%;
  }

  .resource-name-cell {
    color: var(--foreground);
    font-weight: 600;
  }

  .sample-row span,
  .resource-list td {
    overflow-wrap: anywhere;
  }

  .resource-link,
  .detail-link {
    align-items: center;
    color: var(--primary);
    display: inline-flex;
    font-size: 0.85rem;
    font-weight: 600;
    gap: 0.4rem;
    text-decoration: none;
  }

  .resource-link {
    justify-self: start;
  }

  .resource-empty-action {
    color: var(--muted-foreground);
    justify-self: start;
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
    .detail-topline {
      align-items: flex-start;
      flex-direction: column;
    }

    .job-heading {
      align-items: flex-start;
      flex-direction: column;
    }

    .sample-heading {
      align-items: flex-start;
      display: grid;
      gap: 0.2rem;
      justify-content: start;
    }

    .detail-grid {
      grid-template-columns: 1fr;
    }

    .detail-summary-strip {
      gap: 0.5rem;
      grid-template-columns: 1fr;
    }

    .summary-tile {
      padding: 0.65rem;
    }

    .summary-tile.warning {
      border-color: color-mix(in oklab, var(--color-warning) 26%, transparent);
    }

    .summary-tile.action {
      border-color: color-mix(in oklab, var(--destructive) 22%, transparent);
    }

    .summary-tile.warning strong {
      color: var(--color-warning-foreground);
    }

    .summary-tile.action strong {
      color: var(--destructive);
    }

    .summary-tile strong {
      font-size: 0.95rem;
      font-weight: 650;
    }

    .detail-issue-callout {
      align-items: flex-start;
      grid-template-columns: 1fr;
      scroll-margin-bottom: var(--mobile-scroll-clearance, 9rem);
    }

    .detail-side {
      position: static;
    }

    :global(.detail-tab-list) {
      margin-inline: -0.35rem;
      padding-inline: 0.35rem;
    }

    .resource-list th,
    .resource-list td {
      min-width: 11rem;
    }

    .resource-overflow-action {
      align-items: flex-start;
      display: grid;
      justify-content: stretch;
    }
  }
</style>
