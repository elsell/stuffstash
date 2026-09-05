import { infiniteQueryOptions } from '@tanstack/react-query';
import type { SearchAssetsQuery } from '../../application/search/SearchAssetsQuery';
import type { AssetBrowseCheckoutFilter, AssetBrowseLifecycleFilter, AssetBrowseSort } from '../../application/home/InventorySummaryRepository';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { browseScopeToKind, type BrowseScope } from '../screens/SearchScreenPresentation';

export type BrowseCriteria = {
  readonly query: string;
  readonly scope: BrowseScope;
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly checkoutState: AssetBrowseCheckoutFilter;
  readonly sort: AssetBrowseSort;
  readonly tagIds: readonly string[];
};

export function browseInfiniteQueryOptions(
  scopeId: string,
  tenantId: string,
  inventoryId: string,
  criteria: BrowseCriteria,
  query: Pick<SearchAssetsQuery, 'execute'>
) {
  return infiniteQueryOptions({
    queryKey: mobileQueryKeys.browsePages(scopeId, tenantId, inventoryId, criteria),
    initialPageParam: undefined as string | undefined,
    queryFn: async ({ signal, pageParam }) => ({ ...await query.execute({
      ...criteria, query: criteria.query.trim(), tagIds: [...new Set(criteria.tagIds)].sort(),
      kind: browseScopeToKind(criteria.scope), cursor: pageParam, limit: 20, signal
    }), criteria }),
    getNextPageParam: (page) => page.hasMore ? page.nextCursor : undefined
  });
}
