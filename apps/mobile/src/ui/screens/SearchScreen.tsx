import { useCallback, useEffect, useRef, useState } from 'react';
import type { RefObject } from 'react';
import { router, useFocusEffect } from 'expo-router';
import {
  ActivityIndicator,
  FlatList,
  Pressable,
  StyleSheet,
  Text,
  TextInput,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type {
  AssetBrowseKindFilter,
  AssetBrowseLifecycleFilter,
  AssetBrowseSort
} from '../../application/home/InventorySummaryRepository';
import {
  SearchAssetsQuery,
  SearchAssetsViewModel
} from '../../application/search/SearchAssetsQuery';
import { AssetCard } from '../components/AssetCard';
import { colors, radius, spacing } from '../theme/tokens';
import {
  buildSearchFilterGroupPlacement,
  focusSearchInput
} from './SearchScreenPresentation';

type SearchScreenProps = {
  readonly searchAssetsQuery: SearchAssetsQuery;
};

type SearchState =
  | { readonly status: 'loading'; readonly results: SearchAssetsViewModel }
  | { readonly status: 'ready'; readonly results: SearchAssetsViewModel }
  | { readonly status: 'error'; readonly message: string; readonly results: SearchAssetsViewModel };

const pageSize = 20;

const emptyResults: SearchAssetsViewModel = {
  query: '',
  mode: 'browse',
  lifecycleState: 'active',
  kind: 'all',
  sort: 'updated_desc',
  assets: [],
  hasMore: false
};

export function SearchScreen({ searchAssetsQuery }: SearchScreenProps) {
  const [query, setQuery] = useState('');
  const [lifecycleState, setLifecycleState] = useState<AssetBrowseLifecycleFilter>('active');
  const [kind, setKind] = useState<AssetBrowseKindFilter>('all');
  const [sort, setSort] = useState<AssetBrowseSort>('updated_desc');
  const [state, setState] = useState<SearchState>({ status: 'loading', results: emptyResults });
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
    let isCurrent = true;
    const requestId = nextRequestId(requestSequence);

    searchAssetsQuery
      .execute({
        query: '',
        lifecycleState,
        kind,
        sort,
        limit: pageSize
      })
      .then((results) => {
        if (isCurrent && isCurrentRequest(requestSequence, requestId)) {
          setState({ status: 'ready', results });
        }
      })
      .catch((error: unknown) => {
        if (isCurrent && isCurrentRequest(requestSequence, requestId)) {
          setState({
            status: 'error',
            message: readableError(error, 'Browse failed.'),
            results: emptyResults
          });
        }
      });

    return () => {
      isCurrent = false;
    };
  }, [searchAssetsQuery]);

  async function loadFirstPage(next: {
    readonly query?: string;
    readonly lifecycleState?: AssetBrowseLifecycleFilter;
    readonly kind?: AssetBrowseKindFilter;
    readonly sort?: AssetBrowseSort;
  } = {}): Promise<void> {
    const requestId = nextRequestId(requestSequence);
    const nextQuery = next.query ?? query;
    const nextLifecycleState = next.lifecycleState ?? lifecycleState;
    const nextKind = next.kind ?? kind;
    const nextSort = next.sort ?? sort;

    setIsLoadingMore(false);
    setIsRefreshing(false);
    setState({ status: 'loading', results: state.results });

    try {
      const results = await searchAssetsQuery.execute({
        query: nextQuery,
        lifecycleState: nextLifecycleState,
        kind: nextKind,
        sort: nextSort,
        limit: pageSize
      });
      if (isCurrentRequest(requestSequence, requestId)) {
        setState({ status: 'ready', results });
      }
    } catch (error) {
      if (isCurrentRequest(requestSequence, requestId)) {
        setState({
          status: 'error',
          message: readableError(error, 'Browse failed.'),
          results: state.results
        });
      }
    }
  }

  async function refreshResults(): Promise<void> {
    const requestId = nextRequestId(requestSequence);
    setIsRefreshing(true);

    try {
      const results = await searchAssetsQuery.execute({
        query: state.results.query,
        lifecycleState,
        kind,
        sort,
        limit: pageSize
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
      const nextPage = await searchAssetsQuery.execute({
        query: state.results.query,
        cursor: state.results.nextCursor,
        lifecycleState,
        kind,
        sort,
        limit: pageSize
      });
      if (isCurrentRequest(requestSequence, requestId)) {
        setState({
          status: 'ready',
          results: {
            ...nextPage,
            assets: [...state.results.assets, ...nextPage.assets]
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

  function updateLifecycleState(nextLifecycleState: AssetBrowseLifecycleFilter): void {
    setLifecycleState(nextLifecycleState);
    void loadFirstPage({ lifecycleState: nextLifecycleState });
  }

  function updateKind(nextKind: AssetBrowseKindFilter): void {
    setKind(nextKind);
    void loadFirstPage({ kind: nextKind });
  }

  function updateSort(nextSort: AssetBrowseSort): void {
    setSort(nextSort);
    void loadFirstPage({ sort: nextSort });
  }

  return (
    <SafeAreaView style={styles.shell} edges={['top', 'left', 'right']}>
      <FlatList
        data={state.results.assets}
        keyExtractor={(asset) => asset.id}
        columnWrapperStyle={styles.cardRow}
        contentContainerStyle={styles.content}
        keyboardShouldPersistTaps="handled"
        numColumns={2}
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
            kind={kind}
            lifecycleState={lifecycleState}
            query={query}
            resultCount={state.results.assets.length}
            resultsMode={state.results.mode}
            searchInputRef={searchInputRef}
            searchInputFocused={isSearchFocused}
            sort={sort}
            statusMessage={state.status === 'error' ? state.message : undefined}
            onChangeKind={updateKind}
            onChangeLifecycleState={updateLifecycleState}
            onChangeQuery={setQuery}
            onChangeSort={updateSort}
            onSearchBlur={() => setIsSearchFocused(false)}
            onSearchFocus={() => setIsSearchFocused(true)}
            onSubmit={submitSearch}
          />
        }
        ListEmptyComponent={
          state.status === 'loading' ? null : <Text style={styles.emptyText}>No matching assets.</Text>
        }
        renderItem={({ item }) => (
          <AssetCard asset={item} onPress={() => router.push(`/assets/${item.id}`)} />
        )}
      />
    </SafeAreaView>
  );
}

export function SearchHeader({
  isLoading,
  kind,
  lifecycleState,
  query,
  resultCount,
  resultsMode,
  searchInputRef,
  searchInputFocused,
  sort,
  statusMessage,
  onChangeKind,
  onChangeLifecycleState,
  onChangeQuery,
  onChangeSort,
  onSearchBlur,
  onSearchFocus,
  onSubmit
}: {
  readonly isLoading: boolean;
  readonly kind: AssetBrowseKindFilter;
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly query: string;
  readonly resultCount: number;
  readonly resultsMode: SearchAssetsViewModel['mode'];
  readonly searchInputRef: RefObject<TextInput | null>;
  readonly searchInputFocused: boolean;
  readonly sort: AssetBrowseSort;
  readonly statusMessage?: string;
  readonly onChangeKind: (kind: AssetBrowseKindFilter) => void;
  readonly onChangeLifecycleState: (lifecycleState: AssetBrowseLifecycleFilter) => void;
  readonly onChangeQuery: (query: string) => void;
  readonly onChangeSort: (sort: AssetBrowseSort) => void;
  readonly onSearchBlur: () => void;
  readonly onSearchFocus: () => void;
  readonly onSubmit: () => void;
}) {
  const filterGroups = buildSearchFilterGroupPlacement(resultsMode);

  return (
    <View>
      <Text style={styles.title}>Browse</Text>
      <View style={styles.searchRow}>
        <TextInput
          accessibilityLabel="Search inventory"
          autoCapitalize="none"
          ref={searchInputRef}
          onChangeText={onChangeQuery}
          onBlur={onSearchBlur}
          onFocus={onSearchFocus}
          onSubmitEditing={onSubmit}
          placeholder="Search assets"
          placeholderTextColor={colors.textMuted}
          returnKeyType="search"
          style={[styles.searchInput, searchInputFocused ? styles.searchInputFocused : null]}
          value={query}
        />
        <Pressable accessibilityRole="button" onPress={onSubmit} style={styles.searchButton}>
          {isLoading ? (
            <ActivityIndicator color={colors.onAction} />
          ) : (
            <Text style={styles.searchButtonText}>Search</Text>
          )}
        </Pressable>
      </View>
      <View style={styles.filterPanel}>
        {filterGroups.map((group) => {
          if (group.key === 'status') {
            return (
              <FilterGroup
                isLast={group.isLast}
                key={group.key}
                label="Status"
                options={[
                  { label: 'Active', value: 'active' },
                  { label: 'Archived', value: 'archived' },
                  { label: 'All', value: 'all' }
                ]}
                selectedValue={lifecycleState}
                onChange={onChangeLifecycleState}
              />
            );
          }

          if (group.key === 'type') {
            return (
              <FilterGroup
                isLast={group.isLast}
                key={group.key}
                label="Type"
                options={[
                  { label: 'All', value: 'all' },
                  { label: 'Items', value: 'item' },
                  { label: 'Containers', value: 'container' },
                  { label: 'Locations', value: 'location' }
                ]}
                selectedValue={kind}
                onChange={onChangeKind}
              />
            );
          }

          return (
            <FilterGroup
              isLast={group.isLast}
              key={group.key}
              label="Sort"
              options={[
                { label: 'Recent', value: 'updated_desc' },
                { label: 'Stable', value: 'id_asc' }
              ]}
              selectedValue={sort}
              onChange={onChangeSort}
            />
          );
        })}
      </View>
      {statusMessage ? <Text style={styles.errorText}>{statusMessage}</Text> : null}
      <Text style={styles.resultCount}>
        {resultCount.toString()} {resultsMode === 'search' ? 'search results' : 'assets'}
      </Text>
    </View>
  );
}

function FilterGroup<T extends string>({
  isLast,
  label,
  options,
  selectedValue,
  onChange
}: {
  readonly isLast: boolean;
  readonly label: string;
  readonly options: ReadonlyArray<{ readonly label: string; readonly value: T }>;
  readonly selectedValue: T;
  readonly onChange: (value: T) => void;
}) {
  return (
    <View style={[styles.filterGroup, isLast ? styles.filterGroupLast : null]}>
      <Text style={styles.filterLabel}>{label}</Text>
      <View style={styles.filterOptions}>
        {options.map((option) => {
          const isSelected = option.value === selectedValue;
          return (
            <Pressable
              accessibilityRole="button"
              accessibilityState={{ selected: isSelected }}
              key={option.value}
              onPress={() => onChange(option.value)}
              style={[styles.filterButton, isSelected ? styles.filterButtonSelected : null]}
            >
              <Text style={[styles.filterText, isSelected ? styles.filterTextSelected : null]}>
                {option.label}
              </Text>
            </Pressable>
          );
        })}
      </View>
    </View>
  );
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
    padding: spacing.lg,
    paddingBottom: spacing.xl
  },
  title: {
    color: colors.text,
    fontSize: 30,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 36,
    marginBottom: spacing.md
  },
  searchRow: {
    flexDirection: 'row',
    gap: spacing.sm,
    marginBottom: spacing.sm
  },
  searchInput: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    color: colors.text,
    flex: 1,
    fontSize: 16,
    minHeight: 46,
    paddingHorizontal: spacing.md
  },
  searchInputFocused: {
    borderColor: colors.focusRing,
    borderWidth: 2
  },
  searchButton: {
    alignItems: 'center',
    backgroundColor: colors.action,
    borderRadius: radius.md,
    justifyContent: 'center',
    minHeight: 46,
    minWidth: 86,
    paddingHorizontal: spacing.md
  },
  searchButtonText: {
    color: colors.onAction,
    fontSize: 15,
    fontWeight: '800',
    letterSpacing: 0
  },
  filterPanel: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    marginBottom: spacing.sm
  },
  filterGroup: {
    borderBottomColor: colors.border,
    borderBottomWidth: 1,
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.sm
  },
  filterGroupLast: {
    borderBottomWidth: 0
  },
  filterLabel: {
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '800',
    letterSpacing: 0,
    marginBottom: spacing.xs
  },
  filterOptions: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.xs
  },
  filterButton: {
    alignItems: 'center',
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    minHeight: 36,
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  },
  filterButtonSelected: {
    backgroundColor: colors.selected,
    borderColor: colors.accent
  },
  filterText: {
    color: colors.textMuted,
    fontSize: 13,
    fontWeight: '800',
    letterSpacing: 0
  },
  filterTextSelected: {
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
    fontSize: 13,
    fontWeight: '700',
    letterSpacing: 0,
    marginBottom: spacing.md
  },
  emptyText: {
    color: colors.textMuted,
    fontSize: 15,
    lineHeight: 22
  },
  footer: {
    alignItems: 'center',
    paddingVertical: spacing.md
  },
  cardRow: {
    gap: spacing.sm,
    marginBottom: spacing.sm
  }
});
