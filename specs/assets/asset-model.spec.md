# Asset Model Spec

## Purpose

Stuff Stash assets represent things in an inventory.

The asset model must stay lean, flexible, and useful across many categories without making the product feel like it is only for electronics, tools, medicine, groceries, or any other narrow category.

## Scope

This spec covers the initial asset model direction, asset create/update behavior, movement inside an inventory, and the first persistence shape for assets.

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
- Assets must support an optional custom asset type once the custom asset type slice is implemented.
- Custom asset types are user-defined metadata/classification overlays on normal assets. They must not replace the base asset kind enumeration.
- Asset kind controls base containment behavior. Custom asset type controls user-facing classification and which type-specific custom field definitions apply.
- Asset custom field values must be validated against tenant-scoped and inventory-scoped custom field definitions.
- Asset create may accept non-empty custom field values when every value is validated against the effective custom field definitions for the target inventory and the asset's custom asset type.
- Asset update may replace title, description, parent asset reference, and custom field values.
- Asset update must validate custom field values against the same effective custom field definitions used by asset create.
- Asset update must not change asset ID, tenant ID, inventory ID, kind, custom asset type, or lifecycle state in the first update slice.
- Asset movement is represented by updating an asset's parent asset reference.
- Cross-inventory movement is not supported in the first update slice.
- Clearing `parentAssetId` moves an asset to the inventory root.
- Non-empty custom field values must be rejected when the relevant definition is missing or validation fails.
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
- `custom_asset_type_id`: optional reference to a custom asset type once custom asset types are implemented.
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
- Custom asset type must not change the base containment rules. A `medicine` custom asset type can still be an `item`, `container`, or `location` depending on its base asset kind.
- `custom_fields` must default to an empty object, not null.
- Domain and application code must not manipulate raw JSONB.
- Update persistence must preserve the same tenant, inventory, kind, and lifecycle state for the target asset.
- Update persistence must defensively reject cross-tenant and cross-inventory parent references.

## Containment

- Assets may be nested or bundled through the shared containment model.
- Container and location assets may contain other assets.
- Item assets must not contain other assets unless their kind is changed through an explicit future operation.
- The system must support hierarchical place-like assets inside an inventory.
- The system must support moving assets and location-like containers without excessive friction.
- Moving a container or location must move the node with its existing descendants; children must not be rewritten one by one.
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
- Tests must verify asset nesting, moving an asset between parents, moving a container or location with descendants, moving to root, self-parent rejection, cycle rejection, and item-as-parent rejection.
- Tests must verify location-like assets are persisted through the same asset repository path as other assets.
- Security-sensitive asset behavior must have adversarial end-to-end tests before public interaction points expose it.

## Open Questions

- How should consumables be modeled?
- Can assets move between inventories, or should cross-inventory movement remain an export/import-style workflow?
- What asset fields should be first-class rather than custom fields?
- Can users convert an existing asset from one kind to another?
- Can users change an existing asset's custom asset type, and how should values for no-longer-applicable custom fields be handled?
