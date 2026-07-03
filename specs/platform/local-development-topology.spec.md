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
- The Docker Compose self-host/evaluation topology must include Garage as the default blob storage service so media behavior is production-shaped by default.
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

- `dex`: local OIDC provider fixture.

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
- Local Dex may configure `STUFF_STASH_OIDC_CLIENT_IDS` when the API must accept ID tokens issued to more than one local client, such as the API verification fixture and browser public client.

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
- The repository must provide an explicit real-SpiceDB adapter verification command that starts pinned local SpiceDB, runs the adapter integration tests, and cleans up.
- The repository must provide an explicit Garage blob storage verification command that starts a digest-pinned Garage image, runs the S3-compatible blob adapter integration test, and cleans up. Any custom image override must also be pinned with `@sha256:`.
