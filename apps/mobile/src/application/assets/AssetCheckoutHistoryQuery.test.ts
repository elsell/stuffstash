import { describe, expect, it } from 'vitest';
import {
  AssetCheckoutHistoryPage,
  AssetCheckoutHistoryQuery,
  AssetCheckoutHistoryRepository
} from './AssetCheckoutHistoryQuery';

class FakeCheckoutHistoryRepository implements AssetCheckoutHistoryRepository {
  input:
    | {
        readonly assetId: string;
        readonly limit: number;
      }
    | undefined;

  async listAssetCheckoutHistory(input: {
    readonly assetId: string;
    readonly limit: number;
  }): Promise<AssetCheckoutHistoryPage> {
    this.input = input;
    return {
      records: [
        {
          id: 'checkout-open',
          state: 'open',
          checkedOutAt: '2026-06-25T12:30:00Z',
          checkedOutByPrincipalId: 'principal-home',
          checkoutDetails: 'Using at my desk'
        },
        {
          id: 'checkout-returned',
          state: 'returned',
          checkedOutAt: '2026-06-20T09:00:00Z',
          checkedOutByPrincipalId: 'principal-home',
          checkoutDetails: 'Loaned to Jamie',
          returnedAt: '2026-06-21T10:15:00Z',
          returnedByPrincipalId: 'principal-editor',
          returnDetails: 'Back in the tool bin'
        }
      ],
      hasMore: true
    };
  }
}

class EmptyCheckoutHistoryRepository implements AssetCheckoutHistoryRepository {
  async listAssetCheckoutHistory(): Promise<AssetCheckoutHistoryPage> {
    return { records: [], hasMore: false };
  }
}

class LongCheckoutDetailsRepository implements AssetCheckoutHistoryRepository {
  async listAssetCheckoutHistory(): Promise<AssetCheckoutHistoryPage> {
    return {
      records: [
        {
          id: 'checkout-long',
          state: 'open',
          checkedOutAt: '2026-06-25T12:30:00Z',
          checkedOutByPrincipalId: 'principal-home',
          checkoutDetails: 'A'.repeat(240)
        }
      ],
      hasMore: false
    };
  }
}

describe('AssetCheckoutHistoryQuery', () => {
  it('builds mobile checkout history view models for one asset', async () => {
    const repository = new FakeCheckoutHistoryRepository();
    const query = new AssetCheckoutHistoryQuery(repository);

    await expect(query.execute({
      assetId: 'asset-filter-kit',
      limit: 10
    })).resolves.toMatchObject({
      assetId: 'asset-filter-kit',
      hasMore: true,
      records: [
        {
          id: 'checkout-open',
          title: 'Checked out',
          statusLabel: 'Checked out',
          subtitle: expect.stringContaining('Principal principal-home'),
          checkedOutLabel: expect.stringContaining('Checked out Jun 25, 2026'),
          checkoutDetails: 'Using at my desk'
        },
        {
          id: 'checkout-returned',
          title: 'Returned',
          statusLabel: 'Returned',
          returnedLabel: expect.stringContaining('Principal principal-editor'),
          checkoutDetails: 'Loaned to Jamie',
          returnDetails: 'Back in the tool bin'
        }
      ]
    });
    expect(repository.input).toEqual({
      assetId: 'asset-filter-kit',
      limit: 10
    });
  });

  it('bounds checkout details before rendering', async () => {
    const query = new AssetCheckoutHistoryQuery(new LongCheckoutDetailsRepository());

    const history = await query.execute({ assetId: 'asset-filter-kit' });

    expect(history.records[0]?.checkoutDetails).toHaveLength(180);
    expect(history.records[0]?.checkoutDetails?.endsWith('...')).toBe(true);
  });

  it('shows an empty message when the asset has no checkout records', async () => {
    const query = new AssetCheckoutHistoryQuery(new EmptyCheckoutHistoryRepository());

    await expect(query.execute({ assetId: 'asset-filter-kit' })).resolves.toMatchObject({
      records: [],
      emptyTitle: 'No checkout history yet',
      emptyMessage: 'Checkouts and returns for this asset will appear here.'
    });
  });

  it('rejects blank asset IDs', async () => {
    const query = new AssetCheckoutHistoryQuery(new FakeCheckoutHistoryRepository());

    await expect(query.execute({ assetId: '   ' })).rejects.toThrow('Asset ID is required.');
  });
});
