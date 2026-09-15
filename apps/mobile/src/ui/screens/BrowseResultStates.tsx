import { NativeCommandButton } from '../components/NativeCommandButton';
import { StyleSheet, Text, View } from 'react-native';
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
      {presentation.actionLabel && presentation.onAction ? (
        <NativeCommandButton label={presentation.actionLabel} onPress={presentation.onAction} prominence="primary" />
      ) : null}
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
      <NativeCommandButton label="Retry" onPress={onRetry} prominence="primary" />
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
      <NativeCommandButton label="Try again" onPress={onRetry} />
    </View>
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
      alignItems: 'stretch',
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
    paginationFooter: {
      alignItems: 'stretch',
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
