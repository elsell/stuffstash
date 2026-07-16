export type SettingsPrincipal = {
  readonly id: string;
  readonly email?: string;
};

export type SettingsDiagnostics = {
  readonly apiBaseUrl: string;
  readonly appVersion: string;
  readonly authenticationMode: 'oidc-sso' | 'unconfigured';
};

export type SettingsTenantScope = {
  readonly id: string;
  readonly name: string;
  readonly permissions: readonly string[];
};

export type SettingsInventoryScope = {
  readonly id: string;
  readonly name: string;
  readonly permissions: readonly string[];
};

export type SettingsViewModel = {
  readonly principal: {
    readonly id: string;
    readonly primaryLabel: string;
  };
  readonly selectedTenant: SettingsTenantScope;
  readonly selectedInventory: SettingsInventoryScope;
  readonly serverUrl: string;
  readonly appVersion: string;
  readonly authenticationMode: SettingsDiagnostics['authenticationMode'];
};

export interface CurrentPrincipalRepository {
  getCurrentPrincipal(): Promise<SettingsPrincipal>;
}

export interface SettingsDiagnosticsProvider {
  getDiagnostics(): SettingsDiagnostics;
}

export interface SettingsScopeRepository {
  getSelectedScope(): Promise<{
    readonly tenant: SettingsTenantScope;
    readonly inventory: SettingsInventoryScope;
  }>;
}

export class SettingsQuery {
  constructor(
    private readonly principals: CurrentPrincipalRepository,
    private readonly diagnostics: SettingsDiagnosticsProvider,
    private readonly scope: SettingsScopeRepository
  ) {}

  async execute(): Promise<SettingsViewModel> {
    const [principal, diagnostics, selectedScope] = await Promise.all([
      this.principals.getCurrentPrincipal(),
      Promise.resolve(this.diagnostics.getDiagnostics()),
      this.scope.getSelectedScope()
    ]);

    return {
      principal: {
        id: principal.id,
        primaryLabel: principal.email ?? 'Signed in'
      },
      selectedTenant: selectedScope.tenant,
      selectedInventory: selectedScope.inventory,
      serverUrl: diagnostics.apiBaseUrl,
      appVersion: diagnostics.appVersion,
      authenticationMode: diagnostics.authenticationMode
    };
  }
}
