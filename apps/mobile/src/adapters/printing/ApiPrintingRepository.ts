import { PrintRequestRejected } from '../../application/printing/PrintSubmission';
import { PrintingClient, LabelsClient, StuffStashAPIError } from '@stuff-stash/api-client';
import type { PrintCatalog, PrintJob, PrintOutcome, PrintScope, PrintSelection, PrintSettings, PrintTemplate, PrintingRepository, RegisteredPrinter } from '../../application/printing/PrintingWorkspace';
import { assertReadActive } from '../../application/shared/ReadRequest';
import { labelBlobBytes } from '../labels/LabelBlobBytes';

type WireSettings = Awaited<ReturnType<PrintingClient['settings']>>;
type WireJob = Awaited<ReturnType<PrintingClient['job']>>;
export class ApiPrintingRepository implements PrintingRepository {
  constructor(private readonly client: PrintingClient, private readonly labels: LabelsClient) {}
  async catalog(scope: PrintScope, signal: AbortSignal): Promise<PrintCatalog> {
    const [printers, connectors, templates, settings] = await Promise.all([
      allPages(cursor => this.client.printers(scope, cursor), signal),
      allPages(cursor => this.client.connectors(scope, cursor), signal),
      this.client.templates(scope), this.client.settings(scope)
    ]);
    assertReadActive(signal);
    return { printers: printers.map(printer => ({ id: printer.id, name: printer.name, retired: printer.retired, readiness: printer.readiness, readinessReason: printer.readinessReason, reportedAt: printer.reportedAt, revision: printer.revision,
      mediaFingerprint: printer.mediaFingerprint, mediaName: printer.media.name, media: { presetId: printer.media.presetId, version: printer.media.version, widthMicrometers: printer.media.widthMicrometers,
        heightMicrometers: printer.media.heightMicrometers, margins: printer.media.marginsMicrometers, resolutionDPI: printer.media.resolutionDpi, rasterWidth: printer.media.rasterWidth,
        rasterHeight: printer.media.rasterHeight, orientation: printer.media.orientation, colorMode: printer.media.colorMode, cutPolicy: printer.media.cutPolicy, displayRotation: printer.media.displayRotation } })),
      connectors: connectors.map(connector => ({ id: connector.id, name: connector.name, state: connector.state, availability: connector.availability, lastSeenAt: connector.lastSeenAt, printerIds: connector.printerIds ?? [] })),
      templates: templates.map(template => ({ id: template.id, version: template.version, name: template.name, showReference: template.defaults.show_reference, supportsReference: (template.options ?? []).includes('show_reference') })), settings: mapSettings(settings) };
  }
  async saveSettings(scope: PrintScope, settings: PrintSettings) {
    return mapSettings(await this.client.saveSettings(scope, { ...settings, template: { id: settings.template.id, version: settings.template.version, options: { showReference: settings.template.showReference } } }));
  }
  async preview(scope: PrintScope, assetId: string, printer: RegisteredPrinter, template: PrintTemplate, signal: AbortSignal) {
    await this.labels.provision(scope.tenantId, scope.inventoryId, assetId, signal); assertReadActive(signal);
    const media = printer.media;
    const rendered = await this.client.render(scope, assetId, { format: 'png', template: { id: template.id, version: template.version, options: { show_reference: template.showReference } },
      media: { preset_id: media.presetId, version: media.version, width_micrometers: media.widthMicrometers, height_micrometers: media.heightMicrometers,
        margins_micrometers: media.margins, resolution_dpi: media.resolutionDPI, raster_width: media.rasterWidth, raster_height: media.rasterHeight, orientation: media.orientation,
        color_mode: media.colorMode, cut_policy: media.cutPolicy, display_rotation: media.displayRotation } });
    assertReadActive(signal);
    const bytes = await labelBlobBytes(await this.client.content(scope, rendered.id)); assertReadActive(signal);
    return { fingerprint: rendered.selectionFingerprint, file: { bytes, format: 'png' as const, width: rendered.widthPixels, height: rendered.heightPixels, rotation: rendered.displayRotation } };
  }
  async submit(scope: PrintScope, assetId: string, selection: PrintSelection, key: string) {
    return requestJob(() => this.client.createJob(scope, assetId, wireSelection(selection), key));
  }
  async reprint(scope: PrintScope, predecessorId: string, selection: PrintSelection, key: string) {
    return requestJob(() => this.client.reprint(scope, predecessorId, wireSelection(selection), key));
  }
  async test(scope: PrintScope, printerId: string, selection: PrintSelection, key: string) {
    return requestJob(() => this.client.testJob(scope, printerId, wireSelection(selection), key));
  }
  async job(scope: PrintScope, jobId: string, signal: AbortSignal) { const job = await this.client.job(scope, jobId); assertReadActive(signal); return mapJob(job); }
  async jobs(scope: PrintScope, signal: AbortSignal) { const page = await this.client.jobs(scope); assertReadActive(signal); return page.items.map(mapJob); }
  async resolve(scope: PrintScope, job: PrintJob, outcome: PrintOutcome) { return mapJob(await this.client.resolve(scope, job.id, { revision: job.revision, acknowledgeUncertainty: true, reportedOutcome: outcome })); }
  async cancel(scope: PrintScope, job: PrintJob) { return mapJob(await this.client.cancel(scope, job.id, job.revision)); }
}
function mapSettings(value: WireSettings): PrintSettings { return { revision: value.revision, defaultPrinterId: value.defaultPrinterId, printOnCreateDefault: value.printOnCreateDefault,
  template: { id: value.template.id, version: value.template.version, showReference: value.template.options.showReference } }; }
function mapJob(value: WireJob): PrintJob { return { id: value.id, assetId: value.assetId, kind: value.kind, predecessor: value.predecessor, printerId: value.printerId, status: value.status, revision: value.revision, copies: value.copies,
  latestAttemptId: value.attempts?.at(-1)?.id, idleConfirmed: !!value.attempts?.at(-1)?.idleConfirmedAt, resolution: value.resolution ? { reportedOutcome: reportedOutcome(value.resolution.reportedOutcome) } : undefined,
  completedCopies: Math.max(0, ...(value.attempts ?? []).map(attempt => attempt.completedCopies)) }; }
async function allPages<T>(read: (cursor?: string) => Promise<{ items: T[]; nextCursor?: string }>, signal: AbortSignal) {
  const items: T[] = []; const seen = new Set<string>(); let cursor: string | undefined;
  do { assertReadActive(signal); const page = await read(cursor); assertReadActive(signal); items.push(...page.items); cursor = page.nextCursor;
    if (cursor) { if (seen.has(cursor)) throw new Error('Invalid printing pagination'); seen.add(cursor); }
  } while (cursor);
  return items;
}

function reportedOutcome(value: string): PrintOutcome { return value === "printed" || value === "not_printed" ? value : "unknown"; }

function wireSelection(selection: PrintSelection) {
  return { printerId: selection.printerId, expectedMediaFingerprint: selection.mediaFingerprint, copies: selection.copies,
    templateId: selection.template.id, templateVersion: selection.template.version, templateOptions: { showReference: selection.template.showReference }, previewFingerprint: selection.previewFingerprint };
}
async function requestJob(send: () => Promise<WireJob>) {
  try { return mapJob(await send()); }
  catch (error) { if (error instanceof StuffStashAPIError && [400, 409, 422].includes(error.status)) throw new PrintRequestRejected(); throw error; }
}
