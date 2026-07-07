import { useCallback, useEffect, useRef, useState } from 'react';
import type { RefObject } from 'react';
import { router, useFocusEffect } from 'expo-router';
import {
  ActivityIndicator,
  FlatList,
  Image,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Search, SlidersHorizontal, X } from 'lucide-react-native';
import type { AddAssetPhotosCommand } from '../../application/assets/AddAssetPhotosCommand';
import type { AssetCheckoutCommand } from '../../application/assets/AssetCheckoutCommand';
import type { AssetLifecycleCommand } from '../../application/assets/AssetLifecycleCommand';
import type { DeleteAssetPhotoCommand } from '../../application/assets/DeleteAssetPhotoCommand';
import type {
  AssetBrowseCheckoutFilter,
  AssetBrowseLifecycleFilter,
  AssetBrowseSort
} from '../../application/home/InventorySummaryRepository';
import type { AssetCardViewModel } from '../../application/assets/AssetViewModels';
import type { AssetDetailQuery } from '../../application/assets/AssetDetailQuery';
import type { InventoryMapQuery } from '../../application/assets/InventoryMapQuery';
import {
  LocationBrowserItemViewModel,
  LocationsQuery
} from '../../application/locations/LocationsQuery';
import type { PhotoSelectionQuery } from '../../application/add/PhotoSelectionQuery';
import { SearchAssetsQuery } from '../../application/search/SearchAssetsQuery';
import { AssetCard } from '../components/AssetCard';
import { colors, radius, spacing } from '../theme/tokens';
import {
  BrowseScope,
  browseScopeToKind,
  buildBrowseScopeOptions,
  focusSearchInput,
  locationRowsFromAssetCards,
  searchResultSummaryLabel
} from './SearchScreenPresentation';
import { BrowseSurfaceControl, InventoryMapScreen } from './InventoryMapScreen';
import type { InventoryMapSurface } from './InventoryMapPresentation';

type SearchScreenProps = {
  readonly initialScope?: BrowseScope;
  readonly addAssetPhotosCommand: AddAssetPhotosCommand;
  readonly assetCheckoutCommand: AssetCheckoutCommand;
  readonly assetDetailQuery: AssetDetailQuery;
  readonly assetLifecycleCommand: AssetLifecycleCommand;
  readonly deleteAssetPhotoCommand: DeleteAssetPhotoCommand;
  readonly inventoryMapQuery: InventoryMapQuery;
  readonly locationsQuery: LocationsQuery;
  readonly photoSelectionQuery: PhotoSelectionQuery;
  readonly searchAssetsQuery: SearchAssetsQuery;
};

type BrowseResults = {
  readonly scope: BrowseScope;
  readonly query: string;
  readonly assets: readonly AssetCardViewModel[];
  readonly locations: readonly LocationBrowserItemViewModel[];
  readonly nextCursor?: string;
  readonly hasMore: boolean;
};

type BrowseState =
  | { readonly status: 'loading'; readonly results: BrowseResults }
  | { readonly status: 'ready'; readonly results: BrowseResults }
  | { readonly status: 'error'; readonly message: string; readonly results: BrowseResults };

type BrowseListItem =
  | { readonly type: 'asset'; readonly asset: AssetCardViewModel }
  | { readonly type: 'place'; readonly location: LocationBrowserItemViewModel };

const pageSize = 20;

const emptyResults: BrowseResults = {
  scope: 'all',
  query: '',
  assets: [],
  locations: [],
  hasMore: false
};

export function SearchScreen({
  initialScope = 'all',
  addAssetPhotosCommand,
  assetCheckoutCommand,
  assetDetailQuery,
  assetLifecycleCommand,
  deleteAssetPhotoCommand,
  inventoryMapQuery,
  locationsQuery,
  photoSelectionQuery,
  searchAssetsQuery
}: SearchScreenProps) {
  const [query, setQuery] = useState('');
  const [scope, setScope] = useState<BrowseScope>(initialScope);
  const mapPathStore = useRef(new Map<string, readonly string[]>());
  const [surface, setSurface] = useState<InventoryMapSurface>('list');
  const [lifecycleState, setLifecycleState] = useState<AssetBrowseLifecycleFilter>('active');
  const [checkoutState, setCheckoutState] = useState<AssetBrowseCheckoutFilter>('any');
  const [sort, setSort] = useState<AssetBrowseSort>('updated_desc');
  const [state, setState] = useState<BrowseState>({ status: 'loading', results: emptyResults });
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const requestSequence = useRef(0);
  const searchInputRef = useRef<TextInput>(null);
  const [isSearchFocused, setIsSearchFocused] = useState(false);

  useFocusEffect(
    useCallback(() => {
      const focusTimer = setTimeout(() => {
        focusSearchInput(searchInputRef);
      }, 120);

      return () => {
        clearTimeout(focusTimer);
      };
    }, [])
  );

  useEffect(() => {
    setScope(initialScope);
    setQuery('');
    void loadFirstPage({ query: '', scope: initialScope });
  }, [initialScope, locationsQuery, searchAssetsQuery]);

  async function loadFirstPage(next: {
    readonly query?: string;
    readonly lifecycleState?: AssetBrowseLifecycleFilter;
    readonly checkoutState?: AssetBrowseCheckoutFilter;
    readonly scope?: BrowseScope;
    readonly sort?: AssetBrowseSort;
  } = {}): Promise<void> {
    const requestId = nextRequestId(requestSequence);
    const nextQuery = next.query ?? query;
    const nextLifecycleState = next.lifecycleState ?? lifecycleState;
    const nextCheckoutState = next.checkoutState ?? checkoutState;
    const nextScope = next.scope ?? scope;
    const nextSort = next.sort ?? sort;
    const loadingResults: BrowseResults = {
      scope: nextScope,
      query: nextQuery.trim(),
      assets: [],
      locations: [],
      hasMore: false
    };

    setIsLoadingMore(false);
    setIsRefreshing(false);
    setState({ status: 'loading', results: loadingResults });

    try {
      const results = await loadBrowseResults({
        query: nextQuery,
        lifecycleState: nextLifecycleState,
        checkoutState: nextCheckoutState,
        scope: nextScope,
        sort: nextSort
      });
      if (isCurrentRequest(requestSequence, requestId)) {
        setState({ status: 'ready', results });
      }
    } catch (error) {
      if (isCurrentRequest(requestSequence, requestId)) {
        setState({
          status: 'error',
          message: readableError(error, 'Browse failed.'),
          results: loadingResults
        });
      }
    }
  }

  async function loadBrowseResults({
    cursor,
    lifecycleState: nextLifecycleState,
    checkoutState: nextCheckoutState,
    query: nextQuery,
    scope: nextScope,
    sort: nextSort
  }: {
    readonly cursor?: string;
    readonly lifecycleState: AssetBrowseLifecycleFilter;
    readonly checkoutState: AssetBrowseCheckoutFilter;
    readonly query: string;
    readonly scope: BrowseScope;
    readonly sort: AssetBrowseSort;
  }): Promise<BrowseResults> {
    if (nextScope === 'places') {
      const [results, locations] = await Promise.all([
        searchAssetsQuery.execute({
          query: nextQuery,
          cursor,
          lifecycleState: nextLifecycleState,
          checkoutState: nextCheckoutState,
          kind: browseScopeToKind(nextScope),
          sort: nextSort,
          limit: pageSize
        }),
        locationsQuery.execute()
      ]);
      return {
        scope: nextScope,
        query: results.query,
        assets: [],
        locations: locationRowsFromAssetCards(results.assets, locations.locations),
        nextCursor: results.nextCursor,
        hasMore: results.hasMore
      };
    }

    const results = await searchAssetsQuery.execute({
      query: nextQuery,
      cursor,
      lifecycleState: nextLifecycleState,
      checkoutState: nextCheckoutState,
      kind: browseScopeToKind(nextScope),
      sort: nextSort,
      limit: pageSize
    });

    return {
      scope: nextScope,
      query: results.query,
      assets: results.assets,
      locations: [],
      nextCursor: results.nextCursor,
      hasMore: results.hasMore
    };
  }

  async function refreshResults(): Promise<void> {
    const requestId = nextRequestId(requestSequence);
    setIsRefreshing(true);

    try {
      const results = await loadBrowseResults({
        query: state.results.query,
        lifecycleState,
        checkoutState,
        scope,
        sort
      });
      if (isCurrentRequest(requestSequence, requestId)) {
        setQuery(state.results.query);
        setState({ status: 'ready', results });
      }
    } catch (error) {
      if (isCurrentRequest(requestSequence, requestId)) {
        setState({
          status: 'error',
          message: readableError(error, 'Refresh failed.'),
          results: state.results
        });
      }
    } finally {
      if (isCurrentRequest(requestSequence, requestId)) {
        setIsRefreshing(false);
      }
    }
  }

  async function loadNextPage(): Promise<void> {
    if (
      state.status === 'loading' ||
      !state.results.hasMore ||
      !state.results.nextCursor ||
      isRefreshing ||
      isLoadingMore
    ) {
      return;
    }

    const requestId = nextRequestId(requestSequence);
    setIsLoadingMore(true);

    try {
      const nextPage = await loadBrowseResults({
        query: state.results.query,
        cursor: state.results.nextCursor,
        lifecycleState,
        checkoutState,
        scope,
        sort
      });
      if (isCurrentRequest(requestSequence, requestId)) {
        setState({
          status: 'ready',
          results: {
            ...nextPage,
            assets: [...state.results.assets, ...nextPage.assets],
            locations: [...state.results.locations, ...nextPage.locations]
          }
        });
      }
    } catch (error) {
      if (isCurrentRequest(requestSequence, requestId)) {
        setState({
          status: 'error',
          message: readableError(error, 'Could not load more assets.'),
          results: state.results
        });
      }
    } finally {
      if (isCurrentRequest(requestSequence, requestId)) {
        setIsLoadingMore(false);
      }
    }
  }

  function submitSearch(): void {
    focusSearchInput(searchInputRef);
    void loadFirstPage({ query });
  }

  function clearSearch(): void {
    setQuery('');
    void loadFirstPage({ query: '' });
  }

  function updateScope(nextScope: BrowseScope): void {
    setScope(nextScope);
    void loadFirstPage({ scope: nextScope });
  }

  function updateLifecycleState(nextLifecycleState: AssetBrowseLifecycleFilter): void {
    setLifecycleState(nextLifecycleState);
    void loadFirstPage({ lifecycleState: nextLifecycleState });
  }

  function updateCheckoutState(nextCheckoutState: AssetBrowseCheckoutFilter): void {
    setCheckoutState(nextCheckoutState);
    void loadFirstPage({ checkoutState: nextCheckoutState });
  }

  function updateSort(nextSort: AssetBrowseSort): void {
    setSort(nextSort);
    void loadFirstPage({ sort: nextSort });
  }

  const listItems = toBrowseListItems(state.results);
  const isPlacesScope = scope === 'places';

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
        key={isPlacesScope ? 'places' : 'assets'}
        data={listItems}
        keyExtractor={keyBrowseListItem}
        columnWrapperStyle={isPlacesScope ? undefined : styles.cardRow}
        contentContainerStyle={styles.content}
        keyboardShouldPersistTaps="handled"
        numColumns={isPlacesScope ? 1 : 2}
        refreshing={isRefreshing}
        onEndReached={loadNextPage}
        onEndReachedThreshold={0.55}
        onRefresh={refreshResults}
        ListFooterComponent={
          isLoadingMore ? (
            <View style={styles.footer}>
              <ActivityIndicator color={colors.accent} />
            </View>
          ) : null
        }
        ListHeaderComponent={
          <SearchHeader
            isLoading={state.status === 'loading'}
            lifecycleState={lifecycleState}
            checkoutState={checkoutState}
            query={query}
            resultCount={listItems.length}
            scope={scope}
            selectedSurface={surface}
            searchInputRef={searchInputRef}
            searchInputFocused={isSearchFocused}
            sort={sort}
            statusMessage={state.status === 'error' ? state.message : undefined}
            submittedQuery={state.results.query}
            onChangeSurface={setSurface}
            onChangeLifecycleState={updateLifecycleState}
            onChangeCheckoutState={updateCheckoutState}
            onChangeQuery={setQuery}
            onChangeScope={updateScope}
            onChangeSort={updateSort}
            onClearQuery={clearSearch}
            onSearchBlur={() => setIsSearchFocused(false)}
            onSearchFocus={() => setIsSearchFocused(true)}
            onSubmit={submitSearch}
          />
        }
        ListEmptyComponent={
          state.status === 'loading' ? null : (
            <EmptyBrowseState query={state.results.query} scope={scope} onClear={clearSearch} />
          )
        }
        renderItem={({ item }) => {
          if (item.type === 'place') {
            return <PlaceRow location={item.location} />;
          }

          return <AssetCard asset={item.asset} onPress={() => router.push(`/assets/${item.asset.id}`)} />;
        }}
      />
    </SafeAreaView>
  );
}

export function SearchHeader({
  isLoading,
  lifecycleState,
  checkoutState,
  query,
  resultCount,
  scope,
  selectedSurface,
  searchInputRef,
  searchInputFocused,
  sort,
  statusMessage,
  submittedQuery,
  onChangeSurface,
  onChangeLifecycleState,
  onChangeCheckoutState,
  onChangeQuery,
  onChangeScope,
  onChangeSort,
  onClearQuery,
  onSearchBlur,
  onSearchFocus,
  onSubmit
}: {
  readonly isLoading: boolean;
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly checkoutState: AssetBrowseCheckoutFilter;
  readonly query: string;
  readonly resultCount: number;
  readonly scope: BrowseScope;
  readonly selectedSurface: InventoryMapSurface;
  readonly searchInputRef: RefObject<TextInput | null>;
  readonly searchInputFocused: boolean;
  readonly sort: AssetBrowseSort;
  readonly statusMessage?: string;
  readonly submittedQuery: string;
  readonly onChangeSurface: (surface: InventoryMapSurface) => void;
  readonly onChangeLifecycleState: (lifecycleState: AssetBrowseLifecycleFilter) => void;
  readonly onChangeCheckoutState: (checkoutState: AssetBrowseCheckoutFilter) => void;
  readonly onChangeQuery: (query: string) => void;
  readonly onChangeScope: (scope: BrowseScope) => void;
  readonly onChangeSort: (sort: AssetBrowseSort) => void;
  readonly onClearQuery: () => void;
  readonly onSearchBlur: () => void;
  readonly onSearchFocus: () => void;
  readonly onSubmit: () => void;
}) {
  const summaryLabel = searchResultSummaryLabel({
    lifecycleState,
    query: submittedQuery,
    resultCount,
    scope,
    sort
  });

  return (
    <View>
      <View style={styles.headerTopRow}>
        <View style={styles.titleBlock}>
          <Text style={styles.title}>Browse</Text>
          <Text numberOfLines={1} style={styles.resultCount}>{summaryLabel}</Text>
        </View>
        <BrowseSurfaceControl selectedSurface={selectedSurface} onChangeSurface={onChangeSurface} />
      </View>
      <View style={[styles.searchBar, searchInputFocused ? styles.searchBarFocused : null]}>
        <Search color={colors.textMuted} size={19} strokeWidth={2.5} />
        <TextInput
          accessibilityLabel="Search inventory"
          autoCapitalize="none"
          ref={searchInputRef}
          onChangeText={onChangeQuery}
          onBlur={onSearchBlur}
          onFocus={onSearchFocus}
          onSubmitEditing={onSubmit}
          placeholder="Search things, places, boxes"
          placeholderTextColor={colors.textMuted}
          returnKeyType="search"
          style={styles.searchInput}
          value={query}
        />
        {query.length > 0 ? (
          <Pressable
            accessibilityLabel="Clear search"
            accessibilityRole="button"
            hitSlop={10}
            onPress={onClearQuery}
            style={styles.clearButton}
          >
            <X color={colors.textMuted} size={18} strokeWidth={2.5} />
          </Pressable>
        ) : null}
        {isLoading ? <ActivityIndicator color={colors.accent} size="small" /> : null}
      </View>
      <ScopeControl selectedScope={scope} onChangeScope={onChangeScope} />
      <RefinementBar
        lifecycleState={lifecycleState}
        checkoutState={checkoutState}
        scope={scope}
        searchMode={submittedQuery.trim().length > 0}
        sort={sort}
        onChangeLifecycleState={onChangeLifecycleState}
        onChangeCheckoutState={onChangeCheckoutState}
        onChangeSort={onChangeSort}
      />
      {statusMessage ? <Text style={styles.errorText}>{statusMessage}</Text> : null}
    </View>
  );
}

function ScopeControl({
  selectedScope,
  onChangeScope
}: {
  readonly selectedScope: BrowseScope;
  readonly onChangeScope: (scope: BrowseScope) => void;
}) {
  return (
    <View style={styles.scopeControl}>
      {buildBrowseScopeOptions().map((option) => {
        const selected = option.value === selectedScope;
        return (
          <Pressable
            accessibilityRole="button"
            accessibilityState={{ selected }}
            key={option.value}
            onPress={() => onChangeScope(option.value)}
            style={[styles.scopeButton, selected ? styles.scopeButtonSelected : null]}
          >
            <Text style={[styles.scopeText, selected ? styles.scopeTextSelected : null]}>
              {option.label}
            </Text>
          </Pressable>
        );
      })}
    </View>
  );
}

function RefinementBar({
  checkoutState,
  lifecycleState,
  searchMode,
  scope,
  sort,
  onChangeCheckoutState,
  onChangeLifecycleState,
  onChangeSort
}: {
  readonly checkoutState: AssetBrowseCheckoutFilter;
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly searchMode: boolean;
  readonly scope: BrowseScope;
  readonly sort: AssetBrowseSort;
  readonly onChangeCheckoutState: (checkoutState: AssetBrowseCheckoutFilter) => void;
  readonly onChangeLifecycleState: (lifecycleState: AssetBrowseLifecycleFilter) => void;
  readonly onChangeSort: (sort: AssetBrowseSort) => void;
}) {
  return (
    <ScrollView
      horizontal
      showsHorizontalScrollIndicator={false}
      contentContainerStyle={styles.refinementBar}
    >
      <View style={styles.refinementIcon}>
        <SlidersHorizontal color={colors.textMuted} size={17} strokeWidth={2.5} />
      </View>
      <FilterChip
        label="Active"
        selected={lifecycleState === 'active'}
        onPress={() => onChangeLifecycleState('active')}
      />
      <FilterChip
        label="Archived"
        selected={lifecycleState === 'archived'}
        onPress={() => onChangeLifecycleState('archived')}
      />
      <FilterChip
        label="All status"
        selected={lifecycleState === 'all'}
        onPress={() => onChangeLifecycleState('all')}
      />
      <View style={styles.refinementDivider} />
      <FilterChip
        label="Any checkout"
        selected={checkoutState === 'any'}
        onPress={() => onChangeCheckoutState('any')}
      />
      <FilterChip
        label="Checked out"
        selected={checkoutState === 'checked_out'}
        onPress={() => onChangeCheckoutState('checked_out')}
      />
      <FilterChip
        label="Available"
        selected={checkoutState === 'available'}
        onPress={() => onChangeCheckoutState('available')}
      />
      {!searchMode ? (
        <>
          <View style={styles.refinementDivider} />
          <FilterChip
            label="Recent"
            selected={sort === 'updated_desc'}
            onPress={() => onChangeSort('updated_desc')}
          />
          <FilterChip
            label="Stable"
            selected={sort === 'id_asc'}
            onPress={() => onChangeSort('id_asc')}
          />
        </>
      ) : null}
    </ScrollView>
  );
}

function FilterChip({
  label,
  selected,
  onPress
}: {
  readonly label: string;
  readonly selected: boolean;
  readonly onPress: () => void;
}) {
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityState={{ selected }}
      onPress={onPress}
      style={[styles.filterChip, selected ? styles.filterChipSelected : null]}
    >
      <Text style={[styles.filterChipText, selected ? styles.filterChipTextSelected : null]}>
        {label}
      </Text>
    </Pressable>
  );
}

function PlaceRow({ location }: { readonly location: LocationBrowserItemViewModel }) {
  return (
    <Pressable
      accessibilityRole="button"
      onPress={() => router.push(`/locations/${location.id}`)}
      style={styles.placeRow}
    >
      <View style={styles.placeImageFrame}>
        {location.photo ? (
          <Image
            accessibilityIgnoresInvertColors
            source={{ uri: location.photo.uri, headers: location.photo.headers }}
            style={styles.placeImage}
          />
        ) : (
          <Text style={styles.placeImageLabel}>Place</Text>
        )}
      </View>
      <View style={styles.placeBody}>
        <View style={styles.placeHeader}>
          <Text numberOfLines={1} style={styles.placeTitle}>{location.title}</Text>
          <Text style={location.photoLabel === 'Photo ready' ? styles.photoReady : styles.photoNeeded}>
            {location.photoLabel}
          </Text>
        </View>
        {location.description ? (
          <Text numberOfLines={2} style={styles.placeDescription}>{location.description}</Text>
        ) : null}
        <Text style={styles.placeCount}>{location.containedAssetCountLabel}</Text>
        <Text numberOfLines={1} style={styles.recentAssetLabel}>{location.recentAssetLabel}</Text>
      </View>
    </Pressable>
  );
}

function EmptyBrowseState({
  query,
  scope,
  onClear
}: {
  readonly query: string;
  readonly scope: BrowseScope;
  readonly onClear: () => void;
}) {
  const hasQuery = query.trim().length > 0;
  return (
    <View style={styles.emptyPanel}>
      <Text style={styles.emptyTitle}>
        {hasQuery ? `No matches for "${query}"` : `No ${scope === 'all' ? 'things' : scope} yet`}
      </Text>
      <Text style={styles.emptyText}>
        {hasQuery
          ? 'Try a broader search or switch scopes.'
          : 'New inventory activity will appear here.'}
      </Text>
      {hasQuery ? (
        <Pressable accessibilityRole="button" onPress={onClear} style={styles.emptyAction}>
          <Text style={styles.emptyActionText}>Clear search</Text>
        </Pressable>
      ) : null}
    </View>
  );
}

function toBrowseListItems(results: BrowseResults): readonly BrowseListItem[] {
  if (results.scope === 'places') {
    return results.locations.map((location) => ({ type: 'place', location }));
  }

  return results.assets.map((asset) => ({ type: 'asset', asset }));
}

function keyBrowseListItem(item: BrowseListItem): string {
  return item.type === 'place' ? `place:${item.location.id}` : `asset:${item.asset.id}`;
}

function readableError(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

function nextRequestId(requestSequence: { current: number }): number {
  requestSequence.current += 1;
  return requestSequence.current;
}

function isCurrentRequest(
  requestSequence: { readonly current: number },
  requestId: number
): boolean {
  return requestSequence.current === requestId;
}

const styles = StyleSheet.create({
  shell: {
    flex: 1,
    backgroundColor: colors.background
  },
  content: {
    paddingHorizontal: spacing.md,
    paddingTop: spacing.sm,
    paddingBottom: spacing.xl
  },
  headerTopRow: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.md,
    marginBottom: spacing.xs
  },
  titleBlock: {
    flex: 1,
    minWidth: 0
  },
  title: {
    color: colors.text,
    fontSize: 25,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 30
  },
  searchBar: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    minHeight: 44,
    paddingHorizontal: spacing.sm
  },
  searchBarFocused: {
    borderColor: colors.focusRing,
    borderWidth: 2
  },
  searchInput: {
    color: colors.text,
    flex: 1,
    fontSize: 15,
    minHeight: 44,
    paddingVertical: 0
  },
  clearButton: {
    alignItems: 'center',
    minHeight: 44,
    minWidth: 32,
    justifyContent: 'center'
  },
  scopeControl: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.md,
    flexDirection: 'row',
    gap: 2,
    marginTop: spacing.xs,
    padding: 2
  },
  scopeButton: {
    alignItems: 'center',
    borderRadius: radius.sm,
    flex: 1,
    minHeight: 44,
    justifyContent: 'center',
    paddingHorizontal: spacing.xs
  },
  scopeButtonSelected: {
    backgroundColor: colors.surface,
    shadowColor: '#000000',
    shadowOpacity: 0.08,
    shadowRadius: 8,
    shadowOffset: { width: 0, height: 2 }
  },
  scopeText: {
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 0
  },
  scopeTextSelected: {
    color: colors.text
  },
  refinementBar: {
    alignItems: 'center',
    gap: spacing.xs,
    paddingBottom: spacing.xs,
    paddingTop: spacing.xs
  },
  refinementIcon: {
    alignItems: 'center',
    height: 44,
    justifyContent: 'center',
    width: 28
  },
  refinementDivider: {
    backgroundColor: colors.border,
    height: 22,
    marginHorizontal: spacing.xs,
    width: 1
  },
  filterChip: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    minHeight: 44,
    justifyContent: 'center',
    paddingHorizontal: spacing.sm
  },
  filterChipSelected: {
    backgroundColor: colors.selected,
    borderColor: colors.accent
  },
  filterChipText: {
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 0
  },
  filterChipTextSelected: {
    color: colors.accentStrong
  },
  errorText: {
    color: colors.warning,
    fontSize: 14,
    lineHeight: 20,
    marginBottom: spacing.sm
  },
  resultCount: {
    color: colors.textMuted,
    fontSize: 11,
    fontWeight: '800',
    letterSpacing: 0,
    marginTop: 0
  },
  emptyPanel: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    gap: spacing.xs,
    padding: spacing.md
  },
  emptyTitle: {
    color: colors.text,
    fontSize: 18,
    fontWeight: '900',
    letterSpacing: 0
  },
  emptyText: {
    color: colors.textMuted,
    fontSize: 15,
    lineHeight: 22
  },
  emptyAction: {
    alignSelf: 'flex-start',
    backgroundColor: colors.action,
    borderRadius: radius.sm,
    marginTop: spacing.sm,
    minHeight: 38,
    justifyContent: 'center',
    paddingHorizontal: spacing.md
  },
  emptyActionText: {
    color: colors.onAction,
    fontSize: 14,
    fontWeight: '900',
    letterSpacing: 0
  },
  footer: {
    alignItems: 'center',
    paddingVertical: spacing.md
  },
  cardRow: {
    gap: spacing.sm,
    marginBottom: spacing.sm
  },
  placeRow: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    marginBottom: spacing.sm,
    overflow: 'hidden',
    padding: spacing.sm
  },
  placeImageFrame: {
    alignItems: 'center',
    aspectRatio: 1,
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.sm,
    justifyContent: 'center',
    overflow: 'hidden',
    width: 92
  },
  placeImageLabel: {
    color: colors.accentStrong,
    fontSize: 18,
    fontWeight: '900',
    letterSpacing: 0
  },
  placeImage: {
    height: '100%',
    width: '100%'
  },
  placeBody: {
    flex: 1,
    gap: 3,
    justifyContent: 'center',
    minWidth: 0
  },
  placeHeader: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.sm
  },
  placeTitle: {
    color: colors.text,
    flex: 1,
    fontSize: 17,
    fontWeight: '900',
    letterSpacing: 0
  },
  placeDescription: {
    color: colors.textMuted,
    fontSize: 13,
    lineHeight: 18
  },
  placeCount: {
    color: colors.accentStrong,
    fontSize: 13,
    fontWeight: '900',
    letterSpacing: 0
  },
  recentAssetLabel: {
    color: colors.textMuted,
    fontSize: 12,
    lineHeight: 17
  },
  photoReady: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.sm,
    color: colors.accentStrong,
    fontSize: 11,
    fontWeight: '800',
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.xs,
    paddingVertical: 4
  },
  photoNeeded: {
    backgroundColor: colors.warningSurface,
    borderRadius: radius.sm,
    color: colors.warning,
    fontSize: 11,
    fontWeight: '800',
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.xs,
    paddingVertical: 4
  }
});
