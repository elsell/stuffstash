import { assetId } from '../../domain/assets/AssetSummary';
import type { AssetCheckoutInput, InventorySummaryRepository } from '../home/InventorySummaryRepository';

export type AssetCheckoutAction = 'checkout' | 'return';

export type AssetCheckoutCommandInput = {
  readonly action: AssetCheckoutAction | string;
  readonly assetId: string;
  readonly details?: string;
};

export class AssetCheckoutCommand {
  constructor(private readonly inventories: InventorySummaryRepository) {}

  async execute(input: AssetCheckoutCommandInput): Promise<void> {
    const selectedAssetId = assetId(input.assetId);
    const checkoutInput: AssetCheckoutInput = { details: input.details };

    switch (input.action) {
      case 'checkout':
        if (!this.inventories.checkoutAsset) {
          throw new Error('Asset checkout is not available.');
        }
        await this.inventories.checkoutAsset(selectedAssetId, checkoutInput);
        return;
      case 'return':
        if (!this.inventories.returnAsset) {
          throw new Error('Asset return is not available.');
        }
        await this.inventories.returnAsset(selectedAssetId, checkoutInput);
        return;
      default:
        throw new Error('Unsupported asset checkout action.');
    }
  }
}
