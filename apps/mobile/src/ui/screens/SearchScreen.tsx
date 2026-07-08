import { useCallback, useEffect, useRef, useState } from 'react';
import type { ReactNode, RefObject } from 'react';
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
import type { InventoryAssetTagsQuery, AssetTagOptionViewModel } from '../../application/assets/InventoryAssetTagsQuery';
import { SearchAssetsQuery } from '../../application/search/SearchAssetsQuery';
import { AssetCard } from '../components/AssetCard';
import { colors, radius, spacing } from '../theme/tokens';
import {
  BrowseScope,
  browseScopeToKind,
  buildBrowseScopeOptions,
  focusSearchInput,
  locationRowsFromAssetCards,
  searchResultSummaryLabel,
  shouldAutoFocusSearchInput
} from './SearchScreenPresentation';
import { BrowseSurfaceControl, InventoryMapScreen } from './InventoryMapScreen';
import type { InventoryMapSurface } from './InventoryMapPresentation';
import { navigateToAssetTagSearch } from './AssetTagSearchNavigation';

type SearchScreenProps = {
  readonly initialScope?: BrowseScope;
  readonly initialQuery?: string;
  readonly initialTagIds?: readonly string[];
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
const compactControlHitSlop = { top: 6, bottom: 6, left: 6, right: 6 } as const;

const emptyResults: BrowseResults = {
  scope: 'all',
  query: '',
  assets: [],
  locations: [],
  hasMore: false
};

export function SearchScreen({
  initialScope = 'all',
  initialQuery = '',
  initialTagIds = [],
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
  const [query, setQuery] = useState(initialQuery);
  const [scope, setScope] = useState<BrowseScope>(initialScope);
  const mapPathStore = useRef(new Map<string, readonly string[]>());
  const [surface, setSurface] = useState<InventoryMapSurface>('list');
  const [lifecycleState, setLifecycleState] = useState<AssetBrowseLifecycleFilter>('active');
  const [checkoutState, setCheckoutState] = useState<AssetBrowseCheckoutFilter>('any');
  const [sort, setSort] = useState<AssetBrowseSort>('updated_desc');
  const [state, setState] = useState<BrowseState>({ status: 'loading', results: emptyResults });
  const [tagFilters, setTagFilters] = useState<readonly AssetTagOptionViewModel[]>([]);
  const [selectedTagIds, setSelectedTagIds] = useState<readonly string[]>(initialTagIds);
  const [filtersExpanded, setFiltersExpanded] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const requestSequence = useRef(0);
  const searchInputRef = useRef<TextInput>(null);
  const [isSearchFocused, setIsSearchFocused] = useState(false);

  useFocusEffect(
    useCallback(() => {
      if (!shouldAutoFocusSearchInput(initialTagIds)) {
        return undefined;
      }
      const focusTimer = setTimeout(() => {
        focusSearchInput(searchInputRef);
      }, 120);

      return () => {
        clearTimeout(focusTimer);
      };
    }, [initialTagIds.join('|')])
  );

  useEffect(() => {
    const nextQuery = initialQuery.trim();
    const nextTagIds = uniqueTagIds(initialTagIds);
    setScope(initialScope);
    setQuery(nextQuery);
    setSelectedTagIds(nextTagIds);
    void loadFirstPage({ query: nextQuery, scope: initialScope, tagIds: nextTagIds });
  }, [initialQuery, initialScope, initialTagIds.join('|'), locationsQuery, searchAssetsQuery]);

  useEffect(() => {
    let isCurrent = true;

    inventoryAssetTagsQuery
      .execute()
      .then((tags) => {
        if (isCurrent) {
          setTagFilters(tags);
        }
      })
      .catch(() => {
        if (isCurrent) {
          setTagFilters([]);
        }
      });

    return () => {
      isCurrent = false;
    };
  }, [inventoryAssetTagsQuery]);

  async function loadFirstPage(next: {
    readonly query?: string;
    readonly lifecycleState?: AssetBrowseLifecycleFilter;
    readonly checkoutState?: AssetBrowseCheckoutFilter;
    readonly scope?: BrowseScope;
    readonly sort?: AssetBrowseSort;
    readonly tagIds?: readonly string[];
  } = {}): Promise<void> {
    const requestId = nextRequestId(requestSequence);
    const nextQuery = next.query ?? query;
    const nextLifecycleState = next.lifecycleState ?? lifecycleState;
    const nextCheckoutState = next.checkoutState ?? checkoutState;
    const nextScope = next.scope ?? scope;
    const nextSort = next.sort ?? sort;
    const nextTagIds = next.tagIds ?? selectedTagIds;
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
        sort: nextSort,
        tagIds: nextTagIds
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
    sort: nextSort,
    tagIds
  }: {
    readonly cursor?: string;
    readonly lifecycleState: AssetBrowseLifecycleFilter;
    readonly checkoutState: AssetBrowseCheckoutFilter;
    readonly query: string;
    readonly scope: BrowseScope;
    readonly sort: AssetBrowseSort;
    readonly tagIds: readonly string[];
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
          limit: pageSize,
          tagIds
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
      limit: pageSize,
      tagIds
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
        sort,
        tagIds: selectedTagIds
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
        sort,
        tagIds: selectedTagIds
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

  function searchByTag(tag: AssetTagOptionViewModel): void {
    const nextTagIds = toggleTagId(selectedTagIds, tag.id);
    setSelectedTagIds(nextTagIds);
    void loadFirstPage({ tagIds: nextTagIds });
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

  function clearFilters(): void {
    setScope('all');
    setLifecycleState('active');
    setCheckoutState('any');
    setSort('updated_desc');
    setSelectedTagIds([]);
    void loadFirstPage({
      scope: 'all',
      lifecycleState: 'active',
      checkoutState: 'any',
      sort: 'updated_desc',
      tagIds: []
    });
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
            filtersExpanded={filtersExpanded}
            query={query}
            resultCount={listItems.length}
            scope={scope}
            selectedSurface={surface}
            searchInputRef={searchInputRef}
            searchInputFocused={isSearchFocused}
            sort={sort}
            statusMessage={state.status === 'error' ? state.message : undefined}
            submittedQuery={state.results.query}
            selectedTagIds={selectedTagIds}
            tagFilters={tagFilters}
            onChangeSurface={setSurface}
            onChangeLifecycleState={updateLifecycleState}
            onChangeCheckoutState={updateCheckoutState}
            onChangeQuery={setQuery}
            onChangeScope={updateScope}
            onChangeSort={updateSort}
            onClearQuery={clearSearch}
            onClearFilters={clearFilters}
            onToggleTag={searchByTag}
            onSearchBlur={() => setIsSearchFocused(false)}
            onSearchFocus={() => setIsSearchFocused(true)}
            onSubmit={submitSearch}
            onToggleFilters={setFiltersExpanded}
          />
        }
        ListEmptyComponent={
          state.status === 'loading' ? null : (
            <EmptyBrowseState
              query={state.results.query}
              scope={scope}
              filtersActive={hasActiveFilters({ checkoutState, lifecycleState, scope, sort, tagCount: selectedTagIds.length })}
              onClearFilters={clearFilters}
              onClearSearch={clearSearch}
            />
          )
        }
        renderItem={({ item }) => {
          if (item.type === 'place') {
            return <PlaceRow location={item.location} />;
          }

          return (
            <AssetCard
              asset={item.asset}
              onPress={() => router.push(`/assets/${item.asset.id}`)}
              onTagPress={(tag) => navigateToAssetTagSearch(router, tag)}
            />
          );
        }}
      />
    </SafeAreaView>
  );
}

export function SearchHeader({
  isLoading,
  lifecycleState,
  checkoutState,
  filtersExpanded,
  query,
  resultCount,
  scope,
  selectedSurface,
  selectedTagIds,
  searchInputRef,
  searchInputFocused,
  sort,
  statusMessage,
  submittedQuery,
  tagFilters,
  onChangeSurface,
  onChangeLifecycleState,
  onChangeCheckoutState,
  onChangeQuery,
  onChangeScope,
  onChangeSort,
  onClearQuery,
  onClearFilters,
  onToggleTag,
  onSearchBlur,
  onSearchFocus,
  onSubmit,
  onToggleFilters
}: {
  readonly isLoading: boolean;
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly checkoutState: AssetBrowseCheckoutFilter;
  readonly filtersExpanded: boolean;
  readonly query: string;
  readonly resultCount: number;
  readonly scope: BrowseScope;
  readonly selectedSurface: InventoryMapSurface;
  readonly selectedTagIds: readonly string[];
  readonly searchInputRef: RefObject<TextInput | null>;
  readonly searchInputFocused: boolean;
  readonly sort: AssetBrowseSort;
  readonly statusMessage?: string;
  readonly submittedQuery: string;
  readonly tagFilters?: readonly AssetTagOptionViewModel[];
  readonly onChangeSurface: (surface: InventoryMapSurface) => void;
  readonly onChangeLifecycleState: (lifecycleState: AssetBrowseLifecycleFilter) => void;
  readonly onChangeCheckoutState: (checkoutState: AssetBrowseCheckoutFilter) => void;
  readonly onChangeQuery: (query: string) => void;
  readonly onChangeScope: (scope: BrowseScope) => void;
  readonly onChangeSort: (sort: AssetBrowseSort) => void;
  readonly onClearQuery: () => void;
  readonly onClearFilters: () => void;
  readonly onToggleTag?: (tag: AssetTagOptionViewModel) => void;
  readonly onSearchBlur: () => void;
  readonly onSearchFocus: () => void;
  readonly onSubmit: () => void;
  readonly onToggleFilters: (expanded: boolean) => void;
}) {
  const summaryLabel = searchResultSummaryLabel({
    lifecycleState,
    query: submittedQuery,
    resultCount,
    scope,
    sort
  });
  const sortedTagFilters = [...(tagFilters ?? [])].sort(compareTagOptions);
  const selectedTagIdSet = new Set(selectedTagIds);
  const filtersActive = hasActiveFilters({ checkoutState, lifecycleState, scope, sort, tagCount: selectedTagIds.length });

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
          placeholder="Search things, places, tags"
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
      <View style={styles.filterSummaryRow}>
        <Pressable
          accessibilityLabel={filtersExpanded ? 'Hide filters' : 'Show filters'}
          accessibilityRole="button"
          accessibilityState={{ expanded: filtersExpanded }}
          hitSlop={compactControlHitSlop}
          onPress={() => onToggleFilters(!filtersExpanded)}
          style={styles.filterToggle}
        >
          <SlidersHorizontal color={colors.action} size={17} strokeWidth={2.5} />
          <Text style={styles.filterToggleText}>Filters</Text>
        </Pressable>
        <Text numberOfLines={1} style={styles.filterSummaryText}>
          {filterSummary({ checkoutState, lifecycleState, scope, sort, tagCount: selectedTagIds.length })}
        </Text>
      </View>
      {filtersExpanded ? (
        <View>
          <View style={styles.filterActionsRow}>
            <Text style={styles.filterActionsTitle}>Filters</Text>
            <Pressable
              accessibilityLabel="Clear filters"
              accessibilityRole="button"
              disabled={!filtersActive}
              hitSlop={compactControlHitSlop}
              onPress={onClearFilters}
              style={[styles.clearFiltersButton, !filtersActive ? styles.clearFiltersButtonDisabled : null]}
            >
              <Text style={[styles.clearFiltersText, !filtersActive ? styles.clearFiltersTextDisabled : null]}>
                Clear filters
              </Text>
            </Pressable>
          </View>
          <FilterSection title="Scope">
            <ScopeControl selectedScope={scope} onChangeScope={onChangeScope} />
          </FilterSection>
          {sortedTagFilters.length > 0 && onToggleTag ? (
            <FilterSection title="Tags">
              <ScrollView
                accessibilityLabel="Tag filters"
                horizontal
                showsHorizontalScrollIndicator={false}
                contentContainerStyle={styles.optionScroller}
              >
                {sortedTagFilters.map((tag) => (
                  <OptionChip
                    key={tag.id}
                    label={tag.label}
                    color={tag.color}
                    selected={selectedTagIdSet.has(tag.id)}
                    accessibilityLabel={`Filter by tag ${tag.label}`}
                    onPress={() => onToggleTag(tag)}
                  />
                ))}
              </ScrollView>
            </FilterSection>
          ) : null}
          <RefinementBar
            lifecycleState={lifecycleState}
            checkoutState={checkoutState}
            searchMode={submittedQuery.trim().length > 0}
            sort={sort}
            onChangeLifecycleState={onChangeLifecycleState}
            onChangeCheckoutState={onChangeCheckoutState}
            onChangeSort={onChangeSort}
          />
        </View>
      ) : null}
      {statusMessage ? <Text style={styles.errorText}>{statusMessage}</Text> : null}
    </View>
  );
}

function compareTagOptions(left: AssetTagOptionViewModel, right: AssetTagOptionViewModel): number {
  return left.label.localeCompare(right.label, undefined, { sensitivity: 'base' });
}

function filterSummary({
  checkoutState,
  lifecycleState,
  scope,
  sort,
  tagCount
}: {
  readonly checkoutState: AssetBrowseCheckoutFilter;
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly scope: BrowseScope;
  readonly sort: AssetBrowseSort;
  readonly tagCount: number;
}): string {
  const scopeLabel = buildBrowseScopeOptions().find((option) => option.value === scope)?.label ?? 'All';
  const tagLabel = tagCount === 0 ? 'No tags' : tagCount === 1 ? '1 tag' : `${tagCount} tags`;
  const lifecycleLabel = lifecycleState === 'active' ? 'Active' : lifecycleState === 'archived' ? 'Archived' : 'All';
  const checkoutLabel = checkoutState === 'checked_out' ? 'Checked out' : checkoutState === 'available' ? 'Available' : 'Any';
  const sortLabel = sort === 'id_asc' ? 'Stable' : 'Recent';
  return `${scopeLabel} · ${tagLabel} · ${lifecycleLabel} · ${checkoutLabel} · ${sortLabel}`;
}

function hasActiveFilters({
  checkoutState,
  lifecycleState,
  scope,
  sort,
  tagCount
}: {
  readonly checkoutState: AssetBrowseCheckoutFilter;
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly scope: BrowseScope;
  readonly sort: AssetBrowseSort;
  readonly tagCount: number;
}): boolean {
  return scope !== 'all' || lifecycleState !== 'active' || checkoutState !== 'any' || sort !== 'updated_desc' || tagCount > 0;
}

function uniqueTagIds(tagIds: readonly string[]): readonly string[] {
  return [...new Set(tagIds.map((id) => id.trim()).filter((id) => id.length > 0))];
}

function toggleTagId(tagIds: readonly string[], tagId: string): readonly string[] {
  return tagIds.includes(tagId)
    ? tagIds.filter((id) => id !== tagId)
    : [...tagIds, tagId];
}

function ScopeControl({
  selectedScope,
  onChangeScope
}: {
  readonly selectedScope: BrowseScope;
  readonly onChangeScope: (scope: BrowseScope) => void;
}) {
  return (
    <View style={styles.optionGroup}>
      {buildBrowseScopeOptions().map((option) => {
        const selected = option.value === selectedScope;
        return (
          <OptionChip
            key={option.value}
            label={option.label}
            selected={selected}
            onPress={() => onChangeScope(option.value)}
          />
        );
      })}
    </View>
  );
}

function RefinementBar({
  checkoutState,
  lifecycleState,
  searchMode,
  sort,
  onChangeCheckoutState,
  onChangeLifecycleState,
  onChangeSort
}: {
  readonly checkoutState: AssetBrowseCheckoutFilter;
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly searchMode: boolean;
  readonly sort: AssetBrowseSort;
  readonly onChangeCheckoutState: (checkoutState: AssetBrowseCheckoutFilter) => void;
  readonly onChangeLifecycleState: (lifecycleState: AssetBrowseLifecycleFilter) => void;
  readonly onChangeSort: (sort: AssetBrowseSort) => void;
}) {
  return (
    <ScrollView
      horizontal
      showsHorizontalScrollIndicator={false}
      contentContainerStyle={styles.refinementSections}
    >
      <FilterSection title="Status" horizontal>
        <OptionChip label="Active" selected={lifecycleState === 'active'} onPress={() => onChangeLifecycleState('active')} />
        <OptionChip label="Archived" selected={lifecycleState === 'archived'} onPress={() => onChangeLifecycleState('archived')} />
        <OptionChip label="All" selected={lifecycleState === 'all'} onPress={() => onChangeLifecycleState('all')} />
      </FilterSection>
      <FilterSection title="Checkout" horizontal>
        <OptionChip label="Any" selected={checkoutState === 'any'} onPress={() => onChangeCheckoutState('any')} />
        <OptionChip label="Checked out" selected={checkoutState === 'checked_out'} onPress={() => onChangeCheckoutState('checked_out')} />
        <OptionChip label="Available" selected={checkoutState === 'available'} onPress={() => onChangeCheckoutState('available')} />
      </FilterSection>
      {!searchMode ? (
        <FilterSection title="Sort" horizontal>
          <OptionChip label="Recent" selected={sort === 'updated_desc'} onPress={() => onChangeSort('updated_desc')} />
          <OptionChip label="Stable" selected={sort === 'id_asc'} onPress={() => onChangeSort('id_asc')} />
        </FilterSection>
      ) : null}
    </ScrollView>
  );
}

function FilterSection({
  title,
  children,
  horizontal = false,
  accessibilityLabel
}: {
  readonly title: string;
  readonly children: ReactNode;
  readonly horizontal?: boolean;
  readonly accessibilityLabel?: string;
}) {
  return (
    <View style={horizontal ? styles.filterSectionHorizontal : styles.filterSection} accessibilityLabel={accessibilityLabel}>
      <Text style={styles.filterSectionTitle}>{title}</Text>
      <View style={horizontal ? styles.optionGroupHorizontal : undefined}>
        {children}
      </View>
    </View>
  );
}

function OptionChip({
  label,
  selected,
  color,
  accessibilityLabel,
  onPress
}: {
  readonly label: string;
  readonly selected: boolean;
  readonly color?: string;
  readonly accessibilityLabel?: string;
  readonly onPress: () => void;
}) {
  return (
    <Pressable
      accessibilityLabel={accessibilityLabel}
      accessibilityRole="button"
      accessibilityState={{ selected }}
      hitSlop={compactControlHitSlop}
      onPress={onPress}
      style={[
        styles.filterChip,
        color ? { borderColor: color } : null,
        selected ? styles.filterChipSelected : null,
        selected && color ? { backgroundColor: `${color}1F`, borderColor: color } : null
      ]}
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
  filtersActive,
  onClearFilters,
  onClearSearch
}: {
  readonly query: string;
  readonly scope: BrowseScope;
  readonly filtersActive: boolean;
  readonly onClearFilters: () => void;
  readonly onClearSearch: () => void;
}) {
  const hasQuery = query.trim().length > 0;
  return (
    <View style={styles.emptyPanel}>
      <Text style={styles.emptyTitle}>
        {hasQuery ? `No matches for "${query}"` : filtersActive ? 'No matches for these filters' : `No ${scope === 'all' ? 'things' : scope} yet`}
      </Text>
      <Text style={styles.emptyText}>
        {hasQuery
          ? 'Try a broader search or switch scopes.'
          : filtersActive
            ? 'Clear filters or choose fewer tags.'
            : 'New inventory activity will appear here.'}
      </Text>
      {hasQuery ? (
        <Pressable accessibilityRole="button" onPress={onClearSearch} style={styles.emptyAction}>
          <Text style={styles.emptyActionText}>Clear search</Text>
        </Pressable>
      ) : null}
      {!hasQuery && filtersActive ? (
        <Pressable accessibilityRole="button" onPress={onClearFilters} style={styles.emptyAction}>
          <Text style={styles.emptyActionText}>Clear filters</Text>
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
  filterSection: {
    gap: spacing.xs,
    marginTop: spacing.xs
  },
  filterSectionHorizontal: {
    gap: spacing.xs,
    marginRight: spacing.md
  },
  filterSectionTitle: {
    color: colors.textMuted,
    fontSize: 11,
    fontWeight: '900',
    letterSpacing: 0
  },
  optionGroup: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.xs
  },
  optionGroupHorizontal: {
    flexDirection: 'row',
    gap: spacing.xs
  },
  filterSummaryRow: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.sm,
    marginTop: spacing.xs
  },
  filterToggle: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    flexDirection: 'row',
    gap: spacing.xs,
    minHeight: 34,
    paddingHorizontal: spacing.sm
  },
  filterToggleText: {
    color: colors.action,
    fontSize: 13,
    fontWeight: '900',
    letterSpacing: 0
  },
  filterSummaryText: {
    color: colors.textMuted,
    flex: 1,
    fontSize: 11,
    fontWeight: '800',
    letterSpacing: 0
  },
  filterActionsRow: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.sm,
    marginTop: spacing.sm
  },
  filterActionsTitle: {
    color: colors.text,
    flex: 1,
    fontSize: 13,
    fontWeight: '900',
    letterSpacing: 0
  },
  clearFiltersButton: {
    alignItems: 'center',
    borderColor: colors.border,
    borderRadius: radius.sm,
    borderWidth: 1,
    justifyContent: 'center',
    minHeight: 32,
    paddingHorizontal: spacing.sm
  },
  clearFiltersButtonDisabled: {
    opacity: 0.5
  },
  clearFiltersText: {
    color: colors.action,
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 0
  },
  clearFiltersTextDisabled: {
    color: colors.textMuted
  },
  refinementSections: {
    gap: spacing.sm,
    paddingBottom: spacing.xs,
    paddingTop: spacing.sm
  },
  optionScroller: {
    gap: spacing.xs,
    minWidth: '100%',
    paddingBottom: 2
  },
  filterChip: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    minHeight: 34,
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
