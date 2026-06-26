import { describe, expect, it } from 'vitest';
import { mergeMobileRuntimeConfigSources, parseMobileRuntimeConfig } from './mobileRuntimeConfigCore';

describe('mobileRuntimeConfig', () => {
  it('parses required Expo public API configuration', () => {
    expect(
      parseMobileRuntimeConfig({
        apiBaseUrl: 'http://192.168.1.97:8080/',
        tenantId: 'tenant-home',
        devToken: 'dev:john:john@example.com'
      })
    ).toEqual({
      apiBaseUrl: 'http://192.168.1.97:8080',
      tenantId: 'tenant-home',
      devToken: 'dev:john:john@example.com',
      voiceDeveloperDiagnosticsEnabled: false
    });
  });

  it('parses explicit mobile voice developer diagnostics opt-in', () => {
    expect(
      parseMobileRuntimeConfig({
        apiBaseUrl: 'http://192.168.1.97:8080/',
        tenantId: 'tenant-home',
        devToken: 'dev:john:john@example.com',
        voiceDeveloperDiagnosticsEnabled: 'true'
      }).voiceDeveloperDiagnosticsEnabled
    ).toBe(true);

    expect(
      parseMobileRuntimeConfig({
        apiBaseUrl: 'http://192.168.1.97:8080/',
        tenantId: 'tenant-home',
        devToken: 'dev:john:john@example.com',
        voiceDeveloperDiagnosticsEnabled: '0'
      }).voiceDeveloperDiagnosticsEnabled
    ).toBe(false);
  });

  it('rejects invalid mobile voice developer diagnostics values', () => {
    expect(() =>
      parseMobileRuntimeConfig({
        apiBaseUrl: 'http://192.168.1.97:8080',
        tenantId: 'tenant-home',
        devToken: 'dev:john',
        voiceDeveloperDiagnosticsEnabled: 'sometimes'
      })
    ).toThrow('EXPO_PUBLIC_STUFF_STASH_VOICE_DIAGNOSTICS_ENABLED');
  });

  it('rejects missing required values', () => {
    expect(() =>
      parseMobileRuntimeConfig({
        apiBaseUrl: 'http://192.168.1.97:8080',
        tenantId: '',
        devToken: 'dev:john'
      })
    ).toThrow('EXPO_PUBLIC_STUFF_STASH_TENANT_ID');
  });

  it('keeps Expo extra values when public env values are absent at runtime', () => {
    expect(
      mergeMobileRuntimeConfigSources(
        {
          apiBaseUrl: 'http://192.168.1.117:8080',
          tenantId: 'tenant-home',
          devToken: 'dev:owner',
          voiceDeveloperDiagnosticsEnabled: 'true'
        },
        {
          apiBaseUrl: undefined,
          tenantId: '',
          devToken: undefined,
          voiceDeveloperDiagnosticsEnabled: undefined
        }
      )
    ).toEqual({
      apiBaseUrl: 'http://192.168.1.117:8080',
      tenantId: 'tenant-home',
      devToken: 'dev:owner',
      voiceDeveloperDiagnosticsEnabled: 'true'
    });
  });
});
