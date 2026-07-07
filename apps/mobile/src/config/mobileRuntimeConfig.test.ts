import { describe, expect, it } from 'vitest';
import { mergeMobileRuntimeConfigSources, parseMobileRuntimeConfig } from './mobileRuntimeConfigCore';

describe('mobileRuntimeConfig', () => {
  it('parses required Expo public API configuration', () => {
    expect(
      parseMobileRuntimeConfig({
        apiBaseUrl: 'http://192.168.1.97:8080/',
        tenantId: 'tenant-home'
      })
    ).toEqual({
      apiBaseUrl: 'http://192.168.1.97:8080',
      tenantId: 'tenant-home',
      voiceDeveloperDiagnosticsEnabled: false,
      directUploadLocalDevelopmentTargetsEnabled: false
    });
  });

  it('parses explicit mobile voice developer diagnostics opt-in', () => {
    expect(
      parseMobileRuntimeConfig({
        apiBaseUrl: 'http://192.168.1.97:8080/',
        tenantId: 'tenant-home',
        voiceDeveloperDiagnosticsEnabled: 'true'
      }).voiceDeveloperDiagnosticsEnabled
    ).toBe(true);

    expect(
      parseMobileRuntimeConfig({
        apiBaseUrl: 'http://192.168.1.97:8080/',
        tenantId: 'tenant-home',
        voiceDeveloperDiagnosticsEnabled: '0'
      }).voiceDeveloperDiagnosticsEnabled
    ).toBe(false);
  });

  it('parses explicit local direct upload target opt-in', () => {
    expect(
      parseMobileRuntimeConfig({
        apiBaseUrl: 'http://192.168.1.97:8080/',
        tenantId: 'tenant-home',
        directUploadLocalDevelopmentTargetsEnabled: 'true'
      }).directUploadLocalDevelopmentTargetsEnabled
    ).toBe(true);

    expect(
      parseMobileRuntimeConfig({
        apiBaseUrl: 'http://192.168.1.97:8080/',
        tenantId: 'tenant-home',
        directUploadLocalDevelopmentTargetsEnabled: '0'
      }).directUploadLocalDevelopmentTargetsEnabled
    ).toBe(false);
  });

  it('rejects invalid local direct upload target values', () => {
    expect(() =>
      parseMobileRuntimeConfig({
        apiBaseUrl: 'http://192.168.1.97:8080',
        tenantId: 'tenant-home',
        directUploadLocalDevelopmentTargetsEnabled: 'sometimes'
      })
    ).toThrow('EXPO_PUBLIC_STUFF_STASH_DIRECT_UPLOAD_LOCAL_TARGETS_ENABLED');
  });

  it('rejects invalid mobile voice developer diagnostics values', () => {
    expect(() =>
      parseMobileRuntimeConfig({
        apiBaseUrl: 'http://192.168.1.97:8080',
        tenantId: 'tenant-home',
        voiceDeveloperDiagnosticsEnabled: 'sometimes'
      })
    ).toThrow('EXPO_PUBLIC_STUFF_STASH_VOICE_DIAGNOSTICS_ENABLED');
  });

  it('rejects missing required values', () => {
    expect(() =>
      parseMobileRuntimeConfig({
        apiBaseUrl: 'http://192.168.1.97:8080',
        tenantId: ''
      })
    ).toThrow('EXPO_PUBLIC_STUFF_STASH_TENANT_ID');
  });

  it('keeps Expo extra values when public env values are absent at runtime', () => {
    expect(
      mergeMobileRuntimeConfigSources(
        {
          apiBaseUrl: 'http://192.168.1.117:8080',
          tenantId: 'tenant-home',
          voiceDeveloperDiagnosticsEnabled: 'true',
          directUploadLocalDevelopmentTargetsEnabled: 'true'
        },
        {
          apiBaseUrl: undefined,
          tenantId: '',
          voiceDeveloperDiagnosticsEnabled: undefined,
          directUploadLocalDevelopmentTargetsEnabled: undefined
        }
      )
    ).toEqual({
      apiBaseUrl: 'http://192.168.1.117:8080',
      tenantId: 'tenant-home',
      voiceDeveloperDiagnosticsEnabled: 'true',
      directUploadLocalDevelopmentTargetsEnabled: 'true'
    });
  });
});
