import type { ReactNode, RefObject } from 'react';
import {
  ActionSheetIOS,
  ActivityIndicator,
  Alert,
  Modal,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View
} from 'react-native';
import { Check, ChevronDown, Search, SlidersHorizontal, X } from 'lucide-react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { AssetTagOptionViewModel } from '../../application/assets/InventoryAssetTagsQuery';
import type {
  AssetBrowseCheckoutFilter,
  AssetBrowseLifecycleFilter,
  AssetBrowseSort
} from '../../application/home/InventorySummaryRepository';
import {
  buildBrowseFilterTokens,
  buildBrowseScopeOptions,
  browseFilterCount,
  searchResultSummaryLabel,
  sortLabel
} from './SearchScreenPresentation';
import type { BrowseFilterToken, BrowseScope } from './SearchScreenPresentation';
import { BrowseSurfaceControl } from './InventoryMapScreen';
import type { InventoryMapSurface } from './InventoryMapPresentation';
import { radius, spacing } from '../theme/tokens';
import type { MobileColorPalette } from '../theme/tokens';

type TagFilterStatus = 'loading' | 'ready' | 'error';

export type BrowseDraftFilters = {
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly checkoutState: AssetBrowseCheckoutFilter;
  readonly tagIds: readonly string[];
};

export type SearchHeaderProps = {
  readonly isLoading: boolean;
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly checkoutState: AssetBrowseCheckoutFilter;
  readonly filtersExpanded: boolean;
  readonly filterDraft: BrowseDraftFilters;
  readonly inventoryContext?: string;
  readonly inventoryContextStatus?: 'loading' | 'ready' | 'error';
  readonly palette: MobileColorPalette;
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
  readonly tagFilterStatus?: TagFilterStatus;
  readonly onApplyFilters: (filters: BrowseDraftFilters) => void;
  readonly onChangeDraftCheckoutState: (checkoutState: AssetBrowseCheckoutFilter) => void;
  readonly onChangeDraftLifecycleState: (lifecycleState: AssetBrowseLifecycleFilter) => void;
  readonly onChangeDraftTagIds: (tagIds: readonly string[]) => void;
  readonly onChangeQuery: (query: string) => void;
  readonly onChangeScope: (scope: BrowseScope) => void;
  readonly onChangeSort: (sort: AssetBrowseSort) => void;
  readonly onChangeSurface: (surface: InventoryMapSurface) => void;
  readonly onClearFilters: () => void;
  readonly onClearQuery: () => void;
  readonly onRemoveFilter: (token: BrowseFilterToken) => void;
  readonly onRetryInventoryContext?: () => void;
  readonly onRetryResults?: () => void;
  readonly onRetryTags?: () => void;
  readonly onSearchBlur: () => void;
  readonly onSearchFocus: () => void;
  readonly onSubmit: () => void;
  readonly onToggleFilters: (expanded: boolean) => void;
};

export function SearchHeader({
  isLoading,
  lifecycleState,
  checkoutState,
  filtersExpanded,
  filterDraft,
  inventoryContext,
  inventoryContextStatus = 'ready',
  palette,
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
  tagFilters = [],
  tagFilterStatus = 'ready',
  onApplyFilters,
  onChangeDraftCheckoutState,
  onChangeDraftLifecycleState,
  onChangeDraftTagIds,
  onChangeQuery,
  onChangeScope,
  onChangeSort,
  onChangeSurface,
  onClearFilters,
  onClearQuery,
  onRemoveFilter,
  onRetryInventoryContext,
  onRetryResults,
  onRetryTags,
  onSearchBlur,
  onSearchFocus,
  onSubmit,
  onToggleFilters
}: SearchHeaderProps) {
  const styles = stylesForPalette(palette);
  const activeFilterCount = browseFilterCount({ lifecycleState, checkoutState, tagIds: selectedTagIds });
  const activeTokens = buildBrowseFilterTokens(
    { lifecycleState, checkoutState, tagIds: selectedTagIds },
    tagFilters
  );
  const summaryLabel = searchResultSummaryLabel({
    hasTagFilters: selectedTagIds.length > 0,
    lifecycleState,
    query: submittedQuery,
    resultCount,
    scope,
    sort
  });
  const isSearchMode = query.trim().length > 0
    || submittedQuery.trim().length > 0
    || selectedTagIds.length > 0;

  return (
    <View style={baseStyles.header}>
      <View style={styles.headerTopRow}>
        <View style={styles.titleBlock}>
          <Text accessibilityRole="header" style={styles.title}>Browse</Text>
          {inventoryContext ? (
            <Text numberOfLines={1} style={styles.inventoryContext}>{inventoryContext}</Text>
          ) : inventoryContextStatus === 'loading' ? (
            <Text numberOfLines={1} style={styles.inventoryContext}>Loading inventory…</Text>
          ) : (
            <View style={styles.inventoryContextError}>
              <Text numberOfLines={1} style={styles.inventoryContext}>Inventory context unavailable</Text>
              {onRetryInventoryContext ? (
                <Pressable accessibilityLabel="Retry inventory context" accessibilityRole="button" onPress={onRetryInventoryContext} style={styles.inventoryContextRetryButton}>
                  <Text style={styles.inventoryContextRetry}>Retry</Text>
                </Pressable>
              ) : null}
            </View>
          )}
        </View>
        <BrowseSurfaceControl palette={palette} selectedSurface={selectedSurface} onChangeSurface={onChangeSurface} />
      </View>

      <View style={[styles.searchBar, searchInputFocused ? styles.searchBarFocused : null]}>
        <Search color={palette.textMuted} size={19} strokeWidth={2.25} />
        <TextInput
          accessibilityLabel="Search names, places, or tags"
          autoCapitalize="none"
          ref={searchInputRef}
          onBlur={onSearchBlur}
          onChangeText={onChangeQuery}
          onFocus={onSearchFocus}
          onSubmitEditing={onSubmit}
          placeholder="Search names, places, or tags"
          placeholderTextColor={palette.textMuted}
          returnKeyType="search"
          style={styles.searchInput}
          value={query}
        />
        {query.length > 0 ? (
          <Pressable
            accessibilityLabel="Clear search"
            accessibilityRole="button"
            onPress={onClearQuery}
            style={({ pressed }) => [styles.iconButton, pressed ? styles.controlPressed : null]}
          >
            <X color={palette.textMuted} size={18} strokeWidth={2.5} />
          </Pressable>
        ) : null}
        {isLoading ? <ActivityIndicator color={palette.accent} size="small" /> : null}
      </View>

      <ScopeControl palette={palette} selectedScope={scope} onChangeScope={onChangeScope} />

      <View style={styles.resultToolsRow}>
        <Text accessibilityLiveRegion="polite" numberOfLines={1} style={styles.resultSummary}>
          {summaryLabel}
        </Text>
        <Pressable
          accessibilityLabel={activeFilterCount > 0 ? `Filters, ${activeFilterCount.toString()} applied` : 'Filters'}
          accessibilityRole="button"
          accessibilityState={{ expanded: filtersExpanded }}
          onPress={() => onToggleFilters(true)}
          style={({ pressed }) => [styles.toolButton, pressed ? styles.controlPressed : null]}
        >
          <SlidersHorizontal color={palette.action} size={17} strokeWidth={2.4} />
          <Text style={styles.toolButtonText}>Filters{activeFilterCount > 0 ? ` ${activeFilterCount.toString()}` : ''}</Text>
        </Pressable>
        <Pressable
          accessibilityLabel={isSearchMode ? 'Sort unavailable during search' : `Sort, ${sortLabel(sort)}`}
          accessibilityRole="button"
          accessibilityState={{ disabled: isSearchMode }}
          disabled={isSearchMode}
          onPress={() => showSortOptions(sort, onChangeSort)}
          style={({ pressed }) => [
            styles.toolButton,
            isSearchMode ? styles.toolButtonDisabled : null,
            pressed ? styles.controlPressed : null
          ]}
        >
          <Text style={[styles.toolButtonText, isSearchMode ? styles.toolButtonTextDisabled : null]}>Sort</Text>
          <ChevronDown color={isSearchMode ? palette.textMuted : palette.action} size={16} strokeWidth={2.4} />
        </Pressable>
      </View>

      {activeTokens.length > 0 ? (
        <ScrollView
          accessibilityLabel="Applied filters"
          horizontal
          showsHorizontalScrollIndicator={false}
          contentContainerStyle={styles.activeFilterRow}
        >
          {activeTokens.map((token) => (
            <Pressable
              accessibilityLabel={`Remove filter ${token.label}`}
              accessibilityRole="button"
              key={token.key}
              onPress={() => onRemoveFilter(token)}
              style={({ pressed }) => [styles.activeFilterToken, pressed ? styles.controlPressed : null]}
            >
              <Text style={styles.activeFilterTokenText}>{token.label}</Text>
              <X color={palette.accentStrong} size={14} strokeWidth={2.5} />
            </Pressable>
          ))}
          {activeTokens.length > 1 ? (
            <Pressable
              accessibilityRole="button"
              onPress={onClearFilters}
              style={({ pressed }) => [styles.clearAllButton, pressed ? styles.controlPressed : null]}
            >
              <Text style={styles.clearAllText}>Clear all</Text>
            </Pressable>
          ) : null}
        </ScrollView>
      ) : null}

      {statusMessage ? (
        <View accessibilityLiveRegion="polite" style={styles.inlineError}>
          <Text style={styles.errorText}>{statusMessage}</Text>
          {onRetryResults ? (
            <Pressable
              accessibilityRole="button"
              onPress={onRetryResults}
              style={({ pressed }) => [styles.retryButton, pressed ? styles.controlPressed : null]}
            >
              <Text style={styles.retryText}>Retry</Text>
            </Pressable>
          ) : null}
        </View>
      ) : null}

      {filtersExpanded ? (
        <BrowseFilterSheet
          draft={filterDraft}
          palette={palette}
          tagFilters={tagFilters}
          tagFilterStatus={tagFilterStatus}
          onApply={() => onApplyFilters(filterDraft)}
          onChangeCheckoutState={onChangeDraftCheckoutState}
          onChangeLifecycleState={onChangeDraftLifecycleState}
          onChangeTagIds={onChangeDraftTagIds}
          onClose={() => onToggleFilters(false)}
          onReset={() => {
            onChangeDraftLifecycleState('active');
            onChangeDraftCheckoutState('any');
            onChangeDraftTagIds([]);
          }}
          onRetryTags={onRetryTags}
        />
      ) : null}
    </View>
  );
}

function ScopeControl({
  palette,
  selectedScope,
  onChangeScope
}: {
  readonly palette: MobileColorPalette;
  readonly selectedScope: BrowseScope;
  readonly onChangeScope: (scope: BrowseScope) => void;
}) {
  const styles = stylesForPalette(palette);
  return (
    <View accessibilityLabel="Browse by kind" accessibilityRole="tablist" style={styles.scopeControl}>
      {buildBrowseScopeOptions().map((option) => {
        const selected = option.value === selectedScope;
        return (
          <Pressable
            accessibilityRole="tab"
            accessibilityState={{ selected }}
            key={option.value}
            onPress={() => onChangeScope(option.value)}
            style={({ pressed }) => [
              styles.scopeButton,
              selected ? styles.scopeButtonSelected : null,
              pressed ? styles.controlPressed : null
            ]}
          >
            <Text style={[styles.scopeText, selected ? styles.scopeTextSelected : null]}>{option.label}</Text>
          </Pressable>
        );
      })}
    </View>
  );
}

function BrowseFilterSheet({
  draft,
  palette,
  tagFilters,
  tagFilterStatus,
  onApply,
  onChangeCheckoutState,
  onChangeLifecycleState,
  onChangeTagIds,
  onClose,
  onReset,
  onRetryTags
}: {
  readonly draft: BrowseDraftFilters;
  readonly palette: MobileColorPalette;
  readonly tagFilters: readonly AssetTagOptionViewModel[];
  readonly tagFilterStatus: TagFilterStatus;
  readonly onApply: () => void;
  readonly onChangeCheckoutState: (state: AssetBrowseCheckoutFilter) => void;
  readonly onChangeLifecycleState: (state: AssetBrowseLifecycleFilter) => void;
  readonly onChangeTagIds: (ids: readonly string[]) => void;
  readonly onClose: () => void;
  readonly onReset: () => void;
  readonly onRetryTags?: () => void;
}) {
  const styles = stylesForPalette(palette);
  const sortedTags = [...tagFilters].sort((left, right) => left.label.localeCompare(right.label, undefined, { sensitivity: 'base' }));
  const selectedTags = new Set(draft.tagIds);
  return (
    <Modal animationType="slide" onRequestClose={onClose} presentationStyle="pageSheet" visible>
      <SafeAreaView style={styles.filterSheet} edges={['top', 'left', 'right', 'bottom']}>
        <View style={styles.sheetHeader}>
          <Pressable accessibilityRole="button" onPress={onClose} style={({ pressed }) => [styles.sheetHeaderButton, pressed ? styles.controlPressed : null]}>
            <Text style={styles.sheetHeaderSecondary}>Cancel</Text>
          </Pressable>
          <Text accessibilityRole="header" style={styles.sheetTitle}>Filters</Text>
          <Pressable accessibilityRole="button" onPress={onReset} style={({ pressed }) => [styles.sheetHeaderButton, pressed ? styles.controlPressed : null]}>
            <Text style={styles.sheetHeaderSecondary}>Reset</Text>
          </Pressable>
        </View>
        <ScrollView contentContainerStyle={styles.sheetContent}>
          <FilterSection palette={palette} title="Status">
            <View style={styles.sheetOptionGroup}>
              <FilterChip label="Active" selected={draft.lifecycleState === 'active'} palette={palette} onPress={() => onChangeLifecycleState('active')} />
              <FilterChip label="Archived" selected={draft.lifecycleState === 'archived'} palette={palette} onPress={() => onChangeLifecycleState('archived')} />
              <FilterChip label="All" selected={draft.lifecycleState === 'all'} palette={palette} onPress={() => onChangeLifecycleState('all')} />
            </View>
          </FilterSection>
          <FilterSection palette={palette} title="Availability">
            <View style={styles.sheetOptionGroup}>
              <FilterChip label="Any" selected={draft.checkoutState === 'any'} palette={palette} onPress={() => onChangeCheckoutState('any')} />
              <FilterChip label="Available" selected={draft.checkoutState === 'available'} palette={palette} onPress={() => onChangeCheckoutState('available')} />
              <FilterChip label="Checked out" selected={draft.checkoutState === 'checked_out'} palette={palette} onPress={() => onChangeCheckoutState('checked_out')} />
            </View>
          </FilterSection>
          <FilterSection palette={palette} title="Tags">
            {tagFilterStatus === 'loading' ? <ActivityIndicator color={palette.accent} /> : null}
            {tagFilterStatus === 'error' ? (
              <View style={styles.tagError}>
                <Text style={styles.sheetSupportingText}>Tags could not be loaded.</Text>
                {onRetryTags ? (
                  <Pressable accessibilityRole="button" onPress={onRetryTags} style={({ pressed }) => [styles.retryButton, pressed ? styles.controlPressed : null]}>
                    <Text style={styles.retryText}>Retry</Text>
                  </Pressable>
                ) : null}
              </View>
            ) : null}
            {tagFilterStatus === 'ready' && sortedTags.length === 0 ? <Text style={styles.sheetSupportingText}>No tags in this inventory.</Text> : null}
            {tagFilterStatus === 'ready' && sortedTags.length > 0 ? (
              <View accessibilityLabel="Tag filters" style={styles.sheetOptionGroup}>
                {sortedTags.map((tag) => (
                  <FilterChip
                    accessibilityLabel={`Filter by tag ${tag.label}`}
                    color={tag.color}
                    key={tag.id}
                    label={tag.label}
                    palette={palette}
                    selected={selectedTags.has(tag.id)}
                    onPress={() => onChangeTagIds(toggleValue(draft.tagIds, tag.id))}
                  />
                ))}
              </View>
            ) : null}
          </FilterSection>
        </ScrollView>
        <View style={styles.sheetFooter}>
          <Pressable accessibilityRole="button" onPress={onApply} style={({ pressed }) => [styles.applyButton, pressed ? styles.applyButtonPressed : null]}>
            <Text style={styles.applyButtonText}>Show results</Text>
          </Pressable>
        </View>
      </SafeAreaView>
    </Modal>
  );
}

function FilterSection({
  title,
  children,
  palette
}: {
  readonly title: string;
  readonly children: ReactNode;
  readonly palette: MobileColorPalette;
}) {
  return (
    <View style={baseStyles.filterSection}>
      <Text accessibilityRole="header" style={[baseStyles.filterSectionTitle, { color: palette.text }]}>{title}</Text>
      {children}
    </View>
  );
}

function FilterChip({
  accessibilityLabel,
  color,
  label,
  palette,
  selected,
  onPress
}: {
  readonly accessibilityLabel?: string;
  readonly color?: string;
  readonly label: string;
  readonly palette: MobileColorPalette;
  readonly selected: boolean;
  readonly onPress: () => void;
}) {
  const styles = stylesForPalette(palette);
  return (
    <Pressable
      accessibilityLabel={accessibilityLabel}
      accessibilityRole="button"
      accessibilityState={{ selected }}
      onPress={onPress}
      style={({ pressed }) => [
        styles.filterChip,
        selected ? styles.filterChipSelected : null,
        color ? { borderColor: color } : null,
        pressed ? styles.controlPressed : null
      ]}
    >
      {selected ? <Check color={palette.accentStrong} size={16} strokeWidth={2.8} /> : null}
      <Text style={[styles.filterChipText, selected ? styles.filterChipTextSelected : null]}>{label}</Text>
    </Pressable>
  );
}

export function showSortOptions(sort: AssetBrowseSort, onChangeSort: (sort: AssetBrowseSort) => void): void {
  const options = ['Recently changed', 'Default order', 'Cancel'];
  const choose = (index: number) => {
    if (index === 0 && sort !== 'updated_desc') onChangeSort('updated_desc');
    if (index === 1 && sort !== 'id_asc') onChangeSort('id_asc');
  };
  if (Platform.OS === 'ios') {
    ActionSheetIOS.showActionSheetWithOptions({ title: 'Sort', options, cancelButtonIndex: 2 }, choose);
    return;
  }
  Alert.alert('Sort', undefined, [
    { text: 'Recently changed', onPress: () => choose(0) },
    { text: 'Default order', onPress: () => choose(1) },
    { text: 'Cancel', style: 'cancel' }
  ]);
}

function toggleValue(values: readonly string[], value: string): readonly string[] {
  return values.includes(value) ? values.filter((item) => item !== value) : [...values, value];
}

const baseStyles = StyleSheet.create({
  header: { marginBottom: spacing.md },
  filterSection: { gap: spacing.sm },
  filterSectionTitle: { fontSize: 17, fontWeight: '700' }
});

export function createBrowseHeaderStyles(palette: MobileColorPalette) {
  return StyleSheet.create({
    headerTopRow: { alignItems: 'center', flexDirection: 'row', gap: spacing.md, marginBottom: spacing.sm },
    titleBlock: { flex: 1, minWidth: 0 },
    title: { color: palette.text, fontSize: 30, fontWeight: '700', lineHeight: 36 },
    inventoryContext: { color: palette.textMuted, fontSize: 13, fontWeight: '600', marginTop: 1 },
    inventoryContextError: { alignItems: 'center', flexDirection: 'row', gap: spacing.sm },
    inventoryContextRetryButton: { alignItems: 'center', justifyContent: 'center', minHeight: 44, minWidth: 44 },
    inventoryContextRetry: { color: palette.action, fontSize: 13, fontWeight: '700', paddingVertical: spacing.xs },
    searchBar: { alignItems: 'center', backgroundColor: palette.surfaceMuted, borderRadius: radius.lg, flexDirection: 'row', gap: spacing.sm, minHeight: 48, paddingLeft: spacing.md, paddingRight: 2 },
    searchBarFocused: { borderColor: palette.action, borderWidth: 2 },
    searchInput: { color: palette.text, flex: 1, fontSize: 16, minHeight: 48, paddingVertical: 0 },
    iconButton: { alignItems: 'center', justifyContent: 'center', minHeight: 44, minWidth: 44 },
    scopeControl: { backgroundColor: palette.surfaceMuted, borderRadius: radius.md, flexDirection: 'row', gap: 2, marginTop: spacing.sm, padding: 2 },
    scopeButton: { alignItems: 'center', borderRadius: radius.sm, flex: 1, justifyContent: 'center', minHeight: 44, paddingHorizontal: 4 },
    scopeButtonSelected: { backgroundColor: palette.elevatedSurface },
    scopeText: { color: palette.textMuted, fontSize: 13, fontWeight: '600' },
    scopeTextSelected: { color: palette.text, fontWeight: '700' },
    resultToolsRow: { alignItems: 'center', flexDirection: 'row', gap: spacing.xs, marginTop: spacing.sm },
    resultSummary: { color: palette.textMuted, flex: 1, fontSize: 13, fontWeight: '600' },
    toolButton: { alignItems: 'center', backgroundColor: palette.surfaceMuted, borderRadius: radius.md, flexDirection: 'row', gap: 4, justifyContent: 'center', minHeight: 44, paddingHorizontal: spacing.sm },
    toolButtonDisabled: { opacity: 0.55 },
    toolButtonText: { color: palette.action, fontSize: 13, fontWeight: '700' },
    toolButtonTextDisabled: { color: palette.textMuted },
    activeFilterRow: { alignItems: 'center', gap: spacing.xs, paddingBottom: spacing.xs, paddingTop: spacing.sm },
    activeFilterToken: { alignItems: 'center', backgroundColor: palette.selected, borderRadius: 999, flexDirection: 'row', gap: spacing.xs, minHeight: 44, paddingHorizontal: spacing.sm },
    activeFilterTokenText: { color: palette.accentStrong, fontSize: 13, fontWeight: '700' },
    clearAllButton: { alignItems: 'center', justifyContent: 'center', minHeight: 44, paddingHorizontal: spacing.sm },
    clearAllText: { color: palette.action, fontSize: 13, fontWeight: '700' },
    inlineError: { alignItems: 'center', backgroundColor: palette.warningSurface, borderRadius: radius.md, flexDirection: 'row', gap: spacing.sm, marginTop: spacing.sm, paddingLeft: spacing.md, paddingRight: spacing.xs, paddingVertical: spacing.xs },
    errorText: { color: palette.warning, flex: 1, fontSize: 14, lineHeight: 20 },
    retryButton: { alignItems: 'center', justifyContent: 'center', minHeight: 44, minWidth: 44, paddingHorizontal: spacing.sm },
    retryText: { color: palette.action, fontSize: 14, fontWeight: '700' },
    filterSheet: { backgroundColor: palette.background, flex: 1 },
    sheetHeader: { alignItems: 'center', borderBottomColor: palette.border, borderBottomWidth: 1, flexDirection: 'row', justifyContent: 'space-between', minHeight: 56, paddingHorizontal: spacing.sm },
    sheetHeaderButton: { alignItems: 'center', justifyContent: 'center', minHeight: 44, minWidth: 72 },
    sheetHeaderSecondary: { color: palette.action, fontSize: 16, fontWeight: '600' },
    sheetTitle: { color: palette.text, fontSize: 17, fontWeight: '700' },
    sheetContent: { gap: spacing.lg, padding: spacing.md },
    sheetOptionGroup: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.sm },
    filterChip: { alignItems: 'center', backgroundColor: palette.surface, borderColor: palette.controlBorder, borderRadius: 999, borderWidth: 1, flexDirection: 'row', gap: spacing.xs, justifyContent: 'center', minHeight: 44, paddingHorizontal: spacing.md },
    filterChipSelected: { backgroundColor: palette.selected, borderColor: palette.accentStrong },
    filterChipText: { color: palette.textMuted, fontSize: 15, fontWeight: '600' },
    filterChipTextSelected: { color: palette.accentStrong, fontWeight: '700' },
    sheetSupportingText: { color: palette.textMuted, fontSize: 15, lineHeight: 21 },
    tagError: { alignItems: 'center', flexDirection: 'row', justifyContent: 'space-between' },
    sheetFooter: { borderTopColor: palette.border, borderTopWidth: 1, padding: spacing.md },
    applyButton: { alignItems: 'center', backgroundColor: palette.action, borderRadius: radius.md, justifyContent: 'center', minHeight: 50 },
    applyButtonPressed: { backgroundColor: palette.actionPressed },
    applyButtonText: { color: palette.onAction, fontSize: 16, fontWeight: '700' },
    controlPressed: { backgroundColor: palette.selected }
  });
}

const styleCache = new WeakMap<object, ReturnType<typeof createBrowseHeaderStyles>>();

function stylesForPalette(palette: MobileColorPalette) {
  const cached = styleCache.get(palette);
  if (cached) return cached;
  const styles = createBrowseHeaderStyles(palette);
  styleCache.set(palette, styles);
  return styles;
}
