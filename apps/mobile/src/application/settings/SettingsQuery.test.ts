import { describe, expect, it } from 'vitest';
import {
  CurrentPrincipalRepository,
  SettingsDiagnosticsProvider,
  SettingsQuery
} from './SettingsQuery';

class FakeCurrentPrincipalRepository implements CurrentPrincipalRepository {
  async getCurrentPrincipal() {
    return {
      id: 'john',
      email: 'john@example.com'
    };
  }
}

class FakeSettingsDiagnosticsProvider implements SettingsDiagnosticsProvider {
  getDiagnostics() {
    return {
      apiBaseUrl: 'http://192.168.1.97:8090',
      appVersion: '0.0.0',
      authenticationMode: 'oidc-sso' as const
    };
  }
}

describe('SettingsQuery', () => {
  it('builds production-shaped settings from principal and diagnostics ports', async () => {
    const query = new SettingsQuery(
      new FakeCurrentPrincipalRepository(),
      new FakeSettingsDiagnosticsProvider()
    );

    await expect(query.execute()).resolves.toEqual({
      currentUserPrimaryLabel: 'john@example.com',
      currentUserSecondaryLabel: 'john',
      aboutRows: [
        { label: 'App', value: 'Stuff Stash' },
        { label: 'Version', value: '0.0.0' }
      ],
      developerRows: [
        { label: 'API', value: 'http://192.168.1.97:8090' },
        { label: 'Auth', value: 'OIDC SSO' }
      ]
    });
  });
});
