# Asset Tags Spec

## Purpose

Stuff Stash must support native tags for fast, familiar inventory organization and import parity with existing home inventory tools.

Tags are short inventory-scoped labels such as `Tools`, `Camping`, or `Kids`. They are not custom fields. Custom fields describe structured per-asset metadata; tags provide reusable labels that can be assigned to many assets and used for browsing, filtering, and search.

## Scope

This spec covers the first native tag slice:

- Inventory-scoped tag definitions.
- Assignment of tags to assets.
- Homebox CSV and live import tag preservation.
- REST API contracts for tag management and asset tag assignment.
- Web and mobile display/edit expectations.
- Audit, authorization, persistence, and tests.

This spec does not define tag hierarchies, global tenant-level tags, tag merge workflows, rule-based automatic tagging, tag permissions independent from inventory permissions, or tag analytics.

## Domain Model

`AssetTag` is an inventory-scoped aggregate.

Fields:

- `id`: ULID.
- `tenantId`: tenant security boundary.
- `inventoryId`: inventory scope.
- `key`: stable normalized key, unique within an inventory.
- `displayName`: user-facing tag name.
- `color`: optional `#RRGGBB` color.
- `lifecycleState`: tag lifecycle state.
- `createdAt`: application time supplied by the injected clock port.
- `updatedAt`: application time supplied by the injected clock port.

The first lifecycle states are:

- `active`: the tag can be assigned and shown normally.
- `archived`: the tag is hidden from normal selection and cannot be newly assigned.

Asset tag assignment is a many-to-many relationship between active assets and active tags in the same tenant and inventory.

## Invariants

- Tag keys are unique inside `(tenantId, inventoryId)`.
- Tag keys must be normalized lowercase ASCII slugs using letters, digits, and hyphens.
- Tag display names must be non-empty after trimming and fit within bounded UI-friendly length.
- Tag colors are optional. Non-empty colors must normalize to uppercase `#RRGGBB`.
- Tags cannot cross tenant or inventory boundaries.
- Asset assignment must fail if the asset does not exist in the requested tenant and inventory.
- Asset assignment must fail if any tag does not exist in the requested tenant and inventory.
- Asset assignment must fail for archived tags.
- Archiving a tag does not delete audit history or imported source history.
- Application time must come from the injected clock port.

## Application Operations

Required commands:

- `CreateAssetTag`
  - Inputs: tenant ID, inventory ID, authenticated principal, display name, optional key, optional color, source, request ID.
  - Output: created tag, or the existing active tag when the normalized key already exists in the inventory.
- `UpdateAssetTag`
  - Inputs: tenant ID, inventory ID, tag ID, authenticated principal, optional display name, optional color, source, request ID.
  - Output: updated tag.
- `ArchiveAssetTag`
  - Inputs: tenant ID, inventory ID, tag ID, authenticated principal, source, request ID.
  - Output: archived tag.
- `SetAssetTags`
  - Inputs: tenant ID, inventory ID, asset ID, authenticated principal, complete tag ID list, source, request ID.
  - Output: updated assignment state through subsequent asset reads.

Required queries:

- `ListAssetTags`: active tags in one inventory, cursor paginated.
- `AssetTagsByAsset`: active assigned tags for one asset.
- `AssetTagsByAssets`: active assigned tags for asset list responses.
- Asset search must match assigned active tag display names and keys for assets the caller is authorized to search.

## REST API

The first REST endpoints are:

- `GET /tenants/{tenantId}/inventories/{inventoryId}/tags`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/tags`
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/tags/{tagId}`
- `DELETE /tenants/{tenantId}/inventories/{inventoryId}/tags/{tagId}`

Creating a tag with a normalized key that already belongs to an active tag in the inventory must return that existing tag instead of failing. This keeps inline tag creation safe for stale clients and concurrent saves. Archived tags keep their keys reserved; creating a new tag with an archived tag key must fail until a future merge or restore workflow is specified.

Asset create and update requests must accept:

- `tagIds`: complete assigned tag ID list.

Asset detail and list responses must include:

- `tags`: ordered compact assigned tag objects with ID, key, display name, and optional color.

Asset search results:

- Must include assets when the search query matches an assigned active tag display name or tag key.
- Must not match archived tags.
- Must preserve the existing tenant, inventory, lifecycle, checkout, and authorization filters before evaluating tag matches.

Authorization:

- Listing tags and reading assigned tags requires `inventory.view`.
- Creating, updating, archiving, and assigning tags requires `inventory.edit_asset`.
- Endpoints must fail safely without revealing cross-tenant, cross-inventory, or unauthorized resource existence.

## Homebox Import

Homebox imports must preserve tags as native Stuff Stash tags.

CSV imports:

- Read `HB.labels` when present.
- Fall back to `HB.tags` for older exports.
- Split values on semicolons and commas.
- Normalize duplicates by tag key.
- Do not create the legacy `homebox-tags` custom field once native tags are supported.

Live imports:

- Read tags from item detail responses.
- Preserve tag names.
- Preserve tag color when Homebox provides one. Accepted Homebox color values are `#RRGGBB` or `RRGGBB`; imported colors must normalize to uppercase `#RRGGBB`.
- Normalize invalid or absent colors to no color.

Import planning must include reusable tag definitions and per-asset tag key assignments. Import execution must create or reuse tag definitions before creating assets, then assign imported assets to the created or reused tags.

Import job counts must include:

- `tags`
- `tagsCreated`
- `tagsExisting`

## User Experience

Web and mobile clients must show assigned tags as compact chips in asset list and detail views.

Tag chips must:

- Show the display name.
- Use the tag color as a small swatch when a color is present.
- Keep the asset title, photo, kind, parent/location, checkout state, and lifecycle state visually higher priority than tags.
- Collapse gracefully on narrow screens without causing row height jumps or text overlap.
- In compact list and card contexts, clients may show the first few assigned tags plus a `+N` overflow chip instead of rendering every tag.

Asset create and edit flows must let users:

- Select existing active tags.
- Remove assigned tags.
- Create a new tag inline with display name and optional color where the UI already supports editing asset metadata.

Web and mobile clients must load active inventory tags through client adapter boundaries, map API tag DTOs into client domain models, and submit complete `tagIds` lists on asset create and update. Clients must not treat generated API DTOs as UI domain models.

Clients must reconcile pending inline tag drafts against known inventory tags by normalized key before calling the tag creation API. If a matching active tag is already known locally, the save must reuse that tag ID instead of issuing a duplicate create request.

The first inline creation behavior may create the tag immediately before saving the asset draft. If asset save then fails, the created tag may remain available in the inventory; the UI must keep that state visible by refreshing the active tag list.

Tag controls must remain secondary to the asset title, kind, parent/location, photo, and checkout state.

## Tests

Required coverage:

- Tag value object normalization and validation.
- Repository tenant and inventory scoping.
- REST authentication, authorization, cross-tenant, cross-inventory, and legitimate-principal paths.
- Asset create/update assignment behavior.
- Asset list/detail assigned tag responses.
- Asset search matches assigned active tag display names and keys.
- Homebox CSV `HB.labels` and `HB.tags` mapping.
- Homebox live tag color preservation.
- Import execution tag create/reuse counts and assignment.
