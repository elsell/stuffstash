import type { ReportedPrintOutcome, LabelTemplate, PrintDefaults, PrintScope, RegisteredPrinter, PrintConnector, PrintJob, PrintPage, LabelSelection, LabelPreview, LabelMedia } from '$lib/domain/printing';
export interface PrintingRepository {
    printers(scope: PrintScope): Promise<RegisteredPrinter[]>;
    connectors(scope: PrintScope): Promise<PrintConnector[]>;
    templates(scope: PrintScope): Promise<LabelTemplate[]>;
    settings(scope: PrintScope): Promise<PrintDefaults>;
    saveSettings(scope: PrintScope, settings: PrintDefaults): Promise<PrintDefaults>;
    updatePrinter(scope: PrintScope, printer: RegisteredPrinter, name: string, retired: boolean): Promise<RegisteredPrinter>;
    preview(scope: PrintScope, assetId: string, selection: LabelSelection, media: LabelMedia): Promise<LabelPreview>;
    createJob(scope: PrintScope, assetId: string, selection: LabelSelection, previewFingerprint: string, key: string): Promise<PrintJob>;
    job(scope: PrintScope, id: string): Promise<PrintJob>;
    jobs(scope: PrintScope, cursor?: string): Promise<PrintPage<PrintJob>>;
    resolve(scope:PrintScope,job:PrintJob,outcome:ReportedPrintOutcome):Promise<PrintJob>;
    cancel(scope: PrintScope, job: PrintJob): Promise<PrintJob>;
}
export const printingWorkspaceContext = Symbol('printingWorkspace');
export interface PrintingWorkspace {
    apiIdentity: string;
    repository: PrintingRepository;
    intents: PrintIntents;
}
export interface PrintIntent {
    readonly locked: boolean;
    readonly selection: LabelSelection | null;
    readonly rendered: LabelPreview | null;
    readonly result: PrintJob | null;
    invalidate(): void;
    preview(selection: LabelSelection, media: LabelMedia, signal?: AbortSignal): Promise<LabelPreview>;
    submit(): Promise<PrintJob>;
}
export interface PrintIntents {
    forAsset(scope: PrintScope, assetId: string): PrintIntent;
    startAnother(scope: PrintScope, assetId: string): PrintIntent;
}
