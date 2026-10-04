import { expect, it } from 'vitest';
import { OpenLabel } from '../src/application/labels/OpenLabel';
import { createScannerRepository, scannerInstance, delayedLabel, retryLabel } from './LabelScannerRepository';
it('models late cancelled lookup and recoverable transport failure without selecting stale scope', async () => {
  const repository = createScannerRepository(() => {});
  const open = new OpenLabel(repository, async () => repository.selected());
  const controller = new AbortController();
  const pending = open.execute({ instanceId: scannerInstance, labelId: delayedLabel }, controller.signal);
  const failure = expect(pending).rejects.toThrow();
  await Promise.resolve(); await Promise.resolve(); controller.abort(); repository.complete(); await failure;
  expect(repository.evidence).toMatchObject({ cancelled: true, lateReplies: 1, selected: 0 });
  const reference = { instanceId: scannerInstance, labelId: retryLabel };
  await expect(open.execute(reference, new AbortController().signal)).rejects.toThrow('Controlled temporary');
  await expect(open.execute(reference, new AbortController().signal)).resolves.toMatchObject({ assetId: 'scanner-drill' });
  expect(repository.evidence.selected).toBe(1);
});
