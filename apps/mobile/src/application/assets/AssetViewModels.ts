import type { AssetSummary } from '../../domain/assets/AssetSummary';

export type AssetCardViewModel = {
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
  readonly heroUri?: string;
  readonly heroHeaders?: Readonly<Record<string, string>>;
  readonly viewerUri?: string;
  readonly viewerHeaders?: Readonly<Record<string, string>>;
  readonly headers?: Readonly<Record<string, string>>;
};

export type AssetDetailViewModel = {
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
    kindLabel: labelAssetKind(asset.kind),
    customTypeLabel: asset.customType,
    description: asset.description,
    locationTrailLabel: labelLocationTrail(asset.locationTrail),
    parentLocationTrail: parentLocationTrail(asset),
    updatedAtLabel: asset.updatedAtLabel,
    photoLabel: asset.hasPhoto ? 'Photo ready' : 'Needs photo',
    ...(asset.currentCheckout ? { checkedOutLabel: 'Checked out' } : {}),
    ...(tags.length > 0 ? { tags } : {}),
    imagePlaceholderLabel: placeholderForKind(asset.kind),
    ...(asset.photo ? { photo: asset.photo } : {})
  };
}

export function toAssetDetailViewModel(
  asset: AssetSummary,
  options: {
    readonly canManageLifecycle?: boolean;
    readonly canEditAsset?: boolean;
    readonly canCreateAsset?: boolean;
    readonly allAssets?: readonly AssetSummary[];
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
    kind: asset.kind,
    parentAssetId: asset.parentAssetId,
    parentLocationTrailLabel: labelParentLocationTrail(asset),
    photos: (asset.photos ?? (asset.photo ? [asset.photo] : [])).map((photo, index) => ({
      id: photo.id,
      fileName: photo.fileName,
      contentType: photo.contentType,
      sizeBytes: photo.sizeBytes,
      label: photo.fileName ?? `Photo ${(index + 1).toString()}`,
      uri: photo.uri,
      heroUri: photo.heroUri,
      heroHeaders: photo.heroHeaders,
      viewerUri: photo.viewerUri,
      viewerHeaders: photo.viewerHeaders,
      headers: photo.headers
    })),
    lifecycleLabel: asset.lifecycleState === 'active' ? 'Active' : 'Archived',
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
    containedAssetsLabel: containedAssets.length === 1 ? '1 thing inside' : `${containedAssets.length.toString()} things inside`,
    containedSpaces: locationContents.spaces,
    containedSpacesLabel: countLabel(locationContents.spaces.length, 'space'),
    containedItems: locationContents.items,
    containedItemsLabel: countLabel(locationContents.items.length, 'item'),
    canContainAssets: asset.kind === 'container' || asset.kind === 'location',
    canAddContainedAssets: canCreateAsset && canEditAsset && asset.lifecycleState === 'active' && (asset.kind === 'container' || asset.kind === 'location')
  };
}

function locationWorkspaceContents(
  location: AssetSummary,
  allAssets: readonly AssetSummary[]
): {
  readonly spaces: readonly AssetCardViewModel[];
  readonly items: readonly AssetContainedItemViewModel[];
} {
  const assetsById = new Map(allAssets.map((asset) => [asset.id, asset]));
  const spaces = allAssets
    .filter((candidate) => candidate.parentAssetId === location.id && candidate.kind !== 'item')
    .slice()
    .sort(compareContainedAssetSummaries)
    .map(toAssetCardViewModel);
  const items = allAssets
    .filter((candidate) => candidate.kind === 'item')
    .map((candidate) => {
      const relativePath = relativePathFromLocation(location.id, candidate, assetsById);
      if (!relativePath) {
        return undefined;
      }
      return {
        ...toAssetCardViewModel(candidate),
        relativePath,
        relativePathLabel: relativePath.length > 0
          ? relativePath.map((crumb) => crumb.title).join(' / ')
          : undefined
      } satisfies AssetContainedItemViewModel;
    })
    .filter((candidate): candidate is AssetContainedItemViewModel => candidate !== undefined)
    .sort(compareContainedItems);

  return { spaces, items };
}

function relativePathFromLocation(
  locationId: AssetSummary['id'],
  item: AssetSummary,
  assetsById: ReadonlyMap<AssetSummary['id'], AssetSummary>
): readonly AssetRelativePathCrumbViewModel[] | undefined {
  const reversedPath: AssetRelativePathCrumbViewModel[] = [];
  const visited = new Set<AssetSummary['id']>([item.id]);
  let parentId = item.parentAssetId;

  while (parentId !== undefined) {
    if (parentId === locationId) {
      return reversedPath.reverse();
    }
    if (visited.has(parentId)) {
      return undefined;
    }
    visited.add(parentId);
    const parent = assetsById.get(parentId);
    if (!parent) {
      return undefined;
    }
    reversedPath.push({ id: parent.id, title: parent.title });
    parentId = parent.parentAssetId;
  }

  return undefined;
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

function countLabel(count: number, noun: string): string {
  return `${count.toString()} ${count === 1 ? noun : `${noun}s`}`;
}

function checkoutLabel(asset: AssetSummary): string {
  if (!asset.currentCheckout) {
    return 'Available';
  }
  const date = new Date(asset.currentCheckout.checkedOutAt);
  if (Number.isNaN(date.getTime())) {
    return 'Checked out';
  }
  return `Checked out ${date.toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  })}`;
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
    return locationTrail[0] ?? 'Unplaced';
  }

  return localTrail.join(' / ');
}

function labelParentLocationTrail(asset: AssetSummary): string {
  if (asset.parentLocationTrail.length === 0) {
    return 'Inventory root';
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
      return 'Container';
    case 'item':
      return 'Item';
    case 'location':
      return 'Place';
  }
}

function placeholderForKind(kind: AssetSummary['kind']): string {
  switch (kind) {
    case 'container':
      return 'Box';
    case 'item':
      return 'Item';
    case 'location':
      return 'Place';
  }
}
