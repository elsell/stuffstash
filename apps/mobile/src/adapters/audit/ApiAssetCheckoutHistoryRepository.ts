import type { AssetCheckout, StuffStashClient } from '@stuff-stash/api-client';
import type {
  AssetCheckoutHistoryPage,
  AssetCheckoutHistoryRepository,
  AssetCheckoutRecord
} from '../../application/assets/AssetCheckoutHistoryQuery';
import type { CurrentInventoryScopeRepository } from '../../application/home/CurrentInventoryScopeQuery';

type CheckoutHistoryApiClient = Pick<StuffStashClient, 'listAssetCheckoutHistory'>;

export class ApiAssetCheckoutHistoryRepository implements AssetCheckoutHistoryRepository {
  constructor(
    private readonly client: CheckoutHistoryApiClient,
    private readonly inventories: CurrentInventoryScopeRepository
  ) {}

  async listAssetCheckoutHistory(input: Parameters<AssetCheckoutHistoryRepository['listAssetCheckoutHistory']>[0]): Promise<AssetCheckoutHistoryPage> {
    const inventory = await this.inventories.getCurrentInventoryScope({ signal: input.signal });
    const page = await this.client.listAssetCheckoutHistory(
      inventory.tenantId,
      inventory.inventoryId,
      input.assetId,
      input.limit,
      input.cursor,
      input.signal
    );

    return {
      records: page.items.map(mapCheckoutRecord),
      hasMore: page.pagination.hasMore,
      nextCursor: page.pagination.nextCursor ?? undefined
    };
  }
}

function mapCheckoutRecord(record: AssetCheckout): AssetCheckoutRecord {
  return {
    id: record.id,
    state: record.state,
    checkedOutAt: record.checkedOutAt,
    checkedOutByPrincipalId: record.checkedOutByPrincipalId,
    checkoutDetails: record.checkoutDetails,
    returnedAt: record.returnedAt,
    returnedByPrincipalId: record.returnedByPrincipalId,
    returnDetails: record.returnDetails
  };
}
