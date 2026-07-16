import { assetId } from '../../domain/assets/AssetSummary';
import type {
  InventoryAssetUpdateRepository
} from '../home/InventorySummaryRepository';
import type { ActiveAssetTagReference, AssetTagCreateRepository, CreateAssetTagDraft } from './AssetTagDraftResolution';
import { createPendingAssetTags, reconcilePendingAssetTagDrafts } from './AssetTagDraftResolution';

export type UpdateAssetCommandInput = {
  readonly assetId: string;
  readonly title: string;
  readonly description: string;
  readonly tagIds?: readonly string[];
  readonly newTags?: readonly CreateAssetTagDraft[];
  readonly activeTags?: readonly ActiveAssetTagReference[];
};

export type UpdateAssetCommandResult = {
  readonly id: string;
  readonly title: string;
  readonly message: string;
  readonly undoableOperationId?: string;
};

export class UpdateAssetCommand {
  constructor(private readonly inventories: InventoryAssetUpdateRepository & AssetTagCreateRepository) {}

  async execute(input: UpdateAssetCommandInput): Promise<UpdateAssetCommandResult> {
    const title = input.title.trim();
    if (title.length === 0) {
      throw new Error('Name is required.');
    }

    const reconciledTags = reconcilePendingAssetTagDrafts({
      selectedTagIds: input.tagIds ?? [],
      pendingTags: input.newTags ?? [],
      activeTags: input.activeTags ?? []
    });
    const createdTagIds = await createPendingAssetTags(this.inventories, reconciledTags.pendingTags);
    const baseTagIds = input.tagIds === undefined && reconciledTags.tagIds.length === 0
      ? undefined
      : reconciledTags.tagIds;
    const tagIds = baseTagIds === undefined && createdTagIds.length === 0
      ? undefined
      : [...(baseTagIds ?? []), ...createdTagIds];

    const updated = await this.inventories.updateAsset({
      assetId: assetId(input.assetId),
      title,
      description: input.description.trim(),
      tagIds
    });

    return {
      id: updated.id,
      title: updated.title,
      message: `Updated ${updated.title}.`,
      undoableOperationId: updated.undoableOperationId
    };
  }
}
