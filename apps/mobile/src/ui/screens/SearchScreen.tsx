import { useEffect, useMemo, useRef, useState } from 'react';
import { router } from 'expo-router';
import {
  ActivityIndicator,
  FlatList,
  TextInput,
  useWindowDimensions,
  View
} from 'react-native';
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
  AssetTagOptionViewModel,
  InventoryAssetTagsQuery
} from '../../application/assets/InventoryAssetTagsQuery';
import type {
  AssetBrowseCheckoutFilter,
  AssetBrowseLifecycleFilter,
  AssetBrowseSort
} from '../../application/home/InventorySummaryRepository';
import type { LocationsQuery, LocationsViewModel } from '../../application/locations/LocationsQuery';
import type { SearchAssetsQuery } from '../../application/search/SearchAssetsQuery';
import { AssetCard } from '../components/AssetCard';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { assetDetailHref } from './AssetDetailNavigation';
import { navigateToAssetTagSearch } from './AssetTagSearchNavigation';
import { BrowsePlaceRow } from './BrowsePlaceRow';
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
  browseContinuationCriteria,
  browseGridCardWidth,
  browseLoadingFlagsForRefresh,
  BrowseFilterToken,
  BrowsePlaceItemViewModel,
  BrowseScope,
  browseScopeToKind,
  cancelPendingBrowseSearch,
  canLoadNextBrowsePage,
  locationRowsFromAssetCards
} from './SearchScreenPresentation';
import { createSearchScreenStyles } from './SearchScreen.styles';

export { SearchHeader } from './BrowseHeader';

type SearchScreenProps = {
  readonly initialScope?: BrowseScope;
  readonly initialQuery?: string;
  readonly initialTagIds?: readonly string[];
  readonly initialLifecycleState?: AssetBrowseLifecycleFilter;
  readonly initialCheckoutState?: AssetBrowseCheckoutFilter;
  readonly initialSort?: AssetBrowseSort;
  readonly addAssetPhotosCommand: AddAssetPhotosCommand;
  readonly assetCheckoutCommand: AssetCheckoutCommand;
  readonly assetDetailQuery: AssetDetailQuery;
  readonly assetLifecycleCommand: AssetLifecycleCommand;
  readonly deleteAssetPhotoCommand: DeleteAssetPhotoCommand;
  readonly inventoryMapQuery: InventoryMapQuery;
  readonly inventoryAssetTagsQuery: InventoryAssetTagsQuery;
  readonly locationsQuery: LocationsQuery;
  readonly photoSelectionQuery: PhotoSelectionQuery;
  readonly searchAssetsQuery: SearchAssetsQuery;
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

const pageSize = 20;

function emptyResults(scope: BrowseScope = 'all', query = ''): BrowseResults {
  return {
    scope,
    query,
    lifecycleState: 'active',
    checkoutState: 'any',
    sort: 'updated_desc',
    tagIds: [],
    assets: [],
    locations: [],
    hasMore: false
  };
}

export function SearchScreen({
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
  inventoryAssetTagsQuery,
  locationsQuery,
  photoSelectionQuery,
  searchAssetsQuery
}: SearchScreenProps) {
  const { fontScale, width } = useWindowDimensions();
  const palette = useAppearancePalette();
  const styles = useMemo(() => createSearchScreenStyles(palette), [palette]);
  const normalizedInitialTags = useMemo(() => uniqueTagIds(initialTagIds), [initialTagIds.join('|')]);
  const [query, setQuery] = useState(initialQuery);
  const [scope, setScope] = useState<BrowseScope>(initialScope);
  const [surface, setSurface] = useState<InventoryMapSurface>('list');
  const [lifecycleState, setLifecycleState] = useState<AssetBrowseLifecycleFilter>(initialLifecycleState);
  const [checkoutState, setCheckoutState] = useState<AssetBrowseCheckoutFilter>(initialCheckoutState);
  const [sort, setSort] = useState<AssetBrowseSort>(initialSort);
  const [selectedTagIds, setSelectedTagIds] = useState<readonly string[]>(normalizedInitialTags);
  const [filterDraft, setFilterDraft] = useState<BrowseDraftFilters>({
    lifecycleState: initialLifecycleState,
    checkoutState: initialCheckoutState,
    tagIds: normalizedInitialTags
  });
  const [filtersExpanded, setFiltersExpanded] = useState(false);
  const [state, setState] = useState<BrowseState>({ status: 'loading', results: emptyResults(initialScope), isInitial: true });
  const [tagFilters, setTagFilters] = useState<readonly AssetTagOptionViewModel[]>([]);
  const [tagFilterStatus, setTagFilterStatus] = useState<'loading' | 'ready' | 'error'>('loading');
  const [inventoryContext, setInventoryContext] = useState<LocationsViewModel>();
  const [inventoryContextStatus, setInventoryContextStatus] = useState<'loading' | 'ready' | 'error'>('loading');
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [isSearchFocused, setIsSearchFocused] = useState(false);
  const requestSequence = useRef(0);
  const queryTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const mapPathStore = useRef(new Map<string, readonly string[]>());
  const searchInputRef = useRef<TextInput>(null);
  const lastRequestedQuery = useRef(initialQuery.trim());
  const latestResults = useRef<BrowseResults>(emptyResults(initialScope));
  const locationCatalog = useRef<Promise<LocationsViewModel> | undefined>(undefined);

  useEffect(() => () => {
    if (queryTimer.current) clearTimeout(queryTimer.current);
  }, []);

  useEffect(() => {
    const nextQuery = initialQuery.trim();
    cancelPendingBrowseSearch(queryTimer, nextQuery);
    const nextTags = uniqueTagIds(initialTagIds);
    setQuery(nextQuery);
    setScope(initialScope);
    setLifecycleState(initialLifecycleState);
    setCheckoutState(initialCheckoutState);
    setSort(initialSort);
    setSelectedTagIds(nextTags);
    setFilterDraft({ lifecycleState: initialLifecycleState, checkoutState: initialCheckoutState, tagIds: nextTags });
    lastRequestedQuery.current = nextQuery;
    void loadFirstPage({
      query: nextQuery,
      scope: initialScope,
      lifecycleState: initialLifecycleState,
      checkoutState: initialCheckoutState,
      sort: initialSort,
      tagIds: nextTags
    });
  }, [
    initialQuery,
    initialScope,
    initialLifecycleState,
    initialCheckoutState,
    initialSort,
    initialTagIds.join('|'),
    locationsQuery,
    searchAssetsQuery
  ]);

  useEffect(() => {
    void loadTagFilters();
    void loadLocationCatalog().catch(() => undefined);
  }, [inventoryAssetTagsQuery, locationsQuery]);

  async function loadTagFilters(): Promise<void> {
    setTagFilterStatus('loading');
    try {
      setTagFilters(await inventoryAssetTagsQuery.execute());
      setTagFilterStatus('ready');
    } catch {
      setTagFilterStatus('error');
    }
  }

  async function loadLocationCatalog(force = false): Promise<LocationsViewModel> {
    if (force || !locationCatalog.current) {
      setInventoryContextStatus('loading');
      locationCatalog.current = locationsQuery.execute();
    }
    try {
      const catalog = await locationCatalog.current;
      setInventoryContext(catalog);
      setInventoryContextStatus('ready');
      return catalog;
    } catch (error) {
      locationCatalog.current = undefined;
      setInventoryContextStatus('error');
      throw error;
    }
  }

  async function loadFirstPage(next: Partial<{
    readonly query: string;
    readonly lifecycleState: AssetBrowseLifecycleFilter;
    readonly checkoutState: AssetBrowseCheckoutFilter;
    readonly scope: BrowseScope;
    readonly sort: AssetBrowseSort;
    readonly tagIds: readonly string[];
  }> = {}): Promise<void> {
    const requestId = nextRequestId(requestSequence);
    const input = {
      query: next.query ?? query,
      lifecycleState: next.lifecycleState ?? lifecycleState,
      checkoutState: next.checkoutState ?? checkoutState,
      scope: next.scope ?? scope,
      sort: next.sort ?? sort,
      tagIds: next.tagIds ?? selectedTagIds
    };
    const previous = latestResults.current;
    const hasPrevious = browseResultCount(previous) > 0;
    const loadingResults = hasPrevious ? previous : emptyResults(input.scope, input.query.trim());
    setIsLoadingMore(false);
    setIsRefreshing(false);
    setState({ status: 'loading', results: loadingResults, isInitial: !hasPrevious });

    try {
      const results = await loadBrowseResults(input);
      if (isCurrentRequest(requestSequence, requestId)) {
        latestResults.current = results;
        setState({ status: 'ready', results });
      }
    } catch (error) {
      if (isCurrentRequest(requestSequence, requestId)) {
        setState({
          status: 'error',
          phase: hasPrevious ? 'replacement' : 'initial',
          message: readableError(error, 'This inventory could not be loaded.'),
          results: loadingResults
        });
      }
    }
  }

  async function loadBrowseResults(input: {
    readonly cursor?: string;
    readonly lifecycleState: AssetBrowseLifecycleFilter;
    readonly checkoutState: AssetBrowseCheckoutFilter;
    readonly query: string;
    readonly scope: BrowseScope;
    readonly sort: AssetBrowseSort;
    readonly tagIds: readonly string[];
  }): Promise<BrowseResults> {
    const results = await searchAssetsQuery.execute({
      query: input.query,
      cursor: input.cursor,
      lifecycleState: input.lifecycleState,
      checkoutState: input.checkoutState,
      kind: browseScopeToKind(input.scope),
      sort: input.sort,
      limit: pageSize,
      tagIds: input.tagIds
    });
    if (input.scope !== 'places') {
      return {
        scope: input.scope,
        query: results.query,
        lifecycleState: input.lifecycleState,
        checkoutState: input.checkoutState,
        sort: input.sort,
        tagIds: input.tagIds,
        assets: results.assets,
        locations: [],
        nextCursor: results.nextCursor,
        hasMore: results.hasMore
      };
    }
    const catalog = await loadLocationCatalog();
    return {
      scope: input.scope,
      query: results.query,
      lifecycleState: input.lifecycleState,
      checkoutState: input.checkoutState,
      sort: input.sort,
      tagIds: input.tagIds,
      assets: [],
      locations: locationRowsFromAssetCards(results.assets, catalog.locations),
      nextCursor: results.nextCursor,
      hasMore: results.hasMore
    };
  }

  async function refreshResults(): Promise<void> {
    const requestId = nextRequestId(requestSequence);
    const loadingFlags = browseLoadingFlagsForRefresh();
    setIsLoadingMore(loadingFlags.isLoadingMore);
    setIsRefreshing(loadingFlags.isRefreshing);
    try {
      if (scope === 'places') await loadLocationCatalog(true);
      const results = await loadBrowseResults({ query: lastRequestedQuery.current, lifecycleState, checkoutState, scope, sort, tagIds: selectedTagIds });
      if (isCurrentRequest(requestSequence, requestId)) {
        latestResults.current = results;
        setState({ status: 'ready', results });
      }
    } catch (error) {
      if (isCurrentRequest(requestSequence, requestId)) {
        setState({ status: 'error', phase: 'replacement', message: readableError(error, 'This inventory could not be refreshed.'), results: latestResults.current });
      }
    } finally {
      if (isCurrentRequest(requestSequence, requestId)) setIsRefreshing(false);
    }
  }

  async function loadNextPage(): Promise<void> {
    const current = latestResults.current;
    const canPage = canLoadNextBrowsePage(
      state.status,
      state.status === 'error' ? state.phase : undefined
    );
    if (!canPage || !current.hasMore || !current.nextCursor || isRefreshing || isLoadingMore) return;
    const requestId = nextRequestId(requestSequence);
    setIsLoadingMore(true);
    try {
      const nextPage = await loadBrowseResults({
        ...browseContinuationCriteria(current),
        cursor: current.nextCursor,
      });
      if (isCurrentRequest(requestSequence, requestId)) {
        const results = {
          ...nextPage,
          assets: [...current.assets, ...nextPage.assets],
          locations: [...current.locations, ...nextPage.locations]
        };
        latestResults.current = results;
        setState({ status: 'ready', results });
      }
    } catch (error) {
      if (isCurrentRequest(requestSequence, requestId)) {
        setState({ status: 'error', phase: 'pagination', message: readableError(error, 'More items could not be loaded.'), results: current });
      }
    } finally {
      if (isCurrentRequest(requestSequence, requestId)) setIsLoadingMore(false);
    }
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

  function updateScope(nextScope: BrowseScope): void {
    const nextQuery = cancelPendingSearch();
    setScope(nextScope);
    void loadFirstPage({ query: nextQuery, scope: nextScope });
  }

  function updateSort(nextSort: AssetBrowseSort): void {
    const nextQuery = cancelPendingSearch();
    setSort(nextSort);
    void loadFirstPage({ query: nextQuery, sort: nextSort });
  }

  function openFilters(expanded: boolean): void {
    if (expanded) setFilterDraft({ lifecycleState, checkoutState, tagIds: selectedTagIds });
    setFiltersExpanded(expanded);
  }

  function applyFilters(filters: BrowseDraftFilters): void {
    const nextQuery = cancelPendingSearch();
    setLifecycleState(filters.lifecycleState);
    setCheckoutState(filters.checkoutState);
    setSelectedTagIds(filters.tagIds);
    setFiltersExpanded(false);
    void loadFirstPage({ ...filters, query: nextQuery });
  }

  function clearFilters(): void {
    applyFilters({ lifecycleState: 'active', checkoutState: 'any', tagIds: [] });
  }

  function removeFilter(token: BrowseFilterToken): void {
    const next = {
      lifecycleState: token.type === 'lifecycle' ? 'active' as const : lifecycleState,
      checkoutState: token.type === 'checkout' ? 'any' as const : checkoutState,
      tagIds: token.type === 'tag' ? selectedTagIds.filter((id) => id !== token.tagId) : selectedTagIds
    };
    applyFilters(next);
  }

  function retryResults(): void {
    if (state.status === 'error' && state.phase === 'pagination') {
      void loadNextPage();
      return;
    }
    void loadFirstPage();
  }

  const listItems = toBrowseListItems(state.results);
  const resultScope = state.results.scope;
  const numColumns = browseColumnCount({ fontScale, scope: resultScope, width });
  const gridCardWidth = browseGridCardWidth(width, numColumns);
  const hasActiveFilters = browseFilterCount({ lifecycleState, checkoutState, tagIds: selectedTagIds }) > 0;
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
          deleteAssetPhotoCommand={deleteAssetPhotoCommand}
          inventoryMapQuery={inventoryMapQuery}
          pathStore={mapPathStore}
          photoSelectionQuery={photoSelectionQuery}
          selectedSurface={surface}
          onChangeSurface={setSurface}
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
        keyboardShouldPersistTaps="handled"
        numColumns={numColumns}
        refreshing={isRefreshing}
        onEndReached={() => void loadNextPage()}
        onEndReachedThreshold={0.55}
        onRefresh={() => void refreshResults()}
        ListHeaderComponent={
          <SearchHeader
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
            statusMessage={state.status === 'error' && state.phase === 'replacement' ? state.message : undefined}
            submittedQuery={state.results.query}
            tagFilters={tagFilters}
            tagFilterStatus={tagFilterStatus}
            onApplyFilters={applyFilters}
            onChangeDraftCheckoutState={(value) => setFilterDraft((draft) => ({ ...draft, checkoutState: value }))}
            onChangeDraftLifecycleState={(value) => setFilterDraft((draft) => ({ ...draft, lifecycleState: value }))}
            onChangeDraftTagIds={(value) => setFilterDraft((draft) => ({ ...draft, tagIds: value }))}
            onChangeQuery={scheduleSearch}
            onChangeScope={updateScope}
            onChangeSort={updateSort}
            onChangeSurface={setSurface}
            onClearFilters={clearFilters}
            onClearQuery={clearSearch}
            onRemoveFilter={removeFilter}
            onRetryResults={retryResults}
            onRetryInventoryContext={() => void loadLocationCatalog(true).catch(() => undefined)}
            onRetryTags={() => void loadTagFilters()}
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
              onAdd={() => router.navigate('/add')}
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
          <BrowsePlaceRow location={item.location} palette={palette} onPress={() => router.push(`/locations/${item.location.id}`)} />
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

function browseResultCount(results: BrowseResults): number {
  return results.assets.length + results.locations.length;
}

function uniqueTagIds(tagIds: readonly string[]): readonly string[] {
  return [...new Set(tagIds.map((id) => id.trim()).filter(Boolean))];
}

function readableError(_error: unknown, fallback: string): string {
  return fallback;
}

function nextRequestId(requestSequence: { current: number }): number {
  requestSequence.current += 1;
  return requestSequence.current;
}

function isCurrentRequest(requestSequence: { readonly current: number }, requestId: number): boolean {
  return requestSequence.current === requestId;
}
