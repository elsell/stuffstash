# Bounded Contexts Spec

## Purpose

Stuff Stash needs clear ownership boundaries before detailed domain specs and code are written.

## Scope

This spec defines the initial bounded contexts and what each one owns.

This spec does not define full aggregates, persistence schemas, REST endpoints, or UI flows.

## Initial Bounded Contexts

### Assets

The asset context owns inventory nodes.

Assets must be flexible enough to represent household items, tools, electronics, medicine, pantry goods, documents, containers, place-like locations, and other things without making any one category the product focus.

The asset context owns:

- Asset identity.
- Asset kind, including item, container, and location-like assets.
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

The location context owns user-facing named place behavior inside an inventory.

Locations are inventory-scoped and hierarchical from the user's point of view. In the initial implementation, a location is backed by an asset with kind `location`, not by a separate persisted location hierarchy.

The location context owns:

- Location language and UX concepts.
- Location hierarchy behavior through asset containment.
- Moving location assets within an inventory.
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

### Audit And History

The audit and history context owns durable records of what changed, who changed it, and how it changed.

The audit and history context owns:

- Audit records for state-changing actions.
- Undo metadata and compensating action behavior once specified.
- Safe audit views scoped by tenant and inventory authorization.

### Search

The search context owns user-facing lookup and discovery across authorized resources.

The search context owns:

- Exact search.
- Fuzzy search.
- Search result types.
- Search filtering through tenant, inventory, and SpiceDB authorization.
- Search read models and indexes once specified.

### Media

The media context owns inventory file and image attachments.

The media context owns:

- Attachment metadata.
- Blob storage ports.
- Attachment authorization.
- Image and file access behavior.

### Data Portability

The data portability context owns import, export, and backup-style data movement.

The data portability context owns:

- JSON export.
- CSV export.
- Future import formats.
- Future media export packaging.

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
- Every state-changing action must produce audit history.
- Search must return only resources the user is authorized to view.
- Media attachments are scoped by tenant and inventory through the resource they attach to.
- Import and export must preserve tenant, inventory, authorization, and audit boundaries.
- Domain packages must not import other domain packages directly.
- Cross-domain coordination must happen through application services, ports, domain events, or other explicitly specified boundaries.
- The repository must maintain a structural check for direct cross-domain imports once domain packages exist.

## Open Questions

- Are consumables assets, quantities, lifecycle events, or a separate concept?
- Should expiration be a first-class domain context or primarily a custom field and policy feature?
- Should search be its own bounded context or remain an adapter/read-model concern at first?
