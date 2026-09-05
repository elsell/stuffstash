import { assetId, type AssetSummary } from '../../domain/assets/AssetSummary';
import type { InventoryId, TenantId } from '../../domain/inventories/InventorySummary';
import type { ReadRequest } from '../shared/ReadRequest';
import { toAssetDetailViewModel, type AssetDetailViewModel } from './AssetViewModels';

export type AssetCoreSnapshot = {
  readonly tenantId: TenantId;
  readonly inventoryId: InventoryId;
  readonly permissions: readonly string[];
  readonly revision: string;
  readonly asset: AssetSummary;
};

export type AssetCoreResult = {
  readonly snapshot: AssetCoreSnapshot;
  readonly view: AssetDetailViewModel;
};

export function assetContentsIdentity(core: AssetCoreSnapshot): string {
  return [
    core.asset.kind,
    core.asset.lifecycleState,
    core.asset.parentAssetId ?? 'root'
  ].join(':');
}

export type AssetCoreRepository = {
  getAssetCore(assetId: AssetSummary['id'], request?: ReadRequest): Promise<AssetCoreSnapshot>;
};

export class AssetCoreQuery {
  constructor(private readonly assets: AssetCoreRepository) {}

  async execute(assetIdValue: string, request: ReadRequest = {}): Promise<AssetCoreResult> {
    const snapshot = await this.assets.getAssetCore(assetId(assetIdValue), request);
    return {
      snapshot,
      view: toAssetDetailViewModel(snapshot.asset, {
        tenantId: snapshot.tenantId,
        inventoryId: snapshot.inventoryId,
        canManageLifecycle: snapshot.permissions.includes('edit_asset'),
        canEditAsset: snapshot.permissions.includes('edit_asset'),
        canCreateAsset: snapshot.permissions.includes('create_asset'),
        isPlacementLoading: snapshot.asset.parentAssetId !== undefined,
        allAssets: []
      })
    };
  }
}
