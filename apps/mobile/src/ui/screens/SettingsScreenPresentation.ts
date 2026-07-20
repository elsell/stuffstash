import type { AppearancePreference } from '../../application/settings/AppearancePreference';
import { spacing } from '../theme/tokens';

export const settingsLayoutMetrics = {
  bottomSpacing: spacing.xl,
  horizontalInset: spacing.md,
  minimumTouchTarget: 44,
  sectionSpacing: spacing.lg
} as const;

export type SettingsDestination =
  | 'account'
  | 'appearance'
  | 'tenant-settings'
  | 'inventory-settings'
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
  readonly id: 'account' | 'preferences' | 'scope' | 'connection' | 'about';
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

  sections.push({
    id: 'scope',
    title: 'Household and Inventory',
    rows: [
      row('tenant-settings', input.selectedTenant.name, 'Household settings', 'tenant-settings',
        `Open household settings for ${input.selectedTenant.name}`),
      row('inventory-settings', input.selectedInventory.name, `In ${input.selectedTenant.name}`, 'inventory-settings',
        `Open inventory settings for ${input.selectedInventory.name}, in ${input.selectedTenant.name}`)
    ]
  });

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
