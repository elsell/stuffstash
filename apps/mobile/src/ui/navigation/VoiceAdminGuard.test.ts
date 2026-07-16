import { describe, expect, it, vi } from 'vitest';
import type { SettingsLoadState } from '../screens/SettingsScreenState';
import {
  decideVoiceAdminAccess,
  voiceAdminGuardPresentation
} from './VoiceAdminGuard';

vi.mock('react-native', () => ({
  ActivityIndicator: 'ActivityIndicator',
  Pressable: 'Pressable',
  ScrollView: 'ScrollView',
  StyleSheet: { create: (styles: unknown) => styles, hairlineWidth: 1 },
  Text: 'Text',
  useWindowDimensions: () => ({ fontScale: 1 }),
  View: 'View'
}));

vi.mock('lucide-react-native', () => ({ ChevronRight: 'ChevronRight' }));

vi.mock('../theme/AppearanceContext', () => ({
  useAppearancePalette: () => ({
    action: '#0066CC',
    background: '#F7FAFB',
    border: '#C5D0D7',
    danger: '#B42318',
    onAction: '#FFFFFF',
    selected: '#E8F0F5',
    surface: '#FFFFFF',
    text: '#243038',
    textMuted: '#52616B'
  })
}));

// @ts-expect-error Vitest's Vite transform provides raw source imports to structural tests.
const voiceRouteSources = import.meta.glob('../../app/settings/voice/**/*.tsx', {
  eager: true,
  import: 'default',
  query: '?raw'
}) as Record<string, string>;

describe('VoiceAdminGuard', () => {
  it('allows Voice administration only for the selected tenant configure permission', () => {
    expect(decideVoiceAdminAccess(settingsState(['view', 'configure']))).toEqual({
      status: 'allowed'
    });

    expect(decideVoiceAdminAccess(settingsState(['view']))).toEqual({
      status: 'unavailable',
      tenantName: 'Home'
    });
  });

  it('presents a safe tenant-scoped unavailable state with a retry action', () => {
    const presentation = voiceAdminGuardPresentation({
      status: 'unavailable',
      tenantName: 'Home'
    });

    expect(presentation).toEqual({
      title: 'Voice settings unavailable',
      message: 'Only tenant administrators can configure Voice for Home.',
      retryLabel: 'Check Again'
    });
  });

  it('keeps loading and load failures distinct from authorization denial', () => {
    expect(decideVoiceAdminAccess({ status: 'loading' })).toEqual({ status: 'loading' });
    expect(decideVoiceAdminAccess({ status: 'error', message: 'Tenant context unavailable.' }))
      .toEqual({ status: 'error', message: 'Tenant context unavailable.' });
    expect(voiceAdminGuardPresentation({
      status: 'error',
      message: 'Tenant context unavailable.'
    })).toEqual({
      title: 'Could not verify Voice settings access',
      message: 'Tenant context unavailable.',
      retryLabel: 'Retry'
    });
  });

  it('guards every Voice settings route, including nested provider editors', () => {
    expect(Object.keys(voiceRouteSources)).toHaveLength(7);
    for (const source of Object.values(voiceRouteSources)) {
      expect(source).toContain('VoiceAdminGuard');
      expect(source).toContain('<VoiceAdminGuard');
    }
  });
});

function settingsState(permissions: readonly string[]): SettingsLoadState {
  return {
    status: 'ready',
    settings: {
      principal: { id: 'principal-one', primaryLabel: 'owner@example.com' },
      selectedTenant: { id: 'tenant-home', name: 'Home', permissions },
      selectedInventory: { id: 'inventory-home', name: 'Household', permissions: ['view'] },
      serverUrl: 'https://stash.home.test',
      appVersion: '0.0.0',
      authenticationMode: 'oidc-sso'
    }
  };
}
