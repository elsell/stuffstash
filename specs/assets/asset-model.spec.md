# Asset Model Spec

## Purpose

Stuff Stash assets represent things in an inventory.

The asset model must stay lean, flexible, and useful across many categories without making the product feel like it is only for electronics, tools, medicine, groceries, or any other narrow category.

## Scope

This spec covers the initial asset model direction.

This spec does not define the final persistence schema, full lifecycle model, consumable model, search behavior, media attachments, or every REST endpoint.

## Requirements

- An asset is anything that is part of an inventory.
- Asset IDs must use ULIDs.
- An asset must belong to exactly one inventory.
- An asset must be inside the tenant security boundary of its inventory.
- The stable asset core should be lean.
- The initial stable core should include:
  - ID.
  - Inventory ID.
  - Title.
  - Description.
  - Containment reference.
- Assets must support custom field values.
- Asset custom field values must be validated against tenant-scoped and inventory-scoped custom field definitions.
- The asset domain must not be overfit to one product category.

## Containment

- Assets may be nested or bundled through the shared containment model.
- Some assets may behave as containers.
- The system must support placing assets inside other assets when the parent asset is a container-like thing.
- The system must support hierarchical places inside an inventory.
- The system must support moving assets and location-like containers without excessive friction.
- The implementation must avoid duplicating hierarchy logic for locations and container assets.
- Locations and assets are separate domain concepts with a shared containment abstraction.

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
- Tests must verify containment behavior.
- Tests must verify asset nesting or bundling once implemented.
- Security-sensitive asset behavior must have adversarial end-to-end tests before public interaction points expose it.

## Open Questions

- How should consumables be modeled?
- Can assets move between inventories?
- What asset fields should be first-class rather than custom fields?
