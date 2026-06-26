export type MobileRuntimeConfig = {
  readonly apiBaseUrl: string;
  readonly tenantId: string;
  readonly devToken: string;
  readonly voiceDeveloperDiagnosticsEnabled: boolean;
};

export type MobileRuntimeConfigSeed = {
  readonly apiBaseUrl?: string;
  readonly tenantId?: string;
  readonly devToken?: string;
  readonly voiceDeveloperDiagnosticsEnabled: boolean;
};

export type RawMobileRuntimeConfig = {
  readonly apiBaseUrl?: string;
  readonly tenantId?: string;
  readonly devToken?: string;
  readonly voiceDeveloperDiagnosticsEnabled?: string | boolean;
};

export function mergeMobileRuntimeConfigSources(
  expoExtra: RawMobileRuntimeConfig,
  expoPublicEnv: RawMobileRuntimeConfig
): RawMobileRuntimeConfig {
  return {
    apiBaseUrl: preferConfigured(expoPublicEnv.apiBaseUrl, expoExtra.apiBaseUrl),
    tenantId: preferConfigured(expoPublicEnv.tenantId, expoExtra.tenantId),
    devToken: preferConfigured(expoPublicEnv.devToken, expoExtra.devToken),
    voiceDeveloperDiagnosticsEnabled: preferConfigured(
      expoPublicEnv.voiceDeveloperDiagnosticsEnabled,
      expoExtra.voiceDeveloperDiagnosticsEnabled
    )
  };
}

export function parseMobileRuntimeConfig(input: {
  readonly apiBaseUrl?: string;
  readonly tenantId?: string;
  readonly devToken?: string;
  readonly voiceDeveloperDiagnosticsEnabled?: string | boolean;
}): MobileRuntimeConfig {
  const apiBaseUrl = requireValue('EXPO_PUBLIC_STUFF_STASH_API_BASE_URL', input.apiBaseUrl);
  const tenantId = requireValue('EXPO_PUBLIC_STUFF_STASH_TENANT_ID', input.tenantId);
  const devToken = requireValue('EXPO_PUBLIC_STUFF_STASH_DEV_TOKEN', input.devToken);
  const voiceDeveloperDiagnosticsEnabled = optionalBooleanValue(
    'EXPO_PUBLIC_STUFF_STASH_VOICE_DIAGNOSTICS_ENABLED',
    input.voiceDeveloperDiagnosticsEnabled
  );

  return {
    apiBaseUrl: apiBaseUrl.replace(/\/+$/, ''),
    tenantId,
    devToken,
    voiceDeveloperDiagnosticsEnabled
  };
}

function requireValue(name: string, value: string | undefined): string {
  const trimmed = value?.trim() ?? '';
  if (trimmed.length === 0) {
    throw new Error(`Missing mobile runtime configuration value: ${name}.`);
  }

  return trimmed;
}

export function optionalValue(value: string | undefined): string | undefined {
  const trimmed = value?.trim() ?? '';
  return trimmed.length > 0 ? trimmed : undefined;
}

function preferConfigured<T extends string | boolean | undefined>(preferred: T, fallback: T): T {
  if (typeof preferred === 'boolean') {
    return preferred;
  }
  if (typeof preferred === 'string' && preferred.trim().length > 0) {
    return preferred;
  }
  return fallback;
}

export function optionalBooleanValue(name: string, value: string | boolean | undefined): boolean {
  if (typeof value === 'boolean') {
    return value;
  }

  const trimmed = value?.trim().toLowerCase() ?? '';
  if (trimmed.length === 0) {
    return false;
  }

  switch (trimmed) {
    case '1':
    case 'true':
    case 'yes':
      return true;
    case '0':
    case 'false':
    case 'no':
      return false;
    default:
      throw new Error(`Invalid mobile runtime configuration value: ${name}.`);
  }
}
