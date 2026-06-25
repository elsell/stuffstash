import {
  Pressable,
  Image,
  ScrollView,
  StyleSheet,
  Text,
  View
} from 'react-native';
import type { ReactElement } from 'react';
import type { RefreshControlProps } from 'react-native';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import { colors, radius, spacing } from '../theme/tokens';

type AssetDetailViewProps = {
  readonly asset: AssetDetailViewModel;
  readonly isLifecycleActionPending?: boolean;
  readonly onBack?: () => void;
  readonly onArchive?: () => void;
  readonly onRestore?: () => void;
  readonly onDeletePermanently?: () => void;
  readonly refreshControl?: ReactElement<RefreshControlProps>;
};

export function AssetDetailView({
  asset,
  isLifecycleActionPending = false,
  onArchive,
  onBack,
  onDeletePermanently,
  onRestore,
  refreshControl
}: AssetDetailViewProps) {
  return (
    <ScrollView contentContainerStyle={styles.content} refreshControl={refreshControl}>
      {onBack ? (
        <Pressable accessibilityRole="button" onPress={onBack} style={styles.backButton}>
          <Text style={styles.backButtonText}>Back</Text>
        </Pressable>
      ) : null}

      <AssetDetailPanel
        asset={asset}
        isLifecycleActionPending={isLifecycleActionPending}
        onArchive={onArchive}
        onDeletePermanently={onDeletePermanently}
        onRestore={onRestore}
      />
    </ScrollView>
  );
}

export function AssetDetailPanel({
  asset,
  isLifecycleActionPending = false,
  onArchive,
  onDeletePermanently,
  onRestore
}: {
  readonly asset: AssetDetailViewModel;
  readonly isLifecycleActionPending?: boolean;
  readonly onArchive?: () => void;
  readonly onRestore?: () => void;
  readonly onDeletePermanently?: () => void;
}) {
  const showLifecycleActions =
    asset.canArchive || asset.canRestore || asset.canDeletePermanently;

  return (
    <View style={styles.stack}>
      <View style={styles.photoHero}>
        {asset.photo ? (
          <Image
            accessibilityIgnoresInvertColors
            source={{ uri: asset.photo.uri, headers: asset.photo.headers }}
            style={styles.heroImage}
          />
        ) : (
          <Text style={styles.photoPlaceholder}>{asset.imagePlaceholderLabel}</Text>
        )}
        <Text style={styles.photoStatus}>{asset.photoLabel}</Text>
      </View>

      <View style={styles.panel}>
        <View style={styles.badgeRow}>
          <Text style={styles.kindBadge}>{asset.kindLabel}</Text>
          {asset.customTypeLabel ? <Text style={styles.typeBadge}>{asset.customTypeLabel}</Text> : null}
        </View>
        <Text style={styles.title}>{asset.title}</Text>
        <Text style={styles.description}>{asset.description}</Text>

        <View style={styles.metadataList}>
          <MetadataRow label="Location" value={asset.locationTrailLabel} />
          <MetadataRow label="Status" value={asset.lifecycleLabel} />
          <MetadataRow label="Updated" value={asset.updatedAtLabel} />
        </View>

        {showLifecycleActions ? (
          <View style={styles.lifecycleActions}>
            {asset.canArchive ? (
              <Pressable
                accessibilityRole="button"
                accessibilityState={{ disabled: isLifecycleActionPending || !onArchive }}
                disabled={isLifecycleActionPending || !onArchive}
                onPress={onArchive}
                style={[
                  styles.lifecycleAction,
                  styles.archiveAction,
                  isLifecycleActionPending || !onArchive ? styles.disabledAction : null
                ]}
              >
                <Text style={styles.archiveActionText}>Archive</Text>
              </Pressable>
            ) : null}
            {asset.canRestore ? (
              <Pressable
                accessibilityRole="button"
                accessibilityState={{ disabled: isLifecycleActionPending || !onRestore }}
                disabled={isLifecycleActionPending || !onRestore}
                onPress={onRestore}
                style={[
                  styles.lifecycleAction,
                  styles.restoreAction,
                  isLifecycleActionPending || !onRestore ? styles.disabledAction : null
                ]}
              >
                <Text style={styles.restoreActionText}>Restore</Text>
              </Pressable>
            ) : null}
            {asset.canDeletePermanently ? (
              <Pressable
                accessibilityRole="button"
                accessibilityState={{
                  disabled: isLifecycleActionPending || !onDeletePermanently
                }}
                disabled={isLifecycleActionPending || !onDeletePermanently}
                onPress={onDeletePermanently}
                style={[
                  styles.lifecycleAction,
                  styles.deleteAction,
                  isLifecycleActionPending || !onDeletePermanently
                    ? styles.disabledAction
                    : null
                ]}
              >
                <Text style={styles.deleteActionText}>Delete permanently</Text>
              </Pressable>
            ) : null}
          </View>
        ) : null}

        <View style={styles.actionRow}>
          <Pressable
            accessibilityRole="button"
            accessibilityState={{ disabled: true }}
            disabled
            style={[styles.primaryAction, styles.disabledAction]}
          >
            <Text style={styles.primaryActionText}>Edit</Text>
          </Pressable>
          <Pressable
            accessibilityRole="button"
            accessibilityState={{ disabled: true }}
            disabled
            style={[styles.secondaryAction, styles.disabledAction]}
          >
            <Text style={styles.secondaryActionText}>Move</Text>
          </Pressable>
        </View>
      </View>
    </View>
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
    paddingBottom: spacing.xl
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
  photoHero: {
    alignItems: 'center',
    aspectRatio: 4 / 3,
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.md,
    justifyContent: 'center',
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
    marginTop: spacing.xs,
    overflow: 'hidden',
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs,
    position: 'absolute',
    right: spacing.sm,
    top: spacing.sm
  },
  heroImage: {
    height: '100%',
    width: '100%'
  },
  panel: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    padding: spacing.md
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
  actionRow: {
    borderTopColor: colors.border,
    borderTopWidth: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    marginTop: spacing.md,
    paddingTop: spacing.md
  },
  lifecycleActions: {
    borderTopColor: colors.border,
    borderTopWidth: 1,
    gap: spacing.sm,
    marginTop: spacing.md,
    paddingTop: spacing.md
  },
  lifecycleAction: {
    alignItems: 'center',
    borderRadius: radius.md,
    justifyContent: 'center',
    minHeight: 44,
    paddingHorizontal: spacing.md
  },
  archiveAction: {
    backgroundColor: colors.warningSurface,
    borderColor: colors.warning,
    borderWidth: 1
  },
  archiveActionText: {
    color: colors.warning,
    fontSize: 15,
    fontWeight: '800',
    letterSpacing: 0
  },
  restoreAction: {
    backgroundColor: colors.action
  },
  restoreActionText: {
    color: colors.onAction,
    fontSize: 15,
    fontWeight: '800',
    letterSpacing: 0
  },
  deleteAction: {
    borderColor: colors.danger,
    borderWidth: 1
  },
  deleteActionText: {
    color: colors.danger,
    fontSize: 15,
    fontWeight: '800',
    letterSpacing: 0
  },
  primaryAction: {
    alignItems: 'center',
    backgroundColor: colors.action,
    borderRadius: radius.md,
    justifyContent: 'center',
    minHeight: 44,
    paddingHorizontal: spacing.md
  },
  primaryActionText: {
    color: colors.onAction,
    fontSize: 15,
    fontWeight: '800',
    letterSpacing: 0
  },
  secondaryAction: {
    alignItems: 'center',
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    justifyContent: 'center',
    minHeight: 44,
    paddingHorizontal: spacing.md
  },
  secondaryActionText: {
    color: colors.text,
    fontSize: 15,
    fontWeight: '800',
    letterSpacing: 0
  },
  disabledAction: {
    opacity: 0.55
  }
});
