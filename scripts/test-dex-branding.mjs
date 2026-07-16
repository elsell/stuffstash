import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import { existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, symlinkSync, writeFileSync } from 'node:fs';
import { homedir, tmpdir } from 'node:os';
import { join, parse, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = new URL('../', import.meta.url);
const read = (path) => readFileSync(new URL(path, root), 'utf8');
const reviewedGlyphHash = 'ccf66d8d99818fce04d58580ea2c9ccc72996613eb1cf61c5d08cabbe3a86cb7';
mkdirSync(join(resolve('.'), '.stuffstash'), { recursive: true });

const expectedFrontend = [
  'frontend:',
  '  dir: /srv/dex/web',
  '  issuer: Stuff Stash',
  '  theme: light',
].join('\n');

for (const configPath of ['deploy/local/dex/config.yaml', 'deploy/selfhost/dex/config.yaml']) {
  assert.match(read(configPath), new RegExp(escapeRegex(expectedFrontend)), `${configPath} must use the branded Dex frontend`);
}

const temporaryDirectory = mkdtempSync(join(resolve('.'), '.stuffstash', 'dex-branding-test-'));
try {
  const output = join(temporaryDirectory, 'config.yaml');
  const frontendOutput = join(temporaryDirectory, 'dex', 'web');
  execFileSync(process.execPath, [fileURLToPath(new URL('scripts/render-local-dex-config.mjs', root))], {
    env: {
      ...process.env,
      STUFF_STASH_DEX_CONFIG_OUT: output,
      STUFF_STASH_DEX_FRONTEND_DIR: frontendOutput,
      STUFF_STASH_DEX_FRONTEND_OUT: frontendOutput,
    },
    stdio: 'pipe',
  });
  const generatedConfig = readFileSync(output, 'utf8').replaceAll('"', '');
  assert.match(generatedConfig, new RegExp(escapeRegex(expectedFrontend.replace('/srv/dex/web', frontendOutput))), 'generated host config must use its materialized branded Dex frontend');
  for (const path of [
    'robots.txt',
    'static/main.css',
    'templates/approval.html',
    'templates/device.html',
    'templates/device_success.html',
    'templates/error.html',
    'templates/footer.html',
    'templates/header.html',
    'templates/login.html',
    'templates/oob.html',
    'templates/password.html',
    'themes/light/styles.css',
    'themes/light/logo.png',
    'themes/light/favicon.png',
  ]) {
    assert.equal(existsSync(join(frontendOutput, path)), true, `generated host frontend must contain ${path}`);
  }
  assert.equal(sha256(readFileSync(join(frontendOutput, 'themes/light/logo.png'))), reviewedGlyphHash, 'generated host logo must use the reviewed glyph bytes');
  assert.equal(sha256(readFileSync(join(frontendOutput, 'themes/light/favicon.png'))), reviewedGlyphHash, 'generated host favicon must use the reviewed glyph bytes');
} finally {
  rmSync(temporaryDirectory, { recursive: true, force: true });
}

const unsafeTestDirectory = mkdtempSync(join(resolve('.'), '.stuffstash', 'dex-branding-unsafe-'));
try {
  for (const unsafeOutput of [parse(resolve('.')).root, resolve('.'), homedir(), join(tmpdir(), 'unscoped-web')]) {
    assertUnsafeOutput(unsafeOutput, join(unsafeTestDirectory, 'unsafe-config.yaml'));
  }

  const symlinkTarget = join(unsafeTestDirectory, 'target');
  const symlinkDestination = join(unsafeTestDirectory, 'dex', 'web');
  const sentinel = join(symlinkTarget, 'templates', 'sentinel.txt');
  mkdirSync(join(symlinkTarget, 'templates'), { recursive: true });
  mkdirSync(join(unsafeTestDirectory, 'dex'), { recursive: true });
  writeFileSync(sentinel, 'must survive');
  symlinkSync(symlinkTarget, symlinkDestination, 'dir');
  assertUnsafeOutput(symlinkDestination, join(unsafeTestDirectory, 'symlink-config.yaml'));
  assert.equal(readFileSync(sentinel, 'utf8'), 'must survive', 'renderer must not follow a destination symlink into owned entries');
} finally {
  rmSync(unsafeTestDirectory, { recursive: true, force: true });
}

for (const composePath of ['compose.oidc.yaml', 'compose.selfhost.yaml']) {
  const compose = read(composePath);
  assert.match(compose, /deploy\/dex\/theme\/styles\.css:\/srv\/dex\/web\/themes\/light\/styles\.css:ro/, `${composePath} must mount the Stuff Stash theme stylesheet read-only`);
  assert.match(compose, /stuff-stash-glyph\.png:\/srv\/dex\/web\/themes\/light\/logo\.png:ro/, `${composePath} must mount the approved glyph as the Dex logo`);
  assert.match(compose, /stuff-stash-glyph\.png:\/srv\/dex\/web\/themes\/light\/favicon\.png:ro/, `${composePath} must mount the approved glyph as the Dex favicon`);
  assert.match(compose, /deploy\/dex\/templates\/header\.html:\/srv\/dex\/web\/templates\/header\.html:ro/, `${composePath} must mount the reviewed accessible header template read-only`);
  assert.match(compose, /deploy\/dex\/templates\/login\.html:\/srv\/dex\/web\/templates\/login\.html:ro/, `${composePath} must mount the reviewed login template read-only`);
  assert.match(compose, /deploy\/dex\/templates\/password\.html:\/srv\/dex\/web\/templates\/password\.html:ro/, `${composePath} must mount the reviewed password template read-only`);
}

const styles = read('deploy/dex/theme/styles.css');
assert.match(styles, /min-height:\s*44px/, 'theme actions must be at least 44 CSS pixels high');
assert.match(styles, /:focus-visible/, 'theme must provide visible keyboard focus');
assert.match(styles, /@media\s*\(max-width:\s*[^)]+\)/, 'theme must define narrow-screen reflow');
assert.doesNotMatch(styles, /https?:\/\//, 'theme must not fetch remote assets');

const glyphPath = new URL('docs/public/brand/stuff-stash-glyph.png', root);
assert.equal(existsSync(glyphPath), true, 'approved Stuff Stash glyph must exist');
assert.equal(sha256(readFileSync(glyphPath)), reviewedGlyphHash, 'approved Stuff Stash glyph must match the reviewed SHA-256');

for (const templatePath of [
  'deploy/dex/templates/approval.html',
  'deploy/dex/templates/device.html',
  'deploy/dex/templates/device_success.html',
  'deploy/dex/templates/error.html',
  'deploy/dex/templates/header.html',
  'deploy/dex/templates/login.html',
  'deploy/dex/templates/oob.html',
  'deploy/dex/templates/password.html',
]) {
  const template = read(templatePath);
  const visibleTemplateCopy = template.replace(/\{\{[\s\S]*?\}\}/g, '');
  assert.doesNotMatch(visibleTemplateCopy, />[^<]*(?:\bDex\b|connector|fixture)[^<]*</i, `${templatePath} must not expose provider internals`);
}

for (const templatePath of ['deploy/dex/templates/login.html', 'deploy/dex/templates/password.html']) {
  assert.match(read(templatePath), /Sign in/, `${templatePath} must use provider-neutral sign-in language`);
}

const passwordTemplate = read('deploy/dex/templates/password.html');
assert.match(passwordTemplate, /<label for="login">/, 'username label must target the username input');
assert.match(passwordTemplate, /aria-describedby="login-error"/, 'invalid credentials must be associated with the form controls');

const headerTemplate = read('deploy/dex/templates/header.html');
assert.match(headerTemplate, /class="theme-navbar__logo"[^>]*alt=""/, 'decorative brand glyph must have an empty text alternative');

const errorTemplate = read('deploy/dex/templates/error.html');
assert.doesNotMatch(errorTemplate, /\.Err(?:Type|Msg)/, 'sign-in errors must not render raw provider diagnostics');
assert.match(errorTemplate, /start sign-in again/i, 'sign-in errors must provide a calm recovery action');

console.log('Dex branding contract verified.');

function escapeRegex(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function sha256(value) {
  return createHash('sha256').update(value).digest('hex');
}

function assertUnsafeOutput(frontendOutput, configOutput) {
  assert.throws(() => execFileSync(process.execPath, [fileURLToPath(new URL('scripts/render-local-dex-config.mjs', root))], {
    env: {
      ...process.env,
      STUFF_STASH_DEX_CONFIG_OUT: configOutput,
      STUFF_STASH_DEX_FRONTEND_DIR: frontendOutput,
      STUFF_STASH_DEX_FRONTEND_OUT: frontendOutput,
    },
    stdio: 'pipe',
  }), (error) => error.stderr?.toString().includes('refusing unsafe Dex frontend output'), `renderer must reject unsafe frontend output ${frontendOutput} before deletion`);
}
