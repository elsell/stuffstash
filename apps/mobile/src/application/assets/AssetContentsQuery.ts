import type { AssetSummary } from '../../domain/assets/AssetSummary';
import type { ReadRequest } from '../shared/ReadRequest';
import type { AssetCoreSnapshot } from './AssetCoreQuery';
import { toAssetDetailViewModel, type AssetDetailViewModel } from './AssetViewModels';

export type AssetContentsSnapshot = {
  readonly asset: AssetSummary;
  readonly allAssets: readonly AssetSummary[];
};

export type AssetContentsRepository = {
  getAssetContents(core: AssetCoreSnapshot, request?: ReadRequest): Promise<AssetContentsSnapshot>;
};

export class AssetContentsQuery {
  constructor(private readonly contents: AssetContentsRepository) {}

  async execute(core: AssetCoreSnapshot, request: ReadRequest = {}): Promise<AssetDetailViewModel> {
    const snapshot = await this.contents.getAssetContents(core, request);
    return toAssetDetailViewModel(snapshot.asset, {
      tenantId: core.tenantId,
      inventoryId: core.inventoryId,
      canManageLifecycle: core.permissions.includes('edit_asset'),
      canEditAsset: core.permissions.includes('edit_asset'),
      canCreateAsset: core.permissions.includes('create_asset'),
      allAssets: snapshot.allAssets
    });
  }
}
