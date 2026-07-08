---
title: Self-Host Stuff Stash
description: Set up Stuff Stash with Docker Compose, Caddy HTTPS, Dex OIDC, Postgres, SpiceDB, and Garage.
---

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

:::caution[Replace the defaults]
The checked-in passwords, Dex users, object-storage keys, and encryption key are
only first-run examples. Change them before you put real household data in Stuff
Stash.
:::

Start the stack:

```sh
docker compose -f compose.selfhost.yaml up
```

Open [https://stuffstash.localhost:8081](https://stuffstash.localhost:8081).

The Compose stack uses Caddy for browser-facing HTTPS. With the default local
hostname, your browser may ask you to accept or trust Caddy's local certificate
authority.

Sign in with the first-run Dex account:

```text
Email: owner@example.com
Password: password
```

Then follow [First Inventory](../first-inventory/).

To build API and web images from your checkout instead of using published
release images, use the contributor override:

```sh
docker compose -f compose.selfhost.yaml -f compose.selfhost.build.yaml up --build
```

## Hostnames And Ports

The default `.env.example` uses `stuffstash.localhost` so the browser and Docker
containers can agree on the same OIDC issuer:

| URL | Service |
| --- | --- |
| `https://stuffstash.localhost:8081` | Web app |
| `https://stuffstash.localhost:8080` | API |
| `https://stuffstash.localhost:5556/dex` | Dex issuer |
| `https://stuffstash.localhost:3900` | Garage S3 API |

If you run Stuff Stash from another device on your LAN, replace every
`stuffstash.localhost` value in `.env` and your Dex config with the same LAN IP
or DNS name before starting the stack. OIDC is strict: the issuer, web redirect
URI, API origin, and browser-visible URLs must agree.

For a public DNS name, point the name at the server and update `.env` and Dex
before starting the stack.

## Local HTTPS Certificate

Caddy creates a local certificate authority for `stuffstash.localhost`. To avoid
browser certificate errors across the web app, API, Dex, and Garage, trust that
root certificate on the device running the browser:

```sh
mkdir -p .stuffstash/selfhost/caddy
docker compose -f compose.selfhost.yaml cp caddy:/data/caddy/pki/authorities/local/root.crt .stuffstash/selfhost/caddy/root.crt
```

Then import `.stuffstash/selfhost/caddy/root.crt` into your operating system or
browser trust store.

## Replace Defaults

Do not run with the checked-in defaults outside a first local trial. Replace:

| Setting | Why it matters |
| --- | --- |
| Dex users and static clients | The committed Dex config has first-run accounts only. See [Dex Users And Clients](../dex-users/). |
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

Follow [Dex Users And Clients](../dex-users/) before you rely on the deployment.
The bundled Compose topology uses Dex static users from a private config file;
it does not include a Dex user-management UI.

## Verify Persistence

After creating a household, inventory, item, and photo, restart without deleting
volumes:

```sh
docker compose -f compose.selfhost.yaml down
docker compose -f compose.selfhost.yaml up
```

Sign in again. Your inventory and uploaded media should still be available.

To remove containers and all self-host data volumes:

```sh
docker compose -f compose.selfhost.yaml down -v
```

## Operations Notes

- Back up `selfhost-postgres-data`, `selfhost-spicedb-postgres-data`,
  `selfhost-garage-meta`, and `selfhost-garage-data`.
- `selfhost-postgres-data` contains Stuff Stash metadata.
- `selfhost-spicedb-postgres-data` contains authorization relationships.
- `selfhost-garage-meta` and `selfhost-garage-data` contain uploaded media
  metadata and object bytes.
- `selfhost-caddy-data` contains the local Caddy certificate authority and
  certificates.
- Keep `.env` and your private Dex config out of Git.
- Use URL-safe database passwords, or percent-encode reserved characters in
  connection strings.
- Use a real DNS name and publicly trusted certificate before exposing the
  deployment outside your machine.
- Record image versions before upgrades and run the stack after each upgrade so
  migrations can complete.
