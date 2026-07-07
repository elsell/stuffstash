import { Image, Pressable, StyleSheet, Text, View } from 'react-native';
import type { AssetCardViewModel } from '../../application/assets/AssetViewModels';
import { colors, radius, spacing } from '../theme/tokens';
import { AssetTagChips } from './AssetTagChips';

type AssetCardProps = {
  readonly asset: AssetCardViewModel;
  readonly onPress: () => void;
};

export function AssetCard({ asset, onPress }: AssetCardProps) {
  return (
    <Pressable accessibilityRole="button" onPress={onPress} style={styles.card}>
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
      </View>
      <View style={styles.body}>
        <View style={styles.badgeRow}>
          <Text style={styles.kindBadge}>{asset.kindLabel}</Text>
          {asset.customTypeLabel ? <Text style={styles.typeBadge}>{asset.customTypeLabel}</Text> : null}
          {asset.checkedOutLabel ? <Text style={styles.checkoutBadge}>{asset.checkedOutLabel}</Text> : null}
        </View>
        <Text numberOfLines={2} style={styles.title}>
          {asset.title}
        </Text>
        <Text numberOfLines={2} style={styles.description}>
          {asset.description}
        </Text>
        <Text numberOfLines={1} style={styles.meta}>
          {asset.locationTrailLabel}
        </Text>
        <AssetTagChips tags={asset.tags} />
        <View style={styles.footer}>
          <Text style={asset.photoLabel === 'Photo ready' ? styles.photoReady : styles.photoNeeded}>
            {asset.photoLabel}
          </Text>
          <Text style={styles.updated}>{asset.updatedAtLabel}</Text>
        </View>
      </View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    flex: 1,
    minHeight: 286,
    overflow: 'hidden'
  },
  imageFrame: {
    alignItems: 'center',
    aspectRatio: 1,
    backgroundColor: colors.surfaceMuted,
    justifyContent: 'center',
    width: '100%'
  },
  imagePlaceholder: {
    color: colors.accentStrong,
    fontSize: 22,
    fontWeight: '900',
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
  badgeRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.xs,
    minHeight: 24
  },
  kindBadge: {
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
  typeBadge: {
    borderColor: colors.border,
    borderRadius: radius.sm,
    borderWidth: 1,
    color: colors.textMuted,
    fontSize: 11,
    fontWeight: '800',
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.xs,
    paddingVertical: 4
  },
  checkoutBadge: {
    backgroundColor: colors.warningSurface,
    borderRadius: radius.sm,
    color: colors.warning,
    fontSize: 11,
    fontWeight: '800',
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.xs,
    paddingVertical: 4
  },
  title: {
    color: colors.text,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 20,
    minHeight: 40
  },
  description: {
    color: colors.textMuted,
    fontSize: 12,
    lineHeight: 17,
    minHeight: 34
  },
  meta: {
    color: colors.textMuted,
    fontSize: 12,
    lineHeight: 16
  },
  footer: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.xs,
    justifyContent: 'space-between',
    minHeight: 24
  },
  photoReady: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.sm,
    color: colors.accentStrong,
    fontSize: 10,
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
    fontSize: 10,
    fontWeight: '800',
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.xs,
    paddingVertical: 4
  },
  updated: {
    color: colors.textMuted,
    flex: 1,
    fontSize: 10,
    letterSpacing: 0,
    textAlign: 'right'
  }
});
