import { expect, it } from 'vitest';
import { ApiInventorySummaryRepository } from './ApiInventorySummaryRepository';
import { FakeInventoryApiClient } from './testing/InventoryApiClient';

class PathSearchClient extends FakeInventoryApiClient {
  override async searchAssets(...args: Parameters<FakeInventoryApiClient['searchAssets']>) {
    const page = await super.searchAssets(...args);
    return { ...page, items: page.items.map((item) => ({
      ...item,
      ancestorPath: item.asset.parentAssetId ? [{ id: item.asset.parentAssetId, title: 'Garage' }] : []
    })) };
  }
}

const criteria = {
  query: 'filters', limit: 20, lifecycleState: 'active', checkoutState: 'any',
  kind: 'all', sort: 'updated_desc'
} as const;

it('renders the complete API search path without ancestor HTTP reads', async () => {
  const client = new PathSearchClient();
  const repository = new ApiInventorySummaryRepository(client, 'tenant-home');
  const result = await repository.browseAssets(criteria);
  expect(result.assets[0]?.parentLocationTrail).toEqual([{ id: 'asset-garage', title: 'Garage' }]);
  expect(client.getAssetRequests).toEqual([]);
  expect(client.searchAssetRequests).toHaveLength(1);
});

it('keeps ancestor reads for older API results without a path', async () => {
  const client = new FakeInventoryApiClient();
  const result = await new ApiInventorySummaryRepository(client, 'tenant-home').browseAssets(criteria);
  expect(result.assets[0]?.parentLocationTrail).toEqual([{ id: 'asset-garage', title: 'Garage' }]);
  expect(client.getAssetRequests).toEqual([{ inventoryId: 'inventory-home', assetId: 'asset-garage' }]);
});

it('treats an explicit root path as authoritative', async () => {
  const client = new PathSearchClient();
  client.assets[1] = { ...client.assets[1]!, parentAssetId: null };
  const result = await new ApiInventorySummaryRepository(client, 'tenant-home').browseAssets(criteria);
  expect(result.assets[0]?.parentLocationTrail).toEqual([]);
  expect(client.getAssetRequests).toEqual([]);
});
