import React from 'react';
import { setAppStateForTest } from '../../test-support/react-native';
import { describe, expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { InventoryArchiveScreen } from './InventoryArchiveScreen';
import type { ArchiveJob, InventoryArchiveWorkspace } from '../../application/archives/InventoryArchive';

const job: ArchiveJob = { id: 'job', kind: 'restore', state: 'awaiting_approval', phase: 'validation', createdAt: '', expiresAt: '2026-10-03T00:00:00Z', photos: true, otherFiles: true };
describe('mobile archive tasks', () => {
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
