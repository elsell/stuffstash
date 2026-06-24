import { StuffStashAPIError, StuffStashClient } from '@stuff-stash/api-client';
import type { RuntimeConfig } from '$lib/runtimeConfig';
import type { TokenProvider } from '@stuff-stash/api-client';
import type { AddAssetDraft, Asset, SearchResult, WorkspaceData } from '$lib/domain/inventory';
import type { InventoryRepository } from '$lib/ports/inventoryRepository';
import type { WorkspaceObserver } from '$lib/observability/workspaceObserver';
import { mapAsset, mapInventory, mapPrincipal, mapSearchResult, mapTenant } from './inventoryMapper';

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
      if (!this.selectedTenantId) {
        this.observer.record('workspace.loaded', { empty: true });
        return {
          context: {
            principal,
            tenants: [],
            inventories: [],
            selectedTenantId: '',
            selectedInventoryId: '',
            capability: 'editor'
          },
          assets: []
        };
      }
      return await this.selectInventory(this.selectedTenantId, this.selectedInventoryId);
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
        capability: 'editor'
      },
      assets: []
    };
  }

  async selectInventory(tenantId: string, inventoryId: string): Promise<WorkspaceData> {
    this.selectedTenantId = tenantId;
    this.selectedInventoryId = inventoryId;
    const [principal, inventories, assets] = await Promise.all([
      this.client.me(),
      this.client.listInventories(tenantId),
      inventoryId ? this.client.listAssets(tenantId, inventoryId, 100, undefined, 'active') : Promise.resolve({ items: [] })
    ]);
    const selectedInventory = inventories.items.find((inventory) => inventory.id === inventoryId) ?? inventories.items[0];
    this.selectedInventoryId = selectedInventory?.id ?? '';
    this.observer.record('workspace.loaded', {
      inventoryCount: inventories.items.length,
      assetCount: assets.items.length
    });
    return {
      context: {
        principal: mapPrincipal(principal),
        tenants: [{ id: tenantId, name: 'Home' }],
        inventories: inventories.items.map(mapInventory),
        selectedTenantId: tenantId,
        selectedInventoryId: this.selectedInventoryId,
        capability: 'editor'
      },
      assets: assets.items.map(mapAsset)
    };
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
