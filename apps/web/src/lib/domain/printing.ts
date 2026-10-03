export interface PrintScope {
    tenantId: string;
    inventoryId: string;
}
export interface LabelMedia {
    name: string;
    presetId: string;
    version: number;
    widthMicrometers: number;
    heightMicrometers: number;
    marginsMicrometers: {
        left: number;
        right: number;
        top: number;
        bottom: number;
    };
    resolutionDpi: number;
    rasterWidth: number;
    rasterHeight: number;
    orientation: string;
    colorMode: string;
    cutPolicy: string;
    displayRotation: number;
}
export type PrinterReadiness = 'ready' | 'unknown' | 'unavailable' | 'error';
export interface RegisteredPrinter {
    id: string;
    name: string;
    adapterId: string;
    revision: number;
    retired: boolean;
    media: LabelMedia;
    mediaFingerprint: string;
    readiness: PrinterReadiness;
}
export interface PrintConnector {
    id: string;
    name: string;
    state: string;
    authorizationPending: boolean;
    lastSeenAt?: string;
    printerIds: string[];
}
export interface LabelTemplate {
    id: string;
    version: number;
    name: string;
    showReferenceDefault: boolean;
}
export interface PrintDefaults {
    revision: number;
    defaultPrinterId: string | null;
    templateId: string;
    templateVersion: number;
    showReference: boolean;
    printOnCreateDefault: boolean;
}
export type PrintJobStatus = 'queued' | 'claimed' | 'printing' | 'completed' | 'failed' | 'uncertain' | 'canceled';
export interface PrintJob {
    id: string;
    printerId: string;
    assetId?: string;
    status: PrintJobStatus;
    revision: number;
    copies: number;
    completedCopies: number;
    reason: string;
    createdAt: string;
}
export interface LabelSelection {
    printerId: string;
    expectedMediaFingerprint: string;
    templateId: string;
    templateVersion: number;
    showReference: boolean;
    copies: number;
}
export interface LabelPreview {
    bytes: Blob;
    selectionFingerprint: string;
    mediaFingerprint: string;
    displayRotation: number;
    expiresAt: string;
}
export interface PrintPage<T> {
    items: T[];
    nextCursor?: string;
}
export type PrintingFailureKind = 'authentication_required' | 'denied' | 'invalid' | 'conflict' | 'unavailable';
export class PrintingFailure extends Error {
    constructor(readonly kind: PrintingFailureKind) { super(kind); }
}
