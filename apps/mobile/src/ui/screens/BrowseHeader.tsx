import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { X } from 'lucide-react-native';
import type { AssetTagOptionViewModel } from '../../application/assets/InventoryAssetTagsQuery';
import type {
  AssetBrowseCheckoutFilter,
  AssetBrowseLifecycleFilter,
  AssetBrowseSort
} from '../../application/home/InventorySummaryRepository';
import {
  buildBrowseFilterTokens,
  browseFilterCount,
  searchResultSummaryLabel
} from './SearchScreenPresentation';
import type { BrowseFilterToken, BrowseScope } from './SearchScreenPresentation';
import { BrowseSurfaceControl } from './BrowseSurfaceControl';
import type { InventoryMapSurface } from './InventoryMapPresentation';
import { radius, spacing } from '../theme/tokens';
import type { MobileColorPalette } from '../theme/tokens';
import { NativeRefinementButton } from '../components/NativeRefinementButton';


export type SearchHeaderProps = {
  readonly isLoading: boolean;
  readonly lifecycleState: AssetBrowseLifecycleFilter;
  readonly checkoutState: AssetBrowseCheckoutFilter;
  readonly inventoryContext?: string;
  readonly inventoryContextStatus?: 'loading' | 'ready' | 'error';
  readonly palette: MobileColorPalette;
  readonly resultCount: number;
  readonly scope: BrowseScope;
  readonly selectedSurface: InventoryMapSurface;
  readonly selectedTagIds: readonly string[];
  readonly sort: AssetBrowseSort;
  readonly statusMessage?: string;
  readonly submittedQuery: string;
  readonly tagFilters?: readonly AssetTagOptionViewModel[];
  readonly onChangeSurface: (surface: InventoryMapSurface) => void;
  readonly onClearFilters: () => void;
  readonly onRemoveFilter: (token: BrowseFilterToken) => void;
  readonly onRetryInventoryContext?: () => void;
  readonly onRetryResults?: () => void;
  readonly onToggleFilters: () => void;
};

export function SearchHeader({
  isLoading,
  lifecycleState,
  checkoutState,
  inventoryContext,
  inventoryContextStatus = 'ready',
  palette,
  resultCount,
  scope,
  selectedSurface,
  selectedTagIds,
  sort,
  statusMessage,
  submittedQuery,
  tagFilters = [],
  onChangeSurface,
  onClearFilters,
  onRemoveFilter,
  onRetryInventoryContext,
  onRetryResults,
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


  return (
    <View style={baseStyles.header}>
      <View style={styles.headerTopRow}>
        <View style={styles.titleBlock}>
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
          <BrowseSurfaceControl palette={palette} selectedSurface={selectedSurface} onChangeSurface={onChangeSurface} />
        </View>
      </View>


      <View style={styles.resultToolsRow}>
        {isLoading ? <ActivityIndicator accessibilityLabel="Searching inventory" color={palette.accent} size="small" /> : null}
        <Text accessibilityLiveRegion="polite" numberOfLines={1} style={styles.resultSummary}>
          {summaryLabel}
        </Text>
        <NativeRefinementButton
          accessibilityLabel={activeFilterCount > 0 ? `Filters, ${activeFilterCount.toString()} applied` : 'Filters'}
          badgeCount={activeFilterCount}
          iconOnly
          label="Filters"
          onPress={onToggleFilters}
          systemImage="line.3.horizontal.decrease"
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


    </View>
  );
}

const baseStyles = StyleSheet.create({
  header: { marginBottom: spacing.md }
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
