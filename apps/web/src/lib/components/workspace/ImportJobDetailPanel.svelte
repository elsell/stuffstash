<script lang="ts">
  import Activity from '@lucide/svelte/icons/activity';
  import AlertCircle from '@lucide/svelte/icons/alert-circle';
  import CheckCircle2 from '@lucide/svelte/icons/check-circle-2';
  import RefreshCw from '@lucide/svelte/icons/refresh-cw';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import type { ImportJob } from '$lib/domain/inventory';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import * as Tabs from '$lib/components/ui/tabs/index.js';
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

  const COLLAPSED_RESOURCE_LIMIT = 12;

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

  let selectedTab = $state('overview');
  let defaultedTabJobId = $state('');
  let resourcesExpanded = $state(false);
  let issueCount = $derived(detailMessages(job).length);
  let resourcesBounded = $derived(job.resources.length > COLLAPSED_RESOURCE_LIMIT);
  let visibleResources = $derived(resourcesExpanded ? job.resources : job.resources.slice(0, COLLAPSED_RESOURCE_LIMIT));
  let hiddenResourceCount = $derived(Math.max(0, job.resources.length - visibleResources.length));

  $effect(() => {
    if (job.id === defaultedTabJobId) return;
    selectedTab = detailMessages(job).length > 0 ? 'issues' : 'overview';
    resourcesExpanded = false;
    defaultedTabJobId = job.id;
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
  <Card.Content>
    <div class="import-detail-content">
      <div class="detail-hero">
        <div class="job-heading">
          <div>
            <strong>{job.source.name}</strong>
            <span>{sourceDescription(job)}{actorSummary(job) ? ` · ${actorSummary(job)}` : ''}</span>
          </div>
          <Badge variant={job.status === 'failed' || job.status === 'discard_failed' ? 'destructive' : 'secondary'}>
            {statusLabel(job)}
          </Badge>
        </div>
        <div class="detail-topline">
          <div>
            <strong>{phaseLabel(job)}</strong>
            <span>{statusSentence(job)}</span>
          </div>
        </div>
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
              <Tabs.Trigger value="timeline">Timeline</Tabs.Trigger>
            </Tabs.List>

            <Tabs.Content value="overview">
              <section class="detail-panel" aria-label="Import result overview">
                <div class="sample-heading">
                  <h3>Result</h3>
                  <small>{statusSentence(job)}</small>
                </div>
                <ImportCountGrid cells={job.status === 'previewed' ? visiblePreviewCountCells(job) : visibleCountCells(resultCountCells(job))} />
              </section>
            </Tabs.Content>

            <Tabs.Content value="issues">
              <section class="detail-panel" aria-label="Import issues">
                <div class="sample-heading">
                  <h3>Issues</h3>
                  <small>{issueCount === 0 ? 'No issues' : 'Grouped by cause'}</small>
                </div>
                <ImportMessagesList messages={detailMessages(job)} emptyText="No import messages." truncated={job.messages.length === 0 && job.preview.messagesTruncated} />
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
                    {#if job.resources.length > COLLAPSED_RESOURCE_LIMIT}
                      <small>{resourcesExpanded ? `${job.resources.length} records` : `Showing ${visibleResources.length} of ${job.resources.length}`}</small>
                    {/if}
                  </div>
                  <!-- svelte-ignore a11y_no_noninteractive_tabindex (bounded overflow regions need a keyboard focus target) -->
                  <div
                    id="imported-record-summaries"
                    class={resourcesBounded ? 'sample-list resource-list bounded' : 'sample-list resource-list'}
                    role={resourcesBounded ? 'region' : undefined}
                    aria-label={resourcesBounded ? 'Imported record summaries' : undefined}
                    tabindex={resourcesBounded ? 0 : undefined}
                  >
                    {#each visibleResources as resource}
                      <div class="sample-row resource-row">
                        <span>{resourceLabel(resource)}</span>
                        <small>{resourceDiagnosticLabel(resource)} · Imported {new Date(resource.createdAt).toLocaleString()}</small>
                        {#if resourceCanOpen(job, resource)}
                          <a class="resource-link" href={resourceHref(resource)}>Open</a>
                        {/if}
                      </div>
                    {/each}
                  </div>
                  {#if hiddenResourceCount > 0}
                    <div class="resource-overflow-action">
                      <span>{hiddenResourceCount} more imported {hiddenResourceCount === 1 ? 'record' : 'records'} hidden.</span>
                      <Button.Root
                        variant="outline"
                        size="sm"
                        aria-controls="imported-record-summaries"
                        aria-expanded={resourcesExpanded}
                        onclick={() => (resourcesExpanded = true)}
                      >
                        Show more records
                      </Button.Root>
                    </div>
                  {:else if resourcesExpanded && job.resources.length > COLLAPSED_RESOURCE_LIMIT}
                    <div class="resource-overflow-action">
                      <span>All returned record summaries are shown.</span>
                      <Button.Root
                        variant="outline"
                        size="sm"
                        aria-controls="imported-record-summaries"
                        aria-expanded={resourcesExpanded}
                        onclick={() => (resourcesExpanded = false)}
                      >
                        Show fewer
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
          {#if job.cancellationMode}
            <div class="quiet-row">
              <AlertCircle size={16} aria-hidden="true" />
              {cancellationSummary(job)}
            </div>
          {/if}
          <section class="detail-actions" aria-label="Import actions">
            <a class="detail-link" href={auditHistoryHref}>
              <Activity size={16} aria-hidden="true" />
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
    color: hsl(var(--muted-foreground));
    display: block;
    font-size: 0.85rem;
    margin-top: 0.15rem;
    overflow-wrap: anywhere;
  }

  .detail-hero {
    border-bottom: 1px solid hsl(var(--border));
    display: grid;
    gap: 0.75rem;
    padding-bottom: 1rem;
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
    width: 100%;
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
    border-top: 1px solid hsl(var(--border));
    padding-top: 0.75rem;
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

  .resource-list.bounded {
    max-height: min(20rem, 45vh);
    overflow-y: auto;
    padding-right: 0.35rem;
  }

  .resource-overflow-action {
    align-items: center;
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    justify-content: space-between;
  }

  .resource-overflow-action span {
    color: hsl(var(--muted-foreground));
    font-size: 0.82rem;
  }

  .timeline-list {
    gap: 0.35rem;
  }

  .source-options-section {
    border-top: 1px solid hsl(var(--border));
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
    border-left: 2px solid hsl(var(--border));
    color: hsl(var(--foreground));
    font-size: 0.82rem;
    min-width: 0;
    overflow-wrap: anywhere;
    padding: 0.2rem 0 0.2rem 0.55rem;
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
    display: grid;
    gap: 0.55rem;
    grid-template-columns: auto minmax(0, 1fr);
    min-width: 0;
  }

  .timeline-row > span {
    background: hsl(var(--primary));
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
    color: hsl(var(--muted-foreground));
    overflow-wrap: anywhere;
  }

  .sample-row {
    min-width: 0;
  }

  .resource-row {
    align-items: center;
    border-top: 1px solid hsl(var(--border));
    display: grid;
    gap: 0.5rem;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) auto;
    padding-top: 0.55rem;
  }

  .resource-row:first-child {
    border-top: 0;
    padding-top: 0;
  }

  .sample-row span,
  .sample-row small {
    display: block;
    overflow-wrap: anywhere;
  }

  .resource-link,
  .detail-link {
    align-items: center;
    color: hsl(var(--primary));
    display: inline-flex;
    font-size: 0.85rem;
    font-weight: 600;
    gap: 0.4rem;
    text-decoration: none;
  }

  .resource-link {
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

    .detail-side {
      position: static;
    }

    .resource-row {
      gap: 0.35rem;
      grid-template-columns: 1fr;
    }

    .resource-list.bounded {
      max-height: min(14rem, 34vh);
    }

    .resource-overflow-action {
      align-items: flex-start;
      display: grid;
      justify-content: stretch;
    }
  }
</style>
