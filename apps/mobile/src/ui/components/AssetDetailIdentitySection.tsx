import { NativeCommandButton } from './NativeCommandButton';
import { AssetExpirationStatus } from './AssetExpirationStatus';
import { StyleSheet, Text, View } from 'react-native';
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
  assetDetailExceptionMetadataRows,
  assetDetailIdentity,
  assetDetailPlacement,
  visibleAssetDescription
} from './AssetDetailPresentation';

type AssetDetailIdentitySectionProps = {
  readonly asset: AssetDetailViewModel;
  readonly onReturn?: () => void;
  readonly isActionPending?: boolean;
  readonly onParentLocationPress?: (parent: AssetParentLocationCrumbViewModel) => void;
  readonly onTagPress?: (tag: AssetTagViewModel) => void;
};

export function AssetDetailIdentitySection({ asset, onParentLocationPress, onTagPress, onReturn, isActionPending }: AssetDetailIdentitySectionProps) {
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
      </View> : null}

      <View style={styles.contextRow}><AssetDetailAvailabilityStatus asset={asset} />
        {asset.kind !== 'location' && asset.canReturn && onReturn ? <NativeCommandButton label="Return" disabled={isActionPending} onPress={onReturn} /> : null}
      </View>

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

    </View>
  );
}

function AssetDetailAvailabilityStatus({ asset }: { readonly asset: AssetDetailViewModel }) {
  const styles = createStyles(useAppearanceAwarePalette());
  if (asset.kind === 'location') return null;
  return <View accessibilityLabel="Availability" style={styles.contextText}>
    <Text style={styles.placementLabel}>Availability</Text>
    <Text style={styles.placementFallback}>{asset.checkoutLabel}</Text>
    {asset.checkoutActorLabel ? <Text style={styles.classification}>{asset.checkoutActorLabel}</Text> : null}
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
  contextRow: { gap: spacing.xs },
  contextText: { gap: spacing.xs }

  });
}
