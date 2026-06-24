import {
  canCreateInventory,
  canEditInventory,
  type AddAssetDraft,
  type Asset,
  type Inventory,
  type SearchResult,
  type WorkspaceData
} from '$lib/domain/inventory';
import type { InventoryRepository, WorkspaceSeed } from '$lib/ports/inventoryRepository';
import { filterAssets } from '$lib/application/workspace';

export class SeededInventoryRepository implements InventoryRepository {
  private seed: WorkspaceSeed;
  private selectedTenantId: string;
  private selectedInventoryId: string;

  constructor(seed: WorkspaceSeed) {
    this.seed = seed;
    this.selectedTenantId = seed.tenants[0]?.id ?? '';
    this.selectedInventoryId = this.firstInventoryIdForTenant(this.selectedTenantId);
  }

  async loadWorkspace(): Promise<WorkspaceData> {
    return this.workspace();
  }

  async createTenantWithInventory(input: { tenantName: string; inventoryName: string }): Promise<WorkspaceData> {
    const tenant = {
      id: `tenant-${Date.now()}`,
      name: input.tenantName,
      access: {
        relationship: 'owner',
        permissions: ['view', 'create_inventory', 'configure']
      }
    };
    const inventory = {
      id: `inventory-${Date.now()}`,
      tenantId: tenant.id,
      name: input.inventoryName,
      access: {
        relationship: 'owner',
        permissions: ['view', 'create_asset', 'edit_asset', 'share', 'configure']
      }
    };
    this.seed = {
      ...this.seed,
      tenants: [tenant, ...this.seed.tenants],
      inventories: [inventory, ...this.seed.inventories]
    };
    this.selectedTenantId = tenant.id;
    this.selectedInventoryId = inventory.id;
    return this.workspace();
  }

  async createInventory(tenantId: string, inventoryName: string): Promise<WorkspaceData> {
    const tenant = this.seed.tenants.find((candidate) => candidate.id === tenantId);
    if (!canCreateInventory(tenant)) {
      throw new Error('You do not have permission to create inventories in this tenant.');
    }
    const inventory: Inventory = {
      id: `inventory-${Date.now()}`,
      tenantId,
      name: inventoryName,
      access: {
        relationship: 'owner',
        permissions: ['view', 'create_asset', 'edit_asset', 'share', 'configure']
      }
    };
    this.seed = {
      ...this.seed,
      inventories: [inventory, ...this.seed.inventories]
    };
    this.selectedTenantId = tenantId;
    this.selectedInventoryId = inventory.id;
    return this.workspace();
  }

  async selectInventory(tenantId: string, inventoryId: string): Promise<WorkspaceData> {
    this.selectedTenantId = tenantId;
    this.selectedInventoryId = this.seed.inventories.some(
      (inventory) => inventory.tenantId === tenantId && inventory.id === inventoryId
    )
      ? inventoryId
      : this.firstInventoryIdForTenant(tenantId);
    return this.workspace();
  }

  async selectTenant(tenantId: string): Promise<WorkspaceData> {
    this.selectedTenantId = tenantId;
    this.selectedInventoryId = this.firstInventoryIdForTenant(tenantId);
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
    const inventories = this.seed.inventories.filter((inventory) => inventory.tenantId === this.selectedTenantId);
    const selectedInventory = inventories.find((inventory) => inventory.id === this.selectedInventoryId) ?? null;
    return {
      context: {
        principal: this.seed.principal,
        tenants: this.seed.tenants,
        inventories,
        selectedTenantId: this.selectedTenantId,
        selectedInventoryId: this.selectedInventoryId,
        capability: capabilityForInventory(selectedInventory)
      },
      assets: this.workspaceAssets()
    };
  }

  private workspaceAssets(): Asset[] {
    return this.seed.assets.filter(
      (asset) => asset.tenantId === this.selectedTenantId && asset.inventoryId === this.selectedInventoryId
    );
  }

  private firstInventoryIdForTenant(tenantId: string): string {
    return this.seed.inventories.find((inventory) => inventory.tenantId === tenantId)?.id ?? '';
  }
}

function capabilityForInventory(inventory: Inventory | null): 'editor' | 'viewer' {
  if (canEditInventory(inventory)) {
    return 'editor';
  }
  return 'viewer';
}
