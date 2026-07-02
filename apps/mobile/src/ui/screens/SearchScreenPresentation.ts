import type { RefObject } from 'react';
import type { TextInput } from 'react-native';
import type {
  AssetCardViewModel
} from '../../application/assets/AssetViewModels';
import type {
  AssetBrowseKindFilter,
  AssetBrowseLifecycleFilter,
  AssetBrowseSort
} from '../../application/home/InventorySummaryRepository';
import type { LocationBrowserItemViewModel } from '../../application/locations/LocationsQuery';

export type BrowseScope = 'all' | 'places' | 'containers' | 'items';

export type BrowseScopeOption = {
  readonly label: string;
  readonly value: BrowseScope;
};

export function focusSearchInput(inputRef: RefObject<TextInput | null>): void {
  inputRef.current?.focus();
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
  lifecycleState,
  query,
  resultCount,
  scope,
  sort
}: {
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly query: string;
  readonly resultCount: number;
  readonly scope: BrowseScope;
  readonly sort: AssetBrowseSort;
}): string {
  const scopeLabel = labelScope(scope);
  const statusLabel = labelLifecycle(lifecycleState);
  const trimmedQuery = query.trim();
  const queryLabel = trimmedQuery.length > 0 ? ` for "${trimmedQuery}"` : '';

  if (trimmedQuery.length > 0) {
    return `Showing ${resultCount.toString()} ${statusLabel} ${scopeLabel}${queryLabel} · relevance order`;
  }

  if (scope === 'places') {
    return `Showing ${resultCount.toString()} ${statusLabel} ${scopeLabel}${queryLabel} · ${labelSort(sort)}`;
  }

  return `Showing ${resultCount.toString()} ${statusLabel} ${scopeLabel}${queryLabel} · ${labelSort(sort)}`;
}

export function locationRowsFromAssetCards(
  assets: readonly AssetCardViewModel[],
  locations: readonly LocationBrowserItemViewModel[]
): readonly LocationBrowserItemViewModel[] {
  const locationsById = new Map(locations.map((location) => [location.id, location]));
  return assets.map((asset) => {
    const location = locationsById.get(asset.id);
    return {
      id: asset.id,
      title: asset.title,
      description: asset.description,
      containedAssetCountLabel: location?.containedAssetCountLabel ?? 'Contents not summarized',
      recentAssetLabel: location?.recentAssetLabel ?? asset.locationTrailLabel,
      photoLabel: asset.photoLabel,
      photo: asset.photo
    };
  });
}

function labelScope(scope: BrowseScope): string {
  switch (scope) {
    case 'all':
      return 'things';
    case 'places':
      return 'places';
    case 'containers':
      return 'containers';
    case 'items':
      return 'items';
  }
}

function labelLifecycle(lifecycleState: AssetBrowseLifecycleFilter): string {
  switch (lifecycleState) {
    case 'active':
      return 'active';
    case 'archived':
      return 'archived';
    case 'all':
      return 'all';
  }
}

function labelSort(sort: AssetBrowseSort): string {
  switch (sort) {
    case 'updated_desc':
      return 'recent first';
    case 'id_asc':
      return 'stable order';
  }
}
