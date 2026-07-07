import { afterEach, describe, expect, it } from 'vitest';
import { SeededInventoryRepository } from '$lib/adapters/memory/seededInventoryRepository';
import type { ImportJob } from '$lib/domain/inventory';
import {
  CountingImportJobRepository,
  CompletingStartedImportRepository,
  EmptyImportPreviewRepository,
  ImportPreviewCountRepository,
  ImportPreviewGenericInvalidRequestRepository,
  ImportPreviewRecordingRepository,
  ImportStartRecordingRepository,
  PreviewHierarchyRepository,
  PreviewMessageOnlyRepository,
  buttonContaining,
  checkboxContaining,
  cleanupImportWorkspace,
  controlContaining,
  detailsContaining,
  exactButton,
  fileLike,
  mountImportWorkspace,
  openLiveHomeboxSetup,
  seed,
  setFileInputFiles,
  setInputValue,
  waitFor
} from './InventoryImportWorkspace.test-helpers';

afterEach(() => {
  cleanupImportWorkspace();
});

describe('InventoryImportWorkspace import setup and preview', () => {
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

  it('explains empty import history when the user can view but not create import jobs', async () => {
    const viewOnlyImportInventory = {
      ...seed.inventories[0],
      access: { relationship: 'viewer', permissions: ['view', 'view_import_job'] }
    };

    await mountImportWorkspace(new SeededInventoryRepository(structuredClone(seed)), { inventory: viewOnlyImportInventory });

    await waitFor(() => {
      expect(document.body.textContent).toContain('No import runs yet');
      expect(document.body.textContent).toContain('Creating imports requires import job create access.');
    });
    expect(buttonsNamed('New import')).toHaveLength(0);
  });

  it('confirms live Homebox sources with https as the schemeless URL default', async () => {
    const repository = new ImportPreviewRecordingRepository(structuredClone(seed));
    await mountImportWorkspace(repository);

    await waitFor(() => {
      expect(document.body.textContent).toContain('No import runs yet');
    });
    const emptyState = document.body.querySelector<HTMLElement>('.import-history-empty-state');
    const historyHeader = document.body.querySelector<HTMLElement>('.history-header');
    expect(emptyState).toBeTruthy();
    expect(historyHeader).toBeTruthy();
    expect(buttonsNamedWithin(emptyState!, 'New import')).toHaveLength(1);
    expect(buttonsNamedWithin(historyHeader!, 'New import')).toHaveLength(0);

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

  it('shows selected live Homebox connection protections in the preview before start', async () => {
    const repository = new ImportPreviewRecordingRepository(structuredClone(seed));
    await mountImportWorkspace(repository);

    await openLiveHomeboxSetup();

    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-url')!, 'http://homebox.local:7744');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-user')!, 'codex@jsksell.com');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-password')!, 'asldfj3290f!');

    const advanced = detailsContaining('Connection options');
    advanced.open = true;
    advanced.dispatchEvent(new Event('toggle'));
    checkboxContaining('Allow private-network Homebox URL').click();
    checkboxContaining('Allow self-signed TLS certificate').click();

    await waitFor(() => {
      expect(buttonContaining('Confirm connection').disabled).toBe(false);
    });

    buttonContaining('Confirm connection').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Preview import');
      expect(document.body.textContent).toContain('Private-network URLs allowed');
      expect(document.body.textContent).toContain('Self-signed TLS allowed');
      expect(document.body.textContent).toContain('Photos on');
      expect(buttonContaining('Start background import').disabled).toBe(false);
    });
    const sourceOptions = document.body.querySelector<HTMLUListElement>('ul[aria-label="Selected source options"]');
    expect(sourceOptions).toBeTruthy();
    expect(sourceOptions?.querySelectorAll('li')).toHaveLength(3);
    expect(repository.previewInputs[0]).toMatchObject({
      allowPrivateNetwork: true,
      allowInsecureTLS: true
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
    expect(document.body.querySelector<HTMLInputElement>('#homebox-csv')?.getAttribute('autocapitalize')).toBe('none');
    expect(document.body.querySelector<HTMLInputElement>('#homebox-csv')?.getAttribute('autocorrect')).toBe('off');
    expect(document.body.querySelector<HTMLInputElement>('#homebox-csv')?.getAttribute('spellcheck')).toBe('false');

    setFileInputFiles(document.body.querySelector<HTMLInputElement>('#homebox-csv')!, [
      fileLike('homebox.csv', 10 * 1024 * 1024, new TextEncoder().encode('name\nok').buffer)
    ]);

    await waitFor(() => {
      expect(buttonContaining('Prepare preview').disabled).toBe(false);
    });

    buttonContaining('Prepare preview').click();

    await waitFor(() => {
      expect(repository.previewInputs).toHaveLength(1);
      expect(document.body.textContent).toContain('photos unavailable');
      expect(document.body.textContent).toContain('Photos unavailable from CSV');
      expect(document.body.textContent).not.toContain('images disabled');
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

  it('keeps preview issues above plan samples so warnings are not buried', async () => {
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
      expect(document.body.textContent).toContain('Preview import');
      expect(document.body.textContent).toContain('Attachment will be skipped');
      expect(document.body.textContent).toContain('Plan samples');
    });

    const issuesSection = document.body.querySelector<HTMLElement>('.preview-issues-section');
    const samplesSection = document.body.querySelector<HTMLElement>('.preview-samples-section');
    expect(issuesSection).toBeTruthy();
    expect(samplesSection).toBeTruthy();
    expect(issuesSection!.compareDocumentPosition(samplesSection!) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
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
    const repository = new ImportStartRecordingRepository(structuredClone(seed));
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
      expect(repository.startInputs).toHaveLength(1);
      expect(document.body.textContent).toContain('Import is running');
      expect(document.body.textContent).toContain('You can leave this page and return from import history.');
      expect(document.body.querySelector('[aria-current="step"]')?.textContent).toContain('Run');
      expect(document.body.textContent).not.toContain('Step 4 of 4');
    });

    buttonContaining('View in history').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('1 running now');
      expect(document.body.textContent).toContain('Current work');
      expect(document.body.textContent).not.toContain('Import could not be started.');
    });
    expect(repository.startInputs[0]).toMatchObject({
      input: {
        sourceType: 'legacy_homebox',
        baseUrl: 'http://homebox.local:7744',
        username: 'codex@jsksell.com',
        password: 'asldfj3290f!'
      }
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Progress timeline');
      expect(document.body.textContent).toContain('Reading source');
    });
  });

  it('keeps the run handoff current when the started job refreshes', async () => {
    const repository = new CompletingStartedImportRepository(structuredClone(seed));
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
      expect(buttonContaining('Start background import').disabled).toBe(false);
    });
    buttonContaining('Start background import').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Import is running');
      expect(document.body.textContent).toContain('Reading source');
    });

    exactButton('Refresh').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Import finished');
      expect(document.body.textContent).toContain('Completed successfully.');
      expect(document.body.textContent).not.toContain('Import is running');
    });
  });

  it('shows visible refresh progress while the run handoff reloads', async () => {
    let releaseRefresh: () => void = () => {};
    const refreshGate = new Promise<void>((resolve) => {
      releaseRefresh = resolve;
    });
    class SlowRunRefreshRepository extends ImportStartRecordingRepository {
      async listImportJobs(tenantId: string, inventoryId: string): Promise<ImportJob[]> {
        const jobs = await super.listImportJobs(tenantId, inventoryId);
        if (jobs.some((job) => job.status === 'running')) {
          await refreshGate;
        }
        return jobs;
      }
    }
    const repository = new SlowRunRefreshRepository(structuredClone(seed));
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
      expect(buttonContaining('Start background import').disabled).toBe(false);
    });
    buttonContaining('Start background import').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Import is running');
    });

    exactButton('Refresh').click();

    await waitFor(() => {
      expect(buttonContaining('Refreshing').disabled).toBe(true);
      expect(document.body.querySelector('.busy-button-spinner')).toBeTruthy();
    });

    releaseRefresh();

    await waitFor(() => {
      expect(buttonContaining('Refresh').disabled).toBe(false);
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
      expect(document.body.querySelector<HTMLInputElement>('#homebox-password')?.value).toBe('asldfj3290f!');
      expect(buttonContaining('Confirm connection').disabled).toBe(false);
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
      expect(document.body.querySelector<HTMLInputElement>('#homebox-password')?.value).toBe('asldfj3290f!');
      expect(buttonContaining('Confirm connection').disabled).toBe(false);
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

  it('lets users navigate reachable wizard steps without losing preview progress', async () => {
    const repository = new ImportPreviewRecordingRepository(structuredClone(seed));
    await mountImportWorkspace(repository);

    await openLiveHomeboxSetup();
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-url')!, 'http://homebox.local:7744');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-user')!, 'codex@jsksell.com');
    setInputValue(document.body.querySelector<HTMLInputElement>('#homebox-password')!, 'asldfj3290f!');

    await waitFor(() => {
      expect(buttonContaining('Confirm connection').disabled).toBe(false);
    });
    expect(stepButton('Preview')).toBeUndefined();

    buttonContaining('Confirm connection').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Preview import');
      expect(buttonContaining('Start background import').disabled).toBe(false);
      expect(stepButton('Preview')).toBeTruthy();
    });

    requiredStepButton('Connect').click();
    await waitFor(() => {
      expect(document.body.textContent).toContain('Connect to Homebox');
      expect(document.body.querySelector<HTMLInputElement>('#homebox-url')?.value).toBe('http://homebox.local:7744');
      expect(document.body.querySelector<HTMLInputElement>('#homebox-user')?.value).toBe('codex@jsksell.com');
      expect(document.body.querySelector<HTMLInputElement>('#homebox-password')?.value).toBe('asldfj3290f!');
    });

    requiredStepButton('Preview').click();
    await waitFor(() => {
      expect(document.body.textContent).toContain('Preview import');
      expect(buttonContaining('Start background import').disabled).toBe(false);
    });
    expect(repository.previewInputs).toHaveLength(1);

    requiredStepButton('Source').click();
    await waitFor(() => {
      expect(document.body.textContent).toContain('Choose import method');
    });
    expect(stepButton('Connect')).toBeTruthy();
    expect(stepButton('Preview')).toBeTruthy();

    requiredStepButton('Preview').click();
    await waitFor(() => {
      expect(document.body.textContent).toContain('Preview import');
      expect(buttonContaining('Start background import').disabled).toBe(false);
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

});

function buttonsNamed(label: string): HTMLButtonElement[] {
  return buttonsNamedWithin(document.body, label);
}

function buttonsNamedWithin(container: ParentNode, label: string): HTMLButtonElement[] {
  return Array.from(container.querySelectorAll<HTMLButtonElement>('button')).filter((button) => button.textContent?.trim() === label);
}

function stepButton(label: string): HTMLButtonElement | undefined {
  return Array.from(document.body.querySelectorAll<HTMLButtonElement>('.step-progress button')).find((button) => {
    const accessibleLabel = button.getAttribute('aria-label') ?? '';
    return accessibleLabel.startsWith(`${label}, `) || accessibleLabel.startsWith(`Go to ${label}, `);
  });
}

function requiredStepButton(label: string): HTMLButtonElement {
  const button = stepButton(label);
  if (!button) {
    throw new Error(`Missing reachable step ${label}`);
  }
  return button;
}
