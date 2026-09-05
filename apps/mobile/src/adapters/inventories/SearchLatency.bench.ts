import { bench, describe } from 'vitest';
import type { Asset, AssetSearchResult, Page } from '@stuff-stash/api-client';
import { SearchAssetsQuery } from '../../application/search/SearchAssetsQuery';
import { ApiInventorySummaryRepository } from './ApiInventorySummaryRepository';
import { FakeInventoryApiClient } from './testing/InventoryApiClient';

// Synthetic network latency; no network, native build, or production data.
const transportLatencyMs = 50;
const waitForTransport = () => new Promise<void>((resolve) => setTimeout(resolve, transportLatencyMs));

class SearchLatencyTransport extends FakeInventoryApiClient {
  searchReads = 0;
  constructor(depth: number, private readonly includePath: boolean) {
    super();
    const template = this.assets[1]!;
    this.assets = Array.from({ length: depth + 1 }, (_, index) => ({
      ...template,
      id: `search-node-${index}`,
      title: index === 0 ? 'Search target' : `Ancestor ${index}`,
      parentAssetId: index < depth ? `search-node-${index + 1}` : null,
      primaryPhoto: undefined
    }));
  }

  override async searchAssets(tenantId: string, query: string): Promise<Page<AssetSearchResult>> {
    this.searchReads++;
    await waitForTransport();
    return {
      items: this.assets.filter((item) => item.title.toLowerCase().includes(query.toLowerCase())).map((asset) => ({
        type: 'asset', tenantId, inventory: this.inventory, asset,
        ancestorPath: this.includePath ? this.assets.slice(1).reverse().map(({ id, title }) => ({ id, title })) : undefined,
        matches: [{ field: 'title', value: asset.title }]
      })),
      pagination: { limit: 20, hasMore: false, nextCursor: null }
    };
  }

  override async getAsset(tenantId: string, inventoryId: string, id: string): Promise<Asset> {
    await waitForTransport();
    return super.getAsset(tenantId, inventoryId, id);
  }
}

describe('Warm-directory search hydration, no TanStack cache, 50 ms per search/ancestor read', () => {
  for (const includePath of [false, true]) {
  for (const depth of [0, 1, 4]) {
    const client = new SearchLatencyTransport(depth, includePath);
    const query = new SearchAssetsQuery(new ApiInventorySummaryRepository(client, 'tenant-home'));
    bench(`one result, ancestry depth ${depth}, ${includePath ? "API path" : "legacy fallback"}`, async () => {
      client.searchReads = 0;
      client.getAssetRequests.length = 0;
      const result = await query.execute({
        query: 'Search target', limit: 20, kind: 'all', lifecycleState: 'active',
        checkoutState: 'any', sort: 'updated_desc'
      });
      if (result.assets.length !== 1 || result.assets[0]?.title !== 'Search target') {
        throw new Error('Search benchmark must return its matching asset.');
      }
      const trail = result.assets[0].parentLocationTrail.map((crumb) => crumb.title);
      const expectedTrail = Array.from({ length: depth }, (_, index) => `Ancestor ${depth - index}`);
      if (JSON.stringify(trail) !== JSON.stringify(expectedTrail)) {
        throw new Error('Search benchmark must preserve the complete ancestor path.');
      }
      if (client.searchReads !== 1 || client.getAssetRequests.length > (includePath ? 0 : depth)) {
        throw new Error('Search hydration exceeded its baseline request budget.');
      }
    }, { iterations: 5, time: 300, warmupIterations: 1, warmupTime: 0 });
  }
  }
});
