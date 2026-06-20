# Search Spec

## Purpose

Stuff Stash needs search that works across assets, locations, custom asset types, custom fields, and inventories the user can access.

Search should support both exact lookup and fuzzy discovery.

## Scope

This spec covers initial search requirements and the first asset search API slice.

This spec does not define final indexes, ranking, advanced query syntax, highlighting, external search integrations, or conversational disambiguation behavior.

## Requirements

- Search must be tenant-scoped.
- Search results must include only inventories and resources the user is authorized to view.
- SpiceDB filtering must be respected for search.
- Search must discover viewable inventory IDs through a dedicated authorization visibility/query port instead of enumerating every inventory and checking each one individually in application code.
- The authorization visibility/query port must be behind adapters so a SpiceDB lookup implementation, a local in-memory implementation, and future optimized implementations can be swapped without changing search application behavior.
- The search application service must pass authorized inventory IDs into the search repository. The search repository must not decide authorization itself.
- Search must support exact search.
- Search must support fuzzy search.
- Search should search all relevant fields, including asset title, asset description, location names, and custom field values.
- Search should support filtering or matching by custom asset type.
- Search must support custom fields when the user can access the inventory and field definition.
- Search must preserve tenant isolation and inventory isolation.
- Search must support pagination.
- Search must expose stable result types so clients can render assets, locations, inventories, and future result types safely.
- Search implementation must live behind ports and adapters.

## First API Slice

The first API slice is asset search:

- `GET /tenants/{tenantId}/search/assets`
- Query parameter `q` is required, trimmed, and must be between 1 and 120 characters.
- Query parameter `mode` supports `fuzzy` and `exact`. It defaults to `fuzzy`.
- Query parameter `customAssetTypeId` filters to assets with that custom asset type when present.
- Query parameter `lifecycleState` supports `active`, `archived`, and `all`. It defaults to `active`.
- `limit` and `cursor` follow the standard collection pagination contract.
- Results must be ordered deterministically by inventory ID and asset ID for the first slice.
- Results must include a stable `type` field. The first value is `asset`.
- Results must include the matching asset and simple match metadata so clients can explain why a result appeared.
- Location-like assets are returned as assets with kind `location`.
- The first slice searches asset title, description, custom field values, custom asset type key/display name/description, and attachment file name/content type.
- Exact search uses case-insensitive whole-value equality for fields and metadata.
- Fuzzy search uses case-insensitive substring matching.

## Initial Implementation Direction

- PostgreSQL is the initial search backend.
- External search systems should not be added until PostgreSQL is insufficient and a spec justifies the added operational cost.
- Search adapters must not leak PostgreSQL-specific query details into domain logic.
- The first PostgreSQL-backed implementation may scan tenant-scoped candidate rows through GORM and perform matching in adapter code to avoid raw SQL while the data model is small.
- Before larger-scale search, define generated columns, indexes, JSONB search behavior, ranking, and authorization filtering strategy in this spec.
- Before search is used at meaningful data volume, replace adapter-side candidate scanning with indexed PostgreSQL search behavior covered by PostgreSQL-backed tests.
- Before tenants can have many inventories, replace the local in-memory visibility adapter with a production SpiceDB lookup implementation if the current adapter is not already using SpiceDB lookup APIs.

## Conversational Use

- Conversational inventory flows may use search to resolve asset and location references.
- Search results used by conversational flows must respect the initiating user's authorization.
- Fuzzy matches must not cause unsafe actions without clarification or confirmation when ambiguity exists.

## Testing

- Tests must verify exact search, fuzzy search, custom asset type filtering, custom field search, attachment metadata search, pagination, tenant filtering, inventory filtering, lifecycle filtering, and authorization filtering.
- Tests must verify that unauthorized resources do not appear in results.
- PostgreSQL-backed tests are required for PostgreSQL-specific search behavior.

## Open Questions

- What ranking rules should be used first?
- Which custom field types should be searchable first?
- How should search handle aliases and synonyms?
- How should SpiceDB filtering be applied efficiently for larger inventories?
