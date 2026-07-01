import { assetId } from '../../domain/assets/AssetSummary';
import type { InventoryAssetUpdateRepository } from '../home/InventorySummaryRepository';

export type MoveAssetCommandInput = {
  readonly assetId: string;
  readonly parentAssetId?: string;
};

export type MoveAssetCommandResult = {
  readonly id: string;
  readonly title: string;
  readonly message: string;
};

export class MoveAssetCommand {
  constructor(private readonly inventories: InventoryAssetUpdateRepository) {}

  async execute(input: MoveAssetCommandInput): Promise<MoveAssetCommandResult> {
    const targetAssetId = assetId(input.assetId);
    const parentAssetId = input.parentAssetId ? assetId(input.parentAssetId) : null;
    if (parentAssetId === targetAssetId) {
      throw new Error('An asset cannot be moved into itself.');
    }

    const updated = await this.inventories.updateAsset({
      assetId: targetAssetId,
      parentAssetId
    });

    return {
      id: updated.id,
      title: updated.title,
      message: parentAssetId ? `Moved ${updated.title}.` : `Moved ${updated.title} to No parent.`
    };
  }
}
