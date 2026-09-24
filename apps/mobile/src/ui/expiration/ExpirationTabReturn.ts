export function expirationOriginTab(segments: readonly string[]): 'home' | 'search' {
  return segments.includes('(search)') ? 'search' : 'home';
}

/** Root modals cannot infer the underlying tab from their own route segments. */
export function expirationReturnPath(origin: string | string[] | undefined) {
  return origin === 'search' ? '/(tabs)/(search)/expiration' as const : '/(tabs)/(home)/expiration' as const;
}
