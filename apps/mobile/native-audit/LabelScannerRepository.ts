import { LabelFailure, type LabelAsset, type LabelReference } from '../src/application/labels/LabelWorkspace';

export const scannerInstance = '01ARZ3NDEKTSV4RRFFQ69G5FAV';
export const scannerLabel = '01ARZ3NDEKTSV4RRFFQ69G5FAW';
export const retryLabel = '01ARZ3NDEKTSV4RRFFQ69G5FAX';
export const delayedLabel = '01ARZ3NDEKTSV4RRFFQ69G5FAY';
export const scannerLink = `https://old.example/l/v1/${scannerInstance}/${scannerLabel}`;
export const scannerTarget: LabelAsset = { tenantId: 'scanner-household', inventoryId: 'scanner-inventory', assetId: 'scanner-drill', archived: false };

/** Controlled repository: a transport may complete after abort, while OpenLabel must reject it. */
export function createScannerRepository(changed: () => void) {
  let retryFailures = 1; let release: (() => void) | undefined;
  const evidence = { reads: 0, lateReplies: 0, cancelled: false, selected: 0 };
  return {
    evidence,
    async instance() { return scannerInstance; },
    async resolve(reference: LabelReference, signal: AbortSignal) {
      evidence.reads++; changed();
      if (reference.instanceId !== scannerInstance || ![scannerLabel, retryLabel, delayedLabel].includes(reference.labelId)) throw new LabelFailure('invalid_label');
      if (reference.labelId === retryLabel && retryFailures-- > 0) throw new Error('Controlled temporary transport failure');
      if (reference.labelId === delayedLabel) await new Promise<void>(done => {
        release = () => { evidence.lateReplies++; changed(); done(); };
        signal.addEventListener('abort', () => { evidence.cancelled = true; changed(); }, { once: true });
      });
      return scannerTarget;
    },
    complete() { const pending = release; release = undefined; pending?.(); },
    selected() { evidence.selected++; changed(); }
  };
}
