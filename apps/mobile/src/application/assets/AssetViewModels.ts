import { t } from '../../presentation/localization';
import type { AssetSummary } from '../../domain/assets/AssetSummary';

export type AssetCardViewModel = {
  /** Undefined means photo presence has not been established. */
  readonly hasPhoto?: boolean;
  readonly expiration?: AssetSummary['expiration'];
  readonly expirationContext?: AssetSummary['expirationContext'];
  readonly id: string;
  readonly title: string;
  readonly kindLabel: string;
  readonly customTypeLabel?: string;
  readonly description: string;
  readonly locationTrailLabel: string;
  readonly parentLocationTrail: readonly AssetParentLocationCrumbViewModel[];
  readonly updatedAtLabel: string;
  readonly photoLabel: string;
  readonly checkedOutLabel?: string;
  readonly tags?: readonly AssetTagViewModel[];
  readonly searchMatchLabels?: readonly string[];
  readonly imagePlaceholderLabel: string;
  readonly photo?: {
    readonly variant?: 'small' | 'medium' | 'large' | 'original';
    readonly uri: string;
    readonly headers?: Readonly<Record<string, string>>;
  };
};

export type AssetParentLocationCrumbViewModel = {
  readonly id: string;
  readonly title: string;
  readonly isImmediateParent: boolean;
};

export type AssetRelativePathCrumbViewModel = {
  readonly id: string;
  readonly title: string;
};

export type AssetContainedItemViewModel = AssetCardViewModel & {
  readonly relativePath: readonly AssetRelativePathCrumbViewModel[];
  readonly relativePathLabel: string | undefined;
};

export type AssetTagViewModel = {
  readonly id: string;
  readonly label: string;
  readonly color?: string;
};

export type AssetPhotoViewModel = {
  readonly id?: string;
  readonly fileName?: string;
  readonly contentType?: string;
  readonly sizeBytes?: number;
  readonly label: string;
  readonly uri: string;
  readonly variant?: 'small' | 'medium' | 'large' | 'original';
  readonly heroVariant?: 'small' | 'medium' | 'large' | 'original';
  readonly viewerVariant?: 'small' | 'medium' | 'large' | 'original';
  readonly heroUri?: string;
  readonly heroHeaders?: Readonly<Record<string, string>>;
  readonly viewerUri?: string;
  readonly viewerHeaders?: Readonly<Record<string, string>>;
  readonly headers?: Readonly<Record<string, string>>;
};

export type AssetDetailViewModel = {
  readonly expirationContext?: AssetSummary["expirationContext"];
  readonly expiration?: AssetSummary['expiration'];
  readonly customAssetTypeId?: string;
  readonly tenantId?: string;
  readonly inventoryId?: string;
  readonly id: string;
  readonly title: string;
  readonly kind: AssetSummary['kind'];
  readonly kindLabel: string;
  readonly customTypeLabel?: string;
  readonly description: string;
  readonly parentAssetId?: string;
  readonly locationTrailLabel: string;
  readonly parentLocationTrailLabel: string;
  readonly parentLocationTrail: readonly AssetParentLocationCrumbViewModel[];
  readonly isPlacementLoading?: boolean;
  readonly lifecycleLabel: string;
  readonly isActive: boolean;
  readonly canEdit: boolean;
  readonly canMove: boolean;
  readonly canAddPhotos: boolean;
  readonly canArchive: boolean;
  readonly canRestore: boolean;
  readonly canDeletePermanently: boolean;
  readonly isCheckedOut: boolean;
  readonly checkoutLabel: string;
  readonly checkoutActorLabel?: string;
  readonly tags?: readonly AssetTagViewModel[];
  readonly canCheckout: boolean;
  readonly canReturn: boolean;
  readonly containedAssets: readonly AssetCardViewModel[];
  readonly containedAssetsLabel: string;
  readonly containedSpaces: readonly AssetCardViewModel[];
  readonly containedSpacesLabel: string;
  readonly containedItems: readonly AssetContainedItemViewModel[];
  readonly containedItemsLabel: string;
  readonly canContainAssets: boolean;
  readonly canAddContainedAssets: boolean;
  readonly updatedAtLabel: string;
  readonly photoLabel: string;
  readonly imagePlaceholderLabel: string;
  readonly photos: readonly AssetPhotoViewModel[];
  readonly photo?: {
    readonly variant?: 'small' | 'medium' | 'large' | 'original';
    readonly uri: string;
    readonly headers?: Readonly<Record<string, string>>;
  };
};

export function toAssetCardViewModel(asset: AssetSummary): AssetCardViewModel {
  const tags = (asset.tags ?? []).map((tag) => ({
    id: tag.id,
    label: tag.displayName,
    color: tag.color
  }));
  return {
    id: asset.id,
    title: asset.title,
    ...(asset.expiration ? { expiration: asset.expiration } : {}),
    ...(asset.expirationContext ? { expirationContext: asset.expirationContext } : {}),
    kindLabel: labelAssetKind(asset.kind),
    customTypeLabel: asset.customType,
    description: asset.description,
    locationTrailLabel: labelLocationTrail(asset.locationTrail),
    parentLocationTrail: parentLocationTrail(asset),
    updatedAtLabel: asset.updatedAtLabel,
    photoLabel: asset.hasPhoto ? t('mobile.AssetViewModels.photoReady') : t('mobile.AssetViewModels.needsPhoto'),
    hasPhoto: asset.hasPhoto,
    ...(asset.currentCheckout ? { checkedOutLabel: t('mobile.AssetViewModels.checkedOut') } : {}),
    ...(tags.length > 0 ? { tags } : {}),
    imagePlaceholderLabel: placeholderForKind(asset.kind),
    ...(asset.photo ? { photo: asset.photo } : {})
  };
}

export function toAssetDetailViewModel(
  asset: AssetSummary,
  options: {
    readonly tenantId?: string;
    readonly inventoryId?: string;
    readonly canManageLifecycle?: boolean;
    readonly canEditAsset?: boolean;
    readonly canCreateAsset?: boolean;
    readonly allAssets?: readonly AssetSummary[];
    readonly isPlacementLoading?: boolean;
  } = {}
): AssetDetailViewModel {
  const canManageLifecycle = options.canManageLifecycle ?? true;
  const canEditAsset = options.canEditAsset ?? canManageLifecycle;
  const canCreateAsset = options.canCreateAsset ?? canEditAsset;
  const containedAssets = (options.allAssets ?? [])
    .filter((candidate) => candidate.parentAssetId === asset.id)
    .slice()
    .sort(compareContainedAssetSummaries)
    .map(toAssetCardViewModel);
  const locationContents = asset.kind === 'location'
    ? locationWorkspaceContents(asset, options.allAssets ?? [])
    : { spaces: [], items: [] };

  return {
    ...toAssetCardViewModel(asset),
    expiration: asset.expiration,
    expirationContext: asset.expirationContext,
    customAssetTypeId: asset.customAssetTypeId,
    tenantId: options.tenantId ?? '',
    inventoryId: options.inventoryId ?? '',
    kind: asset.kind,
    parentAssetId: asset.parentAssetId,
    parentLocationTrailLabel: labelParentLocationTrail(asset),
    isPlacementLoading: options.isPlacementLoading,
    photos: toAssetPhotoViewModels(asset.photos ?? (asset.photo ? [asset.photo] : [])),
    lifecycleLabel: asset.lifecycleState === 'active' ? t('mobile.AssetViewModels.active') : t('mobile.AssetViewModels.archived'),
    isActive: asset.lifecycleState === 'active',
    canEdit: canEditAsset && asset.lifecycleState === 'active',
    canMove: canEditAsset && asset.lifecycleState === 'active',
    canAddPhotos: canEditAsset && asset.lifecycleState === 'active',
    canArchive: canManageLifecycle && asset.lifecycleState === 'active',
    canRestore: canManageLifecycle && asset.lifecycleState === 'archived',
    canDeletePermanently: canManageLifecycle && asset.lifecycleState === 'archived',
    isCheckedOut: asset.currentCheckout !== undefined,
    checkoutLabel: checkoutLabel(asset),
    // Principal IDs are authorization identifiers, not safe user-facing names.
    // Populate an actor label only when the API exposes a resolved safe profile.
    canCheckout: asset.kind !== 'location'
      && canEditAsset
      && asset.lifecycleState === 'active'
      && asset.currentCheckout === undefined,
    canReturn: canEditAsset && asset.currentCheckout !== undefined,
    containedAssets,
    containedAssetsLabel: t('asset.thingsInside', { count: containedAssets.length }),
    containedSpaces: locationContents.spaces,
    containedSpacesLabel: t('spaces.count', { count: locationContents.spaces.length }),
    containedItems: locationContents.items,
    containedItemsLabel: t('items.count', { count: locationContents.items.length }),
    canContainAssets: asset.kind === 'container' || asset.kind === 'location',
    canAddContainedAssets: canCreateAsset && canEditAsset && asset.lifecycleState === 'active' && (asset.kind === 'container' || asset.kind === 'location')
  };
}

export function toAssetPhotoViewModels(
  photos: AssetSummary['photos'] = []
): readonly AssetPhotoViewModel[] {
  return (photos ?? []).map((photo, index) => ({
    id: photo.id,
    fileName: photo.fileName,
    contentType: photo.contentType,
    sizeBytes: photo.sizeBytes,
    label: photo.fileName ?? t('asset.photoPosition', { position: index + 1 }),
    uri: photo.uri,
    variant: photo.variant,
    heroVariant: photo.heroVariant,
    viewerVariant: photo.viewerVariant,
    heroUri: photo.heroUri,
    heroHeaders: photo.heroHeaders,
    viewerUri: photo.viewerUri,
    viewerHeaders: photo.viewerHeaders,
    headers: photo.headers
  }));
}

function locationWorkspaceContents(
  location: AssetSummary,
  allAssets: readonly AssetSummary[]
): {
  readonly spaces: readonly AssetCardViewModel[];
  readonly items: readonly AssetContainedItemViewModel[];
} {
  const childrenByParent = new Map<AssetSummary['id'], AssetSummary[]>();
  const indexedAssetIds = new Set<AssetSummary['id']>();
  for (const candidate of allAssets) {
    if (indexedAssetIds.has(candidate.id)) {
      continue;
    }
    indexedAssetIds.add(candidate.id);
    const parentAssetId = candidate.parentAssetId;
    if (!parentAssetId) {
      continue;
    }
    const siblings = childrenByParent.get(parentAssetId) ?? [];
    siblings.push(candidate);
    childrenByParent.set(parentAssetId, siblings);
  }
  for (const children of childrenByParent.values()) {
    children.sort(compareContainedAssetSummaries);
  }

  const spaces: AssetCardViewModel[] = [];
  const items: AssetContainedItemViewModel[] = [];
  const visited = new Set<AssetSummary['id']>([location.id]);
  const pending: Array<{
    readonly asset: AssetSummary;
    readonly parentId: AssetSummary['id'];
    readonly relativePath: readonly AssetRelativePathCrumbViewModel[];
  }> = [];
  const rootChildren = childrenByParent.get(location.id) ?? [];
  for (let index = rootChildren.length - 1; index >= 0; index -= 1) {
    const child = rootChildren[index];
    if (child) {
      pending.push({ asset: child, parentId: location.id, relativePath: [] });
    }
  }
  while (pending.length > 0) {
    const entry = pending.pop();
    if (!entry || visited.has(entry.asset.id)) {
      continue;
    }
    visited.add(entry.asset.id);
    if (entry.asset.kind === 'item') {
      items.push({
        ...toAssetCardViewModel(entry.asset),
        relativePath: entry.relativePath,
        relativePathLabel: entry.relativePath.length > 0
          ? entry.relativePath.map((crumb) => crumb.title).join(' / ')
          : undefined
      });
      continue;
    }
    if (entry.parentId === location.id) {
      spaces.push(toAssetCardViewModel(entry.asset));
    }
    const childPath = [
      ...entry.relativePath,
      { id: entry.asset.id, title: entry.asset.title }
    ];
    const children = childrenByParent.get(entry.asset.id) ?? [];
    for (let index = children.length - 1; index >= 0; index -= 1) {
      const child = children[index];
      if (child) {
        pending.push({ asset: child, parentId: entry.asset.id, relativePath: childPath });
      }
    }
  }
  items.sort(compareContainedItems);
  return { spaces, items };
}

function compareContainedItems(
  left: AssetContainedItemViewModel,
  right: AssetContainedItemViewModel
): number {
  const titleOrder = compareStableText(left.title, right.title);
  if (titleOrder !== 0) {
    return titleOrder;
  }
  const pathOrder = compareStableText(left.relativePathLabel ?? '', right.relativePathLabel ?? '');
  if (pathOrder !== 0) {
    return pathOrder;
  }
  return left.id.localeCompare(right.id);
}

function checkoutLabel(asset: AssetSummary): string {
  if (!asset.currentCheckout) {
    return t('mobile.AssetViewModels.available');
  }
  const date = new Date(asset.currentCheckout.checkedOutAt);
  if (Number.isNaN(date.getTime())) {
    return t('mobile.AssetViewModels.checkedOut');
  }
  return t('mobile.AssetViewModels.checkedOut2', { value: String(date.toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  })) });
}

function compareContainedAssetSummaries(left: AssetSummary, right: AssetSummary): number {
  const kindRank = containedKindRank(left.kind) - containedKindRank(right.kind);
  if (kindRank !== 0) {
    return kindRank;
  }

  const titleOrder = compareStableText(left.title, right.title);
  if (titleOrder !== 0) {
    return titleOrder;
  }

  return left.id.localeCompare(right.id);
}

function compareStableText(left: string, right: string): number {
  const leftKey = stableTextSortKey(left);
  const rightKey = stableTextSortKey(right);

  if (leftKey < rightKey) {
    return -1;
  }
  if (leftKey > rightKey) {
    return 1;
  }
  return 0;
}

function stableTextSortKey(value: string): string {
  return value.trim().normalize('NFKD').toLowerCase();
}

function containedKindRank(kind: AssetSummary['kind']): number {
  return kind === 'item' ? 1 : 0;
}

function labelLocationTrail(locationTrail: readonly string[]): string {
  const localTrail = locationTrail.slice(1);

  if (localTrail.length === 0) {
    return locationTrail[0] ?? t('mobile.AssetViewModels.unplaced');
  }

  return localTrail.join(' / ');
}

function labelParentLocationTrail(asset: AssetSummary): string {
  if (asset.parentLocationTrail.length === 0) {
    return t('mobile.AssetViewModels.inventoryRoot');
  }

  return asset.parentLocationTrail.map((segment) => segment.title).join(' / ');
}

function parentLocationTrail(asset: AssetSummary): readonly AssetParentLocationCrumbViewModel[] {
  return asset.parentLocationTrail.map((segment, index) => ({
    id: segment.id,
    title: segment.title,
    isImmediateParent: index === asset.parentLocationTrail.length - 1
  }));
}

function labelAssetKind(kind: AssetSummary['kind']): string {
  switch (kind) {
    case 'container':
      return t('mobile.AssetViewModels.container');
    case 'item':
      return t('mobile.AssetViewModels.item');
    case 'location':
      return t('mobile.AssetViewModels.place');
  }
}

function placeholderForKind(kind: AssetSummary['kind']): string {
  switch (kind) {
    case 'container':
      return t('mobile.AssetViewModels.box');
    case 'item':
      return t('mobile.AssetViewModels.item');
    case 'location':
      return t('mobile.AssetViewModels.place');
  }
}
