import { tick } from 'svelte';
import { mount, unmount } from 'svelte';
import { expect } from 'vitest';
import { SeededInventoryRepository } from '$lib/adapters/memory/seededInventoryRepository';
import type {
  AssetLifecycleFilter,
  ImportJob,
  ImportJobCancellationMode,
  ImportSourceRequest,
  Principal,
  WorkspaceData
} from '$lib/domain/inventory';
import type { WorkspaceSeed } from '$lib/ports/inventoryRepository';
import InventoryImportWorkspaceHarness from './InventoryImportWorkspace.test-harness.svelte';

let component: ReturnType<typeof mount> | null = null;

export const seed: WorkspaceSeed = {
  principal: { id: 'principal-one', email: 'owner@example.test' },
  tenants: [{ id: 'tenant-home', name: 'Home', access: { relationship: 'owner', permissions: ['view'] } }],
  inventories: [
    {
      id: 'inventory-household',
      tenantId: 'tenant-home',
      name: 'Household',
      access: {
        relationship: 'owner',
        permissions: ['view', 'create_asset', 'edit_asset', 'configure', 'view_import_job', 'create_import_job']
      }
    }
  ],
  customAssetTypes: [],
  customFieldDefinitions: [],
  assets: [
    {
      id: 'asset-home',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'item',
      title: 'Passport',
      description: 'Blue folder',
      parentAssetId: null,
      lifecycleState: 'active'
    }
  ]
};

export class ImportPreviewRecordingRepository extends SeededInventoryRepository {
  previewInputs: ImportSourceRequest[] = [];

  async previewImportJob(tenantId: string, inventoryId: string, input: ImportSourceRequest): Promise<ImportJob> {
    this.previewInputs.push(structuredClone(input));
    return super.previewImportJob(tenantId, inventoryId, input);
  }
}

export class ImportStartRecordingRepository extends ImportPreviewRecordingRepository {
  startInputs: Array<{ jobId: string; input: ImportSourceRequest }> = [];

  async startImportJob(tenantId: string, inventoryId: string, jobId: string, input: ImportSourceRequest): Promise<ImportJob> {
    this.startInputs.push({ jobId, input: structuredClone(input) });
    return super.startImportJob(tenantId, inventoryId, jobId, input);
  }
}

export class CompletingStartedImportRepository extends ImportStartRecordingRepository {
  private listCallsAfterStart = 0;

  async listImportJobs(tenantId: string, inventoryId: string): Promise<ImportJob[]> {
    const jobs = await super.listImportJobs(tenantId, inventoryId);
    const started = jobs.find((job) => job.status === 'running');
    if (!started) return jobs;
    this.listCallsAfterStart += 1;
    return jobs.map((job) =>
      job.id === started.id
        ? {
            ...job,
            status: 'succeeded',
            completedAt: '2026-07-06T12:05:00Z',
            updatedAt: '2026-07-06T12:05:00Z',
            progress: { phase: 'terminal', done: 1, total: 1, message: 'Import completed', updatedAt: '2026-07-06T12:05:00Z' },
            progressHistory: [
              ...job.progressHistory,
              { phase: 'terminal', done: 1, total: 1, message: 'Import completed', updatedAt: '2026-07-06T12:05:00Z' }
            ]
          }
        : job
    );
  }
}

export class ImportPreviewGenericInvalidRequestRepository extends SeededInventoryRepository {
  async previewImportJob(_tenantId: string, _inventoryId: string, _input: ImportSourceRequest): Promise<ImportJob> {
    throw Object.assign(new Error('Invalid request.'), { status: 400, code: 'invalid_request' });
  }
}

export class ImportPreviewCountRepository extends ImportPreviewRecordingRepository {
  async previewImportJob(tenantId: string, inventoryId: string, input: ImportSourceRequest): Promise<ImportJob> {
    const job = await super.previewImportJob(tenantId, inventoryId, input);
    job.counts.fields = 2;
    job.counts.locations = 1;
    job.counts.assets = 3;
    job.counts.attachments = 4;
    job.counts.fieldsExisting = 1;
    job.counts.assetsSkipped = 2;
    job.counts.attachmentsSkipped = 1;
    job.counts.warnings = 2;
    job.counts.errors = 1;
    return job;
  }
}

export class EmptyImportPreviewRepository extends ImportPreviewRecordingRepository {
  async previewImportJob(tenantId: string, inventoryId: string, input: ImportSourceRequest): Promise<ImportJob> {
    const job = await super.previewImportJob(tenantId, inventoryId, input);
    job.counts.fields = 0;
    job.counts.locations = 0;
    job.counts.assets = 0;
    job.counts.attachments = 0;
    job.counts.fieldsExisting = 0;
    job.counts.assetsSkipped = 0;
    job.counts.attachmentsSkipped = 0;
    job.counts.warnings = 0;
    job.counts.errors = 0;
    job.preview.fields = [];
    job.preview.locations = [];
    job.preview.assets = [];
    job.preview.attachments = [];
    return job;
  }
}

export class PreviewMessageOnlyRepository extends ImportPreviewRecordingRepository {
  async previewImportJob(tenantId: string, inventoryId: string, input: ImportSourceRequest): Promise<ImportJob> {
    const job = await super.previewImportJob(tenantId, inventoryId, input);
    job.counts.warnings = 1;
    job.messages = [];
    job.preview.messages = [
      {
        code: 'preview-only-warning',
        severity: 'warning',
        summary: 'Attachment will be skipped',
        detail: 'Homebox reported a file without downloadable bytes.'
      }
    ];
    return job;
  }
}

export class PreviewHierarchyRepository extends ImportPreviewRecordingRepository {
  async previewImportJob(tenantId: string, inventoryId: string, input: ImportSourceRequest): Promise<ImportJob> {
    const job = await super.previewImportJob(tenantId, inventoryId, input);
    job.preview.locations = [
      { sourceId: 'loc-garage', kind: 'location', title: 'Garage', archived: false },
      { sourceId: 'loc-shelf', kind: 'location', title: 'Shelf', parentSourceId: 'loc-garage', archived: false }
    ];
    job.preview.assets = [
      { sourceId: 'asset-drill', kind: 'item', title: 'Drill', parentSourceId: 'loc-shelf', archived: false }
    ];
    job.counts.locations = 2;
    job.counts.assets = 1;
    return job;
  }
}

export class TerminalImportJobRepository extends SeededInventoryRepository {
  protected job: ImportJob;

  constructor(seedData: typeof seed) {
    super(seedData);
    this.job = {
      id: 'job-terminal',
      status: 'succeeded',
      actorId: 'owner',
      source: {
        type: 'legacy_homebox',
        name: 'Homebox',
        baseUrl: 'http://homebox.local:7744',
        imageImport: 'enabled',
        allowPrivateNetwork: true,
        allowInsecureTLS: true,
        fingerprint: 'fingerprint-terminal'
      },
      counts: {
        fields: 0,
        locations: 0,
        assets: 1,
        attachments: 0,
        warnings: 2,
        errors: 0,
        fieldsCreated: 0,
        fieldsExisting: 1,
        locationsCreated: 1,
        assetsCreated: 1,
        assetsSkipped: 2,
        attachmentsCreated: 0,
        attachmentsSkipped: 1,
        recordsDiscarded: 0,
        sourceLinksDiscarded: 0
      },
      preview: {
        fields: [{ key: 'serial_number', displayName: 'Serial number', type: 'text' }],
        locations: [{ sourceId: 'loc-garage', kind: 'location', title: 'Garage', archived: false }],
        assets: [{ sourceId: 'asset-drill', kind: 'item', title: 'Cordless drill', parentSourceId: 'loc-garage', archived: false }],
        attachments: [
          {
            sourceId: 'attachment-drill-photo',
            assetSourceId: 'asset-drill',
            fileName: 'drill-photo.jpg',
            contentType: 'image/jpeg',
            sizeBytes: 42_000,
            primary: true
          }
        ],
        messages: [],
        fieldsTruncated: false,
        locationsTruncated: false,
        assetsTruncated: false,
        attachmentsTruncated: false,
        messagesTruncated: false
      },
      progress: { phase: 'terminal', done: 1, total: 1, message: 'Import completed', updatedAt: '2026-07-06T12:05:00Z' },
      progressHistory: [
        { phase: 'ready', done: 1, total: 1, message: 'Preview ready', updatedAt: '2026-07-06T12:00:00Z' },
        { phase: 'reading_source', done: 0, total: 0, message: 'Reading source', updatedAt: '2026-07-06T12:01:00Z' },
        { phase: 'terminal', done: 1, total: 1, message: 'Import completed', updatedAt: '2026-07-06T12:05:00Z' }
      ],
      createdAt: '2026-07-06T12:00:00Z',
      startedAt: '2026-07-06T12:01:00Z',
      completedAt: '2026-07-06T12:05:00Z',
      updatedAt: '2026-07-06T12:05:00Z',
      resources: [],
      messages: []
    };
  }

  protected expectScope(tenantId: string, inventoryId: string, jobId?: string): void {
    expect(tenantId).toBe('tenant-home');
    expect(inventoryId).toBe('inventory-household');
    if (jobId) {
      expect(jobId).toBe(this.job.id);
    }
  }

  async listImportJobs(tenantId: string, inventoryId: string): Promise<ImportJob[]> {
    this.expectScope(tenantId, inventoryId);
    return [this.job];
  }

  async getImportJob(tenantId: string, inventoryId: string, jobId: string): Promise<ImportJob> {
    this.expectScope(tenantId, inventoryId, jobId);
    return this.job;
  }

  async removeImportJobFromHistory(tenantId: string, inventoryId: string, jobId: string): Promise<void> {
    this.expectScope(tenantId, inventoryId, jobId);
  }
}

export class LongActorImportJobRepository extends TerminalImportJobRepository {
  constructor(seedData: typeof seed) {
    super(seedData);
    this.job = {
      ...this.job,
      actorId: 'oidc_vZWJGXPHf8OYYeSghzupLo9vyyywfxu9DKltriM27O9'
    };
  }
}

export class PreviewedImportJobRepository extends TerminalImportJobRepository {
  private previewedJob: ImportJob;

  constructor(seedData: typeof seed) {
    super(seedData);
    this.previewedJob = {
      ...this.job,
      id: 'job-previewed',
      status: 'previewed',
      startedAt: undefined,
      completedAt: undefined,
      progress: { phase: 'ready', done: 1, total: 1, message: 'Preview ready', updatedAt: '2026-07-06T12:00:00Z' },
      progressHistory: [{ phase: 'ready', done: 1, total: 1, message: 'Preview ready', updatedAt: '2026-07-06T12:00:00Z' }]
    };
  }

  async listImportJobs(tenantId: string, inventoryId: string): Promise<ImportJob[]> {
    expect(tenantId).toBe('tenant-home');
    expect(inventoryId).toBe('inventory-household');
    return [this.previewedJob];
  }

  async getImportJob(tenantId: string, inventoryId: string, jobId: string): Promise<ImportJob> {
    expect(tenantId).toBe('tenant-home');
    expect(inventoryId).toBe('inventory-household');
    expect(jobId).toBe(this.previewedJob.id);
    return this.previewedJob;
  }
}

export class PreviewedCSVImportJobRepository extends PreviewedImportJobRepository {
  async listImportJobs(tenantId: string, inventoryId: string): Promise<ImportJob[]> {
    const jobs = await super.listImportJobs(tenantId, inventoryId);
    return [
      {
        ...jobs[0],
        source: {
          ...jobs[0].source,
          type: 'legacy_homebox_csv',
          name: 'Homebox CSV',
          baseUrl: undefined,
          imageImport: 'unavailable'
        }
      }
    ];
  }
}

export class CompletingImportJobRepository extends TerminalImportJobRepository {
  listCalls = 0;

  constructor(seedData: typeof seed) {
    super(seedData);
    this.job = {
      ...this.job,
      status: 'running',
      completedAt: undefined,
      progress: { phase: 'assets', done: 0, total: 1, message: 'Creating assets', updatedAt: '2026-07-06T12:02:00Z' },
      progressHistory: [
        { phase: 'ready', done: 1, total: 1, message: 'Preview ready', updatedAt: '2026-07-06T12:00:00Z' },
        { phase: 'assets', done: 0, total: 1, message: 'Creating assets', updatedAt: '2026-07-06T12:02:00Z' }
      ]
    };
  }

  async listImportJobs(tenantId: string, inventoryId: string): Promise<ImportJob[]> {
    expect(tenantId).toBe('tenant-home');
    expect(inventoryId).toBe('inventory-household');
    this.listCalls += 1;
    if (this.listCalls === 1) {
      return [this.job];
    }
    return [
      {
        ...this.job,
        status: 'succeeded',
        completedAt: '2026-07-06T12:05:00Z',
        updatedAt: '2026-07-06T12:05:00Z',
        progress: { phase: 'terminal', done: 1, total: 1, message: 'Import completed', updatedAt: '2026-07-06T12:05:00Z' },
        progressHistory: [
          ...this.job.progressHistory,
          { phase: 'terminal', done: 1, total: 1, message: 'Import completed', updatedAt: '2026-07-06T12:05:00Z' }
        ]
      }
    ];
  }

  async selectAssetLifecycle(
    tenantId: string,
    inventoryId: string,
    lifecycleState: AssetLifecycleFilter
  ): Promise<WorkspaceData> {
    expect(tenantId).toBe('tenant-home');
    expect(inventoryId).toBe('inventory-household');
    expect(lifecycleState).toBe('active');
    return super.selectAssetLifecycle(tenantId, inventoryId, lifecycleState);
  }
}

export class UnknownProgressImportJobRepository extends TerminalImportJobRepository {
  constructor(seedData: typeof seed) {
    super(seedData);
    this.job = {
      ...this.job,
      status: 'running',
      completedAt: undefined,
      progress: { phase: 'reading_source', done: 0, total: 0, message: 'Reading source', updatedAt: '2026-07-06T12:02:00Z' },
      progressHistory: [{ phase: 'reading_source', done: 0, total: 0, message: 'Reading source', updatedAt: '2026-07-06T12:02:00Z' }]
    };
  }
}

export class TerminalUnknownProgressImportJobRepository extends TerminalImportJobRepository {
  constructor(seedData: typeof seed) {
    super(seedData);
    this.job = {
      ...this.job,
      status: 'succeeded',
      progress: { phase: 'terminal', done: 0, total: 0, message: 'Import completed', updatedAt: '2026-07-06T12:05:00Z' },
      progressHistory: [
        { phase: 'reading_source', done: 0, total: 0, message: 'Reading source', updatedAt: '2026-07-06T12:01:00Z' },
        { phase: 'terminal', done: 0, total: 0, message: 'Import completed', updatedAt: '2026-07-06T12:05:00Z' }
      ]
    };
  }
}

export class DiscardedImportJobRepository extends TerminalImportJobRepository {
  constructor(seedData: typeof seed) {
    super(seedData);
    this.job = {
      ...this.job,
      status: 'cancelled_discarded',
      cancellationMode: 'discard_partial_progress',
      counts: {
        ...this.job.counts,
        assetsCreated: 1,
        recordsDiscarded: 1,
        sourceLinksDiscarded: 1
      },
      progress: {
        phase: 'terminal',
        done: 1,
        total: 1,
        message: 'Import cancelled and partial progress discarded',
        updatedAt: '2026-07-06T12:06:00Z'
      },
      resources: [
        {
          resourceType: 'asset',
          resourceId: 'asset-discarded',
          sourceEntityType: 'asset',
          sourceEntityId: 'source:discarded',
          createdAt: '2026-07-06T12:03:00Z'
        }
      ],
      updatedAt: '2026-07-06T12:06:00Z',
      completedAt: '2026-07-06T12:06:00Z'
    };
  }
}

export class ResourcefulImportJobRepository extends TerminalImportJobRepository {
  constructor(seedData: typeof seed) {
    super(seedData);
    this.job = {
      ...this.job,
      resources: [
        {
          resourceType: 'asset',
          resourceId: 'asset-imported-passport',
          displayName: 'Passport',
          sourceEntityType: 'asset',
          sourceEntityId: 'homebox-item-passport',
          createdAt: '2026-07-06T12:03:00Z'
        },
        {
          resourceType: 'attachment',
          resourceId: 'attachment-passport-photo',
          displayName: 'passport-photo.jpg',
          resourceOwnerId: 'asset-imported-passport',
          sourceEntityType: 'attachment',
          sourceEntityId: 'homebox-photo-passport',
          createdAt: '2026-07-06T12:04:00Z'
        }
      ]
    };
  }
}

export class TerminalPreviewMessageImportJobRepository extends TerminalImportJobRepository {
  constructor(seedData: typeof seed) {
    super(seedData);
    this.job = {
      ...this.job,
      counts: {
        ...this.job.counts,
        warnings: 1
      },
      preview: {
        ...this.job.preview,
        messages: [
          {
            code: 'attachment-skipped',
            severity: 'warning',
            summary: 'Attachment could not be imported',
            detail: 'Homebox reported a file without downloadable bytes.',
            sourceName: 'receipt.png'
          }
        ]
      },
      messages: []
    };
  }
}

export class TerminalJobAndPreviewMessageImportJobRepository extends TerminalPreviewMessageImportJobRepository {
  constructor(seedData: typeof seed) {
    super(seedData);
    this.job = {
      ...this.job,
      messages: [
        {
          code: 'terminal-warning',
          severity: 'warning',
          summary: 'Attachment imported without primary-photo status',
          detail: 'Homebox did not identify a primary image after download.',
          sourceName: 'downloaded-manual.png'
        }
      ]
    };
  }
}

export class DetailOnlyResourcefulImportJobRepository extends TerminalImportJobRepository {
  detailCalls = 0;
  private readonly detailedJobs: ImportJob[];

  constructor(seedData: typeof seed) {
    super(seedData);
    const firstDetail = {
      ...this.job,
      resources: [
        {
          resourceType: 'asset',
          resourceId: 'asset-imported-detail',
          sourceEntityType: 'asset',
          sourceEntityId: 'homebox-detail-asset',
          createdAt: '2026-07-06T12:04:00Z'
        }
      ],
      messages: [
        {
          code: 'detail-message',
          severity: 'warning',
          summary: 'Detail warning from job detail'
        }
      ]
    };
    this.detailedJobs = [
      firstDetail,
      {
        ...firstDetail,
        resources: [
          ...firstDetail.resources,
          {
            resourceType: 'asset',
            resourceId: 'asset-imported-after-refresh',
            sourceEntityType: 'asset',
            sourceEntityId: 'homebox-detail-refresh-asset',
            createdAt: '2026-07-06T12:04:30Z'
          }
        ],
        messages: [
          ...firstDetail.messages,
          {
            code: 'detail-refresh-message',
            severity: 'warning',
            summary: 'Detail warning after refresh'
          }
        ]
      }
    ];
  }

  async listImportJobs(tenantId: string, inventoryId: string): Promise<ImportJob[]> {
    this.expectScope(tenantId, inventoryId);
    return [{ ...this.job, resources: [], messages: [] }];
  }

  async getImportJob(tenantId: string, inventoryId: string, jobId: string): Promise<ImportJob> {
    this.detailCalls += 1;
    this.expectScope(tenantId, inventoryId, jobId);
    return this.detailedJobs[Math.min(this.detailCalls - 1, this.detailedJobs.length - 1)];
  }
}

export class DetailRefreshFailureImportJobRepository extends TerminalImportJobRepository {
  async getImportJob(tenantId: string, inventoryId: string, jobId: string): Promise<ImportJob> {
    this.expectScope(tenantId, inventoryId, jobId);
    throw new Error('provider-stacktrace password=secret');
  }
}

export class DiscardFailedImportJobRepository extends TerminalImportJobRepository {
  constructor(seedData: typeof seed) {
    super(seedData);
    this.job = {
      ...this.job,
      status: 'discard_failed',
      cancellationMode: 'discard_partial_progress',
      progress: {
        phase: 'terminal',
        done: 1,
        total: 1,
        message: 'Import cancellation cleanup failed',
        updatedAt: '2026-07-06T12:06:00Z'
      },
      messages: [
        {
          code: 'import-discard-failed',
          severity: 'error',
          summary: 'Import cancellation cleanup failed',
          detail: 'cleanup will retry'
        }
      ],
      updatedAt: '2026-07-06T12:06:00Z',
      completedAt: '2026-07-06T12:06:00Z'
    };
  }
}

export class CancellableImportJobRepository extends UnknownProgressImportJobRepository {
  cancellationModes: ImportJobCancellationMode[] = [];

  async cancelImportJob(
    tenantId: string,
    inventoryId: string,
    jobId: string,
    mode: ImportJobCancellationMode
  ): Promise<ImportJob> {
    expect(tenantId).toBe('tenant-home');
    expect(inventoryId).toBe('inventory-household');
    expect(jobId).toBe(this.job.id);
    this.cancellationModes.push(mode);
    this.job = {
      ...this.job,
      status: 'cancel_requested',
      cancellationMode: mode,
      progress: { ...this.job.progress, message: 'Cancellation requested' }
    };
    return this.job;
  }
}

export class MultiCancellableImportJobRepository extends CancellableImportJobRepository {
  private secondJob: ImportJob;

  constructor(seedData: typeof seed) {
    super(seedData);
    this.secondJob = {
      ...this.job,
      id: 'job-garage',
      source: {
        ...this.job.source,
        name: 'Garage Homebox',
        baseUrl: 'http://garage-homebox.local:7744',
        fingerprint: 'fingerprint-garage'
      },
      progress: { phase: 'attachments', done: 2, total: 5, message: 'Importing garage shelves', updatedAt: '2026-07-06T12:03:00Z' },
      progressHistory: [
        { phase: 'attachments', done: 2, total: 5, message: 'Importing garage shelves', updatedAt: '2026-07-06T12:03:00Z' }
      ]
    };
  }

  async listImportJobs(tenantId: string, inventoryId: string): Promise<ImportJob[]> {
    expect(tenantId).toBe('tenant-home');
    expect(inventoryId).toBe('inventory-household');
    return [this.job, this.secondJob];
  }

  async cancelImportJob(
    tenantId: string,
    inventoryId: string,
    jobId: string,
    mode: ImportJobCancellationMode
  ): Promise<ImportJob> {
    expect(tenantId).toBe('tenant-home');
    expect(inventoryId).toBe('inventory-household');
    this.cancellationModes.push(mode);
    if (jobId === this.job.id) {
      this.job = {
        ...this.job,
        status: 'cancel_requested',
        cancellationMode: mode,
        progress: { ...this.job.progress, message: 'Cancellation requested' }
      };
      return this.job;
    }
    expect(jobId).toBe(this.secondJob.id);
    this.secondJob = {
      ...this.secondJob,
      status: 'cancel_requested',
      cancellationMode: mode,
      progress: { ...this.secondJob.progress, message: 'Cancellation requested' }
    };
    return this.secondJob;
  }
}

export class CountingImportJobRepository extends SeededInventoryRepository {
  listImportJobCalls = 0;

  async listImportJobs(tenantId: string, inventoryId: string): Promise<ImportJob[]> {
    this.listImportJobCalls += 1;
    return super.listImportJobs(tenantId, inventoryId);
  }
}

export async function mountImportWorkspace(
  repository = new SeededInventoryRepository(structuredClone(seed)),
  options: {
    importSource?: 'homebox' | 'homebox-csv' | null;
    inventory?: WorkspaceSeed['inventories'][number] | null;
    currentPrincipal?: Principal;
    onImportJobInventoryChanged?: (scope: { tenantId: string; inventoryId: string }) => Promise<void>;
    onOpenImportedAssetId?: (assetId: string) => Promise<void>;
    onOpenInventoryAuditHistory?: () => void;
  } = {}
): Promise<SeededInventoryRepository> {
  component = mount(InventoryImportWorkspaceHarness, {
    target: document.body,
    props: {
      tenantId: 'tenant-home',
      inventory: options.inventory === undefined ? seed.inventories[0] : options.inventory,
      repository,
      initialImportSource: options.importSource ?? null,
      currentPrincipal: options.currentPrincipal,
      onImportJobInventoryChanged: options.onImportJobInventoryChanged,
      onOpenImportedAssetId: options.onOpenImportedAssetId,
      onOpenInventoryAuditHistory: options.onOpenInventoryAuditHistory
    }
  });
  return repository;
}

export async function waitFor(assertion: () => void): Promise<void> {
  let lastError: unknown;
  for (let attempt = 0; attempt < 30; attempt += 1) {
    await tick();
    await new Promise((resolve) => window.setTimeout(resolve, 0));
    try {
      assertion();
      return;
    } catch (caught) {
      lastError = caught;
    }
  }
  throw lastError;
}

export async function openLiveHomeboxSetup(): Promise<void> {
  buttonContaining('New import').click();
  await waitFor(() => {
    expect(document.body.textContent).toContain('Choose import method');
  });
  controlContaining('Connect to Homebox').click();

  await waitFor(() => {
    expect(document.body.querySelector<HTMLInputElement>('#homebox-url')).toBeTruthy();
  });
}

export function cleanupImportWorkspace(): void {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
  window.history.replaceState({}, '', '/');
}


export function setInputValue(input: HTMLInputElement, value: string): void {
  const setter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value')?.set;
  setter?.call(input, value);
  input.dispatchEvent(new Event('input', { bubbles: true }));
}

export function setFileInputFiles(input: HTMLInputElement, files: Array<FileLike>): void {
  Object.defineProperty(input, 'files', { value: files, writable: true, configurable: true });
  input.dispatchEvent(new Event('change', { bubbles: true }));
}

type FileLike = Pick<File, 'name' | 'size' | 'lastModified' | 'arrayBuffer'>;

export function fileLike(name: string, size: number, content: ArrayBuffer | Uint8Array | Promise<ArrayBuffer>): FileLike {
  return {
    name,
    size,
    lastModified: 1,
    arrayBuffer: async () => {
      const value = await content;
      if (!(value instanceof Uint8Array)) return value;
      const copy = new ArrayBuffer(value.byteLength);
      new Uint8Array(copy).set(value);
      return copy;
    }
  };
}

export function buttonContaining(text: string): HTMLButtonElement {
  const button = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!button) {
    throw new Error(`Missing button containing ${text}`);
  }
  return button;
}

export function exactButton(text: string): HTMLButtonElement {
  const button = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find(
    (candidate) => candidate.textContent?.trim() === text || candidate.getAttribute('aria-label') === text
  );
  if (!button) {
    throw new Error(`Missing button exactly matching ${text}`);
  }
  return button;
}

export function controlContaining(text: string): HTMLElement {
  const control = Array.from(document.body.querySelectorAll<HTMLElement>('button, a')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!control) {
    throw new Error(`Missing control containing ${text}`);
  }
  return control;
}

export function linkContaining(text: string): HTMLAnchorElement {
  const link = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) => candidate.textContent?.includes(text));
  if (!link) {
    throw new Error(`Missing link containing ${text}`);
  }
  return link;
}

export function detailsContaining(text: string): HTMLDetailsElement {
  const details = Array.from(document.body.querySelectorAll<HTMLDetailsElement>('details')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!details) {
    throw new Error(`Missing details containing ${text}`);
  }
  return details;
}

export function checkboxContaining(text: string): HTMLInputElement {
  const label = Array.from(document.body.querySelectorAll<HTMLLabelElement>('label')).find((candidate) => candidate.textContent?.includes(text));
  const checkbox = label?.querySelector<HTMLInputElement>('input[type="checkbox"]');
  if (!checkbox) {
    throw new Error(`Missing checkbox containing ${text}`);
  }
  return checkbox;
}
