import { assetId } from '../../domain/assets/AssetSummary';
import type { InventorySummaryRepository } from '../home/InventorySummaryRepository';

export type AssetLifecycleAction = 'archive' | 'restore' | 'delete';

export type AssetLifecycleCommandInput = {
  readonly action: AssetLifecycleAction | string;
  readonly assetId: string;
};

export class AssetLifecycleCommand {
  constructor(private readonly inventories: InventorySummaryRepository) {}

  async execute(input: AssetLifecycleCommandInput): Promise<void> {
    const selectedAssetId = assetId(input.assetId);

    switch (input.action) {
      case 'archive':
        await this.inventories.archiveAsset(selectedAssetId);
        return;
      case 'restore':
        await this.inventories.restoreAsset(selectedAssetId);
        return;
      case 'delete':
        await this.inventories.deleteAsset(selectedAssetId);
        return;
      default:
        throw new Error('Unsupported asset lifecycle action.');
    }
  }
}
