import type { RuntimeConfig } from './runtimeConfig';

const verifierKey = 'stuffstash.oidc.verifier';
const stateKey = 'stuffstash.oidc.state';
const returnToKey = 'stuffstash.oidc.returnTo';
const sessionKey = 'stuffstash.oidc.session';
const selectedTenantKey = 'stuffstash.selectedTenantId';
const selectedInventoryKey = 'stuffstash.selectedInventoryId';

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
  storage.removeItem(selectedTenantKey);
  storage.removeItem(selectedInventoryKey);
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

export async function sha256URLSafe(value: string): Promise<string> {
  const bytes = new TextEncoder().encode(value);
  const digest = crypto.subtle ? new Uint8Array(await crypto.subtle.digest('SHA-256', bytes)) : sha256(bytes);
  return base64URL(digest);
}

function base64URL(bytes: Uint8Array): string {
  let binary = '';
  for (const byte of bytes) {
    binary += String.fromCharCode(byte);
  }
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

function sha256(message: Uint8Array): Uint8Array {
  const constants = [
    0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
    0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
    0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
    0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
    0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
    0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
    0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
    0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2
  ];
  const hash = [0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a, 0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19];
  const bitLength = message.length * 8;
  const paddedLength = Math.ceil((message.length + 9) / 64) * 64;
  const padded = new Uint8Array(paddedLength);
  padded.set(message);
  padded[message.length] = 0x80;
  const view = new DataView(padded.buffer);
  view.setUint32(paddedLength - 4, bitLength, false);

  const words = new Uint32Array(64);
  for (let chunk = 0; chunk < padded.length; chunk += 64) {
    for (let index = 0; index < 16; index += 1) {
      words[index] = view.getUint32(chunk + index * 4, false);
    }
    for (let index = 16; index < 64; index += 1) {
      const s0 = rotateRight(words[index - 15], 7) ^ rotateRight(words[index - 15], 18) ^ (words[index - 15] >>> 3);
      const s1 = rotateRight(words[index - 2], 17) ^ rotateRight(words[index - 2], 19) ^ (words[index - 2] >>> 10);
      words[index] = (words[index - 16] + s0 + words[index - 7] + s1) >>> 0;
    }
    let [a, b, c, d, e, f, g, h] = hash;
    for (let index = 0; index < 64; index += 1) {
      const s1 = rotateRight(e, 6) ^ rotateRight(e, 11) ^ rotateRight(e, 25);
      const ch = (e & f) ^ (~e & g);
      const temp1 = (h + s1 + ch + constants[index] + words[index]) >>> 0;
      const s0 = rotateRight(a, 2) ^ rotateRight(a, 13) ^ rotateRight(a, 22);
      const maj = (a & b) ^ (a & c) ^ (b & c);
      const temp2 = (s0 + maj) >>> 0;
      h = g;
      g = f;
      f = e;
      e = (d + temp1) >>> 0;
      d = c;
      c = b;
      b = a;
      a = (temp1 + temp2) >>> 0;
    }
    hash[0] = (hash[0] + a) >>> 0;
    hash[1] = (hash[1] + b) >>> 0;
    hash[2] = (hash[2] + c) >>> 0;
    hash[3] = (hash[3] + d) >>> 0;
    hash[4] = (hash[4] + e) >>> 0;
    hash[5] = (hash[5] + f) >>> 0;
    hash[6] = (hash[6] + g) >>> 0;
    hash[7] = (hash[7] + h) >>> 0;
  }

  const digest = new Uint8Array(32);
  const digestView = new DataView(digest.buffer);
  hash.forEach((word, index) => digestView.setUint32(index * 4, word, false));
  return digest;
}

function rotateRight(value: number, shift: number): number {
  return (value >>> shift) | (value << (32 - shift));
}
