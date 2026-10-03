import type { PrintRequests } from './PrintSubmission';
import type { LabelFile, LabelFiles, LabelMedia, LabelScope, LabelTemplate } from '../labels/LabelWorkspace';
export type PrintScope = LabelScope;
export type PrintTemplate = { readonly id: string; readonly version: number; readonly showReference: boolean };
export type PrintSettings = { readonly revision: number; readonly defaultPrinterId: string | null; readonly template: PrintTemplate; readonly printOnCreateDefault: boolean };
export type PrintMediaPreset = { readonly id: string; readonly version: number; readonly name: string; readonly widthMicrometers: number; readonly heightMicrometers: number };
export type RegisteredPrinter = { readonly id: string; readonly adapterId: string; readonly name: string; readonly revision: number; readonly retired: boolean; readonly readiness: string; readonly readinessReason?: string; readonly reportedAt?: string; readonly mediaName: string; readonly media: LabelMedia; readonly mediaFingerprint: string };
export type RegisteredConnector = { readonly id: string; readonly name: string; readonly state: string; readonly availability: string; readonly lastSeenAt?: string; readonly printerIds: readonly string[] };
export type PrintOutcome = 'printed' | 'not_printed' | 'unknown';
export type PrintJob = { readonly id: string; readonly assetId?: string; readonly kind?: string; readonly predecessor?: string; readonly printerId: string; readonly status: string; readonly revision: number; readonly copies: number; readonly completedCopies: number; readonly latestAttemptId?: string; readonly idleConfirmed?: boolean; readonly resolution?: { readonly reportedOutcome: PrintOutcome } };
export type PrintSelection = { readonly printerId: string; readonly mediaFingerprint: string; readonly template: PrintTemplate; readonly copies: number; readonly previewFingerprint?: string };
export type PrintCatalog = { readonly printers: readonly RegisteredPrinter[]; readonly connectors: readonly RegisteredConnector[]; readonly templates: readonly LabelTemplate[]; readonly settings: PrintSettings };
export interface PrintingRepository {
  catalog(scope: PrintScope, signal: AbortSignal): Promise<PrintCatalog>;
  mediaPresets(scope: PrintScope, printer: RegisteredPrinter, signal: AbortSignal): Promise<readonly PrintMediaPreset[]>;
  configurePrinter(scope: PrintScope, printer: RegisteredPrinter, preset: PrintMediaPreset): Promise<RegisteredPrinter>;
  saveSettings(scope: PrintScope, settings: PrintSettings): Promise<PrintSettings>;
  preview(scope: PrintScope, assetId: string, printer: RegisteredPrinter, template: PrintTemplate, signal: AbortSignal): Promise<{ file: LabelFile; fingerprint: string }>;
  submit(scope: PrintScope, assetId: string, selection: PrintSelection, key: string): Promise<PrintJob>;
  job(scope: PrintScope, jobId: string, signal: AbortSignal): Promise<PrintJob>;
  jobs(scope: PrintScope, signal: AbortSignal): Promise<readonly PrintJob[]>;
  reprint(scope: PrintScope, predecessorId: string, selection: PrintSelection, key: string): Promise<PrintJob>;
  test(scope: PrintScope, printerId: string, selection: PrintSelection, key: string): Promise<PrintJob>;
  resolve(scope: PrintScope, job: PrintJob, outcome: PrintOutcome): Promise<PrintJob>;
  cancel(scope: PrintScope, job: PrintJob): Promise<PrintJob>;
}
export type PrintingWorkspace = { readonly repository: PrintingRepository; readonly files: LabelFiles; readonly requests: PrintRequests; readonly newRequestKey: () => string };

export type CreatePrintRequest = { readonly scope: PrintScope; readonly key: string; readonly selection: PrintSelection };

export function canReprint(job: PrintJob) { return ['completed', 'failed', 'canceled'].includes(job.status); }
