import { assetId, type AssetKind } from '../../domain/assets/AssetSummary';
import type {
  CreateInventoryAssetTagInput,
  CreateInventoryAssetPhotoInput,
  InventorySummaryRepository
} from '../home/InventorySummaryRepository';

export type CreateAssetCommandInput = {
  readonly kind?: AssetKind;
  readonly title: string;
  readonly description: string;
  readonly parentAssetId?: string;
  readonly tagIds?: readonly string[];
  readonly newTags?: readonly CreateInventoryAssetTagInput[];
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

    const createdTagIds = await createNewTags(this.inventories, input.newTags);
    const tagIds = [...(normalizeTagIds(input.tagIds) ?? []), ...createdTagIds];

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

async function createNewTags(
  inventories: InventorySummaryRepository,
  newTags: readonly CreateInventoryAssetTagInput[] | undefined
): Promise<readonly string[]> {
  if ((newTags?.length ?? 0) === 0) {
    return [];
  }
  if (!inventories.createAssetTag) {
    throw new Error('Tag creation is not available.');
  }
  const created = [];
  for (const tag of newTags ?? []) {
    const displayName = tag.displayName.trim();
    if (displayName.length === 0) {
      continue;
    }
    created.push(await inventories.createAssetTag({
      displayName,
      color: tag.color?.trim() || undefined
    }));
  }
  return created.map((tag) => tag.id);
}

function normalizeTagIds(tagIds: readonly string[] | undefined): readonly string[] | undefined {
  const normalized = (tagIds ?? []).map((tagId) => tagId.trim()).filter((tagId) => tagId.length > 0);
  return normalized.length > 0 ? normalized : undefined;
}
