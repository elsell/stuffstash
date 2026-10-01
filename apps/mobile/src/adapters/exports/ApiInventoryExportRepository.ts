import type { InventoryExportClient } from '@stuff-stash/api-client';
import type { InventoryExportFormat, InventoryExportRepository, InventoryExportScope } from '../../application/exports/InventoryExport';
export class ApiInventoryExportRepository implements InventoryExportRepository {
  constructor(private readonly client: InventoryExportClient) {}
  download(scope: InventoryExportScope, format: InventoryExportFormat, signal: AbortSignal) {
    return this.client.download(scope.tenantId, scope.inventoryId, format, signal);
  }
}
