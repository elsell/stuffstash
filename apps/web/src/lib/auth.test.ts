import { describe, expect, it } from 'vitest';
import { completeSignIn, getStoredSession, hasRecentlyCompletedSignIn, sha256URLSafe, signOut, startSignIn } from './auth';
import type { RuntimeConfig } from './runtimeConfig';

const config: RuntimeConfig = {
  apiBaseUrl: 'http://localhost:8080',
  oidcIssuer: 'http://localhost:5556/dex',
  oidcClientId: 'stuff-stash-web-local',
  oidcRedirectUri: 'http://localhost:5173/callback',
  invitationAllowInsecureLocalHTTP: false,
  mediaUploadPolicy: {
    supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'],
    maxBytes: 5242880
  }
};

describe('auth helpers', () => {
  it('preserves an invitation fragment locally without sending it to the identity provider', async () => {
    const storage = new MapStorage();
    let assigned = '';
    const location = {
      pathname: '/invitations/accept',
      search: '?tenant=tenant-one&inventory=inventory-one&invitation=invite-one',
      hash: '#token=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA',
      assign: (value: string) => {
        assigned = value;
      }
    } as unknown as Location;
    const replaced: string[] = [];

    await startSignIn(config, location, storage, {
      state: null,
      replaceState: (_state, _unused, url) => replaced.push(String(url))
    });

    expect(storage.getItem('stuffstash.oidc.returnTo')).toBe(
      '/invitations/accept?tenant=tenant-one&inventory=inventory-one&invitation=invite-one#token=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA'
    );
    expect(assigned).toContain('http://localhost:5556/dex/auth?');
    expect(assigned).not.toContain('AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA');
    expect(replaced).toEqual(['/invitations/accept?tenant=tenant-one&inventory=inventory-one&invitation=invite-one']);
  });

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
    expect(hasRecentlyCompletedSignIn(storage)).toBe(true);
  });

  it.each([
    'https://evil.example/steal',
    '//evil.example/steal',
    '/\\evil.example/steal',
    '/%2f%2fevil.example/steal',
    '/callback?code=loop',
    '/callback/?code=loop',
    '/safe\u0000unsafe'
  ])('rejects unsafe stored return destinations: %s', async (unsafeReturnTo) => {
    const storage = callbackStorage(unsafeReturnTo);
    await expect(
      completeSignIn(
        config,
        'http://localhost:5173/callback?code=code-one&state=state-one',
        async () => Response.json({ id_token: 'id-token', expires_in: 60 }),
        storage
      )
    ).resolves.toBe('/');
    expect(storage.getItem('stuffstash.oidc.returnTo')).toBeNull();
  });

  it('returns to a validated invitation URL including its fragment', async () => {
    const invitationPath = '/invitations/accept?tenant=tenant-one&inventory=inventory-one&invitation=invite-one#token=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA';
    const storage = callbackStorage(invitationPath);
    await expect(
      completeSignIn(
        config,
        'http://localhost:5173/callback?code=code-one&state=state-one',
        async () => Response.json({ id_token: 'id-token', expires_in: 60 }),
        storage
      )
    ).resolves.toBe(invitationPath);
  });

  it('clears one-time callback state when token exchange fails', async () => {
    const storage = callbackStorage('/invitations/accept#token=secret');
    await expect(
      completeSignIn(
        config,
        'http://localhost:5173/callback?code=code-one&state=state-one',
        async () => new Response(null, { status: 500 }),
        storage
      )
    ).rejects.toThrow('Unable to complete sign-in.');
    expect(storage.getItem('stuffstash.oidc.returnTo')).toBeNull();
    expect(storage.getItem('stuffstash.oidc.verifier')).toBeNull();
    expect(storage.getItem('stuffstash.oidc.state')).toBeNull();
  });

  it('remembers only recent completed sign-ins', () => {
    const storage = new MapStorage();
    storage.setItem('stuffstash.oidc.completedAt', String(1000));

    expect(hasRecentlyCompletedSignIn(storage, 1000 + 60_000)).toBe(true);
    expect(hasRecentlyCompletedSignIn(storage, 1000 + 180_000)).toBe(false);
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
    storage.setItem('stuffstash.oidc.returnTo', '/invitations/accept#token=secret');
    storage.setItem('stuffstash.selectedTenantId', 'tenant-one');
    storage.setItem('stuffstash.selectedInventoryId', 'inventory-one');

    signOut(storage);

    expect(getStoredSession(storage)).toBeNull();
    expect(storage.getItem('stuffstash.oidc.returnTo')).toBeNull();
    expect(storage.getItem('stuffstash.selectedTenantId')).toBeNull();
    expect(storage.getItem('stuffstash.selectedInventoryId')).toBeNull();
  });
});

function callbackStorage(returnTo: string): MapStorage {
  const storage = new MapStorage();
  storage.setItem('stuffstash.oidc.state', 'state-one');
  storage.setItem('stuffstash.oidc.verifier', 'verifier-one');
  storage.setItem('stuffstash.oidc.returnTo', returnTo);
  return storage;
}

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
