import { describe, expect, it } from 'vitest';
import { parseMobileRuntimeConfig } from './mobileRuntimeConfigCore';

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
      devToken: 'dev:john:john@example.com'
    });
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
});
