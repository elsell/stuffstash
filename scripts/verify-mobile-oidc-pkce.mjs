#!/usr/bin/env node
import crypto from 'node:crypto';

class CookieJar {
  cookies = new Map();

  store(response) {
    for (const header of setCookieHeaders(response.headers)) {
      const [pair] = header.split(';');
      const separator = pair.indexOf('=');
      if (separator <= 0) {
        continue;
      }
      this.cookies.set(pair.slice(0, separator).trim(), pair.slice(separator + 1).trim());
    }
  }

  header() {
    return [...this.cookies.entries()].map(([name, value]) => `${name}=${value}`).join('; ');
  }
}

const issuer = trimTrailingSlash(requiredEnv('STUFF_STASH_VERIFY_MOBILE_OIDC_ISSUER'));
const clientId = process.env.STUFF_STASH_VERIFY_MOBILE_OIDC_CLIENT_ID || 'stuff-stash-mobile-local';
const redirectUri = process.env.STUFF_STASH_VERIFY_MOBILE_OIDC_REDIRECT_URI || 'stuffstash://auth/callback';
const scopes = (process.env.STUFF_STASH_VERIFY_MOBILE_OIDC_SCOPES || 'openid email profile offline_access')
  .split(/[,\s]+/)
  .map((scope) => scope.trim())
  .filter(Boolean);
const username = process.env.STUFF_STASH_VERIFY_MOBILE_OIDC_USERNAME || 'owner@example.com';
const password = process.env.STUFF_STASH_VERIFY_MOBILE_OIDC_PASSWORD || 'password';
const apiBaseUrl = process.env.STUFF_STASH_VERIFY_MOBILE_OIDC_API_BASE_URL
  ? trimTrailingSlash(process.env.STUFF_STASH_VERIFY_MOBILE_OIDC_API_BASE_URL)
  : '';

const discovery = await fetchJson(`${issuer}/.well-known/openid-configuration`);
assertEqual(discovery.issuer, issuer, 'discovery issuer');
const authorizationEndpoint = requiredDiscoveryUrl(discovery, 'authorization_endpoint');
const tokenEndpoint = requiredDiscoveryUrl(discovery, 'token_endpoint');

const state = base64Url(crypto.randomBytes(24));
const codeVerifier = base64Url(crypto.randomBytes(48));
const codeChallenge = base64Url(crypto.createHash('sha256').update(codeVerifier).digest());
const authURL = new URL(authorizationEndpoint);
authURL.searchParams.set('client_id', clientId);
authURL.searchParams.set('redirect_uri', redirectUri);
authURL.searchParams.set('response_type', 'code');
authURL.searchParams.set('scope', scopes.join(' '));
authURL.searchParams.set('state', state);
authURL.searchParams.set('code_challenge', codeChallenge);
authURL.searchParams.set('code_challenge_method', 'S256');

const jar = new CookieJar();
const authorization = await completeBrowserFlow(authURL, jar);
assertEqual(authorization.state, state, 'authorization callback state');

const tokenResponse = await postToken(tokenEndpoint, {
  grant_type: 'authorization_code',
  client_id: clientId,
  code: authorization.code,
  redirect_uri: redirectUri,
  code_verifier: codeVerifier
});
assertTokenResponse(tokenResponse, 'authorization-code token response');
if (!tokenResponse.refresh_token) {
  fail('authorization-code token response did not include a refresh_token');
}

const refreshResponse = await postToken(tokenEndpoint, {
  grant_type: 'refresh_token',
  client_id: clientId,
  refresh_token: tokenResponse.refresh_token,
  scope: scopes.join(' ')
});
assertTokenResponse(refreshResponse, 'refresh token response');

if (apiBaseUrl) {
  await verifyAPIMobileAuthMetadata(apiBaseUrl);
  await verifyAPIIdentity(apiBaseUrl, refreshResponse.id_token);
}

console.log(`Mobile OIDC PKCE verification passed for issuer ${issuer} and client ${clientId}`);

async function completeBrowserFlow(authURL, jar) {
  let response = await request(authURL, { method: 'GET', jar });
  for (let attempts = 0; attempts < 12; attempts++) {
    const callback = callbackFromRedirect(response);
    if (callback) {
      return callback;
    }

    if (isRedirect(response)) {
      response = await request(absoluteLocation(response), { method: 'GET', jar });
      continue;
    }

    const html = await response.text();
    const redirectLink = redirectLinkFromHTML(html, response.url);
    if (redirectLink) {
      response = await request(redirectLink, { method: 'GET', jar });
      continue;
    }

    const form = parseForm(html, response.url);
    if (!form) {
      fail(`issuer did not return a recognizable login or approval form at ${response.url}`);
    }

    response = await submitForm(form, jar);
  }

  fail('authorization flow did not reach the configured mobile redirect URI');
}

async function submitForm(form, jar) {
  const body = new URLSearchParams(form.inputs);
  if (body.has('login')) {
    body.set('login', username);
  }
  if (body.has('username')) {
    body.set('username', username);
  }
  if (body.has('password')) {
    body.set('password', password);
  }
  if (!body.has('login') && !body.has('username') && !body.has('password')) {
    body.set('approval', 'approve');
  }

  return request(form.action, {
    method: 'POST',
    jar,
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body
  });
}

function parseForm(html, baseURL) {
  const formMatch = html.match(/<form\b[^>]*>/i);
  if (!formMatch) {
    return undefined;
  }
  const actionMatch = formMatch[0].match(/\saction=["']([^"']+)["']/i);
  const action = new URL(actionMatch?.[1] || baseURL, baseURL);
  const inputs = {};
  for (const inputMatch of html.matchAll(/<input\b[^>]*>/gi)) {
    const tag = inputMatch[0];
    const name = attr(tag, 'name');
    if (!name) {
      continue;
    }
    inputs[name] = attr(tag, 'value') || '';
  }
  return { action, inputs };
}

function redirectLinkFromHTML(html, baseURL) {
  const match = html.match(/<a\b[^>]*\shref=["']([^"']+)["']/i);
  if (!match) {
    return undefined;
  }
  return new URL(decodeHTMLAttribute(match[1]), baseURL);
}

function attr(tag, name) {
  const match = tag.match(new RegExp(`\\s${name}=["']([^"']*)["']`, 'i'));
  return match ? decodeHTMLAttribute(match[1]) : undefined;
}

function decodeHTMLAttribute(value) {
  return value
    .replaceAll('&amp;', '&')
    .replaceAll('&quot;', '"')
    .replaceAll('&#39;', "'")
    .replaceAll('&lt;', '<')
    .replaceAll('&gt;', '>');
}

async function postToken(tokenEndpoint, values) {
  const response = await fetch(tokenEndpoint, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams(values),
    redirect: 'manual'
  });
  const body = await response.text();
  if (!response.ok) {
    fail(`token endpoint returned ${response.status}: ${body}`);
  }
  return JSON.parse(body);
}

function assertTokenResponse(response, label) {
  if (!response.id_token) {
    fail(`${label} did not include an id_token`);
  }
  const claims = decodeJwtPayload(response.id_token);
  if (claims.iss !== issuer) {
    fail(`${label} id_token issuer ${claims.iss} did not match ${issuer}`);
  }
  const audiences = Array.isArray(claims.aud) ? claims.aud : [claims.aud];
  if (!audiences.includes(clientId)) {
    fail(`${label} id_token audience ${JSON.stringify(claims.aud)} did not include ${clientId}`);
  }
  if (typeof claims.exp !== 'number' || claims.exp * 1000 <= Date.now()) {
    fail(`${label} id_token is expired or missing exp`);
  }
}

async function verifyAPIMobileAuthMetadata(baseUrl) {
  const metadata = await fetchJson(`${baseUrl}/.well-known/stuff-stash/mobile-auth`);
  const data = metadata.data || {};
  assertEqual(data.issuer, issuer, 'API mobile auth issuer');
  assertEqual(data.clientId, clientId, 'API mobile auth clientId');
  assertEqual(data.redirectUri, redirectUri, 'API mobile auth redirectUri');
  const advertisedScopes = Array.isArray(data.scopes) ? data.scopes : [];
  for (const scope of scopes) {
    if (!advertisedScopes.includes(scope)) {
      fail(`API mobile auth scopes ${JSON.stringify(advertisedScopes)} did not include ${scope}`);
    }
  }
}

async function verifyAPIIdentity(baseUrl, idToken) {
  const response = await fetch(`${baseUrl}/me`, {
    headers: { Authorization: `Bearer ${idToken}` }
  });
  const body = await response.text();
  if (response.status !== 200) {
    fail(`API /me rejected refreshed mobile ID token with ${response.status}: ${body}`);
  }
}

function callbackFromRedirect(response) {
  if (!isRedirect(response)) {
    return undefined;
  }
  const location = response.headers.get('location') || '';
  if (!location.startsWith(redirectUri)) {
    return undefined;
  }
  const callbackURL = new URL(location);
  const error = callbackURL.searchParams.get('error');
  if (error) {
    fail(`authorization callback returned error ${error}: ${callbackURL.searchParams.get('error_description') || ''}`);
  }
  const code = callbackURL.searchParams.get('code');
  const returnedState = callbackURL.searchParams.get('state');
  if (!code || !returnedState) {
    fail('authorization callback did not include code and state');
  }
  return { code, state: returnedState };
}

async function request(url, options = {}) {
  const requestURL = normalizedURL(url);
  const cookieHeader = options.jar?.header() || '';
  const response = await fetch(requestURL, {
    method: options.method,
    headers: {
      ...(cookieHeader ? { Cookie: cookieHeader } : {}),
      ...(options.headers || {})
    },
    body: options.body,
    redirect: 'manual'
  });
  options.jar?.store(response);
  return response;
}

function normalizedURL(url) {
  return new URL(decodeHTMLAttribute(String(url)));
}

function isRedirect(response) {
  return response.status >= 300 && response.status < 400 && response.headers.has('location');
}

function absoluteLocation(response) {
  return new URL(response.headers.get('location'), response.url);
}

async function fetchJson(url) {
  const response = await fetch(url);
  const body = await response.text();
  if (!response.ok) {
    fail(`${url} returned ${response.status}: ${body}`);
  }
  return JSON.parse(body);
}

function requiredDiscoveryUrl(discovery, field) {
  if (typeof discovery[field] !== 'string' || discovery[field].trim().length === 0) {
    fail(`discovery document is missing ${field}`);
  }
  return discovery[field];
}

function decodeJwtPayload(jwt) {
  const payload = jwt.split('.')[1];
  if (!payload) {
    fail('id_token is not a JWT');
  }
  return JSON.parse(Buffer.from(payload, 'base64url').toString('utf8'));
}

function assertEqual(actual, expected, label) {
  if (actual !== expected) {
    fail(`expected ${label} ${expected}, got ${actual}`);
  }
}

function requiredEnv(name) {
  const value = process.env[name]?.trim();
  if (!value) {
    fail(`${name} is required`);
  }
  return value;
}

function trimTrailingSlash(value) {
  return value.replace(/\/+$/, '');
}

function base64Url(bytes) {
  return Buffer.from(bytes).toString('base64url');
}

function fail(message) {
  console.error(`mobile oidc pkce verification failed: ${message}`);
  process.exit(1);
}

function setCookieHeaders(headers) {
  if (typeof headers.getSetCookie === 'function') {
    return headers.getSetCookie();
  }
  const header = headers.get('set-cookie');
  return header ? [header] : [];
}
