import type {
  AddAssetDraft,
  Asset,
  AssetAttachment,
  AssetCheckout,
  AssetCheckoutDraft,
  AssetLifecycleFilter,
  AssetTag,
  AssetTagDraft,
  CheckedOutAsset,
  CustomAssetType,
  CustomFieldDefinition,
  Inventory,
  Principal,
  SearchRequest,
  SearchResult,
  SelectedAttachment,
  SelectedPhoto,
  ImportSourceRequest,
  ImportJob,
  ImportJobCancellationMode,
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
  createAssetTag(tenantId: string, inventoryId: string, draft: AssetTagDraft): Promise<AssetTag>;
  archiveAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset>;
  restoreAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset>;
  deleteAsset(tenantId: string, inventoryId: string, assetId: string): Promise<void>;
  checkoutAsset(tenantId: string, inventoryId: string, assetId: string, draft: AssetCheckoutDraft): Promise<AssetCheckout>;
  returnAsset(tenantId: string, inventoryId: string, assetId: string, draft: AssetCheckoutDraft): Promise<AssetCheckout>;
  listAssetCheckoutHistory(tenantId: string, inventoryId: string, assetId: string): Promise<AssetCheckout[]>;
  listCheckedOutAssets(tenantId: string, inventoryId: string): Promise<CheckedOutAsset[]>;
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
  listImportJobs(tenantId: string, inventoryId: string): Promise<ImportJob[]>;
  previewImportJob(tenantId: string, inventoryId: string, input: ImportSourceRequest): Promise<ImportJob>;
  getImportJob(tenantId: string, inventoryId: string, jobId: string): Promise<ImportJob>;
  startImportJob(tenantId: string, inventoryId: string, jobId: string, input: ImportSourceRequest): Promise<ImportJob>;
  cancelImportJob(
    tenantId: string,
    inventoryId: string,
    jobId: string,
    mode: ImportJobCancellationMode
  ): Promise<ImportJob>;
  removeImportJobFromHistory(tenantId: string, inventoryId: string, jobId: string): Promise<void>;
}

export interface WorkspaceSeed {
  principal: Principal;
  tenants: Tenant[];
  inventories: Inventory[];
  assets: Asset[];
  customAssetTypes: CustomAssetType[];
  customFieldDefinitions: CustomFieldDefinition[];
  assetTags?: AssetTag[];
}
