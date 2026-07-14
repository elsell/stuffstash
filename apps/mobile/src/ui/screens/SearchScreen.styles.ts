import { StyleSheet } from 'react-native';
import type { MobileColorPalette } from '../theme/tokens';
import { spacing } from '../theme/tokens';

export function createSearchScreenStyles(palette: MobileColorPalette) {
  return StyleSheet.create({
    shell: {
      backgroundColor: palette.background,
      flex: 1
    },
    content: {
      paddingBottom: spacing.xl,
      paddingHorizontal: spacing.md,
      paddingTop: spacing.sm
    },
    cardRow: {
      gap: spacing.sm,
      marginBottom: spacing.sm
    },
    singleCardRow: {
      marginBottom: spacing.sm
    },
    footer: {
      alignItems: 'center',
      minHeight: 56,
      justifyContent: 'center',
      paddingVertical: spacing.sm
    }
  });
}
