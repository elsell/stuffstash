export interface PairingInventory {
    tenantId: string;
    inventoryId: string;
    name: string;
    tenantName: string;
}
export interface PairingCandidate {
    id: string;
    name: string;
    adapterId: string;
}
export interface PairingReview {
    rotation?: boolean;
    id: string;
    name: string;
    fingerprint: string;
    candidates: PairingCandidate[];
}
export interface PairingPrinter {
    id: string;
    name: string;
    adapterId: string;
    mediaName: string;
}
export interface PairingMedia {
    key: string;
    name: string;
    adapterId: string;
    presetId: string;
    version: number;
}
export interface PairingSetup {
    printers: PairingPrinter[];
    media: PairingMedia[];
}
export interface PairingSelection {
    candidateId: string;
    destination: string;
    name: string;
    mediaKey: string;
    idempotencyKey: string;
    createdPrinterId?: string;
    registrationStarted?: boolean;
}
export type PairingFailureKind = 'authentication_required' | 'denied' | 'invalid' | 'unavailable';
export class PairingFailure extends Error {
    constructor(readonly kind: PairingFailureKind) { super(kind); }
}
