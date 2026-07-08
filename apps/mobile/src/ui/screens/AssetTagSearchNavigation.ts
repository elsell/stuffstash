export function assetTagSearchHref(label: string) {
  return {
    pathname: '/search',
    params: { query: label }
  } as const;
}

export type AssetTagSearchNavigator = {
  push: (href: ReturnType<typeof assetTagSearchHref>) => void;
};

export type AssetTagSearchSource = {
  readonly label: string;
};

export function navigateToAssetTagSearch(navigator: AssetTagSearchNavigator, tag: AssetTagSearchSource): void {
  navigator.push(assetTagSearchHref(tag.label));
}
