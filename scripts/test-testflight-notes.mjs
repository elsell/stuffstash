import { test } from 'node:test';
import assert from 'node:assert/strict';
import { generateKeyPairSync, verify } from 'node:crypto';
import { createServer } from 'node:http';
import { renderNotes, publishNotes, AppleClient, appleToken } from './testflight-notes.mjs';

test('renders concise release changes without plumbing, duplicates or prefixes', () => {
  const notes = renderNotes('v0.24.4', '90.1', ['fix(mobile): Keep map visible', 'ci: build', 'fix(mobile): Keep map visible', 'feat: Search']);
  assert.match(notes, /0.24.4 \(90.1\)/);
  assert.match(notes, /- Keep map visible\n- Search/);
  assert.doesNotMatch(notes, /fix\(|ci:|build\n/);
  assert.ok(renderNotes('v1.0.0', '1.1', ['fix: ' + 'x'.repeat(5000)]).length <= 4000);
});

const { privateKey, publicKey } = generateKeyPairSync('ec', { namedCurve: 'P-256' });
const credentials = { key: privateKey, keyId: 'ABCDEFGHIJ', issuerId: 'issuer' };
test('signs short-lived Apple JWTs with ES256 P1363 signatures', () => {
  const token = appleToken(credentials, 1000);
  const [header, body, signature] = token.split('.');
  assert.equal(JSON.parse(Buffer.from(header, 'base64url')).alg, 'ES256');
  assert.deepEqual(JSON.parse(Buffer.from(body, 'base64url')), { iss: 'issuer', iat: 1000, exp: 1300, aud: 'appstoreconnect-v1' });
  assert.ok(verify('sha256', Buffer.from(header + '.' + body), { key: publicKey, dsaEncoding: 'ieee-p1363' }, Buffer.from(signature, 'base64url')));
});

async function fixture(options, run) {
  let notes = options.same ? 'New notes' : options.existing ? 'Old notes' : undefined;
  let reads = 0;
  const writes = [];
  const server = createServer(async (req, res) => {
    const url = new URL(req.url, 'http://localhost');
    const payload = [];
    for await (const chunk of req) payload.push(chunk);
    const send = (data, status = 200) => { res.writeHead(status, { 'Content-Type': 'application/json' }); res.end(JSON.stringify(data)); };
    if (options.denied) return send({}, 403);
    assert.match(req.headers.authorization, /^Bearer /);
    if (url.pathname === '/v1/apps') {
      assert.equal(url.searchParams.get('filter[bundleId]'), 'org.stuffstash.mobile');
      return send({ data: [{ id: 'app', attributes: { bundleId: options.wrongApp ? 'another.bundle' : 'org.stuffstash.mobile' } }] });
    }
    if (url.pathname === '/v1/builds') {
      assert.equal(url.searchParams.get('filter[app]'), 'app');
      assert.equal(url.searchParams.get('filter[version]'), '90.1');
      reads++;
      return send({ data: [{ id: 'build', attributes: { version: options.wrongBuild ? '89.1' : '90.1', processingState: options.failed ? 'INVALID' : options.processing && reads === 1 ? 'PROCESSING' : 'VALID' },
        relationships: { preReleaseVersion: { data: { id: 'version' } } } }],
        included: [{ type: 'preReleaseVersions', id: 'version', attributes: { version: options.wrongVersion ? '0.24.3' : '0.24.4', platform: 'IOS' } }] });
    }
    if (url.pathname === '/v1/builds/build/betaBuildLocalizations') return send({ data: [
      { id: 'french', attributes: { locale: 'fr-FR', whatsNew: 'Bonjour' } },
      ...(notes === undefined ? [] : [{ id: 'english', attributes: { locale: 'en-US', whatsNew: notes } }])
    ] });
    if (req.method === 'POST' || req.method === 'PATCH') {
      const data = JSON.parse(Buffer.concat(payload).toString()).data;
      writes.push({ path: url.pathname, method: req.method, data });
      notes = options.readbackMismatch ? 'Different' : data.attributes.whatsNew;
      return send({ data: { id: 'english' } }, 201);
    }
    send({}, 404);
  });
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
  try {
    await run(new AppleClient(credentials, { origin: 'http://127.0.0.1:' + server.address().port }), writes);
  } finally { await new Promise(resolve => server.close(resolve)); }
}
const target = { bundleId: 'org.stuffstash.mobile', tag: 'v0.24.4', buildNumber: '90.1', notes: 'New notes' };
for (const existing of [false, true]) test('publishes exact build notes: ' + (existing ? 'update' : 'create'), async () => {
  let sleeps = 0;
  await fixture({ existing, processing: true }, async (client, writes) => {
    await publishNotes(client, target, { sleep: async () => { sleeps++; } });
    assert.equal(sleeps, 1);
    assert.equal(writes.length, 1);
    assert.equal(writes[0].method, existing ? 'PATCH' : 'POST');
    assert.notEqual(writes[0].path, '/v1/betaBuildLocalizations/french');
    if (!existing) assert.equal(writes[0].data.relationships.build.data.id, 'build');
  });
});
for (const options of [{ denied: true }, { wrongVersion: true }, { wrongApp: true }, { wrongBuild: true }, { failed: true }]) test('rejects unsafe target ' + JSON.stringify(options), async () => {
  await fixture(options, async (client, writes) => {
    await assert.rejects(publishNotes(client, target));
    assert.deepEqual(writes, []);
  });
});

test('does not rewrite notes already published and verified', async () => {
  await fixture({ same: true }, async (client, writes) => {
    await publishNotes(client, target);
    assert.deepEqual(writes, []);
  });
});
test('fails if Apple readback does not match the changelog', async () => {
  await fixture({ readbackMismatch: true }, async client => {
    await assert.rejects(publishNotes(client, target), /verification failed/);
  });
});
