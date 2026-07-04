import { describe, expect, it } from 'vitest';
import type { Asset, AssetAttachment } from '$lib/domain/inventory';
import {
  applyLoadedWorkspaceAssetDetail,
  assetDescriptionText,
  assetEditUnavailableStatus,
  assetFilesStatus,
  loadWorkspaceAssetDetail,
  refreshWorkspaceAssetAttachments
} from './workspaceAssetDetail';

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

  it('applies loaded detail into the selected asset workspace state', () => {
    const original = asset('asset-one');
    const updated = { ...original, title: 'Updated detail title' };
    const detailAttachments = [attachment('photo-one')];

    expect(
      applyLoadedWorkspaceAssetDetail(workspaceData([original]), {
        loaded: true,
        asset: updated,
        attachments: detailAttachments,
        error: ''
      })
    ).toEqual({
      data: workspaceData([updated]),
      loadedAssetDetail: updated,
      selectedAssetId: 'asset-one',
      selectedAssetAttachments: detailAttachments,
      mode: 'asset'
    });
  });

  it('builds asset detail fallback and status presentation', () => {
    expect(assetDescriptionText('Stored in the upstairs closet.')).toBe('Stored in the upstairs closet.');
    expect(assetDescriptionText('')).toBe('No description.');
    expect(assetEditUnavailableStatus(false)).toEqual({
      kind: 'edit-unavailable',
      message: 'Edit actions require asset edit access.'
    });
    expect(assetEditUnavailableStatus(true)).toBeNull();
    expect(assetFilesStatus(0)).toEqual({
      kind: 'files-empty',
      message: 'No active files.'
    });
    expect(assetFilesStatus(1)).toBeNull();
  });
});

function workspaceData(assets: Asset[]): import('$lib/domain/inventory').WorkspaceData {
  return {
    context: {
      principal: { id: 'principal-one', email: 'owner@example.com' },
      tenants: [{ id: 'tenant-home', name: 'Home', access: { relationship: 'owner', permissions: [] } }],
      selectedTenantId: 'tenant-home',
      selectedInventoryId: 'inventory-household',
      inventories: [
        {
          id: 'inventory-household',
          tenantId: 'tenant-home',
          name: 'Household',
          access: { relationship: 'editor', permissions: ['edit_asset'] }
        }
      ],
      assetLifecycleState: 'active',
      mediaUploadPolicy: { supportedContentTypes: ['image/jpeg'], maxBytes: 1024 },
      customAssetTypes: [],
      customFieldDefinitions: [],
      capability: 'editor'
    },
    assets
  };
}

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
