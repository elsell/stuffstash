# Migrations Spec

## Purpose

Stuff Stash needs a pinned, reproducible migration workflow before durable persistence is implemented.

## Scope

This spec chooses the initial migration tool and workflow.

This spec does not define deployment automation.

## Decision

- The first migration tool must be `golang-migrate/migrate`.
- The migration CLI must be pinned to a reviewed version before use.
- Migration files must live under `apps/api/migrations`.
- Migration file names must be ordered and descriptive.
- Migration commands must run from root `make` targets.
- The first migration creates `tenants` and `inventories` for the secure tracer bullet.

## Requirements

- Migrations must be reviewed as code.
- Migrations must be deterministic.
- Migrations must not contain application business logic.
- Migrations must preserve supply-chain pinning rules.
- Migration commands must support PostgreSQL.
- SQLite migration support may be added for local development and test fakes when needed.
- Direct SQL in migration files is allowed because migrations define schema, but direct SQL in application code remains forbidden.

## Commands

The root Makefile should eventually expose:

- `make migrate-up`
- `make migrate-down`
- `make migrate-status`

These commands must require explicit database configuration through environment variables.

Until the migration CLI is pinned in the repository, local Compose may use GORM's schema migration for the first tracer bullet only. Production deployments must use reviewed migration files.

## Testing

- Migration verification must run against PostgreSQL before persistence features are merged.
- Migration tooling must be included in CI once database schema exists.
