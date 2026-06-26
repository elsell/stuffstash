import Constants from 'expo-constants';
import {
  MobileRuntimeConfig,
  MobileRuntimeConfigSeed,
  optionalBooleanValue,
  optionalValue,
  parseMobileRuntimeConfig
} from './mobileRuntimeConfigCore';

export function loadMobileRuntimeConfig(): MobileRuntimeConfig {
  return parseMobileRuntimeConfig({
    ...readExpoExtraConfig(),
    ...readExpoPublicEnvConfig()
  });
}

export function loadMobileRuntimeConfigSeed(): MobileRuntimeConfigSeed {
  const config = {
    ...readExpoExtraConfig(),
    ...readExpoPublicEnvConfig()
  };

  return {
    apiBaseUrl: optionalValue(config.apiBaseUrl),
    tenantId: optionalValue(config.tenantId),
    devToken: optionalValue(config.devToken),
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
          readonly devToken?: string;
          readonly voiceDeveloperDiagnosticsEnabled?: string | boolean;
        };
      }
    | undefined;

  return {
    apiBaseUrl: extra?.stuffStash?.apiBaseUrl,
    tenantId: extra?.stuffStash?.tenantId,
    devToken: extra?.stuffStash?.devToken,
    voiceDeveloperDiagnosticsEnabled: extra?.stuffStash?.voiceDeveloperDiagnosticsEnabled
  };
}

function readExpoPublicEnvConfig() {
  return {
    apiBaseUrl: process.env.EXPO_PUBLIC_STUFF_STASH_API_BASE_URL,
    tenantId: process.env.EXPO_PUBLIC_STUFF_STASH_TENANT_ID,
    devToken: process.env.EXPO_PUBLIC_STUFF_STASH_DEV_TOKEN,
    voiceDeveloperDiagnosticsEnabled: process.env.EXPO_PUBLIC_STUFF_STASH_VOICE_DIAGNOSTICS_ENABLED
  };
}
