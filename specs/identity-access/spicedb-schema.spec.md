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

  permission view = owner + admin
  permission configure = owner + admin
  permission create_inventory = owner + admin
}

definition inventory {
  relation tenant: tenant
  relation owner: user
  relation editor: user
  relation viewer: user

  permission view = owner + editor + viewer + tenant->view
  permission configure = owner + tenant->configure
  permission share = owner + tenant->configure
  permission edit = owner + editor + tenant->configure
}
```

## Tracer Bullet Permissions

The first API slice must check:

- `tenant.create_inventory` before creating an inventory inside a tenant.
- `inventory.view` before returning an inventory.

Creating a tenant grants the creating user the tenant `owner` relationship.

Creating an inventory grants the creating user the inventory `owner` relationship and links the inventory to its tenant.

## Adapter Strategy

- The authorization port must support permission checks.
- The authorization port must support relationship writes needed by the application service.
- The first implementation may use an in-memory fake for tests and the initial tracer bullet.
- The production adapter must use SpiceDB.
- Application services must not import SpiceDB client types.

## Verification

- Tests must prove tenant owner inheritance.
- Tests must prove cross-tenant denial.
- Tests must prove unauthenticated requests never reach privileged behavior.
- Tests must prove a user without a relationship cannot create or list inventory resources.
