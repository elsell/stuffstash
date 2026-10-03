import type { PairingInventory } from '$lib/domain/printPairing';
import type { PrintPairingRepository } from './printPairingRepository';
export type RotationTarget = {
    tenantId: string;
    inventoryId: string;
    connectorId: string;
};
export type RotationConnector = {
    id: string;
    name: string;
    generation: number;
};
export interface PrintRotationRepository extends Pick<PrintPairingRepository, 'inventories' | 'review'> {
    connector(scope: PairingInventory, id: string): Promise<RotationConnector>;
    rotate(pairingId: string, scope: PairingInventory, code: string, connectorId: string, generation: number): Promise<void>;
}
