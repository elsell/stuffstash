import type {
  Asset,
  AssetTag,
  AssetPhotoReference,
  AssetSearchResult,
  Inventory,
  StuffStashClient,
  Tenant
} from '@stuff-stash/api-client';
import {
  AssetBrowsePage,
  AssetBrowsePageInput,
  CreateInventoryAssetInput,
  CreateInventoryAssetTagInput,
  CreateInventoryAssetPhotoInput,
  AddInventoryAssetPhotoInput,
  GetInventoryAssetDetailInput,
  InventoryAssetPhotoDirectUpload,
  InventorySummaryRepository,
  InventoryWorkspace,
  UpdateInventoryAssetInput
} from '../../application/home/InventorySummaryRepository';
import type { InventoryMapAssetRepository } from '../../application/assets/InventoryMapQuery';
import { assetId, AssetSummary, type AssetTagSummary } from '../../domain/assets/AssetSummary';
import {
  AccessRole,
  InventoryId,
  inventoryId,
  InventorySummary,
  tenantId
} from '../../domain/inventories/InventorySummary';
import type { LocationSummary } from '../../domain/locations/LocationSummary';
import {
  directUploadMethod,
  isDirectUploadHTTPTransportAllowed,
  isDirectUploadTargetSupported,
  isLocalDirectUploadURL,
  type DirectUploadTargetPolicy
} from '../uploads/DirectUploadPolicy';

type InventoryApiClient = Pick<
  StuffStashClient,
  | 'listMyTenants'
  | 'listInventories'
  | 'listAssets'
  | 'listAssetTags'
  | 'createAsset'
  | 'createAssetTag'
  | 'updateAsset'
  | 'checkoutAsset'
  | 'returnAsset'
  | 'archiveAsset'
  | 'restoreAsset'
  | 'deleteAsset'
  | 'createAssetAttachment'
  | 'initiateAssetAttachmentDirectUpload'
  | 'completeAssetAttachmentDirectUpload'
  | 'deleteAssetAttachment'
  | 'searchAssets'
  | 'listCheckedOutAssets'
  | 'listAssetAttachments'
  | 'assetAttachmentThumbnailReference'
>;

const inventoryAssetPageSize = 100;

type DirectUploadTransport = {
  upload(input: DirectUploadTransportInput): Promise<boolean>;
};

type DirectUploadTransportInput = {
  readonly upload: InventoryAssetPhotoDirectUpload;
  readonly fileUri: string;
  readonly fileName: string;
  readonly contentType: CreateInventoryAssetPhotoInput['contentType'];
};

export class ApiInventorySummaryRepository implements InventorySummaryRepository, InventoryMapAssetRepository {
  private selectedInventoryId: InventoryId | undefined;
  private readonly directUploadTransport: DirectUploadTransport;

  constructor(
    private readonly client: InventoryApiClient,
    private readonly configuredTenantId: string,
    directUploadTransport?: DirectUploadTransport,
    private readonly sessionScopeId = 'mobile-composition',
    private readonly directUploadPolicy: DirectUploadTargetPolicy = {}
  ) {
    this.directUploadTransport = directUploadTransport ?? new ExpoDirectUploadTransport(directUploadPolicy);
  }

  async getInventoryWorkspace(): Promise<InventoryWorkspace> {
    const tenantsPage = await this.client.listMyTenants(100);
    const tenants = tenantsPage.items;
    const inventoriesByTenant = await Promise.all(
      tenants.map(async (tenant) => {
        const inventoriesPage = await this.client.listInventories(tenant.id, 100);
        return inventoriesPage.items.map((inventory) => ({ tenant, inventory }));
      })
    );
    const availableInventories = inventoriesByTenant.flat();
    const preferredInventory =
      availableInventories.find((item) => item.tenant.id === this.configuredTenantId) ??
      availableInventories[0];
    const defaultInventory =
      availableInventories.find((item) => inventoryId(item.inventory.id) === this.selectedInventoryId) ??
      preferredInventory;

    if (!defaultInventory) {
      throw new Error('API principal did not include any inventories.');
    }

    const inventories = await Promise.all(
      availableInventories.map((item) =>
        this.mapInventoryWithAssets(
          item.tenant,
          item.inventory,
          item.tenant.id === defaultInventory.tenant.id && item.inventory.id === defaultInventory.inventory.id
        )
      )
    );

    return {
      tenants: tenants.map(mapTenant),
      inventories,
      defaultInventoryId: inventoryId(defaultInventory.inventory.id)
    };
  }

  async getDefaultInventorySummary(): Promise<InventorySummary> {
    const workspace = await this.getInventoryWorkspace();
    const inventory =
      workspace.inventories.find((item) => item.id === workspace.defaultInventoryId) ??
      workspace.inventories[0];

    if (!inventory) {
      throw new Error('API workspace did not include an inventory.');
    }

    return inventory;
  }

  async getAssetDetail(input: GetInventoryAssetDetailInput): Promise<AssetSummary> {
    const asset = summaryToApiAsset(input.tenantId, input.inventoryId, input.asset);
    const photos = await this.photosForAsset(asset, { allowAttachmentListFailure: false });

    return {
      ...input.asset,
      hasPhoto: photos.length > 0,
      photos,
      photo: photos[0]
    };
  }

  async selectInventory(selectedInventoryId: InventoryId): Promise<void> {
    const workspace = await this.getInventoryWorkspace();
    const inventory = workspace.inventories.find((item) => item.id === selectedInventoryId);

    if (!inventory) {
      throw new Error('Selected inventory is not available in the configured tenant.');
    }

    this.selectedInventoryId = inventory.id;
  }

  async createAsset(input: CreateInventoryAssetInput): Promise<AssetSummary> {
    const inventory = await this.getDefaultInventorySummary();
    const asset = await this.client.createAsset(inventory.tenantId, inventory.id, {
      kind: input.kind,
      title: input.title,
      description: input.description,
      parentAssetId: input.parentAssetId,
      ...(input.tagIds !== undefined ? { tagIds: [...input.tagIds] } : {})
    });

    return this.mapAssetWithPhoto(
      inventory.name,
      asset,
      inventory.assets.map((item) => summaryToApiAsset(inventory.tenantId, inventory.id, item))
    );
  }

  async createAssetTag(input: CreateInventoryAssetTagInput): Promise<AssetTagSummary> {
    const inventory = await this.getDefaultInventorySummary();
    const tag = await this.client.createAssetTag(inventory.tenantId, inventory.id, {
      displayName: input.displayName,
      ...(input.color !== undefined ? { color: input.color } : {})
    });
    return mapAssetTag(tag);
  }

  async updateAsset(input: UpdateInventoryAssetInput): Promise<AssetSummary> {
    const inventory = await this.getDefaultInventorySummary();
    const asset = await this.client.updateAsset(inventory.tenantId, inventory.id, input.assetId, {
      ...(input.title !== undefined ? { title: input.title } : {}),
      ...(input.description !== undefined ? { description: input.description } : {}),
      ...(input.parentAssetId !== undefined ? { parentAssetId: input.parentAssetId } : {}),
      ...(input.tagIds !== undefined ? { tagIds: [...input.tagIds] } : {})
    });

    const knownAssets = inventory.assets.map((item) =>
      summaryToApiAsset(inventory.tenantId, inventory.id, item)
    );
    return this.mapAssetWithPhoto(inventory.name, asset, knownAssets);
  }

  async addAssetPhoto(
    assetIdValue: AssetSummary['id'],
    input: CreateInventoryAssetPhotoInput
  ): Promise<void> {
    const inventory = await this.getDefaultInventorySummary();
    await this.addInventoryAssetPhoto({
      tenantId: inventory.tenantId,
      inventoryId: inventory.id,
      assetId: assetIdValue,
      ...input
    });
  }

  async addInventoryAssetPhoto(input: AddInventoryAssetPhotoInput): Promise<void> {
    if (input.uri && input.sizeBytes && input.sizeBytes > 0) {
      const directUpload = input.directUpload ?? await this.client.initiateAssetAttachmentDirectUpload(input.tenantId, input.inventoryId, input.assetId, {
        fileName: input.fileName,
        contentType: input.contentType,
        sizeBytes: input.sizeBytes
      });
      if (!isDirectUploadTargetSupported(directUpload.url, this.directUploadPolicy)) {
        throw new Error('Unsupported direct attachment upload target.');
      }
      const uploaded = await this.directUploadTransport.upload({
        upload: directUpload,
        fileUri: input.uri,
        fileName: input.fileName,
        contentType: input.contentType
      });
      if (uploaded) {
        await this.client.completeAssetAttachmentDirectUpload(input.tenantId, input.inventoryId, input.assetId, directUpload.uploadId);
        return;
      }
    }

    await this.client.createAssetAttachment(input.tenantId, input.inventoryId, input.assetId, {
      fileName: input.fileName,
      contentType: input.contentType,
      contentBase64: await attachmentContentBase64(input)
    });
  }

  async deleteAssetPhoto(assetIdValue: AssetSummary['id'], photoId: string): Promise<void> {
    const inventory = await this.getDefaultInventorySummary();
    await this.client.deleteAssetAttachment(inventory.tenantId, inventory.id, assetIdValue, photoId);
  }

  async archiveAsset(assetIdValue: AssetSummary['id']): Promise<void> {
    const inventory = await this.getDefaultInventorySummary();
    await this.client.archiveAsset(inventory.tenantId, inventory.id, assetIdValue);
  }

  async restoreAsset(assetIdValue: AssetSummary['id']): Promise<void> {
    const inventory = await this.getDefaultInventorySummary();
    await this.client.restoreAsset(inventory.tenantId, inventory.id, assetIdValue);
  }

  async deleteAsset(assetIdValue: AssetSummary['id']): Promise<void> {
    const inventory = await this.getDefaultInventorySummary();
    await this.client.deleteAsset(inventory.tenantId, inventory.id, assetIdValue);
  }

  async checkoutAsset(assetIdValue: AssetSummary['id'], input: { readonly details?: string } = {}): Promise<void> {
    const inventory = await this.getDefaultInventorySummary();
    await this.client.checkoutAsset(inventory.tenantId, inventory.id, assetIdValue, input);
  }

  async returnAsset(assetIdValue: AssetSummary['id'], input: { readonly details?: string } = {}): Promise<void> {
    const inventory = await this.getDefaultInventorySummary();
    await this.client.returnAsset(inventory.tenantId, inventory.id, assetIdValue, input);
  }

  async browseAssets(input: AssetBrowsePageInput): Promise<AssetBrowsePage> {
    const workspace = await this.getInventoryWorkspace();
    const inventory =
      workspace.inventories.find((item) => item.id === workspace.defaultInventoryId) ??
      workspace.inventories[0];

    if (!inventory) {
      throw new Error('API workspace did not include an inventory.');
    }

    const knownAssets = inventory.assets.map((item) =>
      summaryToApiAsset(inventory.tenantId, inventory.id, item)
    );
    const hasTagFilters = (input.tagIds?.length ?? 0) > 0;
    if (!hasTagFilters && input.query.trim().length === 0 && input.checkoutState === 'checked_out') {
      return await this.listCheckedOutInventoryAssetPage(inventory, input, knownAssets);
    }
    return input.query.trim().length > 0 || hasTagFilters
      ? await this.searchInventoryAssetPage(inventory, input, knownAssets)
      : await this.listInventoryAssetPage(inventory, input, knownAssets);
  }

  async listActiveInventoryMapAssets(): Promise<{
    readonly sessionScopeId: string;
    readonly tenantId: InventorySummary['tenantId'];
    readonly inventoryId: InventorySummary['id'];
    readonly inventoryName: string;
    readonly permissions: readonly string[];
    readonly assets: readonly AssetSummary[];
  }> {
    const inventory = await this.getDefaultInventoryForMap();
    const activeAssets = await this.listAllActiveInventoryAssets(tenantId(inventory.tenant.id), inventory.inventory.id);
    const assets = await this.mapAssetsWithMapPhotos(inventory.inventory.name, activeAssets);

    return {
      sessionScopeId: this.sessionScopeId,
      tenantId: tenantId(inventory.tenant.id),
      inventoryId: inventoryId(inventory.inventory.id),
      inventoryName: inventory.inventory.name,
      permissions: inventory.inventory.access.permissions,
      assets
    };
  }

  async searchAssets(query: string): Promise<readonly AssetSummary[]> {
    const workspace = await this.getInventoryWorkspace();
    const inventory =
      workspace.inventories.find((item) => item.id === workspace.defaultInventoryId) ??
      workspace.inventories[0];

    if (!inventory) {
      throw new Error('API workspace did not include an inventory.');
    }

    const page = await this.searchSelectedInventoryAssets(inventory.tenantId, inventory.id, query, 50);
    const siblings = inventory.assets.map((item) =>
      summaryToApiAsset(inventory.tenantId, inventory.id, item)
    );
    return Promise.all(
      page.map((item) =>
        this.mapAssetWithPrimaryPhoto(inventory.name, item.asset, siblings)
      )
    );
  }

  async searchLocations(query: string): Promise<readonly LocationSummary[]> {
    const workspace = await this.getInventoryWorkspace();
    const inventory =
      workspace.inventories.find((item) => item.id === workspace.defaultInventoryId) ??
      workspace.inventories[0];

    if (!inventory) {
      throw new Error('API workspace did not include an inventory.');
    }

    const page = await this.client.searchAssets(inventory.tenantId, query, { limit: 50 });
    const locationAssets = page.items
      .filter((item) => item.inventory.id === inventory.id && item.asset.kind === 'location')
      .map((item) => item.asset);
    const knownAssets = inventory.assets.map((item) =>
      summaryToApiAsset(inventory.tenantId, inventory.id, item)
    );

    return Promise.all(
      locationAssets.map(async (location) =>
        mapLocation(location, knownAssets, await this.primaryPhotoForAsset(location))
      )
    );
  }

  private async mapInventoryWithAssets(
    tenant: Tenant,
    inventory: Inventory,
    hydrateFullLocations: boolean
  ): Promise<InventorySummary> {
    const assets = await this.listRecentInventoryAssets(tenant.id, inventory.id);
    const locationSourceAssets = hydrateFullLocations
      ? await this.listAllActiveInventoryAssets(tenant.id, inventory.id)
      : assets;
    const assetTags = await this.listAllInventoryTags(tenant.id, inventory.id);
    const locations = await Promise.all(
      locationSourceAssets
        .filter((asset) => asset.kind === 'location')
        .map(async (location) => mapLocation(location, locationSourceAssets, await this.primaryPhotoForAsset(location)))
    );

    const mappedAssets = await Promise.all(
      assets.map((asset) => this.mapAssetWithPrimaryPhoto(inventory.name, asset, assets))
    );

    return {
      id: inventoryId(inventory.id),
      tenantId: tenantId(tenant.id),
      name: inventory.name,
      role: mapAccessRole(inventory.access.relationship),
      permissions: [...inventory.access.permissions],
      description: '',
      updatedAtLabel: 'Loaded from API',
      locationCount: locations.length,
      locations,
      assets: mappedAssets,
      assetTags: assetTags.map(mapAssetTag)
    };
  }

  private async listRecentInventoryAssets(
    tenantID: string,
    inventoryID: string
  ): Promise<readonly Asset[]> {
    const page = await this.client.listAssets(
      tenantID,
      inventoryID,
      inventoryAssetPageSize,
      undefined,
      'all',
      'updated_desc'
    );
    return page.items;
  }

  private async listAllInventoryTags(
    tenantID: string,
    inventoryID: string
  ): Promise<readonly AssetTag[]> {
    const tags: AssetTag[] = [];
    let cursor: string | undefined;

    do {
      const page = await this.client.listAssetTags(tenantID, inventoryID, 100, cursor);
      tags.push(...page.items);
      cursor = page.pagination.nextCursor ?? undefined;
    } while (cursor);

    return tags;
  }

  private async getDefaultInventoryForMap(): Promise<{
    readonly tenant: Tenant;
    readonly inventory: Inventory;
  }> {
    const tenantsPage = await this.client.listMyTenants(100);
    const tenants = tenantsPage.items;
    const inventoriesByTenant = await Promise.all(
      tenants.map(async (tenant) => {
        const inventoriesPage = await this.client.listInventories(tenant.id, 100);
        return inventoriesPage.items.map((inventory) => ({ tenant, inventory }));
      })
    );
    const inventories = inventoriesByTenant.flat();
    const preferredInventory =
      inventories.find((item) => item.tenant.id === this.configuredTenantId) ??
      inventories[0];
    const defaultInventory =
      inventories.find((item) => inventoryId(item.inventory.id) === this.selectedInventoryId) ??
      preferredInventory;

    if (!defaultInventory) {
      throw new Error('API principal did not include any inventories.');
    }

    return defaultInventory;
  }

  private async listAllActiveInventoryAssets(
    tenantID: string,
    inventoryID: string
  ): Promise<readonly Asset[]> {
    const assets: Asset[] = [];
    let cursor: string | undefined;

    do {
      const page = await this.client.listAssets(
        tenantID,
        inventoryID,
        inventoryAssetPageSize,
        cursor,
        'active',
        'id_asc'
      );
      assets.push(...page.items);
      cursor = page.pagination.nextCursor ?? undefined;
    } while (cursor);

    return assets;
  }

  private async mapAssetWithPhoto(
    inventoryName: string,
    asset: Asset,
    assets: readonly Asset[]
  ): Promise<AssetSummary> {
    const photos = await this.photosForAsset(asset);
    return mapAsset(inventoryName, asset, assets, photos);
  }

  private async mapAssetWithPrimaryPhoto(
    inventoryName: string,
    asset: Asset,
    assets: readonly Asset[]
  ): Promise<AssetSummary> {
    const photo = await this.primaryPhotoForAsset(asset);
    return mapAsset(inventoryName, asset, assets, photo ? [photo] : []);
  }

  private async mapAssetsWithMapPhotos(
    inventoryName: string,
    assets: readonly Asset[]
  ): Promise<readonly AssetSummary[]> {
    return mapWithConcurrency(assets, 6, async (asset) => {
      const photo = await this.primaryMapPhotoForAsset(asset);
      return mapAsset(inventoryName, asset, assets, photo ? [photo] : []);
    });
  }

  private async primaryMapPhotoForAsset(asset: Asset): Promise<NonNullable<AssetSummary['photo']> | undefined> {
    if (!asset.primaryPhoto) {
      return undefined;
    }
    let smallReference: AssetPhotoReference;
    try {
      smallReference = await this.client.assetAttachmentThumbnailReference(
        asset.tenantId,
        asset.inventoryId,
        asset.id,
        asset.primaryPhoto.id,
        'small'
      );
    } catch {
      return undefined;
    }
    return {
      id: asset.primaryPhoto.id,
      fileName: asset.primaryPhoto.fileName,
      contentType: asset.primaryPhoto.contentType,
      sizeBytes: asset.primaryPhoto.sizeBytes,
      uri: smallReference.uri,
      headers: smallReference.headers
    };
  }

  private async primaryPhotoForAsset(asset: Asset): Promise<NonNullable<AssetSummary['photo']> | undefined> {
    if (!asset.primaryPhoto) {
      return undefined;
    }
    const smallReference = await this.client.assetAttachmentThumbnailReference(
      asset.tenantId,
      asset.inventoryId,
      asset.id,
      asset.primaryPhoto.id,
      'small'
    );
    const mediumReference = await this.client.assetAttachmentThumbnailReference(
      asset.tenantId,
      asset.inventoryId,
      asset.id,
      asset.primaryPhoto.id,
      'medium'
    );
    const largeReference = await this.client.assetAttachmentThumbnailReference(
      asset.tenantId,
      asset.inventoryId,
      asset.id,
      asset.primaryPhoto.id,
      'large'
    );
    return {
      id: asset.primaryPhoto.id,
      fileName: asset.primaryPhoto.fileName,
      contentType: asset.primaryPhoto.contentType,
      sizeBytes: asset.primaryPhoto.sizeBytes,
      uri: smallReference.uri,
      heroUri: mediumReference.uri,
      heroHeaders: mediumReference.headers,
      viewerUri: largeReference.uri,
      viewerHeaders: largeReference.headers,
      headers: smallReference.headers
    };
  }

  private async photosForAsset(
    asset: Asset,
    options: { readonly allowAttachmentListFailure?: boolean } = {}
  ): Promise<readonly NonNullable<AssetSummary['photo']>[]> {
    const attachments = [];
    let cursor: string | undefined;

    try {
      do {
        const page = await this.client.listAssetAttachments(
          asset.tenantId,
          asset.inventoryId,
          asset.id,
          50,
          cursor
        );
        attachments.push(...page.items);
        cursor = page.pagination.nextCursor ?? undefined;
      } while (cursor);
    } catch {
      if (options.allowAttachmentListFailure === false) {
        throw new Error('Asset attachments could not be loaded.');
      }
      return [];
    }

    return mapWithConcurrency(
      attachments.filter((item) => item.lifecycleState === 'active' && item.contentType.startsWith('image/')),
      4,
      async (attachment) => {
        const smallReference = await this.client.assetAttachmentThumbnailReference(
          asset.tenantId,
          asset.inventoryId,
          asset.id,
          attachment.id,
          'small'
        );
        const mediumReference = await this.client.assetAttachmentThumbnailReference(
          asset.tenantId,
          asset.inventoryId,
          asset.id,
          attachment.id,
          'medium'
        );
        const largeReference = await this.client.assetAttachmentThumbnailReference(
          asset.tenantId,
          asset.inventoryId,
          asset.id,
          attachment.id,
          'large'
        );
        return {
          id: attachment.id,
          fileName: attachment.fileName,
          contentType: attachment.contentType,
          sizeBytes: attachment.sizeBytes,
          uri: smallReference.uri,
          heroUri: mediumReference.uri,
          heroHeaders: mediumReference.headers,
          viewerUri: largeReference.uri,
          viewerHeaders: largeReference.headers,
          headers: smallReference.headers
        };
      }
    );
  }

  private async searchSelectedInventoryAssets(
    tenantID: string,
    inventoryID: string,
    query: string,
    desiredMatches: number
  ) {
    const matches: AssetSearchResult[] = [];
    let cursor: string | undefined;

    do {
      const page = await this.client.searchAssets(tenantID, query, {
        limit: 50,
        cursor
      });
      matches.push(...page.items.filter((item) => item.inventory.id === inventoryID));
      cursor = page.pagination.nextCursor ?? undefined;
    } while (matches.length < desiredMatches && cursor);

    return matches.slice(0, desiredMatches);
  }

  private async listInventoryAssetPage(
    inventory: InventorySummary,
    input: AssetBrowsePageInput,
    knownAssets: readonly Asset[]
  ): Promise<AssetBrowsePage> {
    const desiredMatches = input.limit ?? 20;
    const selectedAssets: Asset[] = [];
    let cursor = input.cursor;
    let nextCursor: string | undefined;
    let hasMore = false;

    do {
      const pageSize = desiredMatches - selectedAssets.length;
      const page = await this.client.listAssets(
        inventory.tenantId,
        inventory.id,
        pageSize,
        cursor,
        input.lifecycleState,
        input.sort
      );
      selectedAssets.push(...filterAssetsByCheckoutState(
        filterAssetsByKind(page.items, input.kind),
        input.checkoutState
      ));
      nextCursor = page.pagination.nextCursor ?? undefined;
      hasMore = page.pagination.hasMore;
      cursor = nextCursor;
    } while (selectedAssets.length < desiredMatches && hasMore);

    const assets = await Promise.all(
      selectedAssets
        .slice(0, desiredMatches)
        .map((asset) => this.mapAssetWithPrimaryPhoto(inventory.name, asset, knownAssets))
    );

    return {
      assets,
      nextCursor,
      hasMore
    };
  }

  private async searchInventoryAssetPage(
    inventory: InventorySummary,
    input: AssetBrowsePageInput,
    knownAssets: readonly Asset[]
  ): Promise<AssetBrowsePage> {
    const desiredMatches = input.limit ?? 20;
    const selectedResults: AssetSearchResult[] = [];
    let cursor = input.cursor;
    let nextCursor: string | undefined;
    let hasMore = false;

    do {
      const pageSize = desiredMatches - selectedResults.length;
      const page = await this.client.searchAssets(inventory.tenantId, input.query, {
        limit: pageSize,
        cursor,
        inventoryId: inventory.id,
        tagIds: input.tagIds,
        lifecycleState: input.lifecycleState,
        checkoutState: input.checkoutState
      });
      selectedResults.push(
        ...page.items
          .filter((item) => item.inventory.id === inventory.id)
          .filter((item) => assetMatchesKind(item.asset, input.kind))
      );
      nextCursor = page.pagination.nextCursor ?? undefined;
      hasMore = page.pagination.hasMore;
      cursor = nextCursor;
    } while (selectedResults.length < desiredMatches && hasMore);

    const pageResults = selectedResults.slice(0, desiredMatches);
    const assets = await Promise.all(
      pageResults.map((item) =>
        this.mapAssetWithPrimaryPhoto(inventory.name, item.asset, knownAssets)
      )
    );
    const searchMatches = pageResults
      .map((item) => ({
        assetId: assetId(item.asset.id),
        labels: searchMatchLabels(item.matches)
      }))
      .filter((item) => item.labels.length > 0);

    return {
      assets,
      searchMatches,
      nextCursor,
      hasMore
    };
  }

  private async listCheckedOutInventoryAssetPage(
    inventory: InventorySummary,
    input: AssetBrowsePageInput,
    knownAssets: readonly Asset[]
  ): Promise<AssetBrowsePage> {
    const page = await this.client.listCheckedOutAssets(
      inventory.tenantId,
      inventory.id,
      input.limit ?? 20,
      input.cursor
    );
    const selectedAssets = page.items
      .map((item) => item.asset)
      .filter((asset) => input.lifecycleState === 'all' || asset.lifecycleState === input.lifecycleState);
    const assets = await Promise.all(
      filterAssetsByKind(selectedAssets, input.kind).map((asset) =>
        this.mapAssetWithPrimaryPhoto(inventory.name, asset, knownAssets)
      )
    );

    return {
      assets,
      nextCursor: page.pagination.nextCursor ?? undefined,
      hasMore: page.pagination.hasMore
    };
  }
}

class ExpoDirectUploadTransport implements DirectUploadTransport {
  constructor(private readonly directUploadPolicy: DirectUploadTargetPolicy = {}) {}

  async upload(input: DirectUploadTransportInput): Promise<boolean> {
    if (this.directUploadPolicy.allowLocalDevelopmentTargets === true && isLocalDirectUploadURL(input.upload.url)) {
      return false;
    }
    if (!isDirectUploadHTTPTransportAllowed(input.upload.url, this.directUploadPolicy)) {
      throw new Error('Direct attachment upload target must use HTTPS or a private local development host.');
    }
    const FileSystem = await import('expo-file-system/legacy');
    const uploadMethod = directUploadMethod(input.upload.method);
    const result = await FileSystem.uploadAsync(input.upload.url, input.fileUri, {
      httpMethod: uploadMethod,
      headers: input.upload.headers,
      ...(Object.keys(input.upload.formFields).length > 0
        ? {
            uploadType: FileSystem.FileSystemUploadType.MULTIPART,
            fieldName: 'file',
            mimeType: input.contentType,
            parameters: input.upload.formFields
          }
        : {
            uploadType: FileSystem.FileSystemUploadType.BINARY_CONTENT
          })
    });
    if (result.status < 200 || result.status >= 300) {
      throw new Error('Direct attachment upload failed.');
    }
    return true;
  }
}

async function attachmentContentBase64(input: CreateInventoryAssetPhotoInput): Promise<string> {
  if (input.contentBase64) {
    return input.contentBase64;
  }
  if (!input.uri) {
    throw new Error('Attachment content is not available for JSON upload fallback.');
  }
  const FileSystem = await import('expo-file-system/legacy');
  return FileSystem.readAsStringAsync(input.uri, { encoding: FileSystem.EncodingType.Base64 });
}

function mapAccessRole(relationship: string): AccessRole {
  switch (relationship) {
    case 'owner':
    case 'editor':
    case 'viewer':
      return relationship;
    default:
      return 'viewer';
  }
}

function mapTenant(tenant: Tenant) {
  return {
    id: tenantId(tenant.id),
    name: tenant.name
  };
}

function mapAssetTag(tag: AssetTag) {
  return {
    id: tag.id,
    key: tag.key,
    displayName: tag.displayName,
    color: tag.color
  };
}

function mapLocation(
  location: Asset,
  assets: readonly Asset[],
  photo?: AssetSummary['photo']
): LocationSummary {
  const children = assets.filter((asset) => asset.parentAssetId === location.id);
  const recentChildren = sortAssetsByUpdatedDesc(children).slice(0, 3);

  return {
    id: assetId(location.id),
    inventoryId: inventoryId(location.inventoryId),
    title: location.title,
    description: location.description || 'Location asset',
    containedAssetCount: children.length,
    recentAssetTitles: recentChildren.map((asset) => asset.title),
    hasPhoto: photo !== undefined,
    photo
  };
}

function filterAssetsByKind(
  assets: readonly Asset[],
  kind: AssetBrowsePageInput['kind']
): readonly Asset[] {
  if (kind === 'all') {
    return assets;
  }

  return assets.filter((asset) => asset.kind === kind);
}

function assetMatchesKind(asset: Asset, kind: AssetBrowsePageInput['kind']): boolean {
  return kind === 'all' || asset.kind === kind;
}

function filterAssetsByCheckoutState(
  assets: readonly Asset[],
  checkoutState: AssetBrowsePageInput['checkoutState']
): readonly Asset[] {
  if (checkoutState === 'checked_out') {
    return assets.filter((asset) => asset.currentCheckout !== undefined);
  }
  if (checkoutState === 'available') {
    return assets.filter((asset) => asset.currentCheckout === undefined);
  }
  return assets;
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

function mapAsset(
  inventoryName: string,
  asset: Asset,
  assets: readonly Asset[],
  photos: readonly NonNullable<AssetSummary['photo']>[] = []
): AssetSummary {
  const parent = asset.parentAssetId
    ? assets.find((candidate) => candidate.id === asset.parentAssetId)
    : undefined;
  const ancestorTitles = ancestorTrail(asset, assets).map((ancestor) => ancestor.title);
  const photo = photos[0];

  return {
    id: assetId(asset.id),
    title: asset.title,
    kind: asset.kind,
    lifecycleState: asset.lifecycleState,
    parentAssetId: asset.parentAssetId ? assetId(asset.parentAssetId) : undefined,
    locationLabel: parent?.title ?? 'Inventory root',
    locationTrail: [inventoryName, ...ancestorTitles, asset.title].filter(isString),
    description: asset.description,
    updatedAtLabel: updatedAtLabel(asset),
    hasPhoto: photo !== undefined,
    photos,
    photo,
    currentCheckout: asset.currentCheckout,
    tags: asset.tags
  };
}

function searchMatchLabels(matches: readonly { readonly field: string }[]): readonly string[] {
  const labels: string[] = [];
  const seen = new Set<string>();
  for (const match of matches) {
    const label = searchMatchFieldLabel(match.field);
    if (seen.has(label)) {
      continue;
    }
    labels.push(label);
    seen.add(label);
  }
  return labels;
}

function searchMatchFieldLabel(field: string): string {
  switch (field) {
    case 'tag_display_name':
    case 'tag_key':
      return 'Tag';
    case 'title':
      return 'Title';
    case 'description':
      return 'Description';
    case 'location':
    case 'path':
      return 'Location';
    case 'custom_field':
      return 'Custom field';
    default:
      return humanizeSearchMatchField(field);
  }
}

function humanizeSearchMatchField(field: string): string {
  const label = field.trim().replace(/[_-]+/g, ' ');
  if (label.length === 0) {
    return 'Match';
  }
  return label.charAt(0).toUpperCase() + label.slice(1);
}

function ancestorTrail(asset: Asset, assets: readonly Asset[]): readonly Asset[] {
  const byID = new Map(assets.map((candidate) => [candidate.id, candidate]));
  const ancestors: Asset[] = [];
  const seen = new Set<string>([asset.id]);
  let parentID = asset.parentAssetId ?? undefined;

  while (parentID && !seen.has(parentID)) {
    seen.add(parentID);
    const parent = byID.get(parentID);
    if (!parent) {
      break;
    }
    ancestors.unshift(parent);
    parentID = parent.parentAssetId ?? undefined;
  }

  return ancestors;
}

function isString(value: string | undefined): value is string {
  return typeof value === 'string' && value.length > 0;
}

function updatedAtLabel(asset: Asset): string {
  const timestamp = asset.updatedAt || asset.createdAt;
  if (!timestamp) {
    return 'Loaded from API';
  }
  const date = new Date(timestamp);
  if (Number.isNaN(date.getTime())) {
    return 'Loaded from API';
  }
  return `Updated ${date.toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  })}`;
}

function summaryToApiAsset(
  tenantID: string,
  inventoryID: string,
  asset: AssetSummary
): Asset {
  return {
    id: asset.id,
    tenantId: tenantID,
    inventoryId: inventoryID,
    kind: asset.kind,
    title: asset.title,
    description: asset.description,
    parentAssetId: asset.parentAssetId ?? null,
    lifecycleState: asset.lifecycleState,
    tags: [...(asset.tags ?? [])],
    customFields: {},
    createdAt: '',
    updatedAt: '',
    currentCheckout: asset.currentCheckout
  };
}

async function mapWithConcurrency<Input, Output>(
  items: readonly Input[],
  concurrency: number,
  mapper: (item: Input) => Promise<Output>
): Promise<readonly Output[]> {
  const results = new Array<Output>(items.length);
  let nextIndex = 0;
  const workerCount = Math.min(concurrency, items.length);

  await Promise.all(Array.from({ length: workerCount }, async () => {
    while (nextIndex < items.length) {
      const index = nextIndex;
      nextIndex += 1;
      results[index] = await mapper(items[index]);
    }
  }));

  return results;
}
