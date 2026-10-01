export type ExportFormat = 'json' | 'csv';
export type ExportScope = { tenantId: string; inventoryId: string };
export interface InventoryExportRepository {
  download(scope: ExportScope, format: ExportFormat, signal: AbortSignal): Promise<string>;
}
export interface ExportFileDelivery {
  save(file: { content: string; fileName: string; contentType: string }, signal: AbortSignal): Promise<void>;
}
export const inventoryExportContext = Symbol('inventoryExport');
