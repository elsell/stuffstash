import type { AuditRecord } from '$lib/domain/inventory';
import type { Pagination } from './pagination';

export interface AuditRecordPage {
  items: AuditRecord[];
  pagination: Pagination;
}

export interface InventoryAuditRepository {
  listTenantAuditRecords(tenantId: string, cursor?: string, signal?: AbortSignal): Promise<AuditRecordPage>;
  listInventoryAuditRecords(tenantId: string, inventoryId: string, cursor?: string, signal?: AbortSignal): Promise<AuditRecordPage>;
}
