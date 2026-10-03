import { PrintRequests } from '../application/printing/PrintSubmission';
import type { PrintCatalog, PrintJob, PrintOutcome, PrintSettings, PrintingRepository, PrintingWorkspace } from '../application/printing/PrintingWorkspace';
const scope = { tenantId: 'tenant', inventoryId: 'inventory' };
export class PrintingFake implements PrintingRepository {
  settings: PrintSettings = { revision: 1, defaultPrinterId: 'printer', template: { id: 'qr-title', version: 1, showReference: true }, printOnCreateDefault: false };
  readonly submitted = new Map<string, PrintJob>(); drop = false; deny = false; releases = 0; previews = 0;
  async catalog(): Promise<PrintCatalog> {
    if (this.deny) throw new Error('Forbidden');
    return { settings: this.settings, templates: [{ id: 'qr-title', version: 1, name: 'QR and title', supportsReference: true, showReference: true }], connectors: [],
      printers: [{ id: 'printer', name: 'Garage', readiness: 'unavailable', retired: false, revision: 1, mediaName: '29 × 90 mm', mediaFingerprint: 'media', media: { widthMicrometers: 29000, heightMicrometers: 90000, margins: { left: 1, right: 1, top: 1, bottom: 1 }, resolutionDPI: 300, rasterWidth: 306, rasterHeight: 991, orientation: 'portrait', colorMode: 'monochrome', cutPolicy: 'cut', displayRotation: 90 } }] };
  }
  async saveSettings(_scope: typeof scope, value: PrintSettings) { if (value.revision !== this.settings.revision) throw new Error('Conflict'); this.settings = { ...value, revision: value.revision + 1 }; return this.settings; }
  async preview() { this.previews++; return { file: { bytes: new Uint8Array([1]), format: 'png' as const, width: 306, height: 991, rotation: 90 }, fingerprint: 'selection' }; }
  async submit(_scope: typeof scope, assetId: string, selection: { printerId: string }, key: string) {
    const job = this.submitted.get(key) ?? { id: key, assetId, printerId: selection.printerId, status: 'queued', revision: 1, copies: 1, completedCopies: 0 };
    this.submitted.set(key, job);
    if (this.drop) { this.drop = false; throw new Error('Response lost'); } return job;
  }
  async jobs() { return [...this.submitted.values()]; }
  async job(_scope: typeof scope, id: string) { const job = this.submitted.get(id); if (!job) throw new Error('Missing'); return job; }
  dropResolution = false;
  async resolve(_scope: typeof scope, job: PrintJob, outcome: PrintOutcome) {
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
  workspace(): PrintingWorkspace { return { repository: this, requests: new PrintRequests(this, () => 'request'), newRequestKey: () => 'request', files: { preview: async () => ({ uri: 'file:///private/label.png', release: () => { this.releases++; } }), deliver: async () => {} } }; }
}
