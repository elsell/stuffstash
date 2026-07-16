---
title: Self-Host Stuff Stash
description: Run Stuff Stash on your home network with the supported release bundle.
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

## 2. Configure

```sh
./scripts/configure-selfhost.sh
./scripts/selfhost-preflight.sh
```

The setup command detects the server's LAN IPv4 address and prints the URL to
open. If it chooses the wrong address, remove `.env` and `.stuffstash`, then run
it again with `--host 192.168.1.40`.

:::caution[Your LAN can reach the example accounts]
Ports `8081`, `8080`, `5556`, and `3900` listen on the local network. Anyone
who can reach the server can use the example credentials below. Do not forward
these ports to the internet.
:::

:::note[Host firewall enabled?]
Allow inbound TCP ports `8081`, `8080`, `5556`, and `3900` on the server.
:::

## 3. Start

```sh
docker compose -f compose.selfhost.yaml up -d
```

[Trust Caddy's local certificate](../self-host-operations/#trust-the-local-certificate)
on each device that will open Stuff Stash. Then open the URL printed by the
setup command and sign in:

```text
Email: owner@example.com
Password: password
```

:::note[Want private accounts?]
[Replace the example credentials](../self-host-operations/#replace-the-example-credentials)
before adding data. The simple reset deletes existing volumes.
:::

Next, create your [First Inventory](../first-inventory/).

## Stop

```sh
docker compose -f compose.selfhost.yaml down
```

Your data stays in Docker volumes. See [Self-Host
Operations](../self-host-operations/) to replace the example users, keep the
address stable, back up, upgrade, or remove the installation.
