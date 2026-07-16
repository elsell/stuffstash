---
title: Self-Host Operations
description: Manage certificates, users, backups, and upgrades for Stuff Stash.
---

Start with [Self-Host Stuff Stash](../self-hosting/). Use this page when you
need to change or maintain the installation.

## Trust The Local Certificate

Caddy creates one local certificate authority for the web app, API, Dex, and
Garage. Export its root certificate:

```sh
mkdir -p .stuffstash/selfhost/caddy
docker compose -f compose.selfhost.yaml cp caddy:/data/caddy/pki/authorities/local/root.crt .stuffstash/selfhost/caddy/root.crt
```

Copy `root.crt` to each device that opens Stuff Stash, then import it using the
matching instructions below.

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

Run in an administrator PowerShell window:

```powershell
Import-Certificate -FilePath .\root.crt -CertStoreLocation Cert:\LocalMachine\Root
```

</details>

Firefox may use its own certificate store. If the warning remains, open
**Settings → Privacy & Security → Certificates → View Certificates →
Authorities**, then import the root.

## Replace The Example Credentials

The defaults are reasonable for a trusted home network, but they are public.
Replace them before exposing Stuff Stash to a wider network.

Do this before adding data because the clean reset removes the example stack:

1. Stop and reset with `docker compose -f compose.selfhost.yaml down -v`.
2. Replace the [Dex users and clients](../dex-users/).
3. Replace every `change-me-` value in `.env` and generate a provider key with
   `openssl rand -base64 32`.
4. Run `./scripts/selfhost-preflight.sh --strict`.
5. Start with `docker compose -f compose.selfhost.yaml up -d`, then trust its
   new Caddy root.

Keep `.env` and the private Dex config out of Git. Do not change database or
Garage credentials on an installation with data unless you also rotate them in
those services.

## Keep The Address Stable

Reserve the server's IPv4 address in your router so bookmarks, certificates,
and OIDC callbacks do not change. If the address changes before you have data,
stop the stack, remove `.env` and `.stuffstash`, and run the setup again with
the new address.

### Optional DNS Name

A local DNS name is optional. Stop the stack, replace the LAN IP everywhere in
`.env` and the private Dex config, run preflight, then start it again. The name
must resolve to the server on every client device; trust the new Caddy root if
the certificate authority changed.

## What To Back Up

Back up `.env`, your private Dex config, and these Docker volumes together:

| Volume | Contents |
| --- | --- |
| `stuffstash_selfhost-postgres-data` | Inventory metadata and audit history |
| `stuffstash_selfhost-spicedb-postgres-data` | Authorization relationships |
| `stuffstash_selfhost-garage-meta` | Garage object metadata |
| `stuffstash_selfhost-garage-data` | Uploaded files |
| `stuffstash_selfhost-caddy-data` | Local CA and certificates |

Stop the stack before copying the volumes. Start it afterward, and test a
restore before relying on the backup.

## Upgrade

1. Back up the files and volumes above.
2. Download and verify the new bundle in an empty directory.
3. Copy the new `.env.example` to `.env`, then carry over your changed values;
   do not replace the new file wholesale. Move the private Dex config too.
4. Stop the old bundle with `docker compose -f compose.selfhost.yaml down`.
5. Run `./scripts/selfhost-preflight.sh`, then start the new bundle.

The fixed Compose project name reuses the existing volumes. Check the app and
an uploaded image after every upgrade. Database migrations may require the
pre-upgrade backup to roll back.

## Remove Everything

:::danger[This deletes household data]
After making any needed backup, remove containers and all named volumes:

```sh
docker compose -f compose.selfhost.yaml down -v
```
:::
