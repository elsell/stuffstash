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
- Locations are non-portable place concepts. They may be moved within the
  containment hierarchy, but they must not be checked out through the asset
  checkout workflow.

## Mobile Place Workspace

- Mobile may use `Place` as the concise navigation and kind label for a
  location asset while preserving `location` as the domain and API kind.
- A place workspace must prioritize the place identity and the question
  "what is here?" over generic asset-maintenance controls.
- A root-level place must not present `No location` as if placement data were
  missing. Mobile may omit the parent-placement row or use quiet top-level
  context. Nested places must retain structured, clickable parent breadcrumbs.
- Place photos must use the same shared asset-detail gallery, viewer, upload,
  and removal behavior as items and containers. The no-photo state must keep a
  stable but compact media region with one obvious add-photo affordance.
- A place workspace must distinguish direct navigable spaces from items found
  anywhere beneath the place. Direct child locations and containers form the
  navigable `Spaces` section. Descendant items form the `Items` section and
  retain a structured relative path when they are nested below another space.
- Counts and headings must state their scope. Copy such as `In Garage` or
  `Items in Garage` must not describe only immediate children as though it
  represented every descendant.
- Place-specific copy should use natural spatial language such as `In Garage`
  and `Nothing here yet`; container-specific copy may continue to use
  `Inside`.
- The contents heading must remain visually adjacent to its rows. Spatial and
  maintenance controls must not create a large gap between a section heading
  and the content it labels.
- The filled primary action for an editable active place is `Add item here`.
  `Move items here` remains available with quieter treatment. Generic
  operations must name their target clearly, such as `Move place`, and belong
  in a quiet management area or overflow rather than between a contents
  heading and its rows.

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
