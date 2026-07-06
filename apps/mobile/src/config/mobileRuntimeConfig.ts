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
    voiceDeveloperDiagnosticsEnabled: optionalBooleanValue(
      'EXPO_PUBLIC_STUFF_STASH_VOICE_DIAGNOSTICS_ENABLED',
      config.voiceDeveloperDiagnosticsEnabled
    )
  };
}

function readExpoExtraConfig() {
  const extra = Constants.expoConfig?.extra as
    | {
        readonly stuffStash?: {
          readonly apiBaseUrl?: string;
          readonly tenantId?: string;
          readonly voiceDeveloperDiagnosticsEnabled?: string | boolean;
        };
      }
    | undefined;

  return {
    apiBaseUrl: extra?.stuffStash?.apiBaseUrl,
    tenantId: extra?.stuffStash?.tenantId,
    voiceDeveloperDiagnosticsEnabled: extra?.stuffStash?.voiceDeveloperDiagnosticsEnabled
  };
}

function readExpoPublicEnvConfig() {
  return {
    apiBaseUrl: process.env.EXPO_PUBLIC_STUFF_STASH_API_BASE_URL,
    tenantId: process.env.EXPO_PUBLIC_STUFF_STASH_TENANT_ID,
    voiceDeveloperDiagnosticsEnabled: process.env.EXPO_PUBLIC_STUFF_STASH_VOICE_DIAGNOSTICS_ENABLED
  };
}
