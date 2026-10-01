import type { ExportFileDelivery, ExportFormat, ExportScope, InventoryExportRepository } from '$lib/ports/inventoryExport';

export class ExportInventory {
  constructor(private readonly repository: InventoryExportRepository, private readonly files: ExportFileDelivery) {}
  async execute(scope: ExportScope, format: ExportFormat, signal: AbortSignal): Promise<void> {
    signal.throwIfAborted();
    const content = await this.repository.download(scope, format, signal);
    signal.throwIfAborted();
    await this.files.save({ content, fileName: `stuff-stash-inventory.${format}`, contentType: format === 'json' ? 'application/json;charset=utf-8' : 'text/csv;charset=utf-8' }, signal);
  }
}
