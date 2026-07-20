import { useState, type ReactElement } from 'react';
import type { RefreshControlProps } from 'react-native';
import {
  FlatList,
  Pressable,
  StyleSheet,
  Text,
  View
} from 'react-native';
import type {
  AssetDetailViewModel,
  AssetParentLocationCrumbViewModel,
  AssetTagViewModel
} from '../../application/assets/AssetViewModels';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { AssetDetailPhotoGallery } from './AssetDetailPhotoGallery';
import {
  AssetDetailIdentitySection
} from './AssetDetailIdentitySection';
import {
  ContainedContentsSearch,
  ContainedSpatialActions,
  ContainedWorkspaceListItemView,
  ContainedWorkspaceMaintenance,
  containedWorkspaceItems,
  shouldShowContainedContentsSearch
} from './AssetContainedWorkspace';
import { assetDetailUpdatedMetadata } from './AssetDetailPresentation';
import { appKeyboardDismissMode } from './AppTextInput';

export {
  containedAssetRowAccessibilityLabel,
  containedWorkspaceItems
} from './AssetContainedWorkspace';

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
  readonly overflowMenu?: ReactElement;
  readonly onChildPress?: (assetId: string) => void;
  readonly onParentLocationPress?: (parent: AssetParentLocationCrumbViewModel) => void;
  readonly onAddHere?: () => void;
  readonly onMoveThingsHere?: () => void;
  readonly refreshControl?: ReactElement<RefreshControlProps>;
};

export function assetDetailNavigationTitle(asset: Pick<AssetDetailViewModel, 'kind'>): 'Place' | 'Details' {
  return asset.kind === 'location' ? 'Place' : 'Details';
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
  overflowMenu,
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
  const showContentsSearch = shouldShowContainedContentsSearch(asset);
  const workspaceItems = asset.canContainAssets
    ? containedWorkspaceItems(asset, showContentsSearch ? contentsQuery : '')
    : [];
  const updatedMetadata = assetDetailUpdatedMetadata(asset);

  return (
    <FlatList
      contentContainerStyle={styles.content}
      data={workspaceItems}
      keyExtractor={(item) => item.key}
      keyboardDismissMode={appKeyboardDismissMode()}
      keyboardShouldPersistTaps="handled"
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
          {onBack || overflowMenu ? (
            <View style={styles.topActions}>
              {onBack ? (
                <Pressable accessibilityRole="button" onPress={onBack} style={styles.backButton}>
                  <Text style={styles.backButtonText}>Back</Text>
                </Pressable>
              ) : <View />}
              {overflowMenu ?? null}
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
            <ContainedContentsSearch onChangeQuery={setContentsQuery} query={contentsQuery} />
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
    fontWeight: '500'
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
    fontSize: 14
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
  footerStack: {
    gap: spacing.lg,
    paddingTop: spacing.lg
  },
  updatedText: {
    color: palette.textMuted,
    fontSize: 13
  }
  });
}
