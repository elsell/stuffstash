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
  const trimmedHost = host.trim();
  if (!trimmedHost) {
    return undefined;
  }
  try {
    const url = new URL(`http://${trimmedHost}`);
    return isPrivateIPv4(url.hostname) ? url.origin : undefined;
  } catch {
    return undefined;
  }
}

function isPrivateIPv4(hostname: string): boolean {
  const octets = hostname.split('.').map(Number);
  if (octets.length !== 4 || octets.some((octet) => !Number.isInteger(octet) || octet < 0 || octet > 255)) {
    return false;
  }
  return octets[0] === 10
    || (octets[0] === 172 && octets[1] >= 16 && octets[1] <= 31)
    || (octets[0] === 192 && octets[1] === 168)
    || (octets[0] === 169 && octets[1] === 254);
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
