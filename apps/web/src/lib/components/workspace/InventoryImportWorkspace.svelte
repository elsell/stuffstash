<script lang="ts">
  import AlertCircle from '@lucide/svelte/icons/alert-circle';
  import RefreshCw from '@lucide/svelte/icons/refresh-cw';
  import { tick, untrack } from 'svelte';
  import {
    canCreateImportJob,
    canViewImportJobs,
    type ImportJob,
    type ImportJobCancellationMode,
    type Inventory,
    type ImportSourceRequest,
    type Principal
  } from '$lib/domain/inventory';
  import type { InventoryRepository } from '$lib/ports/inventoryRepository';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import { fileToBase64 } from '$lib/application/fileEncoding';
  import { workspaceRouteHref, type ImportDetailTabRoute, type ImportSourceRoute } from '$lib/application/workspaceRoute';
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import {
    buildImportSourceRequest,
    importSourceRequestKey,
    readableImportActionError,
    type ImportCSVSelection,
    type ImportSourceChoice
  } from '$lib/application/workspaceImportRequest';
  import {
    canRemoveJobFromHistory,
    canRequestCancellation,
    importIssueTone,
    progressSummary,
    sourceDescription,
    terminalJobMayHaveChangedInventory
  } from './importWorkspacePresentation';
  import ImportJobDetailPanel from './ImportJobDetailPanel.svelte';
  import ImportJobHistory from './ImportJobHistory.svelte';
  import ImportJobConfirmationPanel from './ImportJobConfirmationPanel.svelte';
  import ImportJobRunHandoff from './ImportJobRunHandoff.svelte';
  import ImportPreviewPanel from './ImportPreviewPanel.svelte';
  import ImportSourceChoiceStep from './ImportSourceChoiceStep.svelte';
  import ImportSourceSetup from './ImportSourceSetup.svelte';

  type FlowStep = 'history' | 'source' | 'setup' | 'preview' | 'run' | 'detail';
  type ImportFlowStepID = 'source' | 'connect' | 'preview' | 'run';

  const MAX_CSV_BYTES = 10 * 1024 * 1024;

  type Props = {
    tenantId: string;
    inventory: Inventory | null;
    currentPrincipal?: Principal;
    repository: InventoryRepository;
    importSource: ImportSourceRoute;
    importJobId: string | null;
    importTab: ImportDetailTabRoute | null;
    onImportSourceChange: (source: ImportSourceRoute) => void;
    onImportJobSelectionChange: (jobId: string | null, tab?: ImportDetailTabRoute | null) => void;
    onImportJobTabChange: (tab: ImportDetailTabRoute | null) => void;
    onImportJobInventoryChanged: (scope: ImportJobInventoryRefreshScope) => Promise<void>;
    onOpenImportedAssetId: (assetId: string) => Promise<void>;
    onOpenInventoryAuditHistory: () => void;
  };

  type ImportJobInventoryRefreshScope = {
    tenantId: string;
    inventoryId: string;
  };

  let {
    tenantId,
    inventory,
    currentPrincipal,
    repository,
    importSource,
    importJobId,
    importTab,
    onImportSourceChange,
    onImportJobSelectionChange,
    onImportJobTabChange,
    onImportJobInventoryChanged,
    onOpenImportedAssetId,
    onOpenInventoryAuditHistory
  }: Props = $props();

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
  let startedJob = $state<ImportJob | null>(null);
  let selectedJob = $state<ImportJob | null>(null);
  let detailTab = $state<ImportDetailTabRoute>('overview');
  let appliedImportJobRouteKey = $state('');
  let suppressDetailTabRouteUpdate = false;
  let cancelIntent = $state<{ job: ImportJob } | null>(null);
  let removeIntent = $state<{ job: ImportJob } | null>(null);
  let loading = $state(false);
  let busy = $state(false);
  let detailLoading = $state(false);
  let manualRefreshLoading = $state(false);
  let error = $state('');
  let notice = $state('');
  let fileSelectionToken = 0;
  let jobLoadSequence = 0;
  let visibleJobLoadSequence = 0;
  let detailLoadSequence = 0;
  let actionSequence = 0;
  let importWorkspaceElement: HTMLElement | null = null;

  let activeJobs = $derived(jobs.filter((job) => ['running', 'cancel_requested'].includes(job.status)));
  let draftJobs = $derived(jobs.filter((job) => job.status === 'previewed'));
  let currentWorkJobs = $derived([...activeJobs, ...draftJobs]);
  let terminalJobs = $derived(jobs.filter((job) => !['running', 'cancel_requested', 'previewed'].includes(job.status)));
  let completedJobs = $derived(jobs.filter((job) => job.status === 'succeeded' && importIssueTone(job) !== 'action'));
  let attentionJobs = $derived(jobs.filter((job) => importIssueTone(job) === 'action'));
  let canViewImports = $derived(canViewImportJobs(inventory));
  let canCreateImports = $derived(canCreateImportJob(inventory));
  let currentRequestKey = $derived(requestKey());
  let previewReady = $derived(previewJob?.status === 'previewed' && previewJob.counts.errors === 0 && previewedRequestKey === currentRequestKey);
  let previewStale = $derived(Boolean(previewJob && previewedRequestKey && previewedRequestKey !== currentRequestKey));
  let canConfirmSource = $derived(
    Boolean(canCreateImports && (sourceChoice === 'homebox_csv' ? contentBase64 : baseUrl.trim() && username.trim() && password))
  );
  let availableWizardSteps = $derived(reachableWizardSteps());

  $effect(() => {
    tenantId;
    inventory?.id;
    untrack(() => {
      void loadJobs();
    });
  });

  $effect(() => {
    importJobId;
    importTab;
    jobs;
    loading;
    canViewImports;
    untrack(() => {
      applyImportJobRoute();
    });
  });

  $effect(() => {
    const tab = detailTab;
    if (suppressDetailTabRouteUpdate || step !== 'detail' || !selectedJob) return;
    if (tab === importTab) return;
    onImportJobTabChange(tab);
  });

  $effect(() => {
    if (!canCreateImports && (step === 'source' || step === 'setup' || step === 'preview')) {
      step = 'history';
      previewJob = null;
      previewedRequestKey = '';
      onImportSourceChange(null);
      return;
    }
    if (importJobId) {
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
      reconcileStartedJob(nextJobs);
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
    startedJob = null;
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
    startedJob = null;
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
    startedJob = null;
    appliedImportJobRouteKey = '';
    onImportJobSelectionChange(null);
    onImportSourceChange(null);
    step = 'history';
  }

  function openJob(job: ImportJob): void {
    showJobDetail(job, null);
    onImportJobSelectionChange(job.id, null);
  }

  function showJobDetail(job: ImportJob, routeTab: ImportDetailTabRoute | null): void {
    selectedJob = job;
    detailTab = routeTab ?? defaultDetailTab(job);
    suppressDetailTabRouteUpdate = true;
    queueMicrotask(() => {
      suppressDetailTabRouteUpdate = false;
    });
    cancelIntent = null;
    removeIntent = null;
    error = '';
    notice = '';
    step = 'detail';
    void scrollImportWorkspaceToTop();
    void loadJobDetail(job);
  }

  async function scrollImportWorkspaceToTop(): Promise<void> {
    await tick();
    if (typeof navigator !== 'undefined' && navigator.userAgent.includes('jsdom')) {
      return;
    }
    try {
      window.scrollTo({ top: 0, left: 0, behavior: 'auto' });
      return;
    } catch {
      // Some test environments expose scrollTo without implementing it.
    }
    if (typeof importWorkspaceElement?.scrollIntoView === 'function') {
      importWorkspaceElement.scrollIntoView({ block: 'start', inline: 'nearest' });
    }
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

  async function loadJobDetail(job: ImportJob): Promise<void> {
    const scope = currentImportScope();
    if (!scope) return;
    const sequence = (detailLoadSequence += 1);
    detailLoading = true;
    try {
      const detail = await repository.getImportJob(scope.tenantId, scope.inventoryId, job.id);
      if (sequence !== detailLoadSequence || !isCurrentScope(scope) || selectedJob?.id !== job.id) return;
      const mergedDetail = { ...detail, actor: detail.actor ?? selectedJob.actor };
      selectedJob = mergedDetail;
      jobs = jobs.map((candidate) => (candidate.id === detail.id ? mergeImportJobDetailSnapshot(mergedDetail, candidate) : candidate));
    } catch {
      if (sequence !== detailLoadSequence || !isCurrentScope(scope) || selectedJob?.id !== job.id) return;
      notice = 'Import details could not be refreshed.';
    } finally {
      if (sequence === detailLoadSequence) {
        detailLoading = false;
      }
    }
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
      startedJob = job;
      previewJob = null;
      previewedRequestKey = '';
      clearSourceSecretAndFileState();
      onImportSourceChange(null);
      step = 'run';
      notice = '';
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
        onImportJobSelectionChange(null);
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

  function errorMessage(value: unknown, fallback: string): string {
    return readableImportActionError(value, fallback);
  }

  function sameCSVSelection(left: ImportCSVSelection | null, right: ImportCSVSelection): boolean {
    return Boolean(left && left.name === right.name && left.size === right.size && left.lastModified === right.lastModified);
  }

  function reconcileSelectedJob(nextJobs: ImportJob[]): void {
    if (!selectedJob) return;
    const summary = nextJobs.find((job) => job.id === selectedJob?.id);
    selectedJob = summary ? mergeImportJobDetailSnapshot(selectedJob, summary) : selectedJob;
  }

  function applyImportJobRoute(): void {
    if (!importJobId) {
      if (step === 'detail' && selectedJob && !importSource) {
        selectedJob = null;
        appliedImportJobRouteKey = '';
        step = 'history';
      }
      return;
    }
    if (!canViewImports) return;
    const routeKey = `${importJobId}:${importTab ?? ''}`;
    if (appliedImportJobRouteKey === routeKey && selectedJob?.id === importJobId) {
      if (importTab && detailTab !== importTab) {
        detailTab = importTab;
      }
      return;
    }
    if (selectedJob?.id === importJobId && step === 'detail') {
      appliedImportJobRouteKey = routeKey;
      if (importTab && detailTab !== importTab) {
        detailTab = importTab;
      }
      return;
    }
    const routedJob = jobs.find((job) => job.id === importJobId) ?? (selectedJob?.id === importJobId ? selectedJob : null);
    if (!routedJob) {
      if (!loading && jobs.length > 0) {
        error = 'That import run is not available in this inventory.';
        step = 'history';
      }
      return;
    }
    appliedImportJobRouteKey = routeKey;
    showJobDetail(routedJob, importTab);
  }

  function defaultDetailTab(job: ImportJob): ImportDetailTabRoute {
    return importIssueTone(job) === 'none' ? 'overview' : 'issues';
  }

  function reconcileStartedJob(nextJobs: ImportJob[]): void {
    if (!startedJob) return;
    startedJob = nextJobs.find((job) => job.id === startedJob?.id) ?? startedJob;
  }

  function mergeImportJobDetailSnapshot(detail: ImportJob, summary: ImportJob): ImportJob {
    return {
      ...detail,
      ...summary,
      actor: summary.actor ?? detail.actor,
      resources: summary.resources.length > 0 || summary.status === 'cancelled_discarded' ? summary.resources : detail.resources,
      messages: summary.messages.length > 0 ? summary.messages : detail.messages
    };
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
    if (attentionJobs.length > 0) return `${attentionJobs.length} ${attentionJobs.length === 1 ? 'import requires' : 'imports require'} action.`;
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

  function resourceAssetId(resource: ImportJob['resources'][number]): string | null {
    if (resource.resourceType === 'asset') return resource.resourceId;
    if (resource.resourceType === 'attachment' && resource.resourceOwnerId) return resource.resourceOwnerId;
    return null;
  }

  function openImportedResource(event: MouseEvent, resource: ImportJob['resources'][number]): void {
    if (!shouldHandleWorkspaceLinkClick(event)) return;
    const assetId = resourceAssetId(resource);
    if (!assetId) return;
    event.preventDefault();
    void onOpenImportedAssetId(assetId);
  }

  function openInventoryAuditHistory(event: MouseEvent): void {
    if (!shouldHandleWorkspaceLinkClick(event)) return;
    event.preventDefault();
    onOpenInventoryAuditHistory();
  }

  function importSourceRouteForChoice(choice: ImportSourceChoice): Exclude<ImportSourceRoute, null> {
    return choice === 'homebox_csv' ? 'homebox-csv' : 'homebox';
  }

  function importSourceHref(choice: ImportSourceChoice): string {
    return workspaceRouteHref({ mode: 'import', importSource: importSourceRouteForChoice(choice) }, tenantId, inventory?.id ?? null);
  }

  function reachableWizardSteps(): ImportFlowStepID[] {
    const steps: ImportFlowStepID[] = ['source'];
    if (sourceChoice) {
      steps.push('connect');
    }
    if (previewJob) {
      steps.push('preview');
    }
    if (startedJob) {
      steps.push('run');
    }
    return steps;
  }

  function navigateWizardStep(target: ImportFlowStepID): void {
    if (!canCreateImports || !availableWizardSteps.includes(target) || busy) return;
    error = '';
    notice = '';
    if (target === 'source') {
      onImportSourceChange(null);
      step = 'source';
      return;
    }
    if (target === 'connect') {
      onImportSourceChange(importSourceRouteForChoice(sourceChoice));
      step = 'setup';
      return;
    }
    if (target === 'preview' && previewJob) {
      onImportSourceChange(importSourceRouteForChoice(sourceChoice));
      step = 'preview';
      return;
    }
    if (target === 'run' && startedJob) {
      onImportSourceChange(null);
      onImportJobSelectionChange(startedJob.id, null);
      step = 'run';
    }
  }

  function returnToSourceChoice(): void {
    onImportSourceChange(null);
    step = 'source';
  }

  function returnToSetupFromPreview(): void {
    onImportSourceChange(importSourceRouteForChoice(sourceChoice));
    step = 'setup';
  }

  async function refreshVisibleImportView(): Promise<void> {
    if (step === 'detail' && selectedJob) {
      await loadJobDetail(selectedJob);
      return;
    }
    const scope = currentImportScope();
    if (!scope || manualRefreshLoading) return;
    manualRefreshLoading = true;
    try {
      await loadJobs({ quiet: true });
    } finally {
      if (isCurrentScope(scope)) {
        manualRefreshLoading = false;
      }
    }
  }

</script>

<section class="import-workspace" bind:this={importWorkspaceElement}>
  <div class="import-toolbar">
    <div>
      <h1>{step === 'history' ? 'Imports' : step === 'detail' ? 'Import details' : step === 'run' ? 'Import running' : 'New import'}</h1>
      <p>
        {#if step === 'history'}
          {`View current and past data imports for ${inventory?.name ?? 'this inventory'}.`}
        {:else if step === 'detail'}
          {selectedJob ? `${selectedJob.source.name} · ${sourceDescription(selectedJob)}` : `Import run for ${inventory?.name ?? 'this inventory'}.`}
        {:else if step === 'run'}
          The job is running in the background.
        {:else}
          Confirm the source, preview the plan, then run it in the background.
        {/if}
      </p>
    </div>
    {#if step === 'history'}
      <Button.Root variant="outline" size="sm" onclick={() => { void loadJobs(); }} disabled={loading || !canViewImports}>
        <Button.BusyContent busy={loading} icon={RefreshCw} label="Refresh" busyLabel="Refreshing" />
      </Button.Root>
    {:else}
      <div class="toolbar-actions">
        {#if step === 'detail' || step === 'run'}
          <Button.Root variant="outline" size="sm" onclick={() => { void refreshVisibleImportView(); }} disabled={loading || detailLoading || manualRefreshLoading || !canViewImports}>
            <Button.BusyContent busy={loading || detailLoading || manualRefreshLoading} icon={RefreshCw} label="Refresh" busyLabel="Refreshing" />
          </Button.Root>
        {/if}
        <Button.Root variant="outline" size="sm" onclick={returnToHistory} disabled={busy}>Back to history</Button.Root>
      </div>
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
  {/if}

  {#if canViewImports}
    {#if step === 'history'}
      <ImportJobHistory
        {jobs}
        {activeJobs}
        {draftJobs}
        {terminalJobs}
        {completedJobs}
        {attentionJobs}
        {currentWorkJobs}
        summaryDescription={historySummaryDescription()}
        {currentPrincipal}
        {canCreateImports}
        {busy}
        onBeginImport={beginImport}
        onOpenJob={openJob}
        onResumePreviewedJob={resumePreviewedJob}
        onRequestCancellation={(job) => (cancelIntent = { job })}
        onRequestRemove={(job) => (removeIntent = { job })}
      />
  {:else if step === 'source'}
    <ImportSourceChoiceStep
      liveHref={importSourceHref('homebox_live')}
      csvHref={importSourceHref('homebox_csv')}
      availableSteps={availableWizardSteps}
      onChoose={chooseSourceFromLink}
      onNavigateStep={navigateWizardStep}
    />
  {:else if step === 'setup'}
    <ImportSourceSetup
      {sourceChoice}
      availableSteps={availableWizardSteps}
      bind:baseUrl
      bind:username
      bind:password
      bind:includeImages
      bind:allowPrivateNetwork
      bind:allowInsecureTLS
      {csvSelection}
      {canConfirmSource}
      {busy}
      {inventory}
      onFileSelected={(event) => { void handleFile(event); }}
      onConfirmSource={() => { void confirmSource(); }}
      onNavigateStep={navigateWizardStep}
      onBack={returnToSourceChoice}
    />
  {:else if step === 'preview'}
    <ImportPreviewPanel
      {previewJob}
      availableSteps={availableWizardSteps}
      {previewReady}
      {previewStale}
      {busy}
      onStart={() => { void start(); }}
      onNavigateStep={navigateWizardStep}
      onBack={returnToSetupFromPreview}
      />
  {:else if step === 'run' && startedJob}
    <ImportJobRunHandoff
      job={startedJob}
      availableSteps={availableWizardSteps}
      onNavigateStep={navigateWizardStep}
      onViewHistory={returnToHistory}
    />
  {:else if step === 'detail' && selectedJob}
    <ImportJobDetailPanel
      job={selectedJob}
      bind:selectedTab={detailTab}
      canRequestCancellation={canRequestCancellation(selectedJob)}
      {detailLoading}
      {canCreateImports}
      {busy}
      {currentPrincipal}
      auditHistoryHref={auditHistoryHref()}
      {resourceCanOpen}
      {resourceHref}
      onOpenAuditHistory={openInventoryAuditHistory}
      onOpenResource={openImportedResource}
      onCancel={requestSelectedJobCancellation}
      onContinue={() => { if (selectedJob) resumePreviewedJob(selectedJob); }}
      onRemove={removeSelectedJobFromHistory}
    />
    {/if}
  {/if}

  {#if canViewImports && (cancelIntent || removeIntent)}
    <div class="import-confirmation-overlay" role="presentation">
      <div
        class="import-confirmation-backdrop"
        aria-hidden="true"
        onclick={() => {
          if (!busy) {
            cancelIntent = null;
            removeIntent = null;
          }
        }}
      ></div>
      <div
        class="import-confirmation-dialog"
        role="dialog"
        aria-modal="true"
        aria-label={cancelIntent ? 'Cancel import' : 'Remove import from history'}
      >
        <ImportJobConfirmationPanel
          cancelJob={cancelIntent?.job ?? null}
          removeJob={removeIntent?.job ?? null}
          {busy}
          onCancelJob={(job, mode) => { void cancel(job, mode); }}
          onDismissCancel={() => (cancelIntent = null)}
          onRemoveJob={(job) => { void removeFromHistory(job); }}
          onDismissRemove={() => (removeIntent = null)}
        />
      </div>
    </div>
  {/if}
</section>

<style>
  .import-workspace {
    display: grid;
    gap: 1rem;
    margin: 0 auto;
    max-width: 76rem;
    padding: 1.25rem;
    width: 100%;
  }

  .import-toolbar,
  .toolbar-actions,
  :global(.check-row) {
    align-items: center;
    display: flex;
    gap: 0.75rem;
  }

  .import-toolbar {
    border-bottom: 1px solid var(--border);
    justify-content: space-between;
    padding-bottom: 0.9rem;
  }

  .toolbar-actions {
    justify-content: flex-end;
  }

  .import-confirmation-overlay {
    align-items: center;
    display: grid;
    inset: 0;
    justify-items: center;
    padding: 1rem;
    position: fixed;
    z-index: 60;
  }

  .import-confirmation-backdrop {
    background: color-mix(in oklab, var(--background) 72%, transparent);
    inset: 0;
    position: absolute;
  }

  .import-confirmation-dialog {
    max-width: min(34rem, 100%);
    position: relative;
    width: 100%;
  }

  h1,
  h2 {
    margin: 0;
  }

  h1 {
    font-size: clamp(1.45rem, 2vw, 1.8rem);
    line-height: 1.2;
  }

  h2 {
    font-size: 1.25rem;
  }

  p {
    color: var(--muted-foreground);
    font-size: 0.95rem;
    margin: 0.25rem 0 0;
  }

  .import-alert {
    background: color-mix(in oklab, var(--destructive) 8%, transparent);
    border: 1px dashed color-mix(in oklab, var(--destructive) 35%, transparent);
    border-radius: 8px;
    color: var(--destructive);
    padding: 1rem;
  }

  .import-notice {
    background: color-mix(in oklab, var(--muted) 55%, transparent);
    border: 1px solid var(--border);
    border-radius: 8px;
    color: var(--foreground);
    padding: 1rem;
  }

  @media (max-width: 860px) {
    .import-workspace {
      gap: 0.9rem;
      padding: 1rem;
      padding-bottom: var(--mobile-scroll-clearance, 7rem);
    }

    .import-toolbar {
      align-items: flex-start;
      flex-wrap: wrap;
      gap: 0.75rem 1rem;
    }

    .import-toolbar > div:first-child {
      flex: 1 1 13rem;
      min-width: 0;
    }

    .toolbar-actions {
      align-items: center;
      flex-direction: row;
      flex-wrap: wrap;
    }
  }
</style>
