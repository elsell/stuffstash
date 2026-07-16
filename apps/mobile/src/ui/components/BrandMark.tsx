import { Image, StyleSheet, Text, View } from 'react-native';
import { radius, spacing, typography, type MobileColorPalette } from '../theme/tokens';
import { useAppearanceAwarePalette } from '../theme/appearance';
import glyph from '../../../assets/brand/stuff-stash-glyph.png';

type BrandMarkProps = {
  readonly size?: 'sm' | 'md';
  readonly showWordmark?: boolean;
};

export function BrandMark({ size = 'md', showWordmark = false }: BrandMarkProps) {
  const styles = createStyles(useAppearanceAwarePalette());
  const imageStyle = size === 'sm' ? styles.imageSmall : styles.image;

  return (
    <View style={styles.row}>
      <Image accessibilityIgnoresInvertColors source={glyph} style={imageStyle} />
      {showWordmark ? <Text style={styles.wordmark}>Stuff Stash</Text> : null}
    </View>
  );
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
  row: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.sm
  },
  image: {
    borderRadius: radius.sm,
    height: 36,
    width: 36
  },
  imageSmall: {
    borderRadius: radius.sm,
    height: 28,
    width: 28
  },
  wordmark: {
    color: colors.brandCharcoalDeep,
    fontSize: 17,
    fontWeight: typography.wordmarkWeight,
    letterSpacing: 0
  }
  });
}
