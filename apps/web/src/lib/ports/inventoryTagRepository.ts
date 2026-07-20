import type { AssetTagDraft, ManagedAssetTag } from '$lib/domain/inventory';
import type { CustomizationPage } from './inventoryCustomizationRepository';

export interface AssetTagUpdate {
  displayName: string;
  color?: string;
}

export interface InventoryTagRepository {
  listManagedAssetTags(tenantId: string, inventoryId: string, cursor?: string): Promise<CustomizationPage<ManagedAssetTag>>;
  createManagedAssetTag(tenantId: string, inventoryId: string, draft: AssetTagDraft): Promise<ManagedAssetTag>;
  updateManagedAssetTag(tenantId: string, inventoryId: string, tagId: string, update: AssetTagUpdate): Promise<ManagedAssetTag>;
  archiveManagedAssetTag(tenantId: string, inventoryId: string, tagId: string): Promise<ManagedAssetTag>;
}
