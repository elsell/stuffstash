import { assetId } from '../../domain/assets/AssetSummary';
import type {
  AssetCheckoutInput,
  AssetCheckoutResult,
  InventorySummaryRepository
} from '../home/InventorySummaryRepository';

export type AssetCheckoutAction = 'checkout' | 'return';

export type AssetCheckoutCommandInput = {
  readonly action: AssetCheckoutAction | string;
  readonly assetId: string;
  readonly details?: string;
};

export type UpdateReturnedCheckoutDetailsCommandInput = {
  readonly assetId: string;
  readonly checkoutId: string;
  readonly details?: string;
};

export type UndoCheckoutOperationCommandInput = {
  readonly operationId: string;
};

export class AssetCheckoutCommand {
  constructor(private readonly inventories: InventorySummaryRepository) {}

  async execute(input: AssetCheckoutCommandInput): Promise<AssetCheckoutResult> {
    const selectedAssetId = assetId(input.assetId);
    const checkoutInput: AssetCheckoutInput = { details: input.details };

    switch (input.action) {
      case 'checkout':
        if (!this.inventories.checkoutAsset) {
          throw new Error('Asset checkout is not available.');
        }
        return await this.inventories.checkoutAsset(selectedAssetId, checkoutInput);
      case 'return':
        if (!this.inventories.returnAsset) {
          throw new Error('Asset return is not available.');
        }
        return await this.inventories.returnAsset(selectedAssetId, checkoutInput);
      default:
        throw new Error('Unsupported asset checkout action.');
    }
  }

  async updateReturnedCheckoutDetails(input: UpdateReturnedCheckoutDetailsCommandInput): Promise<AssetCheckoutResult> {
    if (!this.inventories.updateReturnedCheckoutDetails) {
      throw new Error('Asset return details are not available.');
    }
    return await this.inventories.updateReturnedCheckoutDetails(assetId(input.assetId), input.checkoutId, { details: input.details });
  }

  async undoOperation(input: UndoCheckoutOperationCommandInput): Promise<void> {
    if (!this.inventories.undoInventoryOperation) {
      throw new Error('Undo is not available.');
    }
    await this.inventories.undoInventoryOperation(input.operationId);
  }
}
