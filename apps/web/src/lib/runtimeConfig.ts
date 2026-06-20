export interface RuntimeConfig {
  apiBaseUrl: string;
  oidcIssuer: string;
  oidcClientId: string;
  oidcRedirectUri: string;
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
  return parseRuntimeConfig(await response.json());
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
    oidcRedirectUri: record.oidcRedirectUri as string
  };
}

function trimTrailingSlash(value: string): string {
  return value.trim().replace(/\/+$/, '');
}
