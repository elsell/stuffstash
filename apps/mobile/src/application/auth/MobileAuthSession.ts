export type MobileAuthMetadata = {
  readonly issuer: string;
  readonly clientId: string;
  readonly redirectUri: string;
  readonly scopes: readonly string[];
};

export type MobileAuthSession = {
  readonly apiBaseUrl: string;
  readonly issuer: string;
  readonly clientId: string;
  readonly idToken: string;
  readonly accessToken?: string;
  readonly refreshToken?: string;
  readonly expiresAt: number;
};

export interface MobileAuthMetadataGateway {
  load(apiBaseUrl: string): Promise<MobileAuthMetadata>;
}

export interface NativeOidcClient {
  signIn(metadata: MobileAuthMetadata): Promise<MobileAuthTokenResult>;
  refresh(metadata: MobileAuthMetadata, refreshToken: string): Promise<MobileAuthTokenResult>;
}

export type MobileAuthTokenResult = {
  readonly idToken: string;
  readonly accessToken?: string;
  readonly refreshToken?: string;
  readonly expiresAt: number;
};

export interface MobileAuthSessionStore {
  load(): Promise<MobileAuthSession | undefined>;
  save(session: MobileAuthSession): Promise<void>;
  clear(): Promise<void>;
}

export type MobileAuthStatus =
  | { readonly status: 'signed_out' }
  | { readonly status: 'signed_in'; readonly session: MobileAuthSession };

export class MobileAuthenticationRequiredError extends Error {
  constructor(message = 'Sign in to Stuff Stash.') {
    super(message);
  }
}

export class MobileAuthSessionController {
  private refreshInFlight: { readonly key: string; readonly promise: Promise<MobileAuthSession> } | undefined;

  constructor(
    private readonly store: MobileAuthSessionStore,
    private readonly metadataGateway: MobileAuthMetadataGateway,
    private readonly oidcClient: NativeOidcClient,
    private readonly now: () => number = () => Date.now(),
    private readonly refreshSkewMs = 60_000
  ) {}

  async status(apiBaseUrl?: string): Promise<MobileAuthStatus> {
    const session = await this.store.load();
    if (!session) {
      return { status: 'signed_out' };
    }
    if (apiBaseUrl && session.apiBaseUrl !== apiBaseUrl) {
      return { status: 'signed_out' };
    }
    if (apiBaseUrl) {
      try {
        return { status: 'signed_in', session: await this.validSession(apiBaseUrl) };
      } catch (error) {
        if (error instanceof MobileAuthenticationRequiredError) {
          return { status: 'signed_out' };
        }
        throw error;
      }
    }
    return { status: 'signed_in', session };
  }

  async prepareSignIn(apiBaseUrl: string): Promise<void> {
    await this.metadataGateway.load(apiBaseUrl);
  }

  async signIn(apiBaseUrl: string): Promise<MobileAuthSession> {
    const metadata = await this.metadataGateway.load(apiBaseUrl);
    const tokens = await this.oidcClient.signIn(metadata);
    const session = sessionFromTokens(apiBaseUrl, metadata, tokens);
    await this.store.save(session);
    return session;
  }

  async signOut(): Promise<void> {
    this.refreshInFlight = undefined;
    await this.store.clear();
  }

  async validIdToken(apiBaseUrl: string): Promise<string> {
    const session = await this.validSession(apiBaseUrl);
    return session.idToken;
  }

  async validSession(apiBaseUrl: string): Promise<MobileAuthSession> {
    const session = await this.store.load();
    if (!session || session.apiBaseUrl !== apiBaseUrl || !session.idToken) {
      throw new MobileAuthenticationRequiredError();
    }
    if (!this.shouldRefresh(session)) {
      return session;
    }
    if (!session.refreshToken) {
      await this.store.clear();
      throw new MobileAuthenticationRequiredError('Sign in again to refresh your Stuff Stash session.');
    }
    const key = refreshKey(session);
    if (!this.refreshInFlight || this.refreshInFlight.key !== key) {
      const promise = this.refreshSession(session).finally(() => {
        if (this.refreshInFlight?.key === key) {
          this.refreshInFlight = undefined;
        }
      });
      this.refreshInFlight = { key, promise };
    }
    return this.refreshInFlight.promise;
  }

  private shouldRefresh(session: MobileAuthSession): boolean {
    return session.expiresAt - this.now() <= this.refreshSkewMs;
  }

  private async refreshSession(session: MobileAuthSession): Promise<MobileAuthSession> {
    if (!session.refreshToken) {
      await this.store.clear();
      throw new MobileAuthenticationRequiredError('Sign in again to refresh your Stuff Stash session.');
    }

    try {
      const metadata = await this.metadataGateway.load(session.apiBaseUrl);
      const tokens = await this.oidcClient.refresh(metadata, session.refreshToken);
      const refreshed = sessionFromTokens(session.apiBaseUrl, metadata, {
        ...tokens,
        refreshToken: tokens.refreshToken ?? session.refreshToken
      });
      const current = await this.store.load();
      if (!current || refreshKey(current) !== refreshKey(session)) {
        throw new RefreshSessionSupersededError();
      }
      await this.store.save(refreshed);
      return refreshed;
    } catch (error) {
      if (error instanceof RefreshSessionSupersededError) {
        throw new MobileAuthenticationRequiredError();
      }
      const current = await this.store.load();
      if (current && refreshKey(current) === refreshKey(session)) {
        await this.store.clear();
      }
      if (error instanceof MobileAuthenticationRequiredError) {
        throw error;
      }
      throw new MobileAuthenticationRequiredError('Sign in again to refresh your Stuff Stash session.');
    }
  }
}

class RefreshSessionSupersededError extends Error {}

function refreshKey(session: MobileAuthSession): string {
  return `${session.apiBaseUrl}\u0000${session.issuer}\u0000${session.clientId}\u0000${session.idToken}\u0000${session.refreshToken ?? ''}`;
}

function sessionFromTokens(
  apiBaseUrl: string,
  metadata: MobileAuthMetadata,
  tokens: MobileAuthTokenResult
): MobileAuthSession {
  if (!tokens.idToken || !Number.isFinite(tokens.expiresAt) || tokens.expiresAt <= 0) {
    throw new MobileAuthenticationRequiredError('Sign-in did not return a usable Stuff Stash session.');
  }
  if (requiresRefreshToken(metadata) && !tokens.refreshToken) {
    throw new MobileAuthenticationRequiredError('Sign-in did not return a refreshable Stuff Stash session.');
  }

  return {
    apiBaseUrl,
    issuer: metadata.issuer,
    clientId: metadata.clientId,
    idToken: tokens.idToken,
    accessToken: tokens.accessToken,
    refreshToken: tokens.refreshToken,
    expiresAt: tokens.expiresAt
  };
}

function requiresRefreshToken(metadata: MobileAuthMetadata): boolean {
  return metadata.scopes.some((scope) => scope === 'offline_access');
}
