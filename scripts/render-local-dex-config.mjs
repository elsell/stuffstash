import { copyFileSync, existsSync, lstatSync, mkdirSync, readdirSync, rmSync, writeFileSync } from 'node:fs';
import { homedir } from 'node:os';
import { basename, dirname, isAbsolute, join, parse, relative, resolve, sep } from 'node:path';
import { fileURLToPath } from 'node:url';

const webOrigin = normalizedOrigin(process.env.STUFF_STASH_WEB_ORIGIN || 'http://localhost:5173');
const apiOrigin = normalizedOrigin(process.env.STUFF_STASH_API_ORIGIN || originWithPort(webOrigin, '8080'));
const issuer = process.env.STUFF_STASH_DEX_ISSUER || 'http://dex:5556/dex';
const dexHTTPAddr = process.env.STUFF_STASH_DEX_HTTP_ADDR || '0.0.0.0:5556';
const frontendDir = process.env.STUFF_STASH_DEX_FRONTEND_DIR ?? '/srv/dex/web';
const frontendOutput = process.env.STUFF_STASH_DEX_FRONTEND_OUT || '';
const output = process.env.STUFF_STASH_DEX_CONFIG_OUT || '.stuffstash/local/dex/config.yaml';
const mobileRedirectUri = process.env.STUFF_STASH_OIDC_MOBILE_REDIRECT_URI || 'stuffstash://auth/callback';

const allowedOrigins = unique(['http://localhost:5173', webOrigin]);
const localClientRedirects = unique(['http://localhost:8080/callback', 'http://localhost:5173/callback', `${apiOrigin}/callback`, `${webOrigin}/callback`]);
const webClientRedirects = unique(['http://localhost:5173/callback', `${webOrigin}/callback`]);
const mobileClientRedirects = unique([mobileRedirectUri]);

const config = `# Generated local Dex fixture.
# Do not commit this file. Use deploy/local/dex/config.yaml for the tracked default.

issuer: ${quote(issuer)}

storage:
  type: memory

frontend:
${frontendDir ? `  dir: ${quote(frontendDir)}\n` : ''}  issuer: Stuff Stash
${frontendDir ? '  theme: light\n' : ''}
web:
  http: ${dexHTTPAddr}
  allowedOrigins:
${allowedOrigins.map((origin) => `    - ${quote(origin)}`).join('\n')}

oauth2:
  passwordConnector: local
  skipApprovalScreen: true

staticClients:
  - id: stuff-stash-local
    name: Stuff Stash Local
    secret: stuff-stash-local-secret
    redirectURIs:
${localClientRedirects.map((uri) => `      - ${quote(uri)}`).join('\n')}
  - id: stuff-stash-web-local
    name: Stuff Stash Web Local
    public: true
    redirectURIs:
${webClientRedirects.map((uri) => `      - ${quote(uri)}`).join('\n')}
  - id: stuff-stash-mobile-local
    name: Stuff Stash Mobile Local
    public: true
    redirectURIs:
${mobileClientRedirects.map((uri) => `      - ${quote(uri)}`).join('\n')}
  - id: stuff-stash-wrong-audience
    name: Stuff Stash Wrong Audience Fixture
    secret: stuff-stash-wrong-audience-secret
    redirectURIs:
      - http://localhost:8080/callback

enablePasswordDB: true

staticPasswords:
  - email: owner@example.com
    hash: "$2a$10$2b2cU8CPhOTaGrs1HRQuAueS7JTT5ZHsHSzYiFPm1leZck7Mc8T4W"
    username: owner
    name: Owner User
    emailVerified: true
    preferredUsername: owner
    userID: "11111111-1111-1111-1111-111111111111"
  - email: viewer@example.com
    hash: "$2a$10$2b2cU8CPhOTaGrs1HRQuAueS7JTT5ZHsHSzYiFPm1leZck7Mc8T4W"
    username: viewer
    name: Viewer User
    emailVerified: true
    preferredUsername: viewer
    userID: "22222222-2222-2222-2222-222222222222"
`;

mkdirSync(dirname(output), { recursive: true });
writeFileSync(output, config);
if (frontendOutput) {
  materializeFrontend(frontendOutput);
}
console.log(`Wrote ${output}`);

function normalizedOrigin(value) {
  return new URL(value).origin;
}

function originWithPort(origin, port) {
  const url = new URL(origin);
  url.port = port;
  return url.origin;
}

function unique(values) {
  return [...new Set(values)];
}

function quote(value) {
  return JSON.stringify(value);
}

function materializeFrontend(destination) {
  destination = safeFrontendDestination(destination);
  const source = fileURLToPath(new URL('../deploy/dex', import.meta.url));
  for (const generatedEntry of ['robots.txt', 'static', 'templates', 'themes']) {
    rmSync(join(destination, generatedEntry), { recursive: true, force: true });
  }
  for (const directory of ['static', 'templates', 'themes/light']) {
    mkdirSync(join(destination, directory), { recursive: true });
  }
  copyFileSync(join(source, 'robots.txt'), join(destination, 'robots.txt'));
  copyFileSync(join(source, 'static/main.css'), join(destination, 'static/main.css'));
  for (const template of readdirSync(join(source, 'templates'))) {
    if (template.endsWith('.html')) {
      copyFileSync(join(source, 'templates', template), join(destination, 'templates', template));
    }
  }
  copyFileSync(join(source, 'theme/styles.css'), join(destination, 'themes/light/styles.css'));
  const glyph = fileURLToPath(new URL('../docs/public/brand/stuff-stash-glyph.png', import.meta.url));
  copyFileSync(glyph, join(destination, 'themes/light/logo.png'));
  copyFileSync(glyph, join(destination, 'themes/light/favicon.png'));
}

function safeFrontendDestination(value) {
  const destination = resolve(value);
  const workspace = resolve('.');
  const workspaceRelativePath = relative(workspace, destination);
  const outsideWorkspace = workspaceRelativePath === '..' || workspaceRelativePath.startsWith(`..${sep}`) || isAbsolute(workspaceRelativePath);
  const forbidden = new Set([parse(destination).root, workspace, resolve(homedir())]);
  if (outsideWorkspace || forbidden.has(destination) || basename(destination) !== 'web' || basename(dirname(destination)) !== 'dex') {
    throw new Error(`refusing unsafe Dex frontend output: ${value}; expected a dedicated dex/web child directory`);
  }
  let existingPath = workspace;
  for (const segment of workspaceRelativePath.split(sep)) {
    existingPath = join(existingPath, segment);
    if (existsSync(existingPath) && lstatSync(existingPath).isSymbolicLink()) {
      throw new Error(`refusing unsafe Dex frontend output: ${value}; symlinked paths are not allowed`);
    }
  }
  return destination;
}
