import { describe, expect, it } from 'vitest';
import {
  MobileAuthenticationRequiredError,
  MobileAuthMetadata,
  MobileAuthMetadataGateway,
  MobileAuthSession,
  MobileAuthSessionController,
  MobileAuthSessionStore,
  MobileAuthTokenResult,
  NativeOidcClient
} from './MobileAuthSession';

const metadata: MobileAuthMetadata = {
  issuer: 'https://accounts.example.test',
  clientId: 'stuff-stash-mobile',
  redirectUri: 'stuffstash://auth/callback',
  scopes: ['openid', 'email', 'profile', 'offline_access']
};

describe('MobileAuthSessionController', () => {
  it('signs in through discovered mobile OIDC metadata and stores the session securely', async () => {
    const store = new FakeSessionStore();
    const gateway = new FakeMetadataGateway(metadata);
    const oidc = new FakeOidcClient({
      idToken: 'id-token-1',
      accessToken: 'access-token-1',
      refreshToken: 'refresh-token-1',
      expiresAt: 2_000
    });
    const controller = new MobileAuthSessionController(store, gateway, oidc, () => 1_000);

    const session = await controller.signIn('https://api.example.test');

    expect(gateway.loadedApiBaseUrls).toEqual(['https://api.example.test']);
    expect(oidc.signInMetadata).toEqual([metadata]);
    expect(session).toMatchObject({
      apiBaseUrl: 'https://api.example.test',
      issuer: metadata.issuer,
      clientId: metadata.clientId,
      idToken: 'id-token-1',
      refreshToken: 'refresh-token-1'
    });
    expect(await store.load()).toEqual(session);
  });

  it('rejects sign-in when offline access metadata does not return a refresh token', async () => {
    const store = new FakeSessionStore();
    const oidc = new FakeOidcClient({
      idToken: 'id-token-1',
      accessToken: 'access-token-1',
      expiresAt: 2_000
    });
    const controller = new MobileAuthSessionController(store, new FakeMetadataGateway(metadata), oidc, () => 1_000);

    await expect(controller.signIn('https://api.example.test')).rejects.toBeInstanceOf(
      MobileAuthenticationRequiredError
    );
    expect(await store.load()).toBeUndefined();
  });

  it('returns the stored ID token while it is fresh', async () => {
    const store = new FakeSessionStore(freshSession());
    const oidc = new FakeOidcClient();
    const controller = new MobileAuthSessionController(store, new FakeMetadataGateway(metadata), oidc, () => 1_000);

    await expect(controller.validIdToken('https://api.example.test')).resolves.toBe('fresh-id-token');

    expect(oidc.refreshCalls).toBe(0);
  });

  it('reports a signed-in status with a valid session for the requested instance', async () => {
    const store = new FakeSessionStore(freshSession());
    const controller = new MobileAuthSessionController(store, new FakeMetadataGateway(metadata), new FakeOidcClient(), () => 1_000);

    await expect(controller.status('https://api.example.test')).resolves.toMatchObject({
      status: 'signed_in',
      session: { idToken: 'fresh-id-token' }
    });
  });

  it('reports signed out and clears storage when status cannot refresh an expired session', async () => {
    const store = new FakeSessionStore({ ...freshSession(), expiresAt: 1_020, refreshToken: undefined });
    const controller = new MobileAuthSessionController(store, new FakeMetadataGateway(metadata), new FakeOidcClient(), () => 1_000, 60);

    await expect(controller.status('https://api.example.test')).resolves.toEqual({ status: 'signed_out' });
    expect(await store.load()).toBeUndefined();
  });

  it('refreshes an expiring session before returning a token', async () => {
    const store = new FakeSessionStore({ ...freshSession(), expiresAt: 1_020 });
    const oidc = new FakeOidcClient({
      idToken: 'refreshed-id-token',
      accessToken: 'refreshed-access-token',
      expiresAt: 4_000
    });
    const controller = new MobileAuthSessionController(store, new FakeMetadataGateway(metadata), oidc, () => 1_000, 60);

    await expect(controller.validIdToken('https://api.example.test')).resolves.toBe('refreshed-id-token');

    expect(oidc.refreshCalls).toBe(1);
    expect(await store.load()).toMatchObject({
      idToken: 'refreshed-id-token',
      accessToken: 'refreshed-access-token',
      refreshToken: 'fresh-refresh-token',
      expiresAt: 4_000
    });
  });

  it('coalesces concurrent refresh requests', async () => {
    const store = new FakeSessionStore({ ...freshSession(), expiresAt: 1_020 });
    const oidc = new FakeOidcClient({
      idToken: 'refreshed-id-token',
      refreshToken: 'new-refresh-token',
      expiresAt: 4_000
    });
    oidc.refreshDelay = Promise.resolve();
    const controller = new MobileAuthSessionController(store, new FakeMetadataGateway(metadata), oidc, () => 1_000, 60);

    const tokens = await Promise.all([
      controller.validIdToken('https://api.example.test'),
      controller.validIdToken('https://api.example.test')
    ]);

    expect(tokens).toEqual(['refreshed-id-token', 'refreshed-id-token']);
    expect(oidc.refreshCalls).toBe(1);
  });

  it('does not restore an old session when sign-out wins a refresh race', async () => {
    const store = new FakeSessionStore({ ...freshSession(), expiresAt: 1_020 });
    const oidc = new FakeOidcClient({
      idToken: 'refreshed-id-token',
      refreshToken: 'new-refresh-token',
      expiresAt: 4_000
    });
    const refreshDelay = deferred<void>();
    oidc.refreshDelay = refreshDelay.promise;
    const controller = new MobileAuthSessionController(store, new FakeMetadataGateway(metadata), oidc, () => 1_000, 60);

    const refresh = controller.validIdToken('https://api.example.test');
    await controller.signOut();
    refreshDelay.resolve();

    await expect(refresh).rejects.toBeInstanceOf(MobileAuthenticationRequiredError);
    expect(await store.load()).toBeUndefined();
  });

  it('does not clear a new session when it wins an old refresh race', async () => {
    const store = new FakeSessionStore({ ...freshSession(), expiresAt: 1_020 });
    const oidc = new FakeOidcClient({
      idToken: 'refreshed-id-token',
      refreshToken: 'new-refresh-token',
      expiresAt: 4_000
    });
    const refreshDelay = deferred<void>();
    oidc.refreshDelay = refreshDelay.promise;
    const controller = new MobileAuthSessionController(store, new FakeMetadataGateway(metadata), oidc, () => 1_000, 60);
    const replacementSession = {
      ...freshSession(),
      idToken: 'replacement-id-token',
      refreshToken: 'replacement-refresh-token',
      expiresAt: 10_000
    };

    const refresh = controller.validIdToken('https://api.example.test');
    await store.save(replacementSession);
    refreshDelay.resolve();

    await expect(refresh).rejects.toBeInstanceOf(MobileAuthenticationRequiredError);
    expect(await store.load()).toEqual(replacementSession);
  });

  it('clears stale sessions when refresh fails', async () => {
    const store = new FakeSessionStore({ ...freshSession(), expiresAt: 1_020 });
    const oidc = new FakeOidcClient();
    oidc.refreshError = new Error('provider unavailable');
    const controller = new MobileAuthSessionController(store, new FakeMetadataGateway(metadata), oidc, () => 1_000, 60);

    await expect(controller.validIdToken('https://api.example.test')).rejects.toBeInstanceOf(
      MobileAuthenticationRequiredError
    );

    expect(await store.load()).toBeUndefined();
  });

  it('fails closed when a token is requested for a different instance', async () => {
    const store = new FakeSessionStore(freshSession());
    const controller = new MobileAuthSessionController(store, new FakeMetadataGateway(metadata), new FakeOidcClient(), () => 1_000);

    await expect(controller.validIdToken('https://other.example.test')).rejects.toBeInstanceOf(
      MobileAuthenticationRequiredError
    );
  });

  it('clears session state on sign-out', async () => {
    const store = new FakeSessionStore(freshSession());
    const controller = new MobileAuthSessionController(store, new FakeMetadataGateway(metadata), new FakeOidcClient());

    await controller.signOut();

    expect(await store.load()).toBeUndefined();
  });
});

function freshSession(): MobileAuthSession {
  return {
    apiBaseUrl: 'https://api.example.test',
    issuer: metadata.issuer,
    clientId: metadata.clientId,
    idToken: 'fresh-id-token',
    accessToken: 'fresh-access-token',
    refreshToken: 'fresh-refresh-token',
    expiresAt: 100_000
  };
}

function deferred<T>(): { readonly promise: Promise<T>; readonly resolve: (value: T) => void } {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((resolvePromise) => {
    resolve = resolvePromise;
  });
  return { promise, resolve };
}

class FakeSessionStore implements MobileAuthSessionStore {
  constructor(private session?: MobileAuthSession) {}

  async load(): Promise<MobileAuthSession | undefined> {
    return this.session;
  }

  async save(session: MobileAuthSession): Promise<void> {
    this.session = session;
  }

  async clear(): Promise<void> {
    this.session = undefined;
  }
}

class FakeMetadataGateway implements MobileAuthMetadataGateway {
  readonly loadedApiBaseUrls: string[] = [];

  constructor(private readonly metadata: MobileAuthMetadata) {}

  async load(apiBaseUrl: string): Promise<MobileAuthMetadata> {
    this.loadedApiBaseUrls.push(apiBaseUrl);
    return this.metadata;
  }
}

class FakeOidcClient implements NativeOidcClient {
  readonly signInMetadata: MobileAuthMetadata[] = [];
  refreshCalls = 0;
  refreshError: Error | undefined;
  refreshDelay: Promise<void> | undefined;

  constructor(private readonly result: MobileAuthTokenResult = { idToken: 'id-token', expiresAt: 2_000 }) {}

  async signIn(metadata: MobileAuthMetadata): Promise<MobileAuthTokenResult> {
    this.signInMetadata.push(metadata);
    return this.result;
  }

  async refresh(): Promise<MobileAuthTokenResult> {
    this.refreshCalls += 1;
    if (this.refreshDelay) {
      await this.refreshDelay;
    }
    if (this.refreshError) {
      throw this.refreshError;
    }
    return this.result;
  }
}
