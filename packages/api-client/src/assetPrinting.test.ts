import { expect, it } from 'vitest';
import { StuffStashClient } from './stuffStashClient';
it('uses one scoped create-and-print request identity and exposes its queued job', async () => {
  const created = new Map<string, object>();
  const printLabel = { printerId: 'printer', expectedMediaFingerprint: 'media', templateId: 'qr-title', templateVersion: 1, templateOptions: { showReference: true }, copies: 1 };
  const client = new StuffStashClient({ baseUrl: 'https://api.example.test', tokenProvider: () => 'test-token', fetch: async (input, init) => {
    const request = new Request(input, init);
    expect(request.headers.get('Authorization')).toBe('Bearer test-token');
    const key = request.headers.get('Idempotency-Key'); expect(key).toBe('create-request');
    expect((await request.json()).printLabel).toEqual(printLabel);
    const identity = `${new URL(request.url).pathname}:${key}`;
    const asset = created.get(identity) ?? { id: 'asset', tenantId: 'tenant', inventoryId: 'inventory', title: 'Lamp', kind: 'item', description: '', lifecycleState: 'active', createdAt: '', updatedAt: '', printJobId: 'job' };
    created.set(identity, asset); return Response.json({ data: asset, meta: {} });
  } });
  const input = { kind: 'item' as const, title: 'Lamp', printLabel };
  expect((await client.createAsset('tenant', 'inventory', input, 'create-request')).printJobId).toBe('job');
  expect((await client.createAsset('tenant', 'inventory', input, 'create-request')).printJobId).toBe('job');
  expect(created.size).toBe(1);
});
