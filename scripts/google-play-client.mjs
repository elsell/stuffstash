import { createPrivateKey, sign } from 'node:crypto';

const tokenEndpoint = 'https://oauth2.googleapis.com/token';
const publisherOrigin = 'https://androidpublisher.googleapis.com';

async function requestJSON(fetcher, url, options) {
  let response;
  try {
    response = await fetcher(url, { ...options, redirect: 'error', signal: AbortSignal.timeout(30000) });
  } catch { throw new Error('Google Play connection failed'); }
  if (!response.ok) {
    await response.body?.cancel();
    throw new Error('Google Play request failed (HTTP ' + response.status + ')');
  }
  if (response.status === 204) return null;
  try { return await response.json(); } catch { throw new Error('Invalid Google Play response'); }
}

export async function serviceAccountToken(credentials, { fetcher = fetch, now = () => Math.floor(Date.now() / 1000) } = {}) {
  if (credentials?.type !== 'service_account' || !/^[^\s@]+@[^\s@]+\.iam\.gserviceaccount\.com$/.test(credentials.client_email ?? '') ||
      (credentials.token_uri && credentials.token_uri !== tokenEndpoint)) throw new Error('Invalid Google Play credentials');
  let key;
  try { key = createPrivateKey(credentials.private_key); } catch { throw new Error('Invalid Google Play signing key'); }
  if (key.asymmetricKeyType !== 'rsa' || key.asymmetricKeyDetails.modulusLength < 2048) throw new Error('Invalid Google Play signing key');
  const issued = now();
  const encode = value => Buffer.from(JSON.stringify(value)).toString('base64url');
  const message = encode({ alg: 'RS256', typ: 'JWT' }) + '.' + encode({
    iss: credentials.client_email, scope: 'https://www.googleapis.com/auth/androidpublisher',
    aud: tokenEndpoint, iat: issued, exp: issued + 300,
  });
  const assertion = message + '.' + sign('RSA-SHA256', Buffer.from(message), key).toString('base64url');
  const result = await requestJSON(fetcher, tokenEndpoint, {
    method: 'POST', headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({ grant_type: 'urn:ietf:params:oauth:grant-type:jwt-bearer', assertion }).toString(),
  });
  if (result?.token_type !== 'Bearer' || typeof result.access_token !== 'string' || !result.access_token ||
      /[\r\n]/.test(result.access_token)) throw new Error('Invalid Google Play token');
  return result.access_token;
}

export class GooglePlayClient {
  constructor(token, { fetcher = fetch } = {}) {
    if (typeof token !== 'string' || !token || /[\r\n]/.test(token)) throw new Error('Invalid Google Play token');
    this.token = token; this.fetcher = fetcher;
  }
  request(path, { method = 'GET', body } = {}) {
    if (!/^\/applications\/[A-Za-z0-9_.]+\/edits(?:\/[A-Za-z0-9_-]+(?::commit|\/tracks\/[A-Za-z0-9_%:-]+)?)?$/.test(path)) {
      throw new Error('Invalid Google Play API path');
    }
    return requestJSON(this.fetcher, publisherOrigin + '/androidpublisher/v3' + path, {
      method, headers: { Authorization: 'Bearer ' + this.token, 'Content-Type': 'application/json' },
      ...(body === undefined ? {} : { body: JSON.stringify(body) }),
    });
  }
}
