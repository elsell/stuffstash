---
title: Self-Host Stuff Stash
description: Set up Stuff Stash with Docker Compose, Dex OIDC, Postgres, SpiceDB, and Garage.
---

This is the main way to run Stuff Stash yourself.

Docker Compose starts the web app, API, Dex sign-in, Postgres, SpiceDB, Garage,
migrations, and Garage CORS setup. The stack is designed to survive normal
container restarts as long as you keep the Docker volumes.

## What Runs

| Service | Purpose | Durable state |
| --- | --- | --- |
| Web | Static SvelteKit app served by the web container | Runtime `config.json` |
| API | REST API, auth checks, audit, workers | None by itself |
| Dex | OIDC sign-in | User config file |
| Postgres | Inventory metadata, audit, outboxes | `selfhost-postgres-data` |
| SpiceDB | Relationship-based authorization | `selfhost-spicedb-postgres-data` |
| Garage | S3-compatible media storage | `selfhost-garage-meta`, `selfhost-garage-data` |

The web app is served by Docker Compose. You do not need to run a Vite
development server for self-hosting.

## Quick Start

Clone the repository:

```sh
git clone https://github.com/elsell/stuffstash.git
cd stuffstash
```

Create a private environment file:

```sh
cp .env.example .env
```

Start the stack:

```sh
docker compose -f compose.selfhost.yaml up --build
```

Open:

```text
http://stuffstash.localhost:8081
```

Sign in with the first-run Dex account:

```text
Email: owner@example.com
Password: password
```

Then follow [First Inventory](../first-inventory/).

## Hostnames And Ports

The default `.env.example` uses `stuffstash.localhost` so the browser and Docker
containers can agree on the same OIDC issuer:

| URL | Service |
| --- | --- |
| `http://stuffstash.localhost:8081` | Web app |
| `http://stuffstash.localhost:8080` | API |
| `http://stuffstash.localhost:5556/dex` | Dex issuer |
| `http://stuffstash.localhost:3900` | Garage S3 API |

If you run Stuff Stash from another device on your LAN, replace every
`stuffstash.localhost` value in `.env` and your Dex config with the same LAN IP
or DNS name before starting the stack. OIDC is strict: the issuer, web redirect
URI, API origin, and browser-visible URLs must agree.

For a reverse proxy, use HTTPS origins and route the web app, API, Dex issuer,
and Garage endpoint through the proxy.

## Before Relying On It

The defaults are meant to get you running. Replace these before a household
relies on the deployment:

| Setting | Why it matters |
| --- | --- |
| Dex users and static clients | The committed Dex config has first-run accounts only. |
| `POSTGRES_PASSWORD` | Protects inventory metadata. |
| `SPICEDB_POSTGRES_PASSWORD` | Protects authorization state. |
| `STUFF_STASH_SPICEDB_PRESHARED_KEY` | Protects the SpiceDB API. |
| `STUFF_STASH_S3_ACCESS_KEY` | Garage/S3 access key used by the API. |
| `STUFF_STASH_S3_SECRET_KEY` | Garage/S3 secret key used by the API. |
| `STUFF_STASH_PROVIDER_CREDENTIAL_KEY` | Encrypts provider credentials and temporary import material. |

Generate the provider credential key with:

```sh
openssl rand -base64 32
```

Store secrets in `.env` or your secret manager. Do not commit household secrets.

To customize Dex users, copy the bundled config to a private path and point
`.env` at it:

```sh
mkdir -p .stuffstash/selfhost/dex
cp deploy/selfhost/dex/config.yaml .stuffstash/selfhost/dex/config.yaml
```

Set:

```text
DEX_CONFIG_PATH=.stuffstash/selfhost/dex/config.yaml
```

Then edit the copied Dex config. If you change the hostname, update the Dex
`issuer`, `allowedOrigins`, and redirect URIs to match `.env`.

## Verify Persistence

After creating a household, inventory, item, and photo, restart without deleting
volumes:

```sh
docker compose -f compose.selfhost.yaml down
docker compose -f compose.selfhost.yaml up --build
```

Sign in again. Your inventory and uploaded media should still be available.

To remove containers and all self-host data volumes:

```sh
docker compose -f compose.selfhost.yaml down -v
```

## Operations Notes

- Back up Postgres, SpiceDB Postgres, and Garage volumes.
- Keep `.env` and your private Dex config out of Git.
- Use URL-safe database passwords, or percent-encode reserved characters in
  connection strings.
- Put the web app, API, Dex, and Garage behind TLS before exposing them beyond a
  trusted LAN.
- Record image versions before upgrades and run the stack after each upgrade so
  migrations can complete.
