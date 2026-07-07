import { createRequire } from 'node:module';
import { describe, expect, it, vi } from 'vitest';

const require = createRequire(import.meta.url);

describe('app config', () => {
  it('mirrors mobile runtime configuration into Expo extra', () => {
    vi.stubEnv('EXPO_PUBLIC_STUFF_STASH_API_BASE_URL', 'http://192.168.1.117:8080');
    vi.stubEnv('EXPO_PUBLIC_STUFF_STASH_TENANT_ID', 'tenant-home');
    vi.stubEnv('EXPO_PUBLIC_STUFF_STASH_VOICE_DIAGNOSTICS_ENABLED', 'true');
    vi.stubEnv('EXPO_PUBLIC_STUFF_STASH_DIRECT_UPLOAD_LOCAL_TARGETS_ENABLED', 'true');
    delete require.cache[require.resolve('./app.config.js')];

    const config = require('./app.config.js');

    expect(config.expo.extra.stuffStash).toEqual({
      apiBaseUrl: 'http://192.168.1.117:8080',
      tenantId: 'tenant-home',
      voiceDeveloperDiagnosticsEnabled: 'true',
      directUploadLocalDevelopmentTargetsEnabled: 'true'
    });
    expect(config.expo.plugins).toContain('expo-secure-store');
    expect(config.expo.plugins).toContain('expo-web-browser');
  });
});
