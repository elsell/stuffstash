import Constants from 'expo-constants';
import {
  MobileRuntimeConfig,
  MobileRuntimeConfigSeed,
  RawMobileRuntimeConfig,
  mergeMobileRuntimeConfigSources,
  optionalBooleanValue,
  optionalValue,
  parseMobileRuntimeConfig
} from './mobileRuntimeConfigCore';

export function loadMobileRuntimeConfig(): MobileRuntimeConfig {
  return parseMobileRuntimeConfig(mergeMobileRuntimeConfigSources(readExpoExtraConfig(), readExpoPublicEnvConfig()));
}

export function loadMobileRuntimeConfigSeed(): MobileRuntimeConfigSeed {
  const config = mergeMobileRuntimeConfigSources(readExpoExtraConfig(), readExpoPublicEnvConfig());

  return {
    apiBaseUrl: optionalValue(config.apiBaseUrl),
    tenantId: optionalValue(config.tenantId),
    performanceTelemetryEnabled: optionalBooleanValue('EXPO_PUBLIC_STUFF_STASH_PERFORMANCE_TELEMETRY_ENABLED', config.performanceTelemetryEnabled),
    voiceDeveloperDiagnosticsEnabled: optionalBooleanValue(
      'EXPO_PUBLIC_STUFF_STASH_VOICE_DIAGNOSTICS_ENABLED',
      config.voiceDeveloperDiagnosticsEnabled
    ),
    directUploadLocalDevelopmentTargetsEnabled: optionalBooleanValue(
      'EXPO_PUBLIC_STUFF_STASH_DIRECT_UPLOAD_LOCAL_TARGETS_ENABLED',
      config.directUploadLocalDevelopmentTargetsEnabled
    ),
    invitationOrigin: optionalValue(config.invitationOrigin),
    invitationAllowInsecureLocalHTTP: optionalBooleanValue(
      'EXPO_PUBLIC_STUFF_STASH_INVITATION_ALLOW_INSECURE_LOCAL_HTTP',
      config.invitationAllowInsecureLocalHTTP
    )
  };
}

function readExpoExtraConfig() {
  const extra = Constants.expoConfig?.extra as
    | {
        readonly stuffStash?: {
          readonly apiBaseUrl?: string;
          readonly tenantId?: string;
          readonly performanceTelemetryEnabled?: string | boolean;
          readonly voiceDeveloperDiagnosticsEnabled?: string | boolean;
          readonly directUploadLocalDevelopmentTargetsEnabled?: string | boolean;
          readonly invitationOrigin?: string;
          readonly invitationAllowInsecureLocalHTTP?: string | boolean;
        };
      }
    | undefined;

  return {
    apiBaseUrl: extra?.stuffStash?.apiBaseUrl,
    tenantId: extra?.stuffStash?.tenantId,
    performanceTelemetryEnabled: extra?.stuffStash?.performanceTelemetryEnabled,
    voiceDeveloperDiagnosticsEnabled: extra?.stuffStash?.voiceDeveloperDiagnosticsEnabled,
    directUploadLocalDevelopmentTargetsEnabled: extra?.stuffStash?.directUploadLocalDevelopmentTargetsEnabled,
    invitationOrigin: extra?.stuffStash?.invitationOrigin,
    invitationAllowInsecureLocalHTTP: extra?.stuffStash?.invitationAllowInsecureLocalHTTP
  };
}

function readExpoPublicEnvConfig() {
  return {
    apiBaseUrl: process.env.EXPO_PUBLIC_STUFF_STASH_API_BASE_URL,
    tenantId: process.env.EXPO_PUBLIC_STUFF_STASH_TENANT_ID,
    performanceTelemetryEnabled: process.env.EXPO_PUBLIC_STUFF_STASH_PERFORMANCE_TELEMETRY_ENABLED,
    voiceDeveloperDiagnosticsEnabled: process.env.EXPO_PUBLIC_STUFF_STASH_VOICE_DIAGNOSTICS_ENABLED,
    directUploadLocalDevelopmentTargetsEnabled: process.env.EXPO_PUBLIC_STUFF_STASH_DIRECT_UPLOAD_LOCAL_TARGETS_ENABLED,
    invitationOrigin: process.env.EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN,
    invitationAllowInsecureLocalHTTP: process.env.EXPO_PUBLIC_STUFF_STASH_INVITATION_ALLOW_INSECURE_LOCAL_HTTP
  };
}
