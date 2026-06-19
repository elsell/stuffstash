# Persistence Spec

## Purpose

Stuff Stash needs persistence that supports flexible inventory data while keeping domain logic independent from storage details.

## Scope

This spec covers initial persistence direction for PostgreSQL, SQLite, GORM, migrations, and custom field storage.

This spec does not define every future asset index, backup strategy, or production deployment topology.

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
- Outbox claim lease duration must come from environment-backed configuration.
- Outbox workers must claim events before processing them so multiple API replicas do not process and update the same event at the same time.
- Claimed outbox events must use a lease deadline so a crashed worker does not strand events forever.
- Processing completion or failure must only update an event when the caller still owns the event claim.
- Unrecoverable outbox events must be moved to a dead-letter terminal state instead of being retried forever.
- Dead-lettered events must not be claimed for normal processing.
- Dead-lettering must require ownership of the active event claim.
- A failed outbox event must not prevent later events in the same batch from being attempted.
- Outbox observability must include event ID, event kind, tenant ID, inventory ID when present, attempt count, and failure details when an event fails.
- Dead-letter observability must include event ID, event kind, tenant ID, inventory ID when present, and the terminal reason.

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
  - claim ID
  - claimed until timestamp
  - processed timestamp
  - dead-letter timestamp
  - dead-letter reason
  - timestamps managed by GORM

The `inventories.tenant_id` value must reference `tenants.id`.
Authorization outbox records represent pending relationship grants that must be applied to SpiceDB.

## Initial Asset Schema

The first asset persistence slice must use a unified `assets` table for normal items, movable containers, and place-like locations.

The `assets` table must include:

- `id`
- `tenant_id`
- `inventory_id`
- `parent_asset_id`
- asset kind
- title
- description
- custom field values
- lifecycle state
- timestamps managed by GORM

The table must preserve these invariants:

- `tenant_id` references `tenants.id`.
- `inventory_id` references `inventories.id`.
- `parent_asset_id` references `assets.id` when present.
- A child asset and parent asset must be in the same tenant and inventory.
- Same-tenant and same-inventory parentage must be enforced by application/domain validation and repository adapter defensive checks before commit.
- A plain `parent_asset_id` foreign key is not sufficient by itself.
- An asset must not be its own parent.
- Containment cycles must not be persisted.
- Only container-capable asset kinds may be used as parents.
- Custom field values must be stored as JSONB in PostgreSQL and must default to an empty object.
- The first asset slice may only persist empty custom field objects until custom field definitions are implemented.

The initial asset kind enumeration is:

- `item`
- `container`
- `location`

The initial lifecycle state enumeration is:

- `active`
- `archived`

The first implementation should include indexes that support tenant/inventory scoping and parent-child traversal:

- `(tenant_id, inventory_id)`
- `parent_asset_id`
- `(inventory_id, parent_asset_id)`
- `(inventory_id, kind)`

Additional custom-field JSONB indexes must wait until query behavior is specified.

## Custom Fields

- Custom field definitions must be stored separately from asset custom field values.
- Tenant-scoped custom field definitions must be distinguishable from inventory-scoped custom field definitions.
- In the PostgreSQL adapter, asset custom field values should be stored in a JSONB column.
- The initial `assets.custom_fields` value must be a JSON object.
- The first asset slice must reject non-empty custom field values until custom field definition persistence and validation are implemented.
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
- Tests must cover that location-like nodes are persisted as assets with kind `location`.
- Tests must cover parent-child persistence and movement within a single inventory.
- Tests must cover that cross-tenant and cross-inventory containment is rejected.
- Tests must cover cycle prevention before persistence commits.
- Tests must cover that tenant and inventory state are persisted with authorization outbox records atomically.
- Tests must cover that authorization outbox processing retries failed relationship grants and marks successful grants processed.
- Tests must cover rollback when a domain state write succeeds but the paired authorization outbox insert fails.
- Tests must cover that claimed authorization outbox events are hidden from other claims until their lease expires.
- Tests must cover that processed and failed updates require ownership of the active claim.
- Tests must cover that dead-lettered authorization outbox events are not claimed again.
- Tests must cover that unrecoverable authorization outbox event data is dead-lettered while transient authorization failures remain retryable.
- Security tests must cover that state created while authorization grants fail remains protected until the outbox grant succeeds.

## Open Questions

- What indexing strategy is needed for JSONB custom fields?
- Can assets move between inventories, or must cross-inventory movement be modeled as export/import?
