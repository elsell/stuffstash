import { describe, expect, it } from 'vitest';
import { ArchiveClient } from './archiveClient';

describe('archive transport', () => {
  it('keeps ZIP bytes binary and uses current credentials and stable request keys', async () => {
    const requests: Request[] = [];
    let token = 'first';
    const bytes = new Uint8Array([80, 75, 3, 4, 0, 255]);
    const client = new ArchiveClient({ baseUrl: 'https://api.example.test', tokenProvider: () => token,
      fetch: async (input, init) => {
        const request = new Request(input, init); requests.push(request);
        if (request.url.includes('/content')) return new Response(bytes, { headers: { 'Content-Type': 'application/zip' } });
        return Response.json({ data: { id: 'job', state: 'queued' }, meta: {} });
      }
    });
    await client.upload('home', 'same-request', new Blob([bytes]));
    token = 'second';
    await client.upload('home', 'same-request', new Blob([bytes]));
    const download = await client.download({ tenantId: 'home', inventoryId: 'main' }, 'job');
    expect(new Uint8Array(await new Response(download).arrayBuffer())).toEqual(bytes);
    expect(new Uint8Array(await requests[0]!.arrayBuffer())).toEqual(bytes);
    expect(requests.map(r => r.headers.get('Authorization'))).toEqual(['Bearer first', 'Bearer second', 'Bearer second']);
    expect(requests.slice(0, 2).map(r => r.headers.get('Idempotency-Key'))).toEqual(['same-request', 'same-request']);
    expect(requests[0]!.headers.get('Content-Type')).toBe('application/zip');
    expect(requests[2]!.url).toBe('https://api.example.test/tenants/home/archive-jobs/job/content?inventoryId=main');
    expect(requests.every(r => r.redirect === 'error')).toBe(true);
  });

  it('preserves denial and conflict errors instead of delivering their bodies as files', async () => {
    for (const status of [401, 403, 404, 409, 413, 500]) {
      const client = new ArchiveClient({ baseUrl: 'https://api.example.test', tokenProvider: () => 'token', fetch: async () => Response.json({ error: { code: 'archive_unavailable', message: 'Unavailable.' } }, { status }) });
      await expect(client.download({ tenantId: 'home', inventoryId: 'main' }, 'job')).rejects.toMatchObject({ status });
      await expect(client.upload('home', 'key', new Blob(['zip']))).rejects.toMatchObject({ status });
    }
  });

  it('does not dispatch a cancelled transfer while credentials are pending', async () => {
    let calls = 0;
    let resolve!: (value: string) => void;
    const client = new ArchiveClient({ baseUrl: 'https://api.example.test', tokenProvider: () => new Promise(done => { resolve = done; }), fetch: async () => { calls++; return Response.json({}); } });
    const controller = new AbortController();
    const pending = client.upload('home', 'key', new Blob(['zip']), controller.signal);
    await Promise.resolve(); await Promise.resolve();
    controller.abort(); resolve?.('token');
    await expect(pending).rejects.toMatchObject({ name: 'AbortError' });
    expect(calls).toBe(0);
  });
});
