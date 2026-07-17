import type { RefObject } from 'react';
import type { TextInput } from 'react-native';
import type {
  AssetCardViewModel
} from '../../application/assets/AssetViewModels';
import type {
  AssetBrowseCheckoutFilter,
  AssetBrowseKindFilter,
  AssetBrowseLifecycleFilter,
  AssetBrowseSort
} from '../../application/home/InventorySummaryRepository';
import type { AssetTagOptionViewModel } from '../../application/assets/InventoryAssetTagsQuery';
import type { LocationBrowserItemViewModel } from '../../application/locations/LocationsQuery';
import { spacing } from '../theme/tokens';

export type BrowseScope = 'all' | 'places' | 'containers' | 'items';

export type BrowsePlaceItemViewModel = Pick<
  LocationBrowserItemViewModel,
  'id' | 'title' | 'description' | 'containedAssetCountLabel' | 'recentAssetLabel' | 'photo'
>;

export type BrowseScopeOption = {
  readonly label: string;
  readonly value: BrowseScope;
};

export type BrowseSecondaryFilters = {
  readonly scope: BrowseScope;
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly checkoutState: AssetBrowseCheckoutFilter;
  readonly tagIds: readonly string[];
};

export type BrowseFilterToken =
  | { readonly key: 'scope'; readonly label: string; readonly type: 'scope' }
  | { readonly key: 'lifecycle'; readonly label: string; readonly type: 'lifecycle' }
  | { readonly key: 'checkout'; readonly label: string; readonly type: 'checkout' }
  | { readonly key: string; readonly label: string; readonly type: 'tag'; readonly tagId: string };

export function focusSearchInput(inputRef: RefObject<TextInput | null>): void {
  inputRef.current?.focus();
}

export function shouldAutoFocusSearchInput(_initialTagIds: readonly string[]): boolean {
  return false;
}

export function browseFilterCount(filters: BrowseSecondaryFilters): number {
  return (filters.scope === 'all' ? 0 : 1)
    + (filters.lifecycleState === 'active' ? 0 : 1)
    + (filters.checkoutState === 'any' ? 0 : 1)
    + filters.tagIds.length;
}

export function buildBrowseFilterTokens(
  filters: BrowseSecondaryFilters,
  tagOptions: readonly AssetTagOptionViewModel[]
): readonly BrowseFilterToken[] {
  const tagsById = new Map(tagOptions.map((tag) => [tag.id, tag]));
  const tokens: BrowseFilterToken[] = [];

  if (filters.scope !== 'all') {
    tokens.push({
      key: 'scope',
      label: buildBrowseScopeOptions().find((option) => option.value === filters.scope)?.label ?? 'Type',
      type: 'scope'
    });
  }
  if (filters.lifecycleState !== 'active') {
    tokens.push({
      key: 'lifecycle',
      label: filters.lifecycleState === 'archived' ? 'Archived' : 'All statuses',
      type: 'lifecycle'
    });
  }
  if (filters.checkoutState !== 'any') {
    tokens.push({
      key: 'checkout',
      label: filters.checkoutState === 'checked_out' ? 'Checked out' : 'Available',
      type: 'checkout'
    });
  }
  filters.tagIds.forEach((tagId) => {
    const tag = tagsById.get(tagId);
    tokens.push({
      key: `tag:${tagId}`,
      label: tag?.label ?? 'Tag',
      type: 'tag',
      tagId
    });
  });

  return tokens;
}

export function removeBrowseFilter(
  filters: BrowseSecondaryFilters,
  token: BrowseFilterToken
): BrowseSecondaryFilters {
  return {
    scope: token.type === 'scope' ? 'all' : filters.scope,
    lifecycleState: token.type === 'lifecycle' ? 'active' : filters.lifecycleState,
    checkoutState: token.type === 'checkout' ? 'any' : filters.checkoutState,
    tagIds: token.type === 'tag'
      ? filters.tagIds.filter((tagId) => tagId !== token.tagId)
      : filters.tagIds
  };
}

export function openBrowseFilterDraft(filters: BrowseSecondaryFilters): BrowseSecondaryFilters {
  return { ...filters, tagIds: [...filters.tagIds] };
}

export function commitBrowseFilterDraft(draft: BrowseSecondaryFilters): BrowseSecondaryFilters {
  return { ...draft, tagIds: [...draft.tagIds] };
}

export function sortLabel(sort: AssetBrowseSort): string {
  return sort === 'updated_desc' ? 'Recently changed' : 'Default order';
}

export function browseColumnCount({
  fontScale,
  scope,
  width
}: {
  readonly fontScale: number;
  readonly scope: BrowseScope;
  readonly width: number;
}): 1 | 2 {
  return scope === 'places' || fontScale >= 1.35 || width < 350 ? 1 : 2;
}

export function browseGridCardWidth(width: number, columnCount: 1 | 2): number | undefined {
  return columnCount === 2
    ? Math.floor((width - (spacing.md * 2) - spacing.sm) / 2)
    : undefined;
}

export function canLoadNextBrowsePage(
  status: 'loading' | 'ready' | 'error',
  errorPhase?: 'initial' | 'replacement' | 'pagination'
): boolean {
  return status === 'ready' || (status === 'error' && errorPhase === 'pagination');
}

export function browseContinuationCriteria(criteria: {
  readonly query: string;
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly checkoutState: AssetBrowseCheckoutFilter;
  readonly scope: BrowseScope;
  readonly sort: AssetBrowseSort;
  readonly tagIds: readonly string[];
}) {
  return {
    query: criteria.query,
    lifecycleState: criteria.lifecycleState,
    checkoutState: criteria.checkoutState,
    scope: criteria.scope,
    sort: criteria.sort,
    tagIds: criteria.tagIds
  } as const;
}

export function cancelPendingBrowseSearch(
  timer: { current: ReturnType<typeof setTimeout> | undefined },
  query: string,
  clearTimer: (timer: ReturnType<typeof setTimeout>) => void = clearTimeout
): string {
  if (timer.current) {
    clearTimer(timer.current);
    timer.current = undefined;
  }
  return query.trim();
}

export function browseLoadingFlagsForRefresh() {
  return { isLoadingMore: false, isRefreshing: true } as const;
}

export function buildBrowseScopeOptions(): readonly BrowseScopeOption[] {
  return [
    { label: 'All', value: 'all' },
    { label: 'Places', value: 'places' },
    { label: 'Containers', value: 'containers' },
    { label: 'Items', value: 'items' }
  ];
}

export function browseScopeToKind(scope: BrowseScope): AssetBrowseKindFilter {
  switch (scope) {
    case 'containers':
      return 'container';
    case 'items':
      return 'item';
    case 'places':
      return 'location';
    case 'all':
      return 'all';
  }
}

export function parseBrowseScope(value: string | readonly string[] | undefined): BrowseScope {
  const rawValue = Array.isArray(value) ? value[0] : value;
  switch (rawValue) {
    case 'places':
    case 'containers':
    case 'items':
    case 'all':
      return rawValue;
    default:
      return 'all';
  }
}

export function searchResultSummaryLabel({
  hasTagFilters,
  lifecycleState,
  query,
  resultCount,
  scope,
  sort
}: {
  readonly hasTagFilters?: boolean;
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly query: string;
  readonly resultCount: number;
  readonly scope: BrowseScope;
  readonly sort: AssetBrowseSort;
}): string {
  const trimmedQuery = query.trim();
  if (trimmedQuery.length > 0 || hasTagFilters) {
    return trimmedQuery.length > 0
      ? `${resultCount.toString()} shown for “${trimmedQuery}” · relevance`
      : `${resultCount.toString()} shown · relevance`;
  }
  return `${resultCount.toString()} shown · ${sortLabel(sort)}`;
}

export function locationRowsFromAssetCards(
  assets: readonly AssetCardViewModel[],
  locations: readonly LocationBrowserItemViewModel[]
): readonly BrowsePlaceItemViewModel[] {
  const locationsById = new Map(locations.map((location) => [location.id, location]));
  return assets.map((asset) => {
    const location = locationsById.get(asset.id);
    return {
      id: asset.id,
      title: asset.title,
      description: asset.description,
      containedAssetCountLabel: location?.containedAssetCountLabel ?? 'Contents not summarized',
      recentAssetLabel: location?.recentAssetLabel ?? asset.locationTrailLabel,
      photo: asset.photo
    };
  });
}
