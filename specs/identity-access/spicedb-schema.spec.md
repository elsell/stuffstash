# SpiceDB Schema Spec

## Purpose

Stuff Stash needs an exact first authorization schema before protected tenant and inventory endpoints are implemented.

## Scope

This spec defines the first relationship model for tenants and inventories.

It does not define assets, locations, attachments, invitations, or audit permissions in full.

## Schema

The first SpiceDB schema must model:

```zed
definition user {}

definition tenant {
  relation owner: user
  relation admin: user
  relation viewer: user

  permission view = owner + admin + viewer
  permission view_inventory = owner + admin
  permission configure = owner + admin
  permission create_inventory = owner + admin
}

definition inventory {
  relation tenant: tenant
  relation owner: user
  relation editor: user
  relation viewer: user

  permission view = owner + editor + viewer + tenant->view_inventory
  permission configure = owner + tenant->configure
  permission share = owner + tenant->configure
  permission edit = owner + editor + tenant->configure
  permission edit_asset = edit
  permission create_asset = edit
}
```

The schema must also exist as a checked-in `.zed` file so local and production bootstrapping do not depend on copying schema text from Markdown.

## Tracer Bullet Permissions

The first API slice must check:

- `tenant.create_inventory` before creating an inventory inside a tenant.
- `inventory.view` before returning an inventory.
- `inventory.create_asset` before creating an asset inside an inventory.
- `inventory.edit_asset` before updating or moving an asset inside an inventory.
- `inventory.share` before creating or listing direct inventory access grants.
- `inventory.view` before listing inventory-scoped audit records.
- `tenant.configure` before listing tenant-wide audit records.

Creating a tenant grants the creating user the tenant `owner` relationship.

Creating an inventory grants the creating user the inventory `owner` relationship and links the inventory to its tenant.

Granting inventory ownership must also grant tenant `viewer` on the containing tenant. This lets an inventory owner resolve the tenant and list only visible inventories without granting tenant `create_inventory`, tenant `configure`, or inherited visibility to sibling inventories.

Granting direct inventory `viewer` or `editor` must also grant tenant `viewer` on the containing tenant for the same reason.

Tenant `view_inventory` must stay separate from tenant `view`. Tenant owners and admins inherit inventory view across the tenant through `tenant->view_inventory`; tenant viewers do not.

## Adapter Strategy

- The authorization port must support permission checks.
- The authorization port must support relationship writes needed by the application service.
- The first implementation may use an in-memory fake for tests and the initial tracer bullet.
- The production adapter must use SpiceDB.
- Application services must not import SpiceDB client types.
- The SpiceDB adapter must be selected through environment configuration.
- The SpiceDB adapter must fail closed when checks return anything other than `HAS_PERMISSION`.
- Relationship writes must be idempotent enough for application retries. Already-existing relationships should not create privilege gaps or partial grants.
- Relationship writes that correspond to durable tenant or inventory state must be driven by the authorization outbox, not by a standalone inline call after persistence.
- Direct inventory access grant relationship writes must be driven by the authorization outbox after the durable access grant record is saved.
- The application may attempt to drain the authorization outbox immediately after a write, but failed SpiceDB writes must remain retryable through the outbox.
- Schema bootstrap must be an explicit startup option for local development and deployment automation.
- The default API process must not rewrite the production authorization schema unless explicitly configured to do so.

## Authorization Modes

The API supports these authorization modes:

- `memory`: in-memory relationships for local development and tests.
- `spicedb`: SpiceDB relationship checks and relationship writes.

Any unknown mode must fail startup.

## Verification

- Tests must prove tenant owner inheritance.
- Tests must prove cross-tenant denial.
- Tests must prove unauthenticated requests never reach privileged behavior.
- Tests must prove a user without a relationship cannot create or list inventory resources.
- The SpiceDB adapter must have fake-backed tests for permission mapping, tenant owner grants, inventory owner grants, inventory viewer/editor grants, denied checks, and backend failures.
- The SpiceDB adapter must have an opt-in real-SpiceDB verification path that runs against the pinned local SpiceDB image.
- Real-SpiceDB verification must prove tenant-owner inheritance, inventory-owner visibility, viewer/editor direct grants, edit denial for viewers, share denial for non-owners, unrelated-user denial, and cross-tenant denial.
- Real-SpiceDB verification must not run during default `make test`; it must be an explicit command because it requires Docker.
