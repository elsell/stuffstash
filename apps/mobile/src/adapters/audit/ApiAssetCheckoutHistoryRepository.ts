import type { AssetCheckout, StuffStashClient } from '@stuff-stash/api-client';
import type {
  AssetCheckoutHistoryPage,
  AssetCheckoutHistoryRepository,
  AssetCheckoutRecord
} from '../../application/assets/AssetCheckoutHistoryQuery';
import type { InventorySummaryRepository } from '../../application/home/InventorySummaryRepository';

type CheckoutHistoryApiClient = Pick<StuffStashClient, 'listAssetCheckoutHistory'>;

export class ApiAssetCheckoutHistoryRepository implements AssetCheckoutHistoryRepository {
  constructor(
    private readonly client: CheckoutHistoryApiClient,
    private readonly inventories: Pick<InventorySummaryRepository, 'getDefaultInventorySummary'>
  ) {}

  async listAssetCheckoutHistory(input: {
    readonly assetId: string;
    readonly limit: number;
  }): Promise<AssetCheckoutHistoryPage> {
    const inventory = await this.inventories.getDefaultInventorySummary();
    const page = await this.client.listAssetCheckoutHistory(
      inventory.tenantId,
      inventory.id,
      input.assetId,
      input.limit
    );

    return {
      records: page.items.map(mapCheckoutRecord),
      hasMore: page.pagination.hasMore
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
