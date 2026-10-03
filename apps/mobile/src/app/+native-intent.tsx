/** AuthSession consumes the original Linking event independently of Expo Router. */
export function redirectSystemPath({ path, initial }: { path: string; initial: boolean }): string {
  try {
    const url = new URL(path);
    if (
      url.protocol === 'stuffstash:' && url.hostname === 'auth' &&
      url.pathname === '/callback' && !url.username && !url.password && !url.port
    ) {
      // Warm callbacks must not replace the screen awaiting sign-in. Cold callbacks
      // have no pending request and must never expose their parameters as a route.
      return initial ? '/' : '';
    }
  } catch {
    // Other paths, including relative app routes, remain the router's responsibility.
  }
  return path;
}
