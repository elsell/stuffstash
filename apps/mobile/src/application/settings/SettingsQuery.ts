export type SettingsPrincipal = {
  readonly id: string;
  readonly email?: string;
};

export type SettingsDiagnostics = {
  readonly apiBaseUrl: string;
  readonly appVersion: string;
  readonly authenticationMode: 'local-development-token' | 'unconfigured';
};

export type SettingsRowViewModel = {
  readonly label: string;
  readonly value: string;
};

export type SettingsViewModel = {
  readonly currentUserPrimaryLabel: string;
  readonly currentUserSecondaryLabel: string;
  readonly aboutRows: readonly SettingsRowViewModel[];
  readonly developerRows: readonly SettingsRowViewModel[];
};

export interface CurrentPrincipalRepository {
  getCurrentPrincipal(): Promise<SettingsPrincipal>;
}

export interface SettingsDiagnosticsProvider {
  getDiagnostics(): SettingsDiagnostics;
}

export class SettingsQuery {
  constructor(
    private readonly principals: CurrentPrincipalRepository,
    private readonly diagnostics: SettingsDiagnosticsProvider
  ) {}

  async execute(): Promise<SettingsViewModel> {
    const [principal, diagnostics] = await Promise.all([
      this.principals.getCurrentPrincipal(),
      Promise.resolve(this.diagnostics.getDiagnostics())
    ]);

    return {
      currentUserPrimaryLabel: principal.email ?? principal.id,
      currentUserSecondaryLabel: principal.email ? principal.id : 'Signed in',
      aboutRows: [
        { label: 'App', value: 'Stuff Stash' },
        { label: 'Version', value: diagnostics.appVersion }
      ],
      developerRows: [
        { label: 'API', value: diagnostics.apiBaseUrl },
        { label: 'Auth', value: labelAuthenticationMode(diagnostics.authenticationMode) }
      ]
    };
  }
}

function labelAuthenticationMode(mode: SettingsDiagnostics['authenticationMode']): string {
  switch (mode) {
    case 'local-development-token':
      return 'Local development token';
    case 'unconfigured':
      return 'Not configured';
  }
}
