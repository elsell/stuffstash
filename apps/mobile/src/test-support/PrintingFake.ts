import { PrintRequests } from '../application/printing/PrintSubmission';
import type { PrintCatalog, PrintJob, PrintMediaPreset, RegisteredPrinter, PrintOutcome, PrintSelection, PrintSettings, PrintingRepository, PrintingWorkspace } from '../application/printing/PrintingWorkspace';
const scope = { tenantId: 'tenant', inventoryId: 'inventory' };
export class PrintingFake implements PrintingRepository {
  printer: RegisteredPrinter = { id: 'printer', adapterId: 'brother-ql800', name: 'Garage', readiness: 'unavailable', retired: false, revision: 1, mediaName: '29 × 90 mm', mediaFingerprint: 'media', media: { presetId: 'brother-ql800-29x90', version: 1, widthMicrometers: 29000, heightMicrometers: 89800, margins: { left: 1524, right: 1524, top: 2963, bottom: 2963 }, resolutionDPI: 300, rasterWidth: 306, rasterHeight: 991, orientation: 'feed', colorMode: 'monochrome', cutPolicy: 'after_label', displayRotation: 270 } };
  settings: PrintSettings = { revision: 1, defaultPrinterId: 'printer', template: { id: 'qr-title', version: 1, showReference: true }, printOnCreateDefault: false };
  readonly submitted = new Map<string, PrintJob>(); drop = false; deny = false; releases = 0; previews = 0;
  async catalog(): Promise<PrintCatalog> {
    if (this.deny) throw new Error('Forbidden');
    return { settings: this.settings, templates: [{ id: 'qr-title', version: 1, name: 'QR and title', supportsReference: true, showReference: true }], connectors: [],
      printers: [this.printer] };
  }
  async mediaPresets(selectedScope: typeof scope, printer: RegisteredPrinter) {
    if (this.deny || selectedScope.tenantId !== scope.tenantId || selectedScope.inventoryId !== scope.inventoryId) throw new Error('Forbidden');
    return printer.adapterId === this.printer.adapterId ? [{ id: 'brother-ql800-29x90', version: 1, name: '29 × 90 mm', widthMicrometers: 29000, heightMicrometers: 89800 }] : [];
  }
  async configurePrinter(selectedScope: typeof scope, printer: RegisteredPrinter, preset: PrintMediaPreset) {
    if (this.deny || selectedScope.tenantId !== scope.tenantId || selectedScope.inventoryId !== scope.inventoryId) throw new Error('Forbidden');
    if (printer.id !== this.printer.id || printer.revision !== this.printer.revision || preset.id !== 'brother-ql800-29x90' || preset.version !== 1) throw new Error('Conflict');
    this.printer = { ...this.printer, revision: this.printer.revision + 1, mediaName: preset.name, media: { ...this.printer.media, presetId: preset.id, version: preset.version } };
    return this.printer;
  }
  async saveSettings(_scope: typeof scope, value: PrintSettings) { if (value.revision !== this.settings.revision) throw new Error('Conflict'); this.settings = { ...value, revision: value.revision + 1 }; return this.settings; }
  async preview() { this.previews++; return { file: { bytes: new Uint8Array([1]), format: 'png' as const, width: 306, height: 991, rotation: 90 }, fingerprint: 'selection' }; }
  async submit(_scope: typeof scope, assetId: string, selection: { printerId: string }, key: string) {
    const job = this.submitted.get(key) ?? { id: key, assetId, printerId: selection.printerId, status: 'queued', revision: 1, copies: 1, completedCopies: 0 };
    this.submitted.set(key, job);
    if (this.drop) { this.drop = false; throw new Error('Response lost'); } return job;
  }
  readonly queuedSelections = new Map<string, PrintSelection>();
  private readonly printIntents = new Map<string, string>();
  async reprint(_scope: typeof scope, predecessorId: string, selection: PrintSelection, key: string) {
    if (this.deny) throw new Error('Forbidden');
    const fingerprint = JSON.stringify(['reprint', _scope, predecessorId, selection]);
    if (this.printIntents.has(key) && this.printIntents.get(key) !== fingerprint) throw new Error('Conflict');
    if (this.printIntents.has(key)) return this.submitted.get(key)!;
    const predecessor = this.submitted.get(predecessorId);
    if (!predecessor || !['completed', 'failed', 'canceled'].includes(predecessor.status)) throw new Error('Conflict');
    const job = { id: key, predecessor: predecessorId, assetId: predecessor.assetId, kind: predecessor.kind, printerId: selection.printerId, status: 'queued', revision: 1, copies: selection.copies, completedCopies: 0 };
    this.printIntents.set(key, fingerprint); this.queuedSelections.set(key, selection); this.submitted.set(key, job);
    if (this.drop) { this.drop = false; throw new Error('Response lost'); } return job;
  }
  async test(_scope: typeof scope, printerId: string, selection: PrintSelection, key: string) {
    if (this.deny || printerId !== selection.printerId || selection.copies !== 1) throw new Error('Rejected');
    const fingerprint = JSON.stringify(['test', _scope, printerId, selection]);
    if (this.printIntents.has(key) && this.printIntents.get(key) !== fingerprint) throw new Error('Conflict');
    if (this.printIntents.has(key)) return this.submitted.get(key)!;
    const job = { id: key, kind: 'printer_test', printerId, status: 'queued', revision: 1, copies: 1, completedCopies: 0 };
    this.printIntents.set(key, fingerprint); this.queuedSelections.set(key, selection); this.submitted.set(key, job);
    if (this.drop) { this.drop = false; throw new Error('Response lost'); } return job;
  }
  async jobs() { return [...this.submitted.values()]; }
  async job(_scope: typeof scope, id: string) { const job = this.submitted.get(id); if (!job) throw new Error('Missing'); return job; }
  dropResolution = false; loseResolutionBeforeCommit = false;
  async resolve(_scope: typeof scope, job: PrintJob, outcome: PrintOutcome) {
    if (this.loseResolutionBeforeCommit) { this.loseResolutionBeforeCommit = false; throw new Error('Connection lost'); }
    if (this.deny) throw new Error('Forbidden');
    const current = this.submitted.get(job.id);
    if (current?.resolution?.reportedOutcome === outcome) return current;
    if (!current || current.revision !== job.revision || current.status !== 'uncertain' || !current.idleConfirmed) throw new Error('Conflict');
    const next = { ...current, status: 'failed', revision: current.revision + 1, resolution: { reportedOutcome: outcome } };
    this.submitted.set(job.id, next);
    if (this.dropResolution) { this.dropResolution = false; throw new Error('Response lost'); }
    return next;
  }
  async cancel(_scope: typeof scope, job: PrintJob) { const next = { ...job, status: 'canceled', revision: job.revision + 1 }; this.submitted.set(job.id, next); return next; }
  workspace(): PrintingWorkspace { let count = 0; const newKey = () => ++count === 1 ? 'request' : `request-${count}`; return { repository: this, requests: new PrintRequests(this, newKey), newRequestKey: newKey, files: { preview: async () => ({ uri: 'file:///private/label.png', release: () => { this.releases++; } }), deliver: async () => {} } }; }
}
