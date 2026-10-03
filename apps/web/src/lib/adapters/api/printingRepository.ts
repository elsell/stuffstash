import { PrintingClient, StuffStashAPIError, type TokenProvider } from '@stuff-stash/api-client';
import type { PrintingRepository } from '$lib/ports/printingRepository';
import { PrintingFailure, type PrintScope, type PrintDefaults, type RegisteredPrinter, type PrintConnector, type PrintJob, type LabelSelection, type LabelMedia, type LabelPreview } from '$lib/domain/printing';
export class ApiPrintingRepository implements PrintingRepository {
    private readonly client: PrintingClient;
    constructor(baseUrl: string, tokenProvider: TokenProvider, fetchImpl?: typeof fetch) { this.client = new PrintingClient({ baseUrl, tokenProvider, fetch: fetchImpl }); }
    async printers(scope: PrintScope): Promise<RegisteredPrinter[]> { return guarded(async () => (await collect(cursor => this.client.printers(scope, cursor))).map(mapPrinter)); }
    async connectors(scope: PrintScope): Promise<PrintConnector[]> { return guarded(async () => (await collect(cursor => this.client.connectors(scope, cursor))).map(c => ({ id: c.id, name: c.name, state: c.state, authorizationPending: c.authorizationPending, lastSeenAt: c.lastSeenAt, printerIds: c.printerIds ?? [] }))); }
    async templates(scope: PrintScope) { return guarded(async () => (await this.client.templates(scope)).map(t => ({ id: t.id, version: t.version, name: t.name, showReferenceDefault: t.defaults.show_reference }))); }
    async settings(scope: PrintScope) { return guarded(async () => mapSettings(await this.client.settings(scope))); }
    async saveSettings(scope: PrintScope, s: PrintDefaults) { return guarded(async () => mapSettings(await this.client.saveSettings(scope, { revision: s.revision, defaultPrinterId: s.defaultPrinterId, template: { id: s.templateId, version: s.templateVersion, options: { showReference: s.showReference } }, printOnCreateDefault: s.printOnCreateDefault }))); }
    async updatePrinter(scope: PrintScope, printer: RegisteredPrinter, name: string, retired: boolean) { return guarded(async () => mapPrinter(await this.client.updatePrinter(scope, printer.id, { revision: printer.revision, name, retired }))); }
    async preview(scope: PrintScope, assetId: string, selection: LabelSelection, media: LabelMedia): Promise<LabelPreview> {
        return guarded(async () => {
            const rendered = await this.client.render(scope, assetId, { media: { preset_id: media.presetId, version: media.version, width_micrometers: media.widthMicrometers, height_micrometers: media.heightMicrometers, margins_micrometers: media.marginsMicrometers, resolution_dpi: media.resolutionDpi, raster_width: media.rasterWidth, raster_height: media.rasterHeight, orientation: media.orientation, color_mode: media.colorMode, cut_policy: media.cutPolicy, display_rotation: media.displayRotation }, template: { id: selection.templateId, version: selection.templateVersion, options: { show_reference: selection.showReference } }, format: 'png' });
            if (rendered.mediaFingerprint !== selection.expectedMediaFingerprint)
                throw new PrintingFailure('conflict');
            const bytes = await this.client.content(scope, rendered.id);
            return { bytes, selectionFingerprint: rendered.selectionFingerprint, mediaFingerprint: rendered.mediaFingerprint, displayRotation: rendered.displayRotation, expiresAt: rendered.expiresAt };
        });
    }
    async createJob(scope: PrintScope, assetId: string, s: LabelSelection, previewFingerprint: string, key: string) { return guarded(async () => mapJob(await this.client.createJob(scope, assetId, { printerId: s.printerId, expectedMediaFingerprint: s.expectedMediaFingerprint, templateId: s.templateId, templateVersion: s.templateVersion, templateOptions: { showReference: s.showReference }, copies: s.copies, previewFingerprint }, key))); }
    async job(scope: PrintScope, id: string) { return guarded(async () => mapJob(await this.client.job(scope, id))); }
    async jobs(scope: PrintScope, cursor?: string) { return guarded(async () => { const page = await this.client.jobs(scope, cursor); return { items: page.items.map(mapJob), nextCursor: page.nextCursor }; }); }
    async cancel(scope: PrintScope, job: PrintJob) { return guarded(async () => mapJob(await this.client.cancel(scope, job.id, job.revision))); }
}
function mapPrinter(p: Awaited<ReturnType<PrintingClient['printers']>>['items'][number]): RegisteredPrinter {
    if (!['ready', 'unknown', 'unavailable', 'error'].includes(p.readiness))
        throw new PrintingFailure('unavailable');
    return { id: p.id, name: p.name, adapterId: p.adapterId, revision: p.revision, retired: p.retired, media: { ...p.media }, mediaFingerprint: p.mediaFingerprint, readiness: p.readiness as RegisteredPrinter['readiness'] };
}
function mapSettings(s: Awaited<ReturnType<PrintingClient['settings']>>): PrintDefaults { return { revision: s.revision, defaultPrinterId: s.defaultPrinterId, templateId: s.template.id, templateVersion: s.template.version, showReference: s.template.options.showReference, printOnCreateDefault: s.printOnCreateDefault }; }
function mapJob(j: Awaited<ReturnType<PrintingClient['job']>>): PrintJob {
    if (!['queued', 'claimed', 'printing', 'completed', 'failed', 'uncertain', 'canceled'].includes(j.status))
        throw new PrintingFailure('unavailable');
    const last = j.attempts?.at(-1);
    return { id: j.id, printerId: j.printerId, assetId: j.assetId, status: j.status as PrintJob['status'], revision: j.revision, copies: j.copies, completedCopies: last?.completedCopies ?? 0, reason: last?.reason ?? '', createdAt: j.createdAt };
}
async function collect<T>(next: (cursor?: string) => Promise<{
    items: T[];
    nextCursor?: string;
}>): Promise<T[]> { const items: T[] = []; const seen = new Set<string>(); let cursor: string | undefined; do {
    const page = await next(cursor);
    items.push(...page.items);
    cursor = page.nextCursor;
    if (cursor && seen.has(cursor))
        throw new PrintingFailure('unavailable');
    if (cursor)
        seen.add(cursor);
} while (cursor); return items; }
async function guarded<T>(operation: () => Promise<T>): Promise<T> { try {
    return await operation();
}
catch (error) {
    if (error instanceof PrintingFailure)
        throw error;
    if (error instanceof StuffStashAPIError) {
        if (error.status === 401)
            throw new PrintingFailure('authentication_required');
        if (error.status === 403)
            throw new PrintingFailure('denied');
        if (error.status === 409)
            throw new PrintingFailure('conflict');
        if ([400, 404, 410, 422].includes(error.status))
            throw new PrintingFailure('invalid');
    }
    throw new PrintingFailure('unavailable');
} }
