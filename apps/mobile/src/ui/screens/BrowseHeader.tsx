import { NativeCommandButton } from '../components/NativeCommandButton';
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
  readonly onRetryResults?: () => void;
  readonly onToggleFilters: () => void;
};

export function SearchHeader({
  isLoading,
  lifecycleState,
  checkoutState,
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
      <View style={styles.resultToolsRow}>
        <BrowseSurfaceControl palette={palette} selectedSurface={selectedSurface} onChangeSurface={onChangeSurface} />
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
            <NativeCommandButton label="Retry" onPress={onRetryResults} />
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
    resultToolsRow: { alignItems: 'center', flexDirection: 'row', gap: spacing.xs, marginTop: spacing.sm, minHeight: 44 },
    resultSummary: { color: palette.textMuted, flex: 1, fontSize: 13, fontWeight: '600' },
    activeFilterRow: { alignItems: 'center', gap: spacing.xs, paddingBottom: spacing.xs, paddingTop: spacing.sm },
    activeFilterToken: { alignItems: 'center', justifyContent: 'center', minHeight: 44 },
    activeFilterTokenPressed: { opacity: 0.72 },
    activeFilterTokenPill: { alignItems: 'center', backgroundColor: palette.selected, borderRadius: 999, flexDirection: 'row', gap: spacing.xs, minHeight: 32, paddingHorizontal: spacing.sm },
    activeFilterTokenText: { color: palette.accentStrong, fontSize: 13, fontWeight: '700' },
    clearAllButton: { alignItems: 'center', justifyContent: 'center', minHeight: 44, paddingHorizontal: spacing.sm },
    clearAllText: { color: palette.action, fontSize: 13, fontWeight: '700' },
    inlineError: { alignItems: 'stretch', backgroundColor: palette.warningSurface, borderRadius: radius.md, gap: spacing.sm, marginTop: spacing.sm, paddingLeft: spacing.md, paddingRight: spacing.xs, paddingVertical: spacing.xs },
    errorText: { color: palette.warning, fontSize: 14, lineHeight: 20 },
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
