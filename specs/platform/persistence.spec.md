# Persistence Spec

## Purpose

Stuff Stash needs persistence that supports flexible inventory data while keeping domain logic independent from storage details.

## Scope

This spec covers initial persistence direction for PostgreSQL, SQLite, GORM, migrations, and custom field storage.

This spec does not define every future asset index, backup strategy, retention policy, or production deployment topology.

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
- State-changing application operations must write append-only audit records behind a repository port.
- Audit records must preserve tenant and inventory isolation in storage and read queries.
- Audit metadata must be stored as a JSON object and must not be used as an authorization source.

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
- `inventory_access_grants`
  - tenant ID
  - inventory ID
  - grant key
  - principal ID
  - relationship
  - timestamps managed by GORM
- `custom_field_definitions`
  - `id`
  - `tenant_id`
  - `inventory_id` when scoped to one inventory
  - scope
  - cursor key
  - field key
  - display name
  - field type
  - enum options as JSONB
  - applicability target
  - timestamps managed by GORM
- `custom_asset_types`
  - `id`
  - `tenant_id`
  - `inventory_id` when scoped to one inventory
  - scope
  - cursor key
  - type key
  - display name
  - description
  - timestamps managed by GORM
- `custom_field_definition_asset_types`
  - custom field definition ID
  - custom asset type ID
  - tenant ID
  - inventory ID when the field definition is inventory-scoped
  - timestamps managed by GORM
- `audit_records`
  - `id`
  - tenant ID
  - inventory ID when inventory-scoped
  - principal ID
  - action
  - source
  - target type
  - target ID
  - occurred timestamp
  - request ID
  - metadata as JSONB
  - timestamps managed by GORM

The `inventories.tenant_id` value must reference `tenants.id`.
Authorization outbox records represent pending relationship grants that must be applied to SpiceDB.
Inventory access grant rows represent direct inventory viewer/editor grants. The primary key must be scoped by tenant ID, inventory ID, and grant key so the same principal can receive the same relationship in more than one inventory.

Inventory access grant writes must be committed with a matching authorization outbox event in one transaction. Viewer grants must enqueue `grant_inventory_viewer`; editor grants must enqueue `grant_inventory_editor`.
Repeated writes for the same tenant, inventory, principal, and relationship must be no-ops and must not enqueue another authorization outbox event.

Custom field definitions must be scoped by tenant and optionally by inventory. The first persistence shape must prevent duplicate tenant-scoped keys inside one tenant and duplicate inventory-scoped keys inside one inventory.
PostgreSQL migrations must also enforce the effective-key invariant across scopes so concurrent API replicas cannot create an inventory-scoped definition and a tenant-scoped definition with the same key in one tenant.
That enforcement must include a database-level serialization point, such as a transaction-scoped advisory lock keyed by tenant and field key, not only a read-before-write trigger check.
The database must validate persisted custom field definition key, display name, enum option, scope, and field type shape as a defense-in-depth guard around the domain model.

Custom asset type definitions must be scoped by tenant and optionally by inventory. They must preserve the same effective-key invariant as custom field definitions: no duplicate tenant-scoped type key inside one tenant, no duplicate inventory-scoped type key inside one inventory, and no tenant/inventory effective-key collision.

Custom field definitions that target custom asset types must use `custom_field_definition_asset_types` join rows. A field with applicability `all_assets` must have no join rows. A field with applicability `custom_asset_types` must have at least one join row. Join rows must reference custom asset types available to the field definition's effective scope.

Audit record persistence must be append-only through application ports in the first slice. All repository adapters, including the in-memory adapter, must reject duplicate audit record IDs instead of overwriting existing records. State changes paired with audit writes must fail without partial domain writes when the audit write fails. The database must validate audit action, source, target type, and metadata object shape. Audit list queries must support tenant-wide and inventory-scoped cursor pagination ordered by `(occurred_at, id)`. Inventory references must not cascade-delete audit rows; until delete behavior is specified, inventory deletion must be blocked if audit rows still reference the inventory. Migration `000009_preserve_audit_records_and_ordering` must convert existing audit record inventory references from cascade delete to restrict delete and recreate audit list indexes for `(occurred_at, id)` ordering.

## Initial Asset Schema

The first asset persistence slice must use a unified `assets` table for normal items, movable containers, and place-like locations.

The `assets` table must include:

- `id`
- `tenant_id`
- `inventory_id`
- `parent_asset_id`
- asset kind
- custom asset type ID once custom asset types are implemented
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
- Custom asset type must not affect containment invariants; base asset kind remains authoritative for parent/child behavior.
- Custom asset type references must be tenant/inventory valid when present.
- Custom field values must be stored as JSONB in PostgreSQL and must default to an empty object.
- The repository adapter must be able to round-trip validated custom field values so future custom-field APIs do not require a persistence rewrite.
- The public asset create endpoint may accept non-empty custom field objects only after application validation against effective custom field definitions that apply to the asset's custom asset type.

The initial asset kind enumeration is:

- `item`
- `container`
- `location`

Custom asset types such as `medicine`, `document`, or `battery` must not be added to the base asset kind enumeration. They belong in custom asset type persistence.

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
- Custom asset type definitions must be stored separately from assets and custom field definitions.
- Custom field-to-custom-asset-type targets must be stored in a join table so one field can apply to multiple custom asset types.
- Tenant-scoped custom field definitions must be distinguishable from inventory-scoped custom field definitions.
- Tenant-scoped custom asset type definitions must be distinguishable from inventory-scoped custom asset type definitions.
- In the PostgreSQL adapter, asset custom field values should be stored in a JSONB column.
- The initial `assets.custom_fields` value must be a JSON object.
- Asset custom field values may be non-empty only after application validation against effective custom field definitions that apply to the asset's custom asset type.
- Domain code must not expose or manipulate raw JSONB.
- Repository adapters must map stored custom field values into typed domain values.
- Custom field values must be validated by domain or application services before persistence.
- Repository adapters must update assets without changing their tenant, inventory, kind, or lifecycle state in the first update slice.
- Repository adapters must defensively reject asset updates that would point to a parent in another tenant or inventory, point to an item parent, point to an archived parent, point to the asset itself, or create a containment cycle.

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
- Tests must cover custom field definition persistence, duplicate key rejection, tenant-scoped listing, inventory-scoped listing, and effective definition resolution.
- Tests must cover that location-like nodes are persisted as assets with kind `location`.
- Tests must cover parent-child persistence and movement within a single inventory.
- Tests must cover moving an asset to the inventory root and moving container/location assets with descendants.
- Tests must cover that cross-tenant and cross-inventory containment is rejected.
- Tests must cover cycle prevention before persistence commits.
- Tests must cover that tenant and inventory state are persisted with authorization outbox records atomically.
- Tests must cover that inventory access grants and their authorization outbox records are persisted atomically.
- Tests must cover that duplicate inventory access grants do not enqueue duplicate authorization outbox events.
- Tests must cover that inventory access grant keys are scoped to the inventory, not globally.
- Tests must cover that duplicate audit record IDs are rejected and that paired domain writes roll back when an audit insert fails.
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
