import { describe, expect, it } from 'vitest';
import type {
  Asset,
  AssetTag,
  AssetCheckout,
  CheckedOutAsset,
  AssetPhotoReference,
  AssetSearchResult,
  Attachment,
  DirectUpload,
  Inventory,
  Page,
  Tenant
} from '@stuff-stash/api-client';
import { assetId } from '../../domain/assets/AssetSummary';
import { inventoryId, tenantId } from '../../domain/inventories/InventorySummary';
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
  assets: Asset[] = [
    {
      id: 'asset-garage',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      kind: 'location',
      title: 'Garage',
      description: 'Shelves and bins.',
      parentAssetId: null,
      lifecycleState: 'active',
      customFields: {},
      createdAt: '2026-06-20T10:00:00Z',
      updatedAt: '2026-06-22T10:00:00Z'
    },
    {
      id: 'asset-filters',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      kind: 'item',
      title: 'Furnace filters',
      description: 'Three-pack of filters.',
      parentAssetId: 'asset-garage',
      lifecycleState: 'active',
      customFields: {},
      tags: [{ id: 'tag-workshop', key: 'workshop', displayName: 'Workshop', color: '#2F80ED' }],
      createdAt: '2026-06-21T10:00:00Z',
      updatedAt: '2026-06-23T10:00:00Z',
      primaryPhoto: {
        id: 'attachment-filters-photo',
        fileName: 'filters.jpg',
        contentType: 'image/jpeg',
        sizeBytes: 1024,
        thumbnails: {
          small: '/tenants/tenant-home/inventories/inventory-home/assets/asset-filters/attachments/attachment-filters-photo/thumbnail?variant=small',
          medium: '/tenants/tenant-home/inventories/inventory-home/assets/asset-filters/attachments/attachment-filters-photo/thumbnail?variant=medium',
          large: '/tenants/tenant-home/inventories/inventory-home/assets/asset-filters/attachments/attachment-filters-photo/thumbnail?variant=large'
        }
      }
    }
  ];
  listAssetRequests: Array<{
    readonly inventoryId: string;
    readonly limit?: number;
    readonly cursor?: string;
    readonly lifecycleState?: string;
    readonly sort?: string;
  }> = [];
  listCheckedOutAssetRequests: Array<{
    readonly inventoryId: string;
    readonly limit?: number;
    readonly cursor?: string;
  }> = [];
  listAttachmentRequests: Array<{
    readonly assetId: string;
    readonly limit?: number;
    readonly cursor?: string;
  }> = [];
  listAssetTagRequests: Array<{
    readonly tenantId: string;
    readonly inventoryId: string;
    readonly limit?: number;
    readonly cursor?: string;
  }> = [];
  paginatedAssetTags = false;
  thumbnailRequests: Array<{
    readonly assetId: string;
    readonly attachmentId: string;
    readonly variant: string;
  }> = [];
  createdAssetInput:
    | {
        readonly tenantId: string;
        readonly inventoryId: string;
        readonly title: string;
        readonly parentAssetId?: string;
        readonly tagIds?: readonly string[];
      }
    | undefined;
  createdAssetTagInput:
    | {
        readonly tenantId: string;
        readonly inventoryId: string;
        readonly displayName: string;
        readonly color?: string;
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
  updatedAssetInput:
    | {
        readonly tenantId: string;
        readonly inventoryId: string;
        readonly assetId: string;
        readonly title?: string;
        readonly description?: string;
        readonly parentAssetId?: string | null;
        readonly tagIds?: readonly string[];
      }
    | undefined;
  initiatedDirectUploadInput:
    | {
        readonly tenantId: string;
        readonly inventoryId: string;
        readonly assetId: string;
        readonly fileName: string;
        readonly sizeBytes: number;
      }
    | undefined;
  completedDirectUploadInput:
    | {
        readonly tenantId: string;
        readonly inventoryId: string;
        readonly assetId: string;
        readonly uploadId: string;
      }
    | undefined;
  deletedAttachmentInput:
    | {
        readonly tenantId: string;
        readonly inventoryId: string;
        readonly assetId: string;
        readonly attachmentId: string;
      }
    | undefined;
  directUploadURL = 'https://uploads.example.test/object-one';
  lifecycleInputs: Array<{
    readonly action: 'archive' | 'restore' | 'delete';
    readonly tenantId: string;
    readonly inventoryId: string;
    readonly assetId: string;
  }> = [];
  checkoutInputs: Array<{
    readonly action: 'checkout' | 'return';
    readonly tenantId: string;
    readonly inventoryId: string;
    readonly assetId: string;
    readonly details?: string;
  }> = [];
  searchedQuery: string | undefined;
  searchAssetRequests: Array<{
    readonly tenantId: string;
    readonly query: string;
    readonly cursor?: string;
    readonly inventoryId?: string;
    readonly tagIds?: readonly string[];
    readonly lifecycleState?: string;
    readonly checkoutState?: string;
  }> = [];
  shouldFailAttachmentLookup = false;
  failedThumbnailAssetIds = new Set<string>();

  async listMyTenants(): Promise<Page<Tenant>> {
    return page([this.tenant, this.cabinTenant]);
  }

  async listInventories(tenantId: string): Promise<Page<Inventory>> {
    if (tenantId === this.cabinTenant.id) {
      return page([this.cabinInventory]);
    }

    return page([this.inventory]);
  }

  async listAssets(
    _tenantId: string,
    inventoryId: string,
    limit = 50,
    cursor?: string,
    lifecycleState?: string,
    sort?: string
  ): Promise<Page<Asset>> {
    this.listAssetRequests.push({ inventoryId, limit, cursor, lifecycleState, sort });
    if (inventoryId === this.cabinInventory.id) {
      return page([]);
    }

    const lifecycleAssets = lifecycleState === 'active'
      ? this.assets.filter((asset) => asset.lifecycleState === 'active')
      : lifecycleState === 'archived'
        ? this.assets.filter((asset) => asset.lifecycleState === 'archived')
        : this.assets;
    const sortedAssets = sort === 'updated_desc' ? sortAssetsByUpdatedDesc(lifecycleAssets) : lifecycleAssets;
    const start = cursor ? Number.parseInt(cursor, 10) : 0;
    const items = sortedAssets.slice(start, start + limit);
    const nextCursor =
      start + limit < sortedAssets.length ? (start + limit).toString() : null;

    return pageWithCursor(items, nextCursor);
  }

  async listAssetTags(
    tenantId: string,
    inventoryId: string,
    limit?: number,
    cursor?: string
  ): Promise<Page<AssetTag>> {
    this.listAssetTagRequests.push({ tenantId, inventoryId, limit, cursor });
    if (inventoryId !== this.inventory.id) {
      return page([]);
    }
    if (this.paginatedAssetTags && cursor === undefined) {
      return pageWithCursor([
        {
          id: 'tag-workshop',
          tenantId,
          inventoryId,
          key: 'workshop',
          displayName: 'Workshop',
          color: '#2F80ED',
          lifecycleState: 'active',
          createdAt: '2026-06-20T10:00:00Z',
          updatedAt: '2026-06-20T10:00:00Z'
        }
      ], 'next-tags');
    }
    if (this.paginatedAssetTags && cursor === 'next-tags') {
      return page([
        {
          id: 'tag-camping',
          tenantId,
          inventoryId,
          key: 'camping',
          displayName: 'Camping',
          color: '#2E7D32',
          lifecycleState: 'active',
          createdAt: '2026-06-20T10:00:00Z',
          updatedAt: '2026-06-20T10:00:00Z'
        }
      ]);
    }
    return page([
      {
        id: 'tag-workshop',
        tenantId,
        inventoryId,
        key: 'workshop',
        displayName: 'Workshop',
        color: '#2F80ED',
        lifecycleState: 'active',
        createdAt: '2026-06-20T10:00:00Z',
        updatedAt: '2026-06-20T10:00:00Z'
      }
    ]);
  }

  async createAssetTag(
    tenantId: string,
    inventoryId: string,
    input: { readonly displayName: string; readonly color?: string }
  ): Promise<AssetTag> {
    this.createdAssetTagInput = {
      tenantId,
      inventoryId,
      displayName: input.displayName,
      color: input.color
    };
    return {
      id: 'tag-created',
      tenantId,
      inventoryId,
      key: input.displayName.toLowerCase().replaceAll(' ', '-'),
      displayName: input.displayName,
      color: input.color,
      lifecycleState: 'active',
      createdAt: '2026-06-20T10:00:00Z',
      updatedAt: '2026-06-20T10:00:00Z'
    };
  }

  async listAssetAttachments(
    _tenantId: string,
    _inventoryId: string,
    assetIdValue: string,
    limit?: number,
    cursor?: string
  ): Promise<Page<Attachment>> {
    this.listAttachmentRequests.push({ assetId: assetIdValue, limit, cursor });
    if (this.shouldFailAttachmentLookup) {
      throw new Error('Attachment lookup failed.');
    }

    if (assetIdValue === 'asset-many-photos') {
      const firstPhoto = {
        id: 'attachment-many-one',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        assetId: assetIdValue,
        fileName: 'many-one.jpg',
        contentType: 'image/jpeg',
        sizeBytes: 1024,
        lifecycleState: 'active' as const
      };
      const secondPhoto = {
        ...firstPhoto,
        id: 'attachment-many-two',
        fileName: 'many-two.jpg'
      };
      return cursor ? page([secondPhoto]) : pageWithCursor([firstPhoto], 'next-photo-page');
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
      },
      {
        id: 'attachment-filters-label',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        assetId: assetIdValue,
        fileName: 'filters-label.jpg',
        contentType: 'image/jpeg',
        sizeBytes: 512,
        lifecycleState: 'active'
      }
    ]);
  }

  async assetAttachmentThumbnailReference(
    tenantId: string,
    inventoryId: string,
    assetIdValue: string,
    attachmentId: string,
    variant: 'small' | 'medium' | 'large' = 'small'
  ): Promise<AssetPhotoReference> {
    this.thumbnailRequests.push({ assetId: assetIdValue, attachmentId, variant });
    if (this.failedThumbnailAssetIds.has(assetIdValue)) {
      throw new Error('Thumbnail reference failed.');
    }
    return {
      uri: `https://api.example.test/tenants/${tenantId}/inventories/${inventoryId}/assets/${assetIdValue}/attachments/${attachmentId}/thumbnail?variant=${variant}`,
      headers: { Authorization: 'Bearer dev-token' }
    };
  }

  async createAsset(
    tenantId: string,
    inventoryId: string,
    input: { readonly kind: 'item' | 'container' | 'location'; readonly title: string; readonly description?: string; readonly parentAssetId?: string | null; readonly tagIds?: readonly string[] }
  ): Promise<Asset> {
    this.createdAssetInput = {
      tenantId,
      inventoryId,
      title: input.title,
      parentAssetId: input.parentAssetId ?? undefined,
      tagIds: input.tagIds
    };

    return {
      id: 'asset-created',
      tenantId,
      inventoryId,
      kind: input.kind,
      title: input.title,
      description: input.description ?? '',
      parentAssetId: input.parentAssetId ?? null,
      lifecycleState: 'active',
      customFields: {},
      createdAt: '2026-06-24T10:00:00Z',
      updatedAt: '2026-06-24T10:00:00Z'
    };
  }

  async updateAsset(
    tenantId: string,
    inventoryId: string,
    assetIdValue: string,
    input: { readonly title?: string; readonly description?: string; readonly parentAssetId?: string | null; readonly tagIds?: readonly string[] }
  ): Promise<Asset> {
    this.updatedAssetInput = {
      tenantId,
      inventoryId,
      assetId: assetIdValue,
      title: input.title,
      description: input.description,
      parentAssetId: input.parentAssetId,
      tagIds: input.tagIds
    };
    const current = this.assets.find((asset) => asset.id === assetIdValue);
    if (!current) {
      throw new Error('Asset not found.');
    }
    return {
      ...current,
      title: input.title ?? current.title,
      description: input.description ?? current.description,
      parentAssetId: input.parentAssetId === undefined ? current.parentAssetId : input.parentAssetId,
      updatedAt: '2026-06-25T10:00:00Z'
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

  async initiateAssetAttachmentDirectUpload(
    tenantId: string,
    inventoryId: string,
    assetIdValue: string,
    input: { readonly fileName: string; readonly contentType: 'image/jpeg' | 'image/png' | 'image/webp' | 'application/pdf'; readonly sizeBytes: number }
  ): Promise<DirectUpload> {
    this.initiatedDirectUploadInput = {
      tenantId,
      inventoryId,
      assetId: assetIdValue,
      fileName: input.fileName,
      sizeBytes: input.sizeBytes
    };
    return {
      uploadId: 'upload-one',
      attachmentId: 'attachment-one',
      method: 'PUT',
      url: this.directUploadURL,
      headers: { 'Content-Type': input.contentType },
      formFields: {},
      expiresAt: '2026-06-24T10:15:00Z'
    };
  }

  async completeAssetAttachmentDirectUpload(
    tenantId: string,
    inventoryId: string,
    assetIdValue: string,
    uploadId: string
  ): Promise<Attachment> {
    this.completedDirectUploadInput = {
      tenantId,
      inventoryId,
      assetId: assetIdValue,
      uploadId
    };
    return {
      id: 'attachment-one',
      tenantId,
      inventoryId,
      assetId: assetIdValue,
      fileName: 'uploaded.jpg',
      contentType: 'image/jpeg',
      sizeBytes: 8,
      lifecycleState: 'active'
    };
  }

  async deleteAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetIdValue: string,
    attachmentId: string
  ): Promise<void> {
    this.deletedAttachmentInput = {
      tenantId,
      inventoryId,
      assetId: assetIdValue,
      attachmentId
    };
  }

  async archiveAsset(tenantId: string, inventoryId: string, assetIdValue: string): Promise<Asset> {
    this.lifecycleInputs.push({
      action: 'archive',
      tenantId,
      inventoryId,
      assetId: assetIdValue
    });

    return lifecycleAsset(this.assets, assetIdValue, 'archived');
  }

  async restoreAsset(tenantId: string, inventoryId: string, assetIdValue: string): Promise<Asset> {
    this.lifecycleInputs.push({
      action: 'restore',
      tenantId,
      inventoryId,
      assetId: assetIdValue
    });

    return lifecycleAsset(this.assets, assetIdValue, 'active');
  }

  async deleteAsset(tenantId: string, inventoryId: string, assetIdValue: string): Promise<void> {
    this.lifecycleInputs.push({
      action: 'delete',
      tenantId,
      inventoryId,
      assetId: assetIdValue
    });
  }

  async checkoutAsset(
    tenantId: string,
    inventoryId: string,
    assetIdValue: string,
    input: { readonly details?: string } = {}
  ): Promise<AssetCheckout> {
    this.checkoutInputs.push({
      action: 'checkout',
      tenantId,
      inventoryId,
      assetId: assetIdValue,
      details: input.details
    });
    return checkoutRecord(assetIdValue, 'open', input.details);
  }

  async returnAsset(
    tenantId: string,
    inventoryId: string,
    assetIdValue: string,
    input: { readonly details?: string } = {}
  ): Promise<AssetCheckout> {
    this.checkoutInputs.push({
      action: 'return',
      tenantId,
      inventoryId,
      assetId: assetIdValue,
      details: input.details
    });
    return checkoutRecord(assetIdValue, 'returned', input.details);
  }

  async searchAssets(
    tenantId: string,
    query: string,
    options?: {
      readonly cursor?: string;
      readonly inventoryId?: string;
      readonly tagIds?: readonly string[];
      readonly lifecycleState?: string;
      readonly checkoutState?: string;
    }
  ): Promise<Page<AssetSearchResult>> {
    this.searchedQuery = `${tenantId}:${query}`;
    this.searchAssetRequests.push({
      tenantId,
      query,
      cursor: options?.cursor,
      inventoryId: options?.inventoryId,
      tagIds: options?.tagIds,
      lifecycleState: options?.lifecycleState,
      checkoutState: options?.checkoutState
    });
    const asset = this.assets[1];

    if (!asset) {
      return page([]);
    }

    if (query === 'tagged') {
      return page([
        {
          type: 'asset',
          tenantId,
          inventory: {
            id: this.inventory.id,
            name: this.inventory.name
          },
          asset,
          matches: [
            { field: 'tag_display_name', value: 'Workshop' },
            { field: 'tag_key', value: 'workshop' }
          ]
        }
      ]);
    }

    if (query === 'paged' && options?.cursor === undefined) {
      return pageWithCursor(
        [
          {
            type: 'asset',
            tenantId,
            inventory: {
              id: 'inventory-other',
              name: 'Other inventory'
            },
            asset: {
              ...asset,
              id: 'asset-other-inventory',
              inventoryId: 'inventory-other',
              title: 'Other inventory paged result'
            },
            matches: []
          }
        ],
        'next-page'
      );
    }
    if (query === 'sixth-page') {
      const cursorNumber = options?.cursor ? Number.parseInt(options.cursor, 10) : 0;
      if (cursorNumber < 5) {
        return pageWithCursor(
          [
            {
              type: 'asset',
              tenantId,
              inventory: {
                id: 'inventory-other',
                name: 'Other inventory'
              },
              asset: {
                ...asset,
                id: `asset-other-page-${cursorNumber.toString()}`,
                inventoryId: 'inventory-other',
                title: 'Other inventory page result'
              },
              matches: []
            }
          ],
          (cursorNumber + 1).toString()
        );
      }
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
      },
      {
        type: 'asset',
        tenantId,
        inventory: {
          id: 'inventory-other',
          name: 'Other inventory'
        },
        asset: {
          ...asset,
          id: 'asset-other-inventory',
          inventoryId: 'inventory-other',
          title: 'Other inventory filters'
        },
        matches: []
      }
    ]);
  }

  async listCheckedOutAssets(
    _tenantId: string,
    inventoryId: string,
    limit?: number,
    cursor?: string
  ): Promise<Page<CheckedOutAsset>> {
    this.listCheckedOutAssetRequests.push({ inventoryId, limit, cursor });
    const items = this.assets
      .filter((asset) => asset.currentCheckout !== undefined)
      .map((asset) => ({
        asset,
        checkout: asset.currentCheckout!
      }));
    return page(items);
  }
}

class FakeDirectUploadTransport {
  readonly uploads: Array<{
    readonly url: string;
    readonly fileUri: string;
    readonly fileName: string;
    readonly contentType: string;
  }> = [];

  constructor(private readonly result = true) {}

  async upload(input: {
    readonly upload: DirectUpload;
    readonly fileUri: string;
    readonly fileName: string;
    readonly contentType: string;
  }): Promise<boolean> {
    this.uploads.push({
      url: input.upload.url,
      fileUri: input.fileUri,
      fileName: input.fileName,
      contentType: input.contentType
    });
    return this.result;
  }
}

describe('ApiInventorySummaryRepository', () => {
  it('maps generated API client responses into mobile inventory summaries', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

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
          assetTags: [
            {
              id: 'tag-workshop',
              key: 'workshop',
              displayName: 'Workshop',
              color: '#2F80ED'
            }
          ],
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
              id: 'asset-filters',
              locationLabel: 'Garage',
              locationTrail: ['Home Inventory', 'Garage', 'Furnace filters'],
              hasPhoto: true,
              photos: [
                {
                  id: 'attachment-filters-photo',
                  fileName: 'filters.jpg'
                }
              ],
              photo: {
                uri: 'https://api.example.test/tenants/tenant-home/inventories/inventory-home/assets/asset-filters/attachments/attachment-filters-photo/thumbnail?variant=small',
                heroUri: 'https://api.example.test/tenants/tenant-home/inventories/inventory-home/assets/asset-filters/attachments/attachment-filters-photo/thumbnail?variant=medium',
                heroHeaders: { Authorization: 'Bearer dev-token' },
                viewerUri: 'https://api.example.test/tenants/tenant-home/inventories/inventory-home/assets/asset-filters/attachments/attachment-filters-photo/thumbnail?variant=large',
                viewerHeaders: { Authorization: 'Bearer dev-token' }
              }
            },
            {
              id: 'asset-garage',
              locationLabel: 'Inventory root',
              locationTrail: ['Home Inventory', 'Garage']
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
    expect(client.listAssetTagRequests).toContainEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      limit: 100,
      cursor: undefined
    });
  });

  it('loads every active asset tag page for mobile selection', async () => {
    const client = new FakeInventoryApiClient();
    client.paginatedAssetTags = true;
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getDefaultInventorySummary()).resolves.toMatchObject({
      assetTags: [
        { id: 'tag-workshop', displayName: 'Workshop' },
        { id: 'tag-camping', displayName: 'Camping' }
      ]
    });
    expect(client.listAssetTagRequests).toContainEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      limit: 100,
      cursor: undefined
    });
    expect(client.listAssetTagRequests).toContainEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      limit: 100,
      cursor: 'next-tags'
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

  it('does not list attachments while loading dense inventory summaries', async () => {
    const client = new FakeInventoryApiClient();
    client.shouldFailAttachmentLookup = true;
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getDefaultInventorySummary()).resolves.toMatchObject({
      id: 'inventory-home',
      assets: [
        { id: 'asset-filters', hasPhoto: true },
        { id: 'asset-garage', hasPhoto: false }
      ]
    });
    expect(client.listAttachmentRequests).toEqual([]);
  });

  it('loads the complete active image attachment set for asset detail', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getAssetDetail({
      tenantId: tenantId('tenant-home'),
      inventoryId: inventoryId('inventory-home'),
      asset: {
        id: assetId('asset-filters'),
        title: 'Furnace filters',
        kind: 'item',
        lifecycleState: 'active',
        parentAssetId: assetId('asset-garage'),
        locationLabel: 'Garage',
        locationTrail: ['Home Inventory', 'Garage', 'Furnace filters'],
        description: 'Three-pack of filters.',
        updatedAtLabel: 'Updated today',
        hasPhoto: true
      }
    })).resolves.toMatchObject({
      id: 'asset-filters',
      title: 'Furnace filters',
      hasPhoto: true,
      photos: [
        {
          id: 'attachment-filters-photo',
          fileName: 'filters.jpg',
          uri: 'https://api.example.test/tenants/tenant-home/inventories/inventory-home/assets/asset-filters/attachments/attachment-filters-photo/thumbnail?variant=small',
          heroUri: 'https://api.example.test/tenants/tenant-home/inventories/inventory-home/assets/asset-filters/attachments/attachment-filters-photo/thumbnail?variant=medium',
          viewerUri: 'https://api.example.test/tenants/tenant-home/inventories/inventory-home/assets/asset-filters/attachments/attachment-filters-photo/thumbnail?variant=large'
        },
        {
          id: 'attachment-filters-label',
          fileName: 'filters-label.jpg'
        }
      ]
    });
    expect(client.listAttachmentRequests).toEqual([
      { assetId: 'asset-filters', limit: 50, cursor: undefined }
    ]);
    const detailThumbnailRequests = client.thumbnailRequests.slice(-6);
    expect(detailThumbnailRequests).toHaveLength(6);
    expect(detailThumbnailRequests).toEqual(expect.arrayContaining([
      { assetId: 'asset-filters', attachmentId: 'attachment-filters-photo', variant: 'small' },
      { assetId: 'asset-filters', attachmentId: 'attachment-filters-photo', variant: 'medium' },
      { assetId: 'asset-filters', attachmentId: 'attachment-filters-photo', variant: 'large' },
      { assetId: 'asset-filters', attachmentId: 'attachment-filters-label', variant: 'small' },
      { assetId: 'asset-filters', attachmentId: 'attachment-filters-label', variant: 'medium' },
      { assetId: 'asset-filters', attachmentId: 'attachment-filters-label', variant: 'large' }
    ]));
  });

  it('does not collapse asset detail attachment lookup failures into an empty photo set', async () => {
    const client = new FakeInventoryApiClient();
    client.shouldFailAttachmentLookup = true;
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getAssetDetail({
      tenantId: tenantId('tenant-home'),
      inventoryId: inventoryId('inventory-home'),
      asset: {
        id: assetId('asset-filters'),
        title: 'Furnace filters',
        kind: 'item',
        lifecycleState: 'active',
        locationLabel: 'Garage',
        locationTrail: ['Home Inventory', 'Garage', 'Furnace filters'],
        description: 'Three-pack of filters.',
        updatedAtLabel: 'Updated today',
        hasPhoto: true
      }
    })).rejects.toThrow('Asset attachments could not be loaded.');
  });

  it('requests API-owned updated-descending asset order for mobile recency', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await repository.getDefaultInventorySummary();

    expect(client.listAssetRequests).toContainEqual({
      inventoryId: 'inventory-home',
      limit: 100,
      cursor: undefined,
      lifecycleState: 'all',
      sort: 'updated_desc'
    });
  });

  it('preserves API-provided recency order across asset kinds', async () => {
    const client = new FakeInventoryApiClient();
    client.assets = [
      {
        id: 'asset-new-batteries',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'item',
        title: 'Fresh batteries',
        description: 'Just created from the Add sheet.',
        parentAssetId: 'asset-garage',
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-24T11:00:00Z',
        updatedAt: '2026-06-24T11:00:00Z'
      },
      ...client.assets
    ];
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getDefaultInventorySummary()).resolves.toMatchObject({
      assets: [
        {
          id: 'asset-new-batteries',
          kind: 'item',
          title: 'Fresh batteries'
        },
        {
          id: 'asset-filters',
          kind: 'item',
          title: 'Furnace filters'
        },
        {
          id: 'asset-garage',
          kind: 'location',
          title: 'Garage'
        }
      ]
    });
  });

  it('loads locations from the full active inventory tree instead of only the recent summary page', async () => {
    const client = new FakeInventoryApiClient();
    client.assets = [
      ...Array.from({ length: 100 }, (_, index): Asset => ({
        id: `asset-recent-item-${index.toString().padStart(3, '0')}`,
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'item',
        title: `Recent item ${index.toString().padStart(3, '0')}`,
        description: '',
        parentAssetId: null,
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-25T10:00:00Z',
        updatedAt: `2026-06-25T${(10 + Math.floor(index / 60)).toString().padStart(2, '0')}:${(index % 60).toString().padStart(2, '0')}:00Z`
      })),
      {
        id: 'asset-late-location',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'location',
        title: 'Late page closet',
        description: '',
        parentAssetId: null,
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-20T10:00:00Z',
        updatedAt: '2026-06-20T10:00:00Z'
      },
      {
        id: 'asset-late-child',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'item',
        title: 'Stored blanket',
        description: '',
        parentAssetId: 'asset-late-location',
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-20T10:01:00Z',
        updatedAt: '2026-06-20T10:01:00Z'
      },
      {
        id: 'asset-late-newer-child',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'item',
        title: 'Newer stored blanket',
        description: '',
        parentAssetId: 'asset-late-location',
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-20T10:02:00Z',
        updatedAt: '2026-06-20T10:02:00Z'
      }
    ];
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getDefaultInventorySummary()).resolves.toMatchObject({
      locationCount: 1,
      locations: [
        {
          id: 'asset-late-location',
          title: 'Late page closet',
          containedAssetCount: 2,
          recentAssetTitles: ['Newer stored blanket', 'Stored blanket']
        }
      ],
      assets: expect.not.arrayContaining([
        expect.objectContaining({ id: 'asset-late-location' })
      ])
    });
    expect(client.listAssetRequests.filter((request) => request.inventoryId === 'inventory-home')).toEqual([
      {
        inventoryId: 'inventory-home',
        limit: 100,
        cursor: undefined,
        lifecycleState: 'all',
        sort: 'updated_desc'
      },
      {
        inventoryId: 'inventory-home',
        limit: 100,
        cursor: undefined,
        lifecycleState: 'active',
        sort: 'id_asc'
      },
      {
        inventoryId: 'inventory-home',
        limit: 100,
        cursor: '100',
        lifecycleState: 'active',
        sort: 'id_asc'
      }
    ]);
  });

  it('lists paged selected-inventory assets for browse mode', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(
      repository.browseAssets({
        query: '',
        cursor: undefined,
        limit: 1,
        lifecycleState: 'active',
        checkoutState: 'any',
        kind: 'item',
        sort: 'updated_desc'
      })
    ).resolves.toMatchObject({
      assets: [
        {
          id: 'asset-filters',
          kind: 'item',
          title: 'Furnace filters'
        }
      ],
      hasMore: true
    });
    expect(client.listAssetRequests).toContainEqual({
      inventoryId: 'inventory-home',
      limit: 1,
      cursor: undefined,
      lifecycleState: 'active',
      sort: 'updated_desc'
    });
  });

  it('continues list pagination until a kind-filtered browse page has matches', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(
      repository.browseAssets({
        query: '',
        limit: 1,
        lifecycleState: 'active',
        checkoutState: 'any',
        kind: 'item',
        sort: 'id_asc'
      })
    ).resolves.toMatchObject({
      assets: [
        {
          id: 'asset-filters',
          kind: 'item',
          title: 'Furnace filters'
        }
      ],
      hasMore: false
    });
    expect(client.listAssetRequests).toContainEqual({
      inventoryId: 'inventory-home',
      limit: 1,
      cursor: undefined,
      lifecycleState: 'active',
      sort: 'id_asc'
    });
    expect(client.listAssetRequests).toContainEqual({
      inventoryId: 'inventory-home',
      limit: 1,
      cursor: '1',
      lifecycleState: 'active',
      sort: 'id_asc'
    });
  });

  it('uses the checked-out inventory endpoint for checked-out browse mode', async () => {
    const client = new FakeInventoryApiClient();
    client.assets = [
      {
        ...client.assets[1]!,
        currentCheckout: {
          id: 'checkout-filters',
          state: 'open',
          checkedOutAt: '2026-06-25T12:00:00Z',
          checkedOutByPrincipalId: 'principal-home'
        }
      },
      client.assets[0]!
    ];
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(
      repository.browseAssets({
        query: '',
        limit: 10,
        lifecycleState: 'active',
        checkoutState: 'checked_out',
        kind: 'item',
        sort: 'updated_desc'
      })
    ).resolves.toMatchObject({
      assets: [
        {
          id: 'asset-filters',
          currentCheckout: {
            id: 'checkout-filters',
            state: 'open',
            checkedOutAt: '2026-06-25T12:00:00Z',
            checkedOutByPrincipalId: 'principal-home'
          }
        }
      ],
      hasMore: false
    });
    expect(client.listCheckedOutAssetRequests).toEqual([
      {
        inventoryId: 'inventory-home',
        limit: 10,
        cursor: undefined
      }
    ]);
  });

  it('searches paged selected-inventory assets with lifecycle filtering', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(
      repository.browseAssets({
        query: 'filters',
        cursor: 'next-page',
        limit: 10,
        lifecycleState: 'all',
        checkoutState: 'any',
        kind: 'item',
        sort: 'updated_desc'
      })
    ).resolves.toMatchObject({
      assets: [
        {
          id: 'asset-filters',
          kind: 'item',
          title: 'Furnace filters'
        }
      ]
    });
    expect(client.searchedQuery).toBe('tenant-home:filters');
  });

  it('searches selected-inventory assets with multi-tag filters without replacing the query', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await repository.browseAssets({
      query: '',
      cursor: undefined,
      limit: 10,
      lifecycleState: 'active',
      checkoutState: 'any',
      kind: 'item',
      sort: 'updated_desc',
      tagIds: ['tag-workshop', 'tag-camping']
    });

    expect(client.listAssetRequests).toEqual(expect.arrayContaining([
      expect.objectContaining({
        inventoryId: 'inventory-home',
        lifecycleState: 'all',
        sort: 'updated_desc'
      }),
      expect.objectContaining({
        inventoryId: 'inventory-home',
        lifecycleState: 'active',
        sort: 'id_asc'
      })
    ]));
    expect(client.searchAssetRequests[0]).toMatchObject({
      tenantId: 'tenant-home',
      query: '',
      inventoryId: 'inventory-home',
      tagIds: ['tag-workshop', 'tag-camping'],
      lifecycleState: 'active',
      checkoutState: 'any'
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
        parentAssetId: assetId('asset-garage'),
        tagIds: ['tag-workshop']
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
      parentAssetId: 'asset-garage',
      tagIds: ['tag-workshop']
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
        title: 'Furnace filters',
        tags: [{ id: 'tag-workshop', displayName: 'Workshop', color: '#2F80ED' }]
      }
    ]);
    await expect(repository.searchAssets('filters')).resolves.toHaveLength(1);
    expect(client.searchedQuery).toBe('tenant-home:filters');
  });

  it('maps tag-backed search matches to a user-facing mobile label', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.browseAssets({
      query: 'tagged',
      lifecycleState: 'active',
      checkoutState: 'any',
      kind: 'all',
      sort: 'updated_desc',
      limit: 20
    })).resolves.toMatchObject({
      assets: [
        {
          id: 'asset-filters'
        }
      ],
      searchMatches: [
        {
          assetId: 'asset-filters',
          labels: ['Tag']
        }
      ]
    });
  });

  it('creates asset tags through the generated client wrapper', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.createAssetTag({
      displayName: 'Camping',
      color: '#2F80ED'
    })).resolves.toEqual({
      id: 'tag-created',
      key: 'camping',
      displayName: 'Camping',
      color: '#2F80ED'
    });

    expect(client.createdAssetTagInput).toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      displayName: 'Camping',
      color: '#2F80ED'
    });
  });

  it('updates asset fields and parent placement through the generated client wrapper', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.updateAsset({
      assetId: assetId('asset-filters'),
      title: 'HEPA filters',
      description: 'Replacement filters.',
      parentAssetId: null,
      tagIds: ['tag-workshop']
    })).resolves.toMatchObject({
      id: 'asset-filters',
      title: 'HEPA filters',
      parentAssetId: undefined,
      locationLabel: 'Inventory root'
    });

    expect(client.updatedAssetInput).toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      assetId: 'asset-filters',
      title: 'HEPA filters',
      description: 'Replacement filters.',
      parentAssetId: null,
      tagIds: ['tag-workshop']
    });
  });

  it('maps nested parent trails for asset workspace paths', async () => {
    const client = new FakeInventoryApiClient();
    client.assets = [
      ...client.assets,
      {
        id: 'asset-cabinet',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'container',
        title: 'Big cabinet',
        description: '',
        parentAssetId: 'asset-garage',
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-24T10:00:00Z',
        updatedAt: '2026-06-24T10:00:00Z'
      },
      {
        id: 'asset-shelf',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'container',
        title: 'Second shelf',
        description: '',
        parentAssetId: 'asset-cabinet',
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-25T10:00:00Z',
        updatedAt: '2026-06-25T10:00:00Z'
      }
    ];
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.browseAssets({
      query: '',
      limit: 10,
      lifecycleState: 'all',
      checkoutState: 'any',
      kind: 'all',
      sort: 'updated_desc'
    })).resolves.toMatchObject({
      assets: expect.arrayContaining([
        expect.objectContaining({
          id: 'asset-shelf',
          locationTrail: ['Home Inventory', 'Garage', 'Big cabinet', 'Second shelf']
        })
      ])
    });
  });

  it('uses primary photo summaries for browse cards without paged attachment lookups', async () => {
    const client = new FakeInventoryApiClient();
    client.assets = [
      {
        id: 'asset-many-photos',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'item',
        title: 'Photo album',
        description: '',
        parentAssetId: 'asset-garage',
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-25T10:00:00Z',
        updatedAt: '2026-06-25T10:00:00Z',
        primaryPhoto: {
          id: 'attachment-many-one',
          fileName: 'many-one.jpg',
          contentType: 'image/jpeg',
          sizeBytes: 1024,
          thumbnails: {
            small: '/tenants/tenant-home/inventories/inventory-home/assets/asset-many-photos/attachments/attachment-many-one/thumbnail?variant=small',
            medium: '/tenants/tenant-home/inventories/inventory-home/assets/asset-many-photos/attachments/attachment-many-one/thumbnail?variant=medium',
            large: '/tenants/tenant-home/inventories/inventory-home/assets/asset-many-photos/attachments/attachment-many-one/thumbnail?variant=large'
          }
        }
      },
      ...client.assets
    ];
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.browseAssets({
      query: '',
      limit: 10,
      lifecycleState: 'all',
      checkoutState: 'any',
      kind: 'all',
      sort: 'updated_desc'
    })).resolves.toMatchObject({
      assets: expect.arrayContaining([
        expect.objectContaining({
          id: 'asset-many-photos',
          photos: [
            expect.objectContaining({ id: 'attachment-many-one', fileName: 'many-one.jpg' })
          ]
        })
      ])
    });
    expect(client.listAttachmentRequests).toEqual([]);
  });

  it('prefers direct upload targets when adding asset photos', async () => {
    const client = new FakeInventoryApiClient();
    const directUploads = new FakeDirectUploadTransport();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home', directUploads);

    await repository.addAssetPhoto(assetId('asset-created'), {
      fileName: 'created.jpg',
      contentType: 'image/jpeg',
      uri: 'file:///created.jpg',
      sizeBytes: 4
    });

    expect(client.initiatedDirectUploadInput).toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      assetId: 'asset-created',
      fileName: 'created.jpg',
      sizeBytes: 4
    });
    expect(directUploads.uploads).toEqual([{
      url: 'https://uploads.example.test/object-one',
      fileUri: 'file:///created.jpg',
      fileName: 'created.jpg',
      contentType: 'image/jpeg'
    }]);
    expect(client.completedDirectUploadInput).toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      assetId: 'asset-created',
      uploadId: 'upload-one'
    });
    expect(client.createdAttachmentInput).toBeUndefined();
  });

  it('allows private-network HTTP direct upload targets for local Garage development', async () => {
    const client = new FakeInventoryApiClient();
    client.directUploadURL = 'http://192.168.2.52:3900/stuffstash/object-one';
    const directUploads = new FakeDirectUploadTransport();
    const repository = new ApiInventorySummaryRepository(
      client,
      'tenant-home',
      directUploads,
      'test-scope',
      { allowLocalDevelopmentTargets: true }
    );

    await repository.addAssetPhoto(assetId('asset-created'), {
      fileName: 'created.jpg',
      contentType: 'image/jpeg',
      uri: 'file:///created.jpg',
      sizeBytes: 4
    });

    expect(directUploads.uploads).toEqual([{
      url: 'http://192.168.2.52:3900/stuffstash/object-one',
      fileUri: 'file:///created.jpg',
      fileName: 'created.jpg',
      contentType: 'image/jpeg'
    }]);
    expect(client.completedDirectUploadInput).toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      assetId: 'asset-created',
      uploadId: 'upload-one'
    });
    expect(client.createdAttachmentInput).toBeUndefined();
  });

  it('deletes asset photos through the generated client wrapper', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await repository.deleteAssetPhoto(assetId('asset-filters'), 'attachment-filters-photo');

    expect(client.deletedAttachmentInput).toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      assetId: 'asset-filters',
      attachmentId: 'attachment-filters-photo'
    });
  });

  it('falls back to JSON attachment upload for local-only direct upload targets', async () => {
    const client = new FakeInventoryApiClient();
    client.directUploadURL = 'stuffstash-local://direct-uploads/upload-one';
    const directUploads = new FakeDirectUploadTransport(false);
    const repository = new ApiInventorySummaryRepository(
      client,
      'tenant-home',
      directUploads,
      'test-scope',
      { allowLocalDevelopmentTargets: true }
    );

    await repository.addAssetPhoto(assetId('asset-created'), {
      fileName: 'created.jpg',
      contentType: 'image/jpeg',
      contentBase64: 'ZmFrZQ==',
      uri: 'file:///created.jpg',
      sizeBytes: 4
    });

    expect(client.completedDirectUploadInput).toBeUndefined();
    expect(client.createdAttachmentInput).toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      assetId: 'asset-created',
      fileName: 'created.jpg'
    });
  });

  it('rejects local-only direct upload targets when local development targets are not enabled', async () => {
    const client = new FakeInventoryApiClient();
    client.directUploadURL = 'stuffstash-local://direct-uploads/upload-one';
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home', new FakeDirectUploadTransport(false));

    await expect(
      repository.addAssetPhoto(assetId('asset-created'), {
        fileName: 'created.jpg',
        contentType: 'image/jpeg',
        contentBase64: 'ZmFrZQ==',
        uri: 'file:///created.jpg',
        sizeBytes: 4
      })
    ).rejects.toThrow('Unsupported direct attachment upload target.');

    expect(client.createdAttachmentInput).toBeUndefined();
    expect(client.completedDirectUploadInput).toBeUndefined();
  });

  it('rejects unexpected direct upload target schemes instead of silently falling back', async () => {
    const client = new FakeInventoryApiClient();
    client.directUploadURL = 'ftp://uploads.example.test/object-one';
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home', new FakeDirectUploadTransport());

    await expect(
      repository.addAssetPhoto(assetId('asset-created'), {
        fileName: 'created.jpg',
        contentType: 'image/jpeg',
        contentBase64: 'ZmFrZQ==',
        uri: 'file:///created.jpg',
        sizeBytes: 4
      })
    ).rejects.toThrow('Unsupported direct attachment upload target.');

    expect(client.createdAttachmentInput).toBeUndefined();
    expect(client.completedDirectUploadInput).toBeUndefined();
  });

  it('rejects public cleartext direct upload targets', async () => {
    const client = new FakeInventoryApiClient();
    client.directUploadURL = 'http://uploads.example.test/object-one';
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home', new FakeDirectUploadTransport());

    await expect(
      repository.addAssetPhoto(assetId('asset-created'), {
        fileName: 'created.jpg',
        contentType: 'image/jpeg',
        contentBase64: 'ZmFrZQ==',
        uri: 'file:///created.jpg',
        sizeBytes: 4
      })
    ).rejects.toThrow('Unsupported direct attachment upload target.');

    expect(client.createdAttachmentInput).toBeUndefined();
    expect(client.completedDirectUploadInput).toBeUndefined();
  });

  it('updates asset lifecycle through the generated client wrapper', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await repository.archiveAsset(assetId('asset-filters'));
    await repository.restoreAsset(assetId('asset-filters'));
    await repository.deleteAsset(assetId('asset-filters'));

    expect(client.lifecycleInputs).toEqual([
      {
        action: 'archive',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        assetId: 'asset-filters'
      },
      {
        action: 'restore',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        assetId: 'asset-filters'
      },
      {
        action: 'delete',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        assetId: 'asset-filters'
      }
    ]);
  });

  it('checks out and returns assets through the generated client wrapper', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await repository.checkoutAsset(assetId('asset-filters'), { details: 'using at desk' });
    await repository.returnAsset(assetId('asset-filters'));

    expect(client.checkoutInputs).toEqual([
      {
        action: 'checkout',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        assetId: 'asset-filters',
        details: 'using at desk'
      },
      {
        action: 'return',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        assetId: 'asset-filters',
        details: undefined
      }
    ]);
  });

  it('continues paged tenant search until selected-inventory asset results are found', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.searchAssets('paged')).resolves.toMatchObject([
      {
        id: 'asset-filters',
        title: 'Furnace filters'
      }
    ]);
  });

  it('does not stop selected-inventory search at five tenant search pages', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.searchAssets('sixth-page')).resolves.toMatchObject([
      {
        id: 'asset-filters',
        title: 'Furnace filters'
      }
    ]);
  });

  it('pages active inventory map assets to completion instead of using only the recent summary page', async () => {
    const client = new FakeInventoryApiClient();
    client.assets = Array.from({ length: 102 }, (_, index): Asset => ({
      id: index === 101 ? 'asset-final-child' : `asset-map-${index.toString().padStart(3, '0')}`,
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      kind: index === 0 ? 'location' : 'item',
      title: index === 101 ? 'Final child' : `Map asset ${index.toString().padStart(3, '0')}`,
      description: '',
      parentAssetId: index === 101 ? 'asset-map-000' : null,
      lifecycleState: index === 50 ? 'archived' : 'active',
      customFields: {},
      createdAt: '2026-06-20T10:00:00Z',
      updatedAt: `2026-06-20T10:${index.toString().padStart(2, '0')}:00Z`,
      ...(index > 0 ? {
        primaryPhoto: {
          id: `attachment-map-${index.toString()}`,
          fileName: `map-${index.toString()}.jpg`,
          contentType: 'image/jpeg',
          sizeBytes: 1024,
          thumbnails: {
            small: 'small',
            medium: 'medium',
            large: 'large'
          }
        }
      } : {})
    }));
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home', undefined, 'scope-map-test');

    const mapAssets = await repository.listActiveInventoryMapAssets();

    expect(mapAssets).toMatchObject({
      sessionScopeId: 'scope-map-test',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      inventoryName: 'Home Inventory'
    });
    expect(mapAssets.assets).toHaveLength(101);
    expect(mapAssets.assets.find((asset) => asset.id === 'asset-map-000')).toMatchObject({
      id: 'asset-map-000',
      title: 'Map asset 000'
    });
    expect(mapAssets.assets.find((asset) => asset.id === 'asset-final-child')).toMatchObject({
      id: 'asset-final-child',
      title: 'Final child',
      parentAssetId: 'asset-map-000',
      locationTrail: ['Home Inventory', 'Map asset 000', 'Final child']
    });
    expect(client.thumbnailRequests).toHaveLength(100);
    expect(new Set(client.thumbnailRequests.map((request) => request.variant))).toEqual(new Set(['small']));
    const mapAssetRequests = client.listAssetRequests.filter((request) => request.lifecycleState === 'active');
    expect(mapAssetRequests).toEqual([
      {
        inventoryId: 'inventory-home',
        limit: 100,
        cursor: undefined,
        lifecycleState: 'active',
        sort: 'id_asc'
      },
      {
        inventoryId: 'inventory-home',
        limit: 100,
        cursor: '100',
        lifecycleState: 'active',
        sort: 'id_asc'
      }
    ]);
  });

  it('keeps inventory map structure available when one row thumbnail fails', async () => {
    const client = new FakeInventoryApiClient();
    client.assets = [
      {
        id: 'asset-garage',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'location',
        title: 'Garage',
        description: '',
        parentAssetId: null,
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-20T10:00:00Z',
        updatedAt: '2026-06-20T10:00:00Z',
        primaryPhoto: {
          id: 'attachment-garage',
          fileName: 'garage.jpg',
          contentType: 'image/jpeg',
          sizeBytes: 1024,
          thumbnails: { small: 'small', medium: 'medium', large: 'large' }
        }
      },
      {
        id: 'asset-bin',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'container',
        title: 'Camping bin',
        description: '',
        parentAssetId: 'asset-garage',
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-20T10:00:00Z',
        updatedAt: '2026-06-20T10:00:00Z',
        primaryPhoto: {
          id: 'attachment-bin',
          fileName: 'bin.jpg',
          contentType: 'image/jpeg',
          sizeBytes: 1024,
          thumbnails: { small: 'small', medium: 'medium', large: 'large' }
        }
      }
    ];
    client.failedThumbnailAssetIds.add('asset-garage');
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    const mapAssets = await repository.listActiveInventoryMapAssets();

    expect(mapAssets.assets).toHaveLength(2);
    expect(mapAssets.assets.find((asset) => asset.id === 'asset-garage')).toMatchObject({
      id: 'asset-garage',
      hasPhoto: false,
      photo: undefined
    });
    expect(mapAssets.assets.find((asset) => asset.id === 'asset-bin')).toMatchObject({
      id: 'asset-bin',
      hasPhoto: true,
      photo: {
        uri: 'https://api.example.test/tenants/tenant-home/inventories/inventory-home/assets/asset-bin/attachments/attachment-bin/thumbnail?variant=small'
      }
    });
  });
});

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

function sortAssetsByUpdatedDesc(assets: readonly Asset[]): readonly Asset[] {
  return [...assets].sort((left, right) => {
    const rightTime = Date.parse(right.updatedAt || right.createdAt || '');
    const leftTime = Date.parse(left.updatedAt || left.createdAt || '');
    const timeComparison = safeTimestamp(rightTime) - safeTimestamp(leftTime);

    if (timeComparison !== 0) {
      return timeComparison;
    }

    return right.id.localeCompare(left.id);
  });
}

function safeTimestamp(timestamp: number): number {
  return Number.isNaN(timestamp) ? 0 : timestamp;
}

function lifecycleAsset(
  assets: readonly Asset[],
  assetIdValue: string,
  lifecycleState: Asset['lifecycleState']
): Asset {
  return {
    ...assetById(assets, assetIdValue),
    lifecycleState
  };
}

function assetById(assets: readonly Asset[], assetIdValue: string): Asset {
  const asset = assets.find((candidate) => candidate.id === assetIdValue);

  if (!asset) {
    throw new Error('Asset not found.');
  }

  return asset;
}

function checkoutRecord(
  assetIdValue: string,
  state: AssetCheckout['state'],
  details?: string
): AssetCheckout {
  return {
    id: 'checkout-fake',
    tenantId: 'tenant-home',
    inventoryId: 'inventory-home',
    assetId: assetIdValue,
    state,
    checkoutDetails: details,
    checkedOutAt: '2026-06-24T10:00:00Z',
    checkedOutByPrincipalId: 'principal-mobile',
    returnedAt: state === 'returned' ? '2026-06-24T10:05:00Z' : undefined,
    returnedByPrincipalId: state === 'returned' ? 'principal-mobile' : undefined,
    createdAt: '2026-06-24T10:00:00Z',
    updatedAt: '2026-06-24T10:05:00Z'
  };
}
