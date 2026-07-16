import { describe, expect, it } from 'vitest';
import {
  buildSettingsRootSections,
  settingsLayoutMetrics,
  settingsLayoutMode,
  type SettingsRootPresentationInput
} from './SettingsScreenPresentation';

describe('Settings root presentation', () => {
  it('presents native-feeling grouped navigation instead of equally weighted cards', () => {
    const sections = buildSettingsRootSections(input());

    expect(sections.map((section) => section.id)).toEqual([
      'account',
      'preferences',
      'scope',
      'connection',
      'about'
    ]);
    expect(sections.flatMap((section) => section.rows.map((row) => row.id))).toEqual([
      'account',
      'appearance',
      'tenant-settings',
      'inventory-settings',
      'server',
      'about',
      'diagnostics'
    ]);
    expect(sections.flatMap((section) => section.rows)).toEqual(expect.arrayContaining([
      expect.objectContaining({
        id: 'appearance',
        label: 'Appearance',
        value: 'System',
        destination: 'appearance',
        showsDisclosure: true
      }),
      expect.objectContaining({
        id: 'server',
        label: 'Stuff Stash server',
        value: 'stash.home.test',
        destination: 'connection',
        showsDisclosure: true
      })
    ]));
  });

  it('identifies the named household and inventory settings levels', () => {
    const configurable = buildSettingsRootSections(input({
      selectedTenant: {
        id: 'tenant-home',
        name: 'Home',
        permissions: ['view', 'configure']
      }
    }));
    expect(configurable.flatMap((section) => section.rows)).toEqual(expect.arrayContaining([
      expect.objectContaining({ id: 'tenant-settings', label: 'Home', value: 'Household settings' }),
      expect.objectContaining({ id: 'inventory-settings', label: 'Household', value: 'In Home' })
    ]));
  });

  it('keeps the inventory settings destination readable for viewers', () => {
    const viewer = buildSettingsRootSections(input({
      selectedInventory: {
        id: 'inventory-home',
        name: 'Household',
        permissions: ['view']
      }
    }));
    expect(viewer.flatMap((section) => section.rows).some((row) => row.id === 'inventory-settings')).toBe(true);
  });

  it('keeps opaque and developer-only details behind Diagnostics', () => {
    const sections = buildSettingsRootSections(input());
    const rootRows = sections.flatMap((section) => section.rows);

    expect(rootRows.find((row) => row.id === 'account')).toMatchObject({
      label: 'Account',
      value: 'john@example.com',
      destination: 'account'
    });
    expect(rootRows.find((row) => row.id === 'diagnostics')).toMatchObject({
      label: 'Diagnostics',
      destination: 'diagnostics'
    });
    expect(JSON.stringify(rootRows)).not.toContain('principal-subject-123');
    expect(JSON.stringify(rootRows)).not.toContain('OIDC SSO');
    expect(JSON.stringify(rootRows)).not.toContain('http://stash.home.test/api');
  });

  it('gives every disclosure row a contextual accessibility label', () => {
    const rows = buildSettingsRootSections(input()).flatMap((section) => section.rows);

    expect(rows.every((row) => row.accessibilityRole === 'button')).toBe(true);
    expect(rows.every((row) => row.accessibilityLabel.trim().length > row.label.length)).toBe(true);
  });
});

describe('Settings Dynamic Type presentation', () => {
  it('uses one shared content column and minimum native touch target', () => {
    expect(settingsLayoutMetrics).toEqual({
      bottomSpacing: 32,
      horizontalInset: 16,
      minimumTouchTarget: 44,
      sectionSpacing: 24
    });
  });

  it('stacks values and controls at accessibility content sizes', () => {
    expect(settingsLayoutMode({ fontScale: 1 })).toEqual({
      stacksLabelValueRows: false,
      stacksActionGroups: false,
      stacksChoiceRows: false
    });
    expect(settingsLayoutMode({ fontScale: 1.35 })).toEqual({
      stacksLabelValueRows: true,
      stacksActionGroups: true,
      stacksChoiceRows: true
    });
    expect(settingsLayoutMode({ fontScale: 2.8 })).toEqual({
      stacksLabelValueRows: true,
      stacksActionGroups: true,
      stacksChoiceRows: true
    });
  });
});

function input(overrides: Partial<SettingsRootPresentationInput> = {}): SettingsRootPresentationInput {
  return {
    principal: {
      id: 'principal-subject-123',
      primaryLabel: 'john@example.com'
    },
    appearance: 'system',
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
    voiceReadiness: 'needs_attention',
    serverUrl: 'http://stash.home.test/api',
    appVersion: '0.0.0',
    authenticationMode: 'oidc-sso',
    ...overrides
  };
}
