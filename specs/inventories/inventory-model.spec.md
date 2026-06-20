# Inventory Model Spec

## Purpose

Stuff Stash needs inventories as first-class organizational units inside tenants.

Inventories let a tenant separate access, settings, custom asset types, custom fields, and queries for different sets of assets.

## Scope

This spec covers the initial inventory model.

This spec does not define REST endpoints, UI navigation, billing, import/export, or every inventory setting.

## Requirements

- A tenant may contain multiple inventories.
- An inventory must belong to exactly one tenant.
- Assets must belong to an inventory.
- Locations must belong to an inventory.
- Inventory names must be unique enough within a tenant for users to distinguish them.
- Users may have access to one or more inventories within a tenant.
- Users with access to multiple inventories should be able to query across those inventories when authorized.
- Inventory collection APIs must use cursor pagination and must paginate only inventories visible to the caller.
- Inventory-level settings must be separate from tenant-level settings.
- Inventory-scoped custom asset type definitions must apply only inside that inventory.
- Tenant-scoped custom asset type definitions must be available inside tenant inventories unless a future spec defines override behavior.
- Inventory-scoped custom field definitions must apply only inside that inventory.
- Tenant-scoped custom field definitions must be available inside tenant inventories unless a future spec defines override behavior.

## Examples

A tenant might contain:

- General household inventory.
- Tools inventory.
- Medicine inventory.

The product must not assume that every tenant has only one inventory.

## Testing

- Tests must verify tenant isolation between inventories.
- Tests must verify authorized cross-inventory queries.
- Tests must verify that unauthorized inventories are excluded from queries.
- Tests must verify that inventory-scoped custom fields do not leak into other inventories.
- Tests must verify that inventory-scoped custom asset types do not leak into other inventories.

## Open Questions

- Can inventory settings override tenant settings?
- Can assets move between inventories?
- If assets can move between inventories, what happens to custom field values that are not defined in the destination inventory?
- If assets can move between inventories, what happens to custom asset type assignments that are not defined in the destination inventory?
- Inventories have `active` and `archived` lifecycle states. Archive, restore, and hard-delete behavior is defined by `specs/platform/resource-lifecycle.spec.md`.
