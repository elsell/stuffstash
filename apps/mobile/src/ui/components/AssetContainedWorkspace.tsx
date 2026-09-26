import { NativeActionMenu } from './NativeActionMenu';
import type { ReactElement } from 'react';
import { Image, Pressable, StyleSheet, Text, View } from 'react-native';
import { ChevronRight } from 'lucide-react-native';
import type {
  AssetCardViewModel,
  AssetDetailViewModel
} from '../../application/assets/AssetViewModels';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { useAppearanceAwarePalette } from '../theme/appearance';

import {
  canUseContainedAssetAction,
  containedAssetActions,
  containedAssetRows,
  containedAssetsEmptyState,
  containedAssetsSectionHeading,
  containedItemsEmptyState,
  containedItemsSectionHeading,
  containedSpacesSectionHeading,
  type ContainedAssetsEmptyState,
  type ContainedAssetsSectionHeading,
  type ContainedAssetRowViewModel
} from './ContainedAssetsPresentation';
import { NativeCommandButton } from './NativeCommandButton';

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
  if (spaces.length + items.length === 0) {
    return [{ key: 'contents-empty', kind: 'empty', canClearSearch: isFiltering,
      emptyState: isFiltering
        ? { title: 'No matching contents', message: 'Try another name or path.' }
        : containedAssetsEmptyState(asset) }];
  }
  const spacesHeading = containedSpacesSectionHeading(asset, isFiltering ? {
    visibleCount: spaces.length,
    totalCount: asset.containedSpaces.length
  } : undefined);
  const itemsHeading = containedItemsSectionHeading(asset, isFiltering ? {
    visibleCount: items.length,
    totalCount: asset.containedItems.length
  } : undefined);

  return [
    ...(spaces.length > 0 ? containedSectionItems(
      'spaces', spacesHeading, containedAssetRows(spaces), containedAssetsEmptyState(asset)
    ) : []),
    ...(items.length > 0 ? containedSectionItems(
      'items', itemsHeading, containedAssetRows(items), containedItemsEmptyState(asset)
    ) : [])
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
  const actions = containedAssetActions(asset);
  if (actions.length === 0) return null;
  return <NativeActionMenu accessibilityLabel="Add to contents" disabled={isActionPending}
    trigger={{ kind: 'label', label: 'Add' }} groups={[{ id: 'contents', items: actions.map(action => {
      const onPress = action.kind === 'add_here' ? onAddHere : onMoveThingsHere;
      return { id: action.kind, label: action.label, disabled: !canUseContainedAssetAction({ isActionPending, onPress }), onPress: () => { if (!isActionPending) onPress?.(); } };
    }) }]} />;
}

export function ContainedWorkspaceListItemView({
  item,
  actions,
  onChildPress,
  onClearSearch
}: {
  readonly item: ContainedWorkspaceListItem;
  readonly actions?: ReactElement;
  readonly onChildPress?: (assetId: string) => void;
  readonly onClearSearch: () => void;
}) {
  const styles = createStyles(useAppearanceAwarePalette());
  if (item.kind === 'section') {
    return (
      <View style={styles.sectionHeading}>
        <Text accessibilityRole="header" style={styles.sectionTitle}>{item.heading.title}</Text>
        <Text style={styles.sectionSummary}>{item.heading.summary}</Text>
        {actions}
      </View>
    );
  }
  if (item.kind === 'empty') {
    return (
      <View style={styles.emptySection}>
        {actions}
        <ContainedAssetsEmptyState
          emptyState={item.emptyState}
          onClearSearch={item.canClearSearch ? onClearSearch : undefined}
        />
      </View>
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
        <NativeCommandButton label="Clear search" onPress={onClearSearch} />
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
    sectionHeading: { gap: spacing.xs, paddingBottom: spacing.sm, paddingTop: spacing.sm },
    sectionTitle: { color: palette.text, fontSize: 22, fontWeight: '700' },
    sectionSummary: { color: palette.textMuted, fontSize: 14, fontWeight: '500' },
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
    emptySection: { gap: spacing.sm },
    emptyContainer: { gap: spacing.xs, paddingBottom: spacing.md, paddingTop: spacing.sm },
    emptyContainerTitle: { color: palette.text, fontSize: 17, fontWeight: '600' },
    emptyContainerText: { color: palette.textMuted, fontSize: 15 },
  });
}
