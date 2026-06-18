# Containment Model Spec

## Purpose

Stuff Stash needs a clear model for where things are and what can contain other things.

Users think in a mix of places and containers: garage, shelf, toolbox, drawer, bin, medicine cabinet, and backpack. The model should support that without forcing locations and assets to be the same concept.

## Scope

This spec defines the initial containment direction for assets and locations.

This spec does not define persistence tables, search indexes, REST endpoints, or every move rule.

## Decision

- Assets and locations are separate domain concepts.
- Assets and locations must share containment behavior through a common containment abstraction.
- A location is a named place inside an inventory.
- An asset is a thing inside an inventory.
- Some assets may be containers.
- A container asset may contain other assets.
- A location may contain assets and child locations.

## Requirements

- Containment must be inventory-scoped.
- Containment must preserve tenant isolation through the inventory.
- A contained asset must belong to the same inventory as its container or location.
- A child location must belong to the same inventory as its parent location.
- The system must support moving an asset between locations.
- The system must support moving an asset into a container asset.
- The system must support moving a container asset with its contained assets.
- The system must support moving a location with its child locations and contained assets.
- The system must prevent containment cycles.
- The system must prevent placing a location inside an asset unless a future spec explicitly introduces that behavior.
- The system must support arbitrary location depth unless future performance or usability specs define a limit.
- The system must expose containment through application operations, not direct persistence operations.

## Initial Lifecycle Interaction

- The first asset lifecycle states are `active` and `archived`.
- Archived assets must not be valid containment targets.
- Active contained assets must not silently disappear when a container is archived.
- Archive behavior for containers must be specified before container archiving is implemented.

## Conversational Use

- Conversational flows may create locations and move assets through containment operations.
- Conversational flows may move assets into container assets when the target is unambiguous.
- Ambiguous containment targets must trigger clarification.
- Model output must not bypass containment validation.

## Testing

- Tests must verify moving assets between locations.
- Tests must verify moving assets into container assets.
- Tests must verify moving container assets with children.
- Tests must verify moving locations with children.
- Tests must verify cycle prevention.
- Tests must verify tenant, inventory, and authorization boundaries.

## Open Questions

- What exact user-facing language should distinguish locations from container assets?
- Can a container asset be converted into a non-container asset?
- What should happen when a container with children is archived?
