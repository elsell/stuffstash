import type { ReactNode, RefObject } from 'react';
import {
  ActivityIndicator,
  Modal,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View
} from 'react-native';
import type { TextInput } from 'react-native';
import { Check, Plus, Search, X } from 'lucide-react-native';
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
import { BrowseSurfaceControl } from './BrowseSurfaceControl';
import type { InventoryMapSurface } from './InventoryMapPresentation';
import { radius, spacing } from '../theme/tokens';
import type { MobileColorPalette } from '../theme/tokens';
import { AppTextInput } from '../components/AppTextInput';
import { NativeActionMenu, type NativeActionMenuGroup } from '../components/NativeActionMenu';
import { NativeRefinementButton } from '../components/NativeRefinementButton';
import { NativeSegmentedControl } from '../components/NativeSegmentedControl';

type TagFilterStatus = 'loading' | 'ready' | 'error';

export type BrowseDraftFilters = {
  readonly scope: BrowseScope;
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly checkoutState: AssetBrowseCheckoutFilter;
  readonly tagIds: readonly string[];
};

export type SearchHeaderProps = {
  readonly canAdd?: boolean;
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
  readonly onAdd?: () => void;
  readonly onChangeDraftCheckoutState: (checkoutState: AssetBrowseCheckoutFilter) => void;
  readonly onChangeDraftLifecycleState: (lifecycleState: AssetBrowseLifecycleFilter) => void;
  readonly onChangeDraftScope: (scope: BrowseScope) => void;
  readonly onChangeDraftTagIds: (tagIds: readonly string[]) => void;
  readonly onChangeQuery: (query: string) => void;
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
  canAdd = false,
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
  onAdd,
  onChangeDraftCheckoutState,
  onChangeDraftLifecycleState,
  onChangeDraftScope,
  onChangeDraftTagIds,
  onChangeQuery,
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
  const activeFilterCount = browseFilterCount({ scope, lifecycleState, checkoutState, tagIds: selectedTagIds });
  const activeTokens = buildBrowseFilterTokens(
    { scope, lifecycleState, checkoutState, tagIds: selectedTagIds },
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
        <View style={styles.headerActions}>
          {canAdd && onAdd ? (
            <Pressable
              accessibilityLabel="Add an asset"
              accessibilityRole="button"
              onPress={onAdd}
              style={({ pressed }) => [styles.headerIconButton, pressed ? styles.controlPressed : null]}
            >
              <Plus color={palette.action} size={24} strokeWidth={2.2} />
            </Pressable>
          ) : null}
          <BrowseSurfaceControl palette={palette} selectedSurface={selectedSurface} onChangeSurface={onChangeSurface} />
        </View>
      </View>

      <View style={[styles.searchBar, searchInputFocused ? styles.searchBarFocused : null]}>
        <Search color={palette.textMuted} size={19} strokeWidth={2.25} />
        <AppTextInput
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

      <View style={styles.resultToolsRow}>
        <Text accessibilityLiveRegion="polite" numberOfLines={1} style={styles.resultSummary}>
          {summaryLabel}
        </Text>
        <NativeRefinementButton
          accessibilityLabel={activeFilterCount > 0 ? `Filters, ${activeFilterCount.toString()} applied` : 'Filters'}
          accessibilityState={{ expanded: filtersExpanded }}
          badgeCount={activeFilterCount}
          iconOnly
          label="Filters"
          onPress={() => onToggleFilters(true)}
          systemImage="line.3.horizontal.decrease"
        />
        <NativeActionMenu
          accessibilityLabel={isSearchMode ? 'Sort unavailable during search' : `Sort, ${sortLabel(sort)}`}
          disabled={isSearchMode}
          groups={browseSortMenuGroups(sort, onChangeSort)}
          trigger={{ androidIcon: 'sort', kind: 'icon', systemImage: 'arrow.up.arrow.down' }}
        />
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
              style={({ pressed }) => [styles.activeFilterToken, pressed ? styles.activeFilterTokenPressed : null]}
            >
              <View style={styles.activeFilterTokenPill}>
                <Text style={styles.activeFilterTokenText}>{token.label}</Text>
                <X color={palette.accentStrong} size={14} strokeWidth={2.5} />
              </View>
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
          onChangeScope={onChangeDraftScope}
          onChangeTagIds={onChangeDraftTagIds}
          onClose={() => onToggleFilters(false)}
          onReset={() => {
            onChangeDraftScope('all');
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

function BrowseFilterSheet({
  draft,
  palette,
  tagFilters,
  tagFilterStatus,
  onApply,
  onChangeCheckoutState,
  onChangeLifecycleState,
  onChangeScope,
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
  readonly onChangeScope: (scope: BrowseScope) => void;
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
          <FilterSection palette={palette} title="Type">
            <View accessibilityLabel="Filter by type">
              <NativeSegmentedControl
                colors={palette}
                onChange={onChangeScope}
                segments={buildBrowseScopeOptions()}
                style={styles.sheetSegmentedControl}
                value={draft.scope}
              />
            </View>
          </FilterSection>
          <FilterSection palette={palette} title="Status">
            <View accessibilityLabel="Filter by status">
              <NativeSegmentedControl
                colors={palette}
                onChange={onChangeLifecycleState}
                segments={lifecycleFilterSegments}
                style={styles.sheetSegmentedControl}
                value={draft.lifecycleState}
              />
            </View>
          </FilterSection>
          <FilterSection palette={palette} title="Availability">
            <View accessibilityLabel="Filter by availability">
              <NativeSegmentedControl
                colors={palette}
                onChange={onChangeCheckoutState}
                segments={availabilityFilterSegments}
                style={styles.sheetSegmentedControl}
                value={draft.checkoutState}
              />
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
              <View accessibilityLabel="Tag filters" style={styles.tagList}>
                {sortedTags.map((tag, index) => {
                  const selected = selectedTags.has(tag.id);
                  return (
                    <View key={tag.id}>
                      {index > 0 ? <View accessibilityElementsHidden style={styles.tagSeparator} /> : null}
                      <Pressable
                        accessibilityLabel={`Filter by tag ${tag.label}`}
                        accessibilityRole="checkbox"
                        accessibilityState={{ checked: selected }}
                        onPress={() => onChangeTagIds(toggleValue(draft.tagIds, tag.id))}
                        style={({ pressed }) => [styles.tagRow, pressed ? styles.controlPressed : null]}
                      >
                        <View
                          accessibilityElementsHidden
                          style={[styles.tagColor, tag.color ? { backgroundColor: tag.color } : styles.emptyTagColor]}
                          testID={`tag-color-${tag.id}`}
                        />
                        <Text style={styles.tagLabel}>{tag.label}</Text>
                        <View accessibilityElementsHidden style={styles.tagCheckSpace}>
                          {selected ? <Check color={palette.action} size={20} strokeWidth={2.5} /> : null}
                        </View>
                      </Pressable>
                    </View>
                  );
                })}
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

const lifecycleFilterSegments = [
  { label: 'Active', value: 'active' },
  { label: 'Archived', value: 'archived' },
  { label: 'All', value: 'all' }
] as const;

const availabilityFilterSegments = [
  { label: 'Any', value: 'any' },
  { label: 'Available', value: 'available' },
  { label: 'Checked out', value: 'checked_out' }
] as const;

export function browseSortMenuGroups(
  sort: AssetBrowseSort,
  onChangeSort: (sort: AssetBrowseSort) => void
): readonly NativeActionMenuGroup[] {
  return [{
    id: 'sort',
    items: [
      {
        id: 'updated_desc',
        label: 'Recently changed',
        isSelected: sort === 'updated_desc',
        onPress: () => {
          if (sort !== 'updated_desc') onChangeSort('updated_desc');
        }
      },
      {
        id: 'id_asc',
        label: 'Default order',
        isSelected: sort === 'id_asc',
        onPress: () => {
          if (sort !== 'id_asc') onChangeSort('id_asc');
        }
      }
    ]
  }];
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
    headerActions: { alignItems: 'center', flexDirection: 'row', gap: spacing.xs },
    headerIconButton: { alignItems: 'center', borderRadius: 22, justifyContent: 'center', minHeight: 44, minWidth: 44 },
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
    resultToolsRow: { alignItems: 'center', flexDirection: 'row', gap: spacing.xs, marginTop: spacing.sm, minHeight: 44 },
    resultSummary: { color: palette.textMuted, flex: 1, fontSize: 13, fontWeight: '600' },
    activeFilterRow: { alignItems: 'center', gap: spacing.xs, paddingBottom: spacing.xs, paddingTop: spacing.sm },
    activeFilterToken: { alignItems: 'center', justifyContent: 'center', minHeight: 44 },
    activeFilterTokenPressed: { opacity: 0.72 },
    activeFilterTokenPill: { alignItems: 'center', backgroundColor: palette.selected, borderRadius: 999, flexDirection: 'row', gap: spacing.xs, minHeight: 32, paddingHorizontal: spacing.sm },
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
    sheetSegmentedControl: { width: '100%' },
    tagList: { backgroundColor: palette.surface, borderRadius: radius.md, overflow: 'hidden' },
    tagRow: { alignItems: 'center', flexDirection: 'row', gap: spacing.sm, minHeight: 52, paddingHorizontal: spacing.md, paddingVertical: spacing.xs },
    tagColor: { borderRadius: 8, height: 16, width: 16 },
    emptyTagColor: { backgroundColor: 'transparent', borderColor: palette.controlBorder, borderWidth: 1.5 },
    tagLabel: { color: palette.text, flex: 1, fontSize: 16 },
    tagCheckSpace: { alignItems: 'center', height: 24, justifyContent: 'center', width: 24 },
    tagSeparator: { backgroundColor: palette.border, height: StyleSheet.hairlineWidth, marginLeft: 48 },
    sheetSupportingText: { color: palette.textMuted, fontSize: 15, lineHeight: 21 },
    tagError: { alignItems: 'center', flexDirection: 'row', justifyContent: 'space-between' },
    sheetFooter: { borderTopColor: palette.border, borderTopWidth: 1, padding: spacing.md },
    applyButton: { alignItems: 'center', backgroundColor: palette.action, borderRadius: radius.md, justifyContent: 'center', minHeight: 50 },
    applyButtonPressed: { backgroundColor: palette.actionPressed },
    applyButtonText: { color: palette.onAction, fontSize: 16, fontWeight: '700' },
    controlPressed: { opacity: 0.82 }
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
