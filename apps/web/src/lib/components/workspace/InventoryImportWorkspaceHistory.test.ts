import { afterEach, describe, expect, it } from 'vitest';
import { SeededInventoryRepository } from '$lib/adapters/memory/seededInventoryRepository';
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

    buttonContaining('Details').click();

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
      expect(document.body.textContent).toContain('History');
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
      expect(document.body.textContent).toContain('Source');
      expect(document.body.textContent).toContain('Photos on');
      expect(document.body.textContent).toContain('Private-network URLs allowed');
      expect(document.body.textContent).toContain('Self-signed TLS allowed');
      expect(document.body.textContent).toContain('Preview plan');
      expect(document.body.textContent).toContain('Serial number');
      expect(document.body.textContent).toContain('Garage');
      expect(document.body.textContent).toContain('Cordless drill');
      expect(document.body.textContent).toContain('drill-photo.jpg');
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
      expect(document.body.textContent).toContain('No other import runs to show.');
      expect(historyLedgerText()).not.toContain('Completed with warnings.');
    });
  });

  it('lets attention summary filter directly to imports that need review', async () => {
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
      expect(document.body.textContent).toContain('1 import needs attention');
      expect(document.body.textContent).toContain('1 import has warnings or failed work.');
      expect(document.body.textContent).toContain('2 warnings');
      expect(buttonContaining('Review')).toBeTruthy();
      expect(historyLedgerText()).toContain('Clean Homebox');
      expect(historyLedgerText()).not.toContain('Completed with warnings.');
      expect(historyLedgerText()).not.toContain('Needs attention');
    });

    buttonContaining('Review').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Issues');
      expect(document.body.textContent).toContain('2 affected records');
      expect(document.body.textContent).toContain('Already linked to an earlier import');
    });

    buttonContaining('Back to history').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Show attention history');
    });

    buttonContaining('Show attention history').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Showing imports with warnings, errors, or cleanup work.');
      expect(document.body.textContent).toContain('Homebox');
      expect(document.body.textContent).toContain('Needs attention');
      expect(document.body.textContent).toContain('Details');
      expect(document.body.textContent).not.toContain('Clean Homebox');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Issues');
      expect(document.body.textContent).toContain('2 affected records');
    });
  });

  it('keeps long actor identifiers compact but distinguishable', async () => {
    await mountImportWorkspace(new LongActorImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Prepared by oidc_vZWJGXP...ltriM27O9');
      expect(document.body.textContent).not.toContain('Prepared by signed-in user');
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

    document.body.querySelector<HTMLButtonElement>('button[aria-label^="Remove from history Homebox import"]')?.click();

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

    document.body.querySelector<HTMLButtonElement>('button[aria-label^="Remove from history Homebox import"]')?.click();

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
      expect(document.body.textContent).toContain('Passport');
      expect(document.body.textContent).toContain('passport-photo.jpg');
      expect(document.body.textContent).toContain('Source asset: homebox-item-passport');
      expect(document.body.textContent).toContain('Source attachment: homebox-photo-passport');
      expect(document.body.textContent).not.toContain('Asset · asset-imported-passport');
      expect(document.body.textContent).not.toContain('Photo/file · attachment-passport-photo');
    });
  });

  it('keeps many imported record summaries visually bounded in job detail', async () => {
    class ManyResourcefulImportJobRepository extends ResourcefulImportJobRepository {
      constructor(seedData: typeof seed) {
        super(seedData);
        this.job = {
          ...this.job,
          resources: Array.from({ length: 18 }, (_, index) => ({
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
      expect(document.body.textContent).toContain('History');
    });

    buttonContaining('Details').click();
    await waitFor(() => {
      expect(buttonContaining('Records')).toBeTruthy();
    });
    buttonContaining('Records').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Showing 12 of 18');
      expect(document.body.textContent).toContain('Imported record 12');
      expect(document.body.textContent).not.toContain('Imported record 13');
      expect(document.body.textContent).toContain('6 more imported records hidden.');
    });
    const recordRegion = document.body.querySelector<HTMLElement>('#imported-record-summaries');
    expect(recordRegion?.getAttribute('role')).toBe('region');
    expect(recordRegion?.getAttribute('aria-label')).toBe('Imported record summaries');
    expect(recordRegion?.getAttribute('tabindex')).toBe('0');
    expect(buttonContaining('Show more records').getAttribute('aria-controls')).toBe('imported-record-summaries');
    expect(buttonContaining('Show more records').getAttribute('aria-expanded')).toBe('false');

    buttonContaining('Show more records').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('18 records');
      expect(document.body.textContent).toContain('Imported record 18');
      expect(document.body.textContent).toContain('All returned record summaries are shown.');
    });
    expect(buttonContaining('Show fewer').getAttribute('aria-controls')).toBe('imported-record-summaries');
    expect(buttonContaining('Show fewer').getAttribute('aria-expanded')).toBe('true');
    expect(document.body.querySelector('.resource-list.bounded')).toBeTruthy();
  });

  it('shows preview-preserved warnings in terminal import job detail', async () => {
    await mountImportWorkspace(new TerminalPreviewMessageImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('History');
      expect(document.body.textContent).toContain('Completed with warnings.');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Attachment could not be imported');
      expect(document.body.textContent).toContain('Homebox reported a file without downloadable bytes.');
      expect(document.body.textContent).toContain('receipt.png');
      expect(document.body.textContent).not.toContain('No import messages.');
    });
  });

  it('does not duplicate the review issues action when warning details already open on the issues tab', async () => {
    await mountImportWorkspace(new TerminalPreviewMessageImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('History');
      expect(document.body.textContent).toContain('Completed with warnings.');
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
      expect(document.body.textContent).toContain('History');
      expect(document.body.textContent).toContain('Completed with warnings.');
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
      expect(document.body.textContent).toContain('History');
      expect(document.body.textContent).toContain('Homebox');
      expect(document.body.textContent).not.toContain('Imported records');
    });

    buttonContaining('Details').click();

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

  it('keeps the current import detail visible when detail refresh fails', async () => {
    await mountImportWorkspace(new DetailRefreshFailureImportJobRepository(structuredClone(seed)));

    await waitFor(() => {
      expect(document.body.textContent).toContain('History');
      expect(document.body.textContent).toContain('Completed with warnings.');
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
