# Migrations Spec

## Purpose

Stuff Stash needs a pinned, reproducible migration workflow before durable persistence is implemented.

## Scope

This spec chooses the initial migration tool and workflow.

This spec does not define full deployment automation.

## Decision

- The first migration tool must be `golang-migrate/migrate`.
- Migration execution must use the pinned Go library `github.com/golang-migrate/migrate/v4`.
- Migration files must live under `apps/api/migrations`.
- Migration file names must be ordered and descriptive.
- Migration commands must run from root `make` targets.
- Migration commands must run through the same `stuff-stash` binary and container image as the API.
- Kubernetes deployments should run migrations as a Job or init workflow using the same image with a migration command.
- The API server command must not require building a separate migration image.
- The API server must verify that PostgreSQL schema migrations are clean and current before accepting traffic.
- Destructive rollback commands must not be exposed until a guarded rollback workflow is specified.

## Requirements

- Migrations must be reviewed as code.
- Migrations must be deterministic.
- Migrations must not contain application business logic.
- Migrations must preserve supply-chain pinning rules.
- Migration commands must support PostgreSQL.
- SQLite migration support may be added for local development and test fakes when needed.
- Direct SQL in migration files is allowed because migrations define schema, but direct SQL in application code remains forbidden.
- Migration files must be embedded into the Go binary or otherwise packaged in the same image so Kubernetes jobs can run without extra mounts.
- Migration commands must use `STUFF_STASH_DATABASE_DSN`.
- The API server must not rely on GORM `AutoMigrate` for production schema creation.
- Startup schema verification must not apply migrations; it must only fail fast when the schema is missing, dirty, or behind the embedded migration set.

## Commands

The root Makefile should eventually expose:

- `make migrate-up`
- `make migrate-status`

These commands must use the same binary migration commands that Kubernetes would run.

Local Compose must run migrations before starting the API against Postgres. Kubernetes deployments must do the same with an explicit migration Job or init workflow.

## Testing

- Migration verification must run against PostgreSQL before persistence features are merged.
- Migration tooling must be included in CI once database schema exists.
- Tests must verify migration command behavior with missing configuration and no-op migration states.
- Tests must verify API startup migration checks reject missing, dirty, or outdated schema state before the API serves requests.
