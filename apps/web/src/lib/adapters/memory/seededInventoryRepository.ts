import type { AddAssetDraft, Asset, SearchResult, WorkspaceData } from '$lib/domain/inventory';
import type { InventoryRepository, WorkspaceSeed } from '$lib/ports/inventoryRepository';
import { filterAssets } from '$lib/application/workspace';

export class SeededInventoryRepository implements InventoryRepository {
  private seed: WorkspaceSeed;
  private selectedTenantId: string;
  private selectedInventoryId: string;

  constructor(seed: WorkspaceSeed) {
    this.seed = seed;
    this.selectedTenantId = seed.tenants[0]?.id ?? '';
    this.selectedInventoryId = seed.inventories[0]?.id ?? '';
  }

  async loadWorkspace(): Promise<WorkspaceData> {
    return this.workspace();
  }

  async createTenantWithInventory(input: { tenantName: string; inventoryName: string }): Promise<WorkspaceData> {
    const tenant = { id: `tenant-${Date.now()}`, name: input.tenantName };
    const inventory = { id: `inventory-${Date.now()}`, tenantId: tenant.id, name: input.inventoryName };
    this.seed = {
      ...this.seed,
      tenants: [tenant, ...this.seed.tenants],
      inventories: [inventory, ...this.seed.inventories]
    };
    this.selectedTenantId = tenant.id;
    this.selectedInventoryId = inventory.id;
    return this.workspace();
  }

  async selectInventory(tenantId: string, inventoryId: string): Promise<WorkspaceData> {
    this.selectedTenantId = tenantId;
    this.selectedInventoryId = inventoryId;
    return this.workspace();
  }

  async createAsset(tenantId: string, inventoryId: string, draft: AddAssetDraft): Promise<Asset> {
    const asset: Asset = {
      id: `asset-${Date.now()}`,
      tenantId,
      inventoryId,
      kind: draft.kind,
      title: draft.title,
      description: draft.description,
      parentAssetId: draft.parentAssetId,
      lifecycleState: 'active',
      photo: draft.photos[0]
        ? {
            id: draft.photos[0].id,
            url: draft.photos[0].previewUrl,
            alt: draft.title
          }
        : undefined,
      updatedAt: new Date().toISOString()
    };
    this.seed = { ...this.seed, assets: [asset, ...this.seed.assets] };
    return asset;
  }

  async searchAssets(_tenantId: string, query: string): Promise<SearchResult[]> {
    const inventory = this.seed.inventories.find((candidate) => candidate.id === this.selectedInventoryId);
    return filterAssets(this.workspaceAssets(), query).map((asset) => ({
      type: 'asset',
      asset,
      inventory: {
        id: inventory?.id ?? this.selectedInventoryId,
        name: inventory?.name ?? 'Inventory'
      },
      matches: [{ field: 'title', value: asset.title }]
    }));
  }

  private workspace(): WorkspaceData {
    return {
      context: {
        principal: this.seed.principal,
        tenants: this.seed.tenants,
        inventories: this.seed.inventories.filter((inventory) => inventory.tenantId === this.selectedTenantId),
        selectedTenantId: this.selectedTenantId,
        selectedInventoryId: this.selectedInventoryId,
        capability: 'editor'
      },
      assets: this.workspaceAssets()
    };
  }

  private workspaceAssets(): Asset[] {
    return this.seed.assets.filter((asset) => asset.inventoryId === this.selectedInventoryId);
  }
}
