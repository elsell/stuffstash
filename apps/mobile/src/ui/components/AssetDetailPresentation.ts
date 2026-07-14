import type {
  AssetDetailViewModel,
  AssetParentLocationCrumbViewModel
} from '../../application/assets/AssetViewModels';

export type AssetDetailMetadataRow = {
  readonly label: string;
  readonly value: string;
};

export type AssetDetailLocationContext = {
  readonly label: string;
  readonly value: string;
};

export type AssetDetailIdentityPresentation = {
  readonly title: string;
  readonly classificationLabel: string;
};

export type AssetDetailPlacementPresentation = {
  readonly accessibilityLabel: string;
  readonly crumbs: readonly AssetParentLocationCrumbViewModel[];
  readonly fallbackLabel?: 'No location';
};

export type AssetDetailAvailabilityAction = {
  readonly id: 'check_out' | 'return';
  readonly label: 'Check out' | 'Return';
};

export type AssetDetailMaintenanceAction = {
  readonly id: 'edit' | 'move' | 'add_photos';
  readonly label: 'Edit' | 'Move' | 'Add photos';
};

export type AssetDetailUpdatedMetadata = {
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

export function assetDetailIdentity(
  asset: Pick<AssetDetailViewModel, 'title' | 'kindLabel' | 'customTypeLabel'>
): AssetDetailIdentityPresentation {
  return {
    title: asset.title,
    classificationLabel: [asset.kindLabel, asset.customTypeLabel].filter(isPresent).join(' · ')
  };
}

export function assetDetailPlacement(
  asset: { readonly parentLocationTrail: readonly AssetParentLocationCrumbViewModel[] }
): AssetDetailPlacementPresentation {
  const crumbs = asset.parentLocationTrail;
  if (crumbs.length === 0) {
    return {
      accessibilityLabel: 'Location No location',
      crumbs,
      fallbackLabel: 'No location'
    };
  }

  return {
    accessibilityLabel: `Location ${crumbs.map((crumb) => crumb.title).join(', ')}`,
    crumbs
  };
}

export function assetDetailAvailabilityAction(
  asset: Pick<AssetDetailViewModel, 'canCheckout' | 'canReturn'>
): AssetDetailAvailabilityAction | undefined {
  if (asset.canReturn) {
    return { id: 'return', label: 'Return' };
  }
  if (asset.canCheckout) {
    return { id: 'check_out', label: 'Check out' };
  }
  return undefined;
}

export function assetDetailMaintenanceActions(
  asset: Pick<AssetDetailViewModel, 'canEdit' | 'canMove' | 'canAddPhotos'>
): readonly AssetDetailMaintenanceAction[] {
  return [
    ...(asset.canEdit ? [{ id: 'edit', label: 'Edit' } as const] : []),
    ...(asset.canMove ? [{ id: 'move', label: 'Move' } as const] : []),
    ...(asset.canAddPhotos ? [{ id: 'add_photos', label: 'Add photos' } as const] : [])
  ];
}

export function assetDetailExceptionMetadataRows(
  asset: Pick<
    AssetDetailViewModel,
    'lifecycleLabel' | 'isActive' | 'isCheckedOut' | 'checkoutLabel' | 'checkoutActorLabel'
  >
): readonly AssetDetailMetadataRow[] {
  return [
    ...(!asset.isActive ? [{ label: 'Lifecycle', value: asset.lifecycleLabel }] : []),
    ...(asset.isCheckedOut ? [{
      label: 'Availability',
      value: [asset.checkoutLabel, asset.checkoutActorLabel].filter(isPresent).join(' · ')
    }] : [])
  ];
}

export function assetDetailMetadataRows(
  asset: Pick<
    AssetDetailViewModel,
    | 'lifecycleLabel'
    | 'isActive'
    | 'isCheckedOut'
    | 'checkoutLabel'
    | 'checkoutActorLabel'
    | 'updatedAtLabel'
  >
): readonly AssetDetailMetadataRow[] {
  return assetDetailExceptionMetadataRows(asset);
}

export function assetDetailLocationContext(
  asset:
    | Pick<AssetDetailViewModel, 'locationTrailLabel'>
    | { readonly parentLocationTrail: readonly AssetParentLocationCrumbViewModel[] }
): AssetDetailLocationContext {
  if ('parentLocationTrail' in asset) {
    const placement = assetDetailPlacement(asset);
    return {
      label: 'Location',
      value: placement.fallbackLabel ?? placement.crumbs.map((crumb) => crumb.title).join(' / ')
    };
  }

  return {
    label: 'Location',
    value: asset.locationTrailLabel
  };
}

export function assetDetailUpdatedMetadata(
  asset: Pick<AssetDetailViewModel, 'updatedAtLabel'>
): AssetDetailUpdatedMetadata {
  return { value: asset.updatedAtLabel };
}

export function visibleAssetDescription(
  asset: Pick<AssetDetailViewModel, 'description'>
): string | undefined {
  const description = asset.description.trim();
  return description.length > 0 ? description : undefined;
}

function isPresent(value: string | undefined): value is string {
  return value !== undefined && value.length > 0;
}
