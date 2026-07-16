import { describe, expect, it } from 'vitest';
import {
  buildSettingsRootSections,
  settingsLayoutMode,
  type SettingsRootPresentationInput
} from './SettingsScreenPresentation';

describe('Settings root presentation', () => {
  it('presents native-feeling grouped navigation instead of equally weighted cards', () => {
    const sections = buildSettingsRootSections(input());

    expect(sections.map((section) => section.id)).toEqual([
      'account',
      'preferences',
      'inventory-administration',
      'tenant-administration',
      'connection',
      'about'
    ]);
    expect(sections.flatMap((section) => section.rows.map((row) => row.id))).toEqual([
      'account',
      'appearance',
      'sharing',
      'voice-setup',
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

  it('identifies tenant-wide voice configuration and exposes it only with configure permission', () => {
    const configurable = buildSettingsRootSections(input({
      selectedTenant: {
        id: 'tenant-home',
        name: 'Home',
        permissions: ['view', 'configure']
      }
    }));
    const voiceRow = configurable
      .flatMap((section) => section.rows)
      .find((row) => row.id === 'voice-setup');

    expect(voiceRow).toMatchObject({
      label: 'Voice Setup',
      value: 'Needs attention',
      context: 'Shared with Home',
      destination: 'voice-setup',
      accessibilityLabel: 'Open Voice Setup for the Home tenant. Needs attention',
      showsDisclosure: true
    });

    const viewer = buildSettingsRootSections(input({
      selectedTenant: {
        id: 'tenant-home',
        name: 'Home',
        permissions: ['view']
      }
    }));
    expect(viewer.flatMap((section) => section.rows).some((row) => row.id === 'voice-setup'))
      .toBe(false);
    expect(viewer.some((section) => section.id === 'tenant-administration')).toBe(false);
  });

  it('exposes inventory Sharing only for the selected inventory share permission', () => {
    const sharing = buildSettingsRootSections(input()).flatMap((section) => section.rows)
      .find((row) => row.id === 'sharing');
    expect(sharing).toMatchObject({
      label: 'Sharing',
      value: 'Household',
      destination: 'sharing',
      accessibilityLabel: 'Manage invitations for the Household inventory'
    });

    const viewer = buildSettingsRootSections(input({
      selectedInventory: {
        id: 'inventory-home',
        name: 'Household',
        permissions: ['view']
      }
    }));
    expect(viewer.flatMap((section) => section.rows).some((row) => row.id === 'sharing'))
      .toBe(false);
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
