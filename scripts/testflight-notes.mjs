import { createPrivateKey, sign } from 'node:crypto';
import { execFileSync } from 'node:child_process';
import { pathToFileURL } from 'node:url';
import { setTimeout as sleep } from 'node:timers/promises';

export function renderNotes(tag, buildNumber, subjects) {
  const changes = [...new Set(subjects.flatMap(subject => {
    const match = /^(?:feat|fix|perf)(?:\([^)]*\))?!?:\s*(.+)$/.exec(subject);
    return match ? [match[1]] : [];
  }))];
  const heading = `Stuff Stash ${tag.slice(1)} (${buildNumber})\n\n`;
  const footer = `\n\nFull changelog: https://github.com/elsell/stuffstash/releases/tag/${tag}`;
  const body = changes.length ? changes.map(change => '- ' + change).join('\n') : 'Maintenance and reliability updates. See the full changelog for details.';
  const available = 4000 - heading.length - footer.length;
  return heading + (body.length <= available ? body : body.slice(0, available - 1) + '…') + footer;
}

export function appleToken({ key, keyId, issuerId }, now = Math.floor(Date.now() / 1000)) {
  const encode = value => Buffer.from(JSON.stringify(value)).toString('base64url');
  const message = encode({ alg: 'ES256', kid: keyId, typ: 'JWT' }) + '.' +
    encode({ iss: issuerId, iat: now, exp: now + 300, aud: 'appstoreconnect-v1' });
  return message + '.' + sign('sha256', Buffer.from(message), { key, dsaEncoding: 'ieee-p1363' }).toString('base64url');
}

export class AppleClient {
  constructor(credentials, { origin = 'https://api.appstoreconnect.apple.com', wait = sleep } = {}) {
    this.credentials = credentials;
    this.origin = origin;
    this.wait = wait;
  }
  async request(path, { method = 'GET', body } = {}) {
    const url = new URL(path, this.origin);
    if (url.origin !== this.origin) throw new Error('Refusing cross-origin Apple API request');
    for (let attempt = 0; ; attempt++) {
      let response;
      try {
        response = await fetch(url, {
          method, redirect: 'error', signal: AbortSignal.timeout(30000),
          headers: { Authorization: 'Bearer ' + appleToken(this.credentials), 'Content-Type': 'application/json' },
          ...(body ? { body: JSON.stringify(body) } : {})
        });
      } catch {
        if (method !== 'GET' || attempt === 3) throw new Error('Apple API connection failed');
        await this.wait(1000 * 2 ** attempt);
        continue;
      }
      if (response.ok) return response.json();
      if (method === 'GET' && attempt < 3 && (response.status === 429 || response.status >= 500)) {
        await response.body?.cancel();
        await this.wait(1000 * 2 ** attempt);
        continue;
      }
      // Do not expose Apple's response body, authorization headers or signing key.
      await response.body?.cancel();
      throw new Error('Apple API request failed (HTTP ' + response.status + ')');
    }
  }
}

function exactOne(items, description) {
  if (!Array.isArray(items) || items.length !== 1) throw new Error('Expected exactly one ' + description);
  return items[0];
}
export async function publishNotes(client, target, { sleep: wait = sleep, attempts = 60 } = {}) {
  const { bundleId, tag, buildNumber, notes } = target;
  if (!/^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/.test(tag) ||
      !/^[1-9]\d{0,3}\.[1-9]\d?$/.test(buildNumber) || !notes?.trim() || notes.length > 4000) {
    throw new Error('Invalid TestFlight notes target or text');
  }
  const app = exactOne((await client.request('/v1/apps?' + new URLSearchParams({ 'filter[bundleId]': bundleId, limit: '2' }))).data, 'matching app');
  if (app.attributes.bundleId !== bundleId) throw new Error('App bundle mismatch');
  let build;
  for (let attempt = 0; attempt < attempts; attempt++) {
    const result = await client.request('/v1/builds?' + new URLSearchParams({
      'filter[app]': app.id, 'filter[version]': buildNumber, include: 'preReleaseVersion', limit: '2'
    }));
    if (result.data.length) {
      build = exactOne(result.data, 'matching build');
      const version = result.included?.find(item => item.type === 'preReleaseVersions' && item.id === build.relationships.preReleaseVersion.data.id);
      if (build.attributes.version !== buildNumber || version?.attributes.version !== tag.slice(1) || version?.attributes.platform !== 'IOS') {
        throw new Error('Build version or platform mismatch');
      }
      if (['FAILED', 'INVALID'].includes(build.attributes.processingState)) throw new Error('Apple rejected build processing');
      if (build.attributes.processingState === 'VALID') break;
      if (build.attributes.processingState !== 'PROCESSING') throw new Error('Unknown Apple processing state');
    }
    build = undefined;
    if (attempt + 1 < attempts) await wait(30000);
  }
  if (!build) throw new Error('Timed out waiting for the exact TestFlight build');
  const path = '/v1/builds/' + encodeURIComponent(build.id) + '/betaBuildLocalizations?limit=200';
  const readLocales = async () => {
    const result = await client.request(path);
    if (result.links?.next) throw new Error('Unexpected localization pagination');
    return result.data.filter(item => item.attributes.locale === 'en-US');
  };
  const locales = await readLocales();
  if (locales.length > 1) throw new Error('Ambiguous English localization');
  const existing = locales[0];
  if (existing?.attributes.whatsNew !== notes) {
    const data = { type: 'betaBuildLocalizations', attributes: { whatsNew: notes } };
    if (existing) data.id = existing.id;
    else {
      data.attributes.locale = 'en-US';
      data.relationships = { build: { data: { type: 'builds', id: build.id } } };
    }
    await client.request(existing ? '/v1/betaBuildLocalizations/' + encodeURIComponent(existing.id) : '/v1/betaBuildLocalizations', {
      method: existing ? 'PATCH' : 'POST', body: { data }
    });
  }
  if (exactOne(await readLocales(), 'English localization').attributes.whatsNew !== notes) throw new Error('TestFlight notes verification failed');
}

function mainTarget(env) {
  const tag = env.STUFF_STASH_MOBILE_RELEASE_TAG;
  const buildNumber = env.MOBILE_BUILD_NUMBER;
  if (!/^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/.test(tag ?? '') ||
      !/^[1-9]\d{0,3}\.[1-9]\d?$/.test(buildNumber ?? '')) throw new Error('Invalid release tag or build number');
  const git = (...args) => execFileSync('git', args, { encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] }).trim();
  git('merge-base', '--is-ancestor', tag, 'HEAD');
  const previous = git('tag', '--merged', tag, '--list', 'v[0-9]*.[0-9]*.[0-9]*', '--sort=-v:refname')
    .split('\n').find(value => value !== tag && /^v\d+\.\d+\.\d+$/.test(value));
  const subjects = git('log', '--first-parent', '--format=%s', previous ? previous + '..' + tag : tag).split('\n');
  return { tag, buildNumber, bundleId: env.MOBILE_BUNDLE_ID, notes: renderNotes(tag, buildNumber, subjects) };
}
if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  try {
    const target = mainTarget(process.env);
    if (!target.bundleId || !process.env.APP_STORE_CONNECT_KEY_ID || !process.env.APP_STORE_CONNECT_ISSUER_ID) throw new Error('Missing Apple publication configuration');
    const key = createPrivateKey(Buffer.from(process.env.APP_STORE_CONNECT_API_KEY_BASE64 ?? '', 'base64'));
    if (key.asymmetricKeyType !== 'ec' || key.asymmetricKeyDetails?.namedCurve !== 'prime256v1') throw new Error('Expected an Apple P-256 signing key');
    await publishNotes(new AppleClient({ key, keyId: process.env.APP_STORE_CONNECT_KEY_ID, issuerId: process.env.APP_STORE_CONNECT_ISSUER_ID }), target);
    process.stdout.write('Verified TestFlight changelog for ' + target.tag + ' (' + target.buildNumber + ').\n');
  } catch (error) {
    // Crypto/git errors can contain input material: only publish known operational failures.
    const safe = /^(Apple API|App bundle|Build version|Apple rejected|Unknown Apple|Timed out|TestFlight notes|Expected exactly|Ambiguous|Unexpected localization|Invalid TestFlight|Invalid release|Missing Apple|Expected an Apple)/.test(error.message);
    process.stderr.write((safe ? error.message : 'TestFlight changelog publication failed') + '\n');
    process.exitCode = 1;
  }
}
