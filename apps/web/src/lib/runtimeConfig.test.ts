import { describe, expect, it } from 'vitest';
import { parseRuntimeConfig } from './runtimeConfig';

describe('parseRuntimeConfig', () => {
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
      mediaUploadPolicy: {
        supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'],
        maxBytes: 5242880
      }
    });
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
});
