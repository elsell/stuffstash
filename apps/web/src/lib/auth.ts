import type { RuntimeConfig } from './runtimeConfig';

const verifierKey = 'stuffstash.oidc.verifier';
const stateKey = 'stuffstash.oidc.state';
const returnToKey = 'stuffstash.oidc.returnTo';
const sessionKey = 'stuffstash.oidc.session';

export interface AuthSession {
  idToken: string;
  expiresAt: number;
}

export interface TokenResponse {
  id_token: string;
  expires_in?: number;
}

export function getStoredSession(storage: Storage = window.sessionStorage): AuthSession | null {
  const value = storage.getItem(sessionKey);
  if (!value) {
    return null;
  }
  const session = JSON.parse(value) as AuthSession;
  if (!session.idToken || Date.now() >= session.expiresAt) {
    storage.removeItem(sessionKey);
    return null;
  }
  return session;
}

export function storeSession(session: AuthSession, storage: Storage = window.sessionStorage): void {
  storage.setItem(sessionKey, JSON.stringify(session));
}

export function signOut(storage: Storage = window.sessionStorage): void {
  storage.removeItem(sessionKey);
  storage.removeItem(verifierKey);
  storage.removeItem(stateKey);
  storage.removeItem(returnToKey);
}

export async function startSignIn(
  config: RuntimeConfig,
  location: Location = window.location,
  storage: Storage = window.sessionStorage
): Promise<void> {
  const state = randomURLSafeString(24);
  const verifier = randomURLSafeString(64);
  const challenge = await sha256URLSafe(verifier);
  storage.setItem(stateKey, state);
  storage.setItem(verifierKey, verifier);
  storage.setItem(returnToKey, location.pathname + location.search);

  const params = new URLSearchParams({
    client_id: config.oidcClientId,
    redirect_uri: config.oidcRedirectUri,
    response_type: 'code',
    scope: 'openid email profile',
    state,
    code_challenge: challenge,
    code_challenge_method: 'S256'
  });
  location.assign(`${config.oidcIssuer}/auth?${params.toString()}`);
}

export async function completeSignIn(
  config: RuntimeConfig,
  callbackUrl: string,
  fetchImpl: typeof fetch = fetch,
  storage: Storage = window.sessionStorage
): Promise<string> {
  const url = new URL(callbackUrl);
  const code = url.searchParams.get('code');
  const state = url.searchParams.get('state');
  const expectedState = storage.getItem(stateKey);
  const verifier = storage.getItem(verifierKey);
  if (!code || !state || !expectedState || state !== expectedState || !verifier) {
    throw new Error('Invalid sign-in callback.');
  }

  const response = await fetchImpl(`${config.oidcIssuer}/token`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    },
    body: new URLSearchParams({
      grant_type: 'authorization_code',
      client_id: config.oidcClientId,
      redirect_uri: config.oidcRedirectUri,
      code,
      code_verifier: verifier
    })
  });
  if (!response.ok) {
    throw new Error('Unable to complete sign-in.');
  }
  const token = (await response.json()) as TokenResponse;
  if (!token.id_token) {
    throw new Error('Sign-in response did not include an ID token.');
  }
  storeSession(
    {
      idToken: token.id_token,
      expiresAt: tokenExpiry(token)
    },
    storage
  );

  const returnTo = storage.getItem(returnToKey) ?? '/';
  storage.removeItem(verifierKey);
  storage.removeItem(stateKey);
  storage.removeItem(returnToKey);
  return returnTo;
}

function tokenExpiry(token: TokenResponse): number {
  const expiresIn = typeof token.expires_in === 'number' && token.expires_in > 0 ? token.expires_in : 3600;
  return Date.now() + expiresIn * 1000;
}

function randomURLSafeString(byteLength: number): string {
  const bytes = new Uint8Array(byteLength);
  crypto.getRandomValues(bytes);
  return base64URL(bytes);
}

async function sha256URLSafe(value: string): Promise<string> {
  const bytes = new TextEncoder().encode(value);
  const digest = await crypto.subtle.digest('SHA-256', bytes);
  return base64URL(new Uint8Array(digest));
}

function base64URL(bytes: Uint8Array): string {
  let binary = '';
  for (const byte of bytes) {
    binary += String.fromCharCode(byte);
  }
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}
