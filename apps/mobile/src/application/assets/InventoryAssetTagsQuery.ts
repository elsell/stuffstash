import type { AssetTagSummary } from '../../domain/assets/AssetSummary';
import type { InventorySummaryRepository } from '../home/InventorySummaryRepository';

export type AssetTagOptionViewModel = {
  readonly id: string;
  readonly key: string;
  readonly label: string;
  readonly color?: string;
};

export class InventoryAssetTagsQuery {
  constructor(private readonly inventories: InventorySummaryRepository) {}

  async execute(): Promise<readonly AssetTagOptionViewModel[]> {
    const inventory = await this.inventories.getDefaultInventorySummary();
    return (inventory.assetTags ?? []).map(toTagOption);
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
