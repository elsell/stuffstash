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
      oidcRedirectUri: 'http://localhost:5173/callback'
    });
  });

  it('rejects missing values', () => {
    expect(() => parseRuntimeConfig({ apiBaseUrl: 'http://localhost:8080' })).toThrow(
      'Missing web runtime configuration value'
    );
  });
});
