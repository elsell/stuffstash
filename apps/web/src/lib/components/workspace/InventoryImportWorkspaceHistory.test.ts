import { afterEach, describe, expect, it } from 'vitest';
import { SeededInventoryRepository } from '$lib/adapters/memory/seededInventoryRepository';
import type { ImportJob, ImportJobCancellationMode } from '$lib/domain/inventory';
import importJobDetailPanelSource from './ImportJobDetailPanel.svelte?raw';
import importJobHistorySource from './ImportJobHistory.svelte?raw';
import {
  CancellableImportJobRepository,
  CompletingImportJobRepository,
  DetailOnlyResourcefulImportJobRepository,
  DetailRefreshFailureImportJobRepository,
  DiscardFailedImportJobRepository,
  DiscardedImportJobRepository,
  LongActorImportJobRepository,
  MultiCancellableImportJobRepository,
  PreviewedCSVImportJobRepository,
  PreviewedImportJobRepository,
  ResourcefulImportJobRepository,
  TerminalJobAndPreviewMessageImportJobRepository,
  TerminalImportJobRepository,
  TerminalPreviewMessageImportJobRepository,
  TerminalUnknownProgressImportJobRepository,
  UnknownProgressImportJobRepository,
  buttonContaining,
  cleanupImportWorkspace,
  controlContaining,
  exactButton,
  fileLike,
  linkContaining,
  mountImportWorkspace,
  seed,
  setFileInputFiles,
  setInputValue,
  waitFor
} from './InventoryImportWorkspace.test-helpers';

afterEach(() => {
  cleanupImportWorkspace();
});

describe('InventoryImportWorkspace import history and progress', () => {
  it('keeps import warning styles on semantic warning tokens', () => {
    expect(importJobHistorySource).toContain('var(--color-warning)');
    expect(importJobHistorySource).toContain('var(--color-warning-foreground)');
    expect(importJobHistorySource).not.toContain('var(--color-warning,');
    expect(importJobHistorySource).not.toContain('var(--color-warning-foreground,');

    expect(importJobDetailPanelSource).toContain('var(--color-warning)');
    expect(importJobDetailPanelSource).toContain('var(--color-warning-foreground)');
    expect(importJobDetailPanelSource).not.toContain('var(--color-warning,');
    expect(importJobDetailPanelSource).not.toContain('var(--color-warning-foreground,');
    expect(importJobHistorySource).toContain('.attention-alert');
    expect(importJobHistorySource).toContain('var(--destructive)');
  });

  it('keeps import detail tabs as a scrollable mobile rail instead of clipped equal columns', () => {
    const tabListRule = importJobDetailPanelSource.match(/:global\(\.detail-tab-list\)\s*{(?<body>[^}]*)}/)?.groups?.body ?? '';
    const tabTriggerRule =
      importJobDetailPanelSource.match(/:global\(\.detail-tab-list \[data-slot='tabs-trigger'\]\)\s*{(?<body>[^}]*)}/)?.groups?.body ??
      '';

    expect(tabListRule).toContain('overflow-x: auto');
    expect(tabListRule).toContain('scroll-snap-type: x proximity');
    expect(tabTriggerRule).toContain('flex: 0 0 auto');
    expect(tabTriggerRule).toContain('min-width: max-content');
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
      expect(document.body.textContent).toContain('Runs');
      expect(document.body.textContent).not.toContain('Current work');
      expect(refreshes).toBe(1);
    });
  });

  it('shows visible refresh progress while import history reloads', async () => {
    let releaseRefresh: () => void = () => {};
    const refreshGate = new Promise<void>((resolve) => {
      releaseRefresh = resolve;
    });
    class SlowRefreshImportJobRepository extends CompletingImportJobRepository {
      async listImportJobs(tenantId: string, inventoryId: string) {
        if (this.listCalls >= 1) {
          await refreshGate;
        }
        return super.listImportJobs(tenantId, inventoryId);
      }
    }
    const repository = new SlowRefreshImportJobRepository(structuredClone(seed));
    await mountImportWorkspace(repository);

    await waitFor(() => {
      expect(document.body.textContent).toContain('Current work');
    });

    buttonContaining('Refresh').click();

    await waitFor(() => {
      expect(buttonContaining('Refreshing').disabled).toBe(true);
      expect(document.body.querySelector('.busy-button-spinner')).toBeTruthy();
    });

    releaseRefresh();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Import finished');
      expect(buttonContaining('Refresh').disabled).toBe(false);
    });
  });

  it('shows unknown-total import phases without fake exact progress', async () => {
    await mountImportWorkspace(new UnknownProgressImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Current work');
      expect(document.body.textContent).toContain('Reading source');
      expect(document.body.textContent).toContain('Total not known yet');
    });

    const progressbar = document.body.querySelector<HTMLElement>('[role="progressbar"]');
    expect(progressbar).toBeTruthy();
    expect(progressbar?.classList.contains('indeterminate')).toBe(true);
    expect(progressbar?.getAttribute('aria-label')).toContain('total not known yet');
    expect(progressbar?.hasAttribute('aria-valuenow')).toBe(false);
    expect(progressbar?.querySelector('span')?.getAttribute('style') ?? '').not.toContain('width: 0%');
  });

  it('shows the original preview plan while an import is running', async () => {
    await mountImportWorkspace(new UnknownProgressImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Current work');
      expect(document.body.textContent).toContain('Reading source');
    });

    currentWorkRows()[0].click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Preview plan');
      expect(document.body.textContent).toContain('Original plan');
      expect(document.body.textContent).toContain('Serial number');
      expect(document.body.textContent).toContain('Garage');
      expect(document.body.textContent).toContain('Cordless drill');
      expect(document.body.textContent).toContain('drill-photo.jpg');
    });
  });

  it('does not turn completed unknown-total imports into fake exact progress', async () => {
    await mountImportWorkspace(new TerminalUnknownProgressImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
      expect(document.body.textContent).toContain('Completed');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Progress timeline');
      expect(document.body.textContent).toContain('Completed');
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
      expect(document.body.textContent).toContain('Cancelling');
      expect(document.body.textContent).not.toContain('Discarding imported items');
      expect(Array.from(document.body.querySelectorAll('button')).some((button) => button.textContent?.trim() === 'Cancel')).toBe(false);
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Cancellation is waiting for a safe stopping point.');
      expect(Array.from(document.body.querySelectorAll('button')).some((button) => button.textContent?.trim() === 'Cancel')).toBe(false);
    });
  });

  it('shows one visible cancellation progress label while cancellation is requested', async () => {
    let releaseCancellation: () => void = () => {};
    const cancellationGate = new Promise<void>((resolve) => {
      releaseCancellation = resolve;
    });
    class SlowCancellableImportJobRepository extends CancellableImportJobRepository {
      async cancelImportJob(
        tenantId: string,
        inventoryId: string,
        jobId: string,
        mode: ImportJobCancellationMode
      ): Promise<ImportJob> {
        await cancellationGate;
        return super.cancelImportJob(tenantId, inventoryId, jobId, mode);
      }
    }
    const repository = new SlowCancellableImportJobRepository(structuredClone(seed));
    await mountImportWorkspace(repository);

    await waitFor(() => {
      expect(document.body.textContent).toContain('Current work');
      expect(exactButton('Cancel')).toBeTruthy();
    });

    exactButton('Cancel').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Cancel Homebox?');
    });

    buttonContaining('Keep imported items').click();

    await waitFor(() => {
      expect(buttonContaining('Cancelling import').disabled).toBe(true);
      expect(document.body.textContent?.match(/Cancelling import/g)).toHaveLength(1);
      expect(document.body.querySelector('.busy-button-spinner')).toBeTruthy();
    });

    releaseCancellation();

    await waitFor(() => {
      expect(repository.cancellationModes).toEqual(['keep_partial_progress']);
      expect(document.body.textContent).toContain('Cancellation is waiting for a safe stopping point.');
    });
  });

  it('submits discard cancellation when selected before cancellation is already requested', async () => {
    const repository = new CancellableImportJobRepository(structuredClone(seed));
    await mountImportWorkspace(repository);

    await waitFor(() => {
      expect(document.body.textContent).toContain('Current work');
      expect(exactButton('Cancel')).toBeTruthy();
    });

    exactButton('Cancel').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Cancel Homebox?');
      expect(document.body.textContent).toContain('Stop future work and remove records created by this job. Audit history remains.');
    });

    buttonContaining('Discard imported items').click();

    await waitFor(() => {
      expect(repository.cancellationModes).toEqual(['discard_partial_progress']);
    });
  });

  it('moves focus into the cancellation choices and names the selected job', async () => {
    await mountImportWorkspace(new MultiCancellableImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Garage Homebox');
    });

    const cancelButton = document.body.querySelector<HTMLButtonElement>('button[aria-label^="Cancel Garage Homebox import"]');
    expect(cancelButton).toBeTruthy();
    cancelButton?.click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Cancel Garage Homebox?');
      expect(document.body.textContent).toContain('garage-homebox.local:7744');
      expect(document.body.textContent).toContain('Importing garage shelves');
      expect(document.activeElement?.textContent).toContain('Keep imported items');
    });
  });

  it('returns to import history after removing the selected terminal import job', async () => {
    await mountImportWorkspace(new TerminalImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
      expect(document.body.textContent).toContain('Homebox');
      expect(historyLedgerText()).not.toContain('Prepared by owner');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Progress timeline');
      expect(document.body.textContent).toContain('1 field reused');
      expect(document.body.textContent).toContain('1 location created');
      expect(document.body.textContent).toContain('2 assets skipped');
      expect(document.body.textContent).toContain('1 photo/file skipped');
      expect(document.body.textContent).toContain('2 warnings');
      expect(document.body.textContent).toContain('Method');
      expect(document.body.textContent).toContain('Prepared by owner');
      expect(document.body.textContent).not.toContain('Photos on');
      expect(document.body.textContent).not.toContain('Live Homebox connection');
      expect(document.body.textContent).toContain('Allowed local/private network address');
      expect(document.body.textContent).toContain('Allowed self-signed certificate');
      expect(document.body.textContent).toContain('Preview plan');
      expect(document.body.textContent).toContain('Serial number');
      expect(document.body.textContent).toContain('Garage');
      expect(document.body.textContent).toContain('Cordless drill');
      expect(document.body.textContent).toContain('drill-photo.jpg');
      expect(document.body.textContent).not.toContain('Inventory activity evidence for this run.');
    });

    buttonContaining('More').click();
    await waitFor(() => {
      expect(document.body.textContent).toContain('Inventory activity evidence for this run.');
      expect(document.body.textContent).toContain('Imported records and audit history remain.');
    });
    expect(linkContaining('View audit history').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/settings/activity'
    );

    buttonContaining('Remove from history').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Remove Homebox from history?');
      expect(document.body.textContent).toContain('Imported records and audit history will remain.');
    });

    confirmationButton('Remove from history').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('No import runs yet');
      expect(document.body.textContent).not.toContain('Progress timeline');
      expect(document.body.textContent).not.toContain('Remove from history');
    });
  });

  it('uses the current principal email in import history and detail actor copy when available', async () => {
    await mountImportWorkspace(new TerminalImportJobRepository(structuredClone(seed)), {
      currentPrincipal: { id: 'owner', email: 'owner@example.test' }
    });

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
      expect(historyLedgerText()).not.toContain('Prepared by owner@example.test');
      expect(document.body.textContent).not.toContain('Prepared by owner ·');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Prepared by owner@example.test');
      expect(document.body.textContent).not.toContain('Prepared by owner ·');
    });
  });

  it('uses the resolved import actor email before falling back to opaque principal IDs', async () => {
    class ResolvedActorImportJobRepository extends LongActorImportJobRepository {
      async listImportJobs(tenantId: string, inventoryId: string): Promise<ImportJob[]> {
        const jobs = await super.listImportJobs(tenantId, inventoryId);
        return jobs.map((job) => ({
          ...job,
          actor: { id: job.actorId ?? '', email: 'importer@example.test' }
        }));
      }
    }

    await mountImportWorkspace(new ResolvedActorImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Prepared by importer@example.test');
      expect(document.body.textContent).not.toContain('oidc_vZWJGXPHf8');
    });
  });

  it('opens audit history from import detail through the workspace router callback', async () => {
    let auditOpenCount = 0;
    await mountImportWorkspace(new TerminalImportJobRepository(structuredClone(seed)), {
      onOpenInventoryAuditHistory: () => {
        auditOpenCount += 1;
      }
    });

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Import details');
      expect(document.body.textContent).toContain('Homebox · homebox.local:7744');
    });

    buttonContaining('More').click();

    await waitFor(() => {
      expect(linkContaining('View audit history')).toBeTruthy();
      expect(document.body.textContent).toContain('Inventory activity evidence for this run.');
    });
    expect(linkContaining('View audit history').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/settings/activity'
    );

    linkContaining('View audit history').click();

    await waitFor(() => {
      expect(auditOpenCount).toBe(1);
    });
  });

  it('summarizes terminal import history as a scannable job ledger', async () => {
    await mountImportWorkspace(new TerminalImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
      expect(document.body.textContent).toContain('Warnings');
      expect(document.body.textContent).not.toContain('Action required');
      expect(document.body.textContent).not.toContain('Action required 1');
      expect(document.body.textContent).toContain('Completed');
      expect(document.body.textContent).toContain('Homebox');
      expect(historyLedgerText()).not.toContain('Prepared by owner');
      expect(document.body.textContent).toContain('Jul 6, 2026');
      expect(document.body.textContent).toContain('1 asset created');
      expect(document.body.textContent).not.toContain('No other import runs to show.');
      expect(historyLedgerText()).toContain('Completed with warnings.');
      expect(historyLedgerText()).toContain('Warnings');
    });
  });

  it('keeps warning-only imports in history and filters them separately', async () => {
    class MixedTerminalImportJobRepository extends TerminalImportJobRepository {
      private readonly cleanJob = {
        ...this.job,
        id: 'job-clean',
        source: {
          ...this.job.source,
          name: 'Clean Homebox',
          fingerprint: 'fingerprint-clean'
        },
        counts: {
          ...this.job.counts,
          warnings: 0,
          errors: 0,
          assetsSkipped: 0,
          attachmentsSkipped: 0
        },
        messages: []
      };

      constructor(seedData: typeof seed) {
        super(seedData);
        this.job = {
          ...this.job,
          messages: [
            {
              code: 'duplicate-asset',
              severity: 'warning',
              summary: 'Asset appears to have already been imported',
              detail: 'homebox-source-id duplicate',
              sourceName: 'Sarah Winter Clothes'
            },
            {
              code: 'duplicate-asset',
              severity: 'warning',
              summary: 'Asset appears to have already been imported',
              detail: 'homebox-source-id duplicate',
              sourceName: 'Baby Hats and Socks'
            }
          ]
        };
      }

      async listImportJobs(tenantId: string, inventoryId: string) {
        this.expectScope(tenantId, inventoryId);
        return [this.job, this.cleanJob];
      }

      async getImportJob(tenantId: string, inventoryId: string, jobId: string) {
        this.expectScope(tenantId, inventoryId);
        return jobId === this.cleanJob.id ? this.cleanJob : this.job;
      }
    }

    await mountImportWorkspace(new MixedTerminalImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Homebox');
      expect(document.body.textContent).toContain('Clean Homebox');
      expect(document.body.textContent).toContain('Warnings');
      expect(document.body.textContent).not.toContain('Action required');
      expect(historyLedgerText()).toContain('Completed with warnings.');
      expect(buttonContaining('Review Details')).toBeTruthy();
      expect(historyLedgerText()).toContain('Clean Homebox');
      expect(historyLedgerText()).toContain('Warnings');
      expect(historyLedgerText()).not.toContain('Action required');
      expect(ledgerSourceNames()).toEqual(['Homebox', 'Clean Homebox']);
      expect(columnHeader('Source')?.getAttribute('aria-sort')).toBe('none');
    });

    sortButton('source').click();

    await waitFor(() => {
      expect(columnHeader('Source')?.getAttribute('aria-sort')).toBe('ascending');
      expect(ledgerSourceNames()).toEqual(['Clean Homebox', 'Homebox']);
    });

    sortButton('source').click();

    await waitFor(() => {
      expect(columnHeader('Source')?.getAttribute('aria-sort')).toBe('descending');
      expect(ledgerSourceNames()).toEqual(['Homebox', 'Clean Homebox']);
    });

    buttonContaining('Review Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Issues');
      expect(document.body.textContent).toContain('2 affected records');
      expect(document.body.textContent).toContain('Already linked to an earlier import');
    });

    buttonContaining('Back to history').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Warnings');
    });

    buttonContaining('Warnings').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Warning-only imports.');
      expect(document.body.textContent).toContain('Homebox');
      expect(document.body.textContent).toContain('Warnings');
      expect(document.body.textContent).toContain('Details');
      expect(document.body.textContent).not.toContain('Clean Homebox');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Issues');
      expect(document.body.textContent).toContain('2 affected records');
    });
  });

  it('filters current work without showing an empty terminal ledger', async () => {
    class MixedCurrentAndTerminalImportJobRepository extends CompletingImportJobRepository {
      private readonly previewJob = {
        ...this.job,
        id: 'job-previewed',
        status: 'previewed' as const,
        source: {
          ...this.job.source,
          name: 'Preview Homebox',
          fingerprint: 'fingerprint-previewed'
        },
        progress: { phase: 'ready', done: 1, total: 1, message: 'Preview ready', updatedAt: '2026-07-06T12:00:00Z' },
        completedAt: undefined
      };

      private readonly completedJob = {
        ...this.job,
        id: 'job-completed',
        status: 'succeeded' as const,
        source: {
          ...this.job.source,
          name: 'Completed Homebox',
          fingerprint: 'fingerprint-completed'
        },
        counts: {
          ...this.job.counts,
          warnings: 0,
          errors: 0
        },
        progress: { phase: 'terminal', done: 1, total: 1, message: 'Import completed', updatedAt: '2026-07-06T12:05:00Z' },
        completedAt: '2026-07-06T12:05:00Z'
      };

      async listImportJobs(tenantId: string, inventoryId: string) {
        this.expectScope(tenantId, inventoryId);
        return [this.job, this.previewJob, this.completedJob];
      }
    }

    await mountImportWorkspace(new MixedCurrentAndTerminalImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Current work');
      expect(document.body.textContent).toContain('Completed Homebox');
      expect(statusStripText()).toContain('All runs 3');
      expect(statusButton('All runs')?.getAttribute('aria-pressed')).toBe('true');
    });

    buttonContaining('Running').click();

    await waitFor(() => {
      expect(statusButton('Running')?.getAttribute('aria-pressed')).toBe('true');
      expect(document.body.textContent).toContain('Current work');
      expect(document.body.textContent).toContain('Homebox');
      expect(document.body.textContent).not.toContain('No imports match this filter.');
      expect(document.body.textContent).not.toContain('Completed Homebox');
    });

    buttonContaining('Running').click();
    buttonContaining('Ready to review').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Current work');
      expect(document.body.textContent).toContain('Preview Homebox');
      expect(document.body.textContent).not.toContain('No imports match this filter.');
      expect(document.body.textContent).not.toContain('Completed Homebox');
    });

    currentWorkRows()[0].dispatchEvent(new KeyboardEvent('keydown', { key: ' ', bubbles: true }));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Import details');
      expect(document.body.textContent).toContain('Continue import');
    });
  });

  it('elevates failed and blocking-error imports as action required instead of warnings', async () => {
    class ActionRequiredImportJobRepository extends TerminalImportJobRepository {
      private readonly failedJob = {
        ...this.job,
        id: 'job-failed',
        status: 'failed' as const,
        source: {
          ...this.job.source,
          name: 'Failed Homebox',
          fingerprint: 'fingerprint-failed'
        },
        counts: {
          ...this.job.counts,
          warnings: 2,
          errors: 0
        },
        progress: { phase: 'terminal', done: 0, total: 1, message: 'Import failed', updatedAt: '2026-07-06T12:05:00Z' },
        completedAt: '2026-07-06T12:05:00Z'
      };

      private readonly blockingJob = {
        ...this.job,
        id: 'job-blocking',
        status: 'succeeded' as const,
        source: {
          ...this.job.source,
          name: 'Blocking Homebox',
          fingerprint: 'fingerprint-blocking'
        },
        counts: {
          ...this.job.counts,
          warnings: 1,
          errors: 1
        },
        progress: { phase: 'terminal', done: 1, total: 1, message: 'Import completed', updatedAt: '2026-07-06T12:06:00Z' },
        completedAt: '2026-07-06T12:06:00Z'
      };

      async listImportJobs(tenantId: string, inventoryId: string) {
        this.expectScope(tenantId, inventoryId);
        return [this.failedJob, this.blockingJob];
      }
    }

    await mountImportWorkspace(new ActionRequiredImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Action required');
      expect(document.body.textContent).toContain('2 imports require action');
      expect(document.body.textContent).toContain('Failed Homebox');
      expect(document.body.textContent).toContain('Blocking Homebox');
      expect(document.body.textContent).not.toContain('Warnings 2');
    });

    buttonContaining('Action required').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Imports that need action.');
      expect(historyLedgerText()).toContain('Failed Homebox');
      expect(historyLedgerText()).toContain('Blocking Homebox');
      expect(historyLedgerText()).toContain('Action required');
    });
  });

  it('keeps long actor identifiers compact but distinguishable', async () => {
    await mountImportWorkspace(new LongActorImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Prepared by oidc_vZWJGXP...ltriM27O9');
      expect(document.body.textContent).not.toContain('Prepared by signed-in user');
    });
  });

  it('does not expose discarded import resources as openable records', async () => {
    await mountImportWorkspace(new DiscardedImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
      expect(historyLedgerText()).toContain('Cancelled. Partial progress was discarded.');
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
      expect(document.body.textContent).toContain('Runs');
      expect(document.body.textContent).toContain('Homebox');
    });

    document.body.querySelector<HTMLButtonElement>('button[aria-label^="Remove from history Homebox import"]')?.click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Remove Homebox from history?');
      expect(document.body.textContent).toContain('This only removes the run from the import history list.');
      expect(document.body.textContent).toContain('Keep in history');
      expect(document.body.textContent).toContain('Runs');
      expect(historyLedgerText()).toContain('Completed with warnings.');
    });

    confirmationButton('Keep in history').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
      expect(document.body.textContent).toContain('Homebox');
      expect(document.body.textContent).not.toContain('Remove Homebox from history?');
    });

    document.body.querySelector<HTMLButtonElement>('button[aria-label^="Remove from history Homebox import"]')?.click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Remove Homebox from history?');
    });

    confirmationButton('Remove from history').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('No import runs yet');
      expect(document.body.textContent).not.toContain('Remove Homebox from history?');
    });
  });

  it('shows visible removal progress while hiding an import from history', async () => {
    let releaseRemoval: () => void = () => {};
    const removalGate = new Promise<void>((resolve) => {
      releaseRemoval = resolve;
    });
    class SlowRemoveImportJobRepository extends TerminalImportJobRepository {
      async removeImportJobFromHistory(tenantId: string, inventoryId: string, jobId: string): Promise<void> {
        await removalGate;
        await super.removeImportJobFromHistory(tenantId, inventoryId, jobId);
      }
    }
    await mountImportWorkspace(new SlowRemoveImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
      expect(document.body.textContent).toContain('Homebox');
    });

    document.body.querySelector<HTMLButtonElement>('button[aria-label^="Remove from history Homebox import"]')?.click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Remove Homebox from history?');
    });

    confirmationButton('Remove from history').click();

    await waitFor(() => {
      expect(buttonContaining('Removing from history').disabled).toBe(true);
      expect(document.body.querySelector('.busy-button-spinner')).toBeTruthy();
    });

    releaseRemoval();

    await waitFor(() => {
      expect(document.body.textContent).toContain('No import runs yet');
    });
  });

  it('uses user-facing imported record labels with source IDs as secondary metadata', async () => {
    await mountImportWorkspace(new ResourcefulImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Imported records');
      expect(document.body.textContent).toContain('Passport');
      expect(document.body.textContent).toContain('passport-photo.jpg');
      expect(document.body.textContent).toContain('Source asset: homebox-item-passport');
      expect(document.body.textContent).toContain('Source attachment: homebox-photo-passport');
      expect(document.body.textContent).not.toContain('Asset · asset-imported-passport');
      expect(document.body.textContent).not.toContain('Photo/file · attachment-passport-photo');
    });
  });

  it('opens imported records from detail through the workspace router callback', async () => {
    const openedAssetIds: string[] = [];
    await mountImportWorkspace(new ResourcefulImportJobRepository(structuredClone(seed)), {
      onOpenImportedAssetId: async (assetId) => {
        openedAssetIds.push(assetId);
      }
    });

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(buttonContaining('Records')).toBeTruthy();
    });

    buttonContaining('Records').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Imported records');
      expect(linkContaining('Open')).toBeTruthy();
    });

    const openLink = linkContaining('Open');
    expect(openLink.getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/assets/asset-imported-passport');

    openLink.click();

    await waitFor(() => {
      expect(openedAssetIds).toEqual(['asset-imported-passport']);
    });
  });

  it('uses overview metric tiles as shortcuts to issues and records', async () => {
    await mountImportWorkspace(new ResourcefulImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(activeDetailTabText()).toContain('Issues');
    });

    buttonContaining('Overview').click();

    await waitFor(() => {
      expect(activeDetailTabText()).toContain('Overview');
      expect(document.body.querySelector<HTMLButtonElement>('button[aria-label^="Open issues for"]')).toBeTruthy();
      expect(document.body.querySelector<HTMLButtonElement>('button[aria-label^="Open imported records for"]')).toBeTruthy();
    });

    document.body.querySelector<HTMLButtonElement>('button[aria-label^="Open issues for"]')?.click();

    await waitFor(() => {
      expect(activeDetailTabText()).toContain('Issues');
    });

    buttonContaining('Overview').click();
    document.body.querySelector<HTMLButtonElement>('button[aria-label^="Open imported records for"]')?.click();

    await waitFor(() => {
      expect(activeDetailTabText()).toContain('Records');
      expect(document.body.textContent).toContain('Imported records');
    });
  });

  it('pages many imported record summaries in job detail', async () => {
    class ManyResourcefulImportJobRepository extends ResourcefulImportJobRepository {
      constructor(seedData: typeof seed) {
        super(seedData);
        this.job = {
          ...this.job,
          resources: Array.from({ length: 28 }, (_, index) => ({
            resourceType: 'asset' as const,
            resourceId: `asset-imported-${index + 1}`,
            displayName: `Imported record ${index + 1}`,
            sourceEntityType: 'asset' as const,
            sourceEntityId: `homebox-item-${index + 1}`,
            createdAt: '2026-07-06T12:03:00Z'
          }))
        };
      }
    }

    await mountImportWorkspace(new ManyResourcefulImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
    });

    buttonContaining('Details').click();
    await waitFor(() => {
      expect(buttonContaining('Records')).toBeTruthy();
    });
    buttonContaining('Records').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('1-25 of 28');
      expect(document.body.textContent).toContain('Imported record 25');
      expect(document.body.textContent).not.toContain('Imported record 26');
      expect(document.body.textContent).toContain('Page 1 of 2');
    });
    expect(document.body.querySelector<HTMLElement>('.resource-list')?.getAttribute('role')).toBe('table');
    expect(buttonContaining('Previous').disabled).toBe(true);
    expect(buttonContaining('Next').disabled).toBe(false);

    buttonContaining('Next').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('26-28 of 28');
      expect(document.body.textContent).toContain('Imported record 28');
      expect(document.body.textContent).not.toContain('Imported record 25');
      expect(document.body.textContent).toContain('Page 2 of 2');
    });
    expect(buttonContaining('Previous').disabled).toBe(false);
    expect(buttonContaining('Next').disabled).toBe(true);
  });

  it('shows preview-preserved warnings in terminal import job detail', async () => {
    await mountImportWorkspace(new TerminalPreviewMessageImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
      expect(document.body.textContent).toContain('Warnings');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Attachment could not be imported');
      expect(document.body.textContent).toContain('Homebox reported a file without downloadable bytes.');
      expect(document.body.textContent).toContain('receipt.png');
      expect(document.body.textContent).not.toContain('Open Issues to review warning groups before treating this import as clean.');
      expect(document.body.querySelector('.detail-issue-callout.warning')).toBeFalsy();
      expect(document.body.querySelector('.detail-issue-callout.action')).toBeFalsy();
      expect(document.body.querySelector('.summary-tile.warning')).toBeTruthy();
      expect(document.body.textContent).not.toContain('No import messages.');
    });
  });

  it('uses returned warning messages for detail severity when warning counts are stale', async () => {
    class StaleWarningCountImportJobRepository extends TerminalImportJobRepository {
      constructor(seedData: typeof seed) {
        super(seedData);
        this.job = {
          ...this.job,
          counts: {
            ...this.job.counts,
            warnings: 0,
            errors: 0
          },
          messages: [
            {
              code: 'warning-count-stale',
              severity: 'warning',
              summary: 'Attachment could not be imported',
              detail: 'Homebox returned a warning after counts were calculated.',
              sourceName: 'receipt.png'
            }
          ]
        };
      }
    }

    await mountImportWorkspace(new StaleWarningCountImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
      expect(document.body.textContent).toContain('Homebox');
      expect(statusStripText()).toContain('Warnings 1');
      expect(historyLedgerText()).toContain('Warnings');
    });

    buttonContaining('Warnings').click();

    await waitFor(() => {
      expect(statusButton('Warnings')?.getAttribute('aria-pressed')).toBe('true');
      expect(historyLedgerText()).toContain('Homebox');
      expect(historyLedgerText()).toContain('Warnings');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Attachment could not be imported');
      expect(document.body.textContent).not.toContain('Open Issues to review warning groups before treating this import as clean.');
      expect(document.body.querySelector('.detail-issue-callout.warning')).toBeFalsy();
      expect(document.body.querySelector('.detail-issue-callout.action')).toBeFalsy();
    });
  });

  it('keeps detail issue tabs and stats aligned when returned messages contain exact duplicates', async () => {
    class DuplicateWarningMessageImportJobRepository extends TerminalImportJobRepository {
      constructor(seedData: typeof seed) {
        super(seedData);
        this.job = {
          ...this.job,
          counts: {
            ...this.job.counts,
            warnings: 2,
            errors: 0
          },
          messages: [
            {
              code: 'partial-date',
              severity: 'warning',
              summary: 'Homebox partial date imported as text',
              detail: '0001-09-28',
              sourceId: 'source-compressed-air'
            },
            {
              code: 'partial-date',
              severity: 'warning',
              summary: 'Homebox partial date imported as text',
              detail: '0001-09-28',
              sourceId: 'source-compressed-air'
            },
            {
              code: 'partial-date',
              severity: 'warning',
              summary: 'Homebox partial date imported as text',
              detail: '0001-09-28',
              sourceId: 'source-wood-glue'
            }
          ]
        };
      }
    }

    await mountImportWorkspace(new DuplicateWarningMessageImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(historyLedgerText()).toContain('Warnings');
    });

    buttonContaining('Review Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Issues (2)');
      expect(document.body.textContent).toContain('Warnings 2');
      expect(document.body.textContent).toContain('2 affected records');
      expect(document.body.querySelectorAll('.message-row')).toHaveLength(2);
      expect(document.body.textContent).toContain('Source ID source-compressed-air');
      expect(document.body.textContent).toContain('Source ID source-wood-glue');
    });
  });

  it('uses reported warning counts for detail totals while showing distinct affected messages', async () => {
    class ReportedAndDedupedWarningMessageImportJobRepository extends TerminalImportJobRepository {
      constructor(seedData: typeof seed) {
        super(seedData);
        this.job = {
          ...this.job,
          counts: {
            ...this.job.counts,
            warnings: 4,
            errors: 0
          },
          messages: [
            {
              code: 'partial-date',
              severity: 'warning',
              summary: 'Homebox partial date imported as text',
              detail: '0001-09-28',
              sourceId: 'source-compressed-air'
            },
            {
              code: 'partial-date',
              severity: 'warning',
              summary: 'Homebox partial date imported as text',
              detail: '0001-09-28',
              sourceId: 'source-compressed-air'
            },
            {
              code: 'partial-date',
              severity: 'warning',
              summary: 'Homebox partial date imported as text',
              detail: '0001-09-28',
              sourceId: 'source-wood-glue'
            }
          ]
        };
      }
    }

    await mountImportWorkspace(new ReportedAndDedupedWarningMessageImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(historyLedgerText()).toContain('Warnings');
    });

    buttonContaining('Review Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Issues (4)');
      expect(document.body.textContent).toContain('Warnings 4');
      expect(document.body.textContent).toContain('2 affected records');
      expect(document.body.querySelectorAll('.message-row')).toHaveLength(2);
    });
  });

  it('uses returned error messages for history severity when error counts are stale', async () => {
    class StaleErrorCountImportJobRepository extends TerminalImportJobRepository {
      constructor(seedData: typeof seed) {
        super(seedData);
        this.job = {
          ...this.job,
          counts: {
            ...this.job.counts,
            warnings: 0,
            errors: 0
          },
          messages: [
            {
              code: 'error-count-stale',
              severity: 'error',
              summary: 'Source changed after preview',
              detail: 'Re-preview is required before this import can be trusted.',
              sourceName: 'Homebox'
            }
          ]
        };
      }
    }

    await mountImportWorkspace(new StaleErrorCountImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(statusStripText()).toContain('Action required 1');
      expect(historyLedgerText()).toContain('Action required');
      expect(historyLedgerText()).not.toContain('Warnings');
    });

    buttonContaining('Action required').click();

    await waitFor(() => {
      expect(statusButton('Action required')?.getAttribute('aria-pressed')).toBe('true');
      expect(historyLedgerText()).toContain('Homebox');
      expect(historyLedgerText()).toContain('Action required');
    });
  });

  it('does not duplicate the review issues action when warning details already open on the issues tab', async () => {
    await mountImportWorkspace(new TerminalPreviewMessageImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
      expect(document.body.textContent).toContain('Warnings');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Issues');
      expect(document.body.textContent).toContain('Attachment could not be imported');
    });

    expect(Array.from(document.body.querySelectorAll('button')).some((button) => button.textContent?.trim() === 'Review issues')).toBe(false);
  });

  it('prefers terminal import messages over preview-preserved messages in job detail', async () => {
    await mountImportWorkspace(new TerminalJobAndPreviewMessageImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
      expect(document.body.textContent).toContain('Warnings');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Attachment imported without primary-photo status');
      expect(document.body.textContent).toContain('Homebox did not identify a primary image after download.');
      expect(document.body.textContent).toContain('downloaded-manual.png');
      expect(document.body.textContent).not.toContain('Homebox reported a file without downloadable bytes.');
      expect(document.body.textContent).not.toContain('receipt.png');
    });
  });

  it('loads dedicated import job detail when a history row is opened', async () => {
    const repository = new DetailOnlyResourcefulImportJobRepository(structuredClone(seed));
    await mountImportWorkspace(repository);

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
      expect(document.body.textContent).toContain('Homebox');
      expect(document.body.textContent).not.toContain('Imported records');
    });

    document.body.querySelector<HTMLElement>('.history-row.clickable-row')?.click();

    await waitFor(() => {
      expect(repository.detailCalls).toBe(1);
      expect(document.body.textContent).toContain('Imported records');
      expect(document.body.textContent).toContain('Source asset: homebox-detail-asset');
      expect(document.body.textContent).toContain('Detail warning from job detail');
    });

    exactButton('Refresh').click();

    await waitFor(() => {
      expect(repository.detailCalls).toBe(2);
      expect(document.body.textContent).toContain('Imported records');
      expect(document.body.textContent).toContain('Source asset: homebox-detail-asset');
      expect(document.body.textContent).toContain('Source asset: homebox-detail-refresh-asset');
      expect(document.body.textContent).toContain('Detail warning from job detail');
      expect(document.body.textContent).toContain('Detail warning after refresh');
    });
  });

  it('opens routed import job detail on the requested tab', async () => {
    const repository = new DetailOnlyResourcefulImportJobRepository(structuredClone(seed));
    await mountImportWorkspace(repository, { importJobId: 'job-terminal', importTab: 'records' });

    await waitFor(() => {
      expect(repository.detailCalls).toBe(1);
      expect(document.body.textContent).toContain('Import details');
      expect(document.body.textContent).toContain('Imported records');
      expect(document.body.textContent).toContain('Source asset: homebox-detail-asset');
      expect(activeDetailTabText()).toContain('Records');
    });
  });

  it('reports import detail selection and tab changes to the workspace route', async () => {
    const selectedRoutes: Array<{ jobId: string | null; tab?: string | null }> = [];
    const selectedTabs: Array<string | null> = [];
    await mountImportWorkspace(new TerminalImportJobRepository(structuredClone(seed)), {
      onImportJobSelectionChange: (jobId, tab) => selectedRoutes.push({ jobId, tab }),
      onImportJobTabChange: (tab) => selectedTabs.push(tab)
    });

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
    });

    document.body.querySelector<HTMLElement>('.history-row.clickable-row')?.click();

    await waitFor(() => {
      expect(selectedRoutes).toContainEqual({ jobId: 'job-terminal', tab: null });
      expect(document.body.textContent).toContain('Import details');
    });

    buttonContaining('Records').click();

    await waitFor(() => {
      expect(selectedTabs).toContain('records');
    });
  });

  it('shows visible refresh progress while import details reload', async () => {
    let releaseRefresh: () => void = () => {};
    const refreshGate = new Promise<void>((resolve) => {
      releaseRefresh = resolve;
    });
    class SlowDetailRefreshImportJobRepository extends DetailOnlyResourcefulImportJobRepository {
      async getImportJob(tenantId: string, inventoryId: string, jobId: string) {
        if (this.detailCalls >= 1) {
          await refreshGate;
        }
        return super.getImportJob(tenantId, inventoryId, jobId);
      }
    }
    const repository = new SlowDetailRefreshImportJobRepository(structuredClone(seed));
    await mountImportWorkspace(repository);

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
      expect(document.body.textContent).toContain('Homebox');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(repository.detailCalls).toBe(1);
      expect(document.body.textContent).toContain('Imported records');
    });

    exactButton('Refresh').click();

    await waitFor(() => {
      expect(buttonContaining('Refreshing').disabled).toBe(true);
      expect(document.body.querySelector('.busy-button-spinner')).toBeTruthy();
    });

    releaseRefresh();

    await waitFor(() => {
      expect(repository.detailCalls).toBe(2);
      expect(document.body.textContent).toContain('Detail warning after refresh');
      expect(buttonContaining('Refresh').disabled).toBe(false);
    });
  });

  it('keeps the current import detail visible when detail refresh fails', async () => {
    await mountImportWorkspace(new DetailRefreshFailureImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
      expect(document.body.textContent).toContain('Warnings');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Import details');
      expect(document.body.textContent).toContain('Completed with warnings.');
      expect(document.body.textContent).toContain('Import details could not be refreshed.');
      expect(document.body.textContent).not.toContain('provider-stacktrace');
      expect(document.body.textContent).not.toContain('password=secret');
      expect(document.body.textContent).not.toContain('Import failed before it could finish.');
    });
  });

  it('keeps discard-failed jobs visible without a remove-from-history action', async () => {
    await mountImportWorkspace(new DiscardFailedImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Discard failed');
    });
    expect(document.body.querySelector<HTMLButtonElement>('button[aria-label^="Remove from history Homebox import"]')).toBeFalsy();

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('cleanup will retry');
      expect(document.body.textContent).not.toContain('Open Issues to review what needs action.');
      expect(document.body.querySelector('.detail-issue-callout.action')).toBeFalsy();
      expect(document.body.querySelector('.summary-tile.action')).toBeTruthy();
      expect(document.body.querySelector('.detail-issue-callout.warning')).toBeFalsy();
    });
    buttonContaining('More').click();
    await waitFor(() => {
      expect(document.body.textContent).toContain('Inventory activity evidence for this run.');
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

    currentWorkRows()[0].dispatchEvent(new KeyboardEvent('keydown', { key: ' ', bubbles: true }));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Import details');
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

    exactButton('Back').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Choose import method');
      expect(document.body.querySelector<HTMLInputElement>('#homebox-csv')).toBeFalsy();
    });

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

function historyLedgerText(): string {
  return document.body.querySelector('.history-ledger')?.textContent ?? '';
}

function currentWorkRows(): HTMLElement[] {
  return Array.from(document.body.querySelectorAll<HTMLElement>('.current-work-section .clickable-row'));
}

function ledgerSourceNames(): string[] {
  return Array.from(document.body.querySelectorAll<HTMLElement>('.history-ledger .history-row .history-title strong')).map((node) =>
    node.textContent?.trim() ?? ''
  );
}

function columnHeader(text: string): HTMLElement | null {
  return Array.from(document.body.querySelectorAll<HTMLElement>('.history-ledger-head [role="columnheader"]')).find((node) =>
    node.textContent?.includes(text)
  ) ?? null;
}

function sortButton(label: string): HTMLButtonElement {
  const button = Array.from(document.body.querySelectorAll<HTMLButtonElement>('.history-ledger-head button')).find((candidate) =>
    candidate.getAttribute('aria-label')?.includes(label)
  );
  if (!button) {
    throw new Error(`Missing sort button containing ${label}`);
  }
  return button;
}

function confirmationButton(text: string): HTMLButtonElement {
  const dialog = document.body.querySelector<HTMLElement>('[role="dialog"]');
  const button = Array.from(dialog?.querySelectorAll<HTMLButtonElement>('button') ?? []).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!button) {
    throw new Error(`Missing confirmation button containing ${text}`);
  }
  return button;
}

function activeDetailTabText(): string {
  return (
    Array.from(document.body.querySelectorAll<HTMLElement>('[data-slot="tabs-trigger"]')).find((candidate) =>
      candidate.hasAttribute('data-active') || candidate.getAttribute('aria-selected') === 'true' || candidate.getAttribute('data-state') === 'active'
    )?.textContent ?? ''
  );
}

function statusStripText(): string {
  return document.body.querySelector('.history-status-strip')?.textContent ?? '';
}

function statusButton(label: string): HTMLButtonElement | null {
  return Array.from(document.body.querySelectorAll<HTMLButtonElement>('.history-status-strip button')).find((button) =>
    button.textContent?.includes(label)
  ) ?? null;
}
