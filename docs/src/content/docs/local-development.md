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

Run Postgres, SpiceDB, Dex, migrations, and the API:

```sh
make compose-up-oidc
```

Verify the OIDC and authorization flow:

```sh
make verify-dex-oidc-api
```

Stop the stack:

```sh
make compose-down
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
