import { createRequire } from 'node:module';
import { afterEach, describe, expect, it, vi } from 'vitest';

const require = createRequire(import.meta.url);

describe('app config', () => {
  afterEach(() => {
    vi.unstubAllEnvs();
    delete require.cache[require.resolve('./app.config.js')];
  });

  it('mirrors mobile runtime configuration into Expo extra', () => {
    vi.stubEnv('EXPO_PUBLIC_STUFF_STASH_API_BASE_URL', 'http://192.168.1.117:8080');
    vi.stubEnv('EXPO_PUBLIC_STUFF_STASH_TENANT_ID', 'tenant-home');
    vi.stubEnv('EXPO_PUBLIC_STUFF_STASH_VOICE_DIAGNOSTICS_ENABLED', 'true');
    vi.stubEnv('EXPO_PUBLIC_STUFF_STASH_DIRECT_UPLOAD_LOCAL_TARGETS_ENABLED', 'true');
    vi.stubEnv('EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN', 'https://stash.example.test');
    vi.stubEnv('EXPO_PUBLIC_STUFF_STASH_INVITATION_ALLOW_INSECURE_LOCAL_HTTP', 'false');
    delete require.cache[require.resolve('./app.config.js')];

    const config = require('./app.config.js');

    expect(config.expo.extra.stuffStash).toEqual({
      apiBaseUrl: 'http://192.168.1.117:8080',
      tenantId: 'tenant-home',
      voiceDeveloperDiagnosticsEnabled: 'true',
      directUploadLocalDevelopmentTargetsEnabled: 'true',
      invitationOrigin: 'https://stash.example.test',
      invitationAllowInsecureLocalHTTP: 'false'
    });
    expect(config.expo.plugins).toContain('expo-secure-store');
    expect(config.expo.plugins).toContain('expo-web-browser');
    expect(config.expo.ios.associatedDomains).toEqual(['applinks:stash.example.test']);
    expect(config.expo.android.intentFilters).toEqual([
      expect.objectContaining({
        action: 'VIEW',
        autoVerify: true,
        data: [{ scheme: 'https', host: 'stash.example.test', path: '/invitations/accept' }]
      })
    ]);
  });

  it('accepts an explicitly enabled private LAN HTTP invitation origin without registering platform app links', () => {
    vi.stubEnv('EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN', 'http://192.168.1.117:5173');
    vi.stubEnv('EXPO_PUBLIC_STUFF_STASH_INVITATION_ALLOW_INSECURE_LOCAL_HTTP', 'true');
    delete require.cache[require.resolve('./app.config.js')];

    const config = require('./app.config.js');

    expect(config.expo.extra.stuffStash.invitationOrigin).toBe('http://192.168.1.117:5173');
    expect(config.expo.ios.associatedDomains).toEqual([]);
    expect(config.expo.android.intentFilters).toEqual([]);
  });

  it('rejects private LAN HTTP without opt-in and public HTTP even with opt-in', () => {
    vi.stubEnv('EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN', 'http://192.168.1.117:5173');
    vi.stubEnv('EXPO_PUBLIC_STUFF_STASH_INVITATION_ALLOW_INSECURE_LOCAL_HTTP', 'false');
    delete require.cache[require.resolve('./app.config.js')];
    expect(() => require('./app.config.js')).toThrow('EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN');

    vi.stubEnv('EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN', 'http://8.8.8.8:5173');
    vi.stubEnv('EXPO_PUBLIC_STUFF_STASH_INVITATION_ALLOW_INSECURE_LOCAL_HTTP', 'true');
    delete require.cache[require.resolve('./app.config.js')];
    expect(() => require('./app.config.js')).toThrow('EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN');
  });

  it('fails a production build when the invitation origin is missing', () => {
    vi.stubEnv('EAS_BUILD_PROFILE', 'production');
    vi.stubEnv('EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN', '');
    delete require.cache[require.resolve('./app.config.js')];

    expect(() => require('./app.config.js')).toThrow(
      'EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN is required for production mobile builds.'
    );
  });

  it('rejects private HTTP origins in production even when local HTTP is enabled', () => {
    vi.stubEnv('EAS_BUILD_PROFILE', 'production');
    vi.stubEnv('EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN', 'http://192.168.1.117:5173');
    vi.stubEnv('EXPO_PUBLIC_STUFF_STASH_INVITATION_ALLOW_INSECURE_LOCAL_HTTP', 'true');
    delete require.cache[require.resolve('./app.config.js')];

    expect(() => require('./app.config.js')).toThrow('must use HTTPS for production');
  });

  it('lets an explicit release-build guard require invitation links outside EAS', () => {
    vi.stubEnv('STUFF_STASH_MOBILE_REQUIRE_INVITATION_LINKS', 'true');
    vi.stubEnv('EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN', '');
    delete require.cache[require.resolve('./app.config.js')];

    expect(() => require('./app.config.js')).toThrow(
      'EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN is required for production mobile builds.'
    );
  });

  it.each([
    'http://stash.example.test',
    'https://stash.example.test:8443',
    'https://user@stash.example.test',
    'https://stash.example.test/invitations'
  ])('rejects an invitation origin that cannot back verified platform links: %s', (origin) => {
    vi.stubEnv('EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN', origin);
    delete require.cache[require.resolve('./app.config.js')];

    expect(() => require('./app.config.js')).toThrow(
      'EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN must be a standard-port HTTPS origin.'
    );
  });
});
