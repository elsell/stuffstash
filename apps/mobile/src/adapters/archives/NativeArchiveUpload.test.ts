import { describe, expect, it } from 'vitest';
import { NativeArchiveUpload, type ArchiveUploadModule } from './NativeArchiveUpload';

class Transfers implements ArchiveUploadModule {
  requests: unknown[] = []; cancellations: string[] = [];
  finish!: (value: { status: number; body: string }) => void;
  upload(id: string, url: string, uri: string, headers: Record<string, string>) { this.requests.push({ id, url, uri, headers }); return new Promise<{ status: number; body: string }>(resolve => { this.finish = resolve; }); }
  async cancel(id: string) { this.cancellations.push(id); }
}
describe('native archive upload adapter', () => {
  it('hands the file URI and authenticated operation to native without reading file bytes', async () => {
    const module = new Transfers(), transfer = new NativeArchiveUpload(module, () => 'transfer');
    const result = transfer.send('file:///private/archive.zip', new Request('https://api.example.test/tenants/home/archive-restores', { method: 'POST', headers: { Authorization: 'Bearer current', 'Idempotency-Key': 'stable' } }));
    module.finish({ status: 201, body: '{"data":{"id":"job"}}' });
    expect((await result).status).toBe(201);
    expect(module.requests).toEqual([{ id: 'transfer', url: 'https://api.example.test/tenants/home/archive-restores', uri: 'file:///private/archive.zip', headers: { authorization: 'Bearer current', 'idempotency-key': 'stable' } }]);
  });
  it('cancels native work and rejects late success after leaving the task', async () => {
    const module = new Transfers(), transfer = new NativeArchiveUpload(module, () => 'transfer'), controller = new AbortController();
    const result = transfer.send('file:///archive.zip', new Request('https://api.example.test/tenants/home/archive-restores', { method: 'POST', signal: controller.signal }));
    controller.abort(); module.finish({ status: 201, body: '{}' });
    await expect(result).rejects.toMatchObject({ name: 'AbortError' }); expect(module.cancellations).toEqual(['transfer']);
  });
  it('rejects redirects and oversized native responses', async () => {
    for (const response of [{ status: 302, body: '' }, { status: 200, body: 'x'.repeat(65537) }]) {
      const module = new Transfers(), transfer = new NativeArchiveUpload(module, () => 'transfer');
      const result = transfer.send('file:///archive.zip', new Request('https://api.example.test/tenants/home/archive-restores', { method: 'POST' }));
      module.finish(response); await expect(result).rejects.toThrow();
    }
  });
});
