# Location Model Spec

## Purpose

Stuff Stash needs a flexible way to describe where things are.

Locations must support ordinary household language such as garage, shelf, bin, closet, room, drawer, or wire rack.

## Scope

This spec covers the initial location model direction.

This spec does not decide the final aggregate split between locations and assets, or the fuzzy matching implementation for natural language.

## Requirements

- Locations are inventory-scoped.
- Locations must support hierarchy.
- A location may have a parent location.
- Location nesting may be arbitrarily deep unless a future performance or usability spec defines a limit.
- Locations must be movable within an inventory.
- Moving a location must preserve its children unless a future spec defines a different behavior.
- Locations must be tenant-isolated through their inventory.
- Location names must be usable in conversational inventory flows.
- Location creation from natural language must go through application services, validation, authorization, and audit behavior.

## Relationship To Assets

- Assets may be placed in locations.
- Assets may be placed inside other assets when the parent asset behaves like a container.
- Locations and container assets may share containment behavior.
- The final implementation must decide whether locations and assets are the same model, separate models, or separate models with a shared containment abstraction.

## Testing

- Tests must verify location hierarchy behavior.
- Tests must verify moving locations.
- Tests must verify inventory isolation.
- Tests must verify tenant isolation.
- Tests must verify authorization for creating, moving, renaming, and deleting locations.
- Conversational location creation and movement must have adversarial end-to-end tests before public interaction points expose it.

## Open Questions

- Are locations a special kind of asset or a separate aggregate?
- Can a location be shared independently from its inventory?
- What deletion or archive behavior should locations have?
- How should duplicate location names be handled within different branches of the hierarchy?
