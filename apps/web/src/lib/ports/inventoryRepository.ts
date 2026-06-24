import type {
  AddAssetDraft,
  Asset,
  AssetAttachment,
  AssetLifecycleFilter,
  Inventory,
  Principal,
  SearchResult,
  SelectedPhoto,
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
  selectAssetLifecycle(tenantId: string, inventoryId: string, lifecycleState: AssetLifecycleFilter): Promise<WorkspaceData>;
  getAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset>;
  updateAsset(tenantId: string, inventoryId: string, assetId: string, draft: UpdateAssetDraft): Promise<Asset>;
  createAsset(tenantId: string, inventoryId: string, draft: AddAssetDraft): Promise<Asset>;
  archiveAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset>;
  restoreAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset>;
  deleteAsset(tenantId: string, inventoryId: string, assetId: string): Promise<void>;
  listAssetAttachments(tenantId: string, inventoryId: string, assetId: string): Promise<AssetAttachment[]>;
  uploadAssetPhoto(tenantId: string, inventoryId: string, assetId: string, photo: SelectedPhoto): Promise<AssetAttachment>;
  archiveAssetAttachment(tenantId: string, inventoryId: string, assetId: string, attachmentId: string): Promise<AssetAttachment>;
  restoreAssetAttachment(tenantId: string, inventoryId: string, assetId: string, attachmentId: string): Promise<AssetAttachment>;
  deleteAssetAttachment(tenantId: string, inventoryId: string, assetId: string, attachmentId: string): Promise<void>;
  searchAssets(tenantId: string, query: string): Promise<SearchResult[]>;
}

export interface WorkspaceSeed {
  principal: Principal;
  tenants: Tenant[];
  inventories: Inventory[];
  assets: Asset[];
}
