import type {
  AddAssetDraft,
  Asset,
  Inventory,
  Principal,
  SearchResult,
  Tenant,
  WorkspaceData
} from '$lib/domain/inventory';

export interface InventoryRepository {
  loadWorkspace(): Promise<WorkspaceData>;
  createTenantWithInventory(input: { tenantName: string; inventoryName: string }): Promise<WorkspaceData>;
  selectInventory(tenantId: string, inventoryId: string): Promise<WorkspaceData>;
  createAsset(tenantId: string, inventoryId: string, draft: AddAssetDraft): Promise<Asset>;
  searchAssets(tenantId: string, query: string): Promise<SearchResult[]>;
}

export interface WorkspaceSeed {
  principal: Principal;
  tenants: Tenant[];
  inventories: Inventory[];
  assets: Asset[];
}
