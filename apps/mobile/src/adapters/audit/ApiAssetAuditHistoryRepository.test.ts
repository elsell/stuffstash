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
    readonly limit?: number;
    readonly cursor?: string;
  }> = [];
  pages: Array<Page<AuditRecord>> = [];

  async listInventoryAuditRecords(
    tenantIdValue: string,
    inventoryIdValue: string,
    limit?: number,
    cursor?: string
  ): Promise<Page<AuditRecord>> {
    this.requests.push({ tenantId: tenantIdValue, inventoryId: inventoryIdValue, limit, cursor });
    return this.pages.shift() ?? page([]);
  }
}

describe('ApiAssetAuditHistoryRepository', () => {
  it('filters inventory audit records to the requested asset and follows inventory pages', async () => {
    const client = new FakeAuditApiClient();
    client.pages = [
      pageWithCursor([
        auditRecord({ id: 'audit-filter-update', targetId: 'asset-filters', action: 'asset.updated' }),
        auditRecord({
          id: 'audit-colliding-attachment',
          targetId: 'asset-filters',
          targetType: 'attachment',
          action: 'attachment.created'
        }),
        auditRecord({ id: 'audit-garage-update', targetId: 'asset-garage', action: 'asset.updated' })
      ], 'next-audit-page'),
      page([
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
        limit: 50,
        cursor: undefined
      },
      {
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        limit: 50,
        cursor: 'next-audit-page'
      }
    ]);
  });

  it('does not expose lossy cursor pagination when one inventory page has more matching records than the preview limit', async () => {
    const client = new FakeAuditApiClient();
    client.pages = [
      page([
        auditRecord({ id: 'audit-one', targetId: 'asset-filters', action: 'asset.updated' }),
        auditRecord({ id: 'audit-two', targetId: 'asset-filters', action: 'asset.moved' }),
        auditRecord({ id: 'audit-three', targetId: 'asset-filters', action: 'asset.viewed' })
      ])
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

  it('reports more history may exist when the bounded scan has no matching records yet', async () => {
    const client = new FakeAuditApiClient();
    client.pages = [
      pageWithCursor([
        auditRecord({ id: 'audit-garage-one', targetId: 'asset-garage', action: 'asset.updated' })
      ], 'next-audit-page'),
      pageWithCursor([
        auditRecord({ id: 'audit-garage-two', targetId: 'asset-garage', action: 'asset.moved' })
      ], 'another-page'),
      pageWithCursor([], 'third-page'),
      pageWithCursor([], 'fourth-page'),
      pageWithCursor([], 'fifth-page')
    ];
    const repository = new ApiAssetAuditHistoryRepository(client, new FakeInventoryContext());

    await expect(repository.listAssetAuditHistory({
      assetId: 'asset-filters',
      limit: 2
    })).resolves.toEqual({
      records: [],
      hasMore: true
    });
    expect(client.requests).toHaveLength(5);
  });
});

function auditRecord(input: {
  readonly id: string;
  readonly targetId: string;
  readonly action: string;
  readonly targetType?: string;
  readonly metadata?: Record<string, string>;
  readonly requestId?: string;
}): AuditRecord {
  return {
    id: input.id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-home',
    principalId: 'principal-home',
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
