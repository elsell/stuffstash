import type { AuditRecord, StuffStashClient } from '@stuff-stash/api-client';
import type {
  AssetAuditHistoryPage,
  AssetAuditHistoryRepository,
  AssetAuditRecord
} from '../../application/assets/AssetAuditHistoryQuery';
import type { InventorySummaryRepository } from '../../application/home/InventorySummaryRepository';

type AuditApiClient = Pick<StuffStashClient, 'listAssetAuditRecords'>;

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
    const page = await this.client.listAssetAuditRecords(
      inventory.tenantId,
      inventory.id,
      input.assetId,
      input.limit
    );

    return {
      records: page.items.map(mapAuditRecord),
      hasMore: page.pagination.hasMore
    };
  }
}

function mapAuditRecord(record: AuditRecord): AssetAuditRecord {
  return {
    id: record.id,
    action: record.action,
    source: record.source,
    principalId: record.principalId,
    principal: record.principal,
    targetType: record.targetType,
    targetId: record.targetId,
    occurredAt: record.occurredAt,
    requestId: record.requestId,
    metadata: record.metadata
  };
}
