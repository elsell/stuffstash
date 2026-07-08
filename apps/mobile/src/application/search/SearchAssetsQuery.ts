import {
  AssetCardViewModel,
  toAssetCardViewModel
} from '../assets/AssetViewModels';
import type {
  AssetBrowseCheckoutFilter,
  AssetBrowseKindFilter,
  AssetBrowseLifecycleFilter,
  AssetBrowseSort,
  InventorySummaryRepository
} from '../home/InventorySummaryRepository';

export type SearchAssetsMode = 'browse' | 'search';

export type SearchAssetsQueryInput = {
  readonly query: string;
  readonly cursor?: string;
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly kind: AssetBrowseKindFilter;
  readonly checkoutState: AssetBrowseCheckoutFilter;
  readonly sort: AssetBrowseSort;
  readonly limit?: number;
  readonly tagIds?: readonly string[];
};

export type SearchAssetsViewModel = {
  readonly query: string;
  readonly mode: SearchAssetsMode;
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly kind: AssetBrowseKindFilter;
  readonly checkoutState: AssetBrowseCheckoutFilter;
  readonly sort: AssetBrowseSort;
  readonly tagIds: readonly string[];
  readonly assets: readonly AssetCardViewModel[];
  readonly nextCursor?: string;
  readonly hasMore: boolean;
};

export class SearchAssetsQuery {
  constructor(private readonly inventories: InventorySummaryRepository) {}

  async execute(input: SearchAssetsQueryInput): Promise<SearchAssetsViewModel> {
    const trimmed = input.query.trim();
    const page = await this.inventories.browseAssets({
      ...input,
      query: trimmed
    });
    const searchMatchLabelsByAssetId = new Map(
      (page.searchMatches ?? []).map((match) => [match.assetId, match.labels])
    );

    return {
      query: trimmed,
      mode: trimmed.length > 0 ? 'search' : 'browse',
      lifecycleState: input.lifecycleState,
      kind: input.kind,
      checkoutState: input.checkoutState,
      sort: input.sort,
      tagIds: input.tagIds ?? [],
      assets: page.assets.map((asset) => {
        const card = toAssetCardViewModel(asset);
        const searchMatchLabels = searchMatchLabelsByAssetId.get(asset.id);
        return searchMatchLabels && searchMatchLabels.length > 0
          ? { ...card, searchMatchLabels }
          : card;
      }),
      nextCursor: page.nextCursor,
      hasMore: page.hasMore
    };
  }
}
