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

If you run Stuff Stash from another device on your LAN, pick one hostname or IP
and use it everywhere before the first start. OIDC is strict: the issuer, web
redirect URI, API origin, and browser-visible URLs must agree.

<details>
<summary>LAN or reverse proxy example</summary>

Use a home DNS name if you have one. A LAN IP also works, but DNS is easier to
remember and easier to change later. This example uses `stash.home.arpa`.

In `.env`, change every browser-facing value to the same host:

```text
STUFF_STASH_SELFHOST_HOSTNAME=stash.home.arpa
STUFF_STASH_WEB_ORIGIN=https://stash.home.arpa:8081
STUFF_STASH_API_ORIGIN=https://stash.home.arpa:8080
STUFF_STASH_OIDC_ISSUER=https://stash.home.arpa:5556/dex
STUFF_STASH_WEB_OIDC_REDIRECT_URI=https://stash.home.arpa:8081/callback
STUFF_STASH_CORS_ALLOWED_ORIGINS=https://stash.home.arpa:8081
STUFF_STASH_S3_ENDPOINT=stash.home.arpa:3900
STUFF_STASH_S3_PUBLIC_ENDPOINT=stash.home.arpa:3900
```

In your private Dex config, keep the same host:

```yaml
issuer: https://stash.home.arpa:5556/dex

web:
  allowedOrigins:
    - https://stash.home.arpa:8081

staticClients:
  - id: stuff-stash-web-local
    public: true
    redirectURIs:
      - https://stash.home.arpa:8081/callback
```

If you use a reverse proxy in front of the Compose host, configure the proxy for
the public URLs people will actually open, then put those exact URLs in `.env`
and Dex. Do not mix `localhost`, a LAN IP, and a public DNS name in one
deployment.

The bundled Caddy service is already the HTTPS edge for the Compose stack. For a
simple home deployment, start there before adding another proxy.

</details>

## Verify Startup

After the stack starts, check the pieces from the machine running Docker:

```sh
curl -k https://stuffstash.localhost:8081/
curl -k https://stuffstash.localhost:8080/healthz
curl -k https://stuffstash.localhost:5556/dex/.well-known/openid-configuration
curl -k https://stuffstash.localhost:3900/
```

The API health check should return:

```json
{"service":"stuff-stash","status":"healthy"}
```

Garage should return an anonymous `AccessDenied` response. That is expected; the
browser and API use signed requests for uploads and thumbnails.

You can also check Compose state:

```sh
docker compose -f compose.selfhost.yaml ps
```

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

## Back Up And Restore

For a small home deployment, the simplest backup is a cold volume backup: stop
the stack, archive the volumes, copy `.env`, and copy your private Dex config.

<details>
<summary>Cold backup commands</summary>

Run from the repository root:

```sh
backup_dir=".stuffstash/backups/$(date -u +%Y%m%dT%H%M%SZ)"
compose_project="${COMPOSE_PROJECT_NAME:-$(basename "$PWD")}"
mkdir -p "$backup_dir"

set -a
. ./.env
set +a

docker compose -f compose.selfhost.yaml down

for volume in \
  selfhost-postgres-data \
  selfhost-spicedb-postgres-data \
  selfhost-garage-meta \
  selfhost-garage-data \
  selfhost-caddy-data
do
  docker run --rm \
    -v "${compose_project}_${volume}:/source:ro" \
    -v "$PWD/$backup_dir:/backup" \
    "$POSTGRES_IMAGE" \
    tar -C /source -czf "/backup/${volume}.tgz" .
done

cp .env "$backup_dir/env"
cp "${DEX_CONFIG_PATH:-deploy/selfhost/dex/config.yaml}" "$backup_dir/dex-config.yaml"

docker compose -f compose.selfhost.yaml up -d
```

If you changed the Compose project name or run from a differently named
directory, confirm the real volume names with:

```sh
docker volume ls | grep selfhost
```

</details>

<details>
<summary>Restore commands</summary>

Restore only while the stack is stopped. This replaces the current self-host
data with the backup.

```sh
backup_dir=".stuffstash/backups/20260101T000000Z"
compose_project="${COMPOSE_PROJECT_NAME:-$(basename "$PWD")}"

cp "$backup_dir/env" .env

set -a
. ./.env
set +a

docker compose -f compose.selfhost.yaml down -v

for volume in \
  selfhost-postgres-data \
  selfhost-spicedb-postgres-data \
  selfhost-garage-meta \
  selfhost-garage-data \
  selfhost-caddy-data
do
  docker volume create "${compose_project}_${volume}"
  docker run --rm \
    -v "${compose_project}_${volume}:/target" \
    -v "$PWD/$backup_dir:/backup:ro" \
    "$POSTGRES_IMAGE" \
    tar -C /target -xzf "/backup/${volume}.tgz"
done

dex_config_path="${DEX_CONFIG_PATH:-deploy/selfhost/dex/config.yaml}"
mkdir -p "$(dirname "$dex_config_path")"
cp "$backup_dir/dex-config.yaml" "$dex_config_path"
chmod 600 "$dex_config_path"

docker compose -f compose.selfhost.yaml up -d
```

Sign in and check one inventory, one item, and one uploaded photo before you
trust the restore.

</details>

## Upgrade

Before every upgrade, take a backup and record the image digests you are running:

```sh
mkdir -p .stuffstash/backups
grep -E '^(STUFF_STASH_API_IMAGE|STUFF_STASH_WEB_IMAGE)=' .env \
  > ".stuffstash/backups/images-before-$(date -u +%Y%m%dT%H%M%SZ).txt"
```

Then update the checkout and copy the new published image digest lines from
`.env.example` into your private `.env`:

```sh
git pull --ff-only
grep -E '^(STUFF_STASH_API_IMAGE|STUFF_STASH_WEB_IMAGE)=' .env.example
```

Start the stack and let migrations finish:

```sh
docker compose -f compose.selfhost.yaml up -d
docker compose -f compose.selfhost.yaml ps
```

Verify startup, sign in, and check an existing item and photo.

If the new images fail before you add new data, put the old image digest lines
back in `.env` and run:

```sh
docker compose -f compose.selfhost.yaml up -d
```

If migrations already ran and the old images do not work, restore the backup you
took before upgrading.

## Operations Notes

- Back up `selfhost-postgres-data`, `selfhost-spicedb-postgres-data`,
  `selfhost-garage-meta`, `selfhost-garage-data`, and `selfhost-caddy-data`.
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
