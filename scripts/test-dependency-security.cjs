const assert = require('node:assert/strict');
const { createRequire } = require('node:module');
const { resolve } = require('node:path');
const { test } = require('node:test');

const mobile = createRequire(resolve(__dirname, '../apps/mobile/package.json'));
const expo = createRequire(mobile.resolve('expo/package.json'));
const cli = createRequire(expo.resolve('@expo/cli/package.json'));
const forge = cli('node-forge');
const metro = createRequire(expo.resolve('metro/package.json'));
const micromatch = createRequire(metro.resolve('micromatch/package.json'));
const braces = micromatch('braces');
const docs = createRequire(resolve(__dirname, '../docs/package.json'));
const astro = createRequire(docs.resolve('astro/package.json'));
const CachePolicy = astro('http-cache-semantics');

test('RSA verification accepts valid signatures and rejects extra DigestAlgorithm members', () => {
  const keys = forge.pki.rsa.generateKeyPair({ bits: 1024, e: 65537 });
  const digest = forge.md.sha256.create().update('dependency security regression');
  assert.equal(keys.publicKey.verify(digest.digest().getBytes(), keys.privateKey.sign(digest)), true);
  const a = forge.asn1;
  const digestInfo = a.create(a.Class.UNIVERSAL, a.Type.SEQUENCE, true, [
    a.create(a.Class.UNIVERSAL, a.Type.SEQUENCE, true, [
      a.create(a.Class.UNIVERSAL, a.Type.OID, false, a.oidToDer(forge.pki.oids.sha256).getBytes()),
      a.create(a.Class.UNIVERSAL, a.Type.NULL, false, ''),
      a.create(a.Class.UNIVERSAL, a.Type.OCTETSTRING, false, 'unconsumed data'),
    ]),
    a.create(a.Class.UNIVERSAL, a.Type.OCTETSTRING, false, digest.digest().getBytes()),
  ]);
  const malformed = keys.privateKey.sign(a.toDer(digestInfo).getBytes(), 'NONE');
  assert.throws(() => keys.publicKey.verify(digest.digest().getBytes(), malformed));
});

test('brace expansion bounds nesting without breaking ordinary patterns', () => {
  assert.deepEqual(braces.expand('photo-{small,large}.{png,jpg}'),
    ['photo-small.png', 'photo-small.jpg', 'photo-large.png', 'photo-large.jpg']);
  for (const delimiter of [['{', '}'], ['(', ')']]) {
    assert.throws(() => braces(delimiter[0].repeat(3500) + 'a,b' + delimiter[1].repeat(3500)),
      error => /exceeds max depth/i.test(error.message));
  }
});

test('max-stale cannot bypass shared-cache cookie or proxy-revalidation restrictions', () => {
  const request = { url: 'https://example.test/image.png', method: 'GET', headers: {} };
  const staleRequest = { ...request, headers: { 'cache-control': 'max-stale' } };
  for (const headers of [
    { 'cache-control': 'max-age=60', 'set-cookie': 'session=private' },
    { 'cache-control': 'max-age=60, proxy-revalidate' },
    { 'cache-control': 'no-cache' },
    { 'cache-control': 'no-cache, stale-while-revalidate=600' },
    { 'cache-control': 'no-store' },
    { 'cache-control': 'private, max-age=60' },
  ]) {
    const policy = new CachePolicy(request, { status: 200, headers });
    assert.equal(policy.satisfiesWithoutRevalidation(staleRequest), false);
    assert.equal(policy.evaluateRequest(request).response, undefined);
  }
  const publicPolicy = new CachePolicy(request, { status: 200, headers: { 'cache-control': 'public, max-age=60' } });
  assert.equal(publicPolicy.satisfiesWithoutRevalidation(request), true);
  const ordinaryStale = new CachePolicy(request, { status: 200, headers: { 'cache-control': 'public, max-age=0' } });
  assert.equal(ordinaryStale.satisfiesWithoutRevalidation(staleRequest), true);
});


test('Metro reads PNG asset dimensions through image-size 2', async () => {
  const { mkdtempSync, writeFileSync, rmSync } = require('node:fs');
  const { tmpdir } = require('node:os');
  const directory = mkdtempSync(resolve(tmpdir(), 'stuffstash-asset-security-'));
  try {
    const image = resolve(directory, 'pixel.png');
    writeFileSync(image, Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII=', 'base64'));
    const data = await metro('./src/Assets.js').getAssetData(image, 'pixel.png', [], null, '/assets');
    assert.equal(data.width, 1);
    assert.equal(data.height, 1);
  } finally { rmSync(directory, { recursive: true, force: true }); }
});


test('coverage YAML loading and instrumentation work without the vulnerable sprintf dependency', async () => {
  const { mkdtempSync, writeFileSync, rmSync } = require('node:fs');
  const { tmpdir } = require('node:os');
  const { runInNewContext } = require('node:vm');
  const native = createRequire(mobile.resolve('react-native/package.json'));
  const babelJest = createRequire(native.resolve('babel-jest/package.json'));
  const istanbul = createRequire(babelJest.resolve('babel-plugin-istanbul/package.json'));
  const nyc = createRequire(istanbul.resolve('@istanbuljs/load-nyc-config/package.json'));
  const yaml = createRequire(nyc.resolve('js-yaml/package.json'));
  const argumentParser = yaml('argparse/package.json');
  assert.equal(argumentParser.dependencies?.['sprintf-js'], undefined,
    'coverage configuration still brings the vulnerable formatter into the installed graph');

  const directory = mkdtempSync(resolve(tmpdir(), 'stuffstash-coverage-security-'));
  try {
    writeFileSync(resolve(directory, 'package.json'), '{"name":"coverage-security-fixture"}');
    writeFileSync(resolve(directory, '.nycrc.yaml'),
      'all: true\ncheck-coverage: true\nlines: 80\ninclude:\n  - "subject.js"\nexclude:\n  - "ignored.js"\n');
    const config = await istanbul('@istanbuljs/load-nyc-config').loadNycConfig({ cwd: directory });
    assert.equal(config.all, true);
    assert.equal(config.checkCoverage, true);
    assert.equal(config.lines, 80);
    assert.deepEqual(config.include, ['subject.js']);
    assert.deepEqual(config.exclude, ['ignored.js']);

    const filename = resolve(directory, 'subject.js');
    const result = babelJest('@babel/core').transformSync('module.exports = value => value + 1;', {
      filename, cwd: directory, configFile: false, babelrc: false,
      plugins: [[babelJest('babel-plugin-istanbul'), config]],
    });
    const sandbox = { module: { exports: {} } };
    runInNewContext(result.code, sandbox);
    assert.equal(sandbox.module.exports(41), 42);
    const coverage = sandbox.__coverage__[filename];
    assert.ok(coverage && Object.values(coverage.s).some(count => count > 0),
      'real Babel instrumentation did not record the executed code');
  } finally { rmSync(directory, { recursive: true, force: true }); }
});
