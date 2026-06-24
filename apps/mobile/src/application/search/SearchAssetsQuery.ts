import {
  AssetDetailViewModel,
  AssetCardViewModel,
  toAssetCardViewModel,
  toAssetDetailViewModel
} from '../assets/AssetViewModels';
import type { InventorySummaryRepository } from '../home/InventorySummaryRepository';

export type SearchAssetsViewModel = {
  readonly query: string;
  readonly assets: readonly AssetCardViewModel[];
  readonly assetDetails: readonly AssetDetailViewModel[];
};

export class SearchAssetsQuery {
  constructor(private readonly inventories: InventorySummaryRepository) {}

  async execute(query: string): Promise<SearchAssetsViewModel> {
    const trimmed = query.trim();
    if (trimmed.length === 0) {
      return { query: trimmed, assets: [], assetDetails: [] };
    }

    const assets = await this.inventories.searchAssets(trimmed);

    return {
      query: trimmed,
      assets: assets.map(toAssetCardViewModel),
      assetDetails: assets.map(toAssetDetailViewModel)
    };
  }
}
