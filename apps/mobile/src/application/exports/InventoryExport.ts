import { assertReadActive } from '../shared/ReadRequest';

export type InventoryExportFormat = 'json' | 'csv';
export type InventoryExportScope = { readonly tenantId: string; readonly inventoryId: string };
export type InventoryExportFile = { readonly content: string; readonly format: InventoryExportFormat };
export interface InventoryExportRepository {
  download(scope: InventoryExportScope, format: InventoryExportFormat, signal: AbortSignal): Promise<string>;
}
export interface InventoryExportFileDelivery {
  share(file: InventoryExportFile, signal: AbortSignal): Promise<void>;
}
export interface InventoryExportObserver {
  record(event: { readonly name: 'inventory_export.cleanup_failed' }): void;
}
export class ExportInventoryCommand {
  constructor(private readonly repository: InventoryExportRepository, private readonly files: InventoryExportFileDelivery) {}
  async execute(scope: InventoryExportScope, format: InventoryExportFormat, signal: AbortSignal): Promise<void> {
    assertReadActive(signal);
    const content = await this.repository.download(scope, format, signal);
    assertReadActive(signal);
    await this.files.share({ content, format }, signal);
  }
}
