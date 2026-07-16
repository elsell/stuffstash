import type { AppearancePreference } from '../../application/settings/AppearancePreference';

export type SettingsDestination =
  | 'account'
  | 'appearance'
  | 'sharing'
  | 'voice-setup'
  | 'connection'
  | 'about'
  | 'diagnostics';

export type SettingsRootPresentationInput = {
  readonly principal: {
    readonly id: string;
    readonly primaryLabel: string;
  };
  readonly appearance: AppearancePreference;
  readonly selectedTenant: {
    readonly id: string;
    readonly name: string;
    readonly permissions: readonly string[];
  };
  readonly selectedInventory: {
    readonly id: string;
    readonly name: string;
    readonly permissions: readonly string[];
  };
  readonly voiceReadiness?: string;
  readonly serverUrl: string;
  readonly appVersion: string;
  readonly authenticationMode: 'oidc-sso' | 'unconfigured';
};

export type SettingsRootRow = {
  readonly id: string;
  readonly label: string;
  readonly value?: string;
  readonly context?: string;
  readonly destination: SettingsDestination;
  readonly accessibilityRole: 'button';
  readonly accessibilityLabel: string;
  readonly showsDisclosure: true;
};

export type SettingsRootSection = {
  readonly id: 'account' | 'preferences' | 'inventory-administration' | 'tenant-administration' | 'connection' | 'about';
  readonly title?: string;
  readonly rows: readonly SettingsRootRow[];
};

export function buildSettingsRootSections(
  input: SettingsRootPresentationInput
): readonly SettingsRootSection[] {
  const sections: SettingsRootSection[] = [
    {
      id: 'account',
      rows: [row('account', 'Account', input.principal.primaryLabel, 'account',
        `Open Account settings for ${input.principal.primaryLabel}`)]
    },
    {
      id: 'preferences',
      title: 'Preferences',
      rows: [row('appearance', 'Appearance', appearanceLabel(input.appearance), 'appearance',
        `Open Appearance settings. Current selection ${appearanceLabel(input.appearance)}`)]
    }
  ];

  if (input.selectedInventory.permissions.includes('share')) {
    sections.push({
      id: 'inventory-administration',
      title: 'Inventory',
      rows: [row(
        'sharing',
        'Sharing',
        input.selectedInventory.name,
        'sharing',
        `Manage invitations for the ${input.selectedInventory.name} inventory`
      )]
    });
  }

  if (input.selectedTenant.permissions.includes('configure')) {
    const readiness = voiceReadinessLabel(input.voiceReadiness);
    sections.push({
      id: 'tenant-administration',
      title: 'Tenant Administration',
      rows: [{
        ...row('voice-setup', 'Voice Setup', readiness, 'voice-setup',
          `Open Voice Setup for the ${input.selectedTenant.name} tenant. ${readiness}`),
        context: `Shared with ${input.selectedTenant.name}`
      }]
    });
  }

  sections.push(
    {
      id: 'connection',
      title: 'Connection',
      rows: [row('server', 'Stuff Stash server', serverHostname(input.serverUrl), 'connection',
        `Open Stuff Stash server settings for ${serverHostname(input.serverUrl)}`)]
    },
    {
      id: 'about',
      title: 'About',
      rows: [
        row('about', 'About Stuff Stash', `Version ${input.appVersion}`, 'about',
          `Open About Stuff Stash. Version ${input.appVersion}`),
        row('diagnostics', 'Diagnostics', undefined, 'diagnostics',
          'Open developer and connection Diagnostics')
      ]
    }
  );
  return sections;
}

export function settingsLayoutMode(input: { readonly fontScale: number }) {
  const stacked = input.fontScale >= 1.3;
  return {
    stacksLabelValueRows: stacked,
    stacksActionGroups: stacked,
    stacksChoiceRows: stacked
  };
}

function row(
  id: string,
  label: string,
  value: string | undefined,
  destination: SettingsDestination,
  accessibilityLabel: string
): SettingsRootRow {
  return {
    id,
    label,
    value,
    destination,
    accessibilityRole: 'button',
    accessibilityLabel,
    showsDisclosure: true
  };
}

export function appearanceLabel(value: AppearancePreference): string {
  return value.charAt(0).toUpperCase() + value.slice(1);
}

export function serverHostname(serverUrl: string): string {
  try {
    return new URL(serverUrl).host || serverUrl;
  } catch {
    return serverUrl;
  }
}

function voiceReadinessLabel(readiness: string | undefined): string {
  switch (readiness) {
    case 'ready':
      return 'Ready';
    case 'needs_attention':
      return 'Needs attention';
    default:
      return 'Check setup';
  }
}
