import { describe, expect, it } from 'vitest';
import { approvePairingSelection } from './pairingApproval';
import { PairingFailure, type PairingInventory, type PairingSelection } from '$lib/domain/printPairing';
import type { PrintPairingRepository } from '$lib/ports/printPairingRepository';
class PairingFake implements PrintPairingRepository {
    scope: PairingInventory = { tenantId: 'tenant', inventoryId: 'inventory', name: 'Home', tenantName: 'Family' };
    approved = false;
    unavailable = false;
    loseCreationResponse = false;
    printers = new Map<string, string>();
    async inventories() { return [this.scope]; }
    async review(_id: string, scope: PairingInventory, code: string) {
        if (scope.inventoryId !== this.scope.inventoryId || code !== 'ABCD1234')
            throw new PairingFailure('invalid');
        return { id: 'pair', name: 'Garage', fingerprint: 'sha256:key', candidates: [{ id: 'candidate', name: 'Brother', adapterId: 'brother' }] };
    }
    async setup() { return { printers: [], media: [{ key: 'stock', adapterId: 'brother', presetId: '29x90', version: 1, name: '29 × 90 mm' }] }; }
    async createPrinter(_scope: PairingInventory, _name: string, _media: unknown, key: string) { if (!this.printers.has(key))
        this.printers.set(key, `printer-${this.printers.size + 1}`); if(this.loseCreationResponse){this.loseCreationResponse=false;throw new PairingFailure('unavailable');}
        return this.printers.get(key)!; }
    async approve(_id: string, scope: PairingInventory, code: string, bindings: {
        candidateId: string;
        printerId: string;
    }[]) {
        await this.review(_id, scope, code);
        if (this.unavailable)
            throw new PairingFailure('unavailable');
        if (bindings.length !== 1 || bindings[0].candidateId !== 'candidate' || !Array.from(this.printers.values()).includes(bindings[0].printerId))
            throw new PairingFailure('invalid');
        this.approved = true;
    }
}
describe('pairing approval', () => {
    it('requires reviewed candidate and explicit media; retries reuse created destination', async () => {
        const repo = new PairingFake();
        const review = await repo.review('pair', repo.scope, 'ABCD1234');
        const setup = await repo.setup();
        const selected: PairingSelection[] = [{ candidateId: 'candidate', destination: 'new', name: 'Garage', mediaKey: '', idempotencyKey: 'request' }];
        await expect(approvePairingSelection(repo, 'pair', repo.scope, 'ABCD1234', review, setup, selected)).rejects.toThrow();
        expect(repo.printers.size).toBe(0);
        selected[0].mediaKey = 'stock';
        repo.unavailable = true;
        await expect(approvePairingSelection(repo, 'pair', repo.scope, 'ABCD1234', review, setup, selected)).rejects.toThrow();
        expect(repo.printers.size).toBe(1);
        expect(repo.approved).toBe(false);
        repo.unavailable = false;
        await approvePairingSelection(repo, 'pair', repo.scope, 'ABCD1234', review, setup, selected);
        expect(repo.printers.size).toBe(1);
        expect(repo.approved).toBe(true);
    });
    it('rejects unreviewed or incompatible destinations before creating or approving', async () => {
        const repo = new PairingFake();
        const review = await repo.review('pair', repo.scope, 'ABCD1234');
        const setup = await repo.setup();
        for (const selection of [{ candidateId: 'foreign', destination: 'new', name: 'Garage', mediaKey: 'stock', idempotencyKey: 'request' }, { candidateId: 'candidate', destination: 'foreign', name: '', mediaKey: '', idempotencyKey: 'request' }]) {
            await expect(approvePairingSelection(repo, 'pair', repo.scope, 'ABCD1234', review, setup, [selection])).rejects.toThrow();
        }
        expect(repo.approved).toBe(false);
        expect(repo.printers.size).toBe(0);
    });
});

it('freezes an uncertain creation and retries its exact idempotent request',async()=>{
 const repo=new PairingFake();repo.loseCreationResponse=true;
 const review=await repo.review('pair',repo.scope,'ABCD1234');const setup=await repo.setup();
 const selected:PairingSelection[]=[{candidateId:'candidate',destination:'new',name:'Garage',mediaKey:'stock',idempotencyKey:'request'}];
 await expect(approvePairingSelection(repo,'pair',repo.scope,'ABCD1234',review,setup,selected)).rejects.toThrow();
 expect(selected[0].registrationStarted).toBe(true);expect(selected[0].createdPrinterId).toBeUndefined();
 await approvePairingSelection(repo,'pair',repo.scope,'ABCD1234',review,setup,selected);
 expect(repo.printers.size).toBe(1);expect(repo.approved).toBe(true);
});
