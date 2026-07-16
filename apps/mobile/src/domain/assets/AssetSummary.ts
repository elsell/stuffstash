export type AssetLifecycleState = 'active' | 'archived';

export type AssetKind = 'item' | 'container' | 'location';

export type AssetSummary = {
  readonly id: AssetId;
  readonly title: string;
  readonly kind: AssetKind;
  readonly lifecycleState: AssetLifecycleState;
  readonly parentAssetId?: AssetId;
  readonly locationLabel: string;
  readonly locationTrail: readonly string[];
  readonly parentLocationTrail: readonly AssetLocationTrailSegment[];
  readonly customType?: string;
  readonly description: string;
  readonly updatedAtLabel: string;
  readonly hasPhoto: boolean;
  readonly photos?: readonly AssetPhoto[];
  readonly photo?: AssetPhoto;
  readonly currentCheckout?: CurrentCheckoutSummary;
  readonly tags?: readonly AssetTagSummary[];
  readonly undoableOperationId?: string;
};

export type AssetLocationTrailSegment = {
  readonly id: AssetId;
  readonly title: string;
};

export type AssetTagSummary = {
  readonly id: string;
  readonly key: string;
  readonly displayName: string;
  readonly color?: string;
};

export function assetTagKeyFromDisplayName(value: string): string {
  let key = '';
  let lastHyphen = false;
  for (const character of value.trim().toLowerCase()) {
    if ((character >= 'a' && character <= 'z') || (character >= '0' && character <= '9')) {
      key += character;
      lastHyphen = false;
      continue;
    }
    if (key.length > 0 && !lastHyphen) {
      key += '-';
      lastHyphen = true;
    }
  }
  key = key.replace(/^-+|-+$/g, '');
  if (key.length > 80) {
    key = key.slice(0, 80).replace(/-+$/g, '');
  }
  return key;
}

export type CurrentCheckoutSummary = {
  readonly id: string;
  readonly state: string;
  readonly checkedOutAt: string;
  readonly checkedOutByPrincipalId: string;
};

export type AssetPhoto = {
  readonly id?: string;
  readonly fileName?: string;
  readonly contentType?: string;
  readonly sizeBytes?: number;
  readonly uri: string;
  readonly heroUri?: string;
  readonly heroHeaders?: Readonly<Record<string, string>>;
  readonly viewerUri?: string;
  readonly viewerHeaders?: Readonly<Record<string, string>>;
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
