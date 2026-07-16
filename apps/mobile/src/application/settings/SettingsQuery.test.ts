import { describe, expect, it } from 'vitest';
import {
  CurrentPrincipalRepository,
  SettingsDiagnosticsProvider,
  SettingsScopeRepository,
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

class FakeSettingsScopeRepository implements SettingsScopeRepository {
  async getSelectedScope() {
    return {
      tenant: {
        id: 'tenant-home',
        name: 'Home',
        permissions: ['view', 'configure']
      },
      inventory: {
        id: 'inventory-home',
        name: 'Household',
        permissions: ['view', 'share']
      }
    };
  }
}

describe('SettingsQuery', () => {
  it('builds production-shaped settings from principal and diagnostics ports', async () => {
    const query = new SettingsQuery(
      new FakeCurrentPrincipalRepository(),
      new FakeSettingsDiagnosticsProvider(),
      new FakeSettingsScopeRepository()
    );

    await expect(query.execute()).resolves.toEqual({
      principal: {
        id: 'john',
        primaryLabel: 'john@example.com'
      },
      selectedTenant: {
        id: 'tenant-home',
        name: 'Home',
        permissions: ['view', 'configure']
      },
      selectedInventory: {
        id: 'inventory-home',
        name: 'Household',
        permissions: ['view', 'share']
      },
      serverUrl: 'http://192.168.1.97:8090',
      appVersion: '0.0.0',
      authenticationMode: 'oidc-sso'
    });
  });
});
