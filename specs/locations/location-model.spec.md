# Location Model Spec

## Purpose

Stuff Stash needs a flexible way to describe where things are.

Locations must support ordinary household language such as garage, shelf, bin, closet, room, drawer, or wire rack.

## Scope

This spec covers the initial location model direction from the user's point of view.

This spec does not define separate persistence tables, REST endpoints, or the fuzzy matching implementation for natural language.

## Decision

- Locations are user-facing place concepts backed by assets with kind `location`.
- The system must not implement a separate location hierarchy table for the first asset slice.
- Location behavior must use the shared asset containment model.
- A future spec may reintroduce separate location persistence only if it explains why the unified model is insufficient.

## Requirements

- Locations are inventory-scoped.
- Locations must support hierarchy.
- A location asset may have a parent location or container asset.
- Location nesting may be arbitrarily deep unless a future performance or usability spec defines a limit.
- Locations must be movable within an inventory.
- Moving a location must preserve its children unless a future spec defines a different behavior.
- Locations must be tenant-isolated through their inventory.
- Location names must be usable in conversational inventory flows.
- Location creation from natural language must go through application services, validation, authorization, and audit behavior.

## Relationship To Assets

- Assets may be placed in location assets.
- Assets may be placed inside container assets.
- Locations and container assets share the same containment behavior.
- The domain may expose location-specific operations for clarity, but those operations must delegate to the asset containment model.
- Location creation creates an asset with kind `location`.

## Testing

- Tests must verify location hierarchy behavior.
- Tests must verify moving locations.
- Tests must verify location operations persist location assets, not separate location rows.
- Tests must verify inventory isolation.
- Tests must verify tenant isolation.
- Tests must verify authorization for creating, moving, and renaming locations.
- Location archive and restore use the asset lifecycle rules in `specs/assets/asset-model.spec.md`.
- Permanent deletion behavior for locations is out of scope.
- Conversational location creation and movement must have adversarial end-to-end tests before public interaction points expose it.

## Open Questions

- Can a location be shared independently from its inventory?
- What permanent deletion or bulk subtree archive behavior should locations have?
- How should duplicate location names be handled within different branches of the hierarchy?
