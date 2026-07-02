import {
  Image,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View
} from 'react-native';
import type { ReactElement } from 'react';
import type { RefreshControlProps } from 'react-native';
import { Camera, ChevronLeft, ChevronRight, MoreHorizontal, MoveRight, Pencil, X } from 'lucide-react-native';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import {
  assetPhotoStatusLabel,
  localAssetPhotoOrderNotice,
  orderedAssetPhotos
} from './AssetPhotoWorkspacePresentation';
import {
  canUseContainedAssetAction,
  containedAssetActions,
  containedAssetsEmptyState,
  type ContainedAssetAction
} from './ContainedAssetsPresentation';
import { AssetCard } from './AssetCard';
import { colors, radius, spacing } from '../theme/tokens';

export type AssetPhotoUploadProgressViewModel = {
  readonly index: number;
  readonly fileName: string;
  readonly status: 'pending' | 'uploading' | 'attached' | 'failed';
};

type AssetDetailViewProps = {
  readonly asset: AssetDetailViewModel;
  readonly isActionPending?: boolean;
  readonly photoUploads?: readonly AssetPhotoUploadProgressViewModel[];
  readonly photoOrder?: readonly string[];
  readonly photoStatusMessage?: string;
  readonly workspaceStatusMessage?: string;
  readonly workspaceStatusKind?: 'success' | 'working';
  readonly canRetryPhotos?: boolean;
  readonly onBack?: () => void;
  readonly onEdit?: () => void;
  readonly onMove?: () => void;
  readonly onAddPhotos?: () => void;
  readonly onMovePhoto?: (photoId: string, direction: -1 | 1) => void;
  readonly onPhotoPress?: (photoId: string) => void;
  readonly onRemovePhoto?: (photoId: string) => void;
  readonly onRetryPhotos?: () => void;
  readonly onMoreActions?: () => void;
  readonly onChildPress?: (assetId: string) => void;
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
  onChildPress,
  onEdit,
  onMoreActions,
  onMove,
  onMovePhoto,
  onPhotoPress,
  onRemovePhoto,
  onMoveThingsHere,
  onRetryPhotos,
  photoUploads = [],
  photoOrder = [],
  photoStatusMessage,
  workspaceStatusKind = 'success',
  workspaceStatusMessage,
  refreshControl
}: AssetDetailViewProps) {
  return (
    <ScrollView contentContainerStyle={styles.content} refreshControl={refreshControl}>
      {onBack ? (
        <Pressable accessibilityRole="button" onPress={onBack} style={styles.backButton}>
          <Text style={styles.backButtonText}>Back</Text>
        </Pressable>
      ) : null}

      <View style={styles.stack}>
        <PhotoWorkspace
          asset={asset}
          disabled={isActionPending || !asset.canAddPhotos || !onAddPhotos}
          onAddPhotos={onAddPhotos}
          onMovePhoto={onMovePhoto}
          onPhotoPress={onPhotoPress}
          onRemovePhoto={onRemovePhoto}
          photoOrder={photoOrder}
        />

        <View style={styles.panel}>
          <View style={styles.headerRow}>
            <View style={styles.headerText}>
              <View style={styles.badgeRow}>
                <Text style={styles.kindBadge}>{asset.kindLabel}</Text>
                {asset.customTypeLabel ? <Text style={styles.typeBadge}>{asset.customTypeLabel}</Text> : null}
              </View>
              <Text style={styles.title}>{asset.title}</Text>
            </View>
            {onMoreActions ? (
              <Pressable
                accessibilityLabel="More asset actions"
                accessibilityRole="button"
                disabled={isActionPending}
                onPress={onMoreActions}
                style={[styles.iconButton, isActionPending ? styles.disabledAction : null]}
              >
                <MoreHorizontal color={colors.text} size={22} />
              </Pressable>
            ) : null}
          </View>

          {asset.description.trim().length > 0 ? (
            <Text style={styles.description}>{asset.description}</Text>
          ) : (
            <Text style={styles.emptyDescription}>No description yet.</Text>
          )}

          <View style={styles.primaryActions}>
            <WorkspaceAction
              disabled={isActionPending || !asset.canEdit || !onEdit}
              icon={<Pencil color={colors.onAction} size={18} />}
              label="Edit"
              primary
              onPress={onEdit}
            />
            <WorkspaceAction
              disabled={isActionPending || !asset.canMove || !onMove}
              icon={<MoveRight color={colors.text} size={18} />}
              label="Move"
              onPress={onMove}
            />
            <WorkspaceAction
              disabled={isActionPending || !asset.canAddPhotos || !onAddPhotos}
              icon={<Camera color={colors.text} size={18} />}
              label="Photos"
              onPress={onAddPhotos}
            />
          </View>

          {workspaceStatusMessage ? (
            <View
              accessible
              accessibilityLiveRegion="polite"
              accessibilityRole="alert"
              style={[
                styles.workspaceStatusPanel,
                workspaceStatusKind === 'working' ? styles.workspaceStatusWorkingPanel : null
              ]}
            >
              <Text style={styles.workspaceStatusText}>{workspaceStatusMessage}</Text>
            </View>
          ) : null}

          {photoStatusMessage ? (
            <View style={styles.photoStatusPanel}>
              <Text style={styles.photoStatusText}>{photoStatusMessage}</Text>
              {canRetryPhotos && onRetryPhotos ? (
                <Pressable accessibilityRole="button" onPress={onRetryPhotos} style={styles.retryButton}>
                  <Text style={styles.retryButtonText}>Retry</Text>
                </Pressable>
              ) : null}
            </View>
          ) : null}

          {photoUploads.length > 0 ? <PhotoUploadProgressList uploads={photoUploads} /> : null}

          <View style={styles.metadataList}>
            <MetadataRow label="Location" value={asset.locationTrailLabel} />
            <MetadataRow label="Status" value={asset.lifecycleLabel} />
            <MetadataRow label="Updated" value={asset.updatedAtLabel} />
          </View>
        </View>

        {asset.canContainAssets ? (
          <ContainedAssetsSection
            asset={asset}
            isActionPending={isActionPending}
            onAddHere={onAddHere}
            onChildPress={onChildPress}
            onMoveThingsHere={onMoveThingsHere}
          />
        ) : null}
      </View>
    </ScrollView>
  );
}

function PhotoWorkspace({
  asset,
  disabled,
  onAddPhotos,
  onMovePhoto,
  onPhotoPress,
  onRemovePhoto,
  photoOrder
}: {
  readonly asset: AssetDetailViewModel;
  readonly disabled: boolean;
  readonly onAddPhotos?: () => void;
  readonly onMovePhoto?: (photoId: string, direction: -1 | 1) => void;
  readonly onPhotoPress?: (photoId: string) => void;
  readonly onRemovePhoto?: (photoId: string) => void;
  readonly photoOrder: readonly string[];
}) {
  const orderedPhotos = orderedAssetPhotos(asset.photos, photoOrder);
  const photos = orderedPhotos.length > 0 ? orderedPhotos : undefined;
  const hasLocalOrder = photoOrder.length > 0;

  return (
    <View style={styles.photoWorkspace}>
      {hasLocalOrder ? (
        <View style={styles.localOrderNotice}>
          <Text style={styles.localOrderNoticeText}>{localAssetPhotoOrderNotice}</Text>
        </View>
      ) : null}
      <ScrollView
        horizontal
        showsHorizontalScrollIndicator={false}
        contentContainerStyle={styles.photoStrip}
      >
        {photos ? photos.map((photo, index) => (
          <View key={photo.id ?? photo.uri} style={index === 0 ? styles.photoHero : styles.photoThumb}>
            <Pressable
              accessibilityLabel={`Open ${photo.label}`}
              accessibilityRole="imagebutton"
              disabled={!photo.id || !onPhotoPress}
              onPress={() => photo.id && onPhotoPress ? onPhotoPress(photo.id) : undefined}
              style={styles.photoPressable}
            >
              <Image
                accessibilityIgnoresInvertColors
                accessibilityLabel={photo.label}
                source={index === 0
                  ? { uri: photo.heroUri ?? photo.uri, headers: photo.heroHeaders ?? photo.headers }
                  : { uri: photo.uri, headers: photo.headers }}
                style={styles.heroImage}
              />
              <Text style={styles.photoPosition}>{(index + 1).toString()} / {photos.length.toString()}</Text>
            </Pressable>
            <Text style={styles.photoStatus}>
              {assetPhotoStatusLabel({ hasLocalOrder, index, label: photo.label })}
            </Text>
            {photo.id && onMovePhoto && photos.length > 1 ? (
              <View style={styles.photoReorderControls}>
                <PhotoReorderButton
                  accessibilityLabel={`Move ${photo.label} earlier`}
                  disabled={disabled || index === 0}
                  direction={-1}
                  photoId={photo.id}
                  onMovePhoto={onMovePhoto}
                />
                <PhotoReorderButton
                  accessibilityLabel={`Move ${photo.label} later`}
                  disabled={disabled || index === photos.length - 1}
                  direction={1}
                  photoId={photo.id}
                  onMovePhoto={onMovePhoto}
                />
              </View>
            ) : null}
            {photo.id && onRemovePhoto ? (
              <Pressable
                accessibilityLabel={`Remove ${photo.label}`}
                accessibilityRole="button"
                disabled={disabled}
                onPress={() => onRemovePhoto(photo.id as string)}
                style={[styles.removePhotoButton, disabled ? styles.disabledAction : null]}
              >
                <X color={colors.text} size={18} />
              </Pressable>
            ) : null}
          </View>
        )) : (
          <View style={styles.photoHero}>
            <Text style={styles.photoPlaceholder}>{asset.imagePlaceholderLabel}</Text>
            <Text style={styles.photoStatus}>{asset.photoLabel}</Text>
          </View>
        )}
        <Pressable
          accessibilityRole="button"
          accessibilityState={{ disabled }}
          disabled={disabled}
          onPress={onAddPhotos}
          style={[styles.addPhotoTile, disabled ? styles.disabledAction : null]}
        >
          <Camera color={colors.action} size={26} />
          <Text style={styles.addPhotoText}>Add photos</Text>
        </Pressable>
      </ScrollView>
    </View>
  );
}

function PhotoReorderButton({
  accessibilityLabel,
  direction,
  disabled,
  onMovePhoto,
  photoId
}: {
  readonly accessibilityLabel: string;
  readonly direction: -1 | 1;
  readonly disabled: boolean;
  readonly onMovePhoto: (photoId: string, direction: -1 | 1) => void;
  readonly photoId: string;
}) {
  return (
    <Pressable
      accessibilityLabel={accessibilityLabel}
      accessibilityRole="button"
      accessibilityState={{ disabled }}
      disabled={disabled}
      onPress={() => onMovePhoto(photoId, direction)}
      style={[styles.photoReorderButton, disabled ? styles.disabledAction : null]}
    >
      {direction < 0
        ? <ChevronLeft color={colors.text} size={18} />
        : <ChevronRight color={colors.text} size={18} />}
    </Pressable>
  );
}

function PhotoUploadProgressList({
  uploads
}: {
  readonly uploads: readonly AssetPhotoUploadProgressViewModel[];
}) {
  return (
    <View style={styles.uploadPanel}>
      <Text style={styles.uploadPanelTitle}>Photo uploads</Text>
      {uploads.map((upload) => (
        <View key={`${upload.index.toString()}-${upload.fileName}`} style={styles.uploadRow}>
          <View>
            <Text style={styles.uploadFileName} numberOfLines={1}>{upload.fileName}</Text>
            <Text style={styles.uploadStatusText}>{labelUploadStatus(upload.status)}</Text>
          </View>
          <View style={[styles.uploadStatusPill, upload.status === 'failed' ? styles.uploadFailedPill : null]}>
            <Text style={[styles.uploadStatusPillText, upload.status === 'failed' ? styles.uploadFailedPillText : null]}>
              {labelUploadPill(upload.status)}
            </Text>
          </View>
        </View>
      ))}
    </View>
  );
}

function labelUploadStatus(status: AssetPhotoUploadProgressViewModel['status']): string {
  switch (status) {
    case 'attached':
      return 'Attached to this asset';
    case 'failed':
      return 'Needs retry';
    case 'uploading':
      return 'Uploading original file';
    case 'pending':
      return 'Waiting to upload';
  }
}

function labelUploadPill(status: AssetPhotoUploadProgressViewModel['status']): string {
  switch (status) {
    case 'attached':
      return 'Done';
    case 'failed':
      return 'Failed';
    case 'uploading':
      return 'Now';
    case 'pending':
      return 'Queued';
  }
}

function WorkspaceAction({
  disabled,
  icon,
  label,
  onPress,
  primary = false
}: {
  readonly disabled: boolean;
  readonly icon: ReactElement;
  readonly label: string;
  readonly onPress?: () => void;
  readonly primary?: boolean;
}) {
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityState={{ disabled }}
      disabled={disabled}
      onPress={onPress}
      style={[
        styles.workspaceAction,
        primary ? styles.primaryAction : styles.secondaryAction,
        disabled ? styles.disabledAction : null
      ]}
    >
      {icon}
      <Text style={primary ? styles.primaryActionText : styles.secondaryActionText}>{label}</Text>
    </Pressable>
  );
}

function ContainedAssetsSection({
  asset,
  isActionPending,
  onAddHere,
  onChildPress,
  onMoveThingsHere
}: {
  readonly asset: AssetDetailViewModel;
  readonly isActionPending: boolean;
  readonly onAddHere?: () => void;
  readonly onChildPress?: (assetId: string) => void;
  readonly onMoveThingsHere?: () => void;
}) {
  const actions = containedAssetActions(asset);
  const emptyState = containedAssetsEmptyState(asset);

  return (
    <View style={styles.section}>
      <View style={styles.sectionHeader}>
        <View>
          <Text style={styles.sectionEyebrow}>Inside</Text>
          <Text style={styles.sectionTitle}>{asset.containedAssetsLabel}</Text>
        </View>
      </View>
      {actions.length > 0 ? (
        <View style={styles.containedActionBar}>
          {actions.map((action) => (
            <ContainedAssetActionButton
              key={action.kind}
              action={action}
              isActionPending={isActionPending}
              onPress={action.kind === 'add_here' ? onAddHere : onMoveThingsHere}
            />
          ))}
        </View>
      ) : null}
      {asset.containedAssets.length > 0 ? (
        <View style={styles.childGrid}>
          {asset.containedAssets.map((child) => (
            <AssetCard key={child.id} asset={child} onPress={() => onChildPress?.(child.id)} />
          ))}
        </View>
      ) : (
        <View style={styles.emptyContainer}>
          <Text style={styles.emptyContainerTitle}>{emptyState.title}</Text>
          <Text style={styles.emptyContainerText}>{emptyState.message}</Text>
        </View>
      )}
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
  const canUseAction = canUseContainedAssetAction({ isActionPending, onPress });
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityState={{ disabled: !canUseAction }}
      disabled={!canUseAction}
      onPress={onPress}
      style={[
        styles.containedAction,
        action.isPrimary ? styles.containedPrimaryAction : styles.containedSecondaryAction,
        !canUseAction ? styles.disabledAction : null
      ]}
    >
      <Text style={action.isPrimary ? styles.containedPrimaryActionText : styles.containedSecondaryActionText}>
        {action.label}
      </Text>
    </Pressable>
  );
}

function MetadataRow({ label, value }: { readonly label: string; readonly value: string }) {
  return (
    <View style={styles.metadataRow}>
      <Text style={styles.metadataLabel}>{label}</Text>
      <Text style={styles.metadataValue}>{value}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  content: {
    gap: spacing.md,
    padding: spacing.lg,
    paddingBottom: spacing.xl * 2
  },
  stack: {
    gap: spacing.md
  },
  backButton: {
    alignSelf: 'flex-start',
    minHeight: 40,
    paddingVertical: spacing.xs
  },
  backButtonText: {
    color: colors.action,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  photoWorkspace: {
    gap: spacing.sm
  },
  localOrderNotice: {
    alignSelf: 'flex-start',
    backgroundColor: colors.brandDustyBlueSoft,
    borderRadius: radius.sm,
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  },
  localOrderNoticeText: {
    color: colors.accentStrong,
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 0
  },
  photoStrip: {
    gap: spacing.sm,
    paddingRight: spacing.lg
  },
  photoHero: {
    alignItems: 'center',
    aspectRatio: 4 / 3,
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.md,
    justifyContent: 'center',
    minWidth: 300,
    overflow: 'hidden'
  },
  photoPlaceholder: {
    color: colors.accentStrong,
    fontSize: 34,
    fontWeight: '900',
    letterSpacing: 0
  },
  photoStatus: {
    backgroundColor: colors.surface,
    borderRadius: radius.sm,
    color: colors.textMuted,
    fontSize: 14,
    fontWeight: '800',
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs,
    position: 'absolute',
    right: spacing.sm,
    top: spacing.sm
  },
  photoPressable: {
    height: '100%',
    width: '100%'
  },
  photoPosition: {
    backgroundColor: colors.brandCharcoal,
    borderRadius: radius.sm,
    bottom: spacing.sm,
    color: colors.onAction,
    fontSize: 12,
    fontWeight: '900',
    left: spacing.sm,
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs,
    position: 'absolute'
  },
  heroImage: {
    height: '100%',
    width: '100%'
  },
  addPhotoTile: {
    alignItems: 'center',
    aspectRatio: 3 / 4,
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    gap: spacing.xs,
    justifyContent: 'center',
    minWidth: 132,
    padding: spacing.md
  },
  photoThumb: {
    alignItems: 'center',
    aspectRatio: 4 / 3,
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.md,
    justifyContent: 'center',
    minWidth: 180,
    overflow: 'hidden'
  },
  removePhotoButton: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: 18,
    borderWidth: 1,
    height: 36,
    justifyContent: 'center',
    position: 'absolute',
    right: spacing.sm,
    bottom: spacing.sm,
    width: 36
  },
  photoReorderControls: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: 18,
    borderWidth: 1,
    bottom: spacing.sm,
    flexDirection: 'row',
    gap: 2,
    left: '50%',
    padding: 2,
    position: 'absolute',
    transform: [{ translateX: -47 }]
  },
  photoReorderButton: {
    alignItems: 'center',
    borderRadius: 16,
    minHeight: 44,
    minWidth: 44,
    justifyContent: 'center',
    width: 44
  },
  addPhotoText: {
    color: colors.action,
    fontSize: 14,
    fontWeight: '900',
    letterSpacing: 0,
    textAlign: 'center'
  },
  panel: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    padding: spacing.md
  },
  headerRow: {
    alignItems: 'flex-start',
    flexDirection: 'row',
    gap: spacing.sm
  },
  headerText: {
    flex: 1
  },
  iconButton: {
    alignItems: 'center',
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    height: 44,
    justifyContent: 'center',
    width: 44
  },
  badgeRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.xs,
    marginBottom: spacing.sm
  },
  kindBadge: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.sm,
    color: colors.accentStrong,
    fontSize: 12,
    fontWeight: '800',
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  },
  typeBadge: {
    borderColor: colors.border,
    borderRadius: radius.sm,
    borderWidth: 1,
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '800',
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  },
  title: {
    color: colors.text,
    fontSize: 28,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 34
  },
  description: {
    color: colors.textMuted,
    fontSize: 15,
    lineHeight: 22,
    marginTop: spacing.sm
  },
  emptyDescription: {
    color: colors.textMuted,
    fontSize: 15,
    fontStyle: 'italic',
    lineHeight: 22,
    marginTop: spacing.sm
  },
  primaryActions: {
    borderTopColor: colors.border,
    borderTopWidth: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    marginTop: spacing.md,
    paddingTop: spacing.md
  },
  workspaceAction: {
    alignItems: 'center',
    borderRadius: radius.md,
    flex: 1,
    flexDirection: 'row',
    gap: spacing.xs,
    justifyContent: 'center',
    minHeight: 46,
    paddingHorizontal: spacing.sm
  },
  primaryAction: {
    backgroundColor: colors.action
  },
  primaryActionText: {
    color: colors.onAction,
    fontSize: 14,
    fontWeight: '900',
    letterSpacing: 0
  },
  secondaryAction: {
    borderColor: colors.border,
    borderWidth: 1
  },
  secondaryActionText: {
    color: colors.text,
    fontSize: 14,
    fontWeight: '900',
    letterSpacing: 0
  },
  photoStatusPanel: {
    backgroundColor: colors.brandDustyBlueSoft,
    borderRadius: radius.md,
    gap: spacing.sm,
    marginTop: spacing.md,
    padding: spacing.md
  },
  workspaceStatusPanel: {
    backgroundColor: colors.brandDustyBlueSoft,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    marginTop: spacing.md,
    padding: spacing.md
  },
  workspaceStatusWorkingPanel: {
    backgroundColor: colors.surfaceMuted
  },
  workspaceStatusText: {
    color: colors.text,
    fontSize: 14,
    fontWeight: '800',
    letterSpacing: 0,
    lineHeight: 20
  },
  photoStatusText: {
    color: colors.text,
    fontSize: 14,
    fontWeight: '700',
    letterSpacing: 0,
    lineHeight: 20
  },
  retryButton: {
    alignSelf: 'flex-start',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.sm,
    borderWidth: 1,
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  },
  retryButtonText: {
    color: colors.action,
    fontSize: 13,
    fontWeight: '900',
    letterSpacing: 0
  },
  uploadPanel: {
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    gap: spacing.sm,
    marginTop: spacing.md,
    padding: spacing.md
  },
  uploadPanelTitle: {
    color: colors.textMuted,
    fontSize: 11,
    fontWeight: '900',
    letterSpacing: 0,
    textTransform: 'uppercase'
  },
  uploadRow: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.sm,
    justifyContent: 'space-between'
  },
  uploadFileName: {
    color: colors.text,
    fontSize: 14,
    fontWeight: '800',
    letterSpacing: 0,
    maxWidth: 210
  },
  uploadStatusText: {
    color: colors.textMuted,
    fontSize: 13,
    lineHeight: 18
  },
  uploadStatusPill: {
    backgroundColor: colors.brandDustyBlueSoft,
    borderRadius: radius.sm,
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  },
  uploadStatusPillText: {
    color: colors.accentStrong,
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 0
  },
  uploadFailedPill: {
    backgroundColor: colors.warningSurface
  },
  uploadFailedPillText: {
    color: colors.warning
  },
  metadataList: {
    borderTopColor: colors.border,
    borderTopWidth: 1,
    gap: spacing.sm,
    marginTop: spacing.md,
    paddingTop: spacing.md
  },
  metadataRow: {
    gap: 2
  },
  metadataLabel: {
    color: colors.textMuted,
    fontSize: 11,
    fontWeight: '900',
    letterSpacing: 0,
    textTransform: 'uppercase'
  },
  metadataValue: {
    color: colors.text,
    fontSize: 15,
    fontWeight: '700',
    letterSpacing: 0,
    lineHeight: 21
  },
  section: {
    gap: spacing.md
  },
  sectionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between'
  },
  sectionEyebrow: {
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 0,
    textTransform: 'uppercase'
  },
  sectionTitle: {
    color: colors.text,
    fontSize: 22,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 28
  },
  childGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.sm
  },
  containedActionBar: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.sm
  },
  containedAction: {
    alignItems: 'center',
    borderRadius: radius.md,
    justifyContent: 'center',
    minHeight: 44,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm
  },
  containedPrimaryAction: {
    backgroundColor: colors.action
  },
  containedPrimaryActionText: {
    color: colors.onAction,
    fontSize: 14,
    fontWeight: '900',
    letterSpacing: 0
  },
  containedSecondaryAction: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderWidth: 1
  },
  containedSecondaryActionText: {
    color: colors.text,
    fontSize: 14,
    fontWeight: '900',
    letterSpacing: 0
  },
  emptyContainer: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    gap: spacing.sm,
    padding: spacing.md
  },
  emptyContainerTitle: {
    color: colors.text,
    fontSize: 17,
    fontWeight: '900',
    letterSpacing: 0
  },
  emptyContainerText: {
    color: colors.textMuted,
    fontSize: 14,
    lineHeight: 20
  },
  disabledAction: {
    opacity: 0.55
  }
});
