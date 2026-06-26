import Constants from 'expo-constants';
import {
  MobileRuntimeConfig,
  MobileRuntimeConfigSeed,
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
    devToken: optionalValue(config.devToken)
  };
}

function readExpoExtraConfig() {
  const extra = Constants.expoConfig?.extra as
    | {
        readonly stuffStash?: {
          readonly apiBaseUrl?: string;
          readonly tenantId?: string;
          readonly devToken?: string;
        };
      }
    | undefined;

  return {
    apiBaseUrl: extra?.stuffStash?.apiBaseUrl,
    tenantId: extra?.stuffStash?.tenantId,
    devToken: extra?.stuffStash?.devToken
  };
}

function readExpoPublicEnvConfig() {
  return {
    apiBaseUrl: process.env.EXPO_PUBLIC_STUFF_STASH_API_BASE_URL,
    tenantId: process.env.EXPO_PUBLIC_STUFF_STASH_TENANT_ID,
    devToken: process.env.EXPO_PUBLIC_STUFF_STASH_DEV_TOKEN
  };
}
