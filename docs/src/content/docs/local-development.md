---
title: Development Setup
description: Commands for working on Stuff Stash locally.
---

Use this page when you are changing Stuff Stash itself. If you only want to
evaluate the app, start with [Run Stuff Stash](../self-hosting/).

## Requirements

- Go `1.25.8` or newer.
- Docker with Compose.
- Lefthook.
- pnpm `11.0.7`.

The repository pins dependencies, container images, GitHub Actions, and tooling
versions intentionally. Supply-chain security is part of the project boundary.

Runtime settings are listed in the
[Configuration Reference](../configuration/).

## API

Run the API with in-memory local adapters:

```sh
make run
```

Check health:

```sh
curl http://localhost:8080/healthz
```

Expected response:

```json
{"service":"stuff-stash","status":"healthy"}
```

Run API tests:

```sh
make test
```

## Full Local Stack

Run Postgres, SpiceDB, Dex, migrations, and the API with Docker:

```sh
make compose-up-oidc
```

With the Docker stack, verify the OIDC and authorization flow:

```sh
make verify-dex-oidc-api
```

Stop the Docker stack:

```sh
make compose-down
```

## Host Dex Mobile OIDC

If you use the local Dex binary, start Dex and the API as host processes:

```sh
make dex-local
make run-oidc-local
```

Verify the native mobile authorization-code flow, PKCE exchange, refresh
exchange, API mobile metadata, and `/me` with the refreshed mobile ID token:

```sh
make verify-mobile-oidc-pkce-local
```

For physical-device mobile OIDC testing, run Dex and the API with a LAN-reachable
issuer. Replace `192.168.1.50` with this machine's LAN IP:

```sh
make dex-local STUFF_STASH_LOCAL_HOST=192.168.1.50
make run-oidc-local STUFF_STASH_LOCAL_HOST=192.168.1.50
STUFF_STASH_LOCAL_HOST=192.168.1.50 make verify-mobile-oidc-pkce-local
```

The Docker workflow can also use a LAN issuer:

```sh
make compose-up-oidc-lan STUFF_STASH_LAN_HOST=192.168.1.50
```

## Web App

Start the web app:

```sh
make web-install
make web-dev
```

Open `http://localhost:5173`.

Run web checks:

```sh
make web-test
make web-check
make web-build
```

## API Client

The browser app uses the generated TypeScript API client at its adapter
boundary. When the OpenAPI contract changes, regenerate and check it:

```sh
make api-client-generate
make api-client-test
make api-client-check
make api-client-check-generated
```

## Docs

Run the docs site:

```sh
make docs-install
make docs-dev
```

Build the docs:

```sh
make docs-build
```

## Pre-Commit

Run the configured hooks:

```sh
lefthook run pre-commit --all-files
```

The hook runs Go formatting, Go tests, and structural checks for common project
drift such as raw SQL in Go code, ad hoc prints, oversized Go files, HTTP adapter
organization drift, and GORM adapter catch-all files.
