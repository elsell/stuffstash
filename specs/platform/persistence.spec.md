# Persistence Spec

## Purpose

Stuff Stash needs persistence that supports flexible inventory data while keeping domain logic independent from storage details.

## Scope

This spec covers initial persistence direction for PostgreSQL, SQLite, GORM, migrations, and custom field storage.

This spec does not define the full database schema, migration tool choice, indexing strategy, backup strategy, or production deployment topology.

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

## Custom Fields

- Custom field definitions must be stored separately from asset custom field values.
- Tenant-scoped custom field definitions must be distinguishable from inventory-scoped custom field definitions.
- In the PostgreSQL adapter, asset custom field values should be stored in a JSONB column.
- Domain code must not expose or manipulate raw JSONB.
- Repository adapters must map stored custom field values into typed domain values.
- Custom field values must be validated by domain or application services before persistence.

## Migrations

- A Go-friendly migration tool must be selected before the first durable database schema is implemented.
- The migration tool must be pinned to a known reviewed version.
- Migration commands must be reproducible locally and in CI.
- Migrations must be reviewed as code.
- Migrations must preserve supply-chain rules for pinned tooling and dependencies.
- The migration tool choice remains open until a spec compares the practical trade-offs.

## Testing

- Repository tests must use real behavior, not mocks.
- Repository tests may use SQLite fakes only when the behavior is meaningfully equivalent to the production repository contract.
- PostgreSQL-backed tests are required for PostgreSQL-specific behavior such as JSONB semantics, constraints, indexing, or transaction behavior.
- Tests must cover tenant isolation, inventory isolation, custom field mapping, custom field validation boundaries, and error handling.

## Open Questions

- Which Go migration tool should be used first?
- Which tests must run against PostgreSQL instead of SQLite?
- What indexing strategy is needed for JSONB custom fields?
- How should repository contracts be split across asset, inventory, location, and identity contexts?
