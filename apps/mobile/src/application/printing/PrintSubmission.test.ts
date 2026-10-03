import { expect, it } from 'vitest';
import { PrintSubmission } from './PrintSubmission';
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
