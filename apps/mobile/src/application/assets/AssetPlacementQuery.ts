import type { AssetSummary } from '../../domain/assets/AssetSummary';
import type { ReadRequest } from '../shared/ReadRequest';
import type { AssetCoreSnapshot } from './AssetCoreQuery';
import { toAssetDetailViewModel } from './AssetViewModels';

export interface AssetPlacementRepository {
  getAssetPlacement(core: AssetCoreSnapshot, request?: ReadRequest): Promise<AssetSummary>;
}

export class AssetPlacementQuery {
  constructor(private readonly placement: AssetPlacementRepository) {}
  async execute(core: AssetCoreSnapshot, request: ReadRequest = {}) {
    const asset = await this.placement.getAssetPlacement(core, request);
    return toAssetDetailViewModel(asset, {
      tenantId: core.tenantId, inventoryId: core.inventoryId,
      canManageLifecycle: core.permissions.includes('edit_asset'),
      canEditAsset: core.permissions.includes('edit_asset'),
      canCreateAsset: core.permissions.includes('create_asset'), allAssets: []
    });
  }
}
