type DevRuntimeConfig = Record<string, unknown>;
type DevRuntimeEnvironment = Record<string, string | undefined>;

export function resolveDevRuntimeConfig(
  config: DevRuntimeConfig,
  env: DevRuntimeEnvironment,
  requestHost = ''
): DevRuntimeConfig {
  const configuredWebOrigin = normalizedOrigin(env.VITE_STUFF_STASH_WEB_ORIGIN);
  const requestWebOrigin = configuredWebOrigin ? undefined : trustedPrivateLanOrigin(requestHost);
  const webOrigin = configuredWebOrigin ?? requestWebOrigin;
  const apiBaseUrl = nonEmpty(env.VITE_STUFF_STASH_API_BASE_URL)
    ?? (webOrigin ? originWithPort(webOrigin, '8080') : undefined);
  const oidcIssuer = nonEmpty(env.VITE_STUFF_STASH_OIDC_ISSUER)
    ?? (webOrigin ? `${originWithPort(webOrigin, '5556')}/dex` : undefined);
  const oidcRedirectUri = nonEmpty(env.VITE_STUFF_STASH_OIDC_REDIRECT_URI)
    ?? (webOrigin ? `${webOrigin}/callback` : undefined);

  return {
    ...config,
    apiBaseUrl: apiBaseUrl ? trimTrailingSlash(apiBaseUrl) : config.apiBaseUrl,
    oidcIssuer: oidcIssuer ? trimTrailingSlash(oidcIssuer) : config.oidcIssuer,
    oidcClientId: nonEmpty(env.VITE_STUFF_STASH_OIDC_CLIENT_ID) ?? config.oidcClientId,
    oidcRedirectUri: oidcRedirectUri ?? config.oidcRedirectUri
  };
}

function trustedPrivateLanOrigin(host: string): string | undefined {
  if (!host || host !== host.trim()) {
    return undefined;
  }
  const match = /^(\d{1,3}(?:\.\d{1,3}){3})(?::(\d{1,5}))?$/.exec(host);
  if (!match) {
    return undefined;
  }
  const address = match[1];
  const octets = address.split('.');
  if (!octets.every(isCanonicalOctet) || !isTrustedPrivateIPv4(octets.map(Number))) {
    return undefined;
  }
  const port = match[2];
  if (port && (!isCanonicalDecimal(port) || Number(port) < 1 || Number(port) > 65535)) {
    return undefined;
  }
  return `http://${address}${port ? `:${port}` : ''}`;
}

function isTrustedPrivateIPv4(octets: number[]): boolean {
  return octets[0] === 10
    || (octets[0] === 172 && octets[1] >= 16 && octets[1] <= 31)
    || (octets[0] === 192 && octets[1] === 168)
    || (octets[0] === 169 && octets[1] === 254);
}

function isCanonicalOctet(value: string): boolean {
  return isCanonicalDecimal(value) && Number(value) <= 255;
}

function isCanonicalDecimal(value: string): boolean {
  return String(Number(value)) === value;
}

function normalizedOrigin(value: string | undefined): string | undefined {
  const origin = nonEmpty(value);
  return origin ? new URL(origin).origin : undefined;
}

function originWithPort(origin: string, port: string): string {
  const url = new URL(origin);
  url.port = port;
  return url.origin;
}

function trimTrailingSlash(value: string): string {
  return value.trim().replace(/\/+$/, '');
}

function nonEmpty(value: string | undefined): string | undefined {
  const trimmed = value?.trim();
  return trimmed ? trimmed : undefined;
}
