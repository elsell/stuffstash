import type { PairingInventory, PairingReview, PairingSetup, PairingMedia } from '$lib/domain/printPairing';
export interface PrintPairingRepository {
    inventories(): Promise<PairingInventory[]>;
    review(pairingId: string, scope: PairingInventory, code: string): Promise<PairingReview>;
    setup(scope: PairingInventory): Promise<PairingSetup>;
    createPrinter(scope: PairingInventory, name: string, media: PairingMedia, key: string): Promise<string>;
    approve(pairingId: string, scope: PairingInventory, code: string, bindings: {
        candidateId: string;
        printerId: string;
    }[]): Promise<void>;
}
