export function assetTagSearchHref(tag: AssetTagSearchSource) {
  const params = tag.id
    ? { tagId: tag.id, tagLabel: tag.label }
    : { query: tag.label };
  return {
    pathname: '/search',
    params
  } as const;
}

export type AssetTagSearchNavigator = {
  push: (href: ReturnType<typeof assetTagSearchHref>) => void;
};

export type AssetTagSearchSource = {
  readonly id?: string;
  readonly label: string;
};

export function navigateToAssetTagSearch(navigator: AssetTagSearchNavigator, tag: AssetTagSearchSource): void {
  navigator.push(assetTagSearchHref(tag));
}
