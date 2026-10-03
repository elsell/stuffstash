import { expect, it } from 'vitest';
import { ApiPrintingRepository } from './printingRepository';
import { fakePrintMedia } from '$lib/fakes/printingRepository';
class LabelServer {
    contentReads = 0;
    fingerprint = 'media-v1';
    async fetch(input: RequestInfo | URL, init?: RequestInit) {
        const request = input instanceof Request ? input : new Request(input, init);
        const url = new URL(request.url);
        if (url.origin !== 'https://api.example' || request.headers.get('Authorization') !== 'Bearer human')
            return Response.json({ error: { code: 'unauthenticated', message: 'Sign in' } }, { status: 401 });
        const base = '/tenants/tenant/inventories/inventory';
        if (url.pathname === `${base}/assets/asset/label-renders` && request.method === 'POST')
            return Response.json({ data: { id: 'render', selectionFingerprint: 'selection', mediaFingerprint: this.fingerprint, contentPath: 'https://untrusted.example/secret', displayRotation: 270, expiresAt: '2099-01-01T00:00:00Z' } });
        if (url.pathname === `${base}/label-renders/render/content`) {
            this.contentReads++;
            return new Response(new Uint8Array([137, 80, 78, 71]), { headers: { 'Content-Type': 'image/png' } });
        }
        return Response.json({ error: { code: 'not_found', message: 'Not found' } }, { status: 404 });
    }
}
const scope = { tenantId: 'tenant', inventoryId: 'inventory' };
const selection = { printerId: 'printer', expectedMediaFingerprint: 'media-v1', templateId: 'qr-title', templateVersion: 1, showReference: true, copies: 1 };
it('fetches authenticated preview bytes by scoped render ID and ignores a foreign content path', async () => { const server = new LabelServer(); const repo = new ApiPrintingRepository('https://api.example', () => 'human', server.fetch.bind(server)); const preview = await repo.preview(scope, 'asset', selection, fakePrintMedia); expect(preview.bytes.type).toBe('image/png'); expect(server.contentReads).toBe(1); await expect(repo.preview({ ...scope, inventoryId: 'other' }, 'asset', selection, fakePrintMedia)).rejects.toMatchObject({ kind: 'invalid' }); });
it('rejects a changed media fingerprint before consuming preview content', async () => { const server = new LabelServer(); server.fingerprint = 'changed'; const repo = new ApiPrintingRepository('https://api.example', () => 'human', server.fetch.bind(server)); await expect(repo.preview(scope, 'asset', selection, fakePrintMedia)).rejects.toMatchObject({ kind: 'conflict' }); expect(server.contentReads).toBe(0); });
