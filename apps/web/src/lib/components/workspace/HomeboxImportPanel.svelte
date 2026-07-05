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
    importMessageDetail,
    importMessageTone,
    importMissingInventoryPresentation,
    importPreviewSourceSummary,
    importPlannedCountLabel,
    importSourceOptions,
    importSourceSummary,
    isImportPreviewReady
  } from '$lib/application/workspaceImportPresentation';
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
  let preview = $state<ImportPreview | null>(null);
  let result = $state<ImportApplyResult | null>(null);
  let busy = $state(false);
  let error = $state('');
  let refreshWarning = $state('');
  let previousSourceType = $state(sourceType);
  let linkedSourceOptions = $derived(importSourceOptions(tenantId, inventory?.id ?? null));
  let sourceSummary = $derived(importSourceSummary(sourceType, fileName));
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
  let canApply = $derived(!!preview && !result && blockingErrors.length === 0 && !busy && canImport);
  let applyStatus = $derived(
    importApplyStatus({
      busy,
      hasPreview: !!preview,
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

  async function runPreview(): Promise<void> {
    if (!inventory || !ready || !canImport) {
      return;
    }
    busy = true;
    error = '';
    refreshWarning = '';
    result = null;
    try {
      preview = await repository.previewLegacyHomeboxImport(tenantId, inventory.id, importRequest());
    } catch (caught) {
      error = caught instanceof Error ? caught.message : 'Import preview failed.';
    } finally {
      busy = false;
    }
  }

  async function applyImport(): Promise<void> {
    if (!inventory || !canApply) {
      return;
    }
    busy = true;
    error = '';
    refreshWarning = '';
    try {
      const applied = await repository.applyLegacyHomeboxImport(tenantId, inventory.id, importRequest());
      result = applied;
    } catch (caught) {
      error = caught instanceof Error ? caught.message : 'Import failed.';
      busy = false;
      return;
    }
    try {
      await onImported();
    } catch (caught) {
      refreshWarning = caught instanceof Error ? caught.message : 'Import applied, but the workspace could not refresh.';
    } finally {
      busy = false;
    }
  }

  async function selectCSV(event: Event): Promise<void> {
    const input = event.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    preview = null;
    result = null;
    error = '';
    refreshWarning = '';
    if (!file) {
      fileName = '';
      contentBase64 = '';
      return;
    }
    fileName = file.name;
    contentBase64 = await fileToBase64(file);
  }

  function selectSource(nextSourceType: ImportSourceType): void {
    sourceType = nextSourceType;
    clearImportState();
    onSourceChange(nextSourceType);
  }

  function clearImportState(): void {
    preview = null;
    result = null;
    error = '';
    refreshWarning = '';
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
    if (option === 'images') {
      includeImages = !includeImages;
    } else if (option === 'insecure-tls') {
      allowInsecureTLS = !allowInsecureTLS;
    } else {
      allowPrivateNetwork = !allowPrivateNetwork;
    }
  }

  async function fileToBase64(file: File): Promise<string> {
    const buffer = await file.arrayBuffer();
    const bytes = new Uint8Array(buffer);
    let binary = '';
    const chunkSize = 0x8000;
    for (let index = 0; index < bytes.length; index += chunkSize) {
      binary += String.fromCharCode(...bytes.subarray(index, index + chunkSize));
    }
    return btoa(binary);
  }
</script>

<section class="workspace-main" aria-labelledby="import-title">
  <div class="section-heading">
    <div>
      <h1 id="import-title">Import</h1>
      <p>Bring Homebox records into {inventory?.name ?? 'this inventory'}.</p>
    </div>
    {#if preview}
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
      <form class="settings-panel import-source-panel" onsubmit={(event) => { event.preventDefault(); void runPreview(); }}>
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
            <Input id="homebox-url" bind:value={baseUrl} placeholder="https://homebox.example.com" />
          </div>
          <div class="field-stack">
            <Label for="homebox-username">User</Label>
            <Input id="homebox-username" bind:value={username} autocomplete="username" />
          </div>
          <div class="field-stack">
            <Label for="homebox-password">Password</Label>
            <Input id="homebox-password" bind:value={password} type="password" autocomplete="current-password" />
          </div>
          <div class="import-option-list" aria-label="Live Homebox import options">
            <BinaryOption
              label="Images"
              description="Import Homebox image attachments when available."
              checked={includeImages}
              icon={Image}
              onToggle={() => toggleImportOption('images')}
            />
            <BinaryOption
              label="Self-signed certificate"
              description="Allow a Homebox server with an untrusted TLS certificate."
              checked={allowInsecureTLS}
              onToggle={() => toggleImportOption('insecure-tls')}
            />
            <BinaryOption
              label="Private network address"
              description="Allow connections to private LAN addresses."
              checked={allowPrivateNetwork}
              onToggle={() => toggleImportOption('private-network')}
            />
          </div>
        {:else}
          <div class="field-stack">
            <Label for="homebox-csv">CSV file</Label>
            <Input id="homebox-csv" type="file" accept=".csv,text/csv" onchange={(event) => { void selectCSV(event); }} />
          </div>
          <p class="muted-note">CSV imports do not include image bytes.</p>
        {/if}

        <div class="heading-actions">
          <Button.Root type="submit" disabled={busy || !ready}>{busy ? 'Working' : 'Preview'}</Button.Root>
          <Button.Root
            type="button"
            variant="outline"
            disabled={!canApply}
            aria-describedby={applyStatusId}
            onclick={() => { void applyImport(); }}
          >
            Apply
          </Button.Root>
        </div>
        <p id={applyStatusId} class="muted-note" aria-live={canApply ? undefined : 'polite'}>{applyStatus}</p>
      </form>

      <div class="import-results">
        {#if error}
          <Alert.Root variant="destructive">
            <AlertTriangle aria-hidden="true" />
            <Alert.Title>Import failed</Alert.Title>
            <Alert.Description>{error}</Alert.Description>
          </Alert.Root>
        {/if}

        {#if preview}
          <section class="settings-panel wide" aria-labelledby="import-preview-title">
            <div class="settings-panel-heading">
              <CheckCircle2 aria-hidden="true" />
              <div>
                <h2 id="import-preview-title">{preview.source.name}</h2>
                <p>{importPreviewSourceSummary(preview.source)}</p>
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
              <div class="import-message-list">
                {#each preview.messages.slice(0, 12) as message}
                  <div class="import-message-row">
                    <Badge variant={importMessageTone(message)}>{message.severity}</Badge>
                    <span>
                      <strong>{message.summary}</strong>
                      <small>{importMessageDetail(message)}</small>
                    </span>
                  </div>
                {/each}
              </div>
            {/if}
          </section>

          <section class="settings-panel wide" aria-labelledby="import-fields-title">
            <h2 id="import-fields-title">Field definitions</h2>
            <div class="schema-list import-compact-list">
              {#each preview.fields.slice(0, 10) as field}
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
            <h2 id="import-assets-title">Asset samples</h2>
            <div class="asset-list">
              {#each preview.assetSamples.slice(0, 8) as asset}
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
              <h2 id="import-images-title">Image samples</h2>
              <div class="attachment-section">
                {#each preview.imageSamples.slice(0, 6) as image}
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
