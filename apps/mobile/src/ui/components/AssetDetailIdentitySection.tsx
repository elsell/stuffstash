import { Pressable, StyleSheet, Text, View } from 'react-native';
import { Camera, CheckCircle2, MoveRight, Pencil } from 'lucide-react-native';
import type {
  AssetDetailViewModel,
  AssetParentLocationCrumbViewModel,
  AssetTagViewModel
} from '../../application/assets/AssetViewModels';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { AssetBreadcrumbTrail } from './AssetCard';
import { AssetTagChips } from './AssetTagChips';
import {
  assetDetailAvailabilityAction,
  assetDetailExceptionMetadataRows,
  assetDetailIdentity,
  assetDetailMaintenanceActions,
  assetDetailPlacement,
  visibleAssetDescription
} from './AssetDetailPresentation';

type AssetDetailIdentitySectionProps = {
  readonly asset: AssetDetailViewModel;
  readonly isActionPending: boolean;
  readonly onCheckout?: () => void;
  readonly onEdit?: () => void;
  readonly onMove?: () => void;
  readonly onParentLocationPress?: (parent: AssetParentLocationCrumbViewModel) => void;
  readonly onReturn?: () => void;
  readonly onTagPress?: (tag: AssetTagViewModel) => void;
  readonly showMaintenance?: boolean;
  readonly showAvailability?: boolean;
};

export function AssetDetailIdentitySection({
  asset,
  isActionPending,
  onCheckout,
  onEdit,
  onMove,
  onParentLocationPress,
  onReturn,
  onTagPress,
  showMaintenance = true,
  showAvailability = true
}: AssetDetailIdentitySectionProps) {
  const palette = useAppearanceAwarePalette();
  const styles = createStyles(palette);
  const identity = assetDetailIdentity(asset);
  const placement = assetDetailPlacement(asset);
  const showPlacement = asset.kind !== 'location' || placement.crumbs.length > 0;
  const description = visibleAssetDescription(asset);
  const exceptionRows = assetDetailExceptionMetadataRows(asset);

  return (
    <View style={styles.section}>
      <View style={styles.identity}>
        <Text accessibilityRole="header" style={styles.title}>{identity.title}</Text>
        <Text style={styles.classification}>{identity.classificationLabel}</Text>
      </View>

      {showPlacement ? <View style={styles.placement}>
        <Text style={styles.placementLabel}>Location</Text>
        {placement.crumbs.length > 0 && onParentLocationPress ? (
          <AssetBreadcrumbTrail
            palette={palette}
            prominence="detail"
            segments={placement.crumbs}
            onSegmentPress={onParentLocationPress}
          />
        ) : (
          <Text style={styles.placementFallback}>
            {placement.fallbackLabel ?? placement.crumbs.map((crumb) => crumb.title).join(' / ')}
          </Text>
        )}
      </View> : null}

      {showAvailability ? (
        <AssetDetailAvailabilityButton
          asset={asset}
          isActionPending={isActionPending}
          onCheckout={onCheckout}
          onReturn={onReturn}
        />
      ) : null}

      {exceptionRows.length > 0 ? (
        <View style={styles.exceptionList}>
          {exceptionRows.map((row) => (
            <View key={row.label} style={styles.exceptionRow}>
              <Text style={styles.exceptionLabel}>{row.label}</Text>
              <Text style={styles.exceptionValue}>{row.value}</Text>
            </View>
          ))}
        </View>
      ) : null}

      <AssetTagChips palette={palette} tags={asset.tags} onTagPress={onTagPress} />
      {description ? <Text style={styles.description}>{description}</Text> : null}

      {showMaintenance ? (
        <AssetDetailMaintenanceBar
          asset={asset}
          isActionPending={isActionPending}
          onEdit={onEdit}
          onMove={onMove}
        />
      ) : null}
    </View>
  );
}

export function AssetDetailMaintenanceBar({
  asset,
  includeAddPhotos = false,
  isActionPending,
  onAddPhotos,
  onEdit,
  onMove
}: {
  readonly asset: AssetDetailViewModel;
  readonly includeAddPhotos?: boolean;
  readonly isActionPending: boolean;
  readonly onAddPhotos?: () => void;
  readonly onEdit?: () => void;
  readonly onMove?: () => void;
}) {
  const palette = useAppearanceAwarePalette();
  const styles = createStyles(palette);
  const maintenanceActions = assetDetailMaintenanceActions(asset).filter(
    (action) => includeAddPhotos || action.id !== 'add_photos'
  );
  if (maintenanceActions.length === 0) {
    return null;
  }

  return (
    <View accessibilityLabel="Asset maintenance" style={styles.maintenanceBar}>
      {maintenanceActions.map((action) => {
        const handler = action.id === 'edit' ? onEdit : action.id === 'move' ? onMove : onAddPhotos;
        const disabled = isActionPending || !handler;
        const label = action.id === 'move' && asset.kind === 'location' ? 'Move place' : action.label;
        return (
          <Pressable
            accessibilityLabel={label}
            accessibilityRole="button"
            accessibilityState={{ disabled }}
            disabled={disabled}
            key={action.id}
            onPress={handler}
            style={({ pressed }) => [
              styles.maintenanceAction,
              pressed ? styles.maintenanceActionPressed : null,
              disabled ? styles.disabledAction : null
            ]}
          >
            {action.id === 'edit' ? <Pencil color={palette.action} size={18} /> : null}
            {action.id === 'move' ? <MoveRight color={palette.action} size={18} /> : null}
            {action.id === 'add_photos' ? <Camera color={palette.action} size={18} /> : null}
            <Text style={styles.maintenanceActionText}>{label}</Text>
          </Pressable>
        );
      })}
    </View>
  );
}

export function AssetDetailAvailabilityButton({
  asset,
  isActionPending,
  onCheckout,
  onReturn,
  quiet = false
}: {
  readonly asset: AssetDetailViewModel;
  readonly isActionPending: boolean;
  readonly onCheckout?: () => void;
  readonly onReturn?: () => void;
  readonly quiet?: boolean;
}) {
  const palette = useAppearanceAwarePalette();
  const styles = createStyles(palette);
  const action = assetDetailAvailabilityAction(asset);
  if (!action) {
    return null;
  }
  const handler = action.id === 'return' ? onReturn : onCheckout;
  const disabled = isActionPending || !handler;
  return (
    <Pressable
      accessibilityLabel={action.label}
      accessibilityRole="button"
      accessibilityState={{ disabled }}
      disabled={disabled}
      onPress={handler}
      style={({ pressed }) => [
        styles.availabilityAction,
        quiet ? styles.quietAvailabilityAction : null,
        pressed ? (quiet ? styles.quietActionPressed : styles.actionPressed) : null,
        disabled ? styles.disabledAction : null
      ]}
    >
      {action.id === 'return'
        ? <CheckCircle2 color={quiet ? palette.action : palette.onAction} size={20} />
        : <MoveRight color={quiet ? palette.action : palette.onAction} size={20} />}
      <Text style={[styles.availabilityActionText, quiet ? styles.quietAvailabilityActionText : null]}>
        {action.label}
      </Text>
    </Pressable>
  );
}

function createStyles(palette: MobileColorPalette) {
  return StyleSheet.create({
  section: {
    gap: spacing.md
  },
  identity: {
    gap: spacing.xs
  },
  title: {
    color: palette.text,
    fontSize: 30,
    fontWeight: '700',
    letterSpacing: 0
  },
  classification: {
    color: palette.textMuted,
    fontSize: 15,
    fontWeight: '500'
  },
  placement: {
    gap: spacing.sm
  },
  placementLabel: {
    color: palette.textMuted,
    fontSize: 14,
    fontWeight: '700'
  },
  placementFallback: {
    color: palette.text,
    fontSize: 17,
    fontWeight: '500'
  },
  availabilityAction: {
    alignItems: 'center',
    alignSelf: 'stretch',
    backgroundColor: palette.action,
    borderRadius: radius.md,
    flexDirection: 'row',
    gap: spacing.sm,
    justifyContent: 'center',
    minHeight: 50,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm
  },
  availabilityActionText: {
    color: palette.onAction,
    flexShrink: 1,
    fontSize: 17,
    fontWeight: '600',
    textAlign: 'center'
  },
  quietAvailabilityAction: {
    backgroundColor: palette.surfaceMuted
  },
  quietActionPressed: {
    opacity: 0.82
  },
  quietAvailabilityActionText: {
    color: palette.action
  },
  actionPressed: {
    backgroundColor: palette.actionPressed
  },
  exceptionList: {
    backgroundColor: palette.surfaceMuted,
    borderRadius: radius.md,
    gap: spacing.sm,
    padding: spacing.md
  },
  exceptionRow: {
    gap: 3
  },
  exceptionLabel: {
    color: palette.textMuted,
    fontSize: 13,
    fontWeight: '600'
  },
  exceptionValue: {
    color: palette.text,
    fontSize: 16,
    fontWeight: '500'
  },
  description: {
    color: palette.text,
    fontSize: 17,
    fontWeight: '400'
  },
  maintenanceBar: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.sm
  },
  maintenanceAction: {
    alignItems: 'center',
    backgroundColor: palette.elevatedSurface,
    borderColor: palette.border,
    borderRadius: radius.md,
    borderWidth: StyleSheet.hairlineWidth,
    flexBasis: 140,
    flexGrow: 1,
    flexDirection: 'row',
    gap: spacing.xs,
    justifyContent: 'center',
    minHeight: 46,
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.sm
  },
  maintenanceActionPressed: {
    opacity: 0.82
  },
  maintenanceActionText: {
    color: palette.action,
    flexShrink: 1,
    fontSize: 15,
    fontWeight: '600',
    textAlign: 'center'
  },
  disabledAction: {
    opacity: 0.55
  }
  });
}
