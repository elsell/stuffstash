# Asset Model Spec

## Purpose

Stuff Stash assets represent things in an inventory.

The asset model must stay lean, flexible, and useful across many categories without making the product feel like it is only for electronics, tools, medicine, groceries, or any other narrow category.

## Scope

This spec covers the initial asset model direction and the first persistence shape for assets.

This spec does not define the full lifecycle model, consumable model, search behavior, media attachments, or every REST endpoint.

## Requirements

- An asset is anything that is part of an inventory.
- Place-like locations are represented as assets.
- Asset IDs must use ULIDs.
- An asset must belong to exactly one inventory.
- An asset must be inside the tenant security boundary of its inventory.
- The stable asset core should be lean.
- The initial stable core should include:
  - ID.
  - Tenant ID.
  - Inventory ID.
  - Kind.
  - Title.
  - Description.
  - Parent asset reference.
  - Custom field values.
  - Lifecycle state.
- Assets must support custom field values.
- Asset custom field values must be validated against tenant-scoped and inventory-scoped custom field definitions.
- The first asset slice may only accept an empty custom field map until custom field definitions are implemented.
- Non-empty custom field values must be rejected until the relevant definition and validation behavior exists.
- The asset domain must not be overfit to one product category.
- Asset kind must be represented as a domain enumeration, not ad hoc strings.
- The initial asset kinds are:
  - `item`: a normal thing that can be stored or moved.
  - `container`: a thing that can contain other assets.
  - `location`: a place-like container such as a garage, shelf, closet, room, bin, drawer, or wire rack.
- The domain model may expose location-friendly language to users, but persistence must store location-like nodes as assets.
- The initial lifecycle states are `active` and `archived`.
- New assets are created as `active`.
- The first asset slice must persist lifecycle state but must not expose archive or unarchive operations.
- Archived assets must not be valid containment targets.

## Initial Persistence Shape

The first asset implementation must use one `assets` table for items, containers, and location-like nodes.

The table must include:

- `id`: ULID primary key.
- `tenant_id`: tenant security boundary copied from the inventory for efficient scoping and defensive checks.
- `inventory_id`: inventory that owns the asset.
- `parent_asset_id`: nullable self-reference for containment.
- `kind`: domain asset kind enumeration.
- `title`: short display name.
- `description`: optional longer text.
- `custom_fields`: PostgreSQL JSONB values for validated custom field data.
- `lifecycle_state`: domain lifecycle enumeration.
- timestamps managed by GORM.

Persistence rules:

- `tenant_id` must reference `tenants.id`.
- `inventory_id` must reference `inventories.id`.
- `parent_asset_id` must reference `assets.id` when present.
- A parent asset must be in the same tenant and inventory as the child.
- Same-tenant and same-inventory parentage must be enforced by application/domain validation and by repository adapter defensive checks before commit.
- The first implementation must not rely on a plain self-referencing foreign key alone to enforce parent scope.
- Only `container` and `location` assets may be used as parents.
- `parent_asset_id` must be null for root assets.
- `parent_asset_id` must never reference the asset itself.
- Containment cycles must be prevented before persistence commits.
- `custom_fields` must default to an empty object, not null.
- Domain and application code must not manipulate raw JSONB.

## Containment

- Assets may be nested or bundled through the shared containment model.
- Container and location assets may contain other assets.
- Item assets must not contain other assets unless their kind is changed through an explicit future operation.
- The system must support hierarchical place-like assets inside an inventory.
- The system must support moving assets and location-like containers without excessive friction.
- The implementation must not duplicate hierarchy logic for locations and container assets.
- Locations are user-facing concepts backed by location-kind assets.

## Consumables

Consumables remain an open design question.

The system must eventually support things that can be used up, such as medicine, fertilizer, pantry goods, batteries, or cleaning supplies. Before implementation, a future spec must decide whether consumables are represented as:

- Assets with quantity.
- Asset lots.
- Lifecycle events.
- A separate consumable concept.
- A combination of these.

## Testing

- Tests must verify asset creation, asset updates, custom field validation, tenant isolation, inventory isolation, and authorization.
- Tests must verify containment behavior for item, container, and location asset kinds.
- Tests must verify asset nesting or bundling once implemented.
- Tests must verify location-like assets are persisted through the same asset repository path as other assets.
- Security-sensitive asset behavior must have adversarial end-to-end tests before public interaction points expose it.

## Open Questions

- How should consumables be modeled?
- Can assets move between inventories?
- What asset fields should be first-class rather than custom fields?
- Can users convert an existing asset from one kind to another?
