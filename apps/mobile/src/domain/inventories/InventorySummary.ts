import {
  AssetSummary,
  AssetTagSummary,
  countActiveAssets,
  countAssetsWithPhotos
} from '../assets/AssetSummary';
import type { LocationSummary } from '../locations/LocationSummary';

export type AccessRole = 'owner' | 'editor' | 'viewer';

export type InventoryPermission = 'view' | 'create_asset' | 'edit_asset' | 'share' | 'configure' | string;

export type InventoryId = string & { readonly __brand: 'InventoryId' };

export type TenantId = string & { readonly __brand: 'TenantId' };

export type TenantContext = {
  readonly id: TenantId;
  readonly name: string;
};

export type InventorySummary = {
  readonly id: InventoryId;
  readonly tenantId: TenantId;
  readonly name: string;
  readonly role: AccessRole;
  readonly permissions: readonly InventoryPermission[];
  readonly description: string;
  readonly updatedAtLabel: string;
  readonly locationCount: number;
  readonly locations: readonly LocationSummary[];
  readonly assets: readonly AssetSummary[];
  readonly assetTags?: readonly AssetTagSummary[];
};

export type InventoryOverview = {
  readonly tenantName: string;
  readonly inventoryName: string;
  readonly inventories: readonly InventorySummary[];
  readonly totalAssets: number;
  readonly activeAssets: number;
  readonly photoReadyAssets: number;
  readonly locationCount: number;
  readonly locations: readonly LocationSummary[];
  readonly recentlyUpdatedAssets: readonly AssetSummary[];
};

export function inventoryId(value: string): InventoryId {
  const trimmed = value.trim();
  if (trimmed.length === 0) {
    throw new Error('Inventory ID must not be empty.');
  }

  return trimmed as InventoryId;
}

export function tenantId(value: string): TenantId {
  const trimmed = value.trim();
  if (trimmed.length === 0) {
    throw new Error('Tenant ID must not be empty.');
  }

  return trimmed as TenantId;
}

export function createInventoryOverview(
  tenant: TenantContext,
  inventory: InventorySummary,
  inventories: readonly InventorySummary[]
): InventoryOverview {
  return {
    tenantName: tenant.name,
    inventoryName: inventory.name,
    inventories,
    totalAssets: inventory.assets.length,
    activeAssets: countActiveAssets(inventory.assets),
    photoReadyAssets: countAssetsWithPhotos(inventory.assets),
    locationCount: inventory.locationCount,
    locations: inventory.locations,
    recentlyUpdatedAssets: inventory.assets.slice(0, 4)
  };
}
