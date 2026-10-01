import { describe, expect, it } from 'vitest';
import { ExportInventoryCommand, type InventoryExportFile, type InventoryExportFileDelivery, type InventoryExportRepository } from './InventoryExport';

class DeliveredFiles implements InventoryExportFileDelivery {
  files: InventoryExportFile[] = [];
  async share(file: InventoryExportFile) { this.files.push(file); }
}
describe('inventory export command', () => {
  it('uses the displayed scope and delivers the chosen format', async () => {
    const requests: unknown[] = [];
    const repository: InventoryExportRepository = { async download(scope, format) { requests.push({ scope, format }); return '{"assets":[]}'; } };
    const delivery = new DeliveredFiles();
    await new ExportInventoryCommand(repository, delivery).execute({ tenantId: 'home', inventoryId: 'garage' }, 'json', new AbortController().signal);
    expect(requests).toEqual([{ scope: { tenantId: 'home', inventoryId: 'garage' }, format: 'json' }]);
    expect(delivery.files).toEqual([{ content: '{"assets":[]}', format: 'json' }]);
  });
  it('never opens a late share sheet after scope cancellation', async () => {
    let finish!: (value: string) => void;
    const repository: InventoryExportRepository = { download: () => new Promise(resolve => { finish = resolve; }) };
    const delivery = new DeliveredFiles();
    const cancellation = new AbortController();
    const pending = new ExportInventoryCommand(repository, delivery).execute({ tenantId: 'home', inventoryId: 'garage' }, 'csv', cancellation.signal);
    cancellation.abort(); finish('private content');
    await expect(pending).rejects.toMatchObject({ name: 'AbortError' });
    expect(delivery.files).toEqual([]);
  });
  it('does not share error responses', async () => {
    const delivery = new DeliveredFiles();
    const repository: InventoryExportRepository = { async download() { throw Object.assign(new Error('Forbidden'), { status: 403 }); } };
    await expect(new ExportInventoryCommand(repository, delivery).execute({ tenantId: 'home', inventoryId: 'garage' }, 'json', new AbortController().signal)).rejects.toMatchObject({ status: 403 });
    expect(delivery.files).toEqual([]);
  });
});
