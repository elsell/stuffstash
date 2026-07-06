<script lang="ts">
  import AlertTriangle from '@lucide/svelte/icons/alert-triangle';
  import CheckCircle2 from '@lucide/svelte/icons/check-circle-2';
  import Image from '@lucide/svelte/icons/image';
  import Upload from '@lucide/svelte/icons/upload';
  import * as Alert from '$lib/components/ui/alert/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import {
    buildLegacyHomeboxImportRequest,
    importAppliedDescription,
    importApplyMessagesPresentation,
    importApplyStatus,
    importDeniedPresentation,
    importEmptyPreviewPresentation,
    importFailurePresentation,
    importHiddenCount,
    importHiddenLabel,
    importMessageDetail,
    importMessageTone,
    importMessagesForDisplay,
    importMissingInventoryPresentation,
    importOperationPresentation,
    importPreviewDisplayLimits,
    importPreviewStatus,
    importPreviewSourceSummary,
    importPlannedCountLabel,
    importSampleLimitLabel,
    importWorkflowSteps,
    csvFileTooLargePresentation,
    legacyHomeboxImportRequestKey,
    maxHomeboxCSVBytes,
    importSourceOptions,
    importSourceSummary,
    isImportPreviewReady
  } from '$lib/application/workspaceImportPresentation';
  import { fileToBase64 } from '$lib/application/fileEncoding';
  import type {
    ImportApplyResult,
    ImportPreview,
    ImportSourceType,
    Inventory,
    LegacyHomeboxImportRequest
  } from '$lib/domain/inventory';
  import { hasAccessPermission } from '$lib/domain/inventory';
  import type { InventoryRepository } from '$lib/ports/inventoryRepository';
  import BinaryOption from './BinaryOption.svelte';
  import SegmentedControl from './SegmentedControl.svelte';

  let {
    tenantId,
    inventory,
    repository,
    sourceType = $bindable<ImportSourceType>('legacy_homebox'),
    onSourceChange = (nextSourceType: ImportSourceType) => {
      sourceType = nextSourceType;
    },
    onImported
  }: {
    tenantId: string;
    inventory: Inventory | null;
    repository: InventoryRepository;
    sourceType?: ImportSourceType;
    onSourceChange?: (sourceType: ImportSourceType) => void;
    onImported: () => Promise<void>;
  } = $props();

  let baseUrl = $state('');
  let username = $state('');
  let password = $state('');
  let includeImages = $state(true);
  let allowInsecureTLS = $state(false);
  let allowPrivateNetwork = $state(false);
  let fileName = $state('');
  let contentBase64 = $state('');
  let csvVersion = $state('');
  let preview = $state<ImportPreview | null>(null);
  let previewRequestKey = $state('');
  let result = $state<ImportApplyResult | null>(null);
  let activeOperation = $state<'preview' | 'apply' | null>(null);
  let failure = $state<{ title: string; description: string } | null>(null);
  let refreshWarning = $state('');
  let previousSourceType = $state(sourceType);
  let previewRun = 0;
  let applyRun = 0;
  let fileReadRun = 0;
  let busy = $derived(activeOperation !== null);
  let linkedSourceOptions = $derived(
    importSourceOptions(tenantId, inventory?.id ?? null).map((option) => ({
      ...option,
      disabled: busy
    }))
  );
  let sourceSummary = $derived(importSourceSummary(sourceType, fileName));
  let currentImportRequest = $derived(importRequest());
  let currentImportRequestKey = $derived(legacyHomeboxImportRequestKey(currentImportRequest, csvVersion));
  let currentImportPlanKey = $derived(
    JSON.stringify({
      tenantId,
      inventoryId: inventory?.id ?? '',
      request: currentImportRequestKey
    })
  );
  let previousImportPlanKey = $state('');
  let missingInventoryPresentation = importMissingInventoryPresentation();
  let deniedPresentation = importDeniedPresentation();
  let emptyPreviewPresentation = importEmptyPreviewPresentation();
  let applyMessagesPresentation = importApplyMessagesPresentation();

  let canImport = $derived(hasAccessPermission(inventory?.access, 'configure'));
  let ready = $derived(
    isImportPreviewReady({
      hasInventory: !!inventory,
      sourceType,
      baseUrl,
      username,
      password,
      contentBase64
    })
  );
  let blockingErrors = $derived(preview?.messages.filter((message) => message.severity === 'error') ?? []);
  let warnings = $derived(preview?.messages.filter((message) => message.severity === 'warning') ?? []);
  let hasCurrentPreview = $derived(!!preview && previewRequestKey !== '' && previewRequestKey === currentImportPlanKey);
  let workflowSteps = $derived(
    importWorkflowSteps({
      ready,
      hasPreview: hasCurrentPreview,
      hasBlockingErrors: blockingErrors.length > 0,
      hasResult: !!result,
      activeOperation
    })
  );
  let operationMessage = $derived(activeOperation ? importOperationPresentation(activeOperation) : null);
  let previewMessage = $derived(preview ? importPreviewStatus(preview, blockingErrors.length) : null);
  let displayedPreviewMessages = $derived(preview ? importMessagesForDisplay(preview.messages) : []);
  let hiddenPreviewMessages = $derived(preview ? importHiddenCount(preview.messages.length, displayedPreviewMessages.length) : 0);
  let canApply = $derived(hasCurrentPreview && !result && blockingErrors.length === 0 && !busy && canImport);
  let applyStatus = $derived(
    importApplyStatus({
      activeOperation,
      hasPreview: hasCurrentPreview,
      blockingErrorCount: blockingErrors.length,
      canImport
    })
  );
  const applyStatusId = 'import-apply-status';

  $effect(() => {
    if (sourceType === previousSourceType) {
      return;
    }
    previousSourceType = sourceType;
    clearImportState();
  });

  $effect(() => {
    if (previousImportPlanKey === '') {
      previousImportPlanKey = currentImportPlanKey;
      return;
    }
    if (currentImportPlanKey === previousImportPlanKey) {
      return;
    }
    previousImportPlanKey = currentImportPlanKey;
    invalidatePreviewState();
  });

  async function runPreview(): Promise<void> {
    if (!inventory || !ready || !canImport) {
      return;
    }
    activeOperation = 'preview';
    failure = null;
    refreshWarning = '';
    preview = null;
    previewRequestKey = '';
    result = null;
    const run = ++previewRun;
    const request = currentImportRequest;
    const planKey = currentImportPlanKey;
    const requestSourceType = sourceType;
    try {
      const nextPreview = await repository.previewLegacyHomeboxImport(tenantId, inventory.id, request);
      if (run !== previewRun) {
        return;
      }
      preview = nextPreview;
      previewRequestKey = planKey;
    } catch (caught) {
      if (run !== previewRun) {
        return;
      }
      failure = importFailurePresentation('preview', requestSourceType, caught);
    } finally {
      if (run === previewRun) {
        activeOperation = null;
      }
    }
  }

  async function applyImport(): Promise<void> {
    if (!inventory || !canApply) {
      return;
    }
    activeOperation = 'apply';
    failure = null;
    refreshWarning = '';
    const run = ++applyRun;
    const request = currentImportRequest;
    const planKey = currentImportPlanKey;
    const requestSourceType = sourceType;
    try {
      const applied = await repository.applyLegacyHomeboxImport(tenantId, inventory.id, request);
      if (run !== applyRun || planKey !== currentImportPlanKey) {
        return;
      }
      result = applied;
    } catch (caught) {
      if (run !== applyRun || planKey !== currentImportPlanKey) {
        return;
      }
      failure = importFailurePresentation('apply', requestSourceType, caught);
      activeOperation = null;
      return;
    }
    try {
      await onImported();
    } catch {
      if (run !== applyRun || planKey !== currentImportPlanKey) {
        return;
      }
      refreshWarning = 'Import applied, but the workspace could not refresh. Reload the page to see the latest records.';
    } finally {
      if (run === applyRun) {
        activeOperation = null;
      }
    }
  }

  async function selectCSV(event: Event): Promise<void> {
    if (activeOperation === 'apply') {
      return;
    }
    const input = event.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    const run = ++fileReadRun;
    invalidatePreviewState();
    if (!file) {
      fileName = '';
      contentBase64 = '';
      csvVersion = '';
      return;
    }
    if (file.size > maxHomeboxCSVBytes) {
      fileName = '';
      contentBase64 = '';
      csvVersion = '';
      failure = csvFileTooLargePresentation(file.size);
      return;
    }
    failure = null;
    fileName = file.name;
    contentBase64 = '';
    let nextContentBase64 = '';
    try {
      nextContentBase64 = await fileToBase64(file);
    } catch {
      if (run === fileReadRun) {
        fileName = file.name;
        contentBase64 = '';
        csvVersion = '';
        failure = {
          title: 'CSV could not be read',
          description: 'Choose the Homebox CSV export again and try previewing.'
        };
      }
      return;
    }
    if (run !== fileReadRun || activeOperation !== null) {
      return;
    }
    contentBase64 = nextContentBase64;
    csvVersion = `${file.name}:${file.size}:${file.lastModified}:${run}`;
  }

  function selectSource(nextSourceType: ImportSourceType): void {
    if (activeOperation === 'apply') {
      return;
    }
    sourceType = nextSourceType;
    clearImportState();
    onSourceChange(nextSourceType);
  }

  function clearImportState(): void {
    applyRun += 1;
    fileReadRun += 1;
    invalidatePreviewState();
    failure = null;
    activeOperation = null;
  }

  function invalidatePreviewState(): void {
    previewRun += 1;
    if (activeOperation === 'preview') {
      activeOperation = null;
    }
    preview = null;
    previewRequestKey = '';
    result = null;
    if (!isCSVSelectionFailure(failure)) {
      failure = null;
    }
    refreshWarning = '';
  }

  function isCSVSelectionFailure(message: { title: string; description: string } | null): boolean {
    return message?.title === 'CSV is too large' || message?.title === 'CSV could not be read';
  }

  function importRequest(): LegacyHomeboxImportRequest {
    return buildLegacyHomeboxImportRequest({
      sourceType,
      baseUrl,
      username,
      password,
      includeImages,
      allowInsecureTLS,
      allowPrivateNetwork,
      fileName,
      contentBase64
    });
  }

  function toggleImportOption(option: 'images' | 'insecure-tls' | 'private-network'): void {
    if (busy) {
      return;
    }
    if (option === 'images') {
      includeImages = !includeImages;
    } else if (option === 'insecure-tls') {
      allowInsecureTLS = !allowInsecureTLS;
    } else {
      allowPrivateNetwork = !allowPrivateNetwork;
    }
  }
</script>

<section class="workspace-main" aria-labelledby="import-title">
  <div class="section-heading">
    <div>
      <h1 id="import-title">Import</h1>
      <p>Bring Homebox records into {inventory?.name ?? 'this inventory'}.</p>
    </div>
    {#if hasCurrentPreview && preview}
      <Badge variant={preview.counts.errors > 0 ? 'destructive' : 'secondary'}>
        {importPlannedCountLabel(preview)}
      </Badge>
    {/if}
  </div>

  {#if !inventory}
    <div class="empty-state spacious">
      <h2>{missingInventoryPresentation.title}</h2>
    </div>
  {:else if !canImport}
    <Alert.Root variant="destructive">
      <AlertTriangle aria-hidden="true" />
      <Alert.Title>{deniedPresentation.title}</Alert.Title>
      <Alert.Description>{deniedPresentation.description}</Alert.Description>
    </Alert.Root>
  {:else}
    <div class="import-layout">
      <form class="settings-panel import-source-panel" aria-busy={busy} onsubmit={(event) => { event.preventDefault(); void runPreview(); }}>
        <ol class="import-step-list" aria-label="Import workflow">
          {#each workflowSteps as step}
            <li data-state={step.state}>
              <span aria-hidden="true"></span>
              <div>
                <strong>{step.label}</strong>
                <small>{step.description}</small>
              </div>
            </li>
          {/each}
        </ol>

        <div class="settings-panel-heading">
          <Upload aria-hidden="true" />
          <div>
            <h2>Source</h2>
            <p>{sourceSummary}</p>
          </div>
        </div>

        <SegmentedControl
          label="Import source"
          value={sourceType}
          options={linkedSourceOptions}
          onSelect={(value) => selectSource(value as ImportSourceType)}
        />

        {#if sourceType === 'legacy_homebox'}
          <div class="field-stack">
            <Label for="homebox-url">Homebox URL</Label>
            <Input id="homebox-url" bind:value={baseUrl} placeholder="https://homebox.example.com" disabled={busy} />
          </div>
          <div class="field-stack">
            <Label for="homebox-username">User</Label>
            <Input id="homebox-username" bind:value={username} autocomplete="username" disabled={busy} />
          </div>
          <div class="field-stack">
            <Label for="homebox-password">Password</Label>
            <Input id="homebox-password" bind:value={password} type="password" autocomplete="current-password" disabled={busy} />
          </div>
          <div class="import-option-list" aria-label="Live Homebox import options">
            <BinaryOption
              label="Images"
              description="Import Homebox image attachments when available."
              checked={includeImages}
              icon={Image}
              disabled={busy}
              onToggle={() => toggleImportOption('images')}
            />
            <BinaryOption
              label="Self-signed certificate"
              description="Allow a Homebox server with an untrusted TLS certificate."
              checked={allowInsecureTLS}
              disabled={busy}
              onToggle={() => toggleImportOption('insecure-tls')}
            />
            <BinaryOption
              label="Private network address"
              description="Allow connections to private LAN addresses."
              checked={allowPrivateNetwork}
              disabled={busy}
              onToggle={() => toggleImportOption('private-network')}
            />
          </div>
        {:else}
          <div class="field-stack">
            <Label for="homebox-csv">CSV file</Label>
            <Input id="homebox-csv" type="file" accept=".csv,text/csv" disabled={busy} aria-describedby="homebox-csv-note" onchange={(event) => { void selectCSV(event); }} />
          </div>
          <p id="homebox-csv-note" class="muted-note">CSV imports do not include image bytes. Maximum file size is 10 MB.</p>
        {/if}

        <div class="heading-actions">
          <Button.Root type="submit" disabled={busy || !ready}>{activeOperation === 'preview' ? 'Previewing' : 'Preview'}</Button.Root>
          <Button.Root
            type="button"
            variant="outline"
            disabled={!canApply}
            aria-describedby={applyStatusId}
            onclick={() => { void applyImport(); }}
          >
            {activeOperation === 'apply' ? 'Applying' : 'Apply'}
          </Button.Root>
        </div>
        <p id={applyStatusId} class="muted-note" aria-live={canApply ? undefined : 'polite'}>{applyStatus}</p>
      </form>

      <div class="import-results" aria-busy={busy}>
        {#if operationMessage}
          <Alert.Root>
            <Upload aria-hidden="true" />
            <Alert.Title>{operationMessage.title}</Alert.Title>
            <Alert.Description>{operationMessage.description}</Alert.Description>
          </Alert.Root>
        {/if}

        {#if result}
          <Alert.Root>
            <CheckCircle2 aria-hidden="true" />
            <Alert.Title>Import applied</Alert.Title>
            <Alert.Description>
              {importAppliedDescription(result)}
            </Alert.Description>
          </Alert.Root>
          {#if refreshWarning}
            <Alert.Root variant="default">
              <AlertTriangle aria-hidden="true" />
              <Alert.Title>Workspace refresh needed</Alert.Title>
              <Alert.Description>{refreshWarning}</Alert.Description>
            </Alert.Root>
          {/if}
        {/if}

        {#if failure}
          <Alert.Root variant="destructive">
            <AlertTriangle aria-hidden="true" />
            <Alert.Title>{failure.title}</Alert.Title>
            <Alert.Description>{failure.description}</Alert.Description>
          </Alert.Root>
        {/if}

        {#if hasCurrentPreview && preview}
          <section class="settings-panel wide" aria-labelledby="import-preview-title">
            <div class="settings-panel-heading">
              <CheckCircle2 aria-hidden="true" />
              <div>
                <h2 id="import-preview-title">{previewMessage?.title}</h2>
                <p>{previewMessage?.description}</p>
              </div>
            </div>

            <div class="import-stat-grid">
              <div><strong>{preview.counts.locations}</strong><span>Locations</span></div>
              <div><strong>{preview.counts.assets}</strong><span>Items</span></div>
              <div><strong>{preview.counts.fields}</strong><span>Fields</span></div>
              <div><strong>{preview.counts.attachments}</strong><span>Images</span></div>
              <div><strong>{warnings.length}</strong><span>Warnings</span></div>
              <div><strong>{blockingErrors.length}</strong><span>Errors</span></div>
            </div>

            {#if preview.messages.length > 0}
              <div class="import-section-subhead">
                <h3>Messages</h3>
                <small>{importSampleLimitLabel(preview.messages.length, displayedPreviewMessages.length, 'message')}</small>
              </div>
              <div class="import-message-list">
                {#each displayedPreviewMessages as message}
                  <div class="import-message-row">
                    <Badge variant={importMessageTone(message)}>{message.severity}</Badge>
                    <span>
                      <strong>{message.summary}</strong>
                      <small>{importMessageDetail(message)}</small>
                    </span>
                  </div>
                {/each}
              </div>
              {#if hiddenPreviewMessages > 0}
                <p class="muted-note">{importHiddenLabel(hiddenPreviewMessages, 'message')}</p>
              {/if}
            {/if}
          </section>

          <section class="settings-panel wide" aria-labelledby="import-fields-title">
            <div class="import-section-subhead">
              <h2 id="import-fields-title">Field definitions</h2>
              <small>{importSampleLimitLabel(preview.fields.length, Math.min(preview.fields.length, importPreviewDisplayLimits.fields), 'field')}</small>
            </div>
            <div class="schema-list import-compact-list">
              {#each preview.fields.slice(0, importPreviewDisplayLimits.fields) as field}
                <div class="schema-row">
                  <div>
                    <strong>{field.displayName}</strong>
                    <small>{field.key}</small>
                  </div>
                  <Badge variant="outline">{field.type}</Badge>
                </div>
              {/each}
            </div>
          </section>

          <section class="settings-panel wide" aria-labelledby="import-assets-title">
            <div class="import-section-subhead">
              <h2 id="import-assets-title">Asset samples</h2>
              <small>{importSampleLimitLabel(preview.assetSamples.length, Math.min(preview.assetSamples.length, importPreviewDisplayLimits.assets), 'sample')}</small>
            </div>
            <div class="asset-list">
              {#each preview.assetSamples.slice(0, importPreviewDisplayLimits.assets) as asset}
                <div class="asset-row import-sample-row">
                  <span class="asset-thumb asset-thumb-sm">{asset.kind === 'location' ? 'L' : 'I'}</span>
                  <span class="asset-row-main">
                    <strong>{asset.title}</strong>
                    <small>{asset.kind}{asset.parentSourceId ? ` / parent ${asset.parentSourceId}` : ''}</small>
                  </span>
                  <span class="asset-row-meta"><small>{asset.sourceId}</small></span>
                </div>
              {/each}
            </div>
          </section>

          {#if preview.imageSamples.length > 0}
            <section class="settings-panel wide" aria-labelledby="import-images-title">
              <div class="import-section-subhead">
                <h2 id="import-images-title">Image attachments</h2>
                <small>{importSampleLimitLabel(preview.imageSamples.length, Math.min(preview.imageSamples.length, importPreviewDisplayLimits.images), 'image')}</small>
              </div>
              <div class="attachment-section">
                {#each preview.imageSamples.slice(0, importPreviewDisplayLimits.images) as image}
                  <div class="attachment-row">
                    <span class="asset-thumb asset-thumb-sm"><Image aria-hidden="true" /></span>
                    <span>
                      <strong>{image.fileName}</strong>
                      <small>{image.contentType} / {Math.ceil(image.sizeBytes / 1024)} KB</small>
                    </span>
                    {#if image.primary}<Badge variant="secondary">Primary</Badge>{/if}
                  </div>
                {/each}
              </div>
            </section>
          {/if}
        {:else}
          <div class="empty-state spacious">
            <h2>{emptyPreviewPresentation.title}</h2>
            <p>{emptyPreviewPresentation.description}</p>
          </div>
        {/if}

        {#if result}
          {#if result.messages.length > 0}
            <section class="settings-panel wide" aria-labelledby="import-apply-messages-title">
              <h2 id="import-apply-messages-title">{applyMessagesPresentation.title}</h2>
              <div class="import-message-list">
                {#each result.messages.slice(0, 12) as message}
                  <div class="import-message-row">
                    <Badge variant={importMessageTone(message)}>{message.severity}</Badge>
                    <span>
                      <strong>{message.summary}</strong>
                      <small>{importMessageDetail(message)}</small>
                    </span>
                  </div>
                {/each}
              </div>
            </section>
          {/if}
        {/if}
      </div>
    </div>
  {/if}
</section>
