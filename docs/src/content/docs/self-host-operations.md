---
title: Self-Host Operations
description: Secure, share, back up, and upgrade a Stuff Stash household deployment.
---

Start with [Self-Host Stuff Stash](../self-hosting/). Return here before using
real household data or opening the app to other devices.

## Trust The Local Certificate

Caddy creates one local certificate authority for the web app, API, Dex, and
Garage. Export its root certificate:

```sh
mkdir -p .stuffstash/selfhost/caddy
docker compose -f compose.selfhost.yaml cp caddy:/data/caddy/pki/authorities/local/root.crt .stuffstash/selfhost/caddy/root.crt
```

Import `root.crt` on every device that opens Stuff Stash.

<details>
<summary>macOS</summary>

```sh
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain .stuffstash/selfhost/caddy/root.crt
```

</details>

<details>
<summary>Ubuntu or Debian</summary>

```sh
sudo cp .stuffstash/selfhost/caddy/root.crt /usr/local/share/ca-certificates/stuffstash.crt
sudo update-ca-certificates
```

</details>

<details>
<summary>Fedora or RHEL</summary>

```sh
sudo cp .stuffstash/selfhost/caddy/root.crt /etc/pki/ca-trust/source/anchors/stuffstash.crt
sudo update-ca-trust
```

</details>

<details>
<summary>Windows</summary>

In an administrator PowerShell window:

```powershell
Import-Certificate -FilePath .\root.crt -CertStoreLocation Cert:\LocalMachine\Root
```

</details>

Firefox may use its own store. If the warning remains, open **Settings → Privacy
& Security → Certificates → View Certificates → Authorities**, then import the
root.

## Before Household Use

1. Create private [Dex users and clients](../dex-users/).
2. Replace these example values in `.env`:
   `POSTGRES_PASSWORD`, `SPICEDB_POSTGRES_PASSWORD`,
   `STUFF_STASH_SPICEDB_PRESHARED_KEY`, `STUFF_STASH_S3_ACCESS_KEY`,
   `STUFF_STASH_S3_SECRET_KEY`, and `STUFF_STASH_PROVIDER_CREDENTIAL_KEY`.
3. Run the strict check, then restart:

```sh
docker compose -f compose.selfhost.yaml down
./scripts/selfhost-preflight.sh
docker compose -f compose.selfhost.yaml up -d
```

Generate URL-safe passwords with `openssl rand -hex 24`. Generate the provider
credential key with `openssl rand -base64 32`. Keep `.env` and the private Dex
config out of Git.

## Use Stuff Stash On Your LAN

:::caution[Use DNS, not an IP address]
OIDC discovery requires a DNS hostname. A raw LAN IP is unsupported.
:::

1. Create a local DNS record, such as `stuffstash.home.arpa`, pointing to the server.
2. In `.env`, replace every `stuffstash.localhost` with that DNS name and set
   `STUFF_STASH_BIND_ADDRESS` to the server's LAN address. Use `0.0.0.0` only
   when the server address is not stable.
3. Make the same hostname changes in your private Dex config.
4. Allow TCP ports `8081`, `8080`, `5556`, and `3900` through the server firewall.
5. Stop, check, and start the stack:

```sh
docker compose -f compose.selfhost.yaml down
./scripts/selfhost-preflight.sh
docker compose -f compose.selfhost.yaml up -d
```

Trust the Caddy root on each client device.

The default bind address is `127.0.0.1`, so a new install is reachable only from
the server itself.

## What To Back Up

Back up `.env`, your private Dex config, and these Docker volumes together:

| Volume | Contents |
| --- | --- |
| `stuffstash_selfhost-postgres-data` | Inventory metadata and audit history |
| `stuffstash_selfhost-spicedb-postgres-data` | Authorization relationships |
| `stuffstash_selfhost-garage-meta` | Garage object metadata |
| `stuffstash_selfhost-garage-data` | Uploaded files |
| `stuffstash_selfhost-caddy-data` | Local CA and certificates |

Stop the stack before copying these volumes so Postgres and Garage are
consistent. Start it again afterward, and test restoration before relying on a
backup.

## Upgrade

1. Make a cold backup of the files and volumes listed above.
2. In an empty directory, download and verify the latest release bundle using
   the commands in [Download](../self-hosting/#1-download).
3. Copy the new `.env.example` to `.env`, then carry forward your changed values
   from the old `.env`. Move the private Dex config too. Do not overwrite the
   new example wholesale; a release may add required settings.
4. From the old directory, run `docker compose -f compose.selfhost.yaml down`.
5. From the new directory, run `./scripts/selfhost-preflight.sh`.
6. Run `docker compose -f compose.selfhost.yaml up -d`.

The Compose project name is fixed, so the new bundle reuses the existing data
volumes. Check the app and an uploaded image after every upgrade.

:::caution[Rollback needs the backup]
Database migrations may not run backward. Restore the pre-upgrade volumes before
starting the previous bundle.
:::

## Remove Everything

:::danger[This deletes household data]
After making any needed backup, remove containers and all named volumes:

```sh
docker compose -f compose.selfhost.yaml down -v
```
:::
