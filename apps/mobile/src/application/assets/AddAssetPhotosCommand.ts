import { assetId } from '../../domain/assets/AssetSummary';
import type {
  CreateInventoryAssetPhotoInput,
  InventoryAssetPhotoAddRepository
} from '../home/InventorySummaryRepository';

export type AddAssetPhotosCommandInput = {
  readonly assetId: string;
  readonly photos: readonly CreateInventoryAssetPhotoInput[];
  readonly onPhotoProgress?: (event: AddAssetPhotoProgressEvent) => void;
};

export type AddAssetPhotoProgressStatus = 'uploading' | 'attached' | 'failed';

export type AddAssetPhotoProgressEvent = {
  readonly index: number;
  readonly fileName: string;
  readonly status: AddAssetPhotoProgressStatus;
};

export type AddAssetPhotosCommandResult = {
  readonly attachedCount: number;
  readonly failedCount: number;
  readonly failedPhotos: readonly CreateInventoryAssetPhotoInput[];
  readonly message: string;
  readonly canRetry: boolean;
};

export class AddAssetPhotosCommand {
  constructor(private readonly inventories: InventoryAssetPhotoAddRepository) {}

  async execute(input: AddAssetPhotosCommandInput): Promise<AddAssetPhotosCommandResult> {
    if (input.photos.length === 0) {
      throw new Error('Choose at least one photo.');
    }

    const targetAssetId = assetId(input.assetId);
    let attachedCount = 0;
    let failedCount = 0;
    const failedPhotos: CreateInventoryAssetPhotoInput[] = [];
    for (const [index, photo] of input.photos.entries()) {
      input.onPhotoProgress?.({
        index,
        fileName: photo.fileName,
        status: 'uploading'
      });
      try {
        await this.inventories.addAssetPhoto(targetAssetId, photo);
        input.onPhotoProgress?.({
          index,
          fileName: photo.fileName,
          status: 'attached'
        });
        attachedCount += 1;
      } catch {
        input.onPhotoProgress?.({
          index,
          fileName: photo.fileName,
          status: 'failed'
        });
        failedCount += 1;
        failedPhotos.push(photo);
      }
    }

    return {
      attachedCount,
      failedCount,
      failedPhotos,
      message: photoUploadMessage(attachedCount, failedCount),
      canRetry: failedCount > 0
    };
  }
}

function photoUploadMessage(attachedCount: number, failedCount: number): string {
  if (failedCount === 0) {
    return `${attachedCount.toString()} ${attachedCount === 1 ? 'photo' : 'photos'} added.`;
  }
  if (attachedCount === 0) {
    return 'Photos could not be uploaded.';
  }
  return `${attachedCount.toString()} of ${(attachedCount + failedCount).toString()} photos added.`;
}
