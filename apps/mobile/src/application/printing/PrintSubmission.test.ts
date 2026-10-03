import { expect, it } from 'vitest';
import { PrintRequestRejected, PrintSubmission } from './PrintSubmission';
import type { PrintSelection, PrintingRepository, PrintJob } from './PrintingWorkspace';

it('retains one immutable request after a lost response and never submits a second label', async () => {
  const jobs = new Map<string, PrintJob>(); let drop = true;
  const repository: Pick<PrintingRepository, 'submit'> = { async submit(_scope, assetId, selection, key) {
    const found = jobs.get(key);
    if (found) return found;
    const job = { id: key, assetId, printerId: selection.printerId, status: 'queued', revision: 1, copies: 1, completedCopies: 0 };
    jobs.set(key, job);
    if (drop) { drop = false; throw new Error('Connection lost after commit'); }
    return job;
  } };
  const task = new PrintSubmission(repository, () => 'request');
  const scope = { tenantId: 'tenant', inventoryId: 'inventory' };
  const selection: PrintSelection = { printerId: 'printer', mediaFingerprint: 'media', template: { id: 'qr-title', version: 1, showReference: true }, copies: 1 };
  await expect(task.submit(scope, 'asset', selection)).rejects.toThrow();
  expect(task.locked).toBe(true);
  await expect(task.submit(scope, 'asset', { ...selection, printerId: 'other' })).rejects.toThrow();
  expect((await task.submit(scope, 'asset', selection)).id).toBe('request');
  expect(jobs.size).toBe(1);
});

it('allows correcting a definite first rejection but keeps identity after any ambiguous attempt', async () => {
  const keys: string[] = []; let count = 0; let failure: Error | undefined = new PrintRequestRejected();
  const repository: Pick<PrintingRepository, 'submit'> = { async submit(_scope, assetId, selection, key) {
    keys.push(key); if (failure) throw failure;
    return { id: key, assetId, printerId: selection.printerId, status: 'queued', revision: 1, copies: 1, completedCopies: 0 };
  } };
  const scope = { tenantId: 'tenant', inventoryId: 'inventory' };
  const selection: PrintSelection = { printerId: 'printer', mediaFingerprint: 'media', template: { id: 'qr-title', version: 1, showReference: true }, copies: 1 };
  const task = new PrintSubmission(repository, () => `request-${++count}`);
  await expect(task.submit(scope, 'asset', selection)).rejects.toThrow(); expect(task.locked).toBe(false);
  failure = new Error('Connection lost');
  await expect(task.submit(scope, 'asset', { ...selection, mediaFingerprint: 'corrected' })).rejects.toThrow();
  failure = new PrintRequestRejected(); await expect(task.retry()).rejects.toThrow(); expect(task.locked).toBe(true);
  failure = undefined; await task.retry();
  expect(keys).toEqual(['request-1', 'request-2', 'request-2', 'request-2']);
});
