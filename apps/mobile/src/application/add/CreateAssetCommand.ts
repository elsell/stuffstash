import { assetId, type AssetKind } from '../../domain/assets/AssetSummary';
import type { InventorySummaryRepository } from '../home/InventorySummaryRepository';

export type CreateAssetCommandInput = {
  readonly kind?: AssetKind;
  readonly title: string;
  readonly description: string;
  readonly parentAssetId?: string;
  readonly photos?: readonly CreateAssetPhotoInput[];
};

export type CreateAssetPhotoInput = {
  readonly fileName: string;
  readonly contentType: 'image/jpeg' | 'image/png' | 'image/webp';
  readonly contentBase64: string;
};

export type CreateAssetCommandResult = {
  readonly id: string;
  readonly title: string;
  readonly message: string;
};

export class CreateAssetCommand {
  constructor(private readonly inventories: InventorySummaryRepository) {}

  async execute(input: CreateAssetCommandInput): Promise<CreateAssetCommandResult> {
    const title = input.title.trim();
    if (title.length === 0) {
      throw new Error('Name is required.');
    }

    const asset = await this.inventories.createAsset({
      kind: input.kind ?? 'item',
      title,
      description: input.description.trim(),
      parentAssetId: input.parentAssetId ? assetId(input.parentAssetId) : undefined
    });
    let failedPhotoCount = 0;
    for (const photo of input.photos ?? []) {
      try {
        await this.inventories.addAssetPhoto(asset.id, photo);
      } catch {
        failedPhotoCount += 1;
      }
    }

    return {
      id: asset.id,
      title: asset.title,
      message:
        failedPhotoCount > 0
          ? `Saved ${asset.title}, but ${failedPhotoCount.toString()} photo upload failed.`
          : `Saved ${asset.title}.`
    };
  }
}
