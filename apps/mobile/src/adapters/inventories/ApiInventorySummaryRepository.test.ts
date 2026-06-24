import { describe, expect, it } from 'vitest';
import type {
  Asset,
  AssetPhotoReference,
  AssetSearchResult,
  Attachment,
  Inventory,
  Page,
  Tenant
} from '@stuff-stash/api-client';
import { assetId } from '../../domain/assets/AssetSummary';
import { inventoryId } from '../../domain/inventories/InventorySummary';
import { ApiInventorySummaryRepository } from './ApiInventorySummaryRepository';

class FakeInventoryApiClient {
  readonly tenant: Tenant = {
    id: 'tenant-home',
    name: 'Home',
    access: { relationship: 'owner', permissions: ['view', 'create_inventory', 'configure'] }
  };
  readonly cabinTenant: Tenant = {
    id: 'tenant-cabin',
    name: 'Cabin',
    access: { relationship: 'viewer', permissions: ['view'] }
  };
  readonly inventory: Inventory = {
    id: 'inventory-home',
    tenantId: 'tenant-home',
    name: 'Home Inventory',
    access: { relationship: 'owner', permissions: ['view', 'create_asset', 'edit_asset', 'share', 'configure'] }
  };
  readonly cabinInventory: Inventory = {
    id: 'inventory-cabin',
    tenantId: 'tenant-cabin',
    name: 'Cabin Inventory',
    access: { relationship: 'viewer', permissions: ['view'] }
  };
  readonly assets: readonly Asset[] = [
    {
      id: 'asset-garage',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      kind: 'location',
      title: 'Garage',
      description: 'Shelves and bins.',
      parentAssetId: null,
      lifecycleState: 'active'
    },
    {
      id: 'asset-filters',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      kind: 'item',
      title: 'Furnace filters',
      description: 'Three-pack of filters.',
      parentAssetId: 'asset-garage',
      lifecycleState: 'active'
    }
  ];
  createdAssetInput:
    | {
        readonly tenantId: string;
        readonly inventoryId: string;
        readonly title: string;
        readonly parentAssetId?: string;
      }
    | undefined;
  createdAttachmentInput:
    | {
        readonly tenantId: string;
        readonly inventoryId: string;
        readonly assetId: string;
        readonly fileName: string;
      }
    | undefined;
  searchedQuery: string | undefined;
  shouldFailAttachmentLookup = false;

  async listMyTenants(): Promise<Page<Tenant>> {
    return page([this.tenant, this.cabinTenant]);
  }

  async listInventories(tenantId: string): Promise<Page<Inventory>> {
    if (tenantId === this.cabinTenant.id) {
      return page([this.cabinInventory]);
    }

    return page([this.inventory]);
  }

  async listAssets(_tenantId: string, inventoryId: string): Promise<Page<Asset>> {
    if (inventoryId === this.cabinInventory.id) {
      return page([]);
    }

    return page(this.assets);
  }

  async listAssetAttachments(
    _tenantId: string,
    _inventoryId: string,
    assetIdValue: string
  ): Promise<Page<Attachment>> {
    if (this.shouldFailAttachmentLookup) {
      throw new Error('Attachment lookup failed.');
    }

    if (assetIdValue !== 'asset-filters') {
      return page([]);
    }

    return page([
      {
        id: 'attachment-filters-photo',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        assetId: assetIdValue,
        fileName: 'filters.jpg',
        contentType: 'image/jpeg',
        sizeBytes: 1024,
        lifecycleState: 'active'
      }
    ]);
  }

  async assetAttachmentThumbnailReference(
    tenantId: string,
    inventoryId: string,
    assetIdValue: string,
    attachmentId: string
  ): Promise<AssetPhotoReference> {
    return {
      uri: `https://api.example.test/tenants/${tenantId}/inventories/${inventoryId}/assets/${assetIdValue}/attachments/${attachmentId}/thumbnail?variant=small`,
      headers: { Authorization: 'Bearer dev-token' }
    };
  }

  async createAsset(
    tenantId: string,
    inventoryId: string,
    input: { readonly kind: 'item' | 'container' | 'location'; readonly title: string; readonly description?: string; readonly parentAssetId?: string | null }
  ): Promise<Asset> {
    this.createdAssetInput = {
      tenantId,
      inventoryId,
      title: input.title,
      parentAssetId: input.parentAssetId ?? undefined
    };

    return {
      id: 'asset-created',
      tenantId,
      inventoryId,
      kind: input.kind,
      title: input.title,
      description: input.description ?? '',
      parentAssetId: input.parentAssetId ?? null,
      lifecycleState: 'active'
    };
  }

  async createAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetIdValue: string,
    input: { readonly fileName: string; readonly contentType: 'image/jpeg' | 'image/png' | 'image/webp' | 'application/pdf'; readonly contentBase64: string }
  ): Promise<Attachment> {
    this.createdAttachmentInput = {
      tenantId,
      inventoryId,
      assetId: assetIdValue,
      fileName: input.fileName
    };

    return {
      id: 'attachment-created',
      tenantId,
      inventoryId,
      assetId: assetIdValue,
      fileName: input.fileName,
      contentType: input.contentType,
      sizeBytes: 4,
      lifecycleState: 'active'
    };
  }

  async searchAssets(tenantId: string, query: string): Promise<Page<AssetSearchResult>> {
    this.searchedQuery = `${tenantId}:${query}`;
    const asset = this.assets[1];

    if (!asset) {
      return page([]);
    }

    return page([
      {
        type: 'asset',
        tenantId,
        inventory: {
          id: this.inventory.id,
          name: this.inventory.name
        },
        asset,
        matches: []
      }
    ]);
  }
}

describe('ApiInventorySummaryRepository', () => {
  it('maps generated API client responses into mobile inventory summaries', async () => {
    const repository = new ApiInventorySummaryRepository(
      new FakeInventoryApiClient(),
      'tenant-home'
    );

    await expect(repository.getInventoryWorkspace()).resolves.toMatchObject({
      tenants: [
        { id: 'tenant-home', name: 'Home' },
        { id: 'tenant-cabin', name: 'Cabin' }
      ],
      defaultInventoryId: 'inventory-home',
      inventories: [
        {
          id: 'inventory-home',
          tenantId: 'tenant-home',
          name: 'Home Inventory',
          role: 'owner',
          locationCount: 1,
          locations: [
            {
              id: 'asset-garage',
              title: 'Garage',
              containedAssetCount: 1,
              recentAssetTitles: ['Furnace filters']
            }
          ],
          assets: [
            {
              id: 'asset-garage',
              locationLabel: 'Inventory root',
              locationTrail: ['Home Inventory', 'Garage']
            },
            {
              id: 'asset-filters',
              locationLabel: 'Garage',
              locationTrail: ['Home Inventory', 'Garage', 'Furnace filters'],
              hasPhoto: true,
              photo: {
                uri: 'https://api.example.test/tenants/tenant-home/inventories/inventory-home/assets/asset-filters/attachments/attachment-filters-photo/thumbnail?variant=small'
              }
            }
          ]
        },
        {
          id: 'inventory-cabin',
          tenantId: 'tenant-cabin',
          name: 'Cabin Inventory',
          role: 'viewer',
          locationCount: 0,
          locations: [],
          assets: []
        }
      ]
    });
  });

  it('uses the selected inventory for later default inventory operations', async () => {
    const repository = new ApiInventorySummaryRepository(
      new FakeInventoryApiClient(),
      'tenant-home'
    );

    await repository.selectInventory(inventoryId('inventory-cabin'));

    await expect(repository.getDefaultInventorySummary()).resolves.toMatchObject({
      id: 'inventory-cabin',
      tenantId: 'tenant-cabin',
      name: 'Cabin Inventory'
    });
  });

  it('keeps inventory loading usable when attachment metadata fails', async () => {
    const client = new FakeInventoryApiClient();
    client.shouldFailAttachmentLookup = true;
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getDefaultInventorySummary()).resolves.toMatchObject({
      id: 'inventory-home',
      assets: [
        { id: 'asset-garage', hasPhoto: false },
        { id: 'asset-filters', hasPhoto: false }
      ]
    });
  });

  it('creates and searches assets through the generated client wrapper', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(
      repository.createAsset({
        kind: 'item',
        title: 'USB-C charger pouch',
        description: 'Chargers and spare cables.',
        parentAssetId: assetId('asset-garage')
      })
    ).resolves.toMatchObject({
      id: 'asset-created',
      title: 'USB-C charger pouch',
      locationLabel: 'Garage'
    });
    expect(client.createdAssetInput).toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      title: 'USB-C charger pouch',
      parentAssetId: 'asset-garage'
    });

    await repository.addAssetPhoto(assetId('asset-created'), {
      fileName: 'created.jpg',
      contentType: 'image/jpeg',
      contentBase64: 'ZmFrZQ=='
    });
    expect(client.createdAttachmentInput).toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      assetId: 'asset-created',
      fileName: 'created.jpg'
    });

    await expect(repository.searchAssets('filters')).resolves.toMatchObject([
      {
        id: 'asset-filters',
        title: 'Furnace filters'
      }
    ]);
    expect(client.searchedQuery).toBe('tenant-home:filters');
  });
});

function page<T>(items: readonly T[]): Page<T> {
  return {
    items: [...items],
    pagination: {
      limit: items.length,
      nextCursor: null,
      hasMore: false
    }
  };
}
