import type { AssetSummary } from '../../domain/assets/AssetSummary';

export type AssetCardViewModel = {
  readonly id: string;
  readonly title: string;
  readonly kindLabel: string;
  readonly customTypeLabel?: string;
  readonly description: string;
  readonly locationTrailLabel: string;
  readonly updatedAtLabel: string;
  readonly photoLabel: string;
  readonly imagePlaceholderLabel: string;
  readonly photo?: {
    readonly uri: string;
    readonly headers?: Readonly<Record<string, string>>;
  };
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
  readonly lifecycleLabel: string;
  readonly isActive: boolean;
  readonly canEdit: boolean;
  readonly canMove: boolean;
  readonly canAddPhotos: boolean;
  readonly canArchive: boolean;
  readonly canRestore: boolean;
  readonly canDeletePermanently: boolean;
  readonly containedAssets: readonly AssetCardViewModel[];
  readonly containedAssetsLabel: string;
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
  return {
    id: asset.id,
    title: asset.title,
    kindLabel: labelAssetKind(asset.kind),
    customTypeLabel: asset.customType,
    description: asset.description,
    locationTrailLabel: labelLocationTrail(asset.locationTrail),
    updatedAtLabel: asset.updatedAtLabel,
    photoLabel: asset.hasPhoto ? 'Photo ready' : 'Needs photo',
    imagePlaceholderLabel: placeholderForKind(asset.kind),
    photo: asset.photo
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
    containedAssets,
    containedAssetsLabel: containedAssets.length === 1 ? '1 thing inside' : `${containedAssets.length.toString()} things inside`,
    canContainAssets: asset.kind === 'container' || asset.kind === 'location',
    canAddContainedAssets: canCreateAsset && canEditAsset && asset.lifecycleState === 'active' && (asset.kind === 'container' || asset.kind === 'location')
  };
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
  if (!asset.parentAssetId) {
    return 'Inventory root';
  }

  const localParentTrail = asset.locationTrail.slice(1, -1);

  if (localParentTrail.length === 0) {
    return 'Inventory root';
  }

  return localParentTrail.join(' / ');
}

function labelAssetKind(kind: AssetSummary['kind']): string {
  switch (kind) {
    case 'container':
      return 'Container';
    case 'item':
      return 'Item';
    case 'location':
      return 'Location';
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
