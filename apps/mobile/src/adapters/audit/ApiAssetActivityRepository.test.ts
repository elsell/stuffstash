import { describe, expect, it } from 'vitest';
import type { AssetActivityEntry, Page } from '@stuff-stash/api-client';
import { ApiAssetActivityRepository } from './ApiAssetActivityRepository';

class FakeClient {
  requests: unknown[] = [];

  async listAssetActivity(tenantId: string, inventoryId: string, assetId: string, input: object): Promise<Page<AssetActivityEntry>> {
    this.requests.push({ tenantId, inventoryId, assetId, input });
    return {
      items: [],
      pagination: { limit: 20, hasMore: true, nextCursor: 'next-page' }
    };
  }
}

describe('ApiAssetActivityRepository', () => {
  it('passes the loaded asset scope and cursor directly to the activity endpoint', async () => {
    const client = new FakeClient();
    const repository = new ApiAssetActivityRepository(client);
    await expect(repository.listAssetActivity({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      assetId: 'asset-drill',
      view: 'all',
      limit: 20,
      cursor: 'cursor-one'
    })).resolves.toEqual({ entries: [], hasMore: true, nextCursor: 'next-page' });
    expect(client.requests).toEqual([{
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      assetId: 'asset-drill',
      input: { view: 'all', limit: 20, cursor: 'cursor-one' }
    }]);
  });
});
