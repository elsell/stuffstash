import { assertReadActive } from '../../application/shared/ReadRequest';
import { ReadPageGuard } from '../shared/ReadPageGuard';
import type {
  Asset,
  AssetSearchResult,
  AssetTag,
  CheckedOutAsset,
  Inventory,
  Tenant
} from '@stuff-stash/api-client';
import type { AddAssetContext } from '../../application/add/AddAssetContextQuery';
import type {
  AssetDetailWorkspaceRepository,
  AssetDetailWorkspaceSnapshot
} from '../../application/assets/AssetDetailWorkspaceRepository';
import type { InventoryAssetsSnapshot } from '../../application/assets/InventoryAssetsQuery';
import type { InventoryMapAssetRepository } from '../../application/assets/InventoryMapQuery';
import {
  ignoreInventoryMutations,
  type InventoryMutationObserver
} from '../../application/home/InventoryMutationObserver';
import {
  HomeDashboardSnapshot,
  HomeDashboardSnapshotRepository,
  InventorySummaryRepository,
  InventoryWorkspace
} from '../../application/home/InventorySummaryRepository';
import type { LocationAssetsSnapshot } from '../../application/locations/LocationAssetsQuery';
import type { LocationsSnapshot } from '../../application/locations/LocationsQuery';
import type { ReadRequest } from '../../application/shared/ReadRequest';
import { assetId, AssetSummary, type AssetTagSummary } from '../../domain/assets/AssetSummary';
import {
  InventoryId,
  inventoryId,
  InventorySummary,
  tenantId
} from '../../domain/inventories/InventorySummary';
import type { LocationSummary } from '../../domain/locations/LocationSummary';
import {
  type DirectUploadTargetPolicy
} from '../uploads/DirectUploadPolicy';
import { ExpoDirectUploadTransport, type DirectUploadTransport } from '../uploads/ExpoDirectUploadTransport';
import { ApiInventoryAssetBrowse } from './ApiInventoryAssetBrowse';
import { ApiInventoryAssetCommands } from './ApiInventoryAssetCommands';
import { ApiInventoryAssetDetailReads } from './ApiInventoryAssetDetailReads';
import { ApiInventoryAssetPhotos } from './ApiInventoryAssetPhotos';
import { ApiInventoryAssetTraversal } from './ApiInventoryAssetTraversal';
import { ApiInventoryDirectory } from './ApiInventoryDirectory';
import type { InventoryApiClient } from './InventoryApiClient';
import { ancestorTrail, emptyInventorySummary, mapAccessRole, mapAssetTag, mapLocation, mapTenant, placementAssetsFromFullTree, selectAssetDetailWorkspace, summaryToApiAsset } from './InventoryAssetMapping';

const inventoryAssetPageSize = 100;

type LoadedInventoryWorkspace = {
  readonly workspace: InventoryWorkspace;
  readonly defaultPlacementAssets: readonly Asset[];
};

type MappedInventory = {
  readonly summary: InventorySummary;
  readonly placementAssets: readonly Asset[];
};

export class ApiInventorySummaryRepository implements InventorySummaryRepository, InventoryMapAssetRepository, HomeDashboardSnapshotRepository, AssetDetailWorkspaceRepository {
  private readonly directory: ApiInventoryDirectory;
  private readonly photos: ApiInventoryAssetPhotos;
  private readonly traversal: ApiInventoryAssetTraversal;
  private readonly commands: ApiInventoryAssetCommands;
  private readonly detail: ApiInventoryAssetDetailReads;
  private readonly browse: ApiInventoryAssetBrowse;
  private readonly directUploadTransport: DirectUploadTransport;

  constructor(
    private readonly client: InventoryApiClient,
    configuredTenantId: string,
    directUploadTransport?: DirectUploadTransport,
    private readonly sessionScopeId = 'mobile-composition',
    directUploadPolicy: DirectUploadTargetPolicy = {},
    mutationObserver: InventoryMutationObserver = ignoreInventoryMutations
  ) {
    this.directory = new ApiInventoryDirectory(client, configuredTenantId);
    this.directUploadTransport = directUploadTransport ?? new ExpoDirectUploadTransport(directUploadPolicy);
    this.photos = new ApiInventoryAssetPhotos(client);
    this.traversal = new ApiInventoryAssetTraversal(client);
    this.commands = new ApiInventoryAssetCommands(client, this.directory, this.traversal, this.directUploadTransport, directUploadPolicy, mutationObserver);
    this.detail = new ApiInventoryAssetDetailReads(client, this.directory, this.traversal, this.photos);
    this.browse = new ApiInventoryAssetBrowse(client, this.directory, this.traversal, this.photos);
  }


  createAsset = (...args: Parameters<ApiInventoryAssetCommands['createAsset']>) => this.commands.createAsset(...args);
  createAssetTag = (...args: Parameters<ApiInventoryAssetCommands['createAssetTag']>) => this.commands.createAssetTag(...args);
  updateAsset = (...args: Parameters<ApiInventoryAssetCommands['updateAsset']>) => this.commands.updateAsset(...args);
  addAssetPhoto = (...args: Parameters<ApiInventoryAssetCommands['addAssetPhoto']>) => this.commands.addAssetPhoto(...args);
  addInventoryAssetPhoto = (...args: Parameters<ApiInventoryAssetCommands['addInventoryAssetPhoto']>) => this.commands.addInventoryAssetPhoto(...args);
  deleteAssetPhoto = (...args: Parameters<ApiInventoryAssetCommands['deleteAssetPhoto']>) => this.commands.deleteAssetPhoto(...args);
  archiveAsset = (...args: Parameters<ApiInventoryAssetCommands['archiveAsset']>) => this.commands.archiveAsset(...args);
  restoreAsset = (...args: Parameters<ApiInventoryAssetCommands['restoreAsset']>) => this.commands.restoreAsset(...args);
  deleteAsset = (...args: Parameters<ApiInventoryAssetCommands['deleteAsset']>) => this.commands.deleteAsset(...args);
  checkoutAsset = (...args: Parameters<ApiInventoryAssetCommands['checkoutAsset']>) => this.commands.checkoutAsset(...args);
  returnAsset = (...args: Parameters<ApiInventoryAssetCommands['returnAsset']>) => this.commands.returnAsset(...args);
  updateReturnedCheckoutDetails = (...args: Parameters<ApiInventoryAssetCommands['updateReturnedCheckoutDetails']>) => this.commands.updateReturnedCheckoutDetails(...args);
  undoInventoryOperation = (...args: Parameters<ApiInventoryAssetCommands['undoInventoryOperation']>) => this.commands.undoInventoryOperation(...args);
  getAssetCore = (...args: Parameters<ApiInventoryAssetDetailReads['getAssetCore']>) => this.detail.getAssetCore(...args);
  getAssetPlacement = (...args: Parameters<ApiInventoryAssetDetailReads['getAssetPlacement']>) => this.detail.getAssetPlacement(...args);
  getAssetContents = (...args: Parameters<ApiInventoryAssetDetailReads['getAssetContents']>) => this.detail.getAssetContents(...args);
  getAssetPhotos = (...args: Parameters<ApiInventoryAssetDetailReads['getAssetPhotos']>) => this.detail.getAssetPhotos(...args);
  getAssetDetail = (...args: Parameters<ApiInventoryAssetDetailReads['getAssetDetail']>) => this.detail.getAssetDetail(...args);
  browseAssets = (...args: Parameters<ApiInventoryAssetBrowse['browseAssets']>) => this.browse.browseAssets(...args);

  async getInventoryWorkspace(): Promise<InventoryWorkspace> {
    return (await this.loadInventoryWorkspace()).workspace;
  }

  async getCurrentTenantId(request: ReadRequest = {}): Promise<string> {
    return (await this.getSelectedInventoryIdentity(request.signal)).tenant.id;
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
    const defaultInventory = await this.directory.selected();

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
    const ancestorAssets = await this.traversal.loadAncestorsForAssets(visibleAssets, request.signal);
    const ancestryAssets = placementAssetsFromFullTree(
      [...ancestorAssets, ...visibleAssets],
      visibleAssets
    );
    const mappedCheckedOutAssets = await Promise.all(
      checkedOutAssets.map((item) => this.photos.mapAssetWithPrimaryPhoto(selected.inventory.name, item.asset, ancestryAssets))
    );
    const checkedOutByAssetId = new Map(mappedCheckedOutAssets.map((asset) => [asset.id, asset]));
    const mappedRecentAssets = await Promise.all(
      recentAssets.map((asset) => checkedOutByAssetId.get(assetId(asset.id)) ??
        this.photos.mapAssetWithPrimaryPhoto(selected.inventory.name, asset, ancestryAssets))
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
      this.traversal.listAllActiveInventoryAssets(selected.tenant.id, selected.inventory.id, request.signal)
    ]);
    const ancestryAssets = placementAssetsFromFullTree(activeAssets, recentAssets);
    return {
      inventoryName: selected.inventory.name,
      assets: await Promise.all(
        recentAssets.map((asset) => this.photos.mapAssetWithPrimaryPhoto(
          selected.inventory.name,
          asset,
          ancestryAssets
        ))
      )
    };
  }

  async getLocationsSnapshot(request: ReadRequest = {}): Promise<LocationsSnapshot> {
    const selected = await this.getSelectedInventoryIdentity(request.signal);
    const assets = await this.traversal.listAllActiveInventoryAssets(selected.tenant.id, selected.inventory.id, request.signal);
    const locations = await Promise.all(
      assets
        .filter((asset) => asset.kind === 'location')
        .map(async (location) => mapLocation(
          location,
          assets,
          await this.photos.primaryPhotoForAsset(location)
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
    const assets = await this.traversal.listAllActiveInventoryAssets(selected.tenant.id, selected.inventory.id, request.signal);
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
        this.photos.mapAssetWithPrimaryPhoto(selected.inventory.name, asset, assets)
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

  async getVoiceInventoryContext(request: ReadRequest = {}) {
    const selected = await this.getSelectedInventoryIdentity(request.signal);
    return { tenantId: tenantId(selected.tenant.id), inventoryId: inventoryId(selected.inventory.id), tenantName: selected.tenant.name, inventoryName: selected.inventory.name };
  }

  selectInventory(selectedInventoryId: InventoryId): Promise<void> {
    return this.directory.select(selectedInventoryId);
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
    const activeAssets = await this.traversal.listAllActiveInventoryAssets(tenantId(inventory.tenant.id), inventory.inventory.id, request.signal);
    const assets = await this.photos.mapAssetsWithMapPhotos(inventory.inventory.name, activeAssets);

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
      const ancestors = await this.traversal.loadAssetAncestors(selectedAsset, request.signal);
      const mapped = await this.photos.mapAssetsWithMapPhotos(
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
    const activeAssets = await this.traversal.listAllActiveInventoryAssets(
      tenantId(inventory.tenant.id),
      inventory.inventory.id,
      request.signal
    );
    const selectedFromTraversal = activeAssets.find((asset) => asset.id === selectedAssetId) ?? selectedAsset;
    const workspace = selectAssetDetailWorkspace(selectedFromTraversal, activeAssets);
    const assets = await this.photos.mapAssetsWithMapPhotos(
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
        this.photos.mapAssetWithPrimaryPhoto(
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
    const placementAssets = await this.traversal.listAllActiveInventoryAssets(selected.tenant.id, selected.inventory.id);

    const page = await this.client.searchAssets(inventory.tenantId, query, { limit: 50 });
    const locationAssets = page.items
      .filter((item) => item.inventory.id === inventory.id && item.asset.kind === 'location')
      .map((item) => item.asset);
    const knownAssets = placementAssets;

    return Promise.all(
      locationAssets.map(async (location) =>
        mapLocation(location, knownAssets, await this.photos.primaryPhotoForAsset(location))
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
      ? await this.traversal.listAllActiveInventoryAssets(tenant.id, inventory.id)
      : assets;
    const assetTags = await this.listAllInventoryTags(tenant.id, inventory.id);
    const locations = await Promise.all(
      locationSourceAssets
        .filter((asset) => asset.kind === 'location')
        .map(async (location) => mapLocation(location, locationSourceAssets, await this.photos.primaryPhotoForAsset(location)))
    );

    const ancestryAssets = placementAssetsFromFullTree(locationSourceAssets, assets);
    const mappedAssets = await Promise.all(
      assets.map((asset) => this.photos.mapAssetWithPrimaryPhoto(inventory.name, asset, ancestryAssets))
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
    const guard = new ReadPageGuard();

    do {
      assertReadActive(signal);
      const page = await this.client.listCheckedOutAssets(tenantID, inventoryID, 10, cursor, signal);
      assertReadActive(signal);
      checkedOutAssets.push(...page.items);
      cursor = guard.accept(page.pagination.nextCursor, page.pagination.hasMore);
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
    const guard = new ReadPageGuard();

    do {
      assertReadActive(signal);
      const page = await this.client.listAssetTags(tenantID, inventoryID, 100, cursor, signal);
      assertReadActive(signal);
      tags.push(...page.items);
      cursor = guard.accept(page.pagination.nextCursor, page.pagination.hasMore);
    } while (cursor);

    return tags;
  }

  private getSelectedInventoryIdentity(signal?: AbortSignal) { return this.directory.selected(signal); }

  private loadInventoryDirectory(signal?: AbortSignal) { return this.directory.load(signal); }

  private async searchSelectedInventoryAssets(
    tenantID: string,
    inventoryID: string,
    query: string,
    desiredMatches: number
  ) {
    const matches: AssetSearchResult[] = [];
    let cursor: string | undefined;
    const guard = new ReadPageGuard();

    do {
      const page = await this.client.searchAssets(tenantID, query, {
        limit: 50,
        cursor
      });
      matches.push(...page.items.filter((item) => item.inventory.id === inventoryID));
      cursor = guard.accept(page.pagination.nextCursor, page.pagination.hasMore);
    } while (matches.length < desiredMatches && cursor);

    return matches.slice(0, desiredMatches);
  }


}
