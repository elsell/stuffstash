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

  it('preserves tracked defaults for loopback requests', () => {
    expect(resolveDevRuntimeConfig(config, {}, 'localhost:5173')).toEqual(config);
  });

  it.each(['stuffstash.example:5173', 'evil.test', ''])('does not trust a public or malformed request host %s', (host) => {
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
