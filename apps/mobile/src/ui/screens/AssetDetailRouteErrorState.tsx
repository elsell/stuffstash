import { NativeCommandButton } from '../components/NativeCommandButton';
import { ScrollView, StyleSheet, Text } from 'react-native';
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
        <NativeCommandButton label="Retry asset" onPress={onRetry} />
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
  });
}
