import type { AssetTagSummary } from '../../domain/assets/AssetSummary';
import type { ReadRequest } from '../shared/ReadRequest';

export type AssetTagOptionViewModel = {
  readonly id: string;
  readonly key: string;
  readonly label: string;
  readonly color?: string;
};

export type InventoryAssetTagsRepository = {
  getInventoryAssetTags(request?: ReadRequest): Promise<readonly AssetTagSummary[]>;
};

export class InventoryAssetTagsQuery {
  constructor(private readonly inventories: InventoryAssetTagsRepository) {}

  async execute(request: ReadRequest = {}): Promise<readonly AssetTagOptionViewModel[]> {
    const tags = await this.inventories.getInventoryAssetTags(request);
    return tags.map(toTagOption);
  }
}

function toTagOption(tag: AssetTagSummary): AssetTagOptionViewModel {
  return {
    id: tag.id,
    key: tag.key,
    label: tag.displayName,
    color: tag.color
  };
}
