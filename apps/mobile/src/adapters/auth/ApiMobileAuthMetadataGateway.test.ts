import { describe, expect, it } from 'vitest';
import { ApiMobileAuthMetadataGateway, parseMobileAuthMetadata } from './ApiMobileAuthMetadataGateway';

describe('ApiMobileAuthMetadataGateway', () => {
  it('loads provider-neutral mobile OIDC metadata from the Stuff Stash instance', async () => {
    const requests: string[] = [];
    const gateway = new ApiMobileAuthMetadataGateway(async (input) => {
      requests.push(String(input));
      return Response.json({
        data: {
          issuer: 'https://accounts.example.test/',
          clientId: 'stuff-stash-mobile',
          redirectUri: 'stuffstash://auth/callback',
          scopes: ['openid', 'email', 'profile', 'offline_access', 'email']
        },
        meta: {}
      });
    });

    await expect(gateway.load('https://api.example.test/')).resolves.toEqual({
      issuer: 'https://accounts.example.test',
      clientId: 'stuff-stash-mobile',
      redirectUri: 'stuffstash://auth/callback',
      scopes: ['openid', 'email', 'profile', 'offline_access']
    });
    expect(requests).toEqual(['https://api.example.test/.well-known/stuff-stash/mobile-auth']);
  });

  it('rejects unavailable mobile auth metadata safely', async () => {
    const gateway = new ApiMobileAuthMetadataGateway(async () =>
      Response.json({ error: { code: 'mobile_auth_unavailable' }, meta: {} }, { status: 503 })
    );

    await expect(gateway.load('https://api.example.test')).rejects.toThrow('mobile sign-in is not configured');
  });

  it('rejects malformed metadata before sign-in starts', () => {
    expect(() => parseMobileAuthMetadata({ data: { issuer: '', clientId: 'client', redirectUri: 'stuffstash://auth/callback', scopes: [] } })).toThrow(
      'issuer'
    );
    expect(() => parseMobileAuthMetadata({ data: { issuer: 'https://issuer', clientId: 'client', redirectUri: 'stuffstash://auth/callback', scopes: ['openid', 1] } })).toThrow(
      'invalid scopes'
    );
  });

  it('rejects redirect URIs that do not match the app-owned callback', () => {
    const metadata = {
      issuer: 'https://issuer.example.test',
      clientId: 'client',
      scopes: ['openid']
    };

    expect(() => parseMobileAuthMetadata({ data: { ...metadata, redirectUri: 'https://evil.example.test/callback' } })).toThrow(
      'unsupported redirect URI'
    );
    expect(() => parseMobileAuthMetadata({ data: { ...metadata, redirectUri: 'stuffstash://wrong/callback' } })).toThrow(
      'unsupported redirect URI'
    );
    expect(() => parseMobileAuthMetadata({ data: { ...metadata, redirectUri: 'stuffstash://auth/wrong' } })).toThrow(
      'unsupported redirect URI'
    );
  });
});
