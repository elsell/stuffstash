import { describe, expect, it } from 'vitest';
import { ExpoOidcNativeClientCore, type ExpoOidcFacade, type NativeAuthPromptResult } from '../../adapters/auth/ExpoOidcNativeClientCore';
import { MobileAuthSessionController, type MobileAuthSession, type MobileAuthSessionStore } from '../auth/MobileAuthSession';
import { OnboardingCommand } from './OnboardingCommand';
import { OnboardingProfileFake, onboardingServer } from './OnboardingTestSupport';

describe('combined onboarding through native OIDC boundary', () => {
  it.each([
    ['cancel', { type: 'cancel' }],
    ['forged state', { type: 'success', params: { state: 'forged', code: 'code' } }],
    ['missing code', { type: 'success', params: { state: 'expected' } }]
  ] as const)('rejects %s without creating a session or discovering tenants', async (_name, result) => {
    const f = fixture(result);
    await expect(f.command.connectAndSignIn({ apiBaseUrl: onboardingServer })).rejects.toThrow();
    expect(await f.store.load()).toBeUndefined();
    expect(f.discoveryCalls()).toBe(0);
  });
  it('accepts a matching PKCE callback and discovers authorized setup', async () => {
    const f = fixture({ type: 'success', params: { state: 'expected', code: 'code' } });
    await expect(f.command.connectAndSignIn({ apiBaseUrl: onboardingServer })).resolves.toMatchObject({ step: 'tenant' });
    expect((await f.store.load())?.apiBaseUrl).toBe(onboardingServer);
    expect(f.discoveryCalls()).toBe(1);
  });
});

function fixture(result: NativeAuthPromptResult) {
  let session: MobileAuthSession | undefined;
  const store: MobileAuthSessionStore = {
    async load() { return session; }, async save(value) { session = value; }, async clear() { session = undefined; }
  };
  const facade: ExpoOidcFacade = {
    completeAuthSession() {}, async fetchDiscovery() { return {}; },
    createAuthRequest(config) {
      expect(config.usePKCE).toBe(true);
      return { state: 'expected', codeVerifier: 'verifier', async promptAsync() { return result; } };
    },
    async exchangeCode(config) {
      expect(config.extraParams.code_verifier).toBe('verifier');
      return { idToken: `test.${btoa(JSON.stringify({ iss: 'https://issuer.example.test', aud: 'mobile', exp: 4600 }))}.test`, refreshToken: 'test-refresh-token', expiresIn: 3600, issuedAt: 1_000 };
    },
    async refresh() { throw new Error('Not expected'); }
  };
  const auth = new MobileAuthSessionController(store, { async load() { return {
    issuer: 'https://issuer.example.test', clientId: 'mobile', redirectUri: 'stuffstash://auth/callback', scopes: ['openid', 'offline_access']
  }; } }, new ExpoOidcNativeClientCore(facade), () => 1_000);
  let discoveries = 0;
  const command = new OnboardingCommand(new OnboardingProfileFake(), () => ({
    async listTenants() { discoveries++; if (!session) throw new Error('Unauthenticated'); return []; },
    async listInventories() { return []; },
    async createTenant() { throw new Error('Not expected'); },
    async createInventory() { throw new Error('Not expected'); }
  }), auth);
  return { command, store, discoveryCalls: () => discoveries };
}
