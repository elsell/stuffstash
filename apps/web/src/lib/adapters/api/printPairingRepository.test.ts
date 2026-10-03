import { expect, it } from 'vitest';
import { ApiPrintPairingRepository } from './printPairingRepository';
it('lists only inventories with configure permission from actual wire access summaries', async () => {
    const fetcher: typeof fetch = async (input, init) => {
        const request = new Request(input, init);
        const path = new URL(request.url).pathname;
        if (request.headers.get('Authorization') !== 'Bearer owner')
            return Response.json({ error: { code: 'authentication_required', message: 'Sign in' } }, { status: 401 });
        const data = path.endsWith('/tenants') ? [{ id: 'tenant', name: 'Family', access: { relationship: 'owner', permissions: ['view'] } }] : [
            { id: 'owned', tenantId: 'tenant', name: 'Home', access: { relationship: 'owner', permissions: ['view', 'configure'] } },
            { id: 'read-only', tenantId: 'tenant', name: 'Shared', access: { relationship: 'viewer', permissions: ['view'] } }
        ];
        return Response.json({ data, meta: { pagination: { hasMore: false, nextCursor: null, limit: 100 } } });
    };
    const repository = new ApiPrintPairingRepository('https://stash.example', () => 'owner', fetcher);
    await expect(repository.inventories()).resolves.toEqual([{ tenantId: 'tenant', inventoryId: 'owned', name: 'Home', tenantName: 'Family' }]);
});
