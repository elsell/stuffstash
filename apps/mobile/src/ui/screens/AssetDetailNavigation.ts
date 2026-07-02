export type DeletedAssetNavigator = {
  readonly canGoBack: () => boolean;
  readonly back: () => void;
  readonly replace: (href: '/') => void;
};

export function assetDetailHref(assetId: string) {
  return {
    pathname: '/assets/[assetId]',
    params: { assetId }
  } as const;
}

export function locationAssetDetailHref(locationId: string, assetId: string) {
  return {
    pathname: '/locations/[locationId]/assets/[assetId]',
    params: { assetId, locationId }
  } as const;
}

export function navigateAfterDeletedAsset(navigator: DeletedAssetNavigator): void {
  if (navigator.canGoBack()) {
    navigator.back();
    return;
  }

  navigator.replace('/');
}
