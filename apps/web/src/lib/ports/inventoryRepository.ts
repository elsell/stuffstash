import type {
  AddAssetDraft,
  Asset,
  AssetAttachment,
  AssetLifecycleFilter,
  CustomAssetType,
  CustomFieldDefinition,
  Inventory,
  Principal,
  SearchRequest,
  SearchResult,
  SelectedAttachment,
  SelectedPhoto,
  LegacyHomeboxImportRequest,
  ImportApplyResult,
  ImportPreview,
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
  uploadAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    attachment: SelectedAttachment
  ): Promise<AssetAttachment>;
  uploadAssetPhoto(tenantId: string, inventoryId: string, assetId: string, photo: SelectedPhoto): Promise<AssetAttachment>;
  archiveAssetAttachment(tenantId: string, inventoryId: string, assetId: string, attachmentId: string): Promise<AssetAttachment>;
  restoreAssetAttachment(tenantId: string, inventoryId: string, assetId: string, attachmentId: string): Promise<AssetAttachment>;
  deleteAssetAttachment(tenantId: string, inventoryId: string, assetId: string, attachmentId: string): Promise<void>;
  searchAssets(request: SearchRequest): Promise<SearchResult[]>;
  previewLegacyHomeboxImport(tenantId: string, inventoryId: string, input: LegacyHomeboxImportRequest): Promise<ImportPreview>;
  applyLegacyHomeboxImport(tenantId: string, inventoryId: string, input: LegacyHomeboxImportRequest): Promise<ImportApplyResult>;
}

export interface WorkspaceSeed {
  principal: Principal;
  tenants: Tenant[];
  inventories: Inventory[];
  assets: Asset[];
  customAssetTypes: CustomAssetType[];
  customFieldDefinitions: CustomFieldDefinition[];
}
