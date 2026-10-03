import type { PrintingRepository } from '$lib/ports/printingRepository';
import { PrintingFailure, type PrintScope, type ReportedPrintOutcome, type PrintDefaults, type RegisteredPrinter, type PrintConnector, type PrintJob, type LabelMedia, type LabelSelection } from '$lib/domain/printing';
export const fakePrintMedia: LabelMedia = { name: '29 × 90 mm', presetId: 'brother-ql800-29x90', version: 1, widthMicrometers: 29000, heightMicrometers: 90000, marginsMicrometers: { left: 1546, right: 1546, top: 3047, bottom: 3048 }, resolutionDpi: 300, rasterWidth: 306, rasterHeight: 991, orientation: 'feed', colorMode: 'monochrome', cutPolicy: 'after_label', displayRotation: 270 };
export class FakePrintingRepository implements PrintingRepository {
    readonly scope: PrintScope = { tenantId: 'tenant', inventoryId: 'inventory' };
    canConfigure = true;
    canPrint = true;
    authenticated = true;
    loseNextJobResponse = false;
    loseNextResolutionResponse=false;
    previewBytes = new Blob(['controlled-label'], { type: 'image/png' });
    defaults: PrintDefaults = { revision: 0, defaultPrinterId: 'printer', templateId: 'qr-title', templateVersion: 1, showReference: true, printOnCreateDefault: false };
    destinations: RegisteredPrinter[] = [{ id: 'printer', name: 'Garage Brother', adapterId: 'brother-ql800', revision: 1, retired: false, media: fakePrintMedia, mediaFingerprint: 'media-v1', readiness: 'unavailable' }];
    registrations: PrintConnector[] = [{ id: 'connector', name: 'Garage computer', state: 'active', authorizationPending: false, printerIds: ['printer'], lastSeenAt: '2026-10-03T12:00:00Z' }];
    readonly queued = new Map<string, PrintJob>();
    private readonly requests = new Map<string, {
        fingerprint: string;
        jobId: string;
    }>();
    private check(scope: PrintScope) { if (!this.authenticated)
        throw new PrintingFailure('authentication_required'); if (scope.tenantId !== this.scope.tenantId || scope.inventoryId !== this.scope.inventoryId)
        throw new PrintingFailure('invalid'); }
    async printers(scope: PrintScope) { this.check(scope); return structuredClone(this.destinations); }
    async connectors(scope: PrintScope) { this.check(scope); return structuredClone(this.registrations); }
    async templates(scope: PrintScope) { this.check(scope); return [{ id: 'qr-title', version: 1, name: 'QR and title', showReferenceDefault: true }]; }
    async settings(scope: PrintScope) { this.check(scope); return structuredClone(this.defaults); }
    async saveSettings(scope: PrintScope, next: PrintDefaults) { this.check(scope); if (!this.canConfigure)
        throw new PrintingFailure('denied'); if (next.revision !== this.defaults.revision)
        throw new PrintingFailure('conflict'); if (next.printOnCreateDefault && !this.destinations.some(p => p.id === next.defaultPrinterId && !p.retired))
        throw new PrintingFailure('invalid'); this.defaults = { ...next, revision: next.revision + 1 }; return structuredClone(this.defaults); }
    async updatePrinter(scope: PrintScope, current: RegisteredPrinter, name: string, retired: boolean) { this.check(scope); if (!this.canConfigure)
        throw new PrintingFailure('denied'); const stored = this.destinations.find(p => p.id === current.id); if (!stored)
        throw new PrintingFailure('invalid'); if (stored.revision !== current.revision)
        throw new PrintingFailure('conflict'); Object.assign(stored, { name, retired, revision: stored.revision + 1 }); return structuredClone(stored); }
    async preview(scope: PrintScope, assetId: string, selection: LabelSelection, _media: LabelMedia) { this.check(scope); if (!this.canPrint)
        throw new PrintingFailure('denied'); return { bytes: this.previewBytes, selectionFingerprint: JSON.stringify({ assetId, selection }), mediaFingerprint: selection.expectedMediaFingerprint, displayRotation: 270, expiresAt: '2099-01-01T00:00:00Z' }; }
    async createJob(scope: PrintScope, assetId: string, selection: LabelSelection, previewFingerprint: string, key: string) {
        return this.enqueue(scope, assetId, selection, previewFingerprint, key);
    }
    async reprint(scope: PrintScope, predecessor: string, selection: LabelSelection, previewFingerprint: string, key: string) {
        this.check(scope);
        const prior = this.queued.get(predecessor);
        if (!prior || !prior.assetId) throw new PrintingFailure('invalid');
        if (!['completed', 'failed', 'canceled'].includes(prior.status)) throw new PrintingFailure('conflict');
        return this.enqueue(scope, prior.assetId, selection, previewFingerprint, key, predecessor);
    }
    async testPrinter(scope:PrintScope,printerId:string,selection:LabelSelection,key:string){
        this.check(scope);if(selection.printerId!==printerId||selection.copies!==1)throw new PrintingFailure('invalid');
        return this.enqueue(scope,undefined,selection,JSON.stringify({selection}),key);
    }
    private async enqueue(scope: PrintScope, assetId: string | undefined, selection: LabelSelection, previewFingerprint: string, key: string, predecessor?: string) {
        this.check(scope);
        if (!this.canPrint)
            throw new PrintingFailure('denied');
        const fingerprint = JSON.stringify({ assetId, selection, predecessor });
        if (JSON.stringify({assetId, selection}) !== previewFingerprint)
            throw new PrintingFailure('conflict');
        const prior = this.requests.get(key);
        if (prior) {
            if (prior.fingerprint !== fingerprint)
                throw new PrintingFailure('conflict');
            return structuredClone(this.queued.get(prior.jobId)!);
        }
        const destination = this.destinations.find(p => p.id === selection.printerId && !p.retired);
        if (!destination || destination.mediaFingerprint !== selection.expectedMediaFingerprint || !Number.isInteger(selection.copies) || selection.copies < 1)
            throw new PrintingFailure('conflict');
        const id = `job-${this.queued.size + 1}`;
        const job: PrintJob = { id, assetId, predecessor, printerId: selection.printerId, status: 'queued', revision: 1, copies: selection.copies, completedCopies: 0, reason: '', createdAt: '2026-10-03T12:00:00Z' };
        this.queued.set(id, job);
        this.requests.set(key, { fingerprint, jobId: id });
        if (this.loseNextJobResponse) {
            this.loseNextJobResponse = false;
            throw new PrintingFailure('unavailable');
        }
        return structuredClone(job);
    }
    async job(scope: PrintScope, id: string) { this.check(scope); const job = this.queued.get(id); if (!job)
        throw new PrintingFailure('invalid'); return structuredClone(job); }
    async jobs(scope: PrintScope) { this.check(scope); return { items: structuredClone([...this.queued.values()].reverse()) }; }
    async resolve(scope:PrintScope,current:PrintJob,outcome:ReportedPrintOutcome){
      this.check(scope);if(!this.canPrint)throw new PrintingFailure('denied');
      const job=this.queued.get(current.id);if(!job)throw new PrintingFailure('invalid');
      if(job.resolution){if(job.resolution.reportedOutcome!==outcome||job.resolution.resolvedBy!=='editor')throw new PrintingFailure('conflict');return structuredClone(job);}
      if(job.status!=='uncertain'||job.revision!==current.revision||!job.idleConfirmedAt)throw new PrintingFailure('conflict');
      Object.assign(job,{status:'failed',revision:job.revision+1,resolution:{reportedOutcome:outcome,resolvedBy:'editor',resolvedAt:'2026-10-03T12:01:00Z'}});
      if(this.loseNextResolutionResponse){this.loseNextResolutionResponse=false;throw new PrintingFailure('unavailable');}
      return structuredClone(job);
    }
    async cancel(scope: PrintScope, current: PrintJob) { this.check(scope); if (!this.canPrint)
        throw new PrintingFailure('denied'); const job = this.queued.get(current.id); if (!job || job.revision !== current.revision || !['queued', 'claimed'].includes(job.status))
        throw new PrintingFailure('conflict'); Object.assign(job, { status: 'canceled', revision: job.revision + 1 }); return structuredClone(job); }
}
