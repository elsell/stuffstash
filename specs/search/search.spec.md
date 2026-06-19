# Search Spec

## Purpose

Stuff Stash needs search that works across assets, locations, custom asset types, custom fields, and inventories the user can access.

Search should support both exact lookup and fuzzy discovery.

## Scope

This spec covers initial search requirements.

This spec does not define final indexes, ranking, query syntax, highlighting, or external search integrations.

## Requirements

- Search must be tenant-scoped.
- Search results must include only inventories and resources the user is authorized to view.
- SpiceDB filtering must be respected for search.
- Search must support exact search.
- Search must support fuzzy search.
- Search should search all relevant fields, including asset title, asset description, location names, and custom field values.
- Search should support filtering or matching by custom asset type once custom asset types are implemented.
- Search must support custom fields when the user can access the inventory and field definition.
- Search must preserve tenant isolation and inventory isolation.
- Search must support pagination.
- Search must expose stable result types so clients can render assets, locations, inventories, and future result types safely.
- Search implementation must live behind ports and adapters.

## Initial Implementation Direction

- PostgreSQL is the initial search backend.
- External search systems should not be added until PostgreSQL is insufficient and a spec justifies the added operational cost.
- Search adapters must not leak PostgreSQL-specific query details into domain logic.
- Search indexes and JSONB search behavior must be specified before implementation.

## Conversational Use

- Conversational inventory flows may use search to resolve asset and location references.
- Search results used by conversational flows must respect the initiating user's authorization.
- Fuzzy matches must not cause unsafe actions without clarification or confirmation when ambiguity exists.

## Testing

- Tests must verify exact search, fuzzy search, custom asset type filtering, custom field search, pagination, tenant filtering, inventory filtering, and authorization filtering.
- Tests must verify that unauthorized resources do not appear in results.
- PostgreSQL-backed tests are required for PostgreSQL-specific search behavior.

## Open Questions

- What ranking rules should be used first?
- Which custom field types should be searchable first?
- How should search handle aliases and synonyms?
- How should SpiceDB filtering be applied efficiently for larger inventories?
