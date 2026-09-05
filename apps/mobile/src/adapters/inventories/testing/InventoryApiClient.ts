import type {
  Asset,
  AssetCheckout,
  AssetPhotoReference,
  AssetSearchResult,
  AssetTag,
  Attachment,
  CheckedOutAsset,
  DirectUpload,
  Inventory,
  Page,
  Tenant
} from '@stuff-stash/api-client';

export class FakeInventoryApiClient {
  readonly requestKinds: string[] = [];
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
  additionalHomeInventory: Inventory | undefined;
  listInventoryRequests: string[] = [];
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
  getAssetRequests: Array<{ readonly inventoryId: string; readonly assetId: string }> = [];
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
    readonly action: 'checkout' | 'return' | 'return_details';
    readonly tenantId: string;
    readonly inventoryId: string;
    readonly assetId: string;
    readonly checkoutId?: string;
    readonly details?: string;
  }> = [];
  undoInputs: Array<{
    readonly tenantId: string;
    readonly inventoryId: string;
    readonly operationId: string;
    readonly direction: 'undo' | 'redo';
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
  searchResultAssetOverrides = new Map<string, Partial<Asset>>();

  async listMyTenants(): Promise<Page<Tenant>> {
    this.requestKinds.push('list_tenants');
    return page([this.tenant, this.cabinTenant]);
  }

  async listInventories(tenantId: string): Promise<Page<Inventory>> {
    this.requestKinds.push('list_inventories');
    this.listInventoryRequests.push(tenantId);
    if (tenantId === this.cabinTenant.id) {
      return page([this.cabinInventory]);
    }

    return page([
      this.inventory,
      ...(this.additionalHomeInventory ? [this.additionalHomeInventory] : [])
    ]);
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

  async getAsset(_tenantId: string, inventoryId: string, selectedAssetId: string): Promise<Asset> {
    this.getAssetRequests.push({ inventoryId, assetId: selectedAssetId });
    const asset = this.assets.find((candidate) => candidate.id === selectedAssetId && candidate.inventoryId === inventoryId);
    if (!asset) {
      throw new Error('Asset not found.');
    }
    return asset;
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
    this.requestKinds.push('return_asset');
    this.checkoutInputs.push({
      action: 'return',
      tenantId,
      inventoryId,
      assetId: assetIdValue,
      details: input.details
    });
    return checkoutRecord(assetIdValue, 'returned', input.details);
  }

  async updateReturnedCheckoutDetails(
    tenantId: string,
    inventoryId: string,
    assetIdValue: string,
    checkoutId: string,
    input: { readonly details?: string } = {}
  ): Promise<AssetCheckout> {
    this.checkoutInputs.push({
      action: 'return_details',
      tenantId,
      inventoryId,
      assetId: assetIdValue,
      checkoutId,
      details: input.details
    });
    return checkoutRecord(assetIdValue, 'returned', input.details);
  }

  async applyUndoableOperation(
    tenantId: string,
    inventoryId: string,
    operationId: string,
    direction: 'undo' | 'redo'
  ): Promise<Asset> {
    this.undoInputs.push({ tenantId, inventoryId, operationId, direction });
    return this.assets.find((asset) => asset.id === 'asset-filters') ?? this.assets[0]!;
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
    const baseAsset = this.assets[1];

    if (!baseAsset) {
      return page([]);
    }
    const asset = {
      ...baseAsset,
      ...this.searchResultAssetOverrides.get(baseAsset.id)
    };

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

export class FakeDirectUploadTransport {
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

export function page<T>(items: readonly T[]): Page<T> {
  return pageWithCursor(items, null);
}

export function pageWithCursor<T>(items: readonly T[], nextCursor: string | null): Page<T> {
  return {
    items: [...items],
    pagination: {
      limit: items.length,
      nextCursor,
      hasMore: nextCursor !== null
    }
  };
}

export function sortAssetsByUpdatedDesc(assets: readonly Asset[]): readonly Asset[] {
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

export function safeTimestamp(timestamp: number): number {
  return Number.isNaN(timestamp) ? 0 : timestamp;
}

export function lifecycleAsset(
  assets: readonly Asset[],
  assetIdValue: string,
  lifecycleState: Asset['lifecycleState']
): Asset {
  return {
    ...assetById(assets, assetIdValue),
    lifecycleState
  };
}

export function assetById(assets: readonly Asset[], assetIdValue: string): Asset {
  const asset = assets.find((candidate) => candidate.id === assetIdValue);

  if (!asset) {
    throw new Error('Asset not found.');
  }

  return asset;
}

export function checkoutRecord(
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
    undoableOperationId: 'operation-checkout-fake',
    createdAt: '2026-06-24T10:00:00Z',
    updatedAt: '2026-06-24T10:05:00Z'
  };
}
