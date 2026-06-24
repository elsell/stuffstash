import {
  InventoryId,
  InventorySummary,
  TenantContext
} from '../../domain/inventories/InventorySummary';
import type {
  AssetId,
  AssetKind,
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
  searchAssets(query: string): Promise<readonly AssetSummary[]>;
  searchLocations(query: string): Promise<readonly LocationSummary[]>;
}

export type CreateInventoryAssetInput = {
  readonly kind: AssetKind;
  readonly title: string;
  readonly description: string;
  readonly parentAssetId?: AssetId;
};

export type CreateInventoryAssetPhotoInput = {
  readonly fileName: string;
  readonly contentType: 'image/jpeg' | 'image/png' | 'image/webp';
  readonly contentBase64: string;
};
