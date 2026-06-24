---
title: Self-Hosting
description: Run Stuff Stash with Docker Compose or a split API and web deployment.
---

Stuff Stash is built for self-hosting. The app is intentionally split into two
deployables:

- a Go API service,
- a SvelteKit web app.

The API owns domain behavior, auth checks, persistence, audit history, and media
metadata. The web app is a separate frontend that talks to the API through the
generated OpenAPI client boundary.

## Quick Evaluation

Docker Compose is the easiest way to run the local stack.

```sh
make compose-up-oidc
```

This starts:

- the API,
- Postgres,
- SpiceDB,
- a local Dex OIDC provider,
- the migration job.

Start the web app in another terminal:

```sh
make web-install
make web-dev
```

Open `http://localhost:5173` and sign in with the local Dex test user:

```text
Email: owner@example.com
Password: password
```

The local web config lives at `apps/web/static/config.json`. It points the web
app at the local API and Dex issuer.

## Deployment Shape

Stuff Stash does not ship as one binary because the UI and API are separate on
purpose.

For Docker Compose, run the API image, the web image, Postgres, SpiceDB, and an
OIDC provider.

For Kubernetes-shaped deployments, run the same API and web images as separate
workloads. Configure the web app with the public API base URL and OIDC client
settings. Configure the API with Postgres, SpiceDB, OIDC, and blob storage
environment variables.

## Required Services

| Service | Role |
| --- | --- |
| API | Go backend, migrations, REST API, audit, authorization checks |
| Web | SvelteKit frontend served separately from the API |
| Postgres | Primary metadata store |
| SpiceDB | Relationship-based authorization |
| OIDC provider | SSO sign-in, such as Google or another OIDC issuer |
| Blob storage | Local filesystem for development, S3-compatible storage for deployment |

## Configuration

Runtime configuration comes from environment variables. The most important API
settings are:

| Variable | Purpose |
| --- | --- |
| `STUFF_STASH_AUTH_MODE` | `oidc` for SSO-backed deployments |
| `STUFF_STASH_OIDC_ISSUER` | OIDC issuer URL |
| `STUFF_STASH_OIDC_CLIENT_ID` | API audience/client ID |
| `STUFF_STASH_AUTHZ_MODE` | `spicedb` for relationship-based authorization |
| `STUFF_STASH_SPICEDB_ENDPOINT` | SpiceDB gRPC endpoint |
| `STUFF_STASH_REPOSITORY_MODE` | `postgres` for durable persistence |
| `STUFF_STASH_DATABASE_DSN` | Postgres connection string |
| `STUFF_STASH_BLOB_STORAGE_MODE` | `filesystem` or `s3` |

The web app reads `config.json` at runtime, so the same web image can be pointed
at different API and OIDC endpoints.

## Verify A Local Stack

Run the API verification flow:

```sh
make verify-dex-oidc-api
```

That check uses real Dex-issued ID tokens, SpiceDB authorization, and the same
API user flow used by local development.

Run the web checks:

```sh
make web-test
make web-check
make web-build
```

Run the docs site:

```sh
make docs-install
make docs-dev
```
