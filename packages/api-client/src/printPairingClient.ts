import { createAuthenticatedTransport } from './authenticatedTransport';
import { StuffStashAPIError, type StuffStashClientOptions } from './stuffStashClient';
export interface PairingScope {
    tenantId: string;
    inventoryId: string;
}
export class PrintPairingClient {
    private readonly transport;
    constructor(options: StuffStashClientOptions) { this.transport = createAuthenticatedTransport(options); }
    async review(pairingId: string, scope: PairingScope, userCode: string) {
        const result = await this.transport.POST('/print-connector-pairings/{pairingId}/review', {
            params: { path: { pairingId } }, body: { ...scope, userCode }
        });
        return unwrap(result);
    }
    async approve(pairingId: string, scope: PairingScope, userCode: string, bindings: {
        candidateId: string;
        printerId: string;
    }[]) {
        return unwrap(await this.transport.POST('/print-connector-pairings/{pairingId}/approval', {
            params: { path: { pairingId } }, body: { ...scope, userCode, bindings }
        }));
    }
    async printers(scope: PairingScope, cursor?: string) {
        const result = await this.transport.GET('/tenants/{tenantId}/inventories/{inventoryId}/printers', {
            params: { path: scope, query: { limit: 100, cursor } }
        });
        const data = unwrap(result);
        return { items: data ?? [], nextCursor: result.data?.meta.pagination?.nextCursor ?? undefined };
    }
    async profiles(scope: PairingScope) {
        return unwrap(await this.transport.GET('/tenants/{tenantId}/inventories/{inventoryId}/printer-profiles', { params: { path: scope } })) ?? [];
    }
    async createPrinter(scope: PairingScope, input: {
        name: string;
        adapterId: string;
        presetId: string;
        presetVersion: number;
    }, key: string) {
        return unwrap(await this.transport.POST('/tenants/{tenantId}/inventories/{inventoryId}/printers', {
            params: { path: scope, header: { 'Idempotency-Key': key } }, body: input
        }));
    }
}
function unwrap<T>(result: {
    data?: {
        data: T;
    };
    error?: {
        error: {
            code: string;
            message: string;
        };
    };
    response: Response;
}): T {
    if (!result.response.ok || result.error || !result.data) {
        throw new StuffStashAPIError(result.response.status, result.error?.error.code ?? 'printing_unavailable', result.error?.error.message ?? 'Printing request failed.');
    }
    return result.data.data;
}
