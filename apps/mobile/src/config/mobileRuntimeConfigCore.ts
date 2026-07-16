export type MobileRuntimeConfig = {
  readonly apiBaseUrl: string;
  readonly tenantId: string;
  readonly voiceDeveloperDiagnosticsEnabled: boolean;
  readonly directUploadLocalDevelopmentTargetsEnabled: boolean;
  readonly invitationOrigin?: string;
  readonly invitationAllowInsecureLocalHTTP: boolean;
};

export type MobileRuntimeConfigSeed = {
  readonly apiBaseUrl?: string;
  readonly tenantId?: string;
  readonly voiceDeveloperDiagnosticsEnabled: boolean;
  readonly directUploadLocalDevelopmentTargetsEnabled: boolean;
  readonly invitationOrigin?: string;
  readonly invitationAllowInsecureLocalHTTP: boolean;
};

export type RawMobileRuntimeConfig = {
  readonly apiBaseUrl?: string;
  readonly tenantId?: string;
  readonly voiceDeveloperDiagnosticsEnabled?: string | boolean;
  readonly directUploadLocalDevelopmentTargetsEnabled?: string | boolean;
  readonly invitationOrigin?: string;
  readonly invitationAllowInsecureLocalHTTP?: string | boolean;
};

export function mergeMobileRuntimeConfigSources(
  expoExtra: RawMobileRuntimeConfig,
  expoPublicEnv: RawMobileRuntimeConfig
): RawMobileRuntimeConfig {
  return {
    apiBaseUrl: preferConfigured(expoPublicEnv.apiBaseUrl, expoExtra.apiBaseUrl),
    tenantId: preferConfigured(expoPublicEnv.tenantId, expoExtra.tenantId),
    voiceDeveloperDiagnosticsEnabled: preferConfigured(
      expoPublicEnv.voiceDeveloperDiagnosticsEnabled,
      expoExtra.voiceDeveloperDiagnosticsEnabled
    ),
    directUploadLocalDevelopmentTargetsEnabled: preferConfigured(
      expoPublicEnv.directUploadLocalDevelopmentTargetsEnabled,
      expoExtra.directUploadLocalDevelopmentTargetsEnabled
    ),
    invitationOrigin: preferConfigured(expoPublicEnv.invitationOrigin, expoExtra.invitationOrigin),
    invitationAllowInsecureLocalHTTP: preferConfigured(
      expoPublicEnv.invitationAllowInsecureLocalHTTP,
      expoExtra.invitationAllowInsecureLocalHTTP
    )
  };
}

export function parseMobileRuntimeConfig(input: {
  readonly apiBaseUrl?: string;
  readonly tenantId?: string;
  readonly voiceDeveloperDiagnosticsEnabled?: string | boolean;
  readonly directUploadLocalDevelopmentTargetsEnabled?: string | boolean;
  readonly invitationOrigin?: string;
  readonly invitationAllowInsecureLocalHTTP?: string | boolean;
}): MobileRuntimeConfig {
  const apiBaseUrl = requireValue('EXPO_PUBLIC_STUFF_STASH_API_BASE_URL', input.apiBaseUrl);
  const tenantId = requireValue('EXPO_PUBLIC_STUFF_STASH_TENANT_ID', input.tenantId);
  const voiceDeveloperDiagnosticsEnabled = optionalBooleanValue(
    'EXPO_PUBLIC_STUFF_STASH_VOICE_DIAGNOSTICS_ENABLED',
    input.voiceDeveloperDiagnosticsEnabled
  );
  const directUploadLocalDevelopmentTargetsEnabled = optionalBooleanValue(
    'EXPO_PUBLIC_STUFF_STASH_DIRECT_UPLOAD_LOCAL_TARGETS_ENABLED',
    input.directUploadLocalDevelopmentTargetsEnabled
  );
  const invitationAllowInsecureLocalHTTP = optionalBooleanValue(
    'EXPO_PUBLIC_STUFF_STASH_INVITATION_ALLOW_INSECURE_LOCAL_HTTP',
    input.invitationAllowInsecureLocalHTTP
  );
  const invitationOrigin = optionalInvitationOrigin(
    'EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN',
    input.invitationOrigin,
    invitationAllowInsecureLocalHTTP
  );

  return {
    apiBaseUrl: apiBaseUrl.replace(/\/+$/, ''),
    tenantId,
    voiceDeveloperDiagnosticsEnabled,
    directUploadLocalDevelopmentTargetsEnabled,
    invitationOrigin,
    invitationAllowInsecureLocalHTTP
  };
}

function optionalInvitationOrigin(name: string, value: string | undefined, allowInsecureLocalHTTP: boolean): string | undefined {
  const trimmed = optionalValue(value);
  if (!trimmed) return undefined;
  try {
    const parsed = new URL(trimmed);
    if (
      (parsed.protocol !== 'https:' && !(allowInsecureLocalHTTP && isPrivateLocalHTTPOrigin(parsed))) ||
      parsed.pathname !== '/' ||
      (parsed.protocol === 'https:' && parsed.port !== '') ||
      parsed.search ||
      parsed.hash ||
      parsed.username ||
      parsed.password
    ) {
      throw new Error('invalid');
    }
    return parsed.origin;
  } catch {
    throw new Error(`Invalid mobile runtime configuration value: ${name}.`);
  }
}

function isPrivateLocalHTTPOrigin(origin: URL): boolean {
  if (origin.protocol !== 'http:') return false;
  if (origin.hostname === 'localhost' || origin.hostname === '127.0.0.1' || origin.hostname === '[::1]') return true;
  const parts = origin.hostname.split('.');
  if (parts.length !== 4 || parts.some((part) => !/^\d{1,3}$/.test(part))) return false;
  const octets = parts.map(Number);
  if (octets.some((octet) => octet > 255)) return false;
  return octets[0] === 10 ||
    (octets[0] === 172 && octets[1] >= 16 && octets[1] <= 31) ||
    (octets[0] === 192 && octets[1] === 168);
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
