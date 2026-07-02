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
import { Camera, ChevronRight, MoreHorizontal, MoveRight, Pencil } from 'lucide-react-native';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import {
  assetPhotoStatusLabel
} from './AssetPhotoWorkspacePresentation';
import {
  canUseContainedAssetAction,
  containedAssetActions,
  containedAssetRows,
  containedAssetsEmptyState,
  type ContainedAssetAction,
  type ContainedAssetRowViewModel
} from './ContainedAssetsPresentation';
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
  readonly photoStatusMessage?: string;
  readonly workspaceStatusMessage?: string;
  readonly workspaceStatusKind?: 'success' | 'working';
  readonly canRetryPhotos?: boolean;
  readonly onBack?: () => void;
  readonly onEdit?: () => void;
  readonly onMove?: () => void;
  readonly onAddPhotos?: () => void;
  readonly onPhotoPress?: (photoId: string) => void;
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
  onPhotoPress,
  onMoveThingsHere,
  onRetryPhotos,
  photoUploads = [],
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
          onPhotoPress={onPhotoPress}
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
          ) : null}

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
  onPhotoPress
}: {
  readonly asset: AssetDetailViewModel;
  readonly disabled: boolean;
  readonly onAddPhotos?: () => void;
  readonly onPhotoPress?: (photoId: string) => void;
}) {
  const photos = asset.photos.length > 0 ? asset.photos : undefined;

  return (
    <View style={styles.photoWorkspace}>
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
              {assetPhotoStatusLabel({ index, label: photo.label })}
            </Text>
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
  const childRows = containedAssetRows(asset.containedAssets);

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
      {childRows.length > 0 ? (
        <View style={styles.childList}>
          {childRows.map((child) => (
            <ContainedAssetRowView
              key={child.id}
              asset={child}
              onPress={onChildPress ? () => onChildPress(child.id) : undefined}
            />
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

function ContainedAssetRowView({
  asset,
  onPress
}: {
  readonly asset: ContainedAssetRowViewModel;
  readonly onPress?: () => void;
}) {
  const isDisabled = onPress === undefined;
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityState={{ disabled: isDisabled }}
      disabled={isDisabled}
      onPress={onPress}
      style={[styles.childRow, isDisabled ? styles.disabledAction : null]}
    >
      <View style={styles.childPhoto}>
        {asset.photo ? (
          <Image
            accessibilityIgnoresInvertColors
            source={{ uri: asset.photo.uri, headers: asset.photo.headers }}
            style={styles.childPhotoImage}
          />
        ) : (
          <Text style={styles.childPhotoPlaceholder}>{asset.imagePlaceholderLabel}</Text>
        )}
      </View>
      <View style={styles.childRowText}>
        <Text style={styles.childEyebrow}>{asset.eyebrowLabel}</Text>
        <Text numberOfLines={2} style={styles.childTitle}>{asset.title}</Text>
        <Text numberOfLines={2} style={styles.childSupporting}>{asset.supportingLabel}</Text>
      </View>
      <ChevronRight color={colors.textMuted} size={20} />
    </Pressable>
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
  childList: {
    gap: spacing.sm
  },
  childRow: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    minHeight: 86,
    padding: spacing.sm
  },
  childPhoto: {
    alignItems: 'center',
    aspectRatio: 1,
    backgroundColor: colors.surfaceMuted,
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
    color: colors.accentStrong,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  childRowText: {
    flex: 1,
    gap: 2
  },
  childEyebrow: {
    color: colors.textMuted,
    fontSize: 11,
    fontWeight: '900',
    letterSpacing: 0,
    textTransform: 'uppercase'
  },
  childTitle: {
    color: colors.text,
    fontSize: 17,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 22
  },
  childSupporting: {
    color: colors.textMuted,
    fontSize: 13,
    fontWeight: '700',
    letterSpacing: 0,
    lineHeight: 18
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
