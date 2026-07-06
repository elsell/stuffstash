<script lang="ts">
  import AlertCircle from '@lucide/svelte/icons/alert-circle';
  import CheckCircle2 from '@lucide/svelte/icons/check-circle-2';
  import Clock3 from '@lucide/svelte/icons/clock-3';
  import Database from '@lucide/svelte/icons/database';
  import Eye from '@lucide/svelte/icons/eye';
  import FileText from '@lucide/svelte/icons/file-text';
  import LoaderCircle from '@lucide/svelte/icons/loader-circle';
  import Play from '@lucide/svelte/icons/play';
  import Plus from '@lucide/svelte/icons/plus';
  import RefreshCw from '@lucide/svelte/icons/refresh-cw';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import XCircle from '@lucide/svelte/icons/x-circle';
  import { untrack } from 'svelte';
  import {
    canCreateImportJob,
    canViewImportJobs,
    type ImportJob,
    type ImportJobCancellationMode,
    type Inventory,
    type ImportSourceRequest
  } from '$lib/domain/inventory';
  import type { InventoryRepository } from '$lib/ports/inventoryRepository';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import * as Checkbox from '$lib/components/ui/checkbox/index.js';
  import * as Input from '$lib/components/ui/input/index.js';
  import * as Label from '$lib/components/ui/label/index.js';
  import { fileToBase64 } from '$lib/application/fileEncoding';
  import { workspaceRouteHref, type ImportSourceRoute } from '$lib/application/workspaceRoute';
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import {
    buildImportSourceRequest,
    importSourceRequestKey,
    readableImportActionError,
    type ImportCSVSelection,
    type ImportSourceChoice
  } from '$lib/application/workspaceImportRequest';
  import {
    actorSummary,
    canRemoveJobFromHistory,
    fileSizeLabel,
    historyCountSummary,
    isTerminal,
    jobNeedsAttention,
    jobTimeLabel,
    phaseLabel,
    previewAssetContext,
    previewLocationContext,
    progressBarLabel,
    progressBarStyle,
    progressKnown,
    progressPercent,
    progressSummary,
    progressTimeline,
    resourceDiagnosticLabel,
    resourceLabel,
    resultCountCells,
    sourceDescription,
    sourceSnapshotDescription,
    statusLabel,
    statusSentence,
    statusVariant,
    terminalJobMayHaveChangedInventory,
    visibleCountCells,
    visiblePreviewCountCells,
    visiblePreviewMessages
  } from './importWorkspacePresentation';
  import ImportFlowStepper from './ImportFlowStepper.svelte';
  import ImportJobConfirmationPanel from './ImportJobConfirmationPanel.svelte';

  type FlowStep = 'history' | 'source' | 'setup' | 'preview' | 'detail';

  const MAX_CSV_BYTES = 10 * 1024 * 1024;

  type Props = {
    tenantId: string;
    inventory: Inventory | null;
    repository: InventoryRepository;
    importSource: ImportSourceRoute;
    onImportSourceChange: (source: ImportSourceRoute) => void;
    onImportJobInventoryChanged: (scope: ImportJobInventoryRefreshScope) => Promise<void>;
  };

  type ImportJobInventoryRefreshScope = {
    tenantId: string;
    inventoryId: string;
  };

  let { tenantId, inventory, repository, importSource, onImportSourceChange, onImportJobInventoryChanged }: Props = $props();

  let step = $state<FlowStep>('history');
  let sourceChoice = $state<ImportSourceChoice>('homebox_live');
  let baseUrl = $state('');
  let username = $state('');
  let password = $state('');
  let includeImages = $state(true);
  let allowPrivateNetwork = $state(false);
  let allowInsecureTLS = $state(false);
  let fileName = $state('');
  let contentBase64 = $state('');
  let csvSelection = $state<ImportCSVSelection | null>(null);
  let jobs = $state<ImportJob[]>([]);
  let previewJob = $state<ImportJob | null>(null);
  let previewedRequestKey = $state('');
  let selectedJob = $state<ImportJob | null>(null);
  let cancelIntent = $state<{ job: ImportJob } | null>(null);
  let removeIntent = $state<{ job: ImportJob } | null>(null);
  let loading = $state(false);
  let busy = $state(false);
  let error = $state('');
  let notice = $state('');
  let fileSelectionToken = 0;
  let jobLoadSequence = 0;
  let visibleJobLoadSequence = 0;
  let actionSequence = 0;

  let activeJobs = $derived(jobs.filter((job) => ['running', 'cancel_requested'].includes(job.status)));
  let draftJobs = $derived(jobs.filter((job) => job.status === 'previewed'));
  let currentWorkJobs = $derived([...activeJobs, ...draftJobs]);
  let terminalJobs = $derived(jobs.filter((job) => !['running', 'cancel_requested', 'previewed'].includes(job.status)));
  let completedJobs = $derived(jobs.filter((job) => job.status === 'succeeded'));
  let attentionJobs = $derived(jobs.filter(jobNeedsAttention));
  let canViewImports = $derived(canViewImportJobs(inventory));
  let canCreateImports = $derived(canCreateImportJob(inventory));
  let currentRequestKey = $derived(requestKey());
  let previewReady = $derived(previewJob?.status === 'previewed' && previewJob.counts.errors === 0 && previewedRequestKey === currentRequestKey);
  let previewStale = $derived(Boolean(previewJob && previewedRequestKey && previewedRequestKey !== currentRequestKey));
  let canConfirmSource = $derived(
    Boolean(canCreateImports && (sourceChoice === 'homebox_csv' ? contentBase64 : baseUrl.trim() && username.trim() && password))
  );

  $effect(() => {
    tenantId;
    inventory?.id;
    untrack(() => {
      void loadJobs();
    });
  });

  $effect(() => {
    if (!canCreateImports && (step === 'source' || step === 'setup' || step === 'preview')) {
      step = 'history';
      previewJob = null;
      previewedRequestKey = '';
      onImportSourceChange(null);
      return;
    }
    if (importSource === 'homebox') {
      sourceChoice = 'homebox_live';
      if (canCreateImports && (step === 'history' || step === 'source')) {
        step = 'setup';
      }
    } else if (importSource === 'homebox-csv') {
      sourceChoice = 'homebox_csv';
      if (canCreateImports && (step === 'history' || step === 'source')) {
        step = 'setup';
      }
    } else if (step === 'setup' || step === 'preview') {
      step = 'history';
      previewJob = null;
      previewedRequestKey = '';
      error = '';
      notice = '';
    }
  });

  $effect(() => {
    if (!inventory || !canViewImports || activeJobs.length === 0) return;
    const interval = window.setInterval(() => {
      void loadJobs({ quiet: true });
    }, 2500);
    return () => window.clearInterval(interval);
  });

  async function loadJobs(options: { quiet?: boolean } = {}): Promise<void> {
    const scope = currentImportScope();
    if (!scope) {
      jobs = [];
      previewJob = null;
      selectedJob = null;
      return;
    }
    const sequence = (jobLoadSequence += 1);
    if (!options.quiet) {
      visibleJobLoadSequence = sequence;
    }
    const previousJobs = jobs;
    if (!options.quiet) {
      loading = true;
    }
    if (!options.quiet) {
      error = '';
    }
    if (!options.quiet) {
      notice = '';
    }
    try {
      const nextJobs = await repository.listImportJobs(scope.tenantId, scope.inventoryId);
      if (!isCurrentJobLoad(sequence, scope)) {
        return;
      }
      jobs = nextJobs;
      reconcileSelectedJob(nextJobs);
      await refreshInventoryWhenImportJobsFinish(scope, previousJobs, nextJobs);
    } catch (loadError) {
      if (!options.quiet && isCurrentJobLoad(sequence, scope)) {
        error = errorMessage(loadError, 'Import history could not be loaded.');
      }
    } finally {
      if (!options.quiet && visibleJobLoadSequence === sequence && isCurrentScope(scope)) {
        loading = false;
      }
    }
  }

  function currentImportScope(): ImportJobInventoryRefreshScope | null {
    return inventory && canViewImports ? { tenantId, inventoryId: inventory.id } : null;
  }

  function isCurrentScope(scope: ImportJobInventoryRefreshScope): boolean {
    return tenantId === scope.tenantId && inventory?.id === scope.inventoryId;
  }

  function isCurrentJobLoad(sequence: number, scope: ImportJobInventoryRefreshScope): boolean {
    return sequence === jobLoadSequence && isCurrentScope(scope);
  }

  function startScopedAction(): { sequence: number; scope: ImportJobInventoryRefreshScope } | null {
    const scope = currentImportScope();
    if (!scope) return null;
    return { sequence: (actionSequence += 1), scope };
  }

  function isCurrentAction(action: { sequence: number; scope: ImportJobInventoryRefreshScope }): boolean {
    return action.sequence === actionSequence && isCurrentScope(action.scope);
  }

  function beginImport(): void {
    if (!canCreateImports) return;
    error = '';
    notice = '';
    previewJob = null;
    previewedRequestKey = '';
    selectedJob = null;
    cancelIntent = null;
    removeIntent = null;
    clearSourceSecretAndFileState();
    step = 'source';
  }

  function chooseSource(choice: ImportSourceChoice): void {
    if (!canCreateImports) return;
    sourceChoice = choice;
    error = '';
    notice = '';
    previewJob = null;
    previewedRequestKey = '';
    step = 'setup';
  }

  function chooseSourceFromLink(event: MouseEvent, choice: ImportSourceChoice): void {
    if (!shouldHandleWorkspaceLinkClick(event)) return;
    event.preventDefault();
    onImportSourceChange(importSourceRouteForChoice(choice));
    chooseSource(choice);
  }

  function returnToHistory(): void {
    error = '';
    notice = '';
    cancelIntent = null;
    selectedJob = null;
    onImportSourceChange(null);
    step = 'history';
  }

  function openJob(job: ImportJob): void {
    selectedJob = job;
    cancelIntent = null;
    removeIntent = null;
    error = '';
    notice = '';
    step = 'detail';
  }

  function resumePreviewedJob(job: ImportJob): void {
    if (!canCreateImports || job.status !== 'previewed') return;
    selectedJob = null;
    cancelIntent = null;
    removeIntent = null;
    previewJob = null;
    previewedRequestKey = '';
    clearSourceSecretAndFileState();
    sourceChoice = job.source.type === 'legacy_homebox_csv' ? 'homebox_csv' : 'homebox_live';
    baseUrl = job.source.baseUrl ?? '';
    includeImages = job.source.imageImport !== 'disabled';
    onImportSourceChange(importSourceRouteForChoice(sourceChoice));
    error = '';
    notice = 'Confirm the source again to continue this import. Credentials and CSV contents are not kept in the browser.';
    step = 'setup';
  }

  function clearSourceSecretAndFileState(): void {
    username = '';
    password = '';
    fileName = '';
    contentBase64 = '';
    csvSelection = null;
    allowPrivateNetwork = false;
    allowInsecureTLS = false;
    fileSelectionToken += 1;
  }

  function requestSelectedJobCancellation(): void {
    if (!selectedJob || !canCreateImports) return;
    cancelIntent = { job: selectedJob };
  }

  function removeSelectedJobFromHistory(): void {
    if (!selectedJob || !canCreateImports || !canRemoveJobFromHistory(selectedJob)) return;
    removeIntent = { job: selectedJob };
  }

  async function confirmSource(): Promise<void> {
    const action = startScopedAction();
    if (!action || !canConfirmSource || !canCreateImports) return;
    busy = true;
    error = '';
    notice = '';
    try {
      const job = await repository.previewImportJob(action.scope.tenantId, action.scope.inventoryId, request());
      if (!isCurrentAction(action)) return;
      previewJob = job;
      previewedRequestKey = requestKey();
      jobs = [job, ...jobs.filter((candidate) => candidate.id !== job.id)];
      step = 'preview';
    } catch (previewError) {
      if (!isCurrentAction(action)) return;
      error = errorMessage(previewError, sourceChoice === 'homebox_live' ? 'Homebox connection could not be confirmed.' : 'CSV preview could not be prepared.');
    } finally {
      if (isCurrentAction(action)) {
        busy = false;
      }
    }
  }

  async function start(): Promise<void> {
    const action = startScopedAction();
    if (!action || !previewJob || !previewReady || !canCreateImports) return;
    const jobID = previewJob.id;
    busy = true;
    error = '';
    notice = '';
    try {
      const job = await repository.startImportJob(action.scope.tenantId, action.scope.inventoryId, jobID, request());
      if (!isCurrentAction(action)) return;
      jobs = [job, ...jobs.filter((candidate) => candidate.id !== job.id)];
      previewJob = null;
      previewedRequestKey = '';
      onImportSourceChange(null);
      step = 'history';
      notice = 'Import started. Progress will update in import history.';
    } catch (startError) {
      if (!isCurrentAction(action)) return;
      error = errorMessage(startError, 'Import could not be started. Preview again if the source changed.');
    } finally {
      if (isCurrentAction(action)) {
        busy = false;
      }
    }
  }

  async function cancel(job: ImportJob, mode: ImportJobCancellationMode): Promise<void> {
    const action = startScopedAction();
    if (!action || !canCreateImports) return;
    busy = true;
    error = '';
    notice = '';
    try {
      const next = await repository.cancelImportJob(action.scope.tenantId, action.scope.inventoryId, job.id, mode);
      if (!isCurrentAction(action)) return;
      jobs = jobs.map((candidate) => (candidate.id === next.id ? next : candidate));
      if (selectedJob?.id === next.id) {
        selectedJob = next;
      }
      cancelIntent = null;
    } catch (cancelError) {
      if (!isCurrentAction(action)) return;
      error = errorMessage(cancelError, 'Cancellation could not be requested.');
    } finally {
      if (isCurrentAction(action)) {
        busy = false;
      }
    }
  }

  async function removeFromHistory(job: ImportJob): Promise<void> {
    const action = startScopedAction();
    if (!action || !canCreateImports || !canRemoveJobFromHistory(job)) return;
    busy = true;
    error = '';
    notice = '';
    try {
      await repository.removeImportJobFromHistory(action.scope.tenantId, action.scope.inventoryId, job.id);
      if (!isCurrentAction(action)) return;
      jobs = jobs.filter((candidate) => candidate.id !== job.id);
      removeIntent = null;
      if (selectedJob?.id === job.id) {
        selectedJob = null;
        cancelIntent = null;
        step = 'history';
      }
    } catch (removeError) {
      if (!isCurrentAction(action)) return;
      error = errorMessage(removeError, 'Only completed import jobs can be removed from history.');
    } finally {
      if (isCurrentAction(action)) {
        busy = false;
      }
    }
  }

  async function handleFile(event: Event): Promise<void> {
    const file = (event.currentTarget as HTMLInputElement).files?.[0];
    const token = (fileSelectionToken += 1);
    if (!file) return;
    error = '';
    previewJob = null;
    previewedRequestKey = '';
    if (file.size > MAX_CSV_BYTES) {
      fileName = '';
      contentBase64 = '';
      csvSelection = null;
      error = 'CSV is too large. Choose a Homebox CSV export under 10 MiB.';
      return;
    }
    fileName = file.name;
    const selection = { name: file.name, size: file.size, lastModified: file.lastModified };
    csvSelection = selection;
    try {
      const encoded = await fileToBase64(file);
      if (token !== fileSelectionToken || !sameCSVSelection(csvSelection, selection)) return;
      contentBase64 = encoded;
    } catch {
      if (token !== fileSelectionToken) return;
      fileName = '';
      contentBase64 = '';
      csvSelection = null;
      error = 'CSV could not be read. Choose the Homebox export again.';
    }
  }

  function request(): ImportSourceRequest {
    return buildImportSourceRequest(requestDraft());
  }

  function requestKey(): string {
    return importSourceRequestKey(requestDraft());
  }

  function requestDraft() {
    return {
      sourceChoice,
      baseUrl,
      username,
      password,
      includeImages,
      allowPrivateNetwork,
      allowInsecureTLS,
      fileName,
      contentBase64,
      csvSelection
    };
  }

  function previewReadinessTitle(job: ImportJob): string {
    if (previewStale) return 'Preview needs to be refreshed';
    if (job.counts.errors > 0) return 'Fix blocking issues before importing';
    return 'Ready to start';
  }

  function previewReadinessDescription(job: ImportJob): string {
    if (previewStale) return 'The source settings changed after this preview. Confirm the source again before starting.';
    if (job.counts.errors > 0) return 'Nothing has been saved. Review the blocking messages below and preview again after fixing the source.';
    if (job.counts.warnings > 0) return 'Nothing has been saved. Warnings are shown below so you can decide whether to continue.';
    return 'Nothing has been saved. Start the import when this plan looks right.';
  }

  function previewReadinessBadge(job: ImportJob): string {
    if (previewStale) return 'Re-preview required';
    if (job.counts.errors > 0) return `${job.counts.errors} blocking`;
    if (job.counts.warnings > 0) return `${job.counts.warnings} warnings`;
    return 'Ready';
  }

  function errorMessage(value: unknown, fallback: string): string {
    return readableImportActionError(value, fallback);
  }

  function sameCSVSelection(left: ImportCSVSelection | null, right: ImportCSVSelection): boolean {
    return Boolean(left && left.name === right.name && left.size === right.size && left.lastModified === right.lastModified);
  }

  function reconcileSelectedJob(nextJobs: ImportJob[]): void {
    if (!selectedJob) return;
    selectedJob = nextJobs.find((job) => job.id === selectedJob?.id) ?? selectedJob;
  }

  async function refreshInventoryWhenImportJobsFinish(
    scope: ImportJobInventoryRefreshScope,
    previousJobs: ImportJob[],
    nextJobs: ImportJob[]
  ): Promise<void> {
    const previousActiveJobIds = new Set(previousJobs.filter((job) => ['running', 'cancel_requested'].includes(job.status)).map((job) => job.id));
    const finishedJob = nextJobs.find((job) => previousActiveJobIds.has(job.id) && terminalJobMayHaveChangedInventory(job));
    if (!finishedJob) return;
    try {
      await onImportJobInventoryChanged(scope);
      if (!isCurrentScope(scope)) return;
      notice = 'Import finished. Workspace data has been refreshed.';
    } catch (refreshError) {
      if (!isCurrentScope(scope)) return;
      notice = errorMessage(refreshError, 'Import finished, but workspace data could not be refreshed.');
    }
  }

  function historySummaryDescription(): string {
    if (activeJobs.length > 0) return `${activeJobs.length} running now. You can leave this page and return later.`;
    if (draftJobs.length > 0) return `${draftJobs.length} preview waiting for confirmation.`;
    if (attentionJobs.length > 0) return `${attentionJobs.length} import ${attentionJobs.length === 1 ? 'needs' : 'need'} attention.`;
    if (completedJobs.length > 0) return `${completedJobs.length} completed import ${completedJobs.length === 1 ? 'run' : 'runs'} in this inventory.`;
    return 'No import runs yet.';
  }

  function resourceCanOpen(job: ImportJob, resource: ImportJob['resources'][number]): boolean {
    if (job.status === 'cancelled_discarded') return false;
    return Boolean(inventory && (resource.resourceType === 'asset' || (resource.resourceType === 'attachment' && resource.resourceOwnerId)));
  }

  function resourceHref(resource: ImportJob['resources'][number]): string {
    if (!inventory) return '#';
    if (resource.resourceType === 'asset') {
      return workspaceRouteHref({ mode: 'asset', assetId: resource.resourceId }, tenantId, inventory.id);
    }
    if (resource.resourceType === 'attachment' && resource.resourceOwnerId) {
      return workspaceRouteHref({ mode: 'asset', assetId: resource.resourceOwnerId }, tenantId, inventory.id);
    }
    return '#';
  }

  function auditHistoryHref(): string {
    return workspaceRouteHref({ mode: 'settings', settingsSection: 'activity', auditScope: 'inventory' }, tenantId, inventory?.id ?? null);
  }

  function importSourceRouteForChoice(choice: ImportSourceChoice): Exclude<ImportSourceRoute, null> {
    return choice === 'homebox_csv' ? 'homebox-csv' : 'homebox';
  }

  function importSourceHref(choice: ImportSourceChoice): string {
    return workspaceRouteHref({ mode: 'import', importSource: importSourceRouteForChoice(choice) }, tenantId, inventory?.id ?? null);
  }

</script>

<section class="import-workspace">
  <div class="import-toolbar">
    <div>
      <h1>{step === 'history' ? 'Imports' : step === 'detail' ? 'Import details' : 'New import'}</h1>
      <p>
        {#if step === 'history'}
          {`Review runs for ${inventory?.name ?? 'this inventory'} and resume current work.`}
        {:else if step === 'detail'}
          {`Import run for ${inventory?.name ?? 'this inventory'}.`}
        {:else}
          Confirm the source, preview the plan, then run it in the background.
        {/if}
      </p>
    </div>
    {#if step === 'history'}
      <Button.Root variant="outline" size="sm" onclick={() => { void loadJobs(); }} disabled={loading || !canViewImports}>
        <RefreshCw size={16} aria-hidden="true" />
        Refresh
      </Button.Root>
    {:else}
      <Button.Root variant="outline" size="sm" onclick={returnToHistory} disabled={busy}>Back to history</Button.Root>
    {/if}
  </div>

  {#if error}
    <div class="import-alert" role="alert">{error}</div>
  {/if}
  {#if notice}
    <div class="import-notice" role="status">{notice}</div>
  {/if}

  {#if !canViewImports}
    <Card.Root>
      <Card.Content class="empty-state">
        <AlertCircle size={28} aria-hidden="true" />
        <div>
          <h2>Import access needed</h2>
          <p>You can view this inventory, but importing records requires import job access.</p>
        </div>
      </Card.Content>
    </Card.Root>
  {:else if cancelIntent || removeIntent}
    <ImportJobConfirmationPanel
      cancelJob={cancelIntent?.job ?? null}
      removeJob={removeIntent?.job ?? null}
      {busy}
      onCancelJob={(job, mode) => { void cancel(job, mode); }}
      onDismissCancel={() => (cancelIntent = null)}
      onRemoveJob={(job) => { void removeFromHistory(job); }}
      onDismissRemove={() => (removeIntent = null)}
    />
  {/if}

  {#if canViewImports && !cancelIntent && !removeIntent}
    {#if step === 'history'}
    <div class="history-header">
      <div>
        <h2>Import history</h2>
        <p>{historySummaryDescription()}</p>
      </div>
      <Button.Root onclick={beginImport} disabled={!canCreateImports} variant={currentWorkJobs.length > 0 ? 'outline' : 'default'}>
        <Plus size={16} aria-hidden="true" />
        New import
      </Button.Root>
    </div>

    {#if jobs.length > 0}
      <div class="history-summary-grid" aria-label="Import history summary">
        <div class={activeJobs.length > 0 ? 'summary-active' : ''}>
          <span>Running</span>
          <strong>{activeJobs.length}</strong>
        </div>
        <div class={draftJobs.length > 0 ? 'summary-active' : ''}>
          <span>Ready to review</span>
          <strong>{draftJobs.length}</strong>
        </div>
        <div class={completedJobs.length > 0 ? 'summary-active' : ''}>
          <span>Completed</span>
          <strong>{completedJobs.length}</strong>
        </div>
        <div class={attentionJobs.length > 0 ? 'summary-warning' : ''}>
          <span>Needs attention</span>
          <strong>{attentionJobs.length}</strong>
        </div>
      </div>
    {/if}

    {#if jobs.length === 0}
      <Card.Root>
        <Card.Content class="empty-state">
          <Database size={28} aria-hidden="true" />
          <div>
            <h2>No import runs yet</h2>
            <p>Start with Homebox, preview what Stuff Stash will create, then run it in the background.</p>
          </div>
          <Button.Root onclick={beginImport} disabled={!canCreateImports}>
            <Plus size={16} aria-hidden="true" />
            New import
          </Button.Root>
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
              <Card.Content class="job-card">
                <div class="job-main">
                  <LoaderCircle class="spin" size={18} aria-hidden="true" />
                  <div>
                    <strong>{job.source.name}</strong>
                    <span>{job.status === 'cancel_requested' ? statusSentence(job) : phaseLabel(job)}{actorSummary(job) ? ` · ${actorSummary(job)}` : ''}</span>
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
                  <Button.Root variant="ghost" size="sm" onclick={() => openJob(job)} disabled={busy}>
                    <Eye size={16} aria-hidden="true" />
                    Details
                  </Button.Root>
                  <Button.Root variant="outline" size="sm" onclick={() => (cancelIntent = { job })} disabled={busy || !canCreateImports}>
                    Cancel
                  </Button.Root>
                </div>
              </Card.Content>
            </Card.Root>
          {/each}
          {#each draftJobs as job}
            <div class="history-row">
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
              <Button.Root variant="outline" size="sm" onclick={() => resumePreviewedJob(job)}>Continue</Button.Root>
              <Button.Root variant="ghost" size="sm" onclick={() => openJob(job)}>Details</Button.Root>
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
            <div>
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
              <Button.Root variant="ghost" size="sm" onclick={() => openJob(job)}>
                <Eye size={16} aria-hidden="true" />
                Details
              </Button.Root>
            {#if canRemoveJobFromHistory(job)}
              <Button.Root variant="ghost" size="icon" onclick={() => (removeIntent = { job })} aria-label="Remove import job from history">
                <Trash2 size={16} aria-hidden="true" />
              </Button.Root>
            {/if}
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
  {:else if step === 'source'}
    <Card.Root>
      <Card.Header>
        <ImportFlowStepper current="source" />
        <Card.Title>Choose import method</Card.Title>
        <Card.Description>Pick the path that matches the data you have right now.</Card.Description>
      </Card.Header>
      <Card.Content class="step-content">
        <div class="source-choice-grid" role="group" aria-label="Homebox import method">
          <Button.Root
            variant="outline"
            class="source-card"
            href={importSourceHref('homebox_live')}
            onclick={(event) => chooseSourceFromLink(event, 'homebox_live')}
          >
            <Database size={24} aria-hidden="true" />
            <span>
              <strong>Connect to Homebox</strong>
              <small>Use your Homebox URL and credentials. Best when the instance is reachable and you want photos from the live API.</small>
              <em>Can include photos · checks the source before running</em>
            </span>
          </Button.Root>
          <Button.Root
            variant="outline"
            class="source-card"
            href={importSourceHref('homebox_csv')}
            onclick={(event) => chooseSourceFromLink(event, 'homebox_csv')}
          >
            <FileText size={24} aria-hidden="true" />
            <span>
              <strong>Upload Homebox CSV</strong>
              <small>Use an exported CSV file. Best for offline imports, migrations from an older instance, or when the Homebox API is not reachable.</small>
              <em>No photos in CSV · works without a live server</em>
            </span>
          </Button.Root>
        </div>
        <div class="action-row">
          <Button.Root variant="outline" onclick={returnToHistory}>Cancel</Button.Root>
        </div>
      </Card.Content>
    </Card.Root>
  {:else if step === 'setup'}
    <Card.Root>
      <Card.Header>
        <ImportFlowStepper current="connect" />
        <Card.Title>{sourceChoice === 'homebox_live' ? 'Connect to Homebox' : 'Upload Homebox CSV'}</Card.Title>
        <Card.Description>Stuff Stash will verify the source and build a preview.</Card.Description>
      </Card.Header>
      <Card.Content class="step-content">
        {#if sourceChoice === 'homebox_live'}
          <div class="field-stack">
            <Label.Root for="homebox-url">Homebox URL</Label.Root>
            <Input.Root
              id="homebox-url"
              bind:value={baseUrl}
              placeholder="homebox.example.com or http://homebox.local:3100"
              autocomplete="url"
              autocapitalize="none"
              autocorrect="off"
              inputmode="url"
              spellcheck={false}
            />
            <small class="field-note">Explicit http:// and https:// URLs are preserved. Schemeless hosts try https:// first.</small>
          </div>
          <div class="field-grid">
            <div class="field-stack">
              <Label.Root for="homebox-user">Email</Label.Root>
              <Input.Root
                id="homebox-user"
                type="email"
                bind:value={username}
                autocomplete="username"
                autocapitalize="none"
                autocorrect="off"
                inputmode="email"
                spellcheck={false}
              />
            </div>
            <div class="field-stack">
              <Label.Root for="homebox-password">Password</Label.Root>
              <Input.Root
                id="homebox-password"
                type="password"
                bind:value={password}
                autocomplete="current-password"
                autocapitalize="none"
                autocorrect="off"
                spellcheck={false}
              />
            </div>
          </div>
          <Label.Root class="check-row">
            <Checkbox.Root bind:checked={includeImages} />
            <span>Import photos when Homebox provides them</span>
          </Label.Root>
          <details class="advanced-options">
            <summary>Connection options</summary>
            <div class="advanced-option-list">
              <Label.Root class="check-row">
                <Checkbox.Root bind:checked={allowPrivateNetwork} />
                <span>Allow private-network Homebox URL</span>
              </Label.Root>
              <Label.Root class="check-row">
                <Checkbox.Root bind:checked={allowInsecureTLS} />
                <span>Allow self-signed TLS certificate</span>
              </Label.Root>
            </div>
          </details>
        {:else}
          <div class="field-stack">
            <Label.Root for="homebox-csv">Homebox CSV export</Label.Root>
            <Input.Root id="homebox-csv" type="file" accept=".csv,text/csv" onchange={(event) => { void handleFile(event); }} />
              {#if csvSelection}
                <small class="field-note">{csvSelection.name} · {Math.max(1, Math.round(csvSelection.size / 1024))} KB · photos are not included in CSV exports</small>
              {:else}
                <small class="field-note">CSV files must be 10 MiB or smaller. Homebox CSV exports do not include photos.</small>
              {/if}
          </div>
        {/if}

        <div class="action-row">
          <Button.Root onclick={() => { void confirmSource(); }} disabled={!canConfirmSource || busy || !inventory}>
            <CheckCircle2 size={16} aria-hidden="true" />
            {sourceChoice === 'homebox_live' ? 'Confirm connection' : 'Prepare preview'}
          </Button.Root>
          <Button.Root variant="outline" onclick={() => (step = 'source')} disabled={busy}>Back</Button.Root>
        </div>
      </Card.Content>
    </Card.Root>
  {:else if step === 'preview'}
    <Card.Root>
      <Card.Header>
        <ImportFlowStepper current="preview" />
        <Card.Title>Preview import</Card.Title>
        <Card.Description>Review the plan before starting the background import.</Card.Description>
      </Card.Header>
      <Card.Content class="step-content">
        {#if previewJob}
          <div class={`readiness-panel ${previewStale || previewJob.counts.errors > 0 ? 'needs-attention' : ''}`}>
            <div>
              <strong>{previewReadinessTitle(previewJob)}</strong>
              <span>{previewReadinessDescription(previewJob)}</span>
            </div>
            <Badge variant={previewStale || previewJob.counts.errors > 0 ? 'destructive' : 'secondary'}>
              {previewReadinessBadge(previewJob)}
            </Badge>
          </div>
          <div class="source-summary">
            <span>{sourceDescription(previewJob)}</span>
            <small>{sourceSnapshotDescription(previewJob)}</small>
          </div>
          <div class="summary-grid">
              {#each visiblePreviewCountCells(previewJob) as count}
                <div class:muted-count={count.muted}><strong>{count.value}</strong><span>{` ${count.label}`}</span></div>
              {/each}
          </div>
          <div class="preview-samples">
            <section>
              <div class="sample-heading">
                <h3>Fields</h3>
                {#if previewJob.preview.fieldsTruncated}<small>Showing a sample</small>{/if}
              </div>
              <div class="sample-list">
                {#each previewJob.preview.fields as field}
                  <div class="sample-row">
                    <span>{field.displayName || field.key}</span>
                    <small>{field.key} · {field.type}</small>
                  </div>
                {/each}
                {#if previewJob.preview.fields.length === 0}
                  <div class="quiet-row"><CheckCircle2 size={16} aria-hidden="true" /> No custom fields planned.</div>
                {/if}
              </div>
            </section>
            <section>
              <div class="sample-heading">
                <h3>Locations</h3>
                {#if previewJob.preview.locationsTruncated}<small>Showing a sample</small>{/if}
              </div>
              <div class="sample-list">
                {#each previewJob.preview.locations as item}
                  <div class="sample-row">
                    <span>{item.title}</span>
                    <small>{previewLocationContext(item)}</small>
                  </div>
                {/each}
                {#if previewJob.preview.locations.length === 0}
                  <div class="quiet-row"><CheckCircle2 size={16} aria-hidden="true" /> No locations planned.</div>
                {/if}
              </div>
            </section>
            <section>
              <div class="sample-heading">
                <h3>Assets</h3>
                {#if previewJob.preview.assetsTruncated}<small>Showing a sample</small>{/if}
              </div>
              <div class="sample-list">
                {#each previewJob.preview.assets as item}
                  <div class="sample-row">
                    <span>{item.title}</span>
                    <small>{previewAssetContext(item)}</small>
                  </div>
                {/each}
                {#if previewJob.preview.assets.length === 0}
                  <div class="quiet-row"><CheckCircle2 size={16} aria-hidden="true" /> No asset records planned.</div>
                {/if}
              </div>
            </section>
            <section>
              <div class="sample-heading">
                <h3>Photos/files</h3>
                {#if previewJob.preview.attachmentsTruncated}<small>Showing a sample</small>{/if}
              </div>
              <div class="sample-list">
                {#each previewJob.preview.attachments as attachment}
                  <div class="sample-row">
                    <span>{attachment.fileName || 'Unnamed attachment'}</span>
                    <small>{attachment.contentType || 'unknown type'} · {fileSizeLabel(attachment.sizeBytes)}{attachment.primary ? ' · primary' : ''}</small>
                  </div>
                {/each}
                {#if previewJob.preview.attachments.length === 0}
                  <div class="quiet-row"><CheckCircle2 size={16} aria-hidden="true" /> No photos or files planned.</div>
                {/if}
              </div>
            </section>
          </div>
          <div class="message-list">
            {#each visiblePreviewMessages(previewJob) as message}
              <div class="message-row">
                <Badge variant={message.severity === 'error' ? 'destructive' : 'secondary'}>{message.severity}</Badge>
                <span>{message.summary}{message.detail ? ` · ${message.detail}` : ''}</span>
              </div>
            {/each}
            {#if visiblePreviewMessages(previewJob).length === 0}
              <div class="quiet-row"><CheckCircle2 size={16} aria-hidden="true" /> No blocking issues found.</div>
            {/if}
            {#if previewJob.preview.messagesTruncated}
              <div class="quiet-row"><AlertCircle size={16} aria-hidden="true" /> Showing a sample of import messages.</div>
            {/if}
          </div>
        {/if}

        <div class="action-row">
          <Button.Root onclick={() => { void start(); }} disabled={!previewReady || busy}>
            <Play size={16} aria-hidden="true" />
            Start background import
          </Button.Root>
          <Button.Root variant="outline" onclick={() => (step = 'setup')} disabled={busy}>Back</Button.Root>
        </div>
      </Card.Content>
    </Card.Root>
  {:else if step === 'detail' && selectedJob}
    <Card.Root>
      <Card.Header>
        <Card.Title>{selectedJob.source.name}</Card.Title>
        <Card.Description>{statusLabel(selectedJob)} · {sourceDescription(selectedJob)}{actorSummary(selectedJob) ? ` · ${actorSummary(selectedJob)}` : ''}</Card.Description>
      </Card.Header>
      <Card.Content class="step-content">
          <div class="detail-topline">
            <div>
              <strong>{phaseLabel(selectedJob)}</strong>
              <span>{statusSentence(selectedJob)}</span>
            </div>
            <Badge variant={selectedJob.status === 'failed' || selectedJob.status === 'discard_failed' ? 'destructive' : 'secondary'}>
              {statusLabel(selectedJob)}
            </Badge>
          </div>
          {#if progressKnown(selectedJob) || !isTerminal(selectedJob)}
            <div
              class="progress-track large"
              class:indeterminate={!progressKnown(selectedJob)}
              role="progressbar"
              aria-label={progressBarLabel(selectedJob)}
              aria-valuemin={progressKnown(selectedJob) ? 0 : undefined}
              aria-valuemax={progressKnown(selectedJob) ? 100 : undefined}
              aria-valuenow={progressKnown(selectedJob) ? progressPercent(selectedJob) : undefined}
            >
              <span style={progressBarStyle(selectedJob)}></span>
            </div>
          {/if}
          <section class="timeline-section">
            <div class="sample-heading">
              <h3>Progress timeline</h3>
              <small>{progressTimeline(selectedJob).length} phases</small>
            </div>
            <div class="timeline-list">
              {#each progressTimeline(selectedJob) as progress}
                <div class="timeline-row">
                  <span>{phaseLabel({ ...selectedJob, progress })}</span>
                  <small>
                    {progress.total > 0 ? `${Math.min(progress.done, progress.total)} / ${progress.total}` : ''}
                    {progress.total > 0 && progress.updatedAt ? ' · ' : ''}
                    {progress.updatedAt ? new Date(progress.updatedAt).toLocaleString() : ''}
                  </small>
                </div>
              {/each}
            </div>
          </section>
          <div class="summary-grid">
            {#if selectedJob.status === 'previewed'}
              {#each visiblePreviewCountCells(selectedJob) as count}
                <div class:muted-count={count.muted}><strong>{count.value}</strong><span>{` ${count.label}`}</span></div>
              {/each}
            {:else}
              {#each visibleCountCells(resultCountCells(selectedJob)) as count}
                <div class:muted-count={count.muted}><strong>{count.value}</strong><span>{` ${count.label}`}</span></div>
              {/each}
            {/if}
          </div>
          {#if selectedJob.status === 'previewed'}
            <div class="preview-samples">
              <section>
                <div class="sample-heading">
                  <h3>Fields</h3>
                  {#if selectedJob.preview.fieldsTruncated}<small>Showing a sample</small>{/if}
                </div>
                <div class="sample-list">
                  {#each selectedJob.preview.fields as field}
                    <div class="sample-row">
                      <span>{field.displayName || field.key}</span>
                      <small>{field.key} · {field.type}</small>
                    </div>
                  {/each}
                  {#if selectedJob.preview.fields.length === 0}
                    <div class="quiet-row"><CheckCircle2 size={16} aria-hidden="true" /> No custom fields planned.</div>
                  {/if}
                </div>
              </section>
              <section>
                <div class="sample-heading">
                  <h3>Locations</h3>
                  {#if selectedJob.preview.locationsTruncated}<small>Showing a sample</small>{/if}
                </div>
                <div class="sample-list">
                  {#each selectedJob.preview.locations as item}
                    <div class="sample-row">
                      <span>{item.title}</span>
                      <small>{previewLocationContext(item)}</small>
                    </div>
                  {/each}
                  {#if selectedJob.preview.locations.length === 0}
                    <div class="quiet-row"><CheckCircle2 size={16} aria-hidden="true" /> No locations planned.</div>
                  {/if}
                </div>
              </section>
              <section>
                <div class="sample-heading">
                  <h3>Assets</h3>
                  {#if selectedJob.preview.assetsTruncated}<small>Showing a sample</small>{/if}
                </div>
                <div class="sample-list">
                  {#each selectedJob.preview.assets as item}
                    <div class="sample-row">
                      <span>{item.title}</span>
                      <small>{previewAssetContext(item)}</small>
                    </div>
                  {/each}
                  {#if selectedJob.preview.assets.length === 0}
                    <div class="quiet-row"><CheckCircle2 size={16} aria-hidden="true" /> No asset records planned.</div>
                  {/if}
                </div>
              </section>
              <section>
                <div class="sample-heading">
                  <h3>Photos/files</h3>
                  {#if selectedJob.preview.attachmentsTruncated}<small>Showing a sample</small>{/if}
                </div>
                <div class="sample-list">
                  {#each selectedJob.preview.attachments as attachment}
                    <div class="sample-row">
                      <span>{attachment.fileName || 'Unnamed attachment'}</span>
                      <small>{attachment.contentType || 'unknown type'} · {fileSizeLabel(attachment.sizeBytes)}{attachment.primary ? ' · primary' : ''}</small>
                    </div>
                  {/each}
                  {#if selectedJob.preview.attachments.length === 0}
                    <div class="quiet-row"><CheckCircle2 size={16} aria-hidden="true" /> No photos or files planned.</div>
                  {/if}
                </div>
              </section>
            </div>
          {/if}
          {#if selectedJob.cancellationMode}
            <div class="quiet-row">
              <AlertCircle size={16} aria-hidden="true" />
              {selectedJob.cancellationMode === 'discard_partial_progress'
                ? 'Partial progress discard was requested. Audit history remains.'
                : 'Partial progress was kept.'}
            </div>
          {/if}
          {#if selectedJob.status === 'cancelled_discarded' && selectedJob.resources.length > 0}
            <div class="quiet-row">
              <CheckCircle2 size={16} aria-hidden="true" />
              Records created by this job were discarded. Audit history remains.
            </div>
          {:else if selectedJob.resources.length > 0}
            <section class="resource-section">
              <div class="sample-heading">
                <h3>Imported records</h3>
                {#if selectedJob.resources.length >= 50}<small>Showing a sample</small>{/if}
              </div>
              <div class="sample-list">
                {#each selectedJob.resources as resource}
                  <div class="sample-row resource-row">
                    <span>{resourceLabel(resource)}</span>
                    <small>{resourceDiagnosticLabel(resource)} · Imported {new Date(resource.createdAt).toLocaleString()}</small>
                    {#if resourceCanOpen(selectedJob, resource)}
                      <a class="resource-link" href={resourceHref(resource)}>Open</a>
                    {/if}
                  </div>
                {/each}
              </div>
            </section>
          {/if}
          <div class="message-list">
            {#each selectedJob.messages as message}
              <div class="message-row">
                <Badge variant={message.severity === 'error' ? 'destructive' : 'secondary'}>{message.severity}</Badge>
                <span>{message.summary}{message.detail ? ` · ${message.detail}` : ''}</span>
              </div>
            {/each}
            {#if selectedJob.messages.length === 0}
              <div class="quiet-row"><CheckCircle2 size={16} aria-hidden="true" /> No import messages.</div>
            {/if}
          </div>
          <div class="action-row">
            <a class="detail-link" href={auditHistoryHref()}>
              View audit history
            </a>
            {#if canCreateImports && activeJobs.some((job) => job.id === selectedJob?.id)}
              <Button.Root variant="outline" onclick={requestSelectedJobCancellation} disabled={busy}>
                Cancel
              </Button.Root>
            {/if}
            {#if canCreateImports && selectedJob.status === 'previewed'}
              <Button.Root onclick={() => { if (selectedJob) resumePreviewedJob(selectedJob); }} disabled={busy}>
                Continue import
              </Button.Root>
            {/if}
            {#if canCreateImports && canRemoveJobFromHistory(selectedJob)}
              <Button.Root variant="ghost" onclick={removeSelectedJobFromHistory} disabled={busy}>
                <Trash2 size={16} aria-hidden="true" />
                Remove from history
              </Button.Root>
            {/if}
          </div>
      </Card.Content>
    </Card.Root>
    {/if}
  {/if}
</section>

<style>
  .import-workspace {
    display: grid;
    gap: 1rem;
    padding: 1.25rem;
  }

  .import-toolbar,
  .history-header,
  :global(.job-card),
  .job-main,
  :global(.check-row),
  .quiet-row,
  .message-row,
  .action-row {
    align-items: center;
    display: flex;
    gap: 0.75rem;
  }

  .import-toolbar,
  .history-header,
  :global(.job-card) {
    justify-content: space-between;
  }

  h1,
  h2,
  h3 {
    margin: 0;
  }

  h1 {
    font-size: 1.8rem;
    line-height: 1.2;
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

  .job-section,
  :global(.step-content) {
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

  .history-summary-grid {
    display: grid;
    gap: 0.75rem;
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .history-summary-grid div {
    border: 1px solid hsl(var(--border));
    border-radius: 8px;
    display: grid;
    gap: 0.15rem;
    min-width: 0;
    padding: 0.75rem;
  }

  .history-summary-grid div.summary-active {
    background: hsl(var(--primary) / 0.06);
    border-color: hsl(var(--primary) / 0.28);
  }

  .history-summary-grid div.summary-warning {
    background: hsl(var(--destructive) / 0.06);
    border-color: hsl(var(--destructive) / 0.28);
  }

  .history-summary-grid span {
    color: hsl(var(--muted-foreground));
    font-size: 0.78rem;
  }

  .history-summary-grid strong {
    font-size: 1.45rem;
    line-height: 1;
  }

  .source-choice-grid {
    display: grid;
    gap: 1rem;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .source-choice-grid :global(.source-card) {
    align-items: flex-start;
    display: flex;
    gap: 0.85rem;
    justify-content: flex-start;
    padding: 1rem;
    text-align: left;
    white-space: normal;
  }

  .source-choice-grid :global(.source-card) {
    min-height: 8rem;
  }

  .source-choice-grid strong,
  .source-choice-grid small {
    display: block;
  }

  .source-choice-grid small,
  .source-choice-grid em {
    color: hsl(var(--muted-foreground));
    font-size: 0.85rem;
    font-weight: 400;
    line-height: 1.35;
    margin-top: 0.35rem;
  }

  .source-choice-grid em {
    color: hsl(var(--foreground));
    font-size: 0.78rem;
    font-style: normal;
    font-weight: 600;
  }

  .field-grid {
    display: grid;
    gap: 0.75rem;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .field-stack {
    display: grid;
    gap: 0.35rem;
  }

  .field-note {
    color: hsl(var(--muted-foreground));
    font-size: 0.82rem;
  }

  .action-row {
    flex-wrap: wrap;
  }

  .source-summary,
  .preview-readiness,
  .readiness-panel,
  .detail-topline {
    align-items: center;
    display: flex;
    gap: 0.75rem;
    justify-content: space-between;
  }

  .readiness-panel {
    border: 1px solid hsl(var(--primary) / 0.28);
    border-radius: 8px;
    background: hsl(var(--primary) / 0.06);
    padding: 0.85rem;
  }

  .readiness-panel.needs-attention {
    background: hsl(var(--destructive) / 0.06);
    border-color: hsl(var(--destructive) / 0.3);
  }

  .readiness-panel span {
    color: hsl(var(--muted-foreground));
    display: block;
    font-size: 0.86rem;
    margin-top: 0.2rem;
  }

  .source-summary {
    background: hsl(var(--muted) / 0.45);
    border: 1px solid hsl(var(--border));
    border-radius: 8px;
    padding: 0.75rem;
  }

  .source-summary small {
    color: hsl(var(--muted-foreground));
    font-size: 0.78rem;
  }

  .detail-topline span {
    color: hsl(var(--muted-foreground));
    display: block;
    font-size: 0.85rem;
    margin-top: 0.15rem;
  }

  .progress-track {
    background: hsl(var(--muted));
    border-radius: 999px;
    height: 0.45rem;
    margin-top: 0.45rem;
    overflow: hidden;
    width: min(22rem, 100%);
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

  .summary-grid {
    display: grid;
    gap: 0.75rem;
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .summary-grid div {
    border: 1px solid hsl(var(--border));
    border-radius: 8px;
    padding: 0.75rem;
  }

  .summary-grid div.muted-count {
    background: hsl(var(--muted) / 0.25);
  }

  .summary-grid strong {
    display: block;
    font-size: 1.35rem;
  }

  .summary-grid span,
  .job-main span {
    color: hsl(var(--muted-foreground));
    display: block;
    font-size: 0.82rem;
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
  }

  .history-meta span:not(:last-child)::after {
    color: hsl(var(--border));
    content: "·";
    margin-left: 0.65rem;
  }

  .message-list {
    display: grid;
    gap: 0.5rem;
  }

  .preview-samples {
    display: grid;
    gap: 0.75rem;
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .preview-samples section {
    border: 1px solid hsl(var(--border));
    border-radius: 8px;
    display: grid;
    gap: 0.6rem;
    min-width: 0;
    padding: 0.75rem;
  }

  .sample-heading {
    align-items: baseline;
    display: flex;
    gap: 0.5rem;
    justify-content: space-between;
  }

  .sample-heading small,
  .sample-row small {
    color: hsl(var(--muted-foreground));
    font-size: 0.78rem;
  }

  .sample-list {
    display: grid;
    gap: 0.45rem;
  }

  .timeline-section {
    border: 1px solid hsl(var(--border));
    border-radius: 8px;
    display: grid;
    gap: 0.65rem;
    padding: 0.75rem;
  }

  .timeline-list {
    display: grid;
    gap: 0.35rem;
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

  .history-row {
    align-items: center;
    border: 1px solid hsl(var(--border));
    border-radius: 8px;
    display: grid;
    gap: 0.75rem;
    grid-template-columns: auto minmax(0, 1fr) auto auto;
    padding: 0.8rem;
  }

  .history-row:hover {
    background: hsl(var(--muted) / 0.25);
  }

  .status-icon {
    color: hsl(var(--muted-foreground));
    display: grid;
    place-items: center;
  }

  .advanced-options {
    border: 1px solid hsl(var(--border));
    border-radius: 8px;
    padding: 0.75rem;
  }

  .advanced-options summary {
    color: hsl(var(--muted-foreground));
    cursor: pointer;
    font-size: 0.88rem;
    font-weight: 600;
  }

  .advanced-options[open] summary {
    color: hsl(var(--foreground));
    margin-bottom: 0.7rem;
  }

  .advanced-option-list {
    display: grid;
    gap: 0.6rem;
  }

  :global(.empty-state) {
    align-items: center;
    display: grid;
    gap: 1rem;
    grid-template-columns: auto minmax(0, 1fr) auto;
  }

  .import-alert {
    background: hsl(var(--destructive) / 0.08);
    border: 1px dashed hsl(var(--destructive) / 0.35);
    border-radius: 8px;
    color: hsl(var(--destructive));
    padding: 1rem;
  }

  .import-notice {
    background: hsl(var(--muted) / 0.55);
    border: 1px solid hsl(var(--border));
    border-radius: 8px;
    color: hsl(var(--foreground));
    padding: 1rem;
  }

  :global(.spin) {
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .progress-track span,
    .progress-track.indeterminate span,
    :global(.spin) {
      animation: none;
      transition: none;
    }

    .progress-track.indeterminate span {
      transform: none;
      width: 42%;
    }
  }

  @media (max-width: 860px) {
    .import-toolbar,
    .history-header,
    :global(.job-card),
    .job-main,
    :global(.check-row) {
      align-items: flex-start;
      flex-direction: column;
    }

    .source-choice-grid,
    .field-grid,
    .history-summary-grid,
    .preview-samples,
    .summary-grid,
    .history-row,
    :global(.empty-state) {
      grid-template-columns: 1fr;
    }

    .source-summary,
    .preview-readiness,
    .readiness-panel,
    .detail-topline {
      align-items: flex-start;
      flex-direction: column;
    }

    .history-meta {
      display: grid;
      gap: 0.2rem;
    }

    .history-meta span:not(:last-child)::after {
      content: "";
      margin-left: 0;
    }
  }
</style>
