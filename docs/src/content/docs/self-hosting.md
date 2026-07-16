---
title: Self-Host Stuff Stash
description: Get Stuff Stash running with the supported release bundle.
---

The release bundle runs Docker Compose, Caddy HTTPS, Dex OIDC, Postgres,
SpiceDB, Garage, and the web app. You need Docker with Compose and `curl`.

## 1. Download

```sh
curl -fLO https://github.com/elsell/stuffstash/releases/latest/download/stuffstash-selfhost.tar.gz
curl -fLO https://github.com/elsell/stuffstash/releases/latest/download/stuffstash-selfhost.tar.gz.sha256
sha256sum -c stuffstash-selfhost.tar.gz.sha256
tar -xzf stuffstash-selfhost.tar.gz
cd stuffstash-selfhost
```

:::note[macOS]
Use `shasum -a 256 -c stuffstash-selfhost.tar.gz.sha256` for the checksum.
:::

## 2. Check The Setup

```sh
cp .env.example .env
./scripts/selfhost-preflight.sh --trial
```

Trial mode allows the example users and secrets. It is only for evaluation.

## 3. Start

```sh
docker compose -f compose.selfhost.yaml up -d
```

Before opening the app, [trust Caddy's local certificate
authority](../self-host-operations/#trust-the-local-certificate). Then open
[https://stuffstash.localhost:8081](https://stuffstash.localhost:8081) and sign
in:

```text
Email: owner@example.com
Password: password
```

:::caution[Before adding household data]
Replace the example [Dex users and clients](../dex-users/) and complete the
[household-use checklist](../self-host-operations/#before-household-use).
:::

Next, create your [First Inventory](../first-inventory/).

## Stop

```sh
docker compose -f compose.selfhost.yaml down
```

Your data stays in Docker volumes. For LAN access, backups, upgrades, or full
removal, see [Self-Host Operations](../self-host-operations/).
