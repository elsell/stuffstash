import type {
  Asset,
  AssetSearchResult,
  Inventory,
  StuffStashClient,
  Tenant
} from '@stuff-stash/api-client';
import {
  AssetBrowsePage,
  AssetBrowsePageInput,
  CreateInventoryAssetInput,
  CreateInventoryAssetPhotoInput,
  AddInventoryAssetPhotoInput,
  InventoryAssetPhotoDirectUpload,
  InventorySummaryRepository,
  InventoryWorkspace,
  UpdateInventoryAssetInput
} from '../../application/home/InventorySummaryRepository';
import { assetId, AssetSummary } from '../../domain/assets/AssetSummary';
import {
  AccessRole,
  InventoryId,
  inventoryId,
  InventorySummary,
  tenantId
} from '../../domain/inventories/InventorySummary';
import type { LocationSummary } from '../../domain/locations/LocationSummary';

type InventoryApiClient = Pick<
  StuffStashClient,
  | 'listMyTenants'
  | 'listInventories'
  | 'listAssets'
  | 'createAsset'
  | 'updateAsset'
  | 'archiveAsset'
  | 'restoreAsset'
  | 'deleteAsset'
  | 'createAssetAttachment'
  | 'initiateAssetAttachmentDirectUpload'
  | 'completeAssetAttachmentDirectUpload'
  | 'deleteAssetAttachment'
  | 'searchAssets'
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

export class ApiInventorySummaryRepository implements InventorySummaryRepository {
  private selectedInventoryId: InventoryId | undefined;

  constructor(
    private readonly client: InventoryApiClient,
    private readonly configuredTenantId: string,
    private readonly directUploadTransport: DirectUploadTransport = new ExpoDirectUploadTransport()
  ) {}

  async getInventoryWorkspace(): Promise<InventoryWorkspace> {
    const tenantsPage = await this.client.listMyTenants(100);
    const tenants = tenantsPage.items;
    const inventoriesByTenant = await Promise.all(
      tenants.map(async (tenant) => {
        const inventoriesPage = await this.client.listInventories(tenant.id, 100);
        return Promise.all(
          inventoriesPage.items.map((item) => this.mapInventoryWithAssets(tenant, item))
        );
      })
    );
    const inventories = inventoriesByTenant.flat();
    const preferredInventory =
      inventories.find((inventory) => inventory.tenantId === tenantId(this.configuredTenantId)) ??
      inventories[0];
    const defaultInventory =
      inventories.find((inventory) => inventory.id === this.selectedInventoryId) ??
      preferredInventory;

    if (!defaultInventory) {
      throw new Error('API principal did not include any inventories.');
    }

    return {
      tenants: tenants.map(mapTenant),
      inventories,
      defaultInventoryId: defaultInventory.id
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
      parentAssetId: input.parentAssetId
    });

    return this.mapAssetWithPhoto(
      inventory.name,
      asset,
      inventory.assets.map((item) => summaryToApiAsset(inventory.tenantId, inventory.id, item))
    );
  }

  async updateAsset(input: UpdateInventoryAssetInput): Promise<AssetSummary> {
    const inventory = await this.getDefaultInventorySummary();
    const asset = await this.client.updateAsset(inventory.tenantId, inventory.id, input.assetId, {
      ...(input.title !== undefined ? { title: input.title } : {}),
      ...(input.description !== undefined ? { description: input.description } : {}),
      ...(input.parentAssetId !== undefined ? { parentAssetId: input.parentAssetId } : {})
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
      if (!isDirectUploadTargetSupported(directUpload.url)) {
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
    return input.query.trim().length > 0
      ? await this.searchInventoryAssetPage(inventory, input, knownAssets)
      : await this.listInventoryAssetPage(inventory, input, knownAssets);
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
      page.map((item) => this.mapAssetWithPrimaryPhoto(inventory.name, item.asset, siblings))
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
    inventory: Inventory
  ): Promise<InventorySummary> {
    const assets = await this.listRecentInventoryAssets(tenant.id, inventory.id);
    const locations = await Promise.all(
      assets
        .filter((asset) => asset.kind === 'location')
        .map(async (location) => mapLocation(location, assets, await this.primaryPhotoForAsset(location)))
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
      assets: mappedAssets
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

  private async photosForAsset(asset: Asset): Promise<readonly NonNullable<AssetSummary['photo']>[]> {
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
      return [];
    }

    return Promise.all(
      attachments
        .filter((item) => item.lifecycleState === 'active' && item.contentType.startsWith('image/'))
        .map(async (attachment) => {
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
        })
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
      selectedAssets.push(...filterAssetsByKind(page.items, input.kind));
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
    const selectedAssets: Asset[] = [];
    let cursor = input.cursor;
    let nextCursor: string | undefined;
    let hasMore = false;

    do {
      const pageSize = desiredMatches - selectedAssets.length;
      const page = await this.client.searchAssets(inventory.tenantId, input.query, {
        limit: pageSize,
        cursor,
        lifecycleState: input.lifecycleState
      });
      const inventoryAssets = page.items
        .filter((item) => item.inventory.id === inventory.id)
        .map((item) => item.asset);
      selectedAssets.push(...filterAssetsByKind(inventoryAssets, input.kind));
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
}

class ExpoDirectUploadTransport implements DirectUploadTransport {
  async upload(input: DirectUploadTransportInput): Promise<boolean> {
    if (isLocalDirectUploadURL(input.upload.url)) {
      return false;
    }
    if (!isDirectUploadHTTPTransportAllowed(input.upload.url)) {
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

function isDirectUploadTargetSupported(value: string): boolean {
  return isDirectUploadHTTPTransportAllowed(value) || isLocalDirectUploadURL(value);
}

function isDirectUploadHTTPTransportAllowed(value: string): boolean {
  const parsed = parseHTTPURL(value);
  if (!parsed) {
    return false;
  }
  if (parsed.protocol === 'https:') {
    return true;
  }
  return parsed.protocol === 'http:' && isLocalDevelopmentHost(parsed.hostname);
}

function parseHTTPURL(value: string): URL | undefined {
  try {
    const parsed = new URL(value);
    return parsed.protocol === 'https:' || parsed.protocol === 'http:' ? parsed : undefined;
  } catch {
    return undefined;
  }
}

function isLocalDevelopmentHost(hostname: string): boolean {
  const value = hostname.toLowerCase();
  if (value === 'localhost' || value.endsWith('.local')) {
    return true;
  }
  if (value === '127.0.0.1' || value === '::1' || value === '[::1]') {
    return true;
  }
  const octets = value.split('.').map((part) => Number.parseInt(part, 10));
  if (octets.length !== 4 || octets.some((part) => Number.isNaN(part) || part < 0 || part > 255)) {
    return false;
  }
  const [first, second] = octets;
  return first === 10
    || (first === 172 && second >= 16 && second <= 31)
    || (first === 192 && second === 168);
}

function isLocalDirectUploadURL(value: string): boolean {
  return value.startsWith('stuffstash-local://direct-uploads/');
}

function directUploadMethod(value: string): 'POST' | 'PUT' | 'PATCH' {
  switch (value.toUpperCase()) {
    case 'POST':
      return 'POST';
    case 'PATCH':
      return 'PATCH';
    default:
      return 'PUT';
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

function mapLocation(
  location: Asset,
  assets: readonly Asset[],
  photo?: AssetSummary['photo']
): LocationSummary {
  const children = assets.filter((asset) => asset.parentAssetId === location.id);

  return {
    id: assetId(location.id),
    inventoryId: inventoryId(location.inventoryId),
    title: location.title,
    description: location.description || 'Location asset',
    containedAssetCount: children.length,
    recentAssetTitles: children.slice(0, 3).map((asset) => asset.title),
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
    photo
  };
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
    customFields: {},
    createdAt: '',
    updatedAt: ''
  };
}
