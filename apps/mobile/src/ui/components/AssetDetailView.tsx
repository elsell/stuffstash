import { useState, type ReactElement } from 'react';
import type { RefreshControlProps } from 'react-native';
import {
  FlatList,
  Image,
  Pressable,
  StyleSheet,
  Text,
  TextInput,
  View
} from 'react-native';
import { ChevronRight, MoreHorizontal, MoveRight, Plus } from 'lucide-react-native';
import type {
  AssetCardViewModel,
  AssetDetailViewModel,
  AssetParentLocationCrumbViewModel,
  AssetTagViewModel
} from '../../application/assets/AssetViewModels';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { AssetDetailPhotoGallery } from './AssetDetailPhotoGallery';
import {
  AssetDetailAvailabilityButton,
  AssetDetailIdentitySection,
  AssetDetailMaintenanceBar
} from './AssetDetailIdentitySection';
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
import { assetDetailUpdatedMetadata } from './AssetDetailPresentation';

export type AssetPhotoUploadProgressViewModel = {
  readonly index: number;
  readonly fileName: string;
  readonly status: 'pending' | 'uploading' | 'attached' | 'failed';
};

type AssetDetailViewProps = {
  readonly asset: AssetDetailViewModel;
  readonly isActionPending?: boolean;
  readonly photoUploads?: readonly AssetPhotoUploadProgressViewModel[];
  readonly photoStatusMessage?: string;
  readonly workspaceStatusMessage?: string;
  readonly workspaceStatusKind?: 'success' | 'working';
  readonly canRetryPhotos?: boolean;
  readonly onBack?: () => void;
  readonly onEdit?: () => void;
  readonly onMove?: () => void;
  readonly onCheckout?: () => void;
  readonly onReturn?: () => void;
  readonly onTagPress?: (tag: AssetTagViewModel) => void;
  readonly onAddPhotos?: () => void;
  readonly onPhotoPress?: (photoId: string) => void;
  readonly onRetryPhotos?: () => void;
  readonly onMoreActions?: () => void;
  readonly onChildPress?: (assetId: string) => void;
  readonly onParentLocationPress?: (parent: AssetParentLocationCrumbViewModel) => void;
  readonly onAddHere?: () => void;
  readonly onMoveThingsHere?: () => void;
  readonly refreshControl?: ReactElement<RefreshControlProps>;
};

export function assetDetailNavigationTitle(asset: Pick<AssetDetailViewModel, 'kind'>): 'Place' | 'Details' {
  return asset.kind === 'location' ? 'Place' : 'Details';
}

type ContainedWorkspaceListItem =
  | { readonly key: string; readonly kind: 'section'; readonly heading: ContainedAssetsSectionHeading }
  | { readonly key: string; readonly kind: 'row'; readonly row: ContainedAssetRowViewModel }
  | {
      readonly key: string;
      readonly kind: 'empty';
      readonly emptyState: ContainedAssetsEmptyState;
      readonly canClearSearch?: boolean;
    };

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

export function AssetDetailView({
  asset,
  canRetryPhotos = false,
  isActionPending = false,
  onAddHere,
  onAddPhotos,
  onBack,
  onCheckout,
  onChildPress,
  onEdit,
  onMove,
  onMoreActions,
  onMoveThingsHere,
  onParentLocationPress,
  onPhotoPress,
  onRetryPhotos,
  onReturn,
  onTagPress,
  photoStatusMessage,
  photoUploads = [],
  refreshControl,
  workspaceStatusKind = 'success',
  workspaceStatusMessage
}: AssetDetailViewProps) {
  const palette = useAppearanceAwarePalette();
  const styles = createStyles(palette);
  const [contentsQuery, setContentsQuery] = useState('');
  const showContentsSearch = asset.kind === 'location'
    && asset.containedSpaces.length + asset.containedItems.length >= 20;
  const workspaceItems = asset.canContainAssets
    ? containedWorkspaceItems(asset, showContentsSearch ? contentsQuery : '')
    : [];
  const updatedMetadata = assetDetailUpdatedMetadata(asset);

  return (
    <FlatList
      contentContainerStyle={styles.content}
      data={workspaceItems}
      keyExtractor={(item) => item.key}
      refreshControl={refreshControl}
      renderItem={({ item }) => (
        <ContainedWorkspaceListItemView
          item={item}
          onChildPress={onChildPress}
          onClearSearch={() => setContentsQuery('')}
        />
      )}
      ListHeaderComponent={(
        <View style={styles.headerStack}>
          {onBack || onMoreActions ? (
            <View style={styles.topActions}>
              {onBack ? (
                <Pressable accessibilityRole="button" onPress={onBack} style={styles.backButton}>
                  <Text style={styles.backButtonText}>Back</Text>
                </Pressable>
              ) : <View />}
              {onMoreActions ? (
                <Pressable
                  accessibilityLabel="More actions"
                  accessibilityRole="button"
                  accessibilityState={{ disabled: isActionPending }}
                  disabled={isActionPending}
                  onPress={isActionPending ? undefined : onMoreActions}
                  style={[styles.moreButton, isActionPending ? styles.disabledAction : null]}
                >
                  <MoreHorizontal color={palette.text} size={22} />
                </Pressable>
              ) : null}
            </View>
          ) : null}

          <AssetDetailPhotoGallery
            canAddPhotos={!isActionPending && asset.canAddPhotos}
            contentHorizontalPadding={spacing.md}
            imagePlaceholderLabel={asset.imagePlaceholderLabel}
            onAddPhotos={onAddPhotos}
            onPhotoPress={onPhotoPress}
            photos={asset.photos}
            palette={palette}
          />

          <AssetDetailIdentitySection
            asset={asset}
            isActionPending={isActionPending}
            onCheckout={onCheckout}
            onEdit={onEdit}
            onMove={onMove}
            onParentLocationPress={onParentLocationPress}
            onReturn={onReturn}
            onTagPress={onTagPress}
            showAvailability={!asset.canContainAssets}
            showMaintenance={!asset.canContainAssets}
          />

          <StatusAndProgressSection
            canRetryPhotos={canRetryPhotos}
            onRetryPhotos={onRetryPhotos}
            photoStatusMessage={photoStatusMessage}
            photoUploads={photoUploads}
            workspaceStatusKind={workspaceStatusKind}
            workspaceStatusMessage={workspaceStatusMessage}
          />

          {asset.canContainAssets ? (
            <ContainedSpatialActions
              asset={asset}
              isActionPending={isActionPending}
              onAddHere={onAddHere}
              onMoveThingsHere={onMoveThingsHere}
            />
          ) : null}

          {showContentsSearch ? (
            <View style={styles.contentsSearch}>
              <TextInput
                accessibilityLabel="Search contents"
                autoCapitalize="none"
                onChangeText={setContentsQuery}
                placeholder="Search this place"
                placeholderTextColor={palette.textMuted}
                returnKeyType="search"
                style={styles.contentsSearchInput}
                value={contentsQuery}
              />
              {contentsQuery.length > 0 ? (
                <Pressable
                  accessibilityLabel="Clear contents search"
                  accessibilityRole="button"
                  onPress={() => setContentsQuery('')}
                  style={styles.clearSearchButton}
                >
                  <Text style={styles.clearSearchText}>Clear</Text>
                </Pressable>
              ) : null}
            </View>
          ) : null}
        </View>
      )}
      ListFooterComponent={(
        <View style={styles.footerStack}>
          {asset.canContainAssets ? (
            <ContainedWorkspaceMaintenance
              asset={asset}
              isActionPending={isActionPending}
              onCheckout={onCheckout}
              onEdit={onEdit}
              onMove={onMove}
              onReturn={onReturn}
            />
          ) : null}
          <Text accessibilityLabel={updatedMetadata.value} style={styles.updatedText}>
            {updatedMetadata.value}
          </Text>
        </View>
      )}
    />
  );
}

function ContainedSpatialActions({
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

function ContainedWorkspaceMaintenance({
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

function ContainedWorkspaceListItemView({
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
        accessibilityLabel={`Open asset ${asset.title}`}
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
            <Text style={styles.childPhotoPlaceholder}>{asset.imagePlaceholderLabel}</Text>
          )}
        </View>
        <View style={styles.childRowText}>
          <Text style={styles.childTitle}>{asset.title}</Text>
          <Text style={styles.childEyebrow}>{asset.eyebrowLabel}</Text>
          {asset.supportingLabel ? <Text style={styles.childSupporting}>{asset.supportingLabel}</Text> : null}
        </View>
        <ChevronRight color={palette.textMuted} size={20} />
      </Pressable>
      <View style={styles.childSeparator} />
    </View>
  );
}

function StatusAndProgressSection({
  canRetryPhotos,
  onRetryPhotos,
  photoStatusMessage,
  photoUploads,
  workspaceStatusKind,
  workspaceStatusMessage
}: {
  readonly canRetryPhotos: boolean;
  readonly onRetryPhotos?: () => void;
  readonly photoStatusMessage?: string;
  readonly photoUploads: readonly AssetPhotoUploadProgressViewModel[];
  readonly workspaceStatusKind: 'success' | 'working';
  readonly workspaceStatusMessage?: string;
}) {
  const styles = createStyles(useAppearanceAwarePalette());
  if (!workspaceStatusMessage && !photoStatusMessage && photoUploads.length === 0) {
    return null;
  }
  return (
    <View style={styles.statusSection}>
      {workspaceStatusMessage ? (
        <View
          accessible
          accessibilityLiveRegion="polite"
          accessibilityRole="alert"
          style={[
            styles.statusPanel,
            workspaceStatusKind === 'working' ? styles.workingStatusPanel : null
          ]}
        >
          <Text style={styles.statusText}>{workspaceStatusMessage}</Text>
        </View>
      ) : null}
      {photoStatusMessage ? (
        <View style={styles.statusPanel}>
          <Text style={styles.statusText}>{photoStatusMessage}</Text>
          {canRetryPhotos && onRetryPhotos ? (
            <Pressable accessibilityRole="button" onPress={onRetryPhotos} style={styles.retryButton}>
              <Text style={styles.retryButtonText}>Retry</Text>
            </Pressable>
          ) : null}
        </View>
      ) : null}
      {photoUploads.length > 0 ? <PhotoUploadProgressList uploads={photoUploads} /> : null}
    </View>
  );
}

function PhotoUploadProgressList({ uploads }: { readonly uploads: readonly AssetPhotoUploadProgressViewModel[] }) {
  const styles = createStyles(useAppearanceAwarePalette());
  return (
    <View style={styles.uploadPanel}>
      <Text style={styles.uploadPanelTitle}>Photo uploads</Text>
      {uploads.map((upload) => (
        <View key={`${upload.index.toString()}-${upload.fileName}`} style={styles.uploadRow}>
          <View style={styles.uploadText}>
            <Text style={styles.uploadFileName}>{upload.fileName}</Text>
            <Text style={styles.uploadStatusText}>{uploadStatusLabel(upload.status)}</Text>
          </View>
          <Text style={[styles.uploadPill, upload.status === 'failed' ? styles.uploadFailedPill : null]}>
            {uploadPillLabel(upload.status)}
          </Text>
        </View>
      ))}
    </View>
  );
}

function uploadStatusLabel(status: AssetPhotoUploadProgressViewModel['status']): string {
  switch (status) {
    case 'attached': return 'Attached to this asset';
    case 'failed': return 'Needs retry';
    case 'uploading': return 'Uploading original file';
    case 'pending': return 'Waiting to upload';
  }
}

function uploadPillLabel(status: AssetPhotoUploadProgressViewModel['status']): string {
  switch (status) {
    case 'attached': return 'Done';
    case 'failed': return 'Failed';
    case 'uploading': return 'Now';
    case 'pending': return 'Queued';
  }
}

function createStyles(palette: MobileColorPalette) {
  return StyleSheet.create({
  topActions: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'space-between'
  },
  moreButton: {
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 44,
    minWidth: 44
  },
  content: {
    paddingHorizontal: spacing.md,
    paddingTop: spacing.lg,
    paddingBottom: spacing.xl * 2
  },
  headerStack: {
    gap: spacing.lg,
    paddingBottom: spacing.lg
  },
  backButton: {
    alignSelf: 'flex-start',
    justifyContent: 'center',
    minHeight: 44,
    minWidth: 44
  },
  backButtonText: {
    color: palette.action,
    fontSize: 17,
    fontWeight: '600'
  },
  statusSection: {
    gap: spacing.sm
  },
  statusPanel: {
    backgroundColor: palette.brandDustyBlueSoft,
    borderRadius: radius.md,
    gap: spacing.sm,
    padding: spacing.md
  },
  workingStatusPanel: {
    backgroundColor: palette.surfaceMuted
  },
  statusText: {
    color: palette.text,
    fontSize: 15,
    fontWeight: '500',
    lineHeight: 21
  },
  retryButton: {
    alignItems: 'center',
    alignSelf: 'flex-start',
    backgroundColor: palette.surfaceMuted,
    borderRadius: radius.md,
    justifyContent: 'center',
    minHeight: 44,
    paddingHorizontal: spacing.md
  },
  retryButtonText: {
    color: palette.action,
    fontSize: 15,
    fontWeight: '600'
  },
  uploadPanel: {
    borderColor: palette.border,
    borderRadius: radius.md,
    borderWidth: 1,
    gap: spacing.sm,
    padding: spacing.md
  },
  uploadPanelTitle: {
    color: palette.textMuted,
    fontSize: 13,
    fontWeight: '600'
  },
  uploadRow: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.sm,
    justifyContent: 'space-between'
  },
  uploadText: {
    flex: 1,
    gap: 2
  },
  uploadFileName: {
    color: palette.text,
    fontSize: 15,
    fontWeight: '600'
  },
  uploadStatusText: {
    color: palette.textMuted,
    fontSize: 14,
    lineHeight: 19
  },
  uploadPill: {
    backgroundColor: palette.brandDustyBlueSoft,
    borderRadius: radius.sm,
    color: palette.accentStrong,
    fontSize: 13,
    fontWeight: '600',
    overflow: 'hidden',
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  },
  uploadFailedPill: {
    backgroundColor: palette.warningSurface,
    color: palette.warning
  },
  spatialActions: {
    gap: spacing.md
  },
  maintenanceSection: {
    borderTopColor: palette.border,
    borderTopWidth: 1,
    gap: spacing.md,
    paddingTop: spacing.lg
  },
  footerStack: {
    gap: spacing.lg,
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
  clearSearchText: {
    color: palette.action,
    fontSize: 15,
    fontWeight: '600'
  },
  sectionHeading: {
    gap: 3,
    paddingBottom: spacing.sm,
    paddingTop: spacing.lg
  },
  sectionTitle: {
    color: palette.text,
    fontSize: 22,
    fontWeight: '700',
    lineHeight: 28
  },
  sectionSummary: {
    color: palette.textMuted,
    fontSize: 14,
    fontWeight: '500',
    lineHeight: 20
  },
  spatialAction: {
    alignItems: 'center',
    borderRadius: radius.md,
    flexDirection: 'row',
    gap: spacing.sm,
    justifyContent: 'center',
    minHeight: 50,
    paddingHorizontal: spacing.md
  },
  primarySpatialAction: {
    backgroundColor: palette.action
  },
  secondarySpatialAction: {
    backgroundColor: palette.elevatedSurface,
    borderColor: palette.controlBorder,
    borderWidth: 1
  },
  spatialActionPressed: {
    opacity: 0.82
  },
  primarySpatialText: {
    color: palette.onAction,
    fontSize: 17,
    fontWeight: '600'
  },
  secondarySpatialText: {
    color: palette.action,
    fontSize: 17,
    fontWeight: '600'
  },
  childRow: {
    alignItems: 'center',
    backgroundColor: palette.surface,
    flexDirection: 'row',
    gap: spacing.sm,
    minHeight: 88,
    paddingVertical: spacing.sm
  },
  childRowPressed: {
    backgroundColor: palette.selected
  },
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
  childPhotoImage: {
    height: '100%',
    width: '100%'
  },
  childPhotoPlaceholder: {
    color: palette.accentStrong,
    fontSize: 15,
    fontWeight: '600'
  },
  childRowText: {
    flex: 1,
    gap: 2
  },
  childEyebrow: {
    color: palette.textMuted,
    fontSize: 13,
    fontWeight: '500'
  },
  childTitle: {
    color: palette.text,
    fontSize: 17,
    fontWeight: '600',
    lineHeight: 23
  },
  childSupporting: {
    color: palette.textMuted,
    fontSize: 14,
    lineHeight: 20
  },
  emptyContainer: {
    gap: spacing.xs,
    paddingBottom: spacing.md,
    paddingTop: spacing.sm
  },
  emptyContainerTitle: {
    color: palette.text,
    fontSize: 17,
    fontWeight: '600'
  },
  emptyContainerText: {
    color: palette.textMuted,
    fontSize: 15,
    lineHeight: 21
  },
  emptyClearButton: {
    alignItems: 'center',
    alignSelf: 'flex-start',
    justifyContent: 'center',
    minHeight: 44,
    paddingHorizontal: spacing.sm
  },
  emptyClearText: {
    color: palette.action,
    fontSize: 15,
    fontWeight: '600'
  },
  updatedText: {
    color: palette.textMuted,
    fontSize: 13,
    lineHeight: 18
  },
  disabledAction: {
    opacity: 0.55
  }
  });
}
