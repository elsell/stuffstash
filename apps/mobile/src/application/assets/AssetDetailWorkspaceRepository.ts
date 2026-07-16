import type { AssetSummary } from '../../domain/assets/AssetSummary';
import type { InventoryId, TenantId } from '../../domain/inventories/InventorySummary';

export type AssetDetailWorkspaceSnapshot = {
  readonly tenantId: TenantId;
  readonly inventoryId: InventoryId;
  readonly permissions: readonly string[];
  readonly asset: AssetSummary;
  readonly allAssets: readonly AssetSummary[];
};

export interface AssetDetailWorkspaceRepository {
  getAssetDetailWorkspace(
    assetId: AssetSummary['id']
  ): Promise<AssetDetailWorkspaceSnapshot | undefined>;
}
