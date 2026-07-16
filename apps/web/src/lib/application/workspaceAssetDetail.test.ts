import { describe, expect, it } from 'vitest';
import { AuthenticationRequiredError } from './authenticationRequired';
import type { Asset, AssetAttachment, AssetCheckout } from '$lib/domain/inventory';
import {
  applyLoadedWorkspaceAssetDetail,
  assetDescriptionText,
  assetEditUnavailableStatus,
  assetFilesStatus,
  partitionAssetDetailFields,
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
      repository({ asset: updated, attachments: [attachment('photo-one')], checkoutHistory: [checkout('checkout-one')], calls }),
      'tenant-home',
      'inventory-household',
      'asset-one'
    );

    expect(result).toMatchObject({
      loaded: true,
      asset: updated,
      attachments: [attachment('photo-one')],
      checkoutHistory: [checkout('checkout-one')],
      error: ''
    });
    expect(calls).toEqual([
      'get:tenant-home:inventory-household:asset-one',
      'attachments:tenant-home:inventory-household:asset-one',
      'checkouts:tenant-home:inventory-household:asset-one'
    ]);
  });

  it('replaces unsafe detail failures with calm product copy', async () => {
    const result = await loadWorkspaceAssetDetail(
      repository({ failure: Object.assign(new Error('private database diagnostic'), { safeForUser: false }) }),
      'tenant-home',
      'inventory-household',
      'missing'
    );

    expect(result).toEqual({
      loaded: false,
      asset: null,
      attachments: [],
      checkoutHistory: [],
      error: 'Asset details could not be loaded. Try again.'
    });
  });

  it('preserves detail failure copy explicitly marked safe for users', async () => {
    const result = await loadWorkspaceAssetDetail(
      repository({ failure: Object.assign(new Error('That asset was archived.'), { safeForUser: true }) }),
      'tenant-home',
      'inventory-household',
      'missing'
    );

    expect(result.error).toBe('That asset was archived.');
  });

  it('suppresses generic adapter messages while preserving specific validation guidance', async () => {
    const generic = Object.assign(new Error('Invalid request.'), { safeForUser: true, status: 400, code: 'invalid_request' });
    const specific = Object.assign(new Error('Asset title must not be empty.'), { safeForUser: true, status: 422, code: 'validation_failed' });

    await expect(loadWorkspaceAssetDetail(repository({ failure: generic }), 'tenant-home', 'inventory-household', 'missing'))
      .resolves.toMatchObject({ error: 'Asset details could not be loaded. Try again.' });
    await expect(loadWorkspaceAssetDetail(repository({ failure: specific }), 'tenant-home', 'inventory-household', 'missing'))
      .resolves.toMatchObject({ error: 'Asset title must not be empty.' });
  });

  it('preserves typed authentication failures for the shell session boundary', async () => {
    await expect(
      loadWorkspaceAssetDetail(
        repository({ failure: new AuthenticationRequiredError('expired session diagnostic') }),
        'tenant-home',
        'inventory-household',
        'asset-one'
      )
    ).rejects.toBeInstanceOf(AuthenticationRequiredError);
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
    const checkoutHistory = [checkout('checkout-one')];

    expect(
      applyLoadedWorkspaceAssetDetail(workspaceData([original]), {
        loaded: true,
        asset: updated,
        attachments: detailAttachments,
        checkoutHistory,
        error: ''
      })
    ).toEqual({
      data: workspaceData([updated]),
      loadedAssetDetail: updated,
      selectedAssetId: 'asset-one',
      selectedAssetAttachments: detailAttachments,
      selectedAssetCheckoutHistory: checkoutHistory,
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

  it('separates populated custom fields from unset fields without losing falsey values', () => {
    const definitions = [
      field('serial', 'Serial number'),
      field('insured', 'Insured'),
      field('quantity', 'Quantity'),
      field('notes', 'Notes')
    ];

    expect(partitionAssetDetailFields(definitions, { serial: 'ABC-123', insured: false, quantity: 0, notes: '' }))
      .toEqual({ populated: definitions.slice(0, 3), unset: [definitions[3]] });
  });
});

function field(key: string, displayName: string): import('$lib/domain/inventory').CustomFieldDefinition {
  return {
    id: `field-${key}`,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    scope: 'inventory',
    key,
    displayName,
    type: 'text',
    enumOptions: [],
    applicability: 'all_assets',
    customAssetTypeIds: [],
    lifecycleState: 'active'
  };
}

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
    assets,
    checkedOutAssets: []
  };
}

interface RepositoryOptions {
  asset?: Asset;
  attachments?: AssetAttachment[];
  checkoutHistory?: AssetCheckout[];
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
    },
    async listAssetCheckoutHistory(tenantId: string, inventoryId: string, assetId: string): Promise<AssetCheckout[]> {
      options.calls?.push(`checkouts:${tenantId}:${inventoryId}:${assetId}`);
      if (options.failure) {
        throw options.failure;
      }
      return options.checkoutHistory ?? [];
    }
  };
}

function checkout(id: string): AssetCheckout {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    assetId: 'asset-one',
    state: 'open',
    checkedOutAt: '2026-06-24T11:00:00Z',
    checkedOutByPrincipalId: 'principal-one',
    createdAt: '2026-06-24T11:00:00Z',
    updatedAt: '2026-06-24T11:00:00Z'
  };
}
