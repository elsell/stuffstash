import { expect, it } from 'vitest';
import { PrintPairingClient } from './printPairingClient';
class PairingHTTPFake {
    approved = false;
    readonly requests: Request[] = [];
    fetch: typeof fetch = async (input, init) => {
        const request = new Request(input, init);
        this.requests.push(request);
        if (request.headers.get('Authorization') !== 'Bearer owner')
            return this.failure(403);
        const body = await request.json() as {
            tenantId: string;
            inventoryId: string;
            userCode: string;
            bindings?: {
                candidateId: string;
                printerId: string;
            }[];
        };
        const allowed = new URL(request.url).pathname.endsWith('/review')
            ? ['tenantId', 'inventoryId', 'userCode']
            : ['tenantId', 'inventoryId', 'userCode', 'bindings'];
        if (Object.keys(body).some(key => !allowed.includes(key)))
            return this.failure(422);
        if (body.tenantId !== 'tenant' || body.inventoryId !== 'inventory' || body.userCode !== 'ABCD1234')
            return this.failure(400);
        if (new URL(request.url).pathname.endsWith('/review'))
            return Response.json({ data: { id: 'pair', name: 'Garage', publicKeyFingerprint: 'sha256:key', candidates: [{ id: 'candidate', name: 'Brother', adapterId: 'brother' }] }, meta: {} });
        if (body.bindings?.length !== 1 || body.bindings[0].candidateId !== 'candidate' || body.bindings[0].printerId !== 'printer')
            return this.failure(400);
        this.approved = true;
        return Response.json({ data: { id: 'pair', state: 'approved' }, meta: {} });
    };
    private failure(status: number) { return Response.json({ error: { code: 'denied', message: 'Denied' } }, { status }); }
}
it('uses code-bound authenticated review and approval without URL secrets or device identities', async () => {
    const api = new PairingHTTPFake();
    let token = 'viewer';
    const client = new PrintPairingClient({ baseUrl: 'https://stash.example', tokenProvider: () => token, fetch: api.fetch });
    const scope = { tenantId: 'tenant', inventoryId: 'inventory', name: 'Home', tenantName: 'Household' };
    await expect(client.review('pair', scope, 'ABCD1234')).rejects.toMatchObject({ status: 403 });
    token = 'owner';
    await expect(client.review('pair', { ...scope, inventoryId: 'foreign' }, 'ABCD1234')).rejects.toMatchObject({ status: 400 });
    const review = await client.review('pair', scope, 'ABCD1234');
    expect(review.name).toBe('Garage');
    expect(api.approved).toBe(false);
    await client.approve('pair', scope, 'ABCD1234', [{ candidateId: 'candidate', printerId: 'printer' }]);
    expect(api.approved).toBe(true);
    expect(api.requests.every(request => !request.url.includes('ABCD1234') && request.redirect === 'error')).toBe(true);
    expect(review.candidates?.[0]).not.toHaveProperty('deviceId');
});
