import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('expo-secure-store', () => ({
  getItemAsync: vi.fn(),
  setItemAsync: vi.fn(),
  deleteItemAsync: vi.fn(),
  WHEN_UNLOCKED_THIS_DEVICE_ONLY: 1
}));

vi.mock('expo-auth-session', () => ({
  AuthRequest: class {},
  ResponseType: { Code: 'code' },
  exchangeCodeAsync: vi.fn(),
  fetchDiscoveryAsync: vi.fn(),
  refreshAsync: vi.fn()
}));

vi.mock('expo-web-browser', () => ({
  maybeCompleteAuthSession: vi.fn()
}));

vi.mock('react-native', () => ({
  Platform: { OS: 'ios' }
}));

vi.mock('expo-constants', () => ({
  default: { expoConfig: { extra: {} } }
}));

vi.mock('expo-file-system', () => ({
  Directory: class {},
  File: class {},
  Paths: { document: {} }
}));

vi.mock('expo-file-system/legacy', () => ({}));

vi.mock('expo-audio', () => ({
  createAudioPlayer: vi.fn(),
  requestRecordingPermissionsAsync: vi.fn(),
  setAudioModeAsync: vi.fn(),
  RecordingPresets: {}
}));

vi.mock('expo-audio/src/AudioModule', () => ({
  default: {}
}));

import * as SecureStore from 'expo-secure-store';
import { createMobileComposition } from './mobileComposition';

describe('createMobileComposition', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('notifies the app shell when an authenticated API response requires sign-in', async () => {
    vi.mocked(SecureStore.getItemAsync).mockResolvedValue(JSON.stringify({
      apiBaseUrl: 'http://api.local',
      issuer: 'https://accounts.example.test',
      clientId: 'stuff-stash-mobile',
      idToken: 'stale-id-token',
      refreshToken: 'refresh-token',
      expiresAt: Date.now() + 600_000
    }));
    const fetchMock = vi.fn(async () =>
      Response.json({
        error: {
          code: 'authentication_required',
          message: 'Authentication required.',
          details: []
        },
        meta: {}
      }, { status: 401 })
    );
    vi.stubGlobal('fetch', fetchMock);
    let authRequiredCount = 0;
    const composition = createMobileComposition(
      { apiBaseUrl: 'http://api.local', tenantId: 'tenant-home' },
      { onAuthenticationRequired: () => { authRequiredCount += 1; } }
    );

    await expect(composition.homeDashboardQuery.execute()).rejects.toThrow('Authentication required.');

    expect(authRequiredCount).toBe(1);
    expect(fetchMock).toHaveBeenCalled();
  });

  it('wires customization observability to the production composition and bounded event sink', () => {
    const events: unknown[] = [];
    const composition = createMobileComposition(
      { apiBaseUrl: 'http://api.local', tenantId: 'tenant-home' },
      { onCustomizationEvent: (event) => events.push(event) }
    );

    composition.customizationObservability.record({ name: 'settings.opened' });
    composition.customizationObservability.record({ name: 'settings.level_selected', scope: 'inventory' });

    expect(events).toEqual([
      { name: 'settings.opened' },
      { name: 'settings.level_selected', scope: 'inventory' }
    ]);
    expect(composition.customizationObservability.events()).toEqual(events);
  });
});
