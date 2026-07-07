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
};

export type SearchAssetsViewModel = {
  readonly query: string;
  readonly mode: SearchAssetsMode;
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly kind: AssetBrowseKindFilter;
  readonly checkoutState: AssetBrowseCheckoutFilter;
  readonly sort: AssetBrowseSort;
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

    return {
      query: trimmed,
      mode: trimmed.length > 0 ? 'search' : 'browse',
      lifecycleState: input.lifecycleState,
      kind: input.kind,
      checkoutState: input.checkoutState,
      sort: input.sort,
      assets: page.assets.map(toAssetCardViewModel),
      nextCursor: page.nextCursor,
      hasMore: page.hasMore
    };
  }
}
