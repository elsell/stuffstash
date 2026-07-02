import { describe, expect, it } from 'vitest';
import type { Asset, AssetAttachment } from '$lib/domain/inventory';
import { loadWorkspaceAssetDetail, refreshWorkspaceAssetAttachments } from './workspaceAssetDetail';

function asset(id: string, lifecycleState: Asset['lifecycleState'] = 'active'): Asset {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    kind: 'item',
    title: id,
    description: '',
    parentAssetId: null,
    lifecycleState
  };
}

function attachment(id: string, assetId = 'asset-one'): AssetAttachment {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    assetId,
    fileName: `${id}.jpg`,
    contentType: 'image/jpeg',
    sizeBytes: 12,
    lifecycleState: 'active',
    thumbnailUrl: `blob:${id}`
  };
}

describe('workspace asset detail helpers', () => {
  it('loads asset detail, attachments, and replaces the workspace asset copy', async () => {
    const updated = { ...asset('asset-one'), title: 'Updated passport' };
    const calls: string[] = [];
    const result = await loadWorkspaceAssetDetail(
      repository({ asset: updated, attachments: [attachment('photo-one')], calls }),
      'tenant-home',
      'inventory-household',
      'asset-one'
    );

    expect(result).toMatchObject({
      loaded: true,
      asset: updated,
      attachments: [attachment('photo-one')],
      error: ''
    });
    expect(calls).toEqual([
      'get:tenant-home:inventory-household:asset-one',
      'attachments:tenant-home:inventory-household:asset-one'
    ]);
  });

  it('returns a calm error when detail loading fails', async () => {
    const result = await loadWorkspaceAssetDetail(
      repository({ failure: new Error('Asset not found.') }),
      'tenant-home',
      'inventory-household',
      'missing'
    );

    expect(result).toEqual({
      loaded: false,
      asset: null,
      attachments: [],
      error: 'Asset not found.'
    });
  });

  it('refreshes attachments through the same detail boundary', async () => {
    const calls: string[] = [];
    await expect(
      refreshWorkspaceAssetAttachments(repository({ asset: asset('asset-one'), attachments: [attachment('manual')], calls }), {
        tenantId: 'tenant-home',
        inventoryId: 'inventory-household',
        assetId: 'asset-one'
      })
    ).resolves.toEqual([attachment('manual')]);
    expect(calls).toEqual(['attachments:tenant-home:inventory-household:asset-one']);
  });
});

interface RepositoryOptions {
  asset?: Asset;
  attachments?: AssetAttachment[];
  failure?: Error;
  calls?: string[];
}

function repository(options: RepositoryOptions) {
  return {
    async getAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset> {
      options.calls?.push(`get:${tenantId}:${inventoryId}:${assetId}`);
      if (options.failure) {
        throw options.failure;
      }
      return options.asset ?? asset('asset-one');
    },
    async listAssetAttachments(tenantId: string, inventoryId: string, assetId: string): Promise<AssetAttachment[]> {
      options.calls?.push(`attachments:${tenantId}:${inventoryId}:${assetId}`);
      if (options.failure) {
        throw options.failure;
      }
      return options.attachments ?? [];
    }
  };
}
