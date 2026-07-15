import { describe, expect, it } from 'vitest';
import developmentRuntimeConfig from '../../static/config.json';
import { applyRuntimeConfigOverrides, parseRuntimeConfig } from './runtimeConfig';

describe('parseRuntimeConfig', () => {
  it('keeps the development Files workflow aligned with the initial API attachment types', () => {
    const parsed = parseRuntimeConfig(developmentRuntimeConfig);

    expect(parsed.mediaUploadPolicy).toEqual({
      supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp', 'application/pdf'],
      maxBytes: 25 * 1024 * 1024
    });
  });

  it('normalizes required runtime values', () => {
    expect(
      parseRuntimeConfig({
        apiBaseUrl: 'http://localhost:8080/',
        oidcIssuer: 'http://localhost:5556/dex/',
        oidcClientId: 'stuff-stash-web-local',
        oidcRedirectUri: 'http://localhost:5173/callback'
      })
    ).toEqual({
      apiBaseUrl: 'http://localhost:8080',
      oidcIssuer: 'http://localhost:5556/dex',
      oidcClientId: 'stuff-stash-web-local',
      oidcRedirectUri: 'http://localhost:5173/callback',
      invitationAllowInsecureLocalHTTP: false,
      mediaUploadPolicy: {
        supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp', 'application/pdf'],
        maxBytes: 25 * 1024 * 1024
      }
    });
  });

  it('defaults insecure local invitation HTTP off and accepts only an explicit boolean', () => {
    const base = {
      apiBaseUrl: 'http://localhost:8080',
      oidcIssuer: 'http://localhost:5556/dex',
      oidcClientId: 'stuff-stash-web-local',
      oidcRedirectUri: 'http://localhost:5173/callback'
    };
    expect(parseRuntimeConfig(base).invitationAllowInsecureLocalHTTP).toBe(false);
    expect(parseRuntimeConfig({ ...base, invitationAllowInsecureLocalHTTP: true }).invitationAllowInsecureLocalHTTP).toBe(true);
    expect(parseRuntimeConfig({ ...base, invitationAllowInsecureLocalHTTP: 'true' }).invitationAllowInsecureLocalHTTP).toBe(false);
  });

  it('accepts runtime media upload policy overrides', () => {
    expect(
      parseRuntimeConfig({
        apiBaseUrl: 'http://localhost:8080',
        oidcIssuer: 'http://localhost:5556/dex',
        oidcClientId: 'stuff-stash-web-local',
        oidcRedirectUri: 'http://localhost:5173/callback',
        mediaUploadPolicy: {
          supportedContentTypes: ['image/png'],
          maxBytes: 1234
        }
      }).mediaUploadPolicy
    ).toEqual({
      supportedContentTypes: ['image/png'],
      maxBytes: 1234
    });
  });

  it('rejects missing values', () => {
    expect(() => parseRuntimeConfig({ apiBaseUrl: 'http://localhost:8080' })).toThrow(
      'Missing web runtime configuration value'
    );
  });

  it('derives LAN development endpoints from an optional web origin override', () => {
    const config = parseRuntimeConfig({
      apiBaseUrl: 'http://localhost:8080',
      oidcIssuer: 'http://localhost:5556/dex',
      oidcClientId: 'stuff-stash-web-local',
      oidcRedirectUri: 'http://localhost:5173/callback'
    });

    expect(applyRuntimeConfigOverrides(config, { VITE_STUFF_STASH_WEB_ORIGIN: 'http://192.168.1.50:5173/' })).toMatchObject({
      apiBaseUrl: 'http://192.168.1.50:8080',
      oidcIssuer: 'http://192.168.1.50:5556/dex',
      oidcRedirectUri: 'http://192.168.1.50:5173/callback'
    });
  });

  it('allows explicit runtime config env overrides to win over the derived LAN origin', () => {
    const config = parseRuntimeConfig({
      apiBaseUrl: 'http://localhost:8080',
      oidcIssuer: 'http://localhost:5556/dex',
      oidcClientId: 'stuff-stash-web-local',
      oidcRedirectUri: 'http://localhost:5173/callback'
    });

    expect(
      applyRuntimeConfigOverrides(config, {
        VITE_STUFF_STASH_WEB_ORIGIN: 'http://192.168.1.50:5173',
        VITE_STUFF_STASH_API_BASE_URL: 'http://api.lan:18080/',
        VITE_STUFF_STASH_OIDC_ISSUER: 'http://dex.lan:15556/dex/',
        VITE_STUFF_STASH_OIDC_REDIRECT_URI: 'http://web.lan:5173/callback',
        VITE_STUFF_STASH_OIDC_CLIENT_ID: 'custom-web-client'
      })
    ).toMatchObject({
      apiBaseUrl: 'http://api.lan:18080',
      oidcIssuer: 'http://dex.lan:15556/dex',
      oidcClientId: 'custom-web-client',
      oidcRedirectUri: 'http://web.lan:5173/callback'
    });
  });
});
