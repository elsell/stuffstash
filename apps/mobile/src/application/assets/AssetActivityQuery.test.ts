import { describe, expect, it } from 'vitest';
import {
  AssetActivityEntry,
  AssetActivityPage,
  AssetActivityQuery,
  AssetActivityRepository
} from './AssetActivityQuery';

class FakeActivityRepository implements AssetActivityRepository {
  input: Parameters<AssetActivityRepository['listAssetActivity']>[0] | undefined;

  async listAssetActivity(input: Parameters<AssetActivityRepository['listAssetActivity']>[0]): Promise<AssetActivityPage> {
    this.input = input;
    return {
      entries: [activityEntry()],
      hasMore: true,
      nextCursor: 'next-page'
    };
  }
}

describe('AssetActivityQuery', () => {
  it('uses explicit scope and presents typed changes in homeowner language', async () => {
    const repository = new FakeActivityRepository();
    const query = new AssetActivityQuery(repository);
    const result = await query.execute({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      assetId: 'asset-drill',
      view: 'changes',
      limit: 20
    });

    expect(repository.input).toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      assetId: 'asset-drill',
      view: 'changes',
      limit: 20,
      cursor: undefined
    });
    expect(result).toMatchObject({
      hasMore: true,
      nextCursor: 'next-page',
      records: [{
        id: 'activity-update',
        title: 'Changed name',
        summary: 'Drill → Cordless drill',
        actorLabel: 'alex@example.test',
        sourceLabel: 'App'
      }]
    });
    expect(query.cachedEntry({ tenantId: 'tenant-home', inventoryId: 'inventory-home', assetId: 'asset-drill', activityId: 'activity-update' })).toEqual({ ...activityEntry(), technical: { count: '2' } });
  });

  it('rejects missing tenant, inventory, or asset scope', async () => {
    const query = new AssetActivityQuery(new FakeActivityRepository());
    await expect(query.execute({ tenantId: '', inventoryId: 'inventory-home', assetId: 'asset-drill' }))
      .rejects.toThrow('History scope is required.');
  });

  it('can restore a detail route by paging the explicitly scoped all-events view', async () => {
    const repository = new FakeActivityRepository();
    const query = new AssetActivityQuery(repository);
    await expect(query.loadEntry({ tenantId: 'tenant-home', inventoryId: 'inventory-home', assetId: 'asset-drill', activityId: 'activity-update' }))
      .resolves.toEqual({ ...activityEntry(), technical: { count: '2' } });
    expect(repository.input).toMatchObject({ tenantId: 'tenant-home', inventoryId: 'inventory-home', assetId: 'asset-drill', view: 'all' });
  });

  it('never returns an activity cached under another asset scope', async () => {
    const repository = new FakeActivityRepository();
    const query = new AssetActivityQuery(repository);
    await query.execute({ tenantId: 'tenant-home', inventoryId: 'inventory-home', assetId: 'asset-drill' });
    expect(query.cachedEntry({ tenantId: 'tenant-home', inventoryId: 'inventory-home', assetId: 'asset-saw', activityId: 'activity-update' })).toBeUndefined();
    await query.loadEntry({ tenantId: 'tenant-home', inventoryId: 'inventory-home', assetId: 'asset-saw', activityId: 'activity-update' });
    expect(repository.input).toMatchObject({ assetId: 'asset-saw', view: 'all' });
  });
});

function activityEntry(): AssetActivityEntry {
  return {
    id: 'activity-update',
    principalId: 'principal-home',
    principal: { id: 'principal-home', email: 'alex@example.test' },
    action: 'asset.updated',
    category: 'change',
    source: 'api',
    occurredAt: '2026-07-14T12:30:00Z',
    requestId: 'request-one',
    changes: [{ field: 'title', previousValue: 'Drill', currentValue: 'Cordless drill' }],
    undo: { operationId: 'operation-one', status: 'available' },
    technical: { count: '2', credential: 'must-not-render' }
  };
}
