import { InventoryExportClient } from '@stuff-stash/api-client';
import type { ExportFormat, ExportScope, InventoryExportRepository } from '$lib/ports/inventoryExport';
export class ApiInventoryExportRepository implements InventoryExportRepository {
  constructor(private readonly client: InventoryExportClient) {}
  download(scope: ExportScope, format: ExportFormat, signal: AbortSignal) {
    return this.client.download(scope.tenantId, scope.inventoryId, format, signal);
  }
}
