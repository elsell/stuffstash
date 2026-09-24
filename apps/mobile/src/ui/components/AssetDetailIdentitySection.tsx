import { AssetExpirationStatus } from './AssetExpirationStatus';
import { formatAssetExpiration, expirationStatusLabel } from '../presentation/ExpirationPresentation';
import { StyleSheet, Text, View } from 'react-native';
import { NativeCommandButton } from './NativeCommandButton';
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
  readonly showEditAction?: boolean;
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
  showEditAction = true,
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
  const showPlacement = asset.kind !== 'location' || placement.crumbs.length > 0 || asset.canMove;
  const description = visibleAssetDescription(asset);
  const exceptionRows = assetDetailExceptionMetadataRows(asset);

  return (
    <View style={styles.section}>
      <View style={styles.identity}>
        <Text accessibilityRole="header" style={styles.title}>{identity.title}</Text>
        <Text style={styles.classification}>{identity.classificationLabel}</Text>
      </View>

      <AssetExpirationStatus expiration={asset.expiration} context={asset.expirationContext} />

      {showPlacement ? <View style={styles.contextRow}>
        <View style={styles.contextText}>
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
            {asset.kind === 'location' && placement.crumbs.length === 0 ? 'Top level'
              : placement.fallbackLabel ?? placement.crumbs.map((crumb) => crumb.title).join(' / ')}
          </Text>
        )}
        </View>
        {asset.canMove ? <View style={styles.contextCommand}>
          <NativeCommandButton label={asset.kind === 'location' ? 'Move place' : 'Move'}
            disabled={isActionPending || !onMove} onPress={() => { if (!isActionPending) onMove?.(); }} />
        </View> : null}
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
          showMoveAction={false}
          showEditAction={showEditAction}
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
  showMoveAction = true,
  isActionPending,
  onAddPhotos,
  showEditAction = true,
  onEdit,
  onMove
}: {
  readonly asset: AssetDetailViewModel;
  readonly includeAddPhotos?: boolean;
  readonly showMoveAction?: boolean;
  readonly isActionPending: boolean;
  readonly onAddPhotos?: () => void;
  readonly showEditAction?: boolean;
  readonly onEdit?: () => void;
  readonly onMove?: () => void;
}) {
  const palette = useAppearanceAwarePalette();
  const styles = createStyles(palette);
  const maintenanceActions = assetDetailMaintenanceActions(asset).filter(
    (action) => (showMoveAction || action.id !== 'move') && (showEditAction || action.id !== 'edit') && (includeAddPhotos || action.id !== 'add_photos')
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
          <View key={action.id} style={styles.maintenanceCommand}>
            <NativeCommandButton label={label} disabled={disabled} onPress={() => handler?.()} />
          </View>
        );
      })}
    </View>
  );
}

export function AssetDetailAvailabilityButton({
  asset,
  isActionPending,
  onCheckout,
  onReturn
}: {
  readonly asset: AssetDetailViewModel;
  readonly isActionPending: boolean;
  readonly onCheckout?: () => void;
  readonly onReturn?: () => void;
}) {
  const styles = createStyles(useAppearanceAwarePalette());
  if (asset.kind === 'location') return null;
  const action = assetDetailAvailabilityAction(asset);
  const handler = action?.id === 'return' ? onReturn : onCheckout;
  return <View accessibilityLabel="Availability" style={styles.contextRow}>
    <View style={styles.contextText}>
      <Text style={styles.placementLabel}>Availability</Text>
      <Text style={styles.placementFallback}>{asset.checkoutLabel}</Text>
      {asset.checkoutActorLabel ? <Text style={styles.classification}>{asset.checkoutActorLabel}</Text> : null}
    </View>
    {action ? <View style={styles.contextCommand}>
      <NativeCommandButton label={action.label} disabled={isActionPending || !handler}
        onPress={() => { if (!isActionPending) handler?.(); }} />
    </View> : null}
  </View>;
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
  maintenanceCommand: { width: 160, maxWidth: '100%' },
  contextCommand: { width: 120, maxWidth: '100%' },
  contextRow: { flexDirection: 'row', flexWrap: 'wrap', alignItems: 'center', gap: spacing.md, maxWidth: 560 },
  contextText: { gap: spacing.xs, flexGrow: 1, flexShrink: 1, flexBasis: 160 }

  });
}
