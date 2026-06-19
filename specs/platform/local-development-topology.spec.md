# Local Development Topology Spec

## Purpose

Stuff Stash needs a local stack that proves production-shaped boundaries without making early development slow.

## Scope

This spec covers the first local services used by the secure tracer bullet.

It does not define Kubernetes production deployment, Garage media storage, Google OIDC, or model provider services.

## Decisions

- Local development must run the API, Postgres, and SpiceDB.
- Container images must be pinned by immutable digest.
- Compose is the local orchestration tool.
- The API may use in-memory adapters for a tracer bullet only when a spec states the production adapter that will replace them.
- Postgres is the production persistence target.
- SQLite remains allowed for local-only or test fakes where a spec permits it.
- SpiceDB is the production authorization service.
- Local development may use a deterministic development authentication adapter behind the same authentication port used by OIDC.
- Developers may switch the API to production-shaped OIDC and SpiceDB adapters through environment variables without code changes.
- Local Compose must provide enough SpiceDB configuration for the API to connect to SpiceDB when `STUFF_STASH_AUTHZ_MODE=spicedb`.
- Local SpiceDB uses `serve-testing`; it must not require a preshared key.
- The API's SpiceDB adapter may omit bearer credentials when no preshared key is configured, for local `serve-testing` only.
- The repository must provide a single-command Compose path for local-dev authentication with SpiceDB authorization.
- The repository must provide a non-Compose local SpiceDB path for developer machines where the Docker Compose plugin is unavailable.
- The repository must provide a local verification script for the first secure API flow.

## First Services

The first Compose topology includes:

- `app`: Stuff Stash API.
- `postgres`: local Postgres database.
- `spicedb`: local SpiceDB service.

## Configuration

- All service configuration must come from environment variables.
- Local defaults may be provided in Compose for developer ergonomics.
- Secrets used in local Compose must be clearly local-only.
- Production secrets must never be committed.

## Verification

- `make test` must pass without requiring Compose.
- Compose should be able to start the local service graph when Docker is available.
- The API health endpoint must remain available without authentication.
- Protected endpoints must require authentication even in local development.
- Running with unknown auth or authz modes must fail startup.
- The SpiceDB Compose path must bootstrap the checked-in schema automatically.
- The non-Compose SpiceDB path must bootstrap the checked-in schema automatically.
- The local verification script must cover health, unauthenticated rejection, authenticated identity, tenant creation, inventory creation, and inventory listing.
- The repository must provide an explicit real-SpiceDB adapter verification command that starts pinned local SpiceDB, runs the adapter integration tests, and cleans up.
