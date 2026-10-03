import { PairingFailure, type PairingInventory, type PairingReview, type PairingSetup, type PairingSelection } from '$lib/domain/printPairing';
import type { PrintPairingRepository } from '$lib/ports/printPairingRepository';
export async function approvePairingSelection(repository: PrintPairingRepository, pairingId: string, scope: PairingInventory, code: string, review: PairingReview, setup: PairingSetup, selections: PairingSelection[]): Promise<void> {
    if (review.id !== pairingId)
        throw new PairingFailure('invalid');
    const chosen = selections.filter(s => s.destination !== '');
    if (chosen.length === 0)
        throw new PairingFailure('invalid');
    const candidateIDs = new Set<string>();
    const printerIDs = new Set<string>();
    for (const selection of chosen) {
        const candidate = review.candidates.find(c => c.id === selection.candidateId);
        if (!candidate || candidateIDs.has(candidate.id))
            throw new PairingFailure('invalid');
        candidateIDs.add(candidate.id);
        if (selection.destination === 'new') {
            if (!selection.name.trim() || !setup.media.some(m => m.key === selection.mediaKey && m.adapterId === candidate.adapterId))
                throw new PairingFailure('invalid');
        }
        else {
            if (!setup.printers.some(p => p.id === selection.destination && p.adapterId === candidate.adapterId) || printerIDs.has(selection.destination))
                throw new PairingFailure('invalid');
            printerIDs.add(selection.destination);
        }
    }
    const bindings = [];
    for (const selection of chosen) {
        let printerId = selection.destination;
        if (printerId === 'new') {
            selection.registrationStarted = true;
            const media = setup.media.find(m => m.key === selection.mediaKey)!;
            printerId = selection.createdPrinterId ?? await repository.createPrinter(scope, selection.name.trim(), media, selection.idempotencyKey);
            selection.createdPrinterId = printerId;
        }
        bindings.push({ candidateId: selection.candidateId, printerId });
    }
    await repository.approve(pairingId, scope, code, bindings);
}
