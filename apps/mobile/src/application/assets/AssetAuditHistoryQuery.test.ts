import { describe, expect, it } from 'vitest';
import {
  AssetAuditHistoryPage,
  AssetAuditHistoryQuery,
  AssetAuditHistoryRepository
} from './AssetAuditHistoryQuery';

class FakeAssetAuditHistoryRepository implements AssetAuditHistoryRepository {
  input:
    | {
        readonly assetId: string;
        readonly limit: number;
      }
    | undefined;

  async listAssetAuditHistory(input: {
    readonly assetId: string;
    readonly limit: number;
  }): Promise<AssetAuditHistoryPage> {
    this.input = input;
    return {
      records: [
        {
          id: 'audit-move',
          action: 'asset.moved',
          source: 'api',
          principalId: 'principal-home',
          targetType: 'asset',
          targetId: input.assetId,
          occurredAt: '2026-06-25T12:30:00Z',
          requestId: 'request-one',
          metadata: {
            to: 'Kitchen',
            from: 'Office'
          }
        }
      ],
      hasMore: true
    };
  }
}

class EmptyBoundedHistoryRepository implements AssetAuditHistoryRepository {
  async listAssetAuditHistory(): Promise<AssetAuditHistoryPage> {
    return {
      records: [],
      hasMore: true
    };
  }
}

describe('AssetAuditHistoryQuery', () => {
  it('builds safe mobile audit history view models for one asset', async () => {
    const repository = new FakeAssetAuditHistoryRepository();
    const query = new AssetAuditHistoryQuery(repository);

    await expect(query.execute({
      assetId: 'asset-water-bottle',
      limit: 10
    })).resolves.toMatchObject({
      assetId: 'asset-water-bottle',
      hasMore: true,
      records: [
        {
          id: 'audit-move',
          title: 'Asset Moved',
          subtitle: 'Asset asset-water-bottle',
          occurredAtLabel: expect.stringContaining('Recorded Jun 25, 2026'),
          sourceLabel: 'API',
          principalLabel: 'Principal principal-home',
          requestLabel: 'Request request-one',
          metadataRows: [
            { label: 'From', value: 'Office' },
            { label: 'To', value: 'Kitchen' }
          ]
        }
      ]
    });
    expect(repository.input).toEqual({
      assetId: 'asset-water-bottle',
      limit: 10
    });
  });

  it('does not claim there is no history when a bounded scan may have older records', async () => {
    const query = new AssetAuditHistoryQuery(new EmptyBoundedHistoryRepository());

    await expect(query.execute({ assetId: 'asset-water-bottle' })).resolves.toMatchObject({
      records: [],
      hasMore: true,
      emptyTitle: 'No recent history found',
      emptyMessage: 'Older history may be available in the full audit log.'
    });
  });

  it('rejects blank asset IDs', async () => {
    const query = new AssetAuditHistoryQuery(new FakeAssetAuditHistoryRepository());

    await expect(query.execute({ assetId: '   ' })).rejects.toThrow('Asset ID is required.');
  });
});
