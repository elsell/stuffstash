import { describe, expect, it } from 'vitest';
import { MobileAuthMetadata } from '../../application/auth/MobileAuthSession';
import {
  ExpoOidcFacade,
  ExpoOidcNativeClientCore,
  NativeAccessTokenRequestConfig,
  NativeAuthRequestConfig,
  NativeRefreshTokenRequestConfig,
  NativeTokenResponse
} from './ExpoOidcNativeClientCore';

const metadata: MobileAuthMetadata = {
  issuer: 'https://accounts.example.test',
  clientId: 'stuff-stash-mobile',
  redirectUri: 'stuffstash://auth/callback',
  scopes: ['openid', 'email', 'profile', 'offline_access']
};

describe('ExpoOidcNativeClient', () => {
  it('uses authorization code with PKCE and validates callback state before token exchange', async () => {
    const facade = new FakeExpoOidcFacade();
    facade.promptResult = { type: 'success', params: { code: 'auth-code', state: 'state-1' } };
    facade.exchangeToken = {
      idToken: expiringIDToken(3_601),
      accessToken: 'access-token-1',
      refreshToken: 'refresh-token-1',
      expiresIn: 3600,
      issuedAt: 1
    };
    const client = new ExpoOidcNativeClientCore(facade);

    await expect(client.signIn(metadata)).resolves.toEqual({
      idToken: expiringIDToken(3_601),
      accessToken: 'access-token-1',
      refreshToken: 'refresh-token-1',
      expiresAt: 3_601_000
    });

    expect(facade.completedAuthSessions).toBe(1);
    expect(facade.discoveryIssuers).toEqual(['https://accounts.example.test']);
    expect(facade.authRequestConfig).toMatchObject({
      clientId: 'stuff-stash-mobile',
      redirectUri: 'stuffstash://auth/callback',
      responseType: 'code',
      scopes: ['openid', 'email', 'profile', 'offline_access'],
      usePKCE: true
    });
    expect(facade.exchangeConfig).toMatchObject({
      clientId: 'stuff-stash-mobile',
      code: 'auth-code',
      redirectUri: 'stuffstash://auth/callback',
      scopes: ['openid', 'email', 'profile', 'offline_access'],
      extraParams: { code_verifier: 'verifier-1' }
    });
  });

  it('rejects sign-in callbacks with the wrong state before exchanging a code', async () => {
    const facade = new FakeExpoOidcFacade();
    facade.promptResult = { type: 'success', params: { code: 'auth-code', state: 'attacker-state' } };
    const client = new ExpoOidcNativeClientCore(facade);

    await expect(client.signIn(metadata)).rejects.toThrow('Invalid sign-in callback.');
    expect(facade.exchangeConfig).toBeUndefined();
  });

  it('rejects cancelled sign-in without exchanging a code', async () => {
    const facade = new FakeExpoOidcFacade();
    facade.promptResult = { type: 'cancel', params: {} };
    const client = new ExpoOidcNativeClientCore(facade);

    await expect(client.signIn(metadata)).rejects.toThrow('Sign-in was cancelled.');
    expect(facade.exchangeConfig).toBeUndefined();
  });

  it('rejects sign-in callbacks when PKCE verifier state is missing', async () => {
    const facade = new FakeExpoOidcFacade();
    facade.codeVerifier = undefined;
    facade.promptResult = { type: 'success', params: { code: 'auth-code', state: 'state-1' } };
    const client = new ExpoOidcNativeClientCore(facade);

    await expect(client.signIn(metadata)).rejects.toThrow('usable authorization code');
    expect(facade.exchangeConfig).toBeUndefined();
  });

  it('rejects sign-in callbacks when request and callback state are both missing', async () => {
    const facade = new FakeExpoOidcFacade();
    facade.state = undefined;
    facade.promptResult = { type: 'success', params: { code: 'auth-code' } };
    const client = new ExpoOidcNativeClientCore(facade);

    await expect(client.signIn(metadata)).rejects.toThrow('Invalid sign-in callback.');
    expect(facade.exchangeConfig).toBeUndefined();
  });

  it('refreshes with the discovered provider and configured mobile client', async () => {
    const facade = new FakeExpoOidcFacade();
    facade.refreshTokenResult = {
      idToken: expiringIDToken(4),
      accessToken: 'access-token-2',
      refreshToken: 'refresh-token-2',
      issuedAt: 1
    };
    const client = new ExpoOidcNativeClientCore(facade);

    await expect(client.refresh(metadata, 'refresh-token-1')).resolves.toEqual({
      idToken: expiringIDToken(4),
      accessToken: 'access-token-2',
      refreshToken: 'refresh-token-2',
      expiresAt: 4_000
    });
    expect(facade.refreshConfig).toMatchObject({
      clientId: 'stuff-stash-mobile',
      refreshToken: 'refresh-token-1',
      scopes: ['openid', 'email', 'profile', 'offline_access']
    });
  });

  it('uses the earlier ID-token expiry when expiresIn lasts longer than the ID token', async () => {
    const facade = new FakeExpoOidcFacade();
    facade.exchangeToken = {
      idToken: expiringIDToken(20),
      refreshToken: 'refresh-token-1',
      expiresIn: 3600,
      issuedAt: 1
    };
    facade.promptResult = { type: 'success', params: { code: 'auth-code', state: 'state-1' } };
    const client = new ExpoOidcNativeClientCore(facade);

    await expect(client.signIn(metadata)).resolves.toMatchObject({
      expiresAt: 20_000
    });
  });

  it('rejects token responses without a usable ID token expiry', async () => {
    const facade = new FakeExpoOidcFacade();
    facade.exchangeToken = { idToken: 'bad-token', issuedAt: 1 };
    facade.promptResult = { type: 'success', params: { code: 'auth-code', state: 'state-1' } };
    const client = new ExpoOidcNativeClientCore(facade);

    await expect(client.signIn(metadata)).rejects.toThrow('usable ID token');
  });

  it('rejects ID tokens from a different issuer', async () => {
    const facade = new FakeExpoOidcFacade();
    facade.exchangeToken = {
      idToken: idToken({ iss: 'https://evil.example.test', aud: 'stuff-stash-mobile', exp: 3601 }),
      expiresIn: 3600,
      issuedAt: 1
    };
    facade.promptResult = { type: 'success', params: { code: 'auth-code', state: 'state-1' } };
    const client = new ExpoOidcNativeClientCore(facade);

    await expect(client.signIn(metadata)).rejects.toThrow('usable ID token');
  });

  it('rejects ID tokens for a different audience', async () => {
    const facade = new FakeExpoOidcFacade();
    facade.exchangeToken = {
      idToken: idToken({ iss: 'https://accounts.example.test', aud: 'other-client', exp: 3601 }),
      expiresIn: 3600,
      issuedAt: 1
    };
    facade.promptResult = { type: 'success', params: { code: 'auth-code', state: 'state-1' } };
    const client = new ExpoOidcNativeClientCore(facade);

    await expect(client.signIn(metadata)).rejects.toThrow('usable ID token');
  });

  it('accepts ID tokens whose audience list includes the mobile client', async () => {
    const facade = new FakeExpoOidcFacade();
    facade.exchangeToken = {
      idToken: idToken({
        iss: 'https://accounts.example.test',
        aud: ['other-client', 'stuff-stash-mobile'],
        exp: 3601
      }),
      expiresIn: 3600,
      issuedAt: 1
    };
    facade.promptResult = { type: 'success', params: { code: 'auth-code', state: 'state-1' } };
    const client = new ExpoOidcNativeClientCore(facade);

    await expect(client.signIn(metadata)).resolves.toMatchObject({
      expiresAt: 3_601_000
    });
  });

  it('rejects non-finite token expiry values', async () => {
    const facade = new FakeExpoOidcFacade();
    facade.exchangeToken = { idToken: 'id-token', expiresIn: 3600, issuedAt: Number.POSITIVE_INFINITY };
    facade.promptResult = { type: 'success', params: { code: 'auth-code', state: 'state-1' } };
    const client = new ExpoOidcNativeClientCore(facade);

    await expect(client.signIn(metadata)).rejects.toThrow('usable ID token');
  });

  it('rejects ID-token exp values that are not positive', async () => {
    const facade = new FakeExpoOidcFacade();
    facade.exchangeToken = { idToken: expiringIDToken(0), issuedAt: 1 };
    facade.promptResult = { type: 'success', params: { code: 'auth-code', state: 'state-1' } };
    const client = new ExpoOidcNativeClientCore(facade);

    await expect(client.signIn(metadata)).rejects.toThrow('usable ID token');
  });

  it('rejects refresh responses without a usable ID token', async () => {
    const facade = new FakeExpoOidcFacade();
    facade.refreshTokenResult = { idToken: undefined, expiresIn: 3600, issuedAt: 1 };
    const client = new ExpoOidcNativeClientCore(facade);

    await expect(client.refresh(metadata, 'refresh-token-1')).rejects.toThrow('usable ID token');
  });
});

class FakeExpoOidcFacade implements ExpoOidcFacade {
  completedAuthSessions = 0;
  discoveryIssuers: string[] = [];
  state: string | undefined = 'state-1';
  codeVerifier: string | undefined = 'verifier-1';
  promptResult = { type: 'cancel', params: {} };
  exchangeToken: NativeTokenResponse = { idToken: expiringIDToken(3601), expiresIn: 3600, issuedAt: 1 };
  refreshTokenResult: NativeTokenResponse = { idToken: expiringIDToken(3601), expiresIn: 3600, issuedAt: 1 };
  authRequestConfig: NativeAuthRequestConfig | undefined;
  exchangeConfig: NativeAccessTokenRequestConfig | undefined;
  refreshConfig: NativeRefreshTokenRequestConfig | undefined;

  completeAuthSession(): void {
    this.completedAuthSessions += 1;
  }

  async fetchDiscovery(issuer: string): Promise<unknown> {
    this.discoveryIssuers.push(issuer);
    return { tokenEndpoint: 'https://accounts.example.test/token' };
  }

  createAuthRequest(config: NativeAuthRequestConfig) {
    this.authRequestConfig = config;
    return {
      state: this.state,
      codeVerifier: this.codeVerifier,
      promptAsync: async () => this.promptResult
    };
  }

  async exchangeCode(config: NativeAccessTokenRequestConfig): Promise<NativeTokenResponse> {
    this.exchangeConfig = config;
    return this.exchangeToken;
  }

  async refresh(config: NativeRefreshTokenRequestConfig): Promise<NativeTokenResponse> {
    this.refreshConfig = config;
    return this.refreshTokenResult;
  }
}

function expiringIDToken(exp: number): string {
  return idToken({ iss: 'https://accounts.example.test', aud: 'stuff-stash-mobile', exp });
}

function idToken(claims: Record<string, unknown>): string {
  return `header.${base64UrlEncode(JSON.stringify(claims))}.signature`;
}

function base64UrlEncode(value: string): string {
  return globalThis.btoa(value).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}
