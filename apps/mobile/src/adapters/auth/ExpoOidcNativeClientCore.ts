import {
  MobileAuthMetadata,
  MobileAuthTokenResult,
  NativeOidcClient
} from '../../application/auth/MobileAuthSession';

export type NativeAuthRequestConfig = {
  readonly clientId: string;
  readonly redirectUri: string;
  readonly responseType: 'code';
  readonly scopes: readonly string[];
  readonly usePKCE: true;
};

export type NativeAccessTokenRequestConfig = {
  readonly clientId: string;
  readonly code: string;
  readonly redirectUri: string;
  readonly scopes: readonly string[];
  readonly extraParams: {
    readonly code_verifier: string;
  };
};

export type NativeRefreshTokenRequestConfig = {
  readonly clientId: string;
  readonly refreshToken: string;
  readonly scopes: readonly string[];
};

export type NativeDiscoveryDocument = unknown;

export type NativeAuthRequest = {
  readonly state?: string;
  readonly codeVerifier?: string;
  promptAsync(discovery: NativeDiscoveryDocument): Promise<NativeAuthPromptResult>;
};

export type NativeAuthPromptResult = {
  readonly type: string;
  readonly params?: Record<string, string | undefined>;
};

export type NativeTokenResponse = {
  readonly idToken?: string;
  readonly accessToken?: string;
  readonly refreshToken?: string;
  readonly expiresIn?: number;
  readonly issuedAt: number;
};

export type ExpoOidcFacade = {
  completeAuthSession(): void;
  fetchDiscovery(issuer: string): Promise<NativeDiscoveryDocument>;
  createAuthRequest(config: NativeAuthRequestConfig): NativeAuthRequest;
  exchangeCode(
    config: NativeAccessTokenRequestConfig,
    discovery: NativeDiscoveryDocument
  ): Promise<NativeTokenResponse>;
  refresh(
    config: NativeRefreshTokenRequestConfig,
    discovery: NativeDiscoveryDocument
  ): Promise<NativeTokenResponse>;
};

export class ExpoOidcNativeClientCore implements NativeOidcClient {
  constructor(private readonly expo: ExpoOidcFacade) {
    this.expo.completeAuthSession();
  }

  async signIn(metadata: MobileAuthMetadata): Promise<MobileAuthTokenResult> {
    const discovery = await this.expo.fetchDiscovery(metadata.issuer);
    const request = this.expo.createAuthRequest({
      clientId: metadata.clientId,
      redirectUri: metadata.redirectUri,
      responseType: 'code',
      scopes: [...metadata.scopes],
      usePKCE: true
    });

    const result = await request.promptAsync(discovery);
    if (result.type !== 'success') {
      throw new Error('Sign-in was cancelled.');
    }
    const params = result.params ?? {};
    if (!request.state || !params.state || params.state !== request.state) {
      throw new Error('Invalid sign-in callback.');
    }
    if (!params.code || !request.codeVerifier) {
      throw new Error('Sign-in response did not include a usable authorization code.');
    }

    const token = await this.expo.exchangeCode(
      {
        clientId: metadata.clientId,
        code: params.code,
        redirectUri: metadata.redirectUri,
        scopes: [...metadata.scopes],
        extraParams: {
          code_verifier: request.codeVerifier
        }
      },
      discovery
    );

    return tokenResult(token, metadata);
  }

  async refresh(metadata: MobileAuthMetadata, refreshToken: string): Promise<MobileAuthTokenResult> {
    const discovery = await this.expo.fetchDiscovery(metadata.issuer);
    const token = await this.expo.refresh(
      {
        clientId: metadata.clientId,
        refreshToken,
        scopes: [...metadata.scopes]
      },
      discovery
    );

    return tokenResult(token, metadata);
  }
}

function tokenResult(token: NativeTokenResponse, metadata: MobileAuthMetadata): MobileAuthTokenResult {
  const claims = idTokenClaims(token.idToken);
  const expiresAt = tokenExpiresAt(token, claims);
  if (
    !token.idToken ||
    !claims ||
    claims.iss !== metadata.issuer ||
    !audienceIncludes(claims.aud, metadata.clientId) ||
    expiresAt === undefined ||
    !Number.isFinite(expiresAt) ||
    expiresAt <= 0
  ) {
    throw new Error('Sign-in response did not include a usable ID token.');
  }

  return {
    idToken: token.idToken,
    accessToken: token.accessToken,
    refreshToken: token.refreshToken,
    expiresAt
  };
}

function tokenExpiresAt(token: NativeTokenResponse, claims: IDTokenClaims | undefined): number | undefined {
  if (typeof claims?.exp !== 'number' || !Number.isFinite(claims.exp)) {
    return undefined;
  }
  const idTokenExpiresAt = claims.exp * 1000;
  if (
    typeof token.issuedAt !== 'number' ||
    !Number.isFinite(token.issuedAt) ||
    typeof token.expiresIn !== 'number' ||
    !Number.isFinite(token.expiresIn) ||
    token.expiresIn <= 0
  ) {
    return idTokenExpiresAt;
  }

  return Math.min(idTokenExpiresAt, (token.issuedAt + token.expiresIn) * 1000);
}

type IDTokenClaims = {
  readonly iss?: unknown;
  readonly aud?: unknown;
  readonly exp?: unknown;
};

function idTokenClaims(idToken?: string): IDTokenClaims | undefined {
  const payload = idToken?.split('.')[1];
  if (!payload) {
    return undefined;
  }

  try {
    const normalized = payload.replace(/-/g, '+').replace(/_/g, '/');
    const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=');
    const decoded = globalThis.atob(padded);
    return JSON.parse(decoded) as IDTokenClaims;
  } catch {
    return undefined;
  }
}

function audienceIncludes(audience: unknown, clientId: string): boolean {
  if (typeof audience === 'string') {
    return audience === clientId;
  }
  if (!Array.isArray(audience)) {
    return false;
  }
  return audience.some((value) => value === clientId);
}
