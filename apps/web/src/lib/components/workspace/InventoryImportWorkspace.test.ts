import { tick } from 'svelte';
import { mount, unmount } from 'svelte';
import { afterEach, describe, expect, it } from 'vitest';
import { SeededInventoryRepository } from '$lib/adapters/memory/seededInventoryRepository';
import type {
  AssetLifecycleFilter,
  ImportJob,
  ImportJobCancellationMode,
  ImportSourceRequest,
  WorkspaceData
} from '$lib/domain/inventory';
import type { WorkspaceSeed } from '$lib/ports/inventoryRepository';
import InventoryImportWorkspaceHarness from './InventoryImportWorkspace.test-harness.svelte';

let component: ReturnType<typeof mount> | null = null;

const seed: WorkspaceSeed = {
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

class ImportPreviewRecordingRepository extends SeededInventoryRepository {
  previewInputs: ImportSourceRequest[] = [];

  async previewImportJob(tenantId: string, inventoryId: string, input: ImportSourceRequest): Promise<ImportJob> {
    this.previewInputs.push(structuredClone(input));
    return super.previewImportJob(tenantId, inventoryId, input);
  }
}

class ImportPreviewGenericInvalidRequestRepository extends SeededInventoryRepository {
  async previewImportJob(_tenantId: string, _inventoryId: string, _input: ImportSourceRequest): Promise<ImportJob> {
    throw Object.assign(new Error('Invalid request.'), { status: 400, code: 'invalid_request' });
  }
}

class ImportPreviewCountRepository extends ImportPreviewRecordingRepository {
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

class EmptyImportPreviewRepository extends ImportPreviewRecordingRepository {
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

class PreviewMessageOnlyRepository extends ImportPreviewRecordingRepository {
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

class PreviewHierarchyRepository extends ImportPreviewRecordingRepository {
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

class TerminalImportJobRepository extends SeededInventoryRepository {
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
        fields: [],
        locations: [],
        assets: [],
        attachments: [],
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

  private expectScope(tenantId: string, inventoryId: string, jobId?: string): void {
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

class PreviewedImportJobRepository extends TerminalImportJobRepository {
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

class PreviewedCSVImportJobRepository extends PreviewedImportJobRepository {
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

class CompletingImportJobRepository extends TerminalImportJobRepository {
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

class UnknownProgressImportJobRepository extends TerminalImportJobRepository {
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

class TerminalUnknownProgressImportJobRepository extends TerminalImportJobRepository {
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

class DiscardedImportJobRepository extends TerminalImportJobRepository {
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

class ResourcefulImportJobRepository extends TerminalImportJobRepository {
  constructor(seedData: typeof seed) {
    super(seedData);
    this.job = {
      ...this.job,
      resources: [
        {
          resourceType: 'asset',
          resourceId: 'asset-imported-passport',
          sourceEntityType: 'asset',
          sourceEntityId: 'homebox-item-passport',
          createdAt: '2026-07-06T12:03:00Z'
        },
        {
          resourceType: 'attachment',
          resourceId: 'attachment-passport-photo',
          resourceOwnerId: 'asset-imported-passport',
          sourceEntityType: 'attachment',
          sourceEntityId: 'homebox-photo-passport',
          createdAt: '2026-07-06T12:04:00Z'
        }
      ]
    };
  }
}

class DiscardFailedImportJobRepository extends TerminalImportJobRepository {
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

class CancellableImportJobRepository extends UnknownProgressImportJobRepository {
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

class MultiCancellableImportJobRepository extends CancellableImportJobRepository {
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

class CountingImportJobRepository extends SeededInventoryRepository {
  listImportJobCalls = 0;

  async listImportJobs(tenantId: string, inventoryId: string): Promise<ImportJob[]> {
    this.listImportJobCalls += 1;
    return super.listImportJobs(tenantId, inventoryId);
  }
}

async function mountImportWorkspace(
  repository = new SeededInventoryRepository(structuredClone(seed)),
  options: {
    importSource?: 'homebox' | 'homebox-csv' | null;
    inventory?: WorkspaceSeed['inventories'][number] | null;
    onImportJobInventoryChanged?: (scope: { tenantId: string; inventoryId: string }) => Promise<void>;
  } = {}
): Promise<SeededInventoryRepository> {
  component = mount(InventoryImportWorkspaceHarness, {
    target: document.body,
    props: {
      tenantId: 'tenant-home',
      inventory: options.inventory === undefined ? seed.inventories[0] : options.inventory,
      repository,
      initialImportSource: options.importSource ?? null,
      onImportJobInventoryChanged: options.onImportJobInventoryChanged
    }
  });
  return repository;
}

async function waitFor(assertion: () => void): Promise<void> {
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

async function openLiveHomeboxSetup(): Promise<void> {
  buttonContaining('New import').click();
  await waitFor(() => {
    expect(document.body.textContent).toContain('Choose import method');
  });
  controlContaining('Connect to Homebox').click();

  await waitFor(() => {
    expect(document.body.querySelector<HTMLInputElement>('#homebox-url')).toBeTruthy();
  });
}

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
  window.history.replaceState({}, '', '/');
});

describe('InventoryImportWorkspace', () => {
  it('shows a clear unavailable state and does not load jobs without import view access', async () => {
    const repository = new CountingImportJobRepository(structuredClone(seed));
    const viewerInventory = {
      ...seed.inventories[0],
      access: { relationship: 'viewer', permissions: ['view'] }
    };

    await mountImportWorkspace(repository, { inventory: viewerInventory });

    await waitFor(() => {
      expect(document.body.textContent).toContain('Import access needed');
      expect(document.body.textContent).toContain('importing records requires import job access');
    });
    expect(document.body.textContent).not.toContain('Import history');
    expect(repository.listImportJobCalls).toBe(0);
  });

  it('confirms live Homebox sources with https as the schemeless URL default', async () => {
    const repository = new ImportPreviewRecordingRepository(structuredClone(seed));
    await mountImportWorkspace(repository);

    await waitFor(() => {
      expect(document.body.textContent).toContain('No import runs yet');
    });

    buttonContaining('New import').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Choose import method');
    });

    controlContaining('Connect to Homebox').click();

    await waitFor(() => {
      expect(document.body.querySelector<HTMLInputElement>('#homebox-url')).toBeTruthy();
    });
    expect(document.body.querySelector('[aria-current="step"]')?.textContent).toContain('Connect');
    expect(document.body.textContent).not.toContain('Step 2 of 4');
    expect(document.body.querySelector<HTMLInputElement>('#homebox-url')?.getAttribute('autocapitalize')).toBe('none');
    expect(document.body.querySelector<HTMLInputElement>('#homebox-url')?.getAttribute('autocorrect')).toBe('off');
    expect(document.body.querySelector<HTMLInputElement>('#homebox-url')?.getAttribute('spellcheck')).toBe('false');
    expect(document.body.querySelector<HTMLInputElement>('#homebox-user')?.getAttribute('autocapitalize')).toBe('none');
    expect(document.body.querySelector<HTMLInputElement>('#homebox-user')?.getAttribute('autocorrect')).toBe('off');
    expect(document.body.querySelector<HTMLInputElement>('#homebox-user')?.getAttribute('spellcheck')).toBe('false');
    expect(document.body.querySelector<HTMLInputElement>('#homebox-password')?.getAttribute('autocapitalize')).toBe('none');
    expect(document.body.querySelector<HTMLInputElement>('#homebox-password')?.getAttribute('autocorrect')).toBe('off');
    expect(document.body.querySelector<HTMLInputElement>('#homebox-password')?.getAttribute('spellcheck')).toBe('false');

    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-url')!, 'stuff.jsksell.com');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-user')!, 'codex@jsksell.com');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-password')!, 'asldfj3290f!');

    await waitFor(() => {
      expect(buttonContaining('Confirm connection').disabled).toBe(false);
    });

    buttonContaining('Confirm connection').click();

    await waitFor(() => {
      expect(repository.previewInputs).toHaveLength(1);
      expect(document.body.textContent).toContain('Preview import');
    });

    expect(repository.previewInputs[0]).toMatchObject({
      sourceType: 'legacy_homebox',
      baseUrl: 'https://stuff.jsksell.com',
      username: 'codex@jsksell.com',
      includeImages: true,
      allowPrivateNetwork: false,
      allowInsecureTLS: false
    });
  });

  it('keeps risky live Homebox connection options visually subordinate', async () => {
    await mountImportWorkspace(new SeededInventoryRepository(structuredClone(seed)));

    await openLiveHomeboxSetup();

    const advanced = detailsContaining('Connection options');
    expect(advanced.open).toBe(false);
    expect(advanced.textContent).toContain('Allow private-network Homebox URL');
    expect(advanced.textContent).toContain('Allow self-signed TLS certificate');

    advanced.open = true;
    advanced.dispatchEvent(new Event('toggle'));

    await waitFor(() => {
      expect(checkboxContaining('Allow private-network Homebox URL')).toBeTruthy();
      expect(checkboxContaining('Allow self-signed TLS certificate')).toBeTruthy();
    });
  });

  it('rejects oversized Homebox CSV files before previewing', async () => {
    const repository = new ImportPreviewRecordingRepository(structuredClone(seed));
    await mountImportWorkspace(repository);

    buttonContaining('New import').click();
    await waitFor(() => {
      expect(document.body.textContent).toContain('Choose import method');
    });
    controlContaining('Upload Homebox CSV').click();

    await waitFor(() => {
      expect(document.body.querySelector<HTMLInputElement>('#homebox-csv')).toBeTruthy();
    });

    setFileInputFiles(document.body.querySelector<HTMLInputElement>('#homebox-csv')!, [
      fileLike('homebox.csv', 10 * 1024 * 1024 + 1, new Uint8Array())
    ]);

    await waitFor(() => {
      expect(document.body.textContent).toContain('CSV is too large');
    });

    expect(repository.previewInputs).toHaveLength(0);
  });

  it('does not let a stale CSV read re-enable preview after an oversized file is selected', async () => {
    const repository = new ImportPreviewRecordingRepository(structuredClone(seed));
    let resolveFirstRead: ((value: ArrayBuffer) => void) | undefined;
    await mountImportWorkspace(repository);

    buttonContaining('New import').click();
    await waitFor(() => {
      expect(document.body.textContent).toContain('Choose import method');
    });
    controlContaining('Upload Homebox CSV').click();

    await waitFor(() => {
      expect(document.body.querySelector<HTMLInputElement>('#homebox-csv')).toBeTruthy();
    });

    const fileInput = document.body.querySelector<HTMLInputElement>('#homebox-csv')!;
    setFileInputFiles(fileInput, [
      fileLike(
        'first.csv',
        100,
        new Promise((resolve) => {
          resolveFirstRead = resolve;
        })
      )
    ]);
    setFileInputFiles(fileInput, [fileLike('too-big.csv', 10 * 1024 * 1024 + 1, new Uint8Array())]);
    resolveFirstRead?.(new TextEncoder().encode('name\nstale').buffer);

    await waitFor(() => {
      expect(document.body.textContent).toContain('CSV is too large');
    });

    expect(buttonContaining('Prepare preview').disabled).toBe(true);
    expect(repository.previewInputs).toHaveLength(0);
  });

  it('allows a Homebox CSV at the 10 MiB limit', async () => {
    const repository = new ImportPreviewRecordingRepository(structuredClone(seed));
    await mountImportWorkspace(repository);

    buttonContaining('New import').click();
    await waitFor(() => {
      expect(document.body.textContent).toContain('Choose import method');
    });
    controlContaining('Upload Homebox CSV').click();

    await waitFor(() => {
      expect(document.body.querySelector<HTMLInputElement>('#homebox-csv')).toBeTruthy();
    });

    setFileInputFiles(document.body.querySelector<HTMLInputElement>('#homebox-csv')!, [
      fileLike('homebox.csv', 10 * 1024 * 1024, new TextEncoder().encode('name\nok').buffer)
    ]);

    await waitFor(() => {
      expect(buttonContaining('Prepare preview').disabled).toBe(false);
    });

    buttonContaining('Prepare preview').click();

    await waitFor(() => {
      expect(repository.previewInputs).toHaveLength(1);
    });
  });

  it('shows contextual copy for generic Homebox preview validation failures', async () => {
    await mountImportWorkspace(new ImportPreviewGenericInvalidRequestRepository(structuredClone(seed)));

    buttonContaining('New import').click();
    await waitFor(() => {
      expect(document.body.textContent).toContain('Choose import method');
    });
    controlContaining('Connect to Homebox').click();

    await waitFor(() => {
      expect(document.body.querySelector<HTMLInputElement>('#homebox-url')).toBeTruthy();
    });

    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-url')!, 'stuff.jsksell.com');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-user')!, 'codex@jsksell.com');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-password')!, 'asldfj3290f!');

    await waitFor(() => {
      expect(buttonContaining('Confirm connection').disabled).toBe(false);
    });

    buttonContaining('Confirm connection').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Homebox connection could not be confirmed.');
      expect(document.body.textContent).not.toContain('Invalid request.');
    });
  });

  it('shows preview duplicate, skipped, warning, and blocking counts before start', async () => {
    await mountImportWorkspace(new ImportPreviewCountRepository(structuredClone(seed)));

    await openLiveHomeboxSetup();
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-url')!, 'http://homebox.local:7744');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-user')!, 'codex@jsksell.com');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-password')!, 'asldfj3290f!');

    await waitFor(() => {
      expect(buttonContaining('Confirm connection').disabled).toBe(false);
    });

    buttonContaining('Confirm connection').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Preview import');
      expect(document.body.textContent).not.toContain('Step 3 of 4');
      expect(document.body.textContent).not.toContain('3 Preview');
      expect(document.body.querySelector('[aria-current="step"]')?.textContent).toContain('Preview');
      expect(document.body.textContent).toContain('Fix blocking issues before importing');
      expect(document.body.textContent).toContain('Nothing has been saved.');
      expect(document.body.textContent).not.toContain('Nothing saved');
      expect(document.body.textContent).toContain('4 duplicates/skips');
      expect(document.body.textContent).toContain('2 warnings');
      expect(document.body.textContent).toContain('1 blocking issue');
      expect(buttonContaining('Start background import').disabled).toBe(true);
    });
  });

  it('keeps planned preview count categories visible when nothing will be imported', async () => {
    await mountImportWorkspace(new EmptyImportPreviewRepository(structuredClone(seed)));

    await openLiveHomeboxSetup();
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-url')!, 'http://homebox.local:7744');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-user')!, 'codex@jsksell.com');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-password')!, 'asldfj3290f!');

    await waitFor(() => {
      expect(buttonContaining('Confirm connection').disabled).toBe(false);
    });

    buttonContaining('Confirm connection').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Ready to start');
      expect(document.body.textContent).toContain('Nothing has been saved. Start the import when this plan looks right.');
      expect(document.body.textContent).not.toContain('Nothing saved');
      expect(document.body.textContent).toContain('0 fields');
      expect(document.body.textContent).toContain('0 locations');
      expect(document.body.textContent).toContain('0 assets');
      expect(document.body.textContent).toContain('0 photos/files');
      expect(document.body.textContent).toContain('0 blocking issues');
    });
  });

  it('does not show an empty message state when preview-specific messages are present', async () => {
    await mountImportWorkspace(new PreviewMessageOnlyRepository(structuredClone(seed)));

    await openLiveHomeboxSetup();
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-url')!, 'http://homebox.local:7744');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-user')!, 'codex@jsksell.com');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-password')!, 'asldfj3290f!');

    await waitFor(() => {
      expect(buttonContaining('Confirm connection').disabled).toBe(false);
    });

    buttonContaining('Confirm connection').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Attachment will be skipped');
      expect(document.body.textContent).not.toContain('No blocking issues found.');
    });
  });

  it('keeps preview hierarchy user-facing instead of showing raw parent source IDs', async () => {
    await mountImportWorkspace(new PreviewHierarchyRepository(structuredClone(seed)));

    await openLiveHomeboxSetup();
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-url')!, 'http://homebox.local:7744');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-user')!, 'codex@jsksell.com');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-password')!, 'asldfj3290f!');

    await waitFor(() => {
      expect(buttonContaining('Confirm connection').disabled).toBe(false);
    });

    buttonContaining('Confirm connection').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Garage');
      expect(document.body.textContent).toContain('Shelf');
      expect(document.body.textContent).toContain('inside another imported record');
      expect(document.body.textContent).not.toContain('inside loc-garage');
      expect(document.body.textContent).not.toContain('inside loc-shelf');
    });
  });

  it('returns to import history after starting a durable import', async () => {
    const repository = new ImportPreviewRecordingRepository(structuredClone(seed));
    await mountImportWorkspace(repository);

    buttonContaining('New import').click();
    await waitFor(() => {
      expect(document.body.textContent).toContain('Choose import method');
    });
    controlContaining('Connect to Homebox').click();

    await waitFor(() => {
      expect(document.body.querySelector<HTMLInputElement>('#homebox-url')).toBeTruthy();
    });

    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-url')!, 'http://homebox.local:7744');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-user')!, 'codex@jsksell.com');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-password')!, 'asldfj3290f!');

    await waitFor(() => {
      expect(buttonContaining('Confirm connection').disabled).toBe(false);
    });

    buttonContaining('Confirm connection').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Preview import');
      expect(document.body.textContent).toContain('Homebox source checked for this preview.');
    });

    expect(repository.previewInputs[0]).toMatchObject({
      sourceType: 'legacy_homebox',
      baseUrl: 'http://homebox.local:7744',
      username: 'codex@jsksell.com'
    });

    buttonContaining('Start background import').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Import history');
      expect(document.body.textContent).toContain('Current work');
      expect(document.body.textContent).not.toContain('Import could not be started.');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Progress timeline');
      expect(document.body.textContent).toContain('reading source');
    });
  });

  it('requires a fresh preview after live Homebox connection details change', async () => {
    const repository = new ImportPreviewRecordingRepository(structuredClone(seed));
    await mountImportWorkspace(repository);

    await openLiveHomeboxSetup();
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-url')!, 'http://homebox.local:7744');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-user')!, 'codex@jsksell.com');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-password')!, 'asldfj3290f!');

    await waitFor(() => {
      expect(buttonContaining('Confirm connection').disabled).toBe(false);
    });

    buttonContaining('Confirm connection').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Preview import');
      expect(buttonContaining('Start background import').disabled).toBe(false);
    });

    exactButton('Back').click();
    await waitFor(() => {
      expect(document.body.textContent).toContain('Connect to Homebox');
    });
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-url')!, 'http://homebox.changed.local:7744');

    await waitFor(() => {
      expect(document.body.textContent).toContain('Connect to Homebox');
      expect(buttonContaining('Confirm connection').disabled).toBe(false);
      expect(document.body.textContent).not.toContain('Start background import');
    });

    buttonContaining('Confirm connection').click();

    await waitFor(() => {
      expect(repository.previewInputs).toHaveLength(2);
      expect(document.body.textContent).toContain('Preview import');
      expect(buttonContaining('Start background import').disabled).toBe(false);
    });
    expect(repository.previewInputs[1]).toMatchObject({
      baseUrl: 'http://homebox.changed.local:7744'
    });
  });

  it('requires a fresh preview after live Homebox image options change', async () => {
    const repository = new ImportPreviewRecordingRepository(structuredClone(seed));
    await mountImportWorkspace(repository);

    await openLiveHomeboxSetup();
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-url')!, 'http://homebox.local:7744');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-user')!, 'codex@jsksell.com');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-password')!, 'asldfj3290f!');

    await waitFor(() => {
      expect(buttonContaining('Confirm connection').disabled).toBe(false);
    });

    buttonContaining('Confirm connection').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Preview import');
      expect(buttonContaining('Start background import').disabled).toBe(false);
    });

    exactButton('Back').click();
    await waitFor(() => {
      expect(document.body.textContent).toContain('Connect to Homebox');
    });
    checkboxContaining('Import photos when Homebox provides them').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Connect to Homebox');
      expect(buttonContaining('Confirm connection').disabled).toBe(false);
      expect(document.body.textContent).not.toContain('Start background import');
    });

    buttonContaining('Confirm connection').click();

    await waitFor(() => {
      expect(repository.previewInputs).toHaveLength(2);
      expect(buttonContaining('Start background import').disabled).toBe(false);
    });
    expect(repository.previewInputs[1]).toMatchObject({
      includeImages: false
    });
  });

  it('requires a fresh preview after the selected Homebox CSV changes', async () => {
    const repository = new ImportPreviewRecordingRepository(structuredClone(seed));
    await mountImportWorkspace(repository);

    buttonContaining('New import').click();
    await waitFor(() => {
      expect(document.body.textContent).toContain('Choose import method');
    });
    controlContaining('Upload Homebox CSV').click();

    await waitFor(() => {
      expect(document.body.querySelector<HTMLInputElement>('#homebox-csv')).toBeTruthy();
    });

    setFileInputFiles(document.body.querySelector<HTMLInputElement>('#homebox-csv')!, [
      fileLike('homebox.csv', 100, new TextEncoder().encode('name\nfirst').buffer)
    ]);

    await waitFor(() => {
      expect(buttonContaining('Prepare preview').disabled).toBe(false);
    });

    buttonContaining('Prepare preview').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Preview import');
      expect(buttonContaining('Start background import').disabled).toBe(false);
    });

    exactButton('Back').click();
    await waitFor(() => {
      expect(document.body.querySelector<HTMLInputElement>('#homebox-csv')).toBeTruthy();
    });
    setFileInputFiles(document.body.querySelector<HTMLInputElement>('#homebox-csv')!, [
      fileLike('homebox-updated.csv', 100, new TextEncoder().encode('name\nsecond').buffer)
    ]);

    await waitFor(() => {
      expect(document.body.textContent).toContain('homebox-updated.csv');
      expect(document.body.textContent).not.toContain('Preview import');
      expect(buttonContaining('Prepare preview').disabled).toBe(false);
    });

    expect(repository.previewInputs).toHaveLength(1);
  });

  it('refreshes workspace data when a running import job finishes', async () => {
    const repository = new CompletingImportJobRepository(structuredClone(seed));
    let refreshes = 0;
    await mountImportWorkspace(repository, {
      onImportJobInventoryChanged: async (scope) => {
        expect(scope).toEqual({ tenantId: 'tenant-home', inventoryId: 'inventory-household' });
        refreshes += 1;
      }
    });

    await waitFor(() => {
      expect(document.body.textContent).toContain('Current work');
      expect(document.body.textContent).toContain('Creating assets');
    });

    buttonContaining('Refresh').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Import finished. Workspace data has been refreshed.');
      expect(document.body.textContent).toContain('History');
      expect(document.body.textContent).not.toContain('Current work');
      expect(refreshes).toBe(1);
    });
  });

  it('shows unknown-total import phases without fake exact progress', async () => {
    await mountImportWorkspace(new UnknownProgressImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Current work');
      expect(document.body.textContent).toContain('Reading source');
      expect(document.body.textContent).toContain('total not known yet');
    });

    const progressbar = document.body.querySelector<HTMLElement>('[role="progressbar"]');
    expect(progressbar).toBeTruthy();
    expect(progressbar?.classList.contains('indeterminate')).toBe(true);
    expect(progressbar?.getAttribute('aria-label')).toContain('total not known yet');
    expect(progressbar?.hasAttribute('aria-valuenow')).toBe(false);
    expect(progressbar?.querySelector('span')?.getAttribute('style') ?? '').not.toContain('width: 0%');
  });

  it('does not turn completed unknown-total imports into fake exact progress', async () => {
    await mountImportWorkspace(new TerminalUnknownProgressImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('History');
      expect(document.body.textContent).toContain('Completed');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Progress timeline');
      expect(document.body.textContent).toContain('Import completed');
    });

    expect(document.body.querySelector('[role="progressbar"]')).toBeFalsy();
  });

  it('uses explicit import cancellation choices and submits the selected mode', async () => {
    const repository = new CancellableImportJobRepository(structuredClone(seed));
    await mountImportWorkspace(repository);

    await waitFor(() => {
      expect(document.body.textContent).toContain('Current work');
      expect(exactButton('Cancel')).toBeTruthy();
    });

    exactButton('Cancel').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Cancel Homebox?');
      expect(buttonContaining('Keep imported items')).toBeTruthy();
      expect(buttonContaining('Discard imported items')).toBeTruthy();
      expect(document.body.textContent).toContain('Stop future work and leave anything already imported in the inventory.');
    });

    buttonContaining('Keep imported items').click();

    await waitFor(() => {
      expect(repository.cancellationModes).toEqual(['keep_partial_progress']);
    });

    exactButton('Cancel').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Cancel Homebox?');
      expect(document.body.textContent).toContain('Stop future work and remove records created by this job. Audit history remains.');
    });

    buttonContaining('Discard imported items').click();

    await waitFor(() => {
      expect(repository.cancellationModes).toEqual(['keep_partial_progress', 'discard_partial_progress']);
    });
  });

  it('moves focus into the cancellation choices and names the selected job', async () => {
    await mountImportWorkspace(new MultiCancellableImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Garage Homebox');
    });

    const garageJobCard = Array.from(document.body.querySelectorAll<HTMLElement>('.job-card')).find((candidate) =>
      candidate.textContent?.includes('Garage Homebox')
    );
    expect(garageJobCard).toBeTruthy();
    const cancelButton = Array.from(garageJobCard?.querySelectorAll<HTMLButtonElement>('button') ?? []).find(
      (candidate) => candidate.textContent?.trim() === 'Cancel'
    );
    expect(cancelButton).toBeTruthy();
    cancelButton?.click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Cancel Garage Homebox?');
      expect(document.body.textContent).toContain('http://garage-homebox.local:7744');
      expect(document.body.textContent).toContain('Importing garage shelves');
      expect(document.activeElement?.textContent).toContain('Keep imported items');
    });
  });

  it('returns to import history after removing the selected terminal import job', async () => {
    await mountImportWorkspace(new TerminalImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('History');
      expect(document.body.textContent).toContain('Homebox');
      expect(document.body.textContent).toContain('Prepared by owner');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Progress timeline');
      expect(document.body.textContent).toContain('Remove from history');
      expect(document.body.textContent).toContain('1 field reused');
      expect(document.body.textContent).toContain('1 location created');
      expect(document.body.textContent).toContain('2 assets skipped');
      expect(document.body.textContent).toContain('1 photo/file skipped');
      expect(document.body.textContent).toContain('2 warnings');
    });
    expect(linkContaining('View audit history').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/settings/activity'
    );

    buttonContaining('Remove from history').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Remove Homebox from history?');
      expect(document.body.textContent).toContain('Imported records and audit history will remain.');
    });

    buttonContaining('Remove from history').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Import history');
      expect(document.body.textContent).not.toContain('Progress timeline');
      expect(document.body.textContent).not.toContain('Remove from history');
    });
  });

  it('summarizes terminal import history as a scannable job ledger', async () => {
    await mountImportWorkspace(new TerminalImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('History');
      expect(document.body.textContent).toContain('Running');
      expect(document.body.textContent).toContain('Ready to review');
      expect(document.body.textContent).toContain('Needs attention');
      expect(document.body.textContent).toContain('1 import needs attention');
      expect(document.body.textContent).toContain('Completed with warnings.');
      expect(document.body.textContent).toContain('Completed');
      expect(document.body.textContent).toContain('Homebox');
      expect(document.body.textContent).toContain('Prepared by owner');
      expect(document.body.textContent).toContain('Started Jul 6, 2026');
      expect(document.body.textContent).toContain('Completed Jul 6, 2026');
      expect(document.body.textContent).toContain('1 asset created');
    });
  });

  it('does not expose discarded import resources as openable records', async () => {
    await mountImportWorkspace(new DiscardedImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('History');
      expect(document.body.textContent).toContain('Discarded');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Records created by this job were discarded.');
      expect(document.body.textContent).not.toContain('Imported records');
    });
    expect(document.body.querySelector<HTMLAnchorElement>('a.resource-link')).toBeFalsy();
  });

  it('confirms history-row removal before hiding a terminal import job', async () => {
    await mountImportWorkspace(new TerminalImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('History');
      expect(document.body.textContent).toContain('Homebox');
    });

    document.body.querySelector<HTMLButtonElement>('button[aria-label="Remove import job from history"]')?.click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Remove Homebox from history?');
      expect(document.body.textContent).toContain('This only removes the run from the import history list.');
      expect(document.body.textContent).toContain('Keep in history');
    });

    buttonContaining('Keep in history').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('History');
      expect(document.body.textContent).toContain('Homebox');
      expect(document.body.textContent).not.toContain('Remove Homebox from history?');
    });

    document.body.querySelector<HTMLButtonElement>('button[aria-label="Remove import job from history"]')?.click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Remove Homebox from history?');
    });

    buttonContaining('Remove from history').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('No import runs yet');
      expect(document.body.textContent).not.toContain('Remove Homebox from history?');
    });
  });

  it('uses user-facing imported record labels with source IDs as secondary metadata', async () => {
    await mountImportWorkspace(new ResourcefulImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('History');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Imported records');
      expect(document.body.textContent).toContain('Imported asset');
      expect(document.body.textContent).toContain('Imported photo/file');
      expect(document.body.textContent).toContain('Source asset: homebox-item-passport');
      expect(document.body.textContent).toContain('Source attachment: homebox-photo-passport');
      expect(document.body.textContent).not.toContain('Asset · asset-imported-passport');
      expect(document.body.textContent).not.toContain('Photo/file · attachment-passport-photo');
    });
  });

  it('keeps discard-failed jobs visible without a remove-from-history action', async () => {
    await mountImportWorkspace(new DiscardFailedImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Discard failed');
    });
    expect(document.body.querySelector<HTMLButtonElement>('button[aria-label="Remove import job from history"]')).toBeFalsy();

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('cleanup will retry');
    });
    expect(document.body.textContent).not.toContain('Remove from history');
  });

  it('labels previewed import job actors as prepared rather than started', async () => {
    await mountImportWorkspace(new PreviewedImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Ready to review');
      expect(document.body.textContent).toContain('Ready for your review.');
      expect(document.body.textContent).toContain('Prepared by owner');
      expect(buttonContaining('Continue')).toBeTruthy();
      expect(document.body.textContent).not.toContain('Started by owner');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Continue import');
    });

    buttonContaining('Continue import').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Connect to Homebox');
      expect(document.body.textContent).toContain('Confirm the source again to continue this import.');
      expect(document.body.querySelector('[aria-current="step"]')?.textContent).toContain('Connect');
      expect(document.body.textContent).not.toContain('Step 2 of 4');
      expect(document.body.querySelector<HTMLInputElement>('#homebox-user')?.value).toBe('');
      expect(document.body.querySelector<HTMLInputElement>('#homebox-password')?.value).toBe('');
      expect(buttonContaining('Confirm connection').disabled).toBe(true);
    });
  });

  it('does not reuse stale live Homebox credentials when resuming a previewed job', async () => {
    await mountImportWorkspace(new PreviewedImportJobRepository(structuredClone(seed)));

    exactButton('New import').click();
    await waitFor(() => {
      expect(document.body.textContent).toContain('Choose import method');
    });
    controlContaining('Connect to Homebox').click();

    await waitFor(() => {
      expect(document.body.querySelector<HTMLInputElement>('#homebox-url')).toBeTruthy();
    });
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-url')!, 'http://stale-homebox.local:7744');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-user')!, 'stale@example.test');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-password')!, 'stale-password');

    buttonContaining('Back to history').click();
    await waitFor(() => {
      expect(document.body.textContent).toContain('Current work');
    });
    buttonContaining('Continue').click();

    await waitFor(() => {
      expect(document.body.querySelector<HTMLInputElement>('#homebox-url')?.value).toBe('http://homebox.local:7744');
      expect(document.body.querySelector<HTMLInputElement>('#homebox-user')?.value).toBe('');
      expect(document.body.querySelector<HTMLInputElement>('#homebox-password')?.value).toBe('');
      expect(buttonContaining('Confirm connection').disabled).toBe(true);
    });
  });

  it('does not reuse stale CSV contents when resuming a previewed CSV job', async () => {
    await mountImportWorkspace(new PreviewedCSVImportJobRepository(structuredClone(seed)));

    exactButton('New import').click();
    await waitFor(() => {
      expect(document.body.textContent).toContain('Choose import method');
    });
    controlContaining('Upload Homebox CSV').click();

    await waitFor(() => {
      expect(document.body.querySelector<HTMLInputElement>('#homebox-csv')).toBeTruthy();
    });
    setFileInputFiles(document.body.querySelector<HTMLInputElement>('#homebox-csv')!, [
      fileLike('stale-homebox.csv', 128, new TextEncoder().encode('HB.name\nDrill\n'))
    ]);

    await waitFor(() => {
      expect(document.body.textContent).toContain('stale-homebox.csv');
    });

    buttonContaining('Back to history').click();
    await waitFor(() => {
      expect(document.body.textContent).toContain('Current work');
    });
    buttonContaining('Continue').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Upload Homebox CSV');
      expect(document.body.textContent).not.toContain('stale-homebox.csv');
      expect(buttonContaining('Prepare preview').disabled).toBe(true);
    });
  });

  it('deep-links to source setup routes and exposes source choice hrefs', async () => {
    await mountImportWorkspace(new SeededInventoryRepository(structuredClone(seed)), { importSource: 'homebox-csv' });

    await waitFor(() => {
      expect(document.body.textContent).toContain('Upload Homebox CSV');
      expect(document.body.querySelector<HTMLInputElement>('#homebox-csv')).toBeTruthy();
    });

    buttonContaining('Back').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Import history');
    });

    buttonContaining('New import').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Choose import method');
      expect(document.body.textContent).toContain('Can include photos');
      expect(document.body.textContent).toContain('No photos in CSV');
    });

    expect(controlContaining('Connect to Homebox').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/import/homebox'
    );
    expect(controlContaining('Upload Homebox CSV').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/import/homebox-csv'
    );
  });
});

function setInputValue(input: HTMLInputElement, value: string): void {
  const setter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value')?.set;
  setter?.call(input, value);
  input.dispatchEvent(new Event('input', { bubbles: true }));
}

function setFileInputFiles(input: HTMLInputElement, files: Array<FileLike>): void {
  Object.defineProperty(input, 'files', { value: files, writable: true, configurable: true });
  input.dispatchEvent(new Event('change', { bubbles: true }));
}

type FileLike = Pick<File, 'name' | 'size' | 'lastModified' | 'arrayBuffer'>;

function fileLike(name: string, size: number, content: ArrayBuffer | Uint8Array | Promise<ArrayBuffer>): FileLike {
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

function buttonContaining(text: string): HTMLButtonElement {
  const button = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!button) {
    throw new Error(`Missing button containing ${text}`);
  }
  return button;
}

function exactButton(text: string): HTMLButtonElement {
  const button = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find(
    (candidate) => candidate.textContent?.trim() === text
  );
  if (!button) {
    throw new Error(`Missing button exactly matching ${text}`);
  }
  return button;
}

function controlContaining(text: string): HTMLElement {
  const control = Array.from(document.body.querySelectorAll<HTMLElement>('button, a')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!control) {
    throw new Error(`Missing control containing ${text}`);
  }
  return control;
}

function linkContaining(text: string): HTMLAnchorElement {
  const link = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) => candidate.textContent?.includes(text));
  if (!link) {
    throw new Error(`Missing link containing ${text}`);
  }
  return link;
}

function detailsContaining(text: string): HTMLDetailsElement {
  const details = Array.from(document.body.querySelectorAll<HTMLDetailsElement>('details')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!details) {
    throw new Error(`Missing details containing ${text}`);
  }
  return details;
}

function checkboxContaining(text: string): HTMLInputElement {
  const label = Array.from(document.body.querySelectorAll<HTMLLabelElement>('label')).find((candidate) => candidate.textContent?.includes(text));
  const checkbox = label?.querySelector<HTMLInputElement>('input[type="checkbox"]');
  if (!checkbox) {
    throw new Error(`Missing checkbox containing ${text}`);
  }
  return checkbox;
}
