# Local Development Topology Spec

## Purpose

Stuff Stash needs a local stack that proves production-shaped boundaries without making early development slow.

## Scope

This spec covers the first local services used by the secure tracer bullet.

It does not define Kubernetes production deployment, external Google OIDC rollout, or model provider services.

## Decisions

- Local development must be able to run the API, Postgres, SpiceDB, and a local OIDC provider.
- Container images must be pinned by immutable digest.
- Compose is the local orchestration tool.
- The API may use in-memory adapters for a tracer bullet only when a spec states the production adapter that will replace them.
- Postgres is the production persistence target.
- SQLite remains allowed for local-only or test fakes where a spec permits it.
- SpiceDB is the production authorization service.
- Local development may use a deterministic development authentication adapter behind the same authentication port used by OIDC.
- Local development may use the in-memory authorization adapter only as a development convenience. When paired with a durable repository such as SQLite for phone testing, API bootstrap must rebuild the in-memory authorization graph from durable authorization intent before serving requests so persisted local tenants and inventories do not become unusable after an API restart.
- Local authorization replay must use project-owned authorization/outbox ports. It must not read database tables directly from bootstrap code, must not bypass authorization adapters, and must not run when a production authorization adapter such as SpiceDB is selected.
- Local authorization replay must apply durable tenant-owner, inventory-owner, viewer, editor, and revoke intent in deterministic creation order and ignore dead-lettered authorization intent. Replayed grants and revokes must remain idempotent.
- Local development must also support Dex as a deterministic OIDC issuer so the API can be verified with real OIDC discovery and token verification.
- The Docker Compose self-host topology must include Dex, Postgres, datastore-backed SpiceDB, Garage, the API, and the static web container so the documented path is production-shaped by default.
- The Docker Compose contributor evaluation topology may remain separate for fast local API development, but public self-hosting documentation must not present it as the user-facing happy path.
- The Compose self-host/evaluation topology must be runnable with plain `docker compose` commands. Make targets may remain contributor conveniences, but public self-host documentation must not require GNU Make.
- Developers may switch the API to production-shaped OIDC and SpiceDB adapters through environment variables without code changes.
- Local Compose must provide enough SpiceDB configuration for the API to connect to SpiceDB when `STUFF_STASH_AUTHZ_MODE=spicedb`.
- Local SpiceDB uses `serve-testing`; it must not require a preshared key.
- The API's SpiceDB adapter may omit bearer credentials when no preshared key is configured, for local `serve-testing` only.
- The repository must provide a single-command Compose path for local-dev authentication with SpiceDB authorization.
- The repository must provide a single-command Compose path for Dex OIDC authentication with SpiceDB authorization.
- The repository must provide a non-Compose local SpiceDB path for developer machines where the Docker Compose plugin is unavailable.
- The repository must provide a local verification script for the first secure API flow.
- The repository must provide a local verification command for the Garage/S3-compatible blob storage adapter.

## First Services

The base Compose topology includes:

- `app`: Stuff Stash API.
- `postgres`: local Postgres database.
- `spicedb`: local SpiceDB service.
- `garage`: local Garage S3-compatible blob storage service.

The OIDC Compose override adds:

- `dex`: local OIDC provider fixture. Developers may run Dex either through
  the Compose override or as a host-installed `dex` binary with a generated
  ignored config.

## Configuration

- All service configuration must come from environment variables.
- Local defaults may be provided in Compose for developer ergonomics.
- Secrets used in local Compose must be clearly local-only.
- Production secrets must never be committed.
- Local Garage access keys, secret keys, bucket names, RPC secrets, and admin tokens in Compose are local-only fixtures.
- Compose must configure the API for `STUFF_STASH_BLOB_STORAGE_MODE=s3` with Garage by default.
- Compose must expose a Garage S3 API port for browser direct-upload attempts and local verification. The browser-reachable S3 endpoint must be configurable because `localhost` is correct only when the browser runs on the same host as Docker.
- Dex static users and static client secrets are local-only verification fixtures.
- Dex must be selected through the same `STUFF_STASH_AUTH_MODE=oidc`, `STUFF_STASH_OIDC_ISSUER`, and `STUFF_STASH_OIDC_CLIENT_ID` configuration used by any other OIDC issuer.
- Local Dex may configure `STUFF_STASH_OIDC_CLIENT_IDS` when the API must accept ID tokens issued to more than one local client, such as the API verification fixture, browser public client, and mobile public client.
- Local Dex must provide a public native mobile client for Expo development builds using the app scheme redirect URI specified by `specs/identity-access/mobile-oidc-authentication.spec.md`.
- Physical-device mobile OIDC development must use a generated local Dex config
  whose issuer is reachable from the device and must start the API with that
  same issuer. `make dex-local STUFF_STASH_LOCAL_HOST=<host-ip>` and
  `make run-oidc-local STUFF_STASH_LOCAL_HOST=<host-ip>` are the named
  host-binary workflow for this topology. `make compose-up-oidc-lan
  STUFF_STASH_LAN_HOST=<host-ip>` remains the named Docker Compose workflow when
  Docker is available.

## Verification

- `make test` must pass without requiring Compose.
- Compose should be able to start the local service graph when Docker is available.
- The API health endpoint must remain available without authentication.
- Protected endpoints must require authentication even in local development.
- Running with unknown auth or authz modes must fail startup.
- The SpiceDB Compose path must bootstrap the checked-in schema automatically.
- The non-Compose SpiceDB path must bootstrap the checked-in schema automatically.
- The local verification script must cover health, unauthenticated rejection, authenticated identity, tenant creation, inventory creation, inventory listing with pagination metadata, custom field definition creation/listing, asset creation with validated custom field values, asset update/movement, asset listing with pagination metadata, asset attachment upload/list/download, inventory audit listing with pagination metadata, direct inventory sharing, direct grant listing with pagination metadata, and adversarial viewer permission checks.
- The Dex OIDC verification script must run that same API user flow using two real Dex-issued ID tokens and SpiceDB authorization.
- The Dex OIDC verification script must also reject missing, malformed, unsigned, and wrong-audience OIDC tokens at the HTTP boundary.
- The mobile OIDC verification command must prove the native-app public-client
  authorization-code flow with PKCE against Dex discovery, callback state, token
  exchange, refresh exchange, and API mobile metadata. When an API base URL is
  provided, it must verify that the API accepts a refreshed mobile ID token at
  the HTTP boundary.
- The repository must provide an explicit real-SpiceDB adapter verification command that starts pinned local SpiceDB, runs the adapter integration tests, and cleans up.
- The repository must provide an explicit Garage blob storage verification command that starts a digest-pinned Garage image, runs the S3-compatible blob adapter integration test, and cleans up. Any custom image override must also be pinned with `@sha256:`.

## Self-Hosting Documentation Contract

The public documentation must lead with the lowest-friction path toward a real
self-hosted deployment, not with contributor convenience commands.

The self-hosting path must make these boundaries explicit:

- The happy path must start with one durable Compose command. It must not ask
  self-hosters to start a host Vite development server, wire a separate OIDC
  provider, or choose between SQLite and Postgres before they have seen the app.
- A self-hosted setup must use Dex OIDC, Postgres metadata persistence,
  datastore-backed SpiceDB authorization, durable blob storage, and the static
  web container.
- Bundled Dex is the default OIDC provider for Docker Compose self-hosting.
  Operators may replace Dex with another standards-compliant OIDC provider
  later, but that is an advanced configuration rather than the first-run path.
- The self-host happy path must be restart-durable for authenticated inventory
  access when containers are stopped without removing volumes.
- SQLite is an API runtime mode. It is not a Docker Compose self-hosting path
  unless a dedicated Compose topology wires the API, schema setup, durable file
  mount, and authorization behavior for SQLite.
- The documentation must distinguish browser origins, API origins, OIDC issuer
  URLs, and object-storage endpoints. For OIDC, the issuer used by the browser,
  API, and mobile metadata must be the same reachable issuer for the intended
  client.
- Browser media upload with S3-compatible storage requires a browser-reachable
  `STUFF_STASH_S3_PUBLIC_ENDPOINT` and bucket CORS policy that permits the web
  origin to upload through the S3 API.
- Local fixture secrets, static passwords, provider credential encryption keys,
  database passwords, object-storage keys, OIDC client configuration, and
  SpiceDB credentials must be listed in one operator checklist with guidance to
  replace them before a household relies on the deployment.
- Volume ownership and purpose must be obvious to an operator. The docs must
  name which volumes contain metadata, authorization state, object metadata, and
  object data.

## Durable Self-Host Compose Topology

The repository must provide a durable Docker Compose file for self-hosting. It
must not replace the contributor local Compose path, but public documentation
must lead with the durable self-host path.

The durable self-host Compose topology must:

- Be started with `docker compose -f compose.selfhost.yaml up --build`.
- Read operator-provided runtime values from a private `.env` file copied from
  the committed `.env.example`.
- Run Dex as the default OIDC provider for browser, API, and mobile metadata.
- Treat Dex readiness as OIDC discovery readiness. The API must not start from
  the self-host Compose dependency graph until Dex's discovery document is
  reachable, because the API validates the issuer during startup.
- Run the API with OIDC authentication, SpiceDB authorization, Postgres
  persistence, and Garage/S3-compatible blob storage.
- Run the static web image as a Compose service instead of requiring a host
  Vite development server.
- Generate the web runtime `config.json` from environment values at container
  startup so API base URL, OIDC issuer, web OIDC client ID, redirect URI, and
  media upload policy can change without rebuilding source.
- Generate web CSP values from the configured API, OIDC, and browser-reachable
  Garage/S3 origins at container startup.
- Configure API CORS from the configured web origin.
- Configure the Garage bucket CORS policy for the configured web origin before
  the web service is considered ready for browser upload testing.
- Persist API Postgres data, SpiceDB datastore data, Garage metadata, and Garage
  object data in named volumes.
- Use a datastore-backed SpiceDB service rather than `serve-testing`.
- Keep Redis and Valkey out of this topology until a future spec defines a
  concrete distributed cache, queue, or rate-limiter adapter need. The existing
  in-memory rate limiter remains acceptable for single-replica self-hosted
  Compose.
- Use one browser-and-container-reachable hostname for the web origin, API
  origin, Dex issuer, and Garage public endpoint. The default single-machine
  hostname is `stuffstash.localhost`; LAN and reverse-proxy operators must
  update the hostname consistently before starting the stack.
- Allow operators to provide a private Dex config for household users while
  keeping the committed Dex config suitable only as a first-run local example.
