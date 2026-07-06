import { mkdirSync, writeFileSync } from 'node:fs';
import { dirname } from 'node:path';

const webOrigin = normalizedOrigin(process.env.STUFF_STASH_WEB_ORIGIN || 'http://localhost:5173');
const apiOrigin = normalizedOrigin(process.env.STUFF_STASH_API_ORIGIN || originWithPort(webOrigin, '8080'));
const issuer = process.env.STUFF_STASH_DEX_ISSUER || 'http://dex:5556/dex';
const dexHTTPAddr = process.env.STUFF_STASH_DEX_HTTP_ADDR || '0.0.0.0:5556';
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
