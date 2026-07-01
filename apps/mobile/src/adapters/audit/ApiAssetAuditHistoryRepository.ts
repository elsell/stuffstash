import type { AuditRecord, StuffStashClient } from '@stuff-stash/api-client';
import type {
  AssetAuditHistoryPage,
  AssetAuditHistoryRepository,
  AssetAuditRecord
} from '../../application/assets/AssetAuditHistoryQuery';
import type { InventorySummaryRepository } from '../../application/home/InventorySummaryRepository';

type AuditApiClient = Pick<StuffStashClient, 'listInventoryAuditRecords'>;

const auditPageSize = 50;
const maxInspectedAuditPages = 5;

export class ApiAssetAuditHistoryRepository implements AssetAuditHistoryRepository {
  constructor(
    private readonly client: AuditApiClient,
    private readonly inventories: Pick<InventorySummaryRepository, 'getDefaultInventorySummary'>
  ) {}

  async listAssetAuditHistory(input: {
    readonly assetId: string;
    readonly limit: number;
  }): Promise<AssetAuditHistoryPage> {
    const inventory = await this.inventories.getDefaultInventorySummary();
    const desiredMatches = input.limit;
    const records: AuditRecord[] = [];
    let cursor: string | undefined;
    let inventoryHasMore = false;
    let inspectedPages = 0;

    do {
      const page = await this.client.listInventoryAuditRecords(
        inventory.tenantId,
        inventory.id,
        auditPageSize,
        cursor
      );
      records.push(
        ...page.items.filter((record) =>
          record.targetType === 'asset' && record.targetId === input.assetId
        )
      );
      cursor = page.pagination.nextCursor ?? undefined;
      inventoryHasMore = page.pagination.hasMore;
      inspectedPages += 1;
    } while (records.length <= desiredMatches && inventoryHasMore && inspectedPages < maxInspectedAuditPages);

    return {
      records: records.slice(0, desiredMatches).map(mapAuditRecord),
      hasMore: records.length > desiredMatches || inventoryHasMore
    };
  }
}

function mapAuditRecord(record: AuditRecord): AssetAuditRecord {
  return {
    id: record.id,
    action: record.action,
    source: record.source,
    principalId: record.principalId,
    targetType: record.targetType,
    targetId: record.targetId,
    occurredAt: record.occurredAt,
    requestId: record.requestId,
    metadata: record.metadata
  };
}
