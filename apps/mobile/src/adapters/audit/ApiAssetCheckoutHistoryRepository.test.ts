import { describe, expect, it } from 'vitest';
import type { AssetCheckout, Page } from '@stuff-stash/api-client';
import { inventoryId, tenantId, InventorySummary } from '../../domain/inventories/InventorySummary';
import { ApiAssetCheckoutHistoryRepository } from './ApiAssetCheckoutHistoryRepository';

class FakeInventoryContext {
  readonly inventory: InventorySummary = {
    id: inventoryId('inventory-home'),
    tenantId: tenantId('tenant-home'),
    name: 'Home Inventory',
    role: 'owner',
    permissions: ['view', 'edit_asset'],
    description: '',
    updatedAtLabel: 'Loaded from API',
    locationCount: 0,
    locations: [],
    assets: []
  };

  async getDefaultInventorySummary(): Promise<InventorySummary> {
    return this.inventory;
  }
}

class FakeCheckoutHistoryApiClient {
  requests: Array<{
    readonly tenantId: string;
    readonly inventoryId: string;
    readonly assetId: string;
    readonly limit?: number;
  }> = [];
  pages: Array<Page<AssetCheckout>> = [];

  async listAssetCheckoutHistory(
    tenantIdValue: string,
    inventoryIdValue: string,
    assetIdValue: string,
    limit?: number
  ): Promise<Page<AssetCheckout>> {
    this.requests.push({ tenantId: tenantIdValue, inventoryId: inventoryIdValue, assetId: assetIdValue, limit });
    return this.pages.shift() ?? page([]);
  }
}

describe('ApiAssetCheckoutHistoryRepository', () => {
  it('loads checkout history from the current inventory', async () => {
    const client = new FakeCheckoutHistoryApiClient();
    client.pages = [
      page([
        checkoutRecord({ id: 'checkout-open', state: 'open' }),
        checkoutRecord({
          id: 'checkout-returned',
          state: 'returned',
          returnedAt: '2026-06-26T12:00:00Z',
          returnedByPrincipalId: 'principal-editor',
          returnDetails: 'Back in the bin'
        })
      ])
    ];
    const repository = new ApiAssetCheckoutHistoryRepository(client, new FakeInventoryContext());

    await expect(repository.listAssetCheckoutHistory({
      assetId: 'asset-filter-kit',
      limit: 2
    })).resolves.toEqual({
      records: [
        {
          id: 'checkout-open',
          state: 'open',
          checkedOutAt: '2026-06-25T12:00:00Z',
          checkedOutByPrincipalId: 'principal-home',
          checkoutDetails: 'Using at my desk',
          returnedAt: undefined,
          returnedByPrincipalId: undefined,
          returnDetails: undefined
        },
        {
          id: 'checkout-returned',
          state: 'returned',
          checkedOutAt: '2026-06-25T12:00:00Z',
          checkedOutByPrincipalId: 'principal-home',
          checkoutDetails: 'Using at my desk',
          returnedAt: '2026-06-26T12:00:00Z',
          returnedByPrincipalId: 'principal-editor',
          returnDetails: 'Back in the bin'
        }
      ],
      hasMore: false
    });
    expect(client.requests).toEqual([
      {
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        assetId: 'asset-filter-kit',
        limit: 2
      }
    ]);
  });

  it('preserves the server has-more signal', async () => {
    const client = new FakeCheckoutHistoryApiClient();
    client.pages = [
      pageWithCursor([
        checkoutRecord({ id: 'checkout-open', state: 'open' })
      ], 'next-page')
    ];
    const repository = new ApiAssetCheckoutHistoryRepository(client, new FakeInventoryContext());

    await expect(repository.listAssetCheckoutHistory({
      assetId: 'asset-filter-kit',
      limit: 1
    })).resolves.toMatchObject({
      records: [{ id: 'checkout-open' }],
      hasMore: true
    });
  });
});

function checkoutRecord(input: {
  readonly id: string;
  readonly state: AssetCheckout['state'];
  readonly returnedAt?: string;
  readonly returnedByPrincipalId?: string;
  readonly returnDetails?: string;
}): AssetCheckout {
  return {
    id: input.id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-home',
    assetId: 'asset-filter-kit',
    state: input.state,
    checkedOutAt: '2026-06-25T12:00:00Z',
    checkedOutByPrincipalId: 'principal-home',
    checkoutDetails: 'Using at my desk',
    returnedAt: input.returnedAt,
    returnedByPrincipalId: input.returnedByPrincipalId,
    returnDetails: input.returnDetails,
    createdAt: '2026-06-25T12:00:00Z',
    updatedAt: input.returnedAt ?? '2026-06-25T12:00:00Z'
  };
}

function page<T>(items: readonly T[]): Page<T> {
  return pageWithCursor(items, null);
}

function pageWithCursor<T>(items: readonly T[], nextCursor: string | null): Page<T> {
  return {
    items: [...items],
    pagination: {
      limit: items.length,
      nextCursor,
      hasMore: nextCursor !== null
    }
  };
}
