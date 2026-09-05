import { expect, it } from 'vitest';
import { SearchAssetsQuery } from '../../application/search/SearchAssetsQuery';
import { ApiInventorySummaryRepository } from '../../adapters/inventories/ApiInventorySummaryRepository';
import { FakeInventoryApiClient } from '../../adapters/inventories/testing/InventoryApiClient';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { QueryClientInventoryMutationObserver } from '../../adapters/serverState/QueryClientInventoryMutationObserver';
import { browseInfiniteQueryOptions } from './BrowseInfiniteQuery';

class PathInventoryClient extends FakeInventoryApiClient {
  override async searchAssets(...args: Parameters<FakeInventoryApiClient['searchAssets']>) {
    const page = await super.searchAssets(...args);
    return { ...page, items: page.items.map((item) => ({ ...item,
      ancestorPath: this.assets.filter((asset) => asset.id === item.asset.parentAssetId)
        .map(({ id, title }) => ({ id, title }))
    })) };
  }
}

it('reuses complete search paths and refreshes them after an ancestor rename', async () => {
  const transport = new PathInventoryClient();
  const cache = createMobileQueryClient();
  const query = new SearchAssetsQuery(new ApiInventorySummaryRepository(transport, 'tenant-home'));
  const options = browseInfiniteQueryOptions('composition', 'tenant-home', 'inventory-home', {
    query: 'filters', scope: 'all', lifecycleState: 'active', checkoutState: 'any', sort: 'updated_desc', tagIds: []
  }, query);
  try {
    const first = await cache.fetchInfiniteQuery(options);
    expect(first.pages[0]?.assets[0]?.parentLocationTrail[0]?.title).toBe('Garage');
    await cache.fetchInfiniteQuery(options);
    expect(transport.searchAssetRequests).toHaveLength(1);
    transport.assets[0] = { ...transport.assets[0]!, title: 'Workshop' };
    new QueryClientInventoryMutationObserver(cache, 'composition').onInventoryMutation({
      kind: 'asset_updated', tenantId: 'tenant-home', inventoryId: 'inventory-home', assetId: 'asset-garage'
    });
    const refreshed = await cache.fetchInfiniteQuery(options);
    expect(refreshed.pages[0]?.assets[0]?.parentLocationTrail[0]?.title).toBe('Workshop');
    expect(transport.searchAssetRequests).toHaveLength(2);
    expect(transport.getAssetRequests).toEqual([]);
  } finally {
    cache.clear();
  }
});
