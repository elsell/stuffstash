import { Image, Pressable, StyleSheet, Text, View } from 'react-native';
import type { BrowsePlaceItemViewModel } from './SearchScreenPresentation';
import {
  radius,
  spacing,
  type MobileColorPalette
} from '../theme/tokens';

type BrowsePlaceRowProps = {
  readonly location: BrowsePlaceItemViewModel;
  readonly palette: MobileColorPalette;
  readonly onPress: () => void;
};

export function BrowsePlaceRow({ location, palette, onPress }: BrowsePlaceRowProps) {
  const styles = stylesForPalette(palette);
  const accessibilityHint = [location.description, location.recentAssetLabel]
    .filter((value) => value.trim().length > 0)
    .join('. ');

  return (
    <Pressable
      accessibilityLabel={`Open place ${location.title}, ${location.containedAssetCountLabel}`}
      accessibilityHint={accessibilityHint || undefined}
      accessibilityRole="button"
      onPress={onPress}
      style={({ pressed }) => [styles.row, pressed ? styles.rowPressed : undefined]}
    >
      <View style={styles.imageFrame}>
        {location.photo ? (
          <Image
            accessibilityIgnoresInvertColors
            source={{ uri: location.photo.uri, headers: location.photo.headers }}
            style={styles.image}
          />
        ) : (
          <Text style={styles.imagePlaceholder}>Place</Text>
        )}
      </View>
      <View style={styles.body}>
        <Text numberOfLines={2} style={styles.title}>{location.title}</Text>
        {location.description ? (
          <Text numberOfLines={2} style={styles.description}>{location.description}</Text>
        ) : null}
        <Text style={styles.count}>{location.containedAssetCountLabel}</Text>
        <Text numberOfLines={1} style={styles.recentAssets}>{location.recentAssetLabel}</Text>
      </View>
    </Pressable>
  );
}

function createStyles(palette: MobileColorPalette) {
  return StyleSheet.create({
    row: {
      backgroundColor: palette.surface,
      borderColor: palette.border,
      borderRadius: radius.md,
      borderWidth: StyleSheet.hairlineWidth,
      flexDirection: 'row',
      gap: spacing.sm,
      marginBottom: spacing.sm,
      minHeight: 112,
      overflow: 'hidden',
      padding: spacing.sm
    },
    rowPressed: {
      opacity: 0.82
    },
    imageFrame: {
      alignItems: 'center',
      aspectRatio: 1,
      backgroundColor: palette.surfaceMuted,
      borderRadius: radius.sm,
      justifyContent: 'center',
      overflow: 'hidden',
      width: 92
    },
    imagePlaceholder: {
      color: palette.accentStrong,
      fontSize: 18,
      fontWeight: '700'
    },
    image: {
      height: '100%',
      width: '100%'
    },
    body: {
      flex: 1,
      gap: 3,
      justifyContent: 'center',
      minWidth: 0
    },
    title: {
      color: palette.text,
      fontSize: 17,
      fontWeight: '600',
      lineHeight: 22
    },
    description: {
      color: palette.textMuted,
      fontSize: 13,
      lineHeight: 18
    },
    count: {
      color: palette.accentStrong,
      fontSize: 13,
      fontWeight: '600',
      lineHeight: 18
    },
    recentAssets: {
      color: palette.textMuted,
      fontSize: 13,
      lineHeight: 18
    }
  });
}

const styleCache = new Map<MobileColorPalette, ReturnType<typeof createStyles>>();

function stylesForPalette(palette: MobileColorPalette) {
  const cached = styleCache.get(palette);
  if (cached) return cached;
  const styles = createStyles(palette);
  styleCache.set(palette, styles);
  return styles;
}
