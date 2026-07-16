import { describe, expect, it } from 'vitest';

import { resolveDevRuntimeConfig } from './devRuntimeConfig';

const config = {
  apiBaseUrl: 'http://localhost:8080',
  oidcIssuer: 'http://localhost:5556/dex',
  oidcClientId: 'stuff-stash-web-local',
  oidcRedirectUri: 'http://localhost:5173/callback'
};

describe('development runtime configuration', () => {
  it('keeps OIDC callback and service URLs on a private LAN request origin', () => {
    expect(resolveDevRuntimeConfig(config, {}, '192.168.1.117:5173')).toMatchObject({
      apiBaseUrl: 'http://192.168.1.117:8080',
      oidcIssuer: 'http://192.168.1.117:5556/dex',
      oidcRedirectUri: 'http://192.168.1.117:5173/callback'
    });
  });

  it.each([
    '10.0.0.1:5173',
    '10.255.255.254',
    '172.16.0.1:5173',
    '172.31.255.254',
    '192.168.0.1:5173',
    '192.168.255.254',
    '169.254.0.1:5173',
    '169.254.255.254'
  ])('trusts canonical private IPv4 host %s', (host) => {
    expect(resolveDevRuntimeConfig(config, {}, host).oidcRedirectUri).toBe(`http://${host}/callback`);
  });

  it('preserves tracked defaults for loopback requests', () => {
    expect(resolveDevRuntimeConfig(config, {}, 'localhost:5173')).toEqual(config);
  });

  it.each([
    'stuffstash.example:5173',
    'evil.test',
    '',
    '127.0.0.1:5173',
    '172.15.255.255:5173',
    '172.32.0.0:5173',
    '169.253.255.255:5173',
    '169.255.0.0:5173',
    '192.0.2.1:5173',
    '192.168.1',
    '3232235777',
    '0xc0.0xa8.0x01.0x01:5173',
    '0300.0250.0001.0001:5173',
    '192.168.001.117:5173',
    'user@192.168.1.117:5173',
    '192.168.1.117/path',
    '192.168.1.117?port=5173',
    '192.168.1.117:5173:8080',
    '192.168.1.117:not-a-port',
    '192.168.1.117:70000',
    '256.168.1.117:5173',
    '[::1]:5173',
    '[fd00::1]:5173',
    '[fe80::1]:5173'
  ])('does not trust a public or malformed request host %s', (host) => {
    expect(resolveDevRuntimeConfig(config, {}, host)).toEqual(config);
  });

  it('lets explicit runtime settings win over request-derived LAN values', () => {
    expect(resolveDevRuntimeConfig(config, {
      VITE_STUFF_STASH_API_BASE_URL: 'https://api.example.test/',
      VITE_STUFF_STASH_OIDC_ISSUER: 'https://login.example.test/dex/',
      VITE_STUFF_STASH_OIDC_CLIENT_ID: 'configured-client',
      VITE_STUFF_STASH_OIDC_REDIRECT_URI: 'https://web.example.test/callback'
    }, '192.168.1.117:5173')).toMatchObject({
      apiBaseUrl: 'https://api.example.test',
      oidcIssuer: 'https://login.example.test/dex',
      oidcClientId: 'configured-client',
      oidcRedirectUri: 'https://web.example.test/callback'
    });
  });
});
