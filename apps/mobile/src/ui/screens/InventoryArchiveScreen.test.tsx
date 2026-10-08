import React, { useLayoutEffect, useSyncExternalStore } from 'react';
import { getScreenFocused, setScreenFocused, subscribeScreenFocus } from '../../test-support/navigation';
import { ExportInventoryCommand } from '../../application/exports/InventoryExport';
import { setAppStateForTest } from '../../test-support/react-native';
import { describe, expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { InventoryArchiveScreen } from './InventoryArchiveScreen';
import type { ArchiveJob, InventoryArchiveWorkspace } from '../../application/archives/InventoryArchive';

const job: ArchiveJob = { id: 'job', kind: 'restore', state: 'awaiting_approval', phase: 'validation', createdAt: '', expiresAt: '2026-10-03T00:00:00Z', photos: true, otherFiles: true };
describe('mobile archive tasks', () => {
  it('selects formats without submitting and dispatches one explicit export with inventory scope', async () => {
    const h = new MobileRenderHarness(); const formats: string[] = []; const scopes: unknown[] = [];
    const exports = new ExportInventoryCommand({ download: async (scope, format) => { scopes.push(scope); formats.push(format); return format; } }, { share: async () => {} });
    const workspace = { newRequestKey: () => 'key', files: {}, repository: {
      list: async (scope: unknown) => { expect(scope).toEqual({ tenantId: 'home' }); return { jobs: [] }; },
      create: async () => { formats.push('archive'); return { ...job, kind: 'export', state: 'queued' }; }
    } } as unknown as InventoryArchiveWorkspace;
    try {
      await h.render(<InventoryArchiveScreen workspace={workspace} scope={{ tenantId: 'home', inventoryId: 'inventory' }} exportCommand={exports} onClose={() => {}} onOpen={async () => {}} />);
      await h.press(h.byLabel('Create archive'));
      for (const [label, action] of [['JSON data', 'Export JSON'], ['CSV spreadsheet', 'Export CSV']]) {
        await h.press(h.byLabel('Format')); await h.press(h.byLabel(label));
        expect(formats).toEqual(label === 'JSON data' ? ['archive'] : ['archive', 'json']);
        await h.press(h.byLabel(action));
      }
      expect(formats).toEqual(['archive', 'json', 'csv']);
      expect(scopes).toEqual([{ tenantId: 'home', inventoryId: 'inventory' }, { tenantId: 'home', inventoryId: 'inventory' }]);
    } finally { await h.unmount(); }
  });
  it('shows household mixed activity active first by creation date and reopens without submission', async () => {
    const jobs: ArchiveJob[] = [
      { ...job, id: 'z', kind: 'export', state: 'ready', createdAt: '2026-10-07T00:00:00Z' },
      { ...job, id: 'b', kind: 'export', state: 'queued', createdAt: '2026-10-05T00:00:00Z' },
      { ...job, id: 'a', createdAt: '2026-10-06T00:00:00Z' }
    ];
    let reads = 0;
    const workspace = { newRequestKey: () => { throw new Error('Unexpected submission'); }, files: {}, repository: {
      list: async (scope: unknown) => { expect(scope).toEqual({ tenantId: 'home' }); reads++; return { jobs }; }
    } } as unknown as InventoryArchiveWorkspace;
    for (let visit = 0; visit < 2; visit++) {
      const h = new MobileRenderHarness();
      try {
        await h.render(<InventoryArchiveScreen workspace={workspace} scope={{ tenantId: 'home', inventoryId: 'inventory' }} onClose={() => {}} onOpen={async () => {}} />);
        expect(h.all().filter(node => node.props.testID?.startsWith('archive-job-')).map(node => node.props.testID)).toEqual(['archive-job-a', 'archive-job-b', 'archive-job-z']);
        expect(h.byLabel('Review restore')).toBeDefined(); expect(h.byLabel('Download')).toBeDefined();
        expect(h.allText().join(' ')).toContain('Household activity');
      } finally { await h.unmount(); }
    }
    expect(reads).toBe(2);
  });
  it('keeps Import and the selected file after picking from inventory settings', async () => {
    const h = new MobileRenderHarness(); const uploads: string[] = [];
    const workspace = { newRequestKey: () => 'key', files: { pick: async () => ({ name: 'backup.zip', uri: 'private-file', dispose() {} }) }, repository: {
      list: async () => ({ jobs: [] }), upload: async (_tenant: string, _key: string, uri: string) => { uploads.push(uri); return job; }
    } } as unknown as InventoryArchiveWorkspace;
    try {
      await h.render(<InventoryArchiveScreen workspace={workspace} scope={{ tenantId: 'home', inventoryId: 'inventory' }} onClose={() => {}} onOpen={async () => {}} />);
      await h.change(h.byType('NativeSegmentedControl'), 'Import');
      await h.press(h.byLabel('Choose archive'));
      expect(h.byType('NativeSegmentedControl')?.props.selectedIndex).toBe(1);
      expect(h.allText()).toContain('backup.zip');
      await h.press(h.byLabel('Upload and validate')); expect(uploads).toEqual(['private-file']);
    } finally { await h.unmount(); }
  });
  it('retains the direct format on failure and cancels without sharing late content or duplicate requests', async () => {
    const h = new MobileRenderHarness(); let reads = 0; let finish!: (content: string) => void;
    const formats: string[] = [], delivered: string[] = [];
    const command = new ExportInventoryCommand({ download: async (_scope, format) => {
      formats.push(format); if (++reads === 1) throw { status: 422 };
      return new Promise(resolve => { finish = resolve; });
    } }, { share: async file => { delivered.push(file.content); } });
    const workspace = { repository: { list: async () => ({ jobs: [] }) } } as unknown as InventoryArchiveWorkspace;
    try {
      await h.render(<InventoryArchiveScreen workspace={workspace} scope={{ tenantId: 'home', inventoryId: 'inventory' }} exportCommand={command} onClose={() => {}} onOpen={async () => {}} />);
      await h.press(h.byLabel('Format')); await h.press(h.byLabel('CSV spreadsheet'));
      await h.press(h.byLabel('Export CSV'));
      expect(h.allText().join(' ')).toContain('export limit');
      await h.press(h.byLabel('Export CSV')); await h.press(h.byLabel('Export CSV'));
      expect(formats).toEqual(['csv', 'csv']);
      await h.press(h.byLabel('Cancel export')); await h.run(() => finish('private content'));
      expect(delivered).toEqual([]);
      expect(h.byLabel('Export CSV')?.props.disabled).toBe(false);
    } finally { await h.unmount(); }
  });
  it.each(['scope', 'blur', 'unmount'] as const)('retires a direct export during the %s commit before private content can share', async transition => {
    const h = new MobileRenderHarness(); const delivered: string[] = [];
    let finish!: (content: string) => void; let signal!: AbortSignal; let abortedAtCommit = false;
    const command = new ExportInventoryCommand({ download: (_scope, _format, requestSignal) => {
      signal = requestSignal; return new Promise(resolve => { finish = resolve; });
    } }, { share: async file => { delivered.push(file.content); } });
    const workspace = { repository: { list: async () => ({ jobs: [] }) } } as unknown as InventoryArchiveWorkspace;
    function Host({ inventory = 'main', visible = true }: { inventory?: string; visible?: boolean }) {
      const focused = useSyncExternalStore(subscribeScreenFocus, getScreenFocused);
      useLayoutEffect(() => {
        if (signal && (inventory !== 'main' || !visible || !focused)) { abortedAtCommit = signal.aborted; finish('private content'); }
      }, [inventory, visible, focused]);
      return visible ? <InventoryArchiveScreen workspace={workspace} scope={{ tenantId: 'home', inventoryId: inventory }} exportCommand={command} onClose={() => {}} onOpen={async () => {}} /> : null;
    }
    try {
      await h.render(<Host />); await h.press(h.byLabel('Format')); await h.press(h.byLabel('JSON data')); await h.press(h.byLabel('Export JSON'));
      if (transition === 'scope') await h.render(<Host inventory="other" />);
      else if (transition === 'blur') await h.run(() => setScreenFocused(false));
      else await h.render(<Host visible={false} />);
      expect(abortedAtCommit).toBe(true); expect(delivered).toEqual([]);
    } finally { await h.unmount(); setScreenFocused(true); }
  });
  it('does not offer open or cancel while published access is finalizing', async () => {
    const h = new MobileRenderHarness();
    const workspace = { newRequestKey: () => 'key', files: {}, repository: { list: async () => ({ jobs: [{ ...job, state: 'queued', phase: 'finalization', destinationInventoryId: 'restored' }] }) } } as unknown as InventoryArchiveWorkspace;
    try {
      await h.render(<InventoryArchiveScreen workspace={workspace} scope={{ tenantId: 'home' }} onClose={() => {}} onOpen={async () => {}} />);
      expect(h.allText().join(' ')).toContain('Restoring');
      expect(h.allText().join(' ')).not.toContain('Open inventory');
      expect(h.byLabel('Recent jobs')).toBeUndefined();
      expect(h.byLabel('Open inventory')).toBeUndefined();
    } finally { await h.unmount(); }
  });
  it('reviews a validated archive before approving a new inventory', async () => {
    const h = new MobileRenderHarness(); const approved: string[] = [];
    const workspace = {
      newRequestKey: () => 'key', files: {}, repository: {
        list: async () => ({ jobs: [job] }),
        preview: async () => ({ inventoryName: 'Restored home', assets: 2, tags: 1, photos: 1, otherFiles: 0, customAssetTypes: 0, customFields: 0, omittedAttachments: 0, keyRemappings: [] }),
        approve: async (_tenant: string, _id: string, name: string) => { approved.push(name); return { ...job, state: 'queued' }; },
      }
    } as unknown as InventoryArchiveWorkspace;
    try {
      await h.render(<InventoryArchiveScreen workspace={workspace} scope={{ tenantId: 'empty-household' }} onClose={() => {}} onOpen={async () => {}} />);
      expect(approved).toEqual([]);
      await h.press(h.byLabel('Review restore'));
      expect(h.byLabel('Inventory name')?.props.value).toBe('Restored home');
      expect(approved).toEqual([]);
      await h.press(h.byLabel('Restore inventory'));
      expect(approved).toEqual(['Restored home']);
    } finally { await h.unmount(); }
  });
  it('cancels a pending inventory selection when the task closes', async () => {
    const h = new MobileRenderHarness(); let signal: AbortSignal | undefined;
    let finish!: () => void;
    const workspace = { newRequestKey: () => 'key', files: {}, repository: {
      list: async () => ({ jobs: [{ ...job, state: 'ready', destinationInventoryId: 'new' }] })
    } } as unknown as InventoryArchiveWorkspace;
    await h.render(<InventoryArchiveScreen workspace={workspace} scope={{ tenantId: 'home' }} onClose={() => {}} onOpen={async (_id, requestSignal) => {
      signal = requestSignal; await new Promise<void>(resolve => { finish = resolve; });
    }} />);
    await h.press(h.byLabel('Open inventory'));
    expect(signal?.aborted).toBe(false);
    await h.unmount(); expect(signal?.aborted).toBe(true);
    finish();
  });

  it('preserves older-job pagination when creation finishes before the initial listing', async () => {
    const h = new MobileRenderHarness(); let finish!: (value: { jobs: ArchiveJob[]; nextCursor: string }) => void;
    let reads = 0;
    const workspace = { newRequestKey: () => 'key', files: {}, repository: {
      create: async () => ({ ...job, kind: 'export', state: 'queued' }),
      list: async () => ++reads === 1 ? new Promise(resolve => { finish = resolve; }) : { jobs: [job], nextCursor: 'older' },
    } } as unknown as InventoryArchiveWorkspace;
    try {
      await h.render(<InventoryArchiveScreen workspace={workspace} scope={{ tenantId: 'home', inventoryId: 'inventory' }} onClose={() => {}} onOpen={async () => {}} />);
      await h.press(h.byLabel('Create archive'));
      await h.run(() => finish({ jobs: [], nextCursor: 'obsolete' }));
      await h.run(() => setAppStateForTest('background'));
      await h.run(() => setAppStateForTest('active'));
      expect(h.byLabel('Load more')).toBeDefined();
    } finally { await h.unmount(); setAppStateForTest('active'); }
  });

});
