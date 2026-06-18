# Bounded Contexts Spec

## Purpose

Stuff Stash needs clear ownership boundaries before detailed domain specs and code are written.

## Scope

This spec defines the initial bounded contexts and what each one owns.

This spec does not define full aggregates, persistence schemas, REST endpoints, or UI flows.

## Initial Bounded Contexts

### Assets

The asset context owns inventory items.

Assets must be flexible enough to represent household items, tools, electronics, medicine, pantry goods, documents, containers, and other things without making any one category the product focus.

The asset context owns:

- Asset identity.
- Asset title and description.
- Asset lifecycle once specified.
- Asset containment behavior once specified.
- Asset custom field values.
- Asset domain validation.

### Inventories

The inventory context owns inventory workspaces inside a tenant.

An inventory is the primary unit for organizing, sharing, querying, and configuring a set of assets. A tenant may have multiple inventories, such as a general household inventory, a tools inventory, or a medicine inventory.

The inventory context owns:

- Inventory identity.
- Inventory membership relationships.
- Inventory-scoped settings.
- Inventory-scoped custom field definitions.
- Inventory-level sharing behavior once specified.

### Identity And Access

The identity and access context owns principals, tenants, authentication, and authorization relationships.

The identity and access context owns:

- Users.
- Tenants.
- OIDC and SSO authentication.
- Relationship-based authorization through SpiceDB.
- Tenant security boundaries.
- Sharing relationships inspired by Google Drive-style access.

### Locations

The location context owns named places and containment paths inside an inventory.

Locations are inventory-scoped and hierarchical. The final implementation may model locations and container assets through a shared containment abstraction, but the behavior must be specified before code.

The location context owns:

- Location identity.
- Location names and hierarchy.
- Moving locations within an inventory.
- Resolving spoken or typed location references.

### Agent And Model

The agent and model context owns language, speech, and orchestration adapters for low-friction interaction.

It does not own asset or inventory behavior. It translates natural language into application operations owned by other contexts.

The agent and model context owns:

- Speech-to-text ports.
- Text-to-speech ports.
- Language model ports.
- Conversational action planning.
- Provider adapters.
- Real-time voice and text interaction mechanics once specified.

### Expiration

The expiration context remains a candidate context.

Expiration may become:

- A first-class domain concept.
- A custom field type.
- A policy or reminder capability.
- A combination of these.

This must be decided before expiration-specific behavior is implemented.

## Cross-Context Rules

- Tenants are the top-level security boundary.
- Inventories live inside tenants.
- Assets live inside inventories.
- Locations live inside inventories.
- Users may belong to multiple tenants.
- Users may have access to one or more inventories inside a tenant.
- Custom field definitions may be tenant-scoped or inventory-scoped.
- Tenant-scoped custom field definitions flow down into inventories.
- Inventory-scoped custom field definitions apply only inside that inventory.
- Conversational actions must execute through the same application operations and authorization checks as any other adapter.

## Open Questions

- Should locations and container assets be implemented as one shared containment aggregate or as separate aggregates with a shared containment service?
- Are consumables assets, quantities, lifecycle events, or a separate concept?
- Should expiration be a first-class domain context or primarily a custom field and policy feature?
- Should search be its own bounded context or remain an adapter/read-model concern at first?
- Should audit/history be its own bounded context from the start?
