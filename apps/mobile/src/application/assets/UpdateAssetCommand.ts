import { assetId } from '../../domain/assets/AssetSummary';
import type {
  CreateInventoryAssetTagInput,
  InventoryAssetUpdateRepository
} from '../home/InventorySummaryRepository';

type InventoryAssetTagCreateRepository = {
  createAssetTag?: (input: CreateInventoryAssetTagInput) => Promise<{ readonly id: string }>;
};

export type UpdateAssetCommandInput = {
  readonly assetId: string;
  readonly title: string;
  readonly description: string;
  readonly tagIds?: readonly string[];
  readonly newTags?: readonly CreateInventoryAssetTagInput[];
};

export type UpdateAssetCommandResult = {
  readonly id: string;
  readonly title: string;
  readonly message: string;
};

export class UpdateAssetCommand {
  constructor(private readonly inventories: InventoryAssetUpdateRepository & InventoryAssetTagCreateRepository) {}

  async execute(input: UpdateAssetCommandInput): Promise<UpdateAssetCommandResult> {
    const title = input.title.trim();
    if (title.length === 0) {
      throw new Error('Name is required.');
    }

    const createdTagIds = await createNewTags(this.inventories, input.newTags);
    const baseTagIds = normalizeTagIds(input.tagIds);
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
      message: `Updated ${updated.title}.`
    };
  }
}

async function createNewTags(
  inventories: InventoryAssetTagCreateRepository,
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
  if (tagIds === undefined) {
    return undefined;
  }
  const normalized = (tagIds ?? []).map((tagId) => tagId.trim()).filter((tagId) => tagId.length > 0);
  return normalized;
}
