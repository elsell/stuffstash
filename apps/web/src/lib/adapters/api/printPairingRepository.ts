import {labelMediaName} from '$lib/presentation/printing';
import { PrintPairingClient, StuffStashAPIError, StuffStashClient, type TokenProvider } from '@stuff-stash/api-client';
import type { PrintPairingRepository } from '$lib/ports/printPairingRepository';
import { PairingFailure, type PairingInventory, type PairingMedia, type PairingReview, type PairingSetup } from '$lib/domain/printPairing';
import { collectCursorPages } from './cursorPagination';
export class ApiPrintPairingRepository implements PrintPairingRepository {
    private readonly client: PrintPairingClient;
    private readonly inventoryClient: StuffStashClient;
    constructor(baseUrl: string, tokenProvider: TokenProvider, fetchImpl?: typeof fetch) {
        const options = { baseUrl, tokenProvider, fetch: fetchImpl };
        this.client = new PrintPairingClient(options);
        this.inventoryClient = new StuffStashClient(options);
    }
    async inventories(): Promise<PairingInventory[]> {
        return guarded(async () => {
            const tenants = await collectCursorPages(cursor => this.inventoryClient.listMyTenants(100, cursor));
            const results: PairingInventory[] = [];
            for (const tenant of tenants) {
                const inventories = await collectCursorPages(cursor => this.inventoryClient.listInventories(tenant.id, 100, cursor));
                for (const inventory of inventories) {
                    if (inventory.access.permissions.includes('configure'))
                        results.push({ tenantId: tenant.id, inventoryId: inventory.id, name: inventory.name, tenantName: tenant.name });
                }
            }
            return results;
        });
    }
    async review(pairingId: string, scope: PairingInventory, code: string): Promise<PairingReview> {
        return guarded(async () => {
            const value = await this.client.review(pairingId, scope, code);
            return { rotation: value.rotation, id: value.id, name: value.name, fingerprint: value.publicKeyFingerprint, candidates: (value.candidates ?? []).map(c => ({ id: c.id, name: c.name, adapterId: c.adapterId })) };
        });
    }
    async connector(scope:PairingInventory,id:string){
        return guarded(async()=>{const value=await this.client.connector(scope,id);return {id:value.id,name:value.name,generation:value.generation};});
    }
    async rotate(pairingId:string,scope:PairingInventory,code:string,connectorId:string,generation:number):Promise<void>{
        await guarded(()=>this.client.rotate(pairingId,scope,code,connectorId,generation));
    }
    async setup(scope: PairingInventory): Promise<PairingSetup> {
        return guarded(async () => {
            const printers: PairingSetup['printers'] = [];
            let cursor: string | undefined;
            const cursors = new Set<string>();
            do {
                const page = await this.client.printers(scope, cursor);
                printers.push(...page.items.filter(p => !p.retired).map(p => ({ id: p.id, name: p.name, adapterId: p.adapterId, mediaName: labelMediaName(p.media) })));
                cursor = page.nextCursor;
                if (cursor && cursors.has(cursor))
                    throw new PairingFailure('unavailable');
                if (cursor)
                    cursors.add(cursor);
            } while (cursor);
            const profiles = await this.client.profiles(scope);
            return { printers, media: profiles.flatMap(profile => (profile.media ?? []).map(m => ({ key: `${profile.adapterId}:${m.presetId}:${m.version}`, name: m.name, adapterId: profile.adapterId, presetId: m.presetId, version: m.version }))) };
        });
    }
    async createPrinter(scope: PairingInventory, name: string, media: PairingMedia, key: string): Promise<string> {
        return guarded(async () => (await this.client.createPrinter(scope, { name, adapterId: media.adapterId, presetId: media.presetId, presetVersion: media.version }, key)).id);
    }
    async approve(pairingId: string, scope: PairingInventory, code: string, bindings: {
        candidateId: string;
        printerId: string;
    }[]): Promise<void> {
        await guarded(() => this.client.approve(pairingId, scope, code, bindings));
    }
}
async function guarded<T>(operation: () => Promise<T>): Promise<T> {
    try {
        return await operation();
    }
    catch (error) {
        if (error instanceof PairingFailure)
            throw error;
        if (error instanceof StuffStashAPIError) {
            if (error.status === 401)
                throw new PairingFailure('authentication_required');
            if (error.status === 403)
                throw new PairingFailure('denied');
            if ([400, 404, 409, 410, 422].includes(error.status))
                throw new PairingFailure('invalid');
        }
        throw new PairingFailure('unavailable');
    }
}
