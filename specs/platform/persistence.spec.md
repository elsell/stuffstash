# Persistence Spec

## Purpose

Stuff Stash needs persistence that supports flexible inventory data while keeping domain logic independent from storage details.

## Scope

This spec covers initial persistence direction for PostgreSQL, SQLite, GORM, migrations, and custom field storage.

This spec does not define the full asset schema, indexing strategy, backup strategy, or production deployment topology.

## Requirements

- PostgreSQL is the production database.
- SQLite may be used for local development.
- SQLite may be used for test fakes where that provides useful confidence without external services.
- GORM must be used as the ORM.
- Application code must not use direct SQL.
- Direct SQL is a code smell and requires spec-level justification if ever considered.
- Persistence must live behind repositories and adapters.
- Domain logic must not depend on GORM models, database column names, JSONB structures, migration tools, or SQL dialect details.
- Repository interfaces must be shaped around domain and application needs, not around database tables.
- Tenant isolation and inventory isolation must be represented in persistence.
- Persistence behavior must preserve authorization and tenancy assumptions from the application layer.
- Repository adapter choice must come from `STUFF_STASH_REPOSITORY_MODE`.
- `memory` repository mode is allowed for local tracer bullets and tests.
- `postgres` repository mode must use GORM and `STUFF_STASH_DATABASE_DSN`.
- Local Compose must run the API with `STUFF_STASH_REPOSITORY_MODE=postgres`.
- State changes that require SpiceDB relationship writes must use a transactional outbox.
- Domain state and authorization outbox records must be written in the same database transaction.
- Application code must not save durable tenant or inventory state and then rely only on an inline SpiceDB call for consistency.
- Outbox processing must be retryable and idempotent.
- Outbox processing may run synchronously after writes for low-latency local behavior, but the durable outbox is the consistency mechanism.
- Outbox processing must also have a retry path that does not depend on a future write request.
- Outbox batch size and retry interval must come from environment-backed configuration.
- A failed outbox event must not prevent later events in the same batch from being attempted.
- Outbox observability must include event ID, event kind, tenant ID, inventory ID when present, attempt count, and failure details when an event fails.

## Initial Schema

The first durable schema covers the secure tenant/inventory tracer bullet:

- `tenants`
  - `id`
  - `name`
  - timestamps managed by GORM
- `inventories`
  - `id`
  - `tenant_id`
  - `name`
  - timestamps managed by GORM
- `authorization_outbox_events`
  - `id`
  - event kind
  - principal ID
  - tenant ID
  - inventory ID when applicable
  - attempts and last error
  - processed timestamp
  - timestamps managed by GORM

The `inventories.tenant_id` value must reference `tenants.id`.
Authorization outbox records represent pending relationship grants that must be applied to SpiceDB.

## Custom Fields

- Custom field definitions must be stored separately from asset custom field values.
- Tenant-scoped custom field definitions must be distinguishable from inventory-scoped custom field definitions.
- In the PostgreSQL adapter, asset custom field values should be stored in a JSONB column.
- Domain code must not expose or manipulate raw JSONB.
- Repository adapters must map stored custom field values into typed domain values.
- Custom field values must be validated by domain or application services before persistence.

## Migrations

- The initial migration tool is `golang-migrate/migrate`.
- The migration tool must be pinned to a known reviewed version.
- Migration commands must be reproducible locally and in CI.
- Migrations must be reviewed as code.
- Migrations must preserve supply-chain rules for pinned tooling and dependencies.
- Migration files must live under `apps/api/migrations`.

## Testing

- Repository tests must use real behavior, not mocks.
- Repository tests may use SQLite fakes only when the behavior is meaningfully equivalent to the production repository contract.
- PostgreSQL-backed tests are required for PostgreSQL-specific behavior such as JSONB semantics, constraints, indexing, or transaction behavior.
- Tests must cover tenant isolation, inventory isolation, custom field mapping, custom field validation boundaries, and error handling.
- Tests must cover that tenant and inventory state are persisted with authorization outbox records atomically.
- Tests must cover that authorization outbox processing retries failed relationship grants and marks successful grants processed.
- Tests must cover rollback when a domain state write succeeds but the paired authorization outbox insert fails.
- Security tests must cover that state created while authorization grants fail remains protected until the outbox grant succeeds.

## Open Questions

- What indexing strategy is needed for JSONB custom fields?
- How should repository contracts be split across asset, inventory, location, and identity contexts?
