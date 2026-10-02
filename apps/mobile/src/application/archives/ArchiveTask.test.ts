import { describe, expect, it } from 'vitest';
import { ArchiveTask } from './ArchiveTask';
import type { ArchiveJob, InventoryArchiveWorkspace } from './InventoryArchive';

const job: ArchiveJob = { id: 'job', kind: 'export', state: 'queued', phase: 'execution', createdAt: '', expiresAt: '', photos: true, otherFiles: true };
function setup() {
  const keys: string[] = []; let attempts = 0; let disposed = 0;
  const workspace = {
    newRequestKey: () => `key-${++attempts}`,
    repository: {
      create: async (_scope: unknown, key: string) => { keys.push(key); if (keys.length === 1) throw new Error('lost response'); return job; },
      upload: async (_tenant: string, key: string) => { keys.push(key); if (keys.length === 1) throw new Error('lost response'); return { ...job, kind: 'restore' as const }; },
    },
    files: { pick: async () => ({ name: 'backup.zip', uri: 'file:///backup.zip', dispose: () => { disposed++; } }) },
  } as unknown as InventoryArchiveWorkspace;
  return { workspace, keys, disposed: () => disposed };
}
describe('archive task lifecycle', () => {
  it('reuses a creation key after a lost response but starts a fresh intent after success', async () => {
    const fake = setup(), task = new ArchiveTask(fake.workspace, { tenantId: 'tenant', inventoryId: 'inventory' });
    await expect(task.create({ photos: true, otherFiles: true })).rejects.toThrow('lost response');
    expect(await task.create({ photos: true, otherFiles: true })).toEqual(job);
    await task.create({ photos: true, otherFiles: true });
    expect(fake.keys).toEqual(['key-1', 'key-1', 'key-2']);
  });
  it('keeps an upload source for retry and disposes the owned copy after success', async () => {
    const fake = setup(), task = new ArchiveTask(fake.workspace, { tenantId: 'tenant' });
    await task.pick();
    await expect(task.upload()).rejects.toThrow('lost response');
    expect(fake.disposed()).toBe(0);
    await task.upload(); expect(fake.keys).toEqual(['key-1', 'key-1']); expect(fake.disposed()).toBe(1);
    task.close(); expect(fake.disposed()).toBe(1);
  });
  it('retires local work and disposes a picker result that arrives after closing', async () => {
    const fake = setup(); let resolve!: (value: { name: string; uri: string; dispose(): void }) => void;
    let disposed = false;
    fake.workspace.files.pick = async () => new Promise(done => { resolve = done; });
    const task = new ArchiveTask(fake.workspace, { tenantId: 'tenant' });
    const picking = task.pick(); task.close();
    resolve({ name: 'late.zip', uri: 'file:///late.zip', dispose: () => { disposed = true; } });
    await expect(picking).rejects.toMatchObject({ name: 'AbortError' }); expect(disposed).toBe(true);
    await expect(task.upload()).rejects.toMatchObject({ name: 'AbortError' });
  });
  it('keeps the upload file until cancelled native work has settled', async () => {
    const fake = setup(); let reject!: (error: Error) => void;
    fake.workspace.repository.upload = async () => new Promise((_done, fail) => { reject = fail; });
    const task = new ArchiveTask(fake.workspace, { tenantId: 'tenant' });
    await task.pick(); const upload = task.upload(); task.close();
    expect(fake.disposed()).toBe(0);
    reject(Object.assign(new Error('Cancelled'), { name: 'AbortError' }));
    await expect(upload).rejects.toMatchObject({ name: 'AbortError' });
    expect(fake.disposed()).toBe(1);
  });

});
