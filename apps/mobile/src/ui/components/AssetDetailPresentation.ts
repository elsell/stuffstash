import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';

export type AssetDetailMetadataRow = {
  readonly label: string;
  readonly value: string;
};

export type AssetDetailLocationContext = {
  readonly label: string;
  readonly value: string;
};

export type AssetDetailBadge = {
  readonly label: string;
  readonly kind: 'kind' | 'type';
};

export type AssetDetailSectionRole = 'identity' | 'status' | 'metadata' | 'contained_assets' | 'maintenance_actions';

export type AssetDetailSectionChrome = 'page' | 'status_panel' | 'metadata_context' | 'contained_workspace' | 'utility_toolbar';

export type AssetDetailSectionPresentation = {
  readonly role: AssetDetailSectionRole;
  readonly chrome: AssetDetailSectionChrome;
};

export function assetDetailSectionOrder(
  state: {
    readonly hasWorkspaceStatus: boolean;
    readonly hasPhotoStatus: boolean;
    readonly photoUploadCount: number;
    readonly canContainAssets: boolean;
  }
): readonly AssetDetailSectionPresentation[] {
  const statusSection = state.hasWorkspaceStatus || state.hasPhotoStatus || state.photoUploadCount > 0
    ? [{ role: 'status', chrome: 'status_panel' } as const]
    : [];

  return [
    { role: 'identity', chrome: 'page' },
    ...statusSection,
    { role: 'metadata', chrome: 'metadata_context' },
    ...(state.canContainAssets ? [{ role: 'contained_assets', chrome: 'contained_workspace' } as const] : []),
    { role: 'maintenance_actions', chrome: 'utility_toolbar' }
  ];
}

export const assetDetailSectionsPresentation = assetDetailSectionOrder;

export function assetDetailBadges(
  asset: Pick<AssetDetailViewModel, 'kindLabel' | 'customTypeLabel'>
): readonly AssetDetailBadge[] {
  return [
    { label: asset.kindLabel, kind: 'kind' },
    ...(asset.customTypeLabel ? [{ label: asset.customTypeLabel, kind: 'type' } as const] : [])
  ];
}

export function assetDetailMetadataRows(
  asset: Pick<AssetDetailViewModel, 'lifecycleLabel' | 'updatedAtLabel' | 'checkoutLabel' | 'checkoutActorLabel'>
): readonly AssetDetailMetadataRow[] {
  return [
    { label: 'Status', value: asset.lifecycleLabel },
    { label: 'Checkout', value: asset.checkoutActorLabel ? `${asset.checkoutLabel}. ${asset.checkoutActorLabel}` : asset.checkoutLabel },
    { label: 'Updated', value: asset.updatedAtLabel }
  ];
}

export function assetDetailLocationContext(
  asset: Pick<AssetDetailViewModel, 'locationTrailLabel'>
): AssetDetailLocationContext {
  return {
    label: 'Location',
    value: asset.locationTrailLabel
  };
}

export function visibleAssetDescription(
  asset: Pick<AssetDetailViewModel, 'description'>
): string | undefined {
  const description = asset.description.trim();
  return description.length > 0 ? description : undefined;
}
