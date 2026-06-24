export type AssetLifecycleState = 'active' | 'archived';

export type AssetKind = 'item' | 'container' | 'location';

export type AssetSummary = {
  readonly id: AssetId;
  readonly title: string;
  readonly kind: AssetKind;
  readonly lifecycleState: AssetLifecycleState;
  readonly locationLabel: string;
  readonly locationTrail: readonly string[];
  readonly customType?: string;
  readonly description: string;
  readonly updatedAtLabel: string;
  readonly hasPhoto: boolean;
  readonly photo?: AssetPhoto;
};

export type AssetPhoto = {
  readonly uri: string;
  readonly headers?: Readonly<Record<string, string>>;
};

export type AssetId = string & { readonly __brand: 'AssetId' };

export function assetId(value: string): AssetId {
  const trimmed = value.trim();
  if (trimmed.length === 0) {
    throw new Error('Asset ID must not be empty.');
  }

  return trimmed as AssetId;
}

export function countActiveAssets(assets: readonly AssetSummary[]): number {
  return assets.filter((asset) => asset.lifecycleState === 'active').length;
}

export function countAssetsWithPhotos(assets: readonly AssetSummary[]): number {
  return assets.filter((asset) => asset.hasPhoto).length;
}
