import type { AssetBrowseCheckoutFilter, AssetBrowseLifecycleFilter } from '../../application/home/InventorySummaryRepository';
import type { BrowseScope } from './SearchScreenPresentation';
export type BrowseDraftFilters = {
  readonly scope: BrowseScope;
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly checkoutState: AssetBrowseCheckoutFilter;
  readonly tagIds: readonly string[];
};
