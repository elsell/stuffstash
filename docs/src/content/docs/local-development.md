---
title: Local Development
description: How to run Stuff Stash locally.
---

This page shows the current local workflow.

The app is still early. Today, you can run the Go API, create tenants, create inventories, define custom fields, create, list, update, and move assets, view audit history, share inventory access, persist data in Postgres through Compose, and test the first auth boundary.

## Requirements

- Go 1.25.8 or newer.
- Docker with Compose, for container-based local runs.
- Lefthook, for pre-commit checks.
- pnpm 11.0.7, for this documentation site.

Dependencies and base images are pinned on purpose. This project treats supply-chain security as part of the security boundary.

## Run Tests

From the repository root:

```sh
make test
```

## Run The API Fast

From the repository root:

```sh
make run
```

This uses local development auth, in-memory authorization, and in-memory persistence.

Visit the base URL to see a small API index:

```sh
curl http://localhost:8080/
```

Then check the health endpoint:

```sh
curl http://localhost:8080/healthz
```

Expected response:

```json
{"service":"stuff-stash","status":"healthy"}
```

## Try The Secure API

Local development auth uses deterministic bearer tokens:

```sh
Authorization: Bearer dev:user-one
```

Create a tenant:

```sh
curl -s http://localhost:8080/tenants \
  -H 'Authorization: Bearer dev:user-one' \
  -H 'Content-Type: application/json' \
  -d '{"name":"Home"}'
```

Create an inventory inside that tenant, replacing the tenant ID with the response from the previous command:

```sh
curl -s http://localhost:8080/tenants/<tenant-id>/inventories \
  -H 'Authorization: Bearer dev:user-one' \
  -H 'Content-Type: application/json' \
  -d '{"name":"Tools"}'
```

List visible inventories:

```sh
curl -s 'http://localhost:8080/tenants/<tenant-id>/inventories?limit=50' \
  -H 'Authorization: Bearer dev:user-one'
```

Create a location-like asset:

```sh
curl -s http://localhost:8080/tenants/<tenant-id>/inventories/<inventory-id>/assets \
  -H 'Authorization: Bearer dev:user-one' \
  -H 'Content-Type: application/json' \
  -d '{"kind":"location","title":"Garage"}'
```

Create an item inside that location:

```sh
curl -s http://localhost:8080/tenants/<tenant-id>/inventories/<inventory-id>/assets \
  -H 'Authorization: Bearer dev:user-one' \
  -H 'Content-Type: application/json' \
  -d '{"kind":"item","title":"Fertilizer","parentAssetId":"<garage-asset-id>"}'
```

Create another location under the garage if you want to test movement:

```sh
curl -s http://localhost:8080/tenants/<tenant-id>/inventories/<inventory-id>/assets \
  -H 'Authorization: Bearer dev:user-one' \
  -H 'Content-Type: application/json' \
  -d '{"kind":"location","title":"Shelf","parentAssetId":"<garage-asset-id>"}'
```

Define custom fields before sending custom values on assets:

Keys use lowercase letters, numbers, and hyphens. Initial field types are `text`, `number`, `boolean`, `date`, `url`, and `enum`.

```sh
curl -s http://localhost:8080/tenants/<tenant-id>/custom-field-definitions \
  -H 'Authorization: Bearer dev:user-one' \
  -H 'Content-Type: application/json' \
  -d '{"key":"serial","displayName":"Serial","type":"text"}'

curl -s http://localhost:8080/tenants/<tenant-id>/inventories/<inventory-id>/custom-field-definitions \
  -H 'Authorization: Bearer dev:user-one' \
  -H 'Content-Type: application/json' \
  -d '{"key":"condition","displayName":"Condition","type":"enum","enumOptions":["new","used"]}'
```

Create an item with custom field values:

```sh
curl -s http://localhost:8080/tenants/<tenant-id>/inventories/<inventory-id>/assets \
  -H 'Authorization: Bearer dev:user-one' \
  -H 'Content-Type: application/json' \
  -d '{"kind":"item","title":"Fertilizer","parentAssetId":"<garage-asset-id>","customFields":{"serial":"bag-1","condition":"new"}}'
```

Move or edit an asset:

```sh
curl -s http://localhost:8080/tenants/<tenant-id>/inventories/<inventory-id>/assets/<asset-id> \
  -X PATCH \
  -H 'Authorization: Bearer dev:user-one' \
  -H 'Content-Type: application/json' \
  -d '{"title":"Fertilizer Bag","parentAssetId":"<shelf-asset-id>","customFields":{"serial":"bag-1","condition":"used"}}'
```

Send `parentAssetId` as `null` to move an asset back to the inventory root.

Inventory custom field lists include tenant fields and inventory fields:

```sh
curl -s 'http://localhost:8080/tenants/<tenant-id>/inventories/<inventory-id>/custom-field-definitions?limit=50' \
  -H 'Authorization: Bearer dev:user-one'
```

List assets in the inventory:

```sh
curl -s 'http://localhost:8080/tenants/<tenant-id>/inventories/<inventory-id>/assets?limit=50' \
  -H 'Authorization: Bearer dev:user-one'
```

List inventory audit history:

```sh
curl -s 'http://localhost:8080/tenants/<tenant-id>/inventories/<inventory-id>/audit-records?limit=50' \
  -H 'Authorization: Bearer dev:user-one'
```

Grant another local dev user viewer access:

Viewers can list assets and inventory audit history. Editors can list, create, update, and move assets. Neither can share the inventory with someone else.

```sh
curl -s http://localhost:8080/tenants/<tenant-id>/inventories/<inventory-id>/access-grants \
  -H 'Authorization: Bearer dev:user-one' \
  -H 'Content-Type: application/json' \
  -d '{"principalId":"user-two","relationship":"viewer"}'
```

List direct inventory access grants:

```sh
curl -s 'http://localhost:8080/tenants/<tenant-id>/inventories/<inventory-id>/access-grants?limit=50' \
  -H 'Authorization: Bearer dev:user-one'
```

OpenAPI JSON is generated by Huma:

```sh
curl http://localhost:8080/openapi.json
```

The interactive local API docs are also available at:

```sh
open http://localhost:8080/docs
```

You can run the same tenant, inventory, asset, and sharing flow as a check:

```sh
make verify-local-api
```

## Run With Compose

```sh
make compose-up
```

Compose starts Postgres, runs migrations with the same API image, then starts the API and SpiceDB. By default, the API uses Postgres persistence and in-memory authorization.

To run the API against SpiceDB authorization:

```sh
make compose-up-spicedb
```

This starts the same local stack, keeps Postgres persistence, switches authorization to SpiceDB, and bootstraps the checked-in schema.
Local SpiceDB uses `serve-testing`, so it does not need a preshared key.

In another terminal, run:

```sh
make verify-local-api
```

If port `8080` is already in use, choose another host port:

```sh
STUFF_STASH_HTTP_PORT=18080 make compose-up-spicedb
STUFF_STASH_VERIFY_BASE_URL=http://localhost:18080 make verify-local-api
```

Stop it with:

```sh
make compose-down
```

## Run Migrations

The app binary owns migrations too. That keeps local runs and Kubernetes jobs on the same path.

```sh
make migrate-up
make migrate-status
```

Set `STUFF_STASH_DATABASE_DSN` when you need a database other than the local Compose default.

To verify migration behavior against Postgres:

```sh
make verify-migrations
```

## Run With SpiceDB Without Compose

If your Docker install does not have Compose, use:

```sh
make run-spicedb
```

That starts a local SpiceDB container with `docker run`, runs the API against it, and bootstraps the schema.

In another terminal:

```sh
make verify-local-api
```

Stop the SpiceDB container when you are done:

```sh
make spicedb-down
```

## Verify The SpiceDB Adapter

Run the real SpiceDB adapter checks with Docker:

```sh
make verify-spicedb-adapter
```

This starts the pinned local SpiceDB image, runs the adapter integration test, and removes the test container. If Go is not installed locally, the script runs the test inside the pinned Go builder image.

## Run The Docs

The docs site lives in `docs/`.

```sh
make docs-install
make docs-dev
```

## Auth Modes

Local runs use safe defaults:

- `STUFF_STASH_AUTH_MODE=local-dev`
- `STUFF_STASH_AUTHZ_MODE=memory`
- `STUFF_STASH_REPOSITORY_MODE=memory`

You can switch to the production-shaped adapters with:

- `STUFF_STASH_AUTH_MODE=oidc`
- `STUFF_STASH_AUTHZ_MODE=spicedb`
- `STUFF_STASH_REPOSITORY_MODE=postgres`

OIDC needs an issuer and client ID. A secured SpiceDB deployment needs an endpoint and preshared key. Postgres needs `STUFF_STASH_DATABASE_DSN`. Local Compose already starts Postgres and unauthenticated SpiceDB `serve-testing`, but the API does not use SpiceDB unless you choose the `spicedb` mode.

## Pre-Commit Checks

Run all configured hooks:

```sh
lefthook run pre-commit --all-files
```

The hook runs Go formatting, Go tests, and structural checks for ad hoc prints and raw SQL in Go code.
