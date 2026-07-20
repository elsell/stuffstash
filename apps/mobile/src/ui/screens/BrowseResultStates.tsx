import { Pressable, StyleSheet, Text, View } from 'react-native';
import {
  radius,
  spacing,
  type MobileColorPalette
} from '../theme/tokens';

type BrowseEmptyStateProps =
  | {
      readonly kind: 'inventory';
      readonly inventoryName: string;
      readonly palette: MobileColorPalette;
      readonly onAdd?: () => void;
    }
  | {
      readonly kind: 'search';
      readonly query: string;
      readonly palette: MobileColorPalette;
      readonly onClearSearch: () => void;
    }
  | {
      readonly kind: 'filters';
      readonly palette: MobileColorPalette;
      readonly onClearFilters: () => void;
    };

export function BrowseEmptyState(props: BrowseEmptyStateProps) {
  const styles = stylesForPalette(props.palette);
  const presentation = emptyStatePresentation(props);

  return (
    <View accessibilityLiveRegion="polite" style={styles.statePanel}>
      <Text style={styles.title}>{presentation.title}</Text>
      <Text style={styles.message}>{presentation.message}</Text>
      {presentation.actionLabel && presentation.onAction
        ? browseStateAction({
            label: presentation.actionLabel,
            onPress: presentation.onAction,
            styles
          })
        : null}
    </View>
  );
}

export function BrowseLoadError({
  message,
  palette,
  onRetry
}: {
  readonly message: string;
  readonly palette: MobileColorPalette;
  readonly onRetry: () => void;
}) {
  const styles = stylesForPalette(palette);

  return (
    <View accessibilityLiveRegion="polite" style={styles.statePanel}>
      <Text style={styles.title}>Could not load this inventory</Text>
      <Text style={styles.message}>{message}</Text>
      {browseStateAction({ label: 'Retry', onPress: onRetry, styles })}
    </View>
  );
}

export function BrowsePaginationRetry({
  message,
  palette,
  onRetry
}: {
  readonly message: string;
  readonly palette: MobileColorPalette;
  readonly onRetry: () => void;
}) {
  const styles = stylesForPalette(palette);

  return (
    <View accessibilityLiveRegion="polite" style={styles.paginationFooter}>
      <Text style={styles.paginationMessage}>{message}</Text>
      {browseStateAction({ label: 'Try again', onPress: onRetry, quiet: true, styles })}
    </View>
  );
}

function browseStateAction({
  label,
  onPress,
  styles,
  quiet = false
}: {
  readonly label: string;
  readonly onPress: () => void;
  readonly styles: ReturnType<typeof createStyles>;
  readonly quiet?: boolean;
}) {
  return (
    <Pressable
      accessibilityRole="button"
      onPress={onPress}
      style={({ pressed }) => [
        styles.action,
        quiet ? styles.quietAction : styles.primaryAction,
        pressed ? (quiet ? styles.quietActionPressed : styles.primaryActionPressed) : undefined
      ]}
    >
      <Text style={[styles.actionText, quiet ? styles.quietActionText : styles.primaryActionText]}>
        {label}
      </Text>
    </Pressable>
  );
}

function emptyStatePresentation(props: BrowseEmptyStateProps): {
  readonly title: string;
  readonly message: string;
  readonly actionLabel?: string;
  readonly onAction?: () => void;
} {
  switch (props.kind) {
    case 'inventory':
      return {
        title: `No items in ${props.inventoryName}`,
        message: props.onAdd
          ? 'Add your first item, container, or place.'
          : 'An inventory editor can add the first item, container, or place.',
        actionLabel: props.onAdd ? 'Add item' : undefined,
        onAction: props.onAdd
      };
    case 'search':
      return {
        title: `No results for “${props.query.trim()}”`,
        message: 'Try another search or clear it to browse everything.',
        actionLabel: 'Clear search',
        onAction: props.onClearSearch
      };
    case 'filters':
      return {
        title: 'No items match these filters',
        message: 'Remove a filter to see more of your inventory.',
        actionLabel: 'Clear filters',
        onAction: props.onClearFilters
      };
  }
}

function createStyles(palette: MobileColorPalette) {
  return StyleSheet.create({
    statePanel: {
      alignItems: 'flex-start',
      backgroundColor: palette.surface,
      borderColor: palette.border,
      borderRadius: radius.md,
      borderWidth: 1,
      gap: spacing.xs,
      padding: spacing.md
    },
    title: {
      color: palette.text,
      fontSize: 18,
      fontWeight: '600',
      lineHeight: 24
    },
    message: {
      color: palette.textMuted,
      fontSize: 15,
      lineHeight: 22
    },
    action: {
      alignItems: 'center',
      alignSelf: 'flex-start',
      borderRadius: radius.sm,
      justifyContent: 'center',
      marginTop: spacing.sm,
      minHeight: 44,
      paddingHorizontal: spacing.md
    },
    primaryAction: {
      backgroundColor: palette.action
    },
    primaryActionPressed: {
      backgroundColor: palette.actionPressed
    },
    quietAction: {
      backgroundColor: palette.surface,
      borderColor: palette.border,
      borderWidth: StyleSheet.hairlineWidth
    },
    quietActionPressed: {
      opacity: 0.82
    },
    actionText: {
      fontSize: 15,
      fontWeight: '600'
    },
    primaryActionText: {
      color: palette.onAction
    },
    quietActionText: {
      color: palette.action
    },
    paginationFooter: {
      alignItems: 'center',
      gap: spacing.xs,
      paddingBottom: spacing.sm,
      paddingTop: spacing.md
    },
    paginationMessage: {
      color: palette.textMuted,
      fontSize: 13,
      lineHeight: 18,
      textAlign: 'center'
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
