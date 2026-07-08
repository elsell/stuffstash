export function assetTagSearchHref(label: string) {
  return {
    pathname: '/search',
    params: { query: label }
  } as const;
}
