import { assetId } from '../../domain/assets/AssetSummary';
import type { InventoryAssetPhotoDeletionRepository } from '../home/InventorySummaryRepository';

export type DeleteAssetPhotoCommandInput = {
  readonly assetId: string;
  readonly photoId: string;
};

export type DeleteAssetPhotoCommandResult = {
  readonly message: string;
};

export class DeleteAssetPhotoCommand {
  constructor(private readonly inventories: InventoryAssetPhotoDeletionRepository) {}

  async execute(input: DeleteAssetPhotoCommandInput): Promise<DeleteAssetPhotoCommandResult> {
    const photoId = input.photoId.trim();
    if (photoId.length === 0) {
      throw new Error('Photo ID is required.');
    }

    await this.inventories.deleteAssetPhoto(assetId(input.assetId), photoId);

    return {
      message: 'Photo removed.'
    };
  }
}
