import { Image, Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import type { StyleProp, ViewStyle } from 'react-native';
import type {
  AssetCardViewModel,
  AssetParentLocationCrumbViewModel,
  AssetTagViewModel
} from '../../application/assets/AssetViewModels';
import {
  radius,
  spacing,
  type MobileColorPalette
} from '../theme/tokens';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { AssetTagChips } from './AssetTagChips';

const stylesByPalette = new WeakMap<object, ReturnType<typeof createStyles>>();

function useAssetCardStyles(paletteOverride?: MobileColorPalette) {
  const contextPalette = useAppearancePalette();
  const palette = paletteOverride ?? contextPalette;
  const cached = stylesByPalette.get(palette);
  if (cached) return cached;
  const styles = createStyles(palette);
  stylesByPalette.set(palette, styles);
  return styles;
}

type AssetCardProps = {
  readonly asset: AssetCardViewModel;
  readonly density?: 'standard' | 'compact';
  readonly footerAction?: {
    readonly disabled?: boolean;
    readonly label: string;
    readonly onPress: () => void;
  };
  readonly onPress: () => void;
  readonly onParentLocationPress: (location: AssetParentLocationCrumbViewModel) => void;
  readonly palette?: MobileColorPalette;
  readonly showTags?: boolean;
  readonly style?: StyleProp<ViewStyle>;
  readonly onTagPress?: (tag: AssetTagViewModel) => void;
};

export function AssetCard({
  asset,
  density = 'standard',
  footerAction,
  onParentLocationPress,
  palette: paletteOverride,
  onPress,
  onTagPress,
  showTags = true,
  style
}: AssetCardProps) {
  const isCompact = density === 'compact';
  const styles = useAssetCardStyles(paletteOverride);

  return (
    <View style={[styles.card, isCompact ? styles.compactCard : styles.standardCard, style]}>
      <Pressable
        accessible={false}
        onPress={onPress}
        style={({ pressed }) => [styles.openRegion, pressed ? styles.openRegionPressed : undefined]}
      >
        <View style={styles.imageFrame}>
          {asset.photo ? (
            <Image
              accessibilityIgnoresInvertColors
              source={{ uri: asset.photo.uri, headers: asset.photo.headers }}
              style={styles.assetImage}
            />
          ) : (
            <Text style={styles.imagePlaceholder}>{asset.imagePlaceholderLabel}</Text>
          )}
          {asset.checkedOutLabel ? <Text style={styles.checkoutImageBadge}>{asset.checkedOutLabel}</Text> : null}
        </View>
      </Pressable>
      <View style={styles.body}>
        <Pressable
          accessibilityLabel={`Open asset ${asset.title}`}
          accessibilityRole="button"
          onPress={onPress}
          style={({ pressed }) => [styles.openTextRegion, pressed ? styles.openTextRegionPressed : undefined]}
        >
          <Text numberOfLines={2} style={[styles.title, isCompact ? styles.compactTitle : styles.standardTitle]}>
            {asset.title}
          </Text>
        </Pressable>
        {asset.parentLocationTrail.length > 0 ? (
          <View style={styles.topRow}>
            <AssetBreadcrumbTrail
              palette={paletteOverride}
              segments={asset.parentLocationTrail}
              onSegmentPress={onParentLocationPress}
            />
          </View>
        ) : null}
        {shouldShowAssetCardSupportingDetails(asset, density) ? (
          <Pressable
            accessible={false}
            onPress={onPress}
            style={({ pressed }) => [styles.openTextRegion, pressed ? styles.openTextRegionPressed : undefined]}
          >
            {asset.description.trim().length > 0 ? (
              <Text numberOfLines={2} style={styles.description}>
                {asset.description}
              </Text>
            ) : null}
            {asset.searchMatchLabels && asset.searchMatchLabels.length > 0 ? (
              <Text numberOfLines={1} style={styles.matchMeta}>
                Matched {asset.searchMatchLabels.join(', ')}
              </Text>
            ) : null}
          </Pressable>
        ) : null}
        {showTags ? <AssetTagChips palette={paletteOverride} tags={asset.tags} compact overflowLimit={2} onTagPress={onTagPress} /> : null}
      </View>
      {footerAction ? (
        <Pressable
          accessibilityRole="button"
          disabled={footerAction.disabled}
          onPress={footerAction.onPress}
          style={({ pressed }) => [
            styles.footerAction,
            pressed && !footerAction.disabled ? styles.footerActionPressed : undefined,
            footerAction.disabled ? styles.disabledFooterAction : undefined
          ]}
        >
          <Text style={styles.footerActionText}>{footerAction.label}</Text>
        </Pressable>
      ) : null}
    </View>
  );
}

export function shouldShowAssetCardSupportingDetails(
  asset: AssetCardViewModel,
  density: 'standard' | 'compact'
): boolean {
  return density === 'standard'
    && (asset.description.trim().length > 0 || (asset.searchMatchLabels?.length ?? 0) > 0);
}

type BreadcrumbScroller = {
  readonly scrollToEnd?: (options?: { readonly animated?: boolean }) => void;
};

export function AssetBreadcrumbTrail({
  onSegmentPress,
  palette: paletteOverride,
  prominence = 'compact',
  segments
}: {
  readonly segments: readonly AssetParentLocationCrumbViewModel[];
  readonly onSegmentPress: (location: AssetParentLocationCrumbViewModel) => void;
  readonly palette?: MobileColorPalette;
  readonly prominence?: 'compact' | 'detail';
}) {
  const styles = useAssetCardStyles(paletteOverride);

  if (segments.length === 0) {
    return null;
  }

  let scroller: BreadcrumbScroller | null = null;
  const scrollToMostSpecificParent = () => {
    if (segments.length > 1) {
      scroller?.scrollToEnd?.({ animated: false });
    }
  };

  return (
    <ScrollView
      accessibilityLabel={`Location ${segments.map((segment) => segment.title).join(', ')}`}
      horizontal
      onContentSizeChange={scrollToMostSpecificParent}
      onLayout={scrollToMostSpecificParent}
      ref={(node) => {
        scroller = node;
      }}
      showsHorizontalScrollIndicator={false}
      style={styles.breadcrumbScroller}
      contentContainerStyle={styles.breadcrumbContent}
    >
      {segments.map((segment, index) => (
        <View key={segment.id} style={styles.breadcrumbSegment}>
          {index > 0 ? <Text style={styles.breadcrumbSeparator}>/</Text> : null}
          <Pressable
            accessibilityLabel={`Open location ${segment.title}`}
            accessibilityRole="button"
            onPress={() => onSegmentPress(segment)}
            style={({ pressed }) => [
              styles.breadcrumbButton,
              pressed ? styles.breadcrumbButtonPressed : undefined
            ]}
          >
            <Text
              numberOfLines={1}
              style={[
                styles.breadcrumbText,
                prominence === 'detail' ? styles.detailBreadcrumbText : null,
                segment.isImmediateParent ? styles.immediateBreadcrumbText : styles.ancestorBreadcrumbText
              ]}
            >
              {segment.title}
            </Text>
          </Pressable>
        </View>
      ))}
    </ScrollView>
  );
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
  card: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    overflow: 'hidden'
  },
  standardCard: {
    flex: 1,
    minHeight: 286
  },
  compactCard: {
    minHeight: 210,
    width: 164
  },
  imageFrame: {
    alignItems: 'center',
    aspectRatio: 1,
    backgroundColor: colors.surfaceMuted,
    justifyContent: 'center',
    position: 'relative',
    width: '100%'
  },
  openRegion: {
    width: '100%'
  },
  openRegionPressed: {
    borderColor: colors.controlBorder,
    borderWidth: 2
  },
  openTextRegion: {
    borderRadius: radius.sm,
    gap: spacing.xs,
    minHeight: 44
  },
  openTextRegionPressed: {
    backgroundColor: colors.selected
  },
  imagePlaceholder: {
    color: colors.accentStrong,
    fontSize: 20,
    fontWeight: '700',
    letterSpacing: 0
  },
  assetImage: {
    height: '100%',
    width: '100%'
  },
  body: {
    gap: spacing.xs,
    padding: spacing.sm
  },
  topRow: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.xs,
    minHeight: 28
  },
  breadcrumbScroller: {
    flex: 1,
    minWidth: 0
  },
  breadcrumbContent: {
    alignItems: 'center',
    gap: spacing.xs,
    paddingRight: spacing.xs
  },
  breadcrumbSegment: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.xs
  },
  breadcrumbSeparator: {
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '600',
    letterSpacing: 0
  },
  breadcrumbButton: {
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 44,
    minWidth: 44
  },
  breadcrumbButtonPressed: {
    opacity: 0.62
  },
  breadcrumbText: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.sm,
    fontSize: 12,
    letterSpacing: 0,
    maxWidth: 240,
    overflow: 'hidden',
    paddingHorizontal: spacing.xs,
    paddingVertical: 4
  },
  detailBreadcrumbText: {
    fontSize: 15,
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  },
  ancestorBreadcrumbText: {
    color: colors.textMuted,
    fontWeight: '600'
  },
  immediateBreadcrumbText: {
    color: colors.accentStrong,
    fontWeight: '700'
  },
  checkoutImageBadge: {
    position: 'absolute',
    right: spacing.xs,
    top: spacing.xs,
    backgroundColor: colors.warningSurface,
    borderRadius: radius.sm,
    color: colors.warning,
    fontSize: 11,
    fontWeight: '700',
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.xs,
    paddingVertical: 4
  },
  title: {
    color: colors.text,
    fontSize: 16,
    fontWeight: '700',
    letterSpacing: 0,
    lineHeight: 21
  },
  standardTitle: {
    minHeight: 40
  },
  compactTitle: {
    fontSize: 17,
    lineHeight: 23
  },
  description: {
    color: colors.textMuted,
    fontSize: 13,
    lineHeight: 18,
    minHeight: 36
  },
  matchMeta: {
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '600',
    letterSpacing: 0,
    lineHeight: 16
  },
  footerAction: {
    alignItems: 'center',
    backgroundColor: colors.action,
    borderRadius: radius.sm,
    justifyContent: 'center',
    margin: spacing.sm,
    minHeight: 44
  },
  footerActionPressed: {
    backgroundColor: colors.actionPressed
  },
  disabledFooterAction: {
    opacity: 0.55
  },
  footerActionText: {
    color: colors.onAction,
    fontSize: 14,
    fontWeight: '700',
    letterSpacing: 0
  },
  });
}
