import type {
  Asset,
  AssetTag,
  CheckedOutAsset,
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
  HomeDashboardSnapshot,
  HomeDashboardSnapshotRepository,
  InventoryAssetPhotoDirectUpload,
  InventorySummaryRepository,
  InventoryWorkspace,
  UpdateInventoryAssetInput
} from '../../application/home/InventorySummaryRepository';
import type { InventoryMapAssetRepository } from '../../application/assets/InventoryMapQuery';
import type { AssetCoreSnapshot } from '../../application/assets/AssetCoreQuery';
import type { AssetContentsSnapshot } from '../../application/assets/AssetContentsQuery';
import type { InventoryAssetsSnapshot } from '../../application/assets/InventoryAssetsQuery';
import type {
  AssetDetailWorkspaceRepository,
  AssetDetailWorkspaceSnapshot
} from '../../application/assets/AssetDetailWorkspaceRepository';
import { assetId, AssetSummary, type AssetTagSummary } from '../../domain/assets/AssetSummary';
import {
  AccessRole,
  InventoryId,
  inventoryId,
  InventorySummary,
  tenantId
} from '../../domain/inventories/InventorySummary';
import type { LocationSummary } from '../../domain/locations/LocationSummary';
import type { LocationsSnapshot } from '../../application/locations/LocationsQuery';
import type { LocationAssetsSnapshot } from '../../application/locations/LocationAssetsQuery';
import {
  ignoreInventoryMutations,
  type InventoryMutationKind,
  type InventoryMutationObserver
} from '../../application/home/InventoryMutationObserver';
import type { ReadRequest } from '../../application/shared/ReadRequest';
import type { AddAssetContext } from '../../application/add/AddAssetContextQuery';
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
  | 'getAsset'
  | 'listAssetTags'
  | 'createAsset'
  | 'createAssetTag'
  | 'updateAsset'
  | 'checkoutAsset'
  | 'returnAsset'
  | 'updateReturnedCheckoutDetails'
  | 'applyUndoableOperation'
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

type LoadedInventoryWorkspace = {
  readonly workspace: InventoryWorkspace;
  readonly defaultPlacementAssets: readonly Asset[];
};

type MappedInventory = {
  readonly summary: InventorySummary;
  readonly placementAssets: readonly Asset[];
};

type InventoryDirectory = {
  readonly tenants: readonly Tenant[];
  readonly availableInventories: readonly {
    readonly tenant: Tenant;
    readonly inventory: Inventory;
  }[];
};

type AssetDetailWorkspaceSelection = {
  readonly assets: readonly Asset[];
  readonly photoAssetIds: ReadonlySet<string>;
};

function selectAssetDetailWorkspace(
  root: Asset,
  assets: readonly Asset[]
): AssetDetailWorkspaceSelection {
  const childrenByParent = new Map<string, Asset[]>();
  const indexedIds = new Set<string>();
  for (const asset of assets) {
    if (indexedIds.has(asset.id)) {
      continue;
    }
    indexedIds.add(asset.id);
    if (!asset.parentAssetId) {
      continue;
    }
    const children = childrenByParent.get(asset.parentAssetId) ?? [];
    children.push(asset);
    childrenByParent.set(asset.parentAssetId, children);
  }
  for (const children of childrenByParent.values()) {
    children.sort((left, right) => left.id.localeCompare(right.id));
  }

  if (root.kind === 'container') {
    const directChildren = childrenByParent.get(root.id) ?? [];
    const workspaceAssets = [root, ...directChildren];
    return {
      assets: workspaceAssets,
      photoAssetIds: new Set(workspaceAssets.map((asset) => asset.id))
    };
  }

  const subtree: Asset[] = [];
  const photoAssetIds = new Set<string>([root.id]);
  const visited = new Set<string>();
  const pending = [root];
  while (pending.length > 0) {
    const asset = pending.pop();
    if (!asset || visited.has(asset.id)) {
      continue;
    }
    visited.add(asset.id);
    subtree.push(asset);
    if (asset.kind === 'item' || asset.parentAssetId === root.id) {
      photoAssetIds.add(asset.id);
    }
    const children = childrenByParent.get(asset.id) ?? [];
    for (let index = children.length - 1; index >= 0; index -= 1) {
      const child = children[index];
      if (child) {
        pending.push(child);
      }
    }
  }
  return { assets: subtree, photoAssetIds };
}

export class ApiInventorySummaryRepository implements InventorySummaryRepository, InventoryMapAssetRepository, HomeDashboardSnapshotRepository, AssetDetailWorkspaceRepository {
  private selectedInventoryId: InventoryId | undefined;
  private selectedInventoryIdentity: { readonly tenant: Tenant; readonly inventory: Inventory } | undefined;
  private inventoryDirectoryPromise: Promise<InventoryDirectory> | undefined;
  private readonly directUploadTransport: DirectUploadTransport;

  constructor(
    private readonly client: InventoryApiClient,
    private readonly configuredTenantId: string,
    directUploadTransport?: DirectUploadTransport,
    private readonly sessionScopeId = 'mobile-composition',
    private readonly directUploadPolicy: DirectUploadTargetPolicy = {},
    private readonly mutationObserver: InventoryMutationObserver = ignoreInventoryMutations
  ) {
    this.directUploadTransport = directUploadTransport ?? new ExpoDirectUploadTransport(directUploadPolicy);
  }

  async getInventoryWorkspace(): Promise<InventoryWorkspace> {
    return (await this.loadInventoryWorkspace()).workspace;
  }

  async getCurrentTenantId(): Promise<string> {
    return (await this.getSelectedInventoryIdentity()).tenant.id;
  }

  async getCurrentInventoryScope(request: ReadRequest = {}): Promise<{ readonly tenantId: string; readonly inventoryId: string }> {
    const selected = await this.getSelectedInventoryIdentity(request.signal);
    return {
      tenantId: selected.tenant.id,
      inventoryId: selected.inventory.id
    };
  }

  async getCurrentSettingsScope(request: ReadRequest = {}): Promise<{
    readonly tenantId: string;
    readonly inventory: { readonly id: string; readonly name: string; readonly permissions: readonly string[] };
  }> {
    const selected = await this.getSelectedInventoryIdentity(request.signal);
    return {
      tenantId: selected.tenant.id,
      inventory: {
        id: selected.inventory.id,
        name: selected.inventory.name,
        permissions: [...selected.inventory.access.permissions]
      }
    };
  }

  private async loadInventoryWorkspace(): Promise<LoadedInventoryWorkspace> {
    const { tenants, availableInventories } = await this.loadInventoryDirectory();
    const preferredInventory =
      availableInventories.find((item) => item.tenant.id === this.configuredTenantId) ??
      availableInventories[0];
    const defaultInventory =
      availableInventories.find((item) => inventoryId(item.inventory.id) === this.selectedInventoryId) ??
      preferredInventory;

    if (!defaultInventory) {
      throw new Error('API principal did not include any inventories.');
    }
    this.selectedInventoryIdentity = defaultInventory;

    const mappedInventories = await Promise.all(
      availableInventories.map((item) => {
        const isSelected = item.tenant.id === defaultInventory.tenant.id &&
          item.inventory.id === defaultInventory.inventory.id;
        return isSelected
          ? this.mapInventoryWithAssets(item.tenant, item.inventory, true)
          : Promise.resolve({
              summary: emptyInventorySummary(item.tenant, item.inventory),
              placementAssets: []
            });
      })
    );
    const inventories = mappedInventories.map((item) => item.summary);
    const mappedDefaultInventory = mappedInventories.find(
      (item) => item.summary.id === inventoryId(defaultInventory.inventory.id)
    );

    if (!mappedDefaultInventory) {
      throw new Error('API workspace did not hydrate the selected inventory.');
    }

    return {
      workspace: {
        tenants: tenants.map(mapTenant),
        inventories,
        defaultInventoryId: inventoryId(defaultInventory.inventory.id)
      },
      defaultPlacementAssets: mappedDefaultInventory.placementAssets
    };
  }

  async getDefaultInventorySummary(): Promise<InventorySummary> {
    return (await this.getDefaultInventoryContext()).inventory;
  }

  private async getDefaultInventoryContext(): Promise<{
    readonly inventory: InventorySummary;
    readonly placementAssets: readonly Asset[];
  }> {
    const loadedWorkspace = await this.loadInventoryWorkspace();
    const { workspace } = loadedWorkspace;
    const inventory =
      workspace.inventories.find((item) => item.id === workspace.defaultInventoryId) ??
      workspace.inventories[0];

    if (!inventory) {
      throw new Error('API workspace did not include an inventory.');
    }

    return {
      inventory,
      placementAssets: loadedWorkspace.defaultPlacementAssets
    };
  }

  async getHomeDashboardSnapshot(request: ReadRequest = {}): Promise<HomeDashboardSnapshot> {
    const { tenants, availableInventories } = await this.loadInventoryDirectory(request.signal);
    const selected = await this.getSelectedInventoryIdentity(request.signal);
    const [recentAssets, checkedOutAssets] = await Promise.all([
      this.listHomeRecentAssets(selected.tenant.id, selected.inventory.id, request.signal),
      this.listCheckedOutInventoryAssets(selected.tenant.id, selected.inventory.id, request.signal)
    ]);
    const checkedOutSourceAssets = checkedOutAssets.map((item) => item.asset);
    const visibleAssets = [...recentAssets, ...checkedOutSourceAssets];
    const ancestorAssets = await this.loadAncestorsForAssets(visibleAssets, request.signal);
    const ancestryAssets = placementAssetsFromFullTree(
      [...ancestorAssets, ...visibleAssets],
      visibleAssets
    );
    const mappedCheckedOutAssets = await Promise.all(
      checkedOutAssets.map((item) => this.mapAssetWithPrimaryPhoto(selected.inventory.name, item.asset, ancestryAssets))
    );
    const checkedOutByAssetId = new Map(mappedCheckedOutAssets.map((asset) => [asset.id, asset]));
    const mappedRecentAssets = await Promise.all(
      recentAssets.map((asset) => checkedOutByAssetId.get(assetId(asset.id)) ??
        this.mapAssetWithPrimaryPhoto(selected.inventory.name, asset, ancestryAssets))
    );
    const inventories = availableInventories.map(({ tenant, inventory }) => ({
      id: inventoryId(inventory.id),
      tenantId: tenantId(tenant.id),
      name: inventory.name,
      role: mapAccessRole(inventory.access.relationship),
      permissions: [...inventory.access.permissions],
      description: '',
      updatedAtLabel: 'Loaded from API',
      locationCount: 0,
      locations: [],
      assets: inventory.id === selected.inventory.id ? mappedRecentAssets : [],
      assetTags: []
    } satisfies InventorySummary));
    const workspace: InventoryWorkspace = {
      tenants: tenants.map(mapTenant),
      inventories,
      defaultInventoryId: inventoryId(selected.inventory.id)
    };

    return {
      workspace,
      checkedOutAssets: mappedCheckedOutAssets
    };
  }

  async getInventoryAssetsSnapshot(request: ReadRequest = {}): Promise<InventoryAssetsSnapshot> {
    const selected = await this.getSelectedInventoryIdentity(request.signal);
    const [recentAssets, activeAssets] = await Promise.all([
      this.listRecentInventoryAssets(selected.tenant.id, selected.inventory.id, request.signal),
      this.listAllActiveInventoryAssets(selected.tenant.id, selected.inventory.id, request.signal)
    ]);
    const ancestryAssets = placementAssetsFromFullTree(activeAssets, recentAssets);
    return {
      inventoryName: selected.inventory.name,
      assets: await Promise.all(
        recentAssets.map((asset) => this.mapAssetWithPrimaryPhoto(
          selected.inventory.name,
          asset,
          ancestryAssets
        ))
      )
    };
  }

  async getLocationsSnapshot(request: ReadRequest = {}): Promise<LocationsSnapshot> {
    const selected = await this.getSelectedInventoryIdentity(request.signal);
    const assets = await this.listAllActiveInventoryAssets(selected.tenant.id, selected.inventory.id, request.signal);
    const locations = await Promise.all(
      assets
        .filter((asset) => asset.kind === 'location')
        .map(async (location) => mapLocation(
          location,
          assets,
          await this.primaryPhotoForAsset(location)
        ))
    );
    return {
      canAdd: selected.inventory.access.permissions.includes('create_asset'),
      tenantName: selected.tenant.name,
      inventoryName: selected.inventory.name,
      locations
    };
  }

  async getLocationAssetsSnapshot(locationIdValue: string, request: ReadRequest = {}): Promise<LocationAssetsSnapshot> {
    const selected = await this.getSelectedInventoryIdentity(request.signal);
    const assets = await this.listAllActiveInventoryAssets(selected.tenant.id, selected.inventory.id, request.signal);
    const location = assets.find((candidate) =>
      candidate.id === locationIdValue && candidate.kind === 'location'
    );
    if (!location) {
      throw new Error('Location is not available in the selected inventory.');
    }
    const containedAssets = assets.filter((candidate) =>
      candidate.id !== location.id &&
      ancestorTrail(candidate, assets).some((ancestor) => ancestor.id === location.id)
    );
    return {
      locationId: location.id,
      locationTitle: location.title,
      inventoryName: selected.inventory.name,
      assets: await Promise.all(containedAssets.map((asset) =>
        this.mapAssetWithPrimaryPhoto(selected.inventory.name, asset, assets)
      ))
    };
  }

  async getAddAssetContext(request: ReadRequest = {}): Promise<AddAssetContext> {
    const selected = await this.getSelectedInventoryIdentity(request.signal);
    const tags = await this.listAllInventoryTags(
      selected.tenant.id,
      selected.inventory.id,
      request.signal
    );
    return {
      tenantId: selected.tenant.id,
      tenantName: selected.tenant.name,
      inventoryId: selected.inventory.id,
      inventoryName: selected.inventory.name,
      canAdd: selected.inventory.access.permissions.includes('create_asset'),
      assetTags: tags.map(mapAssetTag)
    };
  }

  async getInventoryAssetTags(request: ReadRequest = {}): Promise<readonly AssetTagSummary[]> {
    const selected = await this.getSelectedInventoryIdentity(request.signal);
    const tags = await this.listAllInventoryTags(
      selected.tenant.id,
      selected.inventory.id,
      request.signal
    );
    return tags.map(mapAssetTag);
  }

  async getAssetCore(
    selectedAssetId: AssetSummary['id'],
    request: ReadRequest = {}
  ): Promise<AssetCoreSnapshot> {
    const selected = await this.getSelectedInventoryIdentity(request.signal);
    const asset = await this.client.getAsset(
      selected.tenant.id,
      selected.inventory.id,
      selectedAssetId,
      request.signal
    );
    return {
      tenantId: tenantId(selected.tenant.id),
      inventoryId: inventoryId(selected.inventory.id),
      permissions: selected.inventory.access.permissions,
      revision: asset.updatedAt,
      asset: mapAsset(selected.inventory.name, asset, [asset])
    };
  }

  async getAssetPlacement(core: AssetCoreSnapshot, request: ReadRequest = {}): Promise<AssetSummary> {
    const selected = await this.getSelectedInventoryIdentity(request.signal);
    if (selected.tenant.id !== core.tenantId || selected.inventory.id !== core.inventoryId) {
      throw new Error('Asset scope no longer matches the selected inventory.');
    }
    const source = summaryToApiAsset(core.tenantId, core.inventoryId, core.asset);
    const ancestors = await this.loadAssetAncestors(source, request.signal);
    return mapAsset(selected.inventory.name, source, [...ancestors, source]);
  }

  async getAssetContents(
    core: AssetCoreSnapshot,
    request: ReadRequest = {}
  ): Promise<AssetContentsSnapshot> {
    const selected = await this.getSelectedInventoryIdentity(request.signal);
    if (
      tenantId(selected.tenant.id) !== core.tenantId
      || inventoryId(selected.inventory.id) !== core.inventoryId
    ) {
      throw new Error('Asset scope no longer matches the selected inventory.');
    }
    const sourceAsset = summaryToApiAsset(core.tenantId, core.inventoryId, core.asset);
    const isActiveContainableAsset = sourceAsset.lifecycleState === 'active'
      && (sourceAsset.kind === 'location' || sourceAsset.kind === 'container');
    if (!isActiveContainableAsset) {
      const ancestors = await this.loadAssetAncestors(sourceAsset, request.signal);
      return {
        asset: mapAsset(selected.inventory.name, sourceAsset, [...ancestors, sourceAsset]),
        allAssets: []
      };
    }

    const activeAssets = await this.listAllActiveInventoryAssets(
      selected.tenant.id,
      selected.inventory.id,
      request.signal
    );
    const selectedFromTraversal = activeAssets.find((asset) => asset.id === sourceAsset.id) ?? sourceAsset;
    const workspace = selectAssetDetailWorkspace(selectedFromTraversal, activeAssets);
    const assets = await this.mapAssetsWithMapPhotos(
      selected.inventory.name,
      workspace.assets,
      activeAssets,
      new Set(workspace.assets.filter((asset) => asset.id !== core.asset.id).map((asset) => asset.id))
    );
    const asset = assets.find((candidate) => candidate.id === core.asset.id);
    if (!asset) {
      throw new Error('Asset is not available in the selected inventory.');
    }
    return { asset, allAssets: assets };
  }

  async getAssetPhotos(
    core: AssetCoreSnapshot,
    request: ReadRequest = {}
  ): Promise<AssetSummary['photos']> {
    const selected = await this.getSelectedInventoryIdentity(request.signal);
    if (
      tenantId(selected.tenant.id) !== core.tenantId
      || inventoryId(selected.inventory.id) !== core.inventoryId
    ) {
      throw new Error('Asset scope no longer matches the selected inventory.');
    }
    const asset = summaryToApiAsset(core.tenantId, core.inventoryId, core.asset);
    return this.photosForAsset(asset, {
      allowAttachmentListFailure: false,
      signal: request.signal
    });
  }

  async getAssetDetail(input: GetInventoryAssetDetailInput, request: ReadRequest = {}): Promise<AssetSummary> {
    const asset = summaryToApiAsset(input.tenantId, input.inventoryId, input.asset);
    const photos = await this.photosForAsset(asset, {
      allowAttachmentListFailure: false,
      signal: request.signal
    });

    return {
      ...input.asset,
      hasPhoto: photos.length > 0,
      photos,
      photo: photos[0]
    };
  }

  async selectInventory(selectedInventoryId: InventoryId): Promise<void> {
    let directory = await this.loadInventoryDirectory();
    let selection = directory.availableInventories.find((item) =>
      inventoryId(item.inventory.id) === selectedInventoryId
    );

    if (!selection) {
      directory = await this.refreshInventoryDirectory();
      selection = directory.availableInventories.find((item) =>
        inventoryId(item.inventory.id) === selectedInventoryId
      );
    }

    if (!selection) {
      throw new Error('Selected inventory is not available in the configured tenant.');
    }

    this.selectedInventoryId = selectedInventoryId;
    this.selectedInventoryIdentity = selection;
  }

  async createAsset(input: CreateInventoryAssetInput): Promise<AssetSummary> {
    const selected = await this.getSelectedInventoryIdentity();
    const inventory = emptyInventorySummary(selected.tenant, selected.inventory);
    const parent = input.parentAssetId ? await this.client.getAsset(selected.tenant.id, selected.inventory.id, input.parentAssetId) : undefined;
    const currentPlacementAssets = parent ? [parent, ...await this.loadAssetAncestors(parent)] : [];
    const asset = await this.client.createAsset(inventory.tenantId, inventory.id, {
      kind: input.kind,
      title: input.title,
      description: input.description,
      parentAssetId: input.parentAssetId,
      ...(input.tagIds !== undefined ? { tagIds: [...input.tagIds] } : {})
    });
    this.observeMutation(
      'asset_created',
      inventory.tenantId,
      inventory.id,
      asset.id,
      ancestorTrail(asset, currentPlacementAssets).map((ancestor) => ancestor.id),
      parent?.kind === 'item' ? parent.id : undefined
    );

    const placementAssets = placementAssetsWithSelectedOverrides(
      currentPlacementAssets,
      [asset]
    );
    return mapAsset(inventory.name, asset, placementAssets);
  }

  async createAssetTag(input: CreateInventoryAssetTagInput): Promise<AssetTagSummary> {
    const selected = await this.getSelectedInventoryIdentity();
    const tag = await this.client.createAssetTag(selected.tenant.id, selected.inventory.id, {
      displayName: input.displayName,
      ...(input.color !== undefined ? { color: input.color } : {})
    });
    this.observeMutation('asset_tag_created', selected.tenant.id, selected.inventory.id);
    return mapAssetTag(tag);
  }

  async updateAsset(input: UpdateInventoryAssetInput): Promise<AssetSummary> {
    const selected = await this.getSelectedInventoryIdentity();
    const inventory = emptyInventorySummary(selected.tenant, selected.inventory);
    const previous = await this.client.getAsset(selected.tenant.id, selected.inventory.id, input.assetId);
    const destination = { ...previous, parentAssetId: input.parentAssetId !== undefined ? input.parentAssetId : previous.parentAssetId };
    const placementAssets = [previous, ...await this.loadAncestorsForAssets([previous, destination])];
    const asset = await this.client.updateAsset(inventory.tenantId, inventory.id, input.assetId, {
      ...(input.title !== undefined ? { title: input.title } : {}),
      ...(input.description !== undefined ? { description: input.description } : {}),
      ...(input.parentAssetId !== undefined ? { parentAssetId: input.parentAssetId } : {}),
      ...(input.tagIds !== undefined ? { tagIds: [...input.tagIds] } : {})
    });
    this.observeMutation(
      'asset_updated',
      inventory.tenantId,
      inventory.id,
      asset.id,
      [
        ...ancestorTrail(previous, placementAssets),
        ...ancestorTrail(asset, placementAssets)
      ].map((ancestor) => ancestor.id)
    );

    const knownAssets = placementAssetsWithSelectedOverrides(
      placementAssets,
      [asset]
    );
    return mapAsset(inventory.name, asset, knownAssets);
  }

  async addAssetPhoto(
    assetIdValue: AssetSummary['id'],
    input: CreateInventoryAssetPhotoInput
  ): Promise<void> {
    const selected = await this.getSelectedInventoryIdentity();
    await this.addInventoryAssetPhoto({
      tenantId: tenantId(selected.tenant.id),
      inventoryId: inventoryId(selected.inventory.id),
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
        this.observeMutation('asset_photo_changed', input.tenantId, input.inventoryId, input.assetId);
        return;
      }
    }

    await this.client.createAssetAttachment(input.tenantId, input.inventoryId, input.assetId, {
      fileName: input.fileName,
      contentType: input.contentType,
      contentBase64: await attachmentContentBase64(input)
    });
    this.observeMutation('asset_photo_changed', input.tenantId, input.inventoryId, input.assetId);
  }

  async deleteAssetPhoto(assetIdValue: AssetSummary['id'], photoId: string): Promise<void> {
    const selected = await this.getSelectedInventoryIdentity();
    await this.client.deleteAssetAttachment(selected.tenant.id, selected.inventory.id, assetIdValue, photoId);
    this.observeMutation('asset_photo_changed', selected.tenant.id, selected.inventory.id, assetIdValue);
  }

  async archiveAsset(assetIdValue: AssetSummary['id']): Promise<void> {
    const selected = await this.getSelectedInventoryIdentity();
    await this.client.archiveAsset(selected.tenant.id, selected.inventory.id, assetIdValue);
    this.observeMutation('asset_lifecycle_changed', selected.tenant.id, selected.inventory.id, assetIdValue);
  }

  async restoreAsset(assetIdValue: AssetSummary['id']): Promise<void> {
    const selected = await this.getSelectedInventoryIdentity();
    await this.client.restoreAsset(selected.tenant.id, selected.inventory.id, assetIdValue);
    this.observeMutation('asset_lifecycle_changed', selected.tenant.id, selected.inventory.id, assetIdValue);
  }

  async deleteAsset(assetIdValue: AssetSummary['id']): Promise<void> {
    const selected = await this.getSelectedInventoryIdentity();
    await this.client.deleteAsset(selected.tenant.id, selected.inventory.id, assetIdValue);
    this.observeMutation('asset_lifecycle_changed', selected.tenant.id, selected.inventory.id, assetIdValue);
  }

  async checkoutAsset(assetIdValue: AssetSummary['id'], input: { readonly details?: string } = {}) {
    const selected = await this.getSelectedInventoryIdentity();
    const checkout = await this.client.checkoutAsset(selected.tenant.id, selected.inventory.id, assetIdValue, input);
    this.observeMutation('asset_checkout_changed', selected.tenant.id, selected.inventory.id, assetIdValue);
    return {
      id: checkout.id,
      assetId: assetId(checkout.assetId),
      undoableOperationId: checkout.undoableOperationId
    };
  }

  async returnAsset(assetIdValue: AssetSummary['id'], input: { readonly details?: string } = {}) {
    const selected = await this.getSelectedInventoryIdentity();
    const checkout = await this.client.returnAsset(selected.tenant.id, selected.inventory.id, assetIdValue, input);
    this.observeMutation('asset_checkout_changed', selected.tenant.id, selected.inventory.id, assetIdValue);
    return {
      id: checkout.id,
      assetId: assetId(checkout.assetId),
      undoableOperationId: checkout.undoableOperationId
    };
  }

  async updateReturnedCheckoutDetails(assetIdValue: AssetSummary['id'], checkoutId: string, input: { readonly details?: string } = {}) {
    const selected = await this.getSelectedInventoryIdentity();
    const checkout = await this.client.updateReturnedCheckoutDetails(selected.tenant.id, selected.inventory.id, assetIdValue, checkoutId, input);
    this.observeMutation('asset_checkout_changed', selected.tenant.id, selected.inventory.id, assetIdValue);
    return {
      id: checkout.id,
      assetId: assetId(checkout.assetId),
      undoableOperationId: checkout.undoableOperationId
    };
  }

  async undoInventoryOperation(operationId: string): Promise<void> {
    const selected = await this.getSelectedInventoryIdentity();
    const asset = await this.client.applyUndoableOperation(selected.tenant.id, selected.inventory.id, operationId, 'undo');
    this.observeMutation('operation_reversed', selected.tenant.id, selected.inventory.id, asset.id, asset.parentAssetId ? [asset.parentAssetId] : []);
  }

  async browseAssets(input: AssetBrowsePageInput): Promise<AssetBrowsePage> {
    const selected = await this.getSelectedInventoryIdentity(input.signal);
    const inventory = emptyInventorySummary(selected.tenant, selected.inventory);
    const hasTagFilters = (input.tagIds?.length ?? 0) > 0;
    if (!hasTagFilters && input.query.trim().length === 0 && input.checkoutState === 'checked_out') {
      return await this.listCheckedOutInventoryAssetPage(inventory, input);
    }
    return input.query.trim().length > 0 || hasTagFilters
      ? await this.searchInventoryAssetPage(inventory, input)
      : await this.listInventoryAssetPage(inventory, input);
  }

  async listActiveInventoryMapAssets(request: ReadRequest = {}): Promise<{
    readonly sessionScopeId: string;
    readonly tenantId: InventorySummary['tenantId'];
    readonly inventoryId: InventorySummary['id'];
    readonly inventoryName: string;
    readonly permissions: readonly string[];
    readonly assets: readonly AssetSummary[];
  }> {
    const inventory = await this.getSelectedInventoryIdentity(request.signal);
    const activeAssets = await this.listAllActiveInventoryAssets(tenantId(inventory.tenant.id), inventory.inventory.id, request.signal);
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

  async getAssetDetailWorkspace(
    selectedAssetId: AssetSummary['id'],
    request: ReadRequest = {}
  ): Promise<AssetDetailWorkspaceSnapshot | undefined> {
    const inventory = await this.getSelectedInventoryIdentity(request.signal);
    const selectedAsset = await this.client.getAsset(
      inventory.tenant.id,
      inventory.inventory.id,
      selectedAssetId,
      request.signal
    );
    const isActiveContainableAsset = selectedAsset.lifecycleState === 'active'
      && (selectedAsset.kind === 'location' || selectedAsset.kind === 'container');
    if (!isActiveContainableAsset) {
      const ancestors = await this.loadAssetAncestors(selectedAsset, request.signal);
      const mapped = await this.mapAssetsWithMapPhotos(
        inventory.inventory.name,
        [selectedAsset],
        [...ancestors, selectedAsset]
      );
      const asset = mapped[0];
      return asset ? {
        tenantId: tenantId(inventory.tenant.id),
        inventoryId: inventoryId(inventory.inventory.id),
        permissions: inventory.inventory.access.permissions,
        asset,
        allAssets: []
      } : undefined;
    }
    const activeAssets = await this.listAllActiveInventoryAssets(
      tenantId(inventory.tenant.id),
      inventory.inventory.id,
      request.signal
    );
    const selectedFromTraversal = activeAssets.find((asset) => asset.id === selectedAssetId) ?? selectedAsset;
    const workspace = selectAssetDetailWorkspace(selectedFromTraversal, activeAssets);
    const assets = await this.mapAssetsWithMapPhotos(
      inventory.inventory.name,
      workspace.assets,
      activeAssets,
      workspace.photoAssetIds
    );
    const asset = assets.find((candidate) => candidate.id === selectedAssetId);
    if (!asset) {
      return undefined;
    }
    return {
      tenantId: tenantId(inventory.tenant.id),
      inventoryId: inventoryId(inventory.inventory.id),
      permissions: inventory.inventory.access.permissions,
      asset,
      allAssets: assets
    };
  }

  private async loadAssetAncestors(asset: Asset, signal?: AbortSignal): Promise<readonly Asset[]> {
    const ancestors: Asset[] = [];
    const visited = new Set<string>([asset.id]);
    let parentId = asset.parentAssetId ?? undefined;
    while (parentId && !visited.has(parentId)) {
      visited.add(parentId);
      const parent = await this.client.getAsset(asset.tenantId, asset.inventoryId, parentId, signal);
      ancestors.unshift(parent);
      parentId = parent.parentAssetId ?? undefined;
    }
    return ancestors;
  }

  private async loadAncestorsForAssets(
    assets: readonly Asset[],
    signal?: AbortSignal
  ): Promise<readonly Asset[]> {
    const knownAssets = new Map(assets.map((asset) => [asset.id, asset]));
    const pendingAssets = new Map<string, Promise<Asset>>();
    const loadParent = async (source: Asset, visited = new Set<string>()): Promise<void> => {
      if (visited.has(source.id)) {
        return;
      }
      const nextVisited = new Set(visited).add(source.id);
      const parentId = source.parentAssetId ?? undefined;
      if (!parentId || parentId === source.id) {
        return;
      }
      let parent = knownAssets.get(parentId);
      if (!parent) {
        let pending = pendingAssets.get(parentId);
        if (!pending) {
          pending = this.client.getAsset(source.tenantId, source.inventoryId, parentId, signal);
          pendingAssets.set(parentId, pending);
        }
        parent = await pending;
        knownAssets.set(parent.id, parent);
      }
      await loadParent(parent, nextVisited);
    };
    await Promise.all(assets.map((asset) => loadParent(asset)));
    const visibleIds = new Set(assets.map((asset) => asset.id));
    return [...knownAssets.values()].filter((asset) => !visibleIds.has(asset.id));
  }

  async listParentCandidates(query: string, request: ReadRequest = {}): Promise<readonly AssetSummary[]> {
    const page = await this.browseAssets({ query, limit: query ? 50 : 5, lifecycleState: 'active', kind: 'all', checkoutState: 'any', sort: 'updated_desc', signal: request.signal });
    return page.assets;
  }

  async searchAssets(query: string): Promise<readonly AssetSummary[]> {
    const { inventory, placementAssets } = await this.getDefaultInventoryContext();

    const page = await this.searchSelectedInventoryAssets(inventory.tenantId, inventory.id, query, 50);
    const siblings = placementAssetsFromFullTree(
      placementAssets,
      inventory.assets.map((item) =>
        summaryToApiAsset(inventory.tenantId, inventory.id, item)
      )
    );
    return Promise.all(
      page.map((item) =>
        this.mapAssetWithPrimaryPhoto(
          inventory.name,
          item.asset,
          siblings,
          { resolveMissingParentFromKnownAsset: true }
        )
      )
    );
  }

  async searchLocations(query: string): Promise<readonly LocationSummary[]> {
    const selected = await this.getSelectedInventoryIdentity();
    const inventory = emptyInventorySummary(selected.tenant, selected.inventory);
    const placementAssets = await this.listAllActiveInventoryAssets(selected.tenant.id, selected.inventory.id);

    const page = await this.client.searchAssets(inventory.tenantId, query, { limit: 50 });
    const locationAssets = page.items
      .filter((item) => item.inventory.id === inventory.id && item.asset.kind === 'location')
      .map((item) => item.asset);
    const knownAssets = placementAssets;

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
  ): Promise<MappedInventory> {
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

    const ancestryAssets = placementAssetsFromFullTree(locationSourceAssets, assets);
    const mappedAssets = await Promise.all(
      assets.map((asset) => this.mapAssetWithPrimaryPhoto(inventory.name, asset, ancestryAssets))
    );

    return {
      summary: {
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
      },
      placementAssets: ancestryAssets
    };
  }

  private async listRecentInventoryAssets(
    tenantID: string,
    inventoryID: string,
    signal?: AbortSignal
  ): Promise<readonly Asset[]> {
    const page = await this.client.listAssets(
      tenantID,
      inventoryID,
      inventoryAssetPageSize,
      undefined,
      'all',
      'updated_desc',
      signal
    );
    return page.items;
  }

  private async listHomeRecentAssets(
    tenantID: string,
    inventoryID: string,
    signal?: AbortSignal
  ): Promise<readonly Asset[]> {
    const page = await this.client.listAssets(
      tenantID,
      inventoryID,
      10,
      undefined,
      'active',
      'updated_desc',
      signal
    );
    return page.items;
  }

  private async listCheckedOutInventoryAssets(
    tenantID: string,
    inventoryID: string,
    signal?: AbortSignal
  ): Promise<readonly CheckedOutAsset[]> {
    const checkedOutAssets: CheckedOutAsset[] = [];
    let cursor: string | undefined;

    do {
      const page = await this.client.listCheckedOutAssets(tenantID, inventoryID, 10, cursor, signal);
      checkedOutAssets.push(...page.items);
      cursor = page.pagination.nextCursor ?? undefined;
    } while (cursor && checkedOutAssets.length < 10);

    return checkedOutAssets.slice(0, 10);
  }

  private async listAllInventoryTags(
    tenantID: string,
    inventoryID: string,
    signal?: AbortSignal
  ): Promise<readonly AssetTag[]> {
    const tags: AssetTag[] = [];
    let cursor: string | undefined;

    do {
      const page = await this.client.listAssetTags(tenantID, inventoryID, 100, cursor, signal);
      tags.push(...page.items);
      cursor = page.pagination.nextCursor ?? undefined;
    } while (cursor);

    return tags;
  }

  private async getSelectedInventoryIdentity(signal?: AbortSignal): Promise<{
    readonly tenant: Tenant;
    readonly inventory: Inventory;
  }> {
    if (this.selectedInventoryIdentity) {
      return this.selectedInventoryIdentity;
    }
    const { availableInventories: inventories } = await this.loadInventoryDirectory(signal);
    const preferredInventory =
      inventories.find((item) => item.tenant.id === this.configuredTenantId) ??
      inventories[0];
    const defaultInventory =
      inventories.find((item) => inventoryId(item.inventory.id) === this.selectedInventoryId) ??
      preferredInventory;

    if (!defaultInventory) {
      throw new Error('API principal did not include any inventories.');
    }

    this.selectedInventoryIdentity = defaultInventory;
    return defaultInventory;
  }

  private observeMutation(
    kind: InventoryMutationKind,
    tenantID: string,
    inventoryID: string,
    assetID?: string,
    relatedAssetIDs: readonly string[] = [],
    promotedParentId?: string
  ): void {
    this.mutationObserver.onInventoryMutation({
      kind,
      tenantId: tenantID,
      inventoryId: inventoryID,
      ...(assetID ? { assetId: assetID } : {}),
      ...(promotedParentId ? { promotedParentId } : {}),
      ...(relatedAssetIDs.length > 0 ? { relatedAssetIds: [...new Set(relatedAssetIDs)] } : {})
    });
  }

  private async loadInventoryDirectory(signal?: AbortSignal): Promise<InventoryDirectory> {
    if (!this.inventoryDirectoryPromise) {
      this.inventoryDirectoryPromise = this.fetchInventoryDirectory(signal).catch((error: unknown) => {
        this.inventoryDirectoryPromise = undefined;
        throw error;
      });
    }
    return this.inventoryDirectoryPromise;
  }

  private async refreshInventoryDirectory(signal?: AbortSignal): Promise<InventoryDirectory> {
    const pending = this.fetchInventoryDirectory(signal).catch((error: unknown) => {
      if (this.inventoryDirectoryPromise === pending) {
        this.inventoryDirectoryPromise = undefined;
      }
      throw error;
    });
    this.inventoryDirectoryPromise = pending;
    return pending;
  }

  private async fetchInventoryDirectory(signal?: AbortSignal): Promise<InventoryDirectory> {
    const tenantsPage = await this.client.listMyTenants(100, undefined, signal);
    const tenants = tenantsPage.items;
    const inventoriesByTenant = await Promise.all(
      tenants.map(async (tenant) => {
        const inventoriesPage = await this.client.listInventories(tenant.id, 100, undefined, signal);
        return inventoriesPage.items.map((inventory) => ({ tenant, inventory }));
      })
    );
    return { tenants, availableInventories: inventoriesByTenant.flat() };
  }

  private async listAllActiveInventoryAssets(
    tenantID: string,
    inventoryID: string,
    signal?: AbortSignal
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
        'id_asc',
        signal
      );
      assets.push(...page.items);
      cursor = page.pagination.nextCursor ?? undefined;
    } while (cursor);

    return assets;
  }

  private async mapAssetWithPhoto(
    inventoryName: string,
    asset: Asset,
    assets: readonly Asset[],
    options: MapAssetOptions = {}
  ): Promise<AssetSummary> {
    const photos = await this.photosForAsset(asset);
    return mapAsset(inventoryName, asset, assets, photos, options);
  }

  private async mapAssetWithPrimaryPhoto(
    inventoryName: string,
    asset: Asset,
    assets: readonly Asset[],
    options: MapAssetOptions = {}
  ): Promise<AssetSummary> {
    const photo = await this.primaryPhotoForAsset(asset);
    return mapAsset(inventoryName, asset, assets, photo ? [photo] : [], options);
  }

  private async mapAssetsWithMapPhotos(
    inventoryName: string,
    assets: readonly Asset[],
    knownAssets: readonly Asset[] = assets,
    photoAssetIds?: ReadonlySet<string>
  ): Promise<readonly AssetSummary[]> {
    return mapWithConcurrency(assets, 6, async (asset) => {
      const photo = !photoAssetIds || photoAssetIds.has(asset.id)
        ? await this.primaryMapPhotoForAsset(asset)
        : undefined;
      return mapAsset(inventoryName, asset, knownAssets, photo ? [photo] : []);
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
    options: { readonly allowAttachmentListFailure?: boolean; readonly signal?: AbortSignal } = {}
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
          cursor,
          options.signal
        );
        attachments.push(...page.items);
        cursor = page.pagination.nextCursor ?? undefined;
      } while (cursor);
    } catch (error) {
      if (options.signal?.aborted) {
        throw error;
      }
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
    input: AssetBrowsePageInput
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
        input.sort,
        input.signal
      );
      selectedAssets.push(...filterAssetsByCheckoutState(
        filterAssetsByKind(page.items, input.kind),
        input.checkoutState
      ));
      nextCursor = page.pagination.nextCursor ?? undefined;
      hasMore = page.pagination.hasMore;
      cursor = nextCursor;
    } while (selectedAssets.length < desiredMatches && hasMore);

    const knownAssets = [...selectedAssets, ...await this.loadAncestorsForAssets(selectedAssets, input.signal)];
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
    input: AssetBrowsePageInput
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
        checkoutState: input.checkoutState,
        signal: input.signal
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
    const selectedAssets = pageResults.map((item) => item.asset);
    const knownAssets = [...selectedAssets, ...await this.loadAncestorsForAssets(selectedAssets, input.signal)];
    const assets = await Promise.all(
      pageResults.map((item) =>
        this.mapAssetWithPrimaryPhoto(
          inventory.name,
          item.asset,
          knownAssets
        )
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
    input: AssetBrowsePageInput
  ): Promise<AssetBrowsePage> {
    const page = await this.client.listCheckedOutAssets(
      inventory.tenantId,
      inventory.id,
      input.limit ?? 20,
      input.cursor,
      input.signal
    );
    const selectedAssets = page.items
      .map((item) => item.asset)
      .filter((asset) => input.lifecycleState === 'all' || asset.lifecycleState === input.lifecycleState);
    const visibleAssets = filterAssetsByKind(selectedAssets, input.kind);
    const knownAssets = [...visibleAssets, ...await this.loadAncestorsForAssets(visibleAssets, input.signal)];
    const assets = await Promise.all(
      visibleAssets.map((asset) =>
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

function emptyInventorySummary(tenant: Tenant, inventory: Inventory): InventorySummary {
  return {
    id: inventoryId(inventory.id),
    tenantId: tenantId(tenant.id),
    name: inventory.name,
    role: mapAccessRole(inventory.access.relationship),
    permissions: [...inventory.access.permissions],
    description: '',
    updatedAtLabel: 'Loaded from API',
    locationCount: 0,
    locations: [],
    assets: [],
    assetTags: []
  };
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
  photos: readonly NonNullable<AssetSummary['photo']>[] = [],
  options: MapAssetOptions = {}
): AssetSummary {
  const knownAsset = assets.find((candidate) => candidate.id === asset.id);
  const shouldResolveMissingParent =
    options.resolveMissingParentFromKnownAsset &&
    (asset.parentAssetId === undefined || asset.parentAssetId === null) &&
    knownAsset?.parentAssetId;
  const parentAssetID = shouldResolveMissingParent
    ? knownAsset.parentAssetId
    : asset.parentAssetId === undefined
      ? knownAsset?.parentAssetId
      : asset.parentAssetId;
  const parent = parentAssetID
    ? assets.find((candidate) => candidate.id === parentAssetID)
    : undefined;
  const assetWithResolvedParent = parentAssetID === asset.parentAssetId
    ? asset
    : { ...asset, parentAssetId: parentAssetID ?? null };
  const ancestors = ancestorTrail(assetWithResolvedParent, assets);
  const photo = photos[0];

  return {
    id: assetId(asset.id),
    title: asset.title,
    kind: asset.kind,
    lifecycleState: asset.lifecycleState,
    parentAssetId: parentAssetID ? assetId(parentAssetID) : undefined,
    locationLabel: parent?.title ?? 'Inventory root',
    locationTrail: [inventoryName, ...ancestors.map((ancestor) => ancestor.title), asset.title].filter(isString),
    parentLocationTrail: ancestors.map((ancestor) => ({
      id: assetId(ancestor.id),
      title: ancestor.title
    })),
    description: asset.description,
    updatedAtLabel: updatedAtLabel(asset),
    hasPhoto: photo !== undefined,
    photos,
    photo,
    currentCheckout: asset.currentCheckout,
    tags: asset.tags,
    undoableOperationId: asset.undoableOperationId
  };
}

type MapAssetOptions = {
  readonly resolveMissingParentFromKnownAsset?: boolean;
};

function placementAssetsFromFullTree(
  ancestryAssets: readonly Asset[],
  selectedAssets: readonly Asset[]
): readonly Asset[] {
  const merged = new Map<string, Asset>();
  for (const asset of ancestryAssets) {
    merged.set(asset.id, asset);
  }
  for (const asset of selectedAssets) {
    if (!merged.has(asset.id)) {
      merged.set(asset.id, asset);
    }
  }
  return [...merged.values()];
}

function placementAssetsWithSelectedOverrides(
  ancestryAssets: readonly Asset[],
  selectedAssets: readonly Asset[]
): readonly Asset[] {
  const merged = new Map<string, Asset>();
  for (const asset of ancestryAssets) {
    merged.set(asset.id, asset);
  }
  for (const asset of selectedAssets) {
    merged.set(asset.id, asset);
  }
  return [...merged.values()];
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
