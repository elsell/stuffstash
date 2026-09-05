import { useEffect, useMemo, useRef, useState } from 'react';
import { router } from 'expo-router';
import {
  ActivityIndicator,
  FlatList,
  useWindowDimensions,
  View
} from 'react-native';
import type { TextInput } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { AddAssetPhotosCommand } from '../../application/assets/AddAssetPhotosCommand';
import type { AssetCheckoutCommand } from '../../application/assets/AssetCheckoutCommand';
import type { AssetLifecycleCommand } from '../../application/assets/AssetLifecycleCommand';
import type { DeleteAssetPhotoCommand } from '../../application/assets/DeleteAssetPhotoCommand';
import type { AssetDetailQuery } from '../../application/assets/AssetDetailQuery';
import type { InventoryMapQuery } from '../../application/assets/InventoryMapQuery';
import type { AssetCardViewModel } from '../../application/assets/AssetViewModels';
import type { PhotoSelectionQuery } from '../../application/add/PhotoSelectionQuery';
import type {
  InventoryAssetTagsQuery
} from '../../application/assets/InventoryAssetTagsQuery';
import type {
  AssetBrowseCheckoutFilter,
  AssetBrowseLifecycleFilter,
  AssetBrowseSort
} from '../../application/home/InventorySummaryRepository';
import type { LocationsQuery } from '../../application/locations/LocationsQuery';
import type { SearchAssetsQuery } from '../../application/search/SearchAssetsQuery';
import { AssetCard } from '../components/AssetCard';
import { appKeyboardDismissMode } from '../components/AppTextInput';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { assetDetailHref } from './AssetDetailNavigation';
import { navigateToAssetTagSearch } from './AssetTagSearchNavigation';
import { BrowsePlaceRow } from './BrowsePlaceRow';
import {
  browseRouteParamsForState,
  consumeLocalBrowseRouteEffect,
  type AppliedBrowseRouteState
} from './BrowseRouteParams';
import {
  BrowseEmptyState,
  BrowseLoadError,
  BrowsePaginationRetry
} from './BrowseResultStates';
import {
  BrowseDraftFilters,
  SearchHeader
} from './BrowseHeader';
import type { InventoryMapSurface } from './InventoryMapPresentation';
import { InventoryMapScreen } from './InventoryMapScreen';
import {
  browseFilterCount,
  browseColumnCount,
  browseGridCardWidth,
  BrowseFilterToken,
  BrowsePlaceItemViewModel,
  BrowseScope,
  cancelPendingBrowseSearch,
  commitBrowseFilterDraft,
  locationRowsFromAssetCards,
  openBrowseFilterDraft,
  removeBrowseFilter
} from './SearchScreenPresentation';
import { createSearchScreenStyles } from './SearchScreen.styles';
import { useInfiniteQuery, useQuery } from '@tanstack/react-query';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileServerStateScope } from '../navigation/MobileServerStateProvider';
import { useMobileInventoryServerQuery } from '../serverState/useMobileInventoryServerQuery';
import { browseInfiniteQueryOptions } from '../serverState/BrowseInfiniteQuery';
import type { InventoryContextQuery } from '../../application/home/InventoryContextQuery';

export { SearchHeader } from './BrowseHeader';

type SearchScreenProps = {
  readonly initialSurface?: InventoryMapSurface;
  readonly initialScope?: BrowseScope;
  readonly initialQuery?: string;
  readonly initialTagIds?: readonly string[];
  readonly initialLifecycleState?: AssetBrowseLifecycleFilter;
  readonly initialCheckoutState?: AssetBrowseCheckoutFilter;
  readonly initialSort?: AssetBrowseSort;
  readonly addAssetPhotosCommand: Pick<AddAssetPhotosCommand, 'execute'>;
  readonly assetCheckoutCommand: Pick<AssetCheckoutCommand, 'execute'>;
  readonly assetDetailQuery: Pick<AssetDetailQuery, 'execute'>;
  readonly assetLifecycleCommand: Pick<AssetLifecycleCommand, 'execute'>;
  readonly deleteAssetPhotoCommand: Pick<DeleteAssetPhotoCommand, 'execute'>;
  readonly inventoryMapQuery: Pick<InventoryMapQuery, 'execute'>;
  readonly inventoryContextQuery: Pick<InventoryContextQuery, 'execute'>;
  readonly inventoryAssetTagsQuery: Pick<InventoryAssetTagsQuery, 'execute'>;
  readonly locationsQuery: Pick<LocationsQuery, 'execute'>;
  readonly photoSelectionQuery: PhotoSelectionQuery;
  readonly searchAssetsQuery: Pick<SearchAssetsQuery, 'execute'>;
};

type BrowseResults = {
  readonly scope: BrowseScope;
  readonly query: string;
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly checkoutState: AssetBrowseCheckoutFilter;
  readonly sort: AssetBrowseSort;
  readonly tagIds: readonly string[];
  readonly assets: readonly AssetCardViewModel[];
  readonly locations: readonly BrowsePlaceItemViewModel[];
  readonly nextCursor?: string;
  readonly hasMore: boolean;
};

type BrowseErrorPhase = 'initial' | 'replacement' | 'pagination';

type BrowseState =
  | { readonly status: 'loading'; readonly results: BrowseResults; readonly isInitial: boolean }
  | { readonly status: 'ready'; readonly results: BrowseResults }
  | { readonly status: 'error'; readonly message: string; readonly phase: BrowseErrorPhase; readonly results: BrowseResults };

type BrowseListItem =
  | { readonly type: 'asset'; readonly asset: AssetCardViewModel }
  | { readonly type: 'place'; readonly location: BrowsePlaceItemViewModel };

export function SearchScreen({
  initialSurface = 'list',
  initialScope = 'all',
  initialQuery = '',
  initialTagIds = [],
  initialLifecycleState = 'active',
  initialCheckoutState = 'any',
  initialSort = 'updated_desc',
  addAssetPhotosCommand,
  assetCheckoutCommand,
  assetDetailQuery,
  assetLifecycleCommand,
  deleteAssetPhotoCommand,
  inventoryMapQuery,
  inventoryContextQuery,
  inventoryAssetTagsQuery,
  locationsQuery,
  photoSelectionQuery,
  searchAssetsQuery
}: SearchScreenProps) {
  const { fontScale, width } = useWindowDimensions();
  const palette = useAppearancePalette();
  const serverState = useMobileServerStateScope();
  const inventoryScope = useQuery({
    queryKey: mobileQueryKeys.inventoryScope(serverState.scopeId),
    queryFn: ({ signal }) => serverState.loadInventoryScope({ signal }),
    staleTime: Infinity
  });
  const styles = useMemo(() => createSearchScreenStyles(palette), [palette]);
  const normalizedInitialTags = useMemo(() => uniqueTagIds(initialTagIds), [initialTagIds.join('|')]);
  const [query, setQuery] = useState(initialQuery);
  const [submittedQuery, setSubmittedQuery] = useState(initialQuery.trim());
  const [scope, setScope] = useState<BrowseScope>(initialScope);
  const [surface, setSurface] = useState<InventoryMapSurface>(initialSurface);
  const [lifecycleState, setLifecycleState] = useState<AssetBrowseLifecycleFilter>(initialLifecycleState);
  const [checkoutState, setCheckoutState] = useState<AssetBrowseCheckoutFilter>(initialCheckoutState);
  const [sort, setSort] = useState<AssetBrowseSort>(initialSort);
  const [selectedTagIds, setSelectedTagIds] = useState<readonly string[]>(normalizedInitialTags);
  const [filterDraft, setFilterDraft] = useState<BrowseDraftFilters>({
    scope: initialScope,
    lifecycleState: initialLifecycleState,
    checkoutState: initialCheckoutState,
    tagIds: normalizedInitialTags
  });
  const [filtersExpanded, setFiltersExpanded] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isSearchFocused, setIsSearchFocused] = useState(false);
  const queryTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const mapPathStore = useRef(new Map<string, readonly string[]>());
  const searchInputRef = useRef<TextInput>(null);
  const lastRequestedQuery = useRef(initialQuery.trim());
  const identity = inventoryScope.data;
  const scopeIdentity = identity ? JSON.stringify([serverState.scopeId, identity.tenantId, identity.inventoryId]) : undefined;
  const browse = useInfiniteQuery({
    ...browseInfiniteQueryOptions(serverState.scopeId, identity?.tenantId ?? 'pending', identity?.inventoryId ?? 'pending', {
      query: submittedQuery, scope, lifecycleState, checkoutState, sort, tagIds: selectedTagIds
    }, searchAssetsQuery),
    enabled: inventoryScope.isSuccess && surface === 'list',
    subscribed: inventoryScope.isSuccess && surface === 'list'
  });
  const context = useMobileInventoryServerQuery({
    key: mobileQueryKeys.inventoryContext,
    query: (signal) => inventoryContextQuery.execute({ signal })
  });
  const tags = useMobileInventoryServerQuery({
    key: mobileQueryKeys.assetTags,
    query: (signal) => inventoryAssetTagsQuery.execute({ signal }),
    enabled: surface === 'list'
  });
  const places = useMobileInventoryServerQuery({
    key: mobileQueryKeys.locations,
    query: (signal) => locationsQuery.execute({ signal }),
    enabled: surface === 'list' && scope === 'places'
  });
  const previous = useRef<{ scope: string; data: NonNullable<typeof browse.data> } | undefined>(undefined);
  if (scopeIdentity && browse.data) previous.current = { scope: scopeIdentity, data: browse.data };
  const data = browse.data ?? (scopeIdentity && previous.current?.scope === scopeIdentity ? previous.current.data : undefined);
  const firstPage = data?.pages[0];
  const loadedCriteria = firstPage?.criteria;
  const cards = data?.pages.flatMap((page) => page.assets) ?? [];
  const results: BrowseResults = {
    ...(loadedCriteria ?? { query: submittedQuery, scope, lifecycleState, checkoutState, sort, tagIds: selectedTagIds }),
    assets: loadedCriteria?.scope === 'places' ? [] : cards,
    locations: loadedCriteria?.scope === 'places' ? locationRowsFromAssetCards(cards, places.data?.locations ?? []) : [],
    hasMore: Boolean(browse.hasNextPage),
    nextCursor: data?.pages.at(-1)?.nextCursor
  };
  const error = inventoryScope.error ?? browse.error;
  const state: BrowseState = error
    ? { status: 'error', results, phase: browse.isFetchNextPageError ? 'pagination' : data ? 'replacement' : 'initial', message: 'This inventory could not be loaded.' }
    : browse.isPending || !identity
      ? { status: 'loading', results, isInitial: !data }
      : { status: 'ready', results };
  const tagFilters = tags.data ?? [];
  const tagFilterStatus = tags.isError ? 'error' : tags.data ? 'ready' : 'loading';
  const inventoryContext = context.data;
  const inventoryContextStatus = context.isError ? 'error' : context.data ? 'ready' : 'loading';
  const isLoadingMore = browse.isFetchingNextPage;
  const localRouteEffectKeys = useRef(new Set<string>());
  useEffect(() => () => {
    if (queryTimer.current) clearTimeout(queryTimer.current);
  }, []);

  useEffect(() => {
    mapPathStore.current.clear();
  }, [scopeIdentity]);

  useEffect(() => {
    const nextQuery = initialQuery.trim();
    const nextTags = uniqueTagIds(initialTagIds);
    const routeKey = browseRouteStateKey({
      surface: initialSurface,
      scope: initialScope,
      query: nextQuery,
      tagIds: nextTags,
      lifecycleState: initialLifecycleState,
      checkoutState: initialCheckoutState,
      sort: initialSort
    });
    if (consumeLocalBrowseRouteEffect(localRouteEffectKeys.current, routeKey)) {
      return;
    }
    cancelPendingBrowseSearch(queryTimer, nextQuery);
    setQuery(nextQuery);
    setSurface(initialSurface);
    setScope(initialScope);
    setLifecycleState(initialLifecycleState);
    setCheckoutState(initialCheckoutState);
    setSort(initialSort);
    setSelectedTagIds(nextTags);
    setFilterDraft({ scope: initialScope, lifecycleState: initialLifecycleState, checkoutState: initialCheckoutState, tagIds: nextTags });
    lastRequestedQuery.current = nextQuery;
    loadFirstPage({ query: nextQuery });
  }, [
    initialQuery,
    initialSurface,
    initialScope,
    initialLifecycleState,
    initialCheckoutState,
    initialSort,
    initialTagIds.join('|'),
    locationsQuery,
    searchAssetsQuery
  ]);

  function loadFirstPage(next: { readonly query?: string } = {}): void {
    setSubmittedQuery((next.query ?? query).trim());
  }

  async function refreshResults(): Promise<void> {
    setIsRefreshing(true);
    try {
      await Promise.all([
        browse.refetch({ cancelRefetch: false }),
        ...(scope === 'places' ? [places.refetch({ cancelRefetch: false })] : [])
      ]);
    } finally { setIsRefreshing(false); }
  }

  async function loadNextPage(): Promise<void> {
    if (!browse.data || browse.isFetching || !browse.hasNextPage || isRefreshing) return;
    await browse.fetchNextPage({ cancelRefetch: false });
  }

  function scheduleSearch(nextQuery: string): void {
    setQuery(nextQuery);
    if (queryTimer.current) clearTimeout(queryTimer.current);
    if (nextQuery.trim() === lastRequestedQuery.current) return;
    queryTimer.current = setTimeout(() => submitQuery(nextQuery), 300);
  }

  function submitQuery(nextQuery = query): void {
    if (queryTimer.current) clearTimeout(queryTimer.current);
    const normalized = nextQuery.trim();
    lastRequestedQuery.current = normalized;
    syncBrowseRoute({ query: normalized });
    void loadFirstPage({ query: normalized });
  }

  function cancelPendingSearch(): string {
    const normalized = cancelPendingBrowseSearch(queryTimer, query);
    lastRequestedQuery.current = normalized;
    return normalized;
  }

  function clearSearch(): void {
    setQuery('');
    submitQuery('');
  }

  function updateSort(nextSort: AssetBrowseSort): void {
    const nextQuery = cancelPendingSearch();
    setSort(nextSort);
    syncBrowseRoute({ query: nextQuery, sort: nextSort });
    loadFirstPage({ query: nextQuery });
  }

  function updateSurface(nextSurface: InventoryMapSurface): void {
    if (nextSurface === surface) return;
    setSurface(nextSurface);
    syncBrowseRoute({ surface: nextSurface });
  }

  function openFilters(expanded: boolean): void {
    if (expanded) setFilterDraft(openBrowseFilterDraft({ scope, lifecycleState, checkoutState, tagIds: selectedTagIds }));
    setFiltersExpanded(expanded);
  }

  function applyFilters(filters: BrowseDraftFilters): void {
    const committed = commitBrowseFilterDraft(filters);
    const nextQuery = cancelPendingSearch();
    setLifecycleState(committed.lifecycleState);
    setCheckoutState(committed.checkoutState);
    setScope(committed.scope);
    setSelectedTagIds(committed.tagIds);
    setFiltersExpanded(false);
    syncBrowseRoute({ ...committed, query: nextQuery });
    loadFirstPage({ query: nextQuery });
  }

  function syncBrowseRoute(next: Partial<AppliedBrowseRouteState>): void {
    const routeState: AppliedBrowseRouteState = {
      surface: next.surface ?? surface,
      scope: next.scope ?? scope,
      query: next.query ?? query,
      tagIds: next.tagIds ?? selectedTagIds,
      lifecycleState: next.lifecycleState ?? lifecycleState,
      checkoutState: next.checkoutState ?? checkoutState,
      sort: next.sort ?? sort
    };
    const currentRouteState: AppliedBrowseRouteState = {
      surface: initialSurface,
      scope: initialScope,
      query: initialQuery.trim(),
      tagIds: uniqueTagIds(initialTagIds),
      lifecycleState: initialLifecycleState,
      checkoutState: initialCheckoutState,
      sort: initialSort
    };
    const routeKey = browseRouteStateKey(routeState);
    if (routeKey === browseRouteStateKey(currentRouteState)) return;
    localRouteEffectKeys.current.add(routeKey);
    router.setParams(browseRouteParamsForState(routeState));
  }

  function clearFilters(): void {
    applyFilters({ scope: 'all', lifecycleState: 'active', checkoutState: 'any', tagIds: [] });
  }

  function removeFilter(token: BrowseFilterToken): void {
    applyFilters(removeBrowseFilter({ scope, lifecycleState, checkoutState, tagIds: selectedTagIds }, token));
  }

  function retryResults(): void {
    if (state.status === 'error' && state.phase === 'pagination') {
      void loadNextPage();
      return;
    }
    if (places.isError && scope === 'places' && !error) {
      void places.refetch({ cancelRefetch: false });
      return;
    }
    void (inventoryScope.isError ? inventoryScope.refetch() : browse.refetch({ cancelRefetch: false }));
  }

  const listItems = toBrowseListItems(state.results);
  const resultScope = state.results.scope;
  const numColumns = browseColumnCount({ fontScale, scope: resultScope, width });
  const gridCardWidth = browseGridCardWidth(width, numColumns);
  const hasActiveFilters = browseFilterCount({ scope, lifecycleState, checkoutState, tagIds: selectedTagIds }) > 0;
  const isInitialError = state.status === 'error' && state.phase === 'initial';
  const isPaginationError = state.status === 'error' && state.phase === 'pagination';

  if (surface === 'map') {
    return (
      <SafeAreaView style={styles.shell} edges={['top', 'left', 'right']}>
        <InventoryMapScreen
          addAssetPhotosCommand={addAssetPhotosCommand}
          assetCheckoutCommand={assetCheckoutCommand}
          assetDetailQuery={assetDetailQuery}
          assetLifecycleCommand={assetLifecycleCommand}
          canAdd={inventoryContext?.canAdd ?? false}
          deleteAssetPhotoCommand={deleteAssetPhotoCommand}
          inventoryMapQuery={inventoryMapQuery}
          pathStore={mapPathStore}
          photoSelectionQuery={photoSelectionQuery}
          selectedSurface={surface}
          onAdd={() => router.navigate('/add')}
          onChangeSurface={updateSurface}
        />
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.shell} edges={['top', 'left', 'right']}>
      <FlatList
        key={`${resultScope}:${numColumns.toString()}`}
        data={listItems}
        keyExtractor={keyBrowseListItem}
        columnWrapperStyle={numColumns === 2 ? styles.cardRow : undefined}
        contentContainerStyle={styles.content}
        keyboardDismissMode={appKeyboardDismissMode()}
        keyboardShouldPersistTaps="handled"
        numColumns={numColumns}
        refreshing={isRefreshing}
        onEndReached={() => void loadNextPage()}
        onEndReachedThreshold={0.55}
        onRefresh={() => void refreshResults()}
        ListHeaderComponent={
          <SearchHeader
            canAdd={inventoryContext?.canAdd ?? false}
            isLoading={state.status === 'loading'}
            lifecycleState={lifecycleState}
            checkoutState={checkoutState}
            filtersExpanded={filtersExpanded}
            filterDraft={filterDraft}
            inventoryContext={inventoryContext?.inventoryName}
            inventoryContextStatus={inventoryContextStatus}
            palette={palette}
            query={query}
            resultCount={listItems.length}
            scope={scope}
            selectedSurface={surface}
            selectedTagIds={selectedTagIds}
            searchInputRef={searchInputRef}
            searchInputFocused={isSearchFocused}
            sort={sort}
            statusMessage={state.status === 'error' && state.phase === 'replacement'
              ? state.message
              : scope === 'places' && places.isError ? 'Place summaries could not load. Your places are still available.' : undefined}
            submittedQuery={state.results.query}
            tagFilters={tagFilters}
            tagFilterStatus={tagFilterStatus}
            onApplyFilters={applyFilters}
            onAdd={() => router.navigate('/add')}
            onChangeDraftCheckoutState={(value) => setFilterDraft((draft) => ({ ...draft, checkoutState: value }))}
            onChangeDraftLifecycleState={(value) => setFilterDraft((draft) => ({ ...draft, lifecycleState: value }))}
            onChangeDraftScope={(value) => setFilterDraft((draft) => ({ ...draft, scope: value }))}
            onChangeDraftTagIds={(value) => setFilterDraft((draft) => ({ ...draft, tagIds: value }))}
            onChangeQuery={scheduleSearch}
            onChangeSort={updateSort}
            onChangeSurface={updateSurface}
            onClearFilters={clearFilters}
            onClearQuery={clearSearch}
            onRemoveFilter={removeFilter}
            onRetryResults={retryResults}
            onRetryInventoryContext={() => void context.refetch({ cancelRefetch: false })}
            onRetryTags={() => void tags.refetch({ cancelRefetch: false })}
            onSearchBlur={() => setIsSearchFocused(false)}
            onSearchFocus={() => setIsSearchFocused(true)}
            onSubmit={() => submitQuery()}
            onToggleFilters={openFilters}
          />
        }
        ListEmptyComponent={
          state.status === 'loading' ? null : isInitialError ? (
            <BrowseLoadError message={state.message} palette={palette} onRetry={retryResults} />
          ) : state.results.query.trim() ? (
            <BrowseEmptyState kind="search" palette={palette} query={state.results.query} onClearSearch={clearSearch} />
          ) : hasActiveFilters ? (
            <BrowseEmptyState kind="filters" palette={palette} onClearFilters={clearFilters} />
          ) : (
            <BrowseEmptyState
              kind="inventory"
              inventoryName={inventoryContext?.inventoryName ?? 'this inventory'}
              palette={palette}
              onAdd={inventoryContext?.canAdd ? () => router.navigate('/add') : undefined}
            />
          )
        }
        ListFooterComponent={
          isPaginationError ? (
            <BrowsePaginationRetry message={state.message} palette={palette} onRetry={retryResults} />
          ) : isLoadingMore ? (
            <View style={styles.footer}><ActivityIndicator color={palette.accent} /></View>
          ) : null
        }
        renderItem={({ item }) => item.type === 'place' ? (
          <BrowsePlaceRow
            location={item.location}
            palette={palette}
            onPress={() => router.push(assetDetailHref(item.location.id))}
          />
        ) : (
          <AssetCard
            asset={item.asset}
            palette={palette}
            style={gridCardWidth
              ? { maxWidth: gridCardWidth, minWidth: gridCardWidth, width: gridCardWidth }
              : styles.singleCardRow}
            onParentLocationPress={(location) => router.push(assetDetailHref(location.id))}
            onPress={() => router.push(assetDetailHref(item.asset.id))}
            onTagPress={(tag) => navigateToAssetTagSearch(router, tag)}
          />
        )}
      />
    </SafeAreaView>
  );
}

function toBrowseListItems(results: BrowseResults): readonly BrowseListItem[] {
  return results.scope === 'places'
    ? results.locations.map((location) => ({ type: 'place' as const, location }))
    : results.assets.map((asset) => ({ type: 'asset' as const, asset }));
}

function keyBrowseListItem(item: BrowseListItem): string {
  return item.type === 'place' ? `place:${item.location.id}` : `asset:${item.asset.id}`;
}

function uniqueTagIds(tagIds: readonly string[]): readonly string[] {
  return [...new Set(tagIds.map((id) => id.trim()).filter(Boolean))];
}

function browseRouteStateKey(state: AppliedBrowseRouteState): string {
  return JSON.stringify({ ...state, query: state.query.trim(), tagIds: uniqueTagIds(state.tagIds) });
}
