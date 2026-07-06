import { describe, expect, it } from 'vitest';
import { MobileAuthSession } from '../../application/auth/MobileAuthSession';
import { ExpoSecureAuthSessionStore, parseStoredSession, SecureStoreModule } from './ExpoSecureAuthSessionStore';

describe('ExpoSecureAuthSessionStore', () => {
  it('persists auth sessions only through the secure store module', async () => {
    const secureStore = new FakeSecureStore();
    const store = new ExpoSecureAuthSessionStore(secureStore);
    const session = authSession();

    await store.save(session);

    expect(secureStore.rawValue()).toContain('id-token');
    expect(secureStore.rawValue()).toContain('refresh-token');
    await expect(store.load()).resolves.toEqual(session);
  });

  it('clears secure session state', async () => {
    const secureStore = new FakeSecureStore(JSON.stringify(authSession()));
    const store = new ExpoSecureAuthSessionStore(secureStore);

    await store.clear();

    await expect(store.load()).resolves.toBeUndefined();
  });

  it('ignores malformed stored session material', () => {
    expect(parseStoredSession('not-json')).toBeUndefined();
    expect(parseStoredSession(JSON.stringify({ idToken: 'token' }))).toBeUndefined();
    expect(parseStoredSession(JSON.stringify({ ...authSession(), expiresAt: 0 }))).toBeUndefined();
    expect(parseStoredSession(JSON.stringify({ ...authSession(), expiresAt: -1 }))).toBeUndefined();
    expect(parseStoredSession(JSON.stringify({ ...authSession(), expiresAt: Number.NaN }))).toBeUndefined();
  });
});

function authSession(): MobileAuthSession {
  return {
    apiBaseUrl: 'https://api.example.test',
    issuer: 'https://accounts.example.test',
    clientId: 'stuff-stash-mobile',
    idToken: 'id-token',
    accessToken: 'access-token',
    refreshToken: 'refresh-token',
    expiresAt: 12_345
  };
}

class FakeSecureStore implements SecureStoreModule {
  readonly WHEN_UNLOCKED_THIS_DEVICE_ONLY = 1;

  constructor(private value: string | null = null) {}

  async getItemAsync(): Promise<string | null> {
    return this.value;
  }

  async setItemAsync(_key: string, value: string): Promise<void> {
    this.value = value;
  }

  async deleteItemAsync(): Promise<void> {
    this.value = null;
  }

  rawValue(): string {
    return this.value ?? '';
  }
}
