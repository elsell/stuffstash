import { assetId } from '../../domain/assets/AssetSummary';
import type { InventoryAssetUpdateRepository } from '../home/InventorySummaryRepository';

export type UpdateAssetCommandInput = {
  readonly assetId: string;
  readonly title: string;
  readonly description: string;
  readonly tagIds?: readonly string[];
};

export type UpdateAssetCommandResult = {
  readonly id: string;
  readonly title: string;
  readonly message: string;
};

export class UpdateAssetCommand {
  constructor(private readonly inventories: InventoryAssetUpdateRepository) {}

  async execute(input: UpdateAssetCommandInput): Promise<UpdateAssetCommandResult> {
    const title = input.title.trim();
    if (title.length === 0) {
      throw new Error('Name is required.');
    }

    const updated = await this.inventories.updateAsset({
      assetId: assetId(input.assetId),
      title,
      description: input.description.trim(),
      tagIds: normalizeTagIds(input.tagIds)
    });

    return {
      id: updated.id,
      title: updated.title,
      message: `Updated ${updated.title}.`
    };
  }
}

function normalizeTagIds(tagIds: readonly string[] | undefined): readonly string[] | undefined {
  const normalized = (tagIds ?? []).map((tagId) => tagId.trim()).filter((tagId) => tagId.length > 0);
  return normalized.length > 0 ? normalized : undefined;
}
