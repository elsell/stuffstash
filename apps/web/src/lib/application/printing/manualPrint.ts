import {PrinterTestRequest} from './printerTest';
import { PrintingFailure, type PrintScope, type LabelSelection, type LabelPreview, type LabelMedia, type PrintJob } from '$lib/domain/printing';
import type { PrintingRepository, PrintIntent, PrintIntents, PrinterTestIntent } from '$lib/ports/printingRepository';
export class ManualPrintRequest implements PrintIntent {
    locked = false;
    selection: LabelSelection | null = null;
    rendered: LabelPreview | null = null;
    result: PrintJob | null = null;
    private ambiguous = false;
    private defaultPrint = false;
    private previewGeneration = 0;
    private pending: Promise<PrintJob> | null = null;
    constructor(private readonly repository: PrintingRepository, private readonly scope: PrintScope, private readonly assetId: string, private readonly key: string, private readonly predecessor?: string) { }
    invalidate() { if (this.locked)
        throw new PrintingFailure('conflict'); this.defaultPrint = false; this.previewGeneration++; this.rendered = null; this.selection = null; }
    async preview(selection: LabelSelection, media: LabelMedia, signal?: AbortSignal) {
        this.invalidate();
        if (!Number.isInteger(selection.copies) || selection.copies < 1)
            throw new PrintingFailure('invalid');
        const generation = this.previewGeneration;
        const captured = { ...selection };
        const rendered = await this.repository.preview(this.scope, this.assetId, captured, media);
        if (generation !== this.previewGeneration || signal?.aborted)
            throw new PrintingFailure('conflict');
        this.selection = captured;
        this.rendered = rendered;
        return rendered;
    }
    submitDefault(selection: LabelSelection): Promise<PrintJob> {
        if (this.locked || this.result || this.pending || this.predecessor || selection.copies !== 1)
            return Promise.reject(new PrintingFailure('conflict'));
        this.invalidate();
        this.selection = {...selection};
        this.defaultPrint = true;
        return this.submit();
    }
    submit(): Promise<PrintJob> {
        if (this.result)
            return Promise.resolve(this.result);
        if (this.pending)
            return this.pending;
        if (!this.selection || (!this.rendered && !this.defaultPrint))
            return Promise.reject(new PrintingFailure('invalid'));
        this.locked = true;
        const operation = this.predecessor
            ? this.repository.reprint(this.scope, this.predecessor, this.selection, this.rendered?.selectionFingerprint ?? '', this.key)
            : this.repository.createJob(this.scope, this.assetId, this.selection, this.rendered?.selectionFingerprint ?? '', this.key);
        this.pending = operation.then(result => { this.result = result; return result; }).catch(error => {
            const definite = error instanceof PrintingFailure && ['invalid', 'conflict', 'denied', 'authentication_required'].includes(error.kind);
            if (!definite)
                this.ambiguous = true;
            if (!this.ambiguous)
                this.locked = false;
            throw error;
        }).finally(() => { this.pending = null; });
        return this.pending;
    }
}
export class SessionPrintIntents implements PrintIntents {
    private readonly tests = new Map<string, PrinterTestIntent>();
    private readonly requests = new Map<string, PrintIntent>();
    constructor(private readonly repository: PrintingRepository, private readonly newKey: () => string) { }
    forPrinterTest(scope:PrintScope,printerId:string){
        const identity=this.identity(scope,printerId);let request=this.tests.get(identity);
        if(!request){request=new PrinterTestRequest(this.repository,{...scope},printerId,this.newKey());this.tests.set(identity,request);}
        return request;
    }
    startAnotherPrinterTest(scope:PrintScope,printerId:string,priorJob:PrintJob){
        const prior=this.forPrinterTest(scope,printerId);
        if(!prior.result||prior.result.id!==priorJob.id||!['completed','failed','canceled'].includes(priorJob.status))throw new PrintingFailure('conflict');
        const request=new PrinterTestRequest(this.repository,{...scope},printerId,this.newKey());this.tests.set(this.identity(scope,printerId),request);return request;
    }
    forAsset(scope: PrintScope, assetId: string) { const key = this.identity(scope, assetId); let request = this.requests.get(key); if (!request) {
        request = this.create(scope, assetId);
        this.requests.set(key, request);
    } return request; }
    forReprint(scope: PrintScope, assetId: string, predecessor: string) {
        const key = JSON.stringify([scope.tenantId, scope.inventoryId, 'reprint', predecessor]);
        let request = this.requests.get(key);
        if (!request) {
            request = new ManualPrintRequest(this.repository, {...scope}, assetId, this.newKey(), predecessor);
            this.requests.set(key, request);
        }
        return request;
    }
    startAnother(scope: PrintScope, assetId: string) { const prior = this.forAsset(scope, assetId); if (prior.locked && !prior.result)
        throw new PrintingFailure('conflict'); const request = this.create(scope, assetId); this.requests.set(this.identity(scope, assetId), request); return request; }
    private create(scope: PrintScope, assetId: string) { return new ManualPrintRequest(this.repository, { ...scope }, assetId, this.newKey()); }
    private identity(scope: PrintScope, assetId: string) { return JSON.stringify([scope.tenantId, scope.inventoryId, assetId]); }
}
