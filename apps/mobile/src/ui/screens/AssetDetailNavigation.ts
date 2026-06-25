export type DeletedAssetNavigator = {
  readonly canGoBack: () => boolean;
  readonly back: () => void;
  readonly replace: (href: '/') => void;
};

export function navigateAfterDeletedAsset(navigator: DeletedAssetNavigator): void {
  if (navigator.canGoBack()) {
    navigator.back();
    return;
  }

  navigator.replace('/');
}
