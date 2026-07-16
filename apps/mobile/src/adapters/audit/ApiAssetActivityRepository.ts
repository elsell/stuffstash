import type { StuffStashClient } from '@stuff-stash/api-client';
import type {
  AssetActivityEntry,
  AssetActivityPage,
  AssetActivityRepository
} from '../../application/assets/AssetActivityQuery';

type ActivityClient = Pick<StuffStashClient, 'listAssetActivity'>;

export class ApiAssetActivityRepository implements AssetActivityRepository {
  constructor(private readonly client: ActivityClient) {}

  async listAssetActivity(input: Parameters<AssetActivityRepository['listAssetActivity']>[0]): Promise<AssetActivityPage> {
    const page = await this.client.listAssetActivity(
      input.tenantId,
      input.inventoryId,
      input.assetId,
      { view: input.view, limit: input.limit, cursor: input.cursor }
    );
    return {
      entries: page.items.map((entry): AssetActivityEntry => ({
        id: entry.id,
        principalId: entry.principalId,
        principal: entry.principal ? { id: entry.principal.id, email: entry.principal.email } : undefined,
        action: entry.action,
        category: entry.category,
        source: entry.source,
        occurredAt: entry.occurredAt,
        requestId: entry.requestId,
        changes: entry.changes.map((change) => ({ ...change })),
        undo: entry.undo ? { ...entry.undo } : undefined,
        technical: { ...entry.technical }
      })),
      nextCursor: page.pagination.nextCursor ?? undefined,
      hasMore: page.pagination.hasMore
    };
  }
}
