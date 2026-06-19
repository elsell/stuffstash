# Containment Model Spec

## Purpose

Stuff Stash needs a clear model for where things are and what can contain other things.

Users think in a mix of places and containers: garage, shelf, toolbox, drawer, bin, medicine cabinet, and backpack.

The model should support that with one containment graph instead of separate asset and location hierarchies.

## Scope

This spec defines the initial containment direction for assets, containers, place-like locations, and movement inside one inventory.

This spec does not define search indexes, cross-inventory movement, bulk archive behavior, permanent deletion, or every future move rule.

## Decision

- Assets and locations are unified in the containment model.
- The persisted node is an asset.
- A location is a place-like asset with asset kind `location`.
- A normal stored thing is an asset with asset kind `item`.
- A movable container thing is an asset with asset kind `container`.
- Container and location assets may contain child assets.
- Item assets may be contained but must not contain children.
- User interfaces, conversational flows, and docs may say "location" when the asset kind is `location`.

## Requirements

- Containment must be inventory-scoped.
- Containment must preserve tenant isolation through the inventory.
- A child asset must belong to the same tenant and inventory as its parent asset.
- A root asset has no parent asset.
- The system must support moving an asset between location assets.
- The system must support moving an asset into a container asset.
- The system must support moving a container asset with its contained assets.
- The system must support moving a location asset with its child location assets and contained assets.
- Moving a container or location asset must preserve all descendants by moving only the selected node.
- Moving an asset to the inventory root must be supported by clearing the parent asset reference.
- Moving an asset between inventories is out of scope for the first movement slice.
- The system must prevent containment cycles.
- The system must prevent an asset from containing itself.
- The system must prevent placing a child under an `item` asset.
- The system must support arbitrary containment depth unless future performance or usability specs define a limit.
- The system must expose containment through application operations, not direct persistence operations.
- The first implementation must enforce containment invariants in application/domain code before repository persistence.
- The repository adapter must defensively preserve tenant and inventory boundaries.

## Initial Lifecycle Interaction

- The first asset lifecycle states are `active` and `archived`.
- Archived assets must not be valid containment targets.
- Active contained assets must not silently disappear when a container is archived.
- Archive and restore behavior is specified in `specs/assets/asset-model.spec.md`.
- The first lifecycle slice rejects archiving assets with active children.

## Conversational Use

- Conversational flows may create location assets and move assets through containment operations.
- Conversational flows may move assets into container assets when the target is unambiguous.
- Ambiguous containment targets must trigger clarification.
- Model output must not bypass containment validation.

## Testing

- Tests must verify moving assets between location assets.
- Tests must verify moving assets into container assets.
- Tests must verify moving container assets with children.
- Tests must verify moving location assets with children.
- Tests must verify moving assets to the root.
- Tests must verify cycle prevention.
- Tests must verify item assets cannot contain children.
- Tests must verify tenant, inventory, and authorization boundaries.

## Open Questions

- What exact user-facing language should distinguish locations from container assets?
- Can a container asset be converted into a non-container asset?
- What should happen when a container with children is archived?
