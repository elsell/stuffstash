import { Image, Pressable, StyleSheet, Text, View } from 'react-native';
import { ChevronRight, MoveRight, Plus } from 'lucide-react-native';
import type {
  AssetCardViewModel,
  AssetDetailViewModel
} from '../../application/assets/AssetViewModels';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { useAppearanceAwarePalette } from '../theme/appearance';
import {
  AssetDetailAvailabilityButton,
  AssetDetailMaintenanceBar
} from './AssetDetailIdentitySection';
import {
  assetDetailAvailabilityAction,
  assetDetailMaintenanceActions
} from './AssetDetailPresentation';
import {
  canUseContainedAssetAction,
  containedAssetActions,
  containedAssetRows,
  containedAssetsEmptyState,
  containedAssetsSectionHeading,
  containedItemsEmptyState,
  containedItemsSectionHeading,
  containedSpacesEmptyState,
  containedSpacesSectionHeading,
  type ContainedAssetAction,
  type ContainedAssetsEmptyState,
  type ContainedAssetsSectionHeading,
  type ContainedAssetRowViewModel
} from './ContainedAssetsPresentation';
import { AppTextInput } from './AppTextInput';

export type ContainedWorkspaceListItem =
  | { readonly key: string; readonly kind: 'section'; readonly heading: ContainedAssetsSectionHeading }
  | { readonly key: string; readonly kind: 'row'; readonly row: ContainedAssetRowViewModel }
  | {
      readonly key: string;
      readonly kind: 'empty';
      readonly emptyState: ContainedAssetsEmptyState;
      readonly canClearSearch?: boolean;
    };

export function shouldShowContainedContentsSearch(asset: AssetDetailViewModel): boolean {
  return asset.kind === 'location' && asset.containedSpaces.length + asset.containedItems.length >= 20;
}

export function containedWorkspaceItems(
  asset: AssetDetailViewModel,
  query: string
): readonly ContainedWorkspaceListItem[] {
  if (asset.kind !== 'location') {
    return containedSectionItems(
      'contained',
      containedAssetsSectionHeading(asset),
      containedAssetRows(asset.containedAssets),
      containedAssetsEmptyState(asset)
    );
  }

  const normalizedQuery = query.trim().toLocaleLowerCase();
  const spaces = filterContainedAssets(asset.containedSpaces, normalizedQuery);
  const items = filterContainedAssets(asset.containedItems, normalizedQuery);
  const isFiltering = normalizedQuery.length > 0;
  const noMatches = isFiltering && spaces.length + items.length === 0;
  const spacesHeading = containedSpacesSectionHeading(asset, isFiltering ? {
    visibleCount: spaces.length,
    totalCount: asset.containedSpaces.length
  } : undefined);
  const itemsHeading = containedItemsSectionHeading(asset, isFiltering ? {
    visibleCount: items.length,
    totalCount: asset.containedItems.length
  } : undefined);

  return [
    ...containedSectionItems(
      'spaces',
      spacesHeading,
      containedAssetRows(spaces),
      isFiltering
        ? { title: 'No matching spaces', message: 'Try another name or path.' }
        : containedSpacesEmptyState()
    ),
    ...containedSectionItems(
      'items',
      itemsHeading,
      containedAssetRows(items),
      isFiltering
        ? { title: 'No matching items', message: 'Try another name or path.' }
        : containedItemsEmptyState(asset),
      noMatches
    )
  ];
}

function filterContainedAssets<T extends AssetCardViewModel & { readonly relativePathLabel?: string }>(
  assets: readonly T[],
  normalizedQuery: string
): readonly T[] {
  if (normalizedQuery.length === 0) {
    return assets;
  }
  return assets.filter((candidate) => [candidate.title, candidate.relativePathLabel]
    .some((value) => value?.toLocaleLowerCase().includes(normalizedQuery)));
}

function containedSectionItems(
  sectionKey: string,
  heading: ContainedAssetsSectionHeading,
  rows: readonly ContainedAssetRowViewModel[],
  emptyState: ContainedAssetsEmptyState,
  canClearSearch = false
): readonly ContainedWorkspaceListItem[] {
  return [
    { key: `${sectionKey}-heading`, kind: 'section', heading },
    ...(rows.length > 0
      ? rows.map((row) => ({ key: `${sectionKey}-${row.id}`, kind: 'row' as const, row }))
      : [{ key: `${sectionKey}-empty`, kind: 'empty' as const, emptyState, canClearSearch }])
  ];
}

export function containedAssetRowAccessibilityLabel(asset: ContainedAssetRowViewModel): string {
  return [`Open asset ${asset.title}`, asset.eyebrowLabel, asset.supportingLabel]
    .filter((value) => value.trim().length > 0)
    .join('. ');
}

export function ContainedContentsSearch({
  onChangeQuery,
  query
}: {
  readonly onChangeQuery: (query: string) => void;
  readonly query: string;
}) {
  const palette = useAppearanceAwarePalette();
  const styles = createStyles(palette);
  return (
    <View style={styles.contentsSearch}>
      <AppTextInput
        accessibilityLabel="Search contents"
        autoCapitalize="none"
        onChangeText={onChangeQuery}
        placeholder="Search this place"
        placeholderTextColor={palette.textMuted}
        returnKeyType="search"
        style={styles.contentsSearchInput}
        value={query}
      />
      {query.length > 0 ? (
        <Pressable
          accessibilityLabel="Clear contents search"
          accessibilityRole="button"
          onPress={() => onChangeQuery('')}
          style={styles.clearSearchButton}
        >
          <Text style={styles.clearSearchText}>Clear</Text>
        </Pressable>
      ) : null}
    </View>
  );
}

export function ContainedSpatialActions({
  asset,
  isActionPending,
  onAddHere,
  onMoveThingsHere
}: {
  readonly asset: AssetDetailViewModel;
  readonly isActionPending: boolean;
  readonly onAddHere?: () => void;
  readonly onMoveThingsHere?: () => void;
}) {
  const styles = createStyles(useAppearanceAwarePalette());
  const actions = containedAssetActions(asset);
  return (
    <View
      accessibilityLabel={asset.kind === 'location' ? 'Place items in this place' : 'Place items in this container'}
      style={styles.spatialActions}
    >
      {actions.map((action) => (
        <ContainedAssetActionButton
          action={action}
          isActionPending={isActionPending}
          key={action.kind}
          onPress={action.kind === 'add_here' ? onAddHere : onMoveThingsHere}
        />
      ))}
    </View>
  );
}

export function ContainedWorkspaceMaintenance({
  asset,
  isActionPending,
  onCheckout,
  onEdit,
  onMove,
  onReturn
}: {
  readonly asset: AssetDetailViewModel;
  readonly isActionPending: boolean;
  readonly onCheckout?: () => void;
  readonly onEdit?: () => void;
  readonly onMove?: () => void;
  readonly onReturn?: () => void;
}) {
  const styles = createStyles(useAppearanceAwarePalette());
  const hasAvailabilityAction = assetDetailAvailabilityAction(asset) !== undefined;
  const hasMaintenanceAction = assetDetailMaintenanceActions(asset)
    .some((action) => action.id === 'edit' || action.id === 'move');
  if (!hasAvailabilityAction && !hasMaintenanceAction) {
    return null;
  }
  return (
    <View accessibilityLabel="Manage this asset" style={styles.maintenanceSection}>
      <AssetDetailAvailabilityButton
        asset={asset}
        isActionPending={isActionPending}
        onCheckout={onCheckout}
        onReturn={onReturn}
        quiet
      />
      <AssetDetailMaintenanceBar
        asset={asset}
        isActionPending={isActionPending}
        onEdit={onEdit}
        onMove={onMove}
      />
    </View>
  );
}

function ContainedAssetActionButton({
  action,
  isActionPending,
  onPress
}: {
  readonly action: ContainedAssetAction;
  readonly isActionPending: boolean;
  readonly onPress?: () => void;
}) {
  const palette = useAppearanceAwarePalette();
  const styles = createStyles(palette);
  const enabled = canUseContainedAssetAction({ isActionPending, onPress });
  return (
    <Pressable
      accessibilityLabel={action.label}
      accessibilityRole="button"
      accessibilityState={{ disabled: !enabled }}
      disabled={!enabled}
      onPress={onPress}
      style={({ pressed }) => [
        styles.spatialAction,
        action.isPrimary ? styles.primarySpatialAction : styles.secondarySpatialAction,
        pressed ? styles.spatialActionPressed : null,
        !enabled ? styles.disabledAction : null
      ]}
    >
      {action.kind === 'add_here'
        ? <Plus color={action.isPrimary ? palette.onAction : palette.action} size={19} />
        : <MoveRight color={palette.action} size={19} />}
      <Text style={action.isPrimary ? styles.primarySpatialText : styles.secondarySpatialText}>
        {action.label}
      </Text>
    </Pressable>
  );
}

export function ContainedWorkspaceListItemView({
  item,
  onChildPress,
  onClearSearch
}: {
  readonly item: ContainedWorkspaceListItem;
  readonly onChildPress?: (assetId: string) => void;
  readonly onClearSearch: () => void;
}) {
  const styles = createStyles(useAppearanceAwarePalette());
  if (item.kind === 'section') {
    return (
      <View style={styles.sectionHeading}>
        <Text accessibilityRole="header" style={styles.sectionTitle}>{item.heading.title}</Text>
        <Text style={styles.sectionSummary}>{item.heading.summary}</Text>
      </View>
    );
  }
  if (item.kind === 'empty') {
    return (
      <ContainedAssetsEmptyState
        emptyState={item.emptyState}
        onClearSearch={item.canClearSearch ? onClearSearch : undefined}
      />
    );
  }
  return (
    <ContainedAssetRowView
      asset={item.row}
      onPress={onChildPress ? () => onChildPress(item.row.id) : undefined}
    />
  );
}

function ContainedAssetsEmptyState({
  emptyState,
  onClearSearch
}: {
  readonly emptyState: ContainedAssetsEmptyState;
  readonly onClearSearch?: () => void;
}) {
  const styles = createStyles(useAppearanceAwarePalette());
  return (
    <View style={styles.emptyContainer}>
      <Text style={styles.emptyContainerTitle}>{emptyState.title}</Text>
      <Text style={styles.emptyContainerText}>{emptyState.message}</Text>
      {onClearSearch ? (
        <Pressable accessibilityRole="button" onPress={onClearSearch} style={styles.emptyClearButton}>
          <Text style={styles.emptyClearText}>Clear search</Text>
        </Pressable>
      ) : null}
    </View>
  );
}

function ContainedAssetRowView({
  asset,
  onPress
}: {
  readonly asset: ContainedAssetRowViewModel;
  readonly onPress?: () => void;
}) {
  const palette = useAppearanceAwarePalette();
  const styles = createStyles(palette);
  return (
    <View>
      <Pressable
        accessibilityLabel={containedAssetRowAccessibilityLabel(asset)}
        accessibilityRole="button"
        disabled={!onPress}
        onPress={onPress}
        style={({ pressed }) => [styles.childRow, pressed ? styles.childRowPressed : null]}
      >
        <View style={styles.childPhoto}>
          {asset.photo ? (
            <Image
              accessibilityIgnoresInvertColors
              accessible={false}
              source={{ uri: asset.photo.uri, headers: asset.photo.headers }}
              style={styles.childPhotoImage}
            />
          ) : (
            <Text allowFontScaling={false} accessible={false} style={styles.childPhotoPlaceholder}>
              {asset.imagePlaceholderLabel}
            </Text>
          )}
        </View>
        <View style={styles.childRowText}>
          <Text style={styles.childTitle}>{asset.title}</Text>
          <Text style={styles.childEyebrow}>{asset.eyebrowLabel}</Text>
          {asset.supportingLabel ? <Text style={styles.childSupporting}>{asset.supportingLabel}</Text> : null}
        </View>
        <ChevronRight accessible={false} color={palette.textMuted} size={20} />
      </Pressable>
      <View style={styles.childSeparator} />
    </View>
  );
}

function createStyles(palette: MobileColorPalette) {
  return StyleSheet.create({
    spatialActions: { gap: spacing.md },
    maintenanceSection: {
      borderTopColor: palette.border,
      borderTopWidth: 1,
      gap: spacing.md,
      paddingTop: spacing.lg
    },
    contentsSearch: {
      alignItems: 'center',
      backgroundColor: palette.elevatedSurface,
      borderColor: palette.controlBorder,
      borderRadius: radius.md,
      borderWidth: 1,
      flexDirection: 'row',
      minHeight: 44,
      paddingLeft: spacing.sm
    },
    contentsSearchInput: {
      color: palette.text,
      flex: 1,
      fontSize: 16,
      minHeight: 44,
      paddingVertical: spacing.sm
    },
    clearSearchButton: {
      alignItems: 'center',
      justifyContent: 'center',
      minHeight: 44,
      minWidth: 60,
      paddingHorizontal: spacing.sm
    },
    clearSearchText: { color: palette.action, fontSize: 15, fontWeight: '600' },
    sectionHeading: { gap: 3, paddingBottom: spacing.sm, paddingTop: spacing.lg },
    sectionTitle: { color: palette.text, fontSize: 22, fontWeight: '700' },
    sectionSummary: { color: palette.textMuted, fontSize: 14, fontWeight: '500' },
    spatialAction: {
      alignItems: 'center',
      borderRadius: radius.md,
      flexDirection: 'row',
      gap: spacing.sm,
      justifyContent: 'center',
      minHeight: 50,
      paddingHorizontal: spacing.md,
      paddingVertical: spacing.sm
    },
    primarySpatialAction: { backgroundColor: palette.action },
    secondarySpatialAction: {
      backgroundColor: palette.elevatedSurface,
      borderColor: palette.border,
      borderWidth: StyleSheet.hairlineWidth
    },
    spatialActionPressed: { opacity: 0.82 },
    primarySpatialText: {
      color: palette.onAction,
      flexShrink: 1,
      fontSize: 17,
      fontWeight: '600',
      textAlign: 'center'
    },
    secondarySpatialText: {
      color: palette.action,
      flexShrink: 1,
      fontSize: 17,
      fontWeight: '600',
      textAlign: 'center'
    },
    childRow: {
      alignItems: 'center',
      backgroundColor: palette.surface,
      flexDirection: 'row',
      gap: spacing.sm,
      minHeight: 88,
      paddingVertical: spacing.sm
    },
    childRowPressed: { opacity: 0.82 },
    childSeparator: {
      backgroundColor: palette.border,
      height: StyleSheet.hairlineWidth,
      marginLeft: 76
    },
    childPhoto: {
      alignItems: 'center',
      aspectRatio: 1,
      backgroundColor: palette.elevatedSurface,
      borderColor: palette.border,
      borderRadius: radius.sm,
      borderWidth: 1,
      justifyContent: 'center',
      overflow: 'hidden',
      width: 64
    },
    childPhotoImage: { height: '100%', width: '100%' },
    childPhotoPlaceholder: { color: palette.accentStrong, fontSize: 15, fontWeight: '600' },
    childRowText: { flex: 1, gap: 2 },
    childEyebrow: { color: palette.textMuted, fontSize: 13, fontWeight: '500' },
    childTitle: { color: palette.text, fontSize: 17, fontWeight: '600' },
    childSupporting: { color: palette.textMuted, fontSize: 14 },
    emptyContainer: { gap: spacing.xs, paddingBottom: spacing.md, paddingTop: spacing.sm },
    emptyContainerTitle: { color: palette.text, fontSize: 17, fontWeight: '600' },
    emptyContainerText: { color: palette.textMuted, fontSize: 15 },
    emptyClearButton: {
      alignItems: 'center',
      alignSelf: 'flex-start',
      justifyContent: 'center',
      minHeight: 44,
      paddingHorizontal: spacing.sm
    },
    emptyClearText: { color: palette.action, fontSize: 15, fontWeight: '600' },
    disabledAction: { opacity: 0.55 }
  });
}
