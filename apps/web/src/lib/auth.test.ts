import { describe, expect, it } from 'vitest';
import { completeSignIn, getStoredSession, sha256URLSafe, signOut } from './auth';
import type { RuntimeConfig } from './runtimeConfig';

const config: RuntimeConfig = {
  apiBaseUrl: 'http://localhost:8080',
  oidcIssuer: 'http://localhost:5556/dex',
  oidcClientId: 'stuff-stash-web-local',
  oidcRedirectUri: 'http://localhost:5173/callback',
  mediaUploadPolicy: {
    supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'],
    maxBytes: 5242880
  }
};

describe('auth helpers', () => {
  it('clears expired sessions', () => {
    const storage = new MapStorage();
    storage.setItem('stuffstash.oidc.session', JSON.stringify({ idToken: 'token', expiresAt: Date.now() - 1 }));

    expect(getStoredSession(storage)).toBeNull();
    expect(storage.getItem('stuffstash.oidc.session')).toBeNull();
  });

  it('exchanges callback codes and stores the id token', async () => {
    const storage = new MapStorage();
    storage.setItem('stuffstash.oidc.state', 'state-one');
    storage.setItem('stuffstash.oidc.verifier', 'verifier-one');
    storage.setItem('stuffstash.oidc.returnTo', '/');

    const returnTo = await completeSignIn(
      config,
      'http://localhost:5173/callback?code=code-one&state=state-one',
      async () => Response.json({ id_token: 'id-token', expires_in: 60 }),
      storage
    );

    expect(returnTo).toBe('/');
    expect(getStoredSession(storage)?.idToken).toBe('id-token');
  });

  it('creates a PKCE challenge when Web Crypto digest is unavailable', async () => {
    const originalSubtle = crypto.subtle;
    Object.defineProperty(crypto, 'subtle', { configurable: true, value: undefined });
    try {
      await expect(sha256URLSafe('dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk')).resolves.toBe(
        'E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM'
      );
    } finally {
      Object.defineProperty(crypto, 'subtle', { configurable: true, value: originalSubtle });
    }
  });

  it('removes stored auth state on sign out', () => {
    const storage = new MapStorage();
    storage.setItem('stuffstash.oidc.session', JSON.stringify({ idToken: 'token', expiresAt: Date.now() + 1000 }));
    storage.setItem('stuffstash.selectedTenantId', 'tenant-one');
    storage.setItem('stuffstash.selectedInventoryId', 'inventory-one');

    signOut(storage);

    expect(getStoredSession(storage)).toBeNull();
    expect(storage.getItem('stuffstash.selectedTenantId')).toBeNull();
    expect(storage.getItem('stuffstash.selectedInventoryId')).toBeNull();
  });
});

class MapStorage implements Storage {
  private readonly values = new Map<string, string>();

  get length(): number {
    return this.values.size;
  }

  clear(): void {
    this.values.clear();
  }

  getItem(key: string): string | null {
    return this.values.get(key) ?? null;
  }

  key(index: number): string | null {
    return Array.from(this.values.keys())[index] ?? null;
  }

  removeItem(key: string): void {
    this.values.delete(key);
  }

  setItem(key: string, value: string): void {
    this.values.set(key, value);
  }
}
