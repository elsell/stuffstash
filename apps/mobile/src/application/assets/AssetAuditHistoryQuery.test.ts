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
            from: 'Office',
            prompt: 'system prompt with model internals',
            transcript: 'move the thing',
            storage_key: 'garage/bucket/blob-secret',
            authorization_relationship: 'inventory#editor@user'
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

class FakeUnsafeMetadataRepository implements AssetAuditHistoryRepository {
  async listAssetAuditHistory(): Promise<AssetAuditHistoryPage> {
    return {
      records: [
        {
          id: 'audit-created',
          action: 'asset.created',
          source: 'voice',
          principalId: 'principal-home',
          targetType: 'asset',
          targetId: 'asset-water-bottle',
          occurredAt: '2026-06-25T12:30:00Z',
          metadata: {
            attachment_count: '2',
            credential: 'secret-token',
            file_name: 'bottle.jpg',
            file_size: '1536000',
            provider_prompt: 'You are an agent',
            raw_transcript: 'put my bottle somewhere',
            signed_url: 'https://garage.example.test/private',
            storage_key: 'tenant/blob/key',
            title: 'Water bottle'
          }
        }
      ],
      hasMore: false
    };
  }
}

class FakeLongMetadataRepository implements AssetAuditHistoryRepository {
  async listAssetAuditHistory(): Promise<AssetAuditHistoryPage> {
    return {
      records: [
        {
          id: 'audit-updated',
          action: 'asset.updated',
          source: 'api',
          principalId: 'principal-home',
          targetType: 'asset',
          targetId: 'asset-water-bottle',
          occurredAt: '2026-06-25T12:30:00Z',
          metadata: {
            title: 'A'.repeat(240)
          }
        }
      ],
      hasMore: false
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

  it('omits unsafe audit metadata from the mobile asset history view model', async () => {
    const repository = new FakeUnsafeMetadataRepository();
    const query = new AssetAuditHistoryQuery(repository);

    await expect(query.execute({ assetId: 'asset-water-bottle' })).resolves.toMatchObject({
      records: [
        {
          metadataRows: [
            { label: 'Attachment Count', value: '2' },
            { label: 'File Name', value: 'bottle.jpg' },
            { label: 'File Size', value: '1536000' },
            { label: 'Title', value: 'Water bottle' }
          ]
        }
      ]
    });
  });

  it('bounds safe audit metadata values before rendering', async () => {
    const query = new AssetAuditHistoryQuery(new FakeLongMetadataRepository());

    const history = await query.execute({ assetId: 'asset-water-bottle' });
    const titleRow = history.records[0]?.metadataRows.find((row) => row.label === 'Title');

    expect(titleRow?.value).toHaveLength(160);
    expect(titleRow?.value.endsWith('...')).toBe(true);
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
