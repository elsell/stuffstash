import { describe, expect, it } from 'vitest';
import type { AuditRecord, Page } from '@stuff-stash/api-client';
import { inventoryId, tenantId, InventorySummary } from '../../domain/inventories/InventorySummary';
import { ApiAssetAuditHistoryRepository } from './ApiAssetAuditHistoryRepository';

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

class FakeAuditApiClient {
  requests: Array<{
    readonly tenantId: string;
    readonly inventoryId: string;
    readonly assetId?: string;
    readonly limit?: number;
  }> = [];
  pages: Array<Page<AuditRecord>> = [];

  async listAssetAuditRecords(
    tenantIdValue: string,
    inventoryIdValue: string,
    assetIdValue: string,
    limit?: number
  ): Promise<Page<AuditRecord>> {
    this.requests.push({ tenantId: tenantIdValue, inventoryId: inventoryIdValue, assetId: assetIdValue, limit });
    return this.pages.shift() ?? page([]);
  }
}

describe('ApiAssetAuditHistoryRepository', () => {
  it('loads asset-scoped audit records from the current inventory', async () => {
    const client = new FakeAuditApiClient();
    client.pages = [
      page([
        auditRecord({ id: 'audit-filter-update', targetId: 'asset-filters', action: 'asset.updated' }),
        auditRecord({
          id: 'audit-filter-move',
          targetId: 'asset-filters',
          action: 'asset.moved',
          metadata: { from: 'Office', to: 'Garage' },
          requestId: 'request-move'
        })
      ])
    ];
    const repository = new ApiAssetAuditHistoryRepository(client, new FakeInventoryContext());

    await expect(repository.listAssetAuditHistory({
      assetId: 'asset-filters',
      limit: 2
    })).resolves.toEqual({
      records: [
        {
          id: 'audit-filter-update',
          action: 'asset.updated',
          source: 'api',
          principalId: 'principal-home',
          principal: { id: 'principal-home', email: 'alex@example.test' },
          targetType: 'asset',
          targetId: 'asset-filters',
          occurredAt: '2026-06-25T12:00:00Z',
          requestId: undefined,
          metadata: {}
        },
        {
          id: 'audit-filter-move',
          action: 'asset.moved',
          source: 'api',
          principalId: 'principal-home',
          principal: { id: 'principal-home', email: 'alex@example.test' },
          targetType: 'asset',
          targetId: 'asset-filters',
          occurredAt: '2026-06-25T12:00:00Z',
          requestId: 'request-move',
          metadata: { from: 'Office', to: 'Garage' }
        }
      ],
      hasMore: false
    });
    expect(client.requests).toEqual([
      {
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        assetId: 'asset-filters',
        limit: 2
      }
    ]);
  });

  it('preserves the server has-more signal for bounded asset history', async () => {
    const client = new FakeAuditApiClient();
    client.pages = [
      pageWithCursor([
        auditRecord({ id: 'audit-one', targetId: 'asset-filters', action: 'asset.updated' }),
        auditRecord({ id: 'audit-two', targetId: 'asset-filters', action: 'asset.moved' })
      ], 'next-page')
    ];
    const repository = new ApiAssetAuditHistoryRepository(client, new FakeInventoryContext());

    await expect(repository.listAssetAuditHistory({
      assetId: 'asset-filters',
      limit: 2
    })).resolves.toMatchObject({
      records: [
        { id: 'audit-one' },
        { id: 'audit-two' }
      ],
      hasMore: true
    });
  });

  it('returns an empty page when the asset endpoint has no records', async () => {
    const client = new FakeAuditApiClient();
    client.pages = [
      page([])
    ];
    const repository = new ApiAssetAuditHistoryRepository(client, new FakeInventoryContext());

    await expect(repository.listAssetAuditHistory({
      assetId: 'asset-filters',
      limit: 2
    })).resolves.toEqual({
      records: [],
      hasMore: false
    });
    expect(client.requests).toHaveLength(1);
  });
});

function auditRecord(input: {
  readonly id: string;
  readonly targetId: string;
  readonly action: string;
  readonly targetType?: string;
  readonly metadata?: Record<string, string>;
  readonly requestId?: string;
  readonly principal?: AuditRecord['principal'];
}): AuditRecord {
  return {
    id: input.id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-home',
    principalId: 'principal-home',
    principal: input.principal ?? { id: 'principal-home', email: 'alex@example.test' },
    action: input.action,
    source: 'api',
    targetType: input.targetType ?? 'asset',
    targetId: input.targetId,
    occurredAt: '2026-06-25T12:00:00Z',
    requestId: input.requestId,
    metadata: input.metadata ?? {}
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
