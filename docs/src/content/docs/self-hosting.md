---
title: Self-Host Stuff Stash
description: Set up Stuff Stash with Docker Compose, OIDC, Postgres, SpiceDB, and Garage.
---

Stuff Stash is designed to run as separate services: a web app, an API,
Postgres, SpiceDB, an OIDC provider, and S3-compatible media storage.

The goal is a Docker Compose path that a homeowner can run on a home server and
trust after restarts. Use the production-like Compose path for that shape. The
older local Compose path remains useful for contributor evaluation with local
Dex fixtures.

Use this page in two ways:

- If you are preparing a real household deployment, start with the
  production-like Compose path.
- If you want the local Dex test accounts and a Vite development server, use the
  Compose evaluation path.

## Target Self-Hosted Shape

A real self-hosted deployment needs these pieces:

| Service | Purpose | Durable state |
| --- | --- | --- |
| Web | SvelteKit static app | Runtime `config.json` |
| API | REST API, auth checks, audit, workers | None by itself |
| Postgres | Inventory metadata, audit, outboxes | Postgres data volume |
| SpiceDB | Relationship-based authorization | SpiceDB Postgres datastore |
| OIDC provider | Sign-in | Provider-specific |
| Garage | S3-compatible media storage | Garage metadata and data volumes |

All browser-visible URLs must agree:

- `STUFF_STASH_WEB_ORIGIN`: where the browser opens the web app.
- `STUFF_STASH_API_ORIGIN`: where the browser reaches the API.
- `STUFF_STASH_DEX_ISSUER` or your external OIDC issuer: the issuer both the
  browser and API can reach.
- `STUFF_STASH_S3_PUBLIC_ENDPOINT`: the Garage or S3 endpoint the browser can
  reach for direct uploads.

For local LAN testing, these are usually based on your server IP, such as
`http://192.168.1.50:5173`, not `localhost`.

## Before You Rely On It

Change these before a household relies on the deployment:

| Setting | Why it matters |
| --- | --- |
| `POSTGRES_PASSWORD` | Protects the metadata database. |
| `STUFF_STASH_S3_ACCESS_KEY` | Garage/S3 access key used by the API. |
| `STUFF_STASH_S3_SECRET_KEY` | Garage/S3 secret key used by the API. |
| `STUFF_STASH_PROVIDER_CREDENTIAL_KEY_ID` | Labels the active credential encryption key. |
| `STUFF_STASH_PROVIDER_CREDENTIAL_KEY` | Base64-encoded 32-byte AES-GCM key for provider credentials and temporary import material. |
| OIDC clients and users | Local Dex users are fixtures, not household accounts. |
| SpiceDB credentials and datastore | `serve-testing` is local-only and not durable. |
| Public origins and callback URLs | OIDC fails when issuer, redirect URI, API origin, or web origin disagree. |

Generate the provider credential key with:

```sh
openssl rand -base64 32
```

Store secrets in a private `.env` file or your secret manager. Do not commit
household secrets.

## Production-Like Compose

This path runs the API, web app, Postgres metadata database, datastore-backed
SpiceDB, Garage, migrations, and Garage CORS setup from Docker Compose.

It does not run an OIDC provider for you. Use a real OIDC issuer, such as your
existing identity provider, and create a public web client with this redirect
URI:

```text
http://localhost:8081/callback
```

For a LAN or reverse-proxied setup, use the matching public web origin instead.

Clone the repository:

```sh
git clone https://github.com/elsell/stuffstash.git
cd stuffstash
```

Create a private environment file:

```sh
cp .env.example .env
```

Edit `.env` before starting. At minimum, replace:

- `STUFF_STASH_OIDC_ISSUER`
- `STUFF_STASH_OIDC_CLIENT_ID`
- `STUFF_STASH_WEB_OIDC_CLIENT_ID`
- `STUFF_STASH_WEB_OIDC_REDIRECT_URI`
- `STUFF_STASH_OIDC_CLIENT_IDS`
- `POSTGRES_PASSWORD`
- `STUFF_STASH_SPICEDB_PRESHARED_KEY`
- `SPICEDB_POSTGRES_PASSWORD`
- `STUFF_STASH_S3_ACCESS_KEY`
- `STUFF_STASH_S3_SECRET_KEY`
- `STUFF_STASH_PROVIDER_CREDENTIAL_KEY`

Generate the provider credential key with:

```sh
openssl rand -base64 32
```

Start the stack:

```sh
docker compose -f compose.selfhost.yaml up --build
```

Open the configured `STUFF_STASH_WEB_ORIGIN`.

The web container generates `/config.json` at startup from `.env`, including the
public API URL, OIDC issuer, web client ID, redirect URI, and upload limit. It
also renders its CSP from the configured API, OIDC, and browser-reachable
Garage/S3 origins.

The API allows browser CORS only from `STUFF_STASH_WEB_ORIGIN`. The
`garage-cors` service configures the Garage bucket CORS policy for the same
origin before the web service is started.

For a trusted LAN test with plain HTTP, the default Garage settings use:

```text
STUFF_STASH_S3_PUBLIC_ENDPOINT=localhost:3900
STUFF_STASH_S3_SECURE=false
```

For a reverse-proxied setup, put Garage behind HTTPS and set:

```text
STUFF_STASH_S3_PUBLIC_ENDPOINT=storage.example.test
STUFF_STASH_S3_SECURE=true
```

Stop containers but keep volumes:

```sh
docker compose -f compose.selfhost.yaml down
```

Remove containers and self-host data volumes:

```sh
docker compose -f compose.selfhost.yaml down -v
```

The self-host volumes are:

| Volume | Contains |
| --- | --- |
| `selfhost-postgres-data` | Stuff Stash metadata, audit, and outboxes. |
| `selfhost-spicedb-postgres-data` | SpiceDB authorization datastore. |
| `selfhost-garage-meta` | Garage object metadata. |
| `selfhost-garage-data` | Garage object data. |

Redis and Valkey are intentionally not part of this stack. The current
in-memory rate limiter is enough for a single API replica.

## Compose Evaluation

This path proves the API, web app, Postgres, Dex OIDC, SpiceDB authorization,
and Garage media storage on one machine. It builds from source and runs the web
app with Vite.

It is not a complete self-hosting runbook yet because:

- SpiceDB uses `serve-testing`, so authorization state is not durable.
- SQLite is supported by the API runtime, but there is no SQLite Compose
  self-hosting topology.

You need Docker with Compose, Node.js, and pnpm through Corepack or an existing
pnpm install.

Clone the repository:

```sh
git clone https://github.com/elsell/stuffstash.git
cd stuffstash
```

Start the current evaluation stack:

```sh
docker compose -f compose.yaml -f compose.oidc.yaml up --build
```

This starts the API, Postgres, SpiceDB, Garage, Dex, and the migration job.
Garage is the local S3-compatible media store.

If a port is already in use, choose alternate host ports:

```sh
STUFF_STASH_HTTP_PORT=18080 \
DEX_HTTP_PORT=15556 \
GARAGE_S3_PORT=13900 \
STUFF_STASH_S3_PUBLIC_ENDPOINT=localhost:13900 \
POSTGRES_PORT=15432 \
SPICEDB_GRPC_PORT=15051 \
docker compose -f compose.yaml -f compose.oidc.yaml up --build
```

If you change the API or Dex port, update `apps/web/static/config.json` before
starting the web app. If you change the Garage browser-facing port, set
`STUFF_STASH_S3_PUBLIC_ENDPOINT` to the same host and port before starting the
API and before applying the Garage CORS policy below.

Start the web app in another terminal:

```sh
corepack pnpm install --frozen-lockfile
corepack pnpm --dir apps/web dev --host 0.0.0.0
```

If `corepack enable` fails on your machine, it usually means Corepack tried to
write global shims under a system directory. You can skip that step when
`corepack pnpm --version` or `pnpm --version` already returns `11.0.7`.

Open `http://localhost:5173` and sign in:

```text
Email: owner@example.com
Password: password
```

Then follow [First Inventory](../first-inventory/).

## LAN OIDC Evaluation

Use this when the browser or mobile app is not running on the same machine as
Docker. Replace `192.168.1.50` with your server's LAN IP:

```sh
export STUFF_STASH_WEB_ORIGIN=http://192.168.1.50:5173
export STUFF_STASH_API_ORIGIN=http://192.168.1.50:8080
export STUFF_STASH_DEX_ISSUER=http://192.168.1.50:5556/dex
export STUFF_STASH_DEX_HTTP_ADDR=0.0.0.0:5556
export STUFF_STASH_OIDC_MOBILE_REDIRECT_URI=stuffstash://auth/callback
node scripts/render-local-dex-config.mjs
```

Start Compose with the generated Dex config and matching API issuer:

```sh
DEX_CONFIG_PATH=.stuffstash/local/dex/config.yaml \
STUFF_STASH_CORS_ALLOWED_ORIGINS="$STUFF_STASH_WEB_ORIGIN" \
STUFF_STASH_OIDC_ISSUER="$STUFF_STASH_DEX_ISSUER" \
docker compose -f compose.yaml -f compose.oidc.yaml up --build
```

Start the web app for the same LAN origin:

```sh
VITE_STUFF_STASH_WEB_ORIGIN="$STUFF_STASH_WEB_ORIGIN" \
corepack pnpm --dir apps/web dev --host 0.0.0.0
```

Check mobile auth metadata before opening a mobile app:

```sh
curl "$STUFF_STASH_API_ORIGIN/.well-known/stuff-stash/mobile-auth"
```

The response should advertise the same issuer you configured in
`STUFF_STASH_DEX_ISSUER`.

## Garage Browser Uploads

Browser direct upload needs two things:

- `STUFF_STASH_S3_PUBLIC_ENDPOINT` must be reachable from the browser.
- The Garage bucket must allow CORS for the web origin.

For local Garage, configure CORS with an S3-compatible CLI after Garage is
running. Use the same browser-facing endpoint that the API uses in
`STUFF_STASH_S3_PUBLIC_ENDPOINT`:

```sh
export STUFF_STASH_S3_ACCESS_KEY="${STUFF_STASH_S3_ACCESS_KEY:-GK0123456789abcdef0123456789abcdef}"
export STUFF_STASH_S3_SECRET_KEY="${STUFF_STASH_S3_SECRET_KEY:-0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef}"
export AWS_DEFAULT_REGION="${STUFF_STASH_S3_REGION:-garage}"
export STUFF_STASH_WEB_ORIGIN="${STUFF_STASH_WEB_ORIGIN:-http://localhost:5173}"
export STUFF_STASH_S3_PUBLIC_ENDPOINT="${STUFF_STASH_S3_PUBLIC_ENDPOINT:-localhost:3900}"
export STUFF_STASH_S3_BUCKET="${STUFF_STASH_S3_BUCKET:-stuffstash}"
export AWS_ACCESS_KEY_ID="$STUFF_STASH_S3_ACCESS_KEY"
export AWS_SECRET_ACCESS_KEY="$STUFF_STASH_S3_SECRET_KEY"

cat > /tmp/stuffstash-garage-cors.json <<EOF
{
  "CORSRules": [
    {
      "AllowedHeaders": ["*"],
      "AllowedMethods": ["GET", "POST"],
      "AllowedOrigins": ["$STUFF_STASH_WEB_ORIGIN"],
      "ExposeHeaders": ["ETag"]
    }
  ]
}
EOF

aws --endpoint-url "http://$STUFF_STASH_S3_PUBLIC_ENDPOINT" \
  s3api put-bucket-cors \
  --bucket "$STUFF_STASH_S3_BUCKET" \
  --cors-configuration file:///tmp/stuffstash-garage-cors.json
```

For production, put Garage or your S3-compatible storage behind TLS and use the
public HTTPS storage endpoint in `STUFF_STASH_S3_PUBLIC_ENDPOINT`.

## Verify The Stack

Check API health:

```sh
curl http://localhost:8080/healthz
```

Check OIDC/API authorization:

```sh
scripts/verify-dex-oidc-api.sh
```

Check browser behavior. These commands are contributor checks for the web app;
they are useful before treating a local setup guide change as verified, but they
are not required to run Stuff Stash:

```sh
corepack pnpm --dir apps/web test
corepack pnpm --dir apps/web check:shadcn
corepack pnpm --dir apps/web check
corepack pnpm --dir apps/web build
```

After signing in, create a household, create an inventory, add a location, add
an item, and upload a small JPEG or PNG. Then reload the page.

Do not use a successful page reload as proof of restart durability. Until the
Compose stack uses persistent SpiceDB storage, a full stack restart can make
previously created inventories inaccessible to the same user.

## Stop And Clean Up

Stop containers but keep volumes:

```sh
docker compose -f compose.yaml -f compose.oidc.yaml down
```

Remove containers and local data volumes:

```sh
docker compose -f compose.yaml -f compose.oidc.yaml down -v
```

The main local evaluation volumes are:

| Volume | Contains |
| --- | --- |
| `postgres-data` | Stuff Stash metadata, audit, and outboxes. |
| `garage-meta` | Garage object metadata. |
| `garage-data` | Garage object data. |

The production-like Compose path uses a separate durable SpiceDB datastore
volume.

## Production Readiness

Before publishing Stuff Stash beyond a trusted LAN, add:

- TLS and a reverse proxy for the web app, API, OIDC issuer, and Garage/S3
  endpoint.
- Real OIDC provider configuration and reviewed callback URLs.
- Private, rotated secrets.
- Backups for Postgres, SpiceDB, and Garage.
- An upgrade process that records image versions and runs migrations.

The split-container deployment shape is the long-term target, but a complete
production runbook is still in progress.
