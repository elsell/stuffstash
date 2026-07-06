import Constants from 'expo-constants';
import type {
  SettingsDiagnostics,
  SettingsDiagnosticsProvider
} from '../../application/settings/SettingsQuery';
import type { MobileRuntimeConfig } from '../../config/mobileRuntimeConfigCore';

export class ExpoSettingsDiagnosticsProvider implements SettingsDiagnosticsProvider {
  constructor(private readonly runtimeConfig: MobileRuntimeConfig | undefined) {}

  getDiagnostics(): SettingsDiagnostics {
    return {
      apiBaseUrl: this.runtimeConfig?.apiBaseUrl ?? 'Not configured',
      appVersion: Constants.expoConfig?.version ?? 'Unknown',
      authenticationMode: this.runtimeConfig ? 'oidc-sso' : 'unconfigured'
    };
  }
}
