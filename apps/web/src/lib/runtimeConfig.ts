import { defaultMediaUploadPolicy, type AttachmentContentType, type MediaUploadPolicy } from '$lib/domain/inventory';

export interface RuntimeConfig {
  apiBaseUrl: string;
  oidcIssuer: string;
  oidcClientId: string;
  oidcRedirectUri: string;
  mediaUploadPolicy: MediaUploadPolicy;
}

const requiredKeys: Array<keyof RuntimeConfig> = [
  'apiBaseUrl',
  'oidcIssuer',
  'oidcClientId',
  'oidcRedirectUri'
];

export async function loadRuntimeConfig(fetchImpl: typeof fetch = fetch): Promise<RuntimeConfig> {
  const response = await fetchImpl('/config.json', { cache: 'no-store' });
  if (!response.ok) {
    throw new Error('Unable to load web runtime configuration.');
  }
  return applyRuntimeConfigOverrides(parseRuntimeConfig(await response.json()), import.meta.env);
}

export function parseRuntimeConfig(value: unknown): RuntimeConfig {
  if (typeof value !== 'object' || value === null) {
    throw new Error('Invalid web runtime configuration.');
  }
  const record = value as Record<string, unknown>;
  for (const key of requiredKeys) {
    if (typeof record[key] !== 'string' || record[key].trim() === '') {
      throw new Error(`Missing web runtime configuration value: ${key}.`);
    }
  }
  return {
    apiBaseUrl: trimTrailingSlash(record.apiBaseUrl as string),
    oidcIssuer: trimTrailingSlash(record.oidcIssuer as string),
    oidcClientId: record.oidcClientId as string,
    oidcRedirectUri: record.oidcRedirectUri as string,
    mediaUploadPolicy: parseMediaUploadPolicy(record.mediaUploadPolicy)
  };
}

function trimTrailingSlash(value: string): string {
  return value.trim().replace(/\/+$/, '');
}

export function applyRuntimeConfigOverrides(config: RuntimeConfig, env: Partial<ImportMetaEnv>): RuntimeConfig {
  const webOrigin = normalizedOrigin(env.VITE_STUFF_STASH_WEB_ORIGIN);
  const apiBaseUrl = nonEmpty(env.VITE_STUFF_STASH_API_BASE_URL) ?? (webOrigin ? originWithPort(webOrigin, '8080') : undefined);
  const oidcIssuer = nonEmpty(env.VITE_STUFF_STASH_OIDC_ISSUER) ?? (webOrigin ? `${originWithPort(webOrigin, '5556')}/dex` : undefined);
  const oidcRedirectUri = nonEmpty(env.VITE_STUFF_STASH_OIDC_REDIRECT_URI) ?? (webOrigin ? `${webOrigin}/callback` : undefined);
  return {
    ...config,
    apiBaseUrl: apiBaseUrl ? trimTrailingSlash(apiBaseUrl) : config.apiBaseUrl,
    oidcIssuer: oidcIssuer ? trimTrailingSlash(oidcIssuer) : config.oidcIssuer,
    oidcClientId: nonEmpty(env.VITE_STUFF_STASH_OIDC_CLIENT_ID) ?? config.oidcClientId,
    oidcRedirectUri: oidcRedirectUri ?? config.oidcRedirectUri
  };
}

function normalizedOrigin(value: string | undefined): string | undefined {
  const origin = nonEmpty(value);
  if (!origin) {
    return undefined;
  }
  return new URL(origin).origin;
}

function originWithPort(origin: string, port: string): string {
  const url = new URL(origin);
  url.port = port;
  return url.origin;
}

function nonEmpty(value: string | undefined): string | undefined {
  const trimmed = value?.trim();
  return trimmed ? trimmed : undefined;
}

function parseMediaUploadPolicy(value: unknown): MediaUploadPolicy {
  if (typeof value !== 'object' || value === null) {
    return defaultMediaUploadPolicy;
  }
  const record = value as Record<string, unknown>;
  const supportedContentTypes = Array.isArray(record.supportedContentTypes)
    ? record.supportedContentTypes.filter(isAttachmentContentType)
    : defaultMediaUploadPolicy.supportedContentTypes;
  const maxBytes = typeof record.maxBytes === 'number' && Number.isFinite(record.maxBytes) && record.maxBytes > 0
    ? Math.floor(record.maxBytes)
    : defaultMediaUploadPolicy.maxBytes;
  return {
    supportedContentTypes: supportedContentTypes.length > 0 ? supportedContentTypes : defaultMediaUploadPolicy.supportedContentTypes,
    maxBytes
  };
}

function isAttachmentContentType(value: unknown): value is AttachmentContentType {
  return value === 'image/jpeg' || value === 'image/png' || value === 'image/webp' || value === 'application/pdf';
}
