import type {
  Asset,
  AssetPhotoReference,
  Inventory,
  StuffStashClient,
  Tenant
} from '@stuff-stash/api-client';
import {
  CreateInventoryAssetInput,
  CreateInventoryAssetPhotoInput,
  InventorySummaryRepository,
  InventoryWorkspace
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
  | 'createAssetAttachment'
  | 'searchAssets'
  | 'listAssetAttachments'
  | 'assetAttachmentThumbnailReference'
>;

export class ApiInventorySummaryRepository implements InventorySummaryRepository {
  private selectedInventoryId: InventoryId | undefined;

  constructor(
    private readonly client: InventoryApiClient,
    private readonly configuredTenantId: string
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

  async addAssetPhoto(
    assetIdValue: AssetSummary['id'],
    input: CreateInventoryAssetPhotoInput
  ): Promise<void> {
    const inventory = await this.getDefaultInventorySummary();
    await this.client.createAssetAttachment(inventory.tenantId, inventory.id, assetIdValue, {
      fileName: input.fileName,
      contentType: input.contentType,
      contentBase64: input.contentBase64
    });
  }

  async searchAssets(query: string): Promise<readonly AssetSummary[]> {
    const workspace = await this.getInventoryWorkspace();
    const inventory =
      workspace.inventories.find((item) => item.id === workspace.defaultInventoryId) ??
      workspace.inventories[0];

    if (!inventory) {
      throw new Error('API workspace did not include an inventory.');
    }

    const page = await this.client.searchAssets(inventory.tenantId, query, 50);
    const siblings = inventory.assets.map((item) =>
      summaryToApiAsset(inventory.tenantId, inventory.id, item)
    );
    return Promise.all(
      page.items.map((item) => this.mapAssetWithPhoto(inventory.name, item.asset, siblings))
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

    const page = await this.client.searchAssets(inventory.tenantId, query, 50);
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
    const assetsPage = await this.client.listAssets(tenant.id, inventory.id, 100, undefined, 'all');
    const assets = assetsPage.items;
    const locations = await Promise.all(
      assets
        .filter((asset) => asset.kind === 'location')
        .map(async (location) => mapLocation(location, assets, await this.primaryPhotoForAsset(location)))
    );

    const mappedAssets = await Promise.all(
      assets.map((asset) => this.mapAssetWithPhoto(inventory.name, asset, assets))
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

  private async mapAssetWithPhoto(
    inventoryName: string,
    asset: Asset,
    assets: readonly Asset[]
  ): Promise<AssetSummary> {
    const photo = await this.primaryPhotoForAsset(asset);
    return mapAsset(inventoryName, asset, assets, photo);
  }

  private async primaryPhotoForAsset(asset: Asset): Promise<AssetPhotoReference | undefined> {
    let attachmentsPage;
    try {
      attachmentsPage = await this.client.listAssetAttachments(
        asset.tenantId,
        asset.inventoryId,
        asset.id,
        1
      );
    } catch {
      return undefined;
    }

    const attachment = attachmentsPage.items.find((item) => item.lifecycleState === 'active');
    if (!attachment || !attachment.contentType.startsWith('image/')) {
      return undefined;
    }

    return this.client.assetAttachmentThumbnailReference(
      asset.tenantId,
      asset.inventoryId,
      asset.id,
      attachment.id
    );
  }
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
  photo?: AssetPhotoReference
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

function mapAsset(
  inventoryName: string,
  asset: Asset,
  assets: readonly Asset[],
  photo?: AssetPhotoReference
): AssetSummary {
  const parent = asset.parentAssetId
    ? assets.find((candidate) => candidate.id === asset.parentAssetId)
    : undefined;

  return {
    id: assetId(asset.id),
    title: asset.title,
    kind: asset.kind,
    lifecycleState: asset.lifecycleState,
    locationLabel: parent?.title ?? 'Inventory root',
    locationTrail: [inventoryName, parent?.title, asset.title].filter(isString),
    description: asset.description,
    updatedAtLabel: 'Loaded from API',
    hasPhoto: photo !== undefined,
    photo
  };
}

function isString(value: string | undefined): value is string {
  return typeof value === 'string' && value.length > 0;
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
    parentAssetId: null,
    lifecycleState: asset.lifecycleState
  };
}
