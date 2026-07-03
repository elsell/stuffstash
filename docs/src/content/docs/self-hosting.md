---
title: Run Stuff Stash
description: Pick the self-hosting path that fits your setup.
---

Stuff Stash runs as a web app and an API. Docker Compose is the fastest way to
evaluate the full stack on one machine.

## Pick A Path

<div class="ss-card-grid">
  <section class="ss-card">
    <span class="ss-card-kicker">Recommended</span>
    <strong>Docker Compose</strong>
    <span>Best for trying Stuff Stash on a laptop, mini PC, or home server.</span>
  </section>
  <section class="ss-card">
    <span class="ss-card-kicker">Topology</span>
    <strong>Split containers</strong>
    <span>Understand the API, web, database, auth, authorization, and storage pieces.</span>
  </section>
  <section class="ss-card">
    <span class="ss-card-kicker">Advanced</span>
    <strong>Kubernetes-shaped</strong>
    <span>Use the same split API and web images as separate workloads.</span>
  </section>
</div>

## Recommended: Docker Compose

**Best for:** evaluation and local self-hosting experiments.

**You need:** Docker with Compose, Node.js with Corepack, an available API port,
and an available web port.

Start the local stack:

```sh
docker compose -f compose.yaml -f compose.oidc.yaml up --build
```

This starts the API, Postgres, SpiceDB, Garage, Dex, and the migration job.
Garage is the local S3-compatible media store.

Start the web app in another terminal:

```sh
corepack enable
corepack pnpm install --frozen-lockfile
corepack pnpm --dir apps/web dev --host 0.0.0.0
```

Open `http://localhost:5173` and sign in:

```text
Email: owner@example.com
Password: password
```

Then follow [First Inventory](../first-inventory/).

The default web config points at `http://localhost:8080`. If you change the API
host port, update `apps/web/static/config.json` before starting the web app.

To stop the stack:

```sh
docker compose -f compose.yaml -f compose.oidc.yaml down
```

## Deployment Shape: Split Containers

**Best for:** planning a home-server or lab deployment.

This is the deployment shape, not a complete production runbook.

Run these pieces:

| Service | What it does |
| --- | --- |
| API | Go backend, migrations, REST API, audit, auth checks |
| Web | SvelteKit frontend served separately from the API |
| Postgres | Primary metadata store |
| SpiceDB | Relationship-based authorization |
| OIDC provider | SSO sign-in |
| Blob storage | Garage or another S3-compatible media store |

The API reads configuration from environment variables. The web app reads
`config.json` at runtime, so the same web image can point at different API and
OIDC endpoints.

Before using this as a production deployment, choose reviewed image references
pinned by immutable digest, configure public URLs and CORS, mount or generate
the web `config.json`, enable the right OIDC and SpiceDB settings, and configure
durable Garage or S3-compatible media storage.

See the [Configuration Reference](../configuration/) for the full runtime
settings list.

## Advanced: Kubernetes-Shaped

**Best for:** operators who already run home or lab services on Kubernetes.

This is topology guidance, not a ready-made manifest set.

Run the API and web images as separate workloads. Provide Postgres, SpiceDB, an
OIDC issuer, and blob storage as external services or cluster workloads. Configure
the web app with the public API URL and OIDC client settings.

The project shape fits Kubernetes, but Docker Compose remains the quickest
evaluation path.

## Important Settings

This is the short list most deployments need first. The full list is in the
[Configuration Reference](../configuration/).

| Variable | Purpose |
| --- | --- |
| `STUFF_STASH_AUTH_MODE` | Use `oidc` for SSO-backed deployments |
| `STUFF_STASH_OIDC_ISSUER` | OIDC issuer URL |
| `STUFF_STASH_OIDC_CLIENT_ID` | API audience/client ID |
| `STUFF_STASH_AUTHZ_MODE` | Use `spicedb` for relationship-based authorization |
| `STUFF_STASH_SPICEDB_ENDPOINT` | SpiceDB gRPC endpoint |
| `STUFF_STASH_REPOSITORY_MODE` | Use `postgres` for durable persistence |
| `STUFF_STASH_DATABASE_DSN` | Postgres connection string |
| `STUFF_STASH_BLOB_STORAGE_MODE` | Use `s3` for Garage-backed media storage |
| `STUFF_STASH_S3_ENDPOINT` | API-reachable Garage or S3 endpoint |
| `STUFF_STASH_S3_PUBLIC_ENDPOINT` | Browser-reachable Garage or S3 endpoint for direct uploads |
| `STUFF_STASH_S3_BUCKET` | Garage or S3 bucket name |

## Verify The Stack

Run the OIDC and authorization verification flow:

```sh
make verify-dex-oidc-api
```

Run web checks:

```sh
make web-test
make web-check
make web-build
```
