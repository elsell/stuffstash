import { StuffStashAPIError, StuffStashClient } from '@stuff-stash/api-client';
import type { RuntimeConfig } from '$lib/runtimeConfig';
import type { TokenProvider } from '@stuff-stash/api-client';
import type { AddAssetDraft, Asset, SearchResult, WorkspaceData } from '$lib/domain/inventory';
import type { InventoryRepository } from '$lib/ports/inventoryRepository';
import type { WorkspaceObserver } from '$lib/observability/workspaceObserver';
import { mapAsset, mapCapability, mapInventory, mapPrincipal, mapSearchResult, mapTenant } from './inventoryMapper';

export class StuffStashInventoryRepository implements InventoryRepository {
  private readonly client: StuffStashClient;
  private selectedTenantId = '';
  private selectedInventoryId = '';

  constructor(
    config: RuntimeConfig,
    tokenProvider: TokenProvider,
    private readonly observer: WorkspaceObserver,
    fetchImpl?: typeof fetch
  ) {
    this.client = new StuffStashClient({
      baseUrl: config.apiBaseUrl,
      tokenProvider,
      fetch: fetchImpl
    });
  }

  async loadWorkspace(): Promise<WorkspaceData> {
    this.observer.record('workspace.load_started');
    try {
      const principal = mapPrincipal(await this.client.me());
      const tenants = (await this.client.listMyTenants()).items.map(mapTenant);
      const selectedTenant = tenants.find((tenant) => tenant.id === this.selectedTenantId) ?? tenants[0] ?? null;
      if (!selectedTenant) {
        this.observer.record('workspace.loaded', { empty: true });
        return {
          context: {
            principal,
            tenants,
            inventories: [],
            selectedTenantId: '',
            selectedInventoryId: '',
            capability: 'viewer'
          },
          assets: []
        };
      }
      return await this.loadTenantWorkspace(principal, tenants, selectedTenant.id, this.selectedInventoryId);
    } catch (error) {
      this.observer.record('workspace.load_failed');
      throw safeError(error);
    }
  }

  async createTenantWithInventory(input: { tenantName: string; inventoryName: string }): Promise<WorkspaceData> {
    const tenant = mapTenant(await this.client.createTenant(input.tenantName));
    const inventory = mapInventory(await this.client.createInventory(tenant.id, input.inventoryName));
    this.selectedTenantId = tenant.id;
    this.selectedInventoryId = inventory.id;
    return {
      context: {
        principal: mapPrincipal(await this.client.me()),
        tenants: [tenant],
        inventories: [inventory],
        selectedTenantId: tenant.id,
        selectedInventoryId: inventory.id,
        capability: mapCapability(inventory)
      },
      assets: []
    };
  }

  async selectInventory(tenantId: string, inventoryId: string): Promise<WorkspaceData> {
    const principal = mapPrincipal(await this.client.me());
    const tenants = (await this.client.listMyTenants()).items.map(mapTenant);
    return this.loadTenantWorkspace(principal, tenants, tenantId, inventoryId);
  }

  async createAsset(tenantId: string, inventoryId: string, draft: AddAssetDraft): Promise<Asset> {
    this.observer.record('workspace.asset_create_started', { kind: draft.kind });
    try {
      const asset = mapAsset(
        await this.client.createAsset(tenantId, inventoryId, {
          kind: draft.kind,
          title: draft.title,
          description: draft.description,
          parentAssetId: draft.parentAssetId
        })
      );
      this.observer.record('workspace.asset_created', { kind: asset.kind });
      return asset;
    } catch (error) {
      this.observer.record('workspace.asset_create_failed', { kind: draft.kind });
      throw safeError(error);
    }
  }

  async searchAssets(tenantId: string, query: string): Promise<SearchResult[]> {
    this.observer.record('workspace.search_started');
    try {
      const page = await this.client.searchAssets(tenantId, query);
      this.observer.record('workspace.search_completed', { resultCount: page.items.length });
      return page.items.map(mapSearchResult);
    } catch (error) {
      this.observer.record('workspace.search_failed');
      throw safeError(error);
    }
  }

  private async loadTenantWorkspace(
    principal: ReturnType<typeof mapPrincipal>,
    tenants: ReturnType<typeof mapTenant>[],
    tenantId: string,
    inventoryId: string
  ): Promise<WorkspaceData> {
    this.selectedTenantId = tenantId;
    const inventories = (await this.client.listInventories(tenantId)).items.map(mapInventory);
    const selectedInventory = inventories.find((inventory) => inventory.id === inventoryId) ?? inventories[0] ?? null;
    this.selectedInventoryId = selectedInventory?.id ?? '';
    const assets = selectedInventory
      ? (await this.client.listAssets(tenantId, selectedInventory.id, 100, undefined, 'active')).items.map(mapAsset)
      : [];
    this.observer.record('workspace.loaded', {
      tenantCount: tenants.length,
      inventoryCount: inventories.length,
      assetCount: assets.length
    });
    return {
      context: {
        principal,
        tenants,
        inventories,
        selectedTenantId: tenantId,
        selectedInventoryId: this.selectedInventoryId,
        capability: mapCapability(selectedInventory)
      },
      assets
    };
  }
}

function safeError(error: unknown): Error {
  if (error instanceof StuffStashAPIError) {
    return new Error(error.message);
  }
  if (error instanceof Error) {
    return error;
  }
  return new Error('Request failed.');
}
