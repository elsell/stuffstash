import type { PrintScope, PrintSelection, PrintingRepository } from './PrintingWorkspace';
export class PrintRequestRejected extends Error { constructor() { super('Print request was not accepted'); } }
/** Owns a single explicit print intent, including retries after an unknown response. */
export class PrintSubmission {
  private request?: { identity: string; key: string; selection: PrintSelection; scope: PrintScope; assetId: string; ambiguous: boolean };
  private running = false;
  constructor(private readonly repository: Pick<PrintingRepository, 'submit'>, private readonly newKey: () => string) {}
  get selection() { return this.request?.selection; }
  get locked() { return !!this.request; }
  async retry() {
    if (!this.request) throw new Error('No submitted request');
    return this.submit(this.request.scope, this.request.assetId, this.request.selection);
  }
  async submit(scope: PrintScope, assetId: string, selection: PrintSelection) {
    if (this.running) throw new Error('Print submission in progress');
    const identity = JSON.stringify([scope.tenantId, scope.inventoryId, assetId, selection]);
    if (this.request && this.request.identity !== identity) throw new Error('Resolve the submitted request before changing it');
    this.request ??= { identity, key: this.newKey(), selection: { ...selection, template: { ...selection.template } }, scope: { ...scope }, assetId, ambiguous: false };
    this.running = true;
    try { return await this.repository.submit(scope, assetId, this.request.selection, this.request.key); }
    catch (error) {
      if (error instanceof PrintRequestRejected && !this.request.ambiguous) this.request = undefined;
      else this.request.ambiguous = true;
      throw error;
    }
    finally { this.running = false; }
  }
}

/** Session-owned intents survive leaving a task; sign-out discards this composition. */
export class PrintRequests {
  private readonly requests = new Map<string, PrintSubmission>();
  constructor(private readonly repository: Pick<PrintingRepository, 'submit' | 'reprint' | 'test'>, private readonly newKey: () => string) {}
  forAsset(scope: PrintScope, assetId: string) { return this.forTarget(scope, assetId, 'asset'); }
  forReprint(scope: PrintScope, predecessorId: string) { return this.forTarget(scope, predecessorId, 'reprint'); }
  forTest(scope: PrintScope, printerId: string) { return this.forTarget(scope, printerId, 'test'); }
  private forTarget(scope: PrintScope, targetId: string, kind: 'asset' | 'reprint' | 'test') {
    const key = JSON.stringify([scope.tenantId, scope.inventoryId, kind, targetId]);
    let request = this.requests.get(key);
    if (!request) {
      const submit: PrintingRepository['submit'] = (selectedScope, selectedTarget, selection, requestKey) => kind === 'asset'
        ? this.repository.submit(selectedScope, selectedTarget, selection, requestKey)
        : kind === 'reprint' ? this.repository.reprint(selectedScope, selectedTarget, selection, requestKey)
        : this.repository.test(selectedScope, selectedTarget, selection, requestKey);
      request = new PrintSubmission({ submit }, this.newKey); this.requests.set(key, request);
    }
    return request;
  }
  settled(scope: PrintScope, targetId: string, kind: 'asset' | 'reprint' | 'test' = 'asset') {
    this.requests.delete(JSON.stringify([scope.tenantId, scope.inventoryId, kind, targetId]));
  }
}
