import { describe, expect, it } from 'vitest';
import { InventoryExportClient } from './inventoryExportClient';

describe('inventory export transport', () => {
  it('downloads both formats with current credentials, scoped URLs and redirects disabled', async () => {
    const requests: Request[] = [];
    let token = 'first';
    const client = new InventoryExportClient({ baseUrl: 'https://api.example.test', tokenProvider: () => token,
      fetch: async (input, init) => {
        const request = new Request(input, init); requests.push(request);
        return new Response(request.url.endsWith('format=csv') ? 'id,title\n1,Box\n' : '{"schemaVersion":1,"assets":[]}', { headers: { 'Content-Type': request.url.endsWith('format=csv') ? 'text/csv' : 'application/json' } });
      }
    });
    expect(await client.download('home', 'main', 'json')).toBe('{"schemaVersion":1,"assets":[]}');
    token = 'second';
    expect(await client.download('home', 'main', 'csv')).toBe('id,title\n1,Box\n');
    expect(requests.map(request => request.headers.get('Authorization'))).toEqual(['Bearer first', 'Bearer second']);
    expect(requests.map(request => request.url)).toEqual(['https://api.example.test/tenants/home/inventories/main/export?format=json', 'https://api.example.test/tenants/home/inventories/main/export?format=csv']);
    expect(requests.every(request => request.redirect === 'error')).toBe(true);
  });

  it('never turns an API failure into a downloadable file', async () => {
    for (const status of [401, 403, 422, 500]) {
      const client = new InventoryExportClient({ baseUrl: 'https://api.example.test', tokenProvider: () => 'token',
        fetch: async () => Response.json({ error: { code: 'export_failed', message: 'Export unavailable.' }, meta: {} }, { status })
      });
      await expect(client.download('home', 'main', 'json')).rejects.toMatchObject({ status, message: 'Export unavailable.' });
    }
  });

  it('cancels before network dispatch if session resolution is pending', async () => {
    let dispatches = 0;
    let resolve!: (value: string) => void;
    const pendingToken = new Promise<string>(done => { resolve = done; });
    const client = new InventoryExportClient({ baseUrl: 'https://api.example.test', tokenProvider: () => pendingToken,
      fetch: async () => { dispatches++; return new Response('private'); }
    });
    const controller = new AbortController();
    const pending = client.download('home', 'main', 'csv', controller.signal);
    controller.abort(); resolve('token');
    await expect(pending).rejects.toMatchObject({ name: 'AbortError' });
    expect(dispatches).toBe(0);
  });
});
