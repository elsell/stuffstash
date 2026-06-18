---
title: Local Development
description: How to run Stuff Stash locally.
---

This page shows the current local workflow.

The app is still early. Today, the runnable part is the Go API health endpoint.

## Requirements

- Go 1.25.x.
- Docker with Compose, for container-based local runs.
- Lefthook, for pre-commit checks.
- pnpm 11.0.7, for this documentation site.

Dependencies and base images are pinned on purpose. This project treats supply-chain security as part of the security boundary.

## Run Tests

From the repository root:

```sh
make test
```

## Run The API

From the repository root:

```sh
make run
```

Then check the health endpoint:

```sh
curl http://localhost:8080/healthz
```

Expected response:

```json
{"service":"stuff-stash","status":"healthy"}
```

## Run With Compose

```sh
make compose-up
```

Stop it with:

```sh
make compose-down
```

## Run The Docs

The docs site lives in `docs/`.

```sh
cd docs
pnpm install --frozen-lockfile
pnpm dev
```

The lockfile is committed, so use frozen installs:

```sh
pnpm install --frozen-lockfile
```

## Pre-Commit Checks

Run all configured hooks:

```sh
lefthook run pre-commit --all-files
```

The hook runs Go formatting, Go tests, and structural checks for ad hoc prints and raw SQL in Go code.
