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

export type AssetDetailViewModel = {
  readonly id: string;
  readonly title: string;
  readonly kindLabel: string;
  readonly customTypeLabel?: string;
  readonly description: string;
  readonly locationTrailLabel: string;
  readonly lifecycleLabel: string;
  readonly updatedAtLabel: string;
  readonly photoLabel: string;
  readonly imagePlaceholderLabel: string;
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

export function toAssetDetailViewModel(asset: AssetSummary): AssetDetailViewModel {
  return {
    ...toAssetCardViewModel(asset),
    lifecycleLabel: asset.lifecycleState === 'active' ? 'Active' : 'Archived'
  };
}

function labelLocationTrail(locationTrail: readonly string[]): string {
  const localTrail = locationTrail.slice(1);

  if (localTrail.length === 0) {
    return locationTrail[0] ?? 'Unplaced';
  }

  return localTrail.join(' / ');
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
