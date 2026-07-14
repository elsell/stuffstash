import type { ReactElement } from 'react';
import type { RefreshControlProps } from 'react-native';
import {
  FlatList,
  Image,
  Pressable,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { ChevronRight, MoreHorizontal, MoveRight, Plus } from 'lucide-react-native';
import type {
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
  type ContainedAssetAction,
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
  const childRows = asset.canContainAssets ? containedAssetRows(asset.containedAssets) : [];
  const updatedMetadata = assetDetailUpdatedMetadata(asset);

  return (
    <FlatList
      contentContainerStyle={styles.content}
      data={childRows}
      keyExtractor={(child) => child.id}
      refreshControl={refreshControl}
      renderItem={({ item }) => (
        <ContainedAssetRowView
          asset={item}
          onPress={onChildPress ? () => onChildPress(item.id) : undefined}
        />
      )}
      ItemSeparatorComponent={() => <View style={styles.childSeparator} />}
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
            canAddPhotos={!asset.canContainAssets && !isActionPending && asset.canAddPhotos}
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
            <ContainedWorkspaceHeader
              asset={asset}
              isActionPending={isActionPending}
              onAddHere={onAddHere}
              onAddPhotos={onAddPhotos}
              onCheckout={onCheckout}
              onEdit={onEdit}
              onMove={onMove}
              onMoveThingsHere={onMoveThingsHere}
              onReturn={onReturn}
            />
          ) : null}
        </View>
      )}
      ListEmptyComponent={asset.canContainAssets ? <ContainedAssetsEmptyState asset={asset} /> : null}
      ListFooterComponent={(
        <Text accessibilityLabel={updatedMetadata.value} style={styles.updatedText}>
          {updatedMetadata.value}
        </Text>
      )}
    />
  );
}

function ContainedWorkspaceHeader({
  asset,
  isActionPending,
  onAddHere,
  onAddPhotos,
  onCheckout,
  onEdit,
  onMove,
  onMoveThingsHere,
  onReturn
}: {
  readonly asset: AssetDetailViewModel;
  readonly isActionPending: boolean;
  readonly onAddHere?: () => void;
  readonly onAddPhotos?: () => void;
  readonly onCheckout?: () => void;
  readonly onEdit?: () => void;
  readonly onMove?: () => void;
  readonly onMoveThingsHere?: () => void;
  readonly onReturn?: () => void;
}) {
  const palette = useAppearanceAwarePalette();
  const styles = createStyles(palette);
  const heading = containedAssetsSectionHeading(asset);
  const actions = containedAssetActions(asset);
  return (
    <View style={styles.containedHeader}>
      <View style={styles.sectionHeading}>
        <Text accessibilityRole="header" style={styles.sectionTitle}>{heading.title}</Text>
        <Text style={styles.sectionSummary}>{heading.summary}</Text>
      </View>
      {actions.map((action) => (
        <ContainedAssetActionButton
          action={action}
          isActionPending={isActionPending}
          key={action.kind}
          onPress={action.kind === 'add_here' ? onAddHere : onMoveThingsHere}
        />
      ))}
      <AssetDetailAvailabilityButton
        asset={asset}
        isActionPending={isActionPending}
        onCheckout={onCheckout}
        onReturn={onReturn}
        quiet
      />
      <AssetDetailMaintenanceBar
        asset={asset}
        includeAddPhotos
        isActionPending={isActionPending}
        onAddPhotos={onAddPhotos}
        onEdit={onEdit}
        onMove={onMove}
      />
    </View>
  );
}

function ContainedAssetsEmptyState({ asset }: { readonly asset: AssetDetailViewModel }) {
  const styles = createStyles(useAppearanceAwarePalette());
  const emptyState = containedAssetsEmptyState(asset);
  return (
    <View style={styles.emptyContainer}>
      <Text style={styles.emptyContainerTitle}>{emptyState.title}</Text>
      <Text style={styles.emptyContainerText}>{emptyState.message}</Text>
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
        <Text style={styles.childEyebrow}>{asset.eyebrowLabel}</Text>
        <Text style={styles.childTitle}>{asset.title}</Text>
        {asset.supportingLabel ? <Text style={styles.childSupporting}>{asset.supportingLabel}</Text> : null}
      </View>
      <ChevronRight color={palette.textMuted} size={20} />
    </Pressable>
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
    padding: spacing.lg,
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
  containedHeader: {
    gap: spacing.md
  },
  sectionHeading: {
    gap: 3
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
    backgroundColor: palette.surfaceMuted
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
    backgroundColor: palette.surfaceMuted,
    borderRadius: radius.sm,
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
    borderTopColor: palette.border,
    borderTopWidth: 1,
    gap: spacing.xs,
    paddingVertical: spacing.lg
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
  updatedText: {
    color: palette.textMuted,
    fontSize: 13,
    lineHeight: 18,
    paddingTop: spacing.lg
  },
  disabledAction: {
    opacity: 0.55
  }
  });
}
