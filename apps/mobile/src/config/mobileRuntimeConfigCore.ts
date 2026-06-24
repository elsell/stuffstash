export type MobileRuntimeConfig = {
  readonly apiBaseUrl: string;
  readonly tenantId: string;
  readonly devToken: string;
};

export function parseMobileRuntimeConfig(input: {
  readonly apiBaseUrl?: string;
  readonly tenantId?: string;
  readonly devToken?: string;
}): MobileRuntimeConfig {
  const apiBaseUrl = requireValue('EXPO_PUBLIC_STUFF_STASH_API_BASE_URL', input.apiBaseUrl);
  const tenantId = requireValue('EXPO_PUBLIC_STUFF_STASH_TENANT_ID', input.tenantId);
  const devToken = requireValue('EXPO_PUBLIC_STUFF_STASH_DEV_TOKEN', input.devToken);

  return {
    apiBaseUrl: apiBaseUrl.replace(/\/+$/, ''),
    tenantId,
    devToken
  };
}

function requireValue(name: string, value: string | undefined): string {
  const trimmed = value?.trim() ?? '';
  if (trimmed.length === 0) {
    throw new Error(`Missing mobile runtime configuration value: ${name}.`);
  }

  return trimmed;
}
