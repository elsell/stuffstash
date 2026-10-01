import { describe, expect, it } from 'vitest';
import { ExportInventory } from './exportInventory';
import type { ExportFileDelivery, InventoryExportRepository } from '$lib/ports/inventoryExport';

class DownloadedFiles implements ExportFileDelivery {
  files: Parameters<ExportFileDelivery['save']>[0][] = [];
  async save(file: Parameters<ExportFileDelivery['save']>[0]) { this.files.push(file); }
}
describe('export inventory delivery', () => {
  it('delivers the chosen scoped file without exposing the inventory name as a filename', async () => {
    const requests: unknown[] = [];
    const repository: InventoryExportRepository = { async download(scope, format) { requests.push({ scope, format }); return 'id,title\n1,Box'; } };
    const files = new DownloadedFiles();
    await new ExportInventory(repository, files).execute({ tenantId: 'home', inventoryId: 'main' }, 'csv', new AbortController().signal);
    expect(requests).toEqual([{ scope: { tenantId: 'home', inventoryId: 'main' }, format: 'csv' }]);
    expect(files.files).toEqual([{ content: 'id,title\n1,Box', fileName: 'stuff-stash-inventory.csv', contentType: 'text/csv;charset=utf-8' }]);
  });
  it('does not deliver a stale export after leaving its scope, even if the repository ignores cancellation', async () => {
    let finish!: (value: string) => void;
    const repository: InventoryExportRepository = { download: () => new Promise(resolve => { finish = resolve; }) };
    const files = new DownloadedFiles();
    const controller = new AbortController();
    const pending = new ExportInventory(repository, files).execute({ tenantId: 'home', inventoryId: 'main' }, 'json', controller.signal);
    controller.abort(); finish('private data');
    await expect(pending).rejects.toMatchObject({ name: 'AbortError' });
    expect(files.files).toEqual([]);
  });
});
