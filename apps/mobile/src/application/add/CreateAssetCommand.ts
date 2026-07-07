import { assetId, type AssetKind } from '../../domain/assets/AssetSummary';
import type { ActiveAssetTagReference, CreateAssetTagDraft } from '../assets/AssetTagDraftResolution';
import { createPendingAssetTags, reconcilePendingAssetTagDrafts } from '../assets/AssetTagDraftResolution';
import type {
  CreateInventoryAssetPhotoInput,
  InventorySummaryRepository
} from '../home/InventorySummaryRepository';

export type CreateAssetCommandInput = {
  readonly kind?: AssetKind;
  readonly title: string;
  readonly description: string;
  readonly parentAssetId?: string;
  readonly tagIds?: readonly string[];
  readonly newTags?: readonly CreateAssetTagDraft[];
  readonly activeTags?: readonly ActiveAssetTagReference[];
  readonly photos?: readonly CreateAssetPhotoInput[];
};

export type CreateAssetPhotoInput = CreateInventoryAssetPhotoInput;

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

    const reconciledTags = reconcilePendingAssetTagDrafts({
      selectedTagIds: input.tagIds ?? [],
      pendingTags: input.newTags ?? [],
      activeTags: input.activeTags ?? []
    });
    const createdTagIds = await createPendingAssetTags(this.inventories, reconciledTags.pendingTags);
    const tagIds = [...reconciledTags.tagIds, ...createdTagIds];

    const asset = await this.inventories.createAsset({
      kind: input.kind ?? 'item',
      title,
      description: input.description.trim(),
      parentAssetId: input.parentAssetId ? assetId(input.parentAssetId) : undefined,
      tagIds: tagIds.length > 0 ? tagIds : undefined
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
