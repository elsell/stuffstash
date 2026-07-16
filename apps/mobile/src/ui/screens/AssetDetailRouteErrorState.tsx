import { Pressable, ScrollView, StyleSheet, Text } from 'react-native';
import { spacing, type MobileColorPalette } from '../theme/tokens';
import { useAppearanceAwarePalette } from '../theme/appearance';

export function AssetDetailRouteErrorState({
  canRetry,
  message,
  onRetry,
  title
}: {
  readonly canRetry: boolean;
  readonly message: string;
  readonly onRetry: () => void;
  readonly title: string;
}) {
  const styles = createStyles(useAppearanceAwarePalette());
  return (
    <ScrollView
      accessibilityLabel="Asset error"
      contentContainerStyle={styles.content}
      style={styles.scroller}
    >
      <Text accessibilityRole="header" style={styles.title}>{title}</Text>
      <Text style={styles.message}>{message}</Text>
      {canRetry ? (
        <Pressable accessibilityRole="button" onPress={onRetry} style={styles.retryButton}>
          <Text style={styles.retryButtonText}>Retry</Text>
        </Pressable>
      ) : null}
    </ScrollView>
  );
}

function createStyles(palette: MobileColorPalette) {
  return StyleSheet.create({
    scroller: { flex: 1 },
    content: {
      alignItems: 'center',
      flexGrow: 1,
      justifyContent: 'center',
      padding: spacing.lg
    },
    title: {
      color: palette.text,
      fontSize: 24,
      fontWeight: '800',
      letterSpacing: 0,
      textAlign: 'center'
    },
    message: {
      color: palette.textMuted,
      fontSize: 16,
      marginTop: spacing.md,
      textAlign: 'center'
    },
    retryButton: {
      alignItems: 'center',
      justifyContent: 'center',
      marginTop: spacing.lg,
      minHeight: 44,
      minWidth: 88,
      paddingHorizontal: spacing.md,
      paddingVertical: spacing.sm
    },
    retryButtonText: {
      color: palette.action,
      flexShrink: 1,
      fontSize: 17,
      fontWeight: '600',
      textAlign: 'center'
    }
  });
}
