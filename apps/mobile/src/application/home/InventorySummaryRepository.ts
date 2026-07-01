import {
  InventoryId,
  InventorySummary,
  TenantContext
} from '../../domain/inventories/InventorySummary';
import type {
  AssetId,
  AssetKind,
  AssetLifecycleState,
  AssetSummary
} from '../../domain/assets/AssetSummary';
import type { LocationSummary } from '../../domain/locations/LocationSummary';

export type InventoryWorkspace = {
  readonly tenants: readonly TenantContext[];
  readonly inventories: readonly InventorySummary[];
  readonly defaultInventoryId: InventoryId;
};

export interface InventorySummaryRepository {
  getInventoryWorkspace(): Promise<InventoryWorkspace>;
  getDefaultInventorySummary(): Promise<InventorySummary>;
  selectInventory(inventoryId: InventoryId): Promise<void>;
  createAsset(input: CreateInventoryAssetInput): Promise<AssetSummary>;
  addAssetPhoto(assetId: AssetId, input: CreateInventoryAssetPhotoInput): Promise<void>;
  addInventoryAssetPhoto?(input: AddInventoryAssetPhotoInput): Promise<void>;
  archiveAsset(assetId: AssetId): Promise<void>;
  restoreAsset(assetId: AssetId): Promise<void>;
  deleteAsset(assetId: AssetId): Promise<void>;
  browseAssets(input: AssetBrowsePageInput): Promise<AssetBrowsePage>;
  searchAssets(query: string): Promise<readonly AssetSummary[]>;
  searchLocations(query: string): Promise<readonly LocationSummary[]>;
}

export interface InventoryAssetUpdateRepository {
  updateAsset(input: UpdateInventoryAssetInput): Promise<AssetSummary>;
}

export interface InventoryAssetPhotoAddRepository {
  addAssetPhoto(assetId: AssetId, input: CreateInventoryAssetPhotoInput): Promise<void>;
}

export interface InventoryAssetPhotoDeletionRepository {
  deleteAssetPhoto(assetId: AssetId, photoId: string): Promise<void>;
}

export type AssetBrowseLifecycleFilter = AssetLifecycleState | 'all';

export type AssetBrowseKindFilter = AssetKind | 'all';

export type AssetBrowseSort = 'updated_desc' | 'id_asc';

export type AssetBrowsePageInput = {
  readonly query: string;
  readonly cursor?: string;
  readonly limit?: number;
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly kind: AssetBrowseKindFilter;
  readonly sort: AssetBrowseSort;
};

export type AssetBrowsePage = {
  readonly assets: readonly AssetSummary[];
  readonly nextCursor?: string;
  readonly hasMore: boolean;
};

export type CreateInventoryAssetInput = {
  readonly kind: AssetKind;
  readonly title: string;
  readonly description: string;
  readonly parentAssetId?: AssetId;
};

export type UpdateInventoryAssetInput = {
  readonly assetId: AssetId;
  readonly title?: string;
  readonly description?: string;
  readonly parentAssetId?: AssetId | null;
};

export type CreateInventoryAssetPhotoInput = {
  readonly fileName: string;
  readonly contentType: 'image/jpeg' | 'image/png' | 'image/webp';
  readonly contentBase64?: string;
  readonly uri?: string;
  readonly sizeBytes?: number;
  readonly directUpload?: InventoryAssetPhotoDirectUpload;
};

export type InventoryAssetPhotoDirectUpload = {
  readonly uploadId: string;
  readonly attachmentId: string;
  readonly method: string;
  readonly url: string;
  readonly headers: Readonly<Record<string, string>>;
  readonly formFields: Readonly<Record<string, string>>;
  readonly expiresAt: string;
};

export type AddInventoryAssetPhotoInput = CreateInventoryAssetPhotoInput & {
  readonly tenantId: TenantContext['id'];
  readonly inventoryId: InventoryId;
  readonly assetId: AssetId;
};
