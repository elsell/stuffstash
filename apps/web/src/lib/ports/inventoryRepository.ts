import type {
  AddAssetDraft,
  Asset,
  Inventory,
  Principal,
  SearchResult,
  Tenant,
  UpdateAssetDraft,
  WorkspaceData
} from '$lib/domain/inventory';

export interface InventoryRepository {
  loadWorkspace(): Promise<WorkspaceData>;
  createTenantWithInventory(input: { tenantName: string; inventoryName: string }): Promise<WorkspaceData>;
  createInventory(tenantId: string, inventoryName: string): Promise<WorkspaceData>;
  selectTenant(tenantId: string): Promise<WorkspaceData>;
  selectInventory(tenantId: string, inventoryId: string): Promise<WorkspaceData>;
  getAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset>;
  updateAsset(tenantId: string, inventoryId: string, assetId: string, draft: UpdateAssetDraft): Promise<Asset>;
  createAsset(tenantId: string, inventoryId: string, draft: AddAssetDraft): Promise<Asset>;
  searchAssets(tenantId: string, query: string): Promise<SearchResult[]>;
}

export interface WorkspaceSeed {
  principal: Principal;
  tenants: Tenant[];
  inventories: Inventory[];
  assets: Asset[];
}
