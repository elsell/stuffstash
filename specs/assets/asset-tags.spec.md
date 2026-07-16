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
- Asset search must support active tag ID facets. When tag IDs are supplied, the result set must include only assets assigned to every selected active tag, while preserving any text query, inventory, lifecycle, checkout, kind, authorization, and pagination filters.

## REST API

The first REST endpoints are:

- `GET /tenants/{tenantId}/inventories/{inventoryId}/tags`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/tags`
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/tags/{tagId}`
- `DELETE /tenants/{tenantId}/inventories/{inventoryId}/tags/{tagId}`

Creating a tag with a normalized key that already belongs to an active tag in the inventory must return that existing tag instead of failing. This keeps inline tag creation safe for stale clients and concurrent saves. Archived tags keep their keys reserved; creating a new tag with an archived tag key must fail until a future merge or restore workflow is specified.

Updating a tag with an empty `color` value must clear the optional color. Omitting `color` must leave the current color unchanged.

Asset create and update requests must accept:

- `tagIds`: complete assigned tag ID list.

Asset detail and list responses must include:

- `tags`: ordered compact assigned tag objects with ID, key, display name, and optional color.

Asset search results:

- Must include assets when the search query matches an assigned active tag display name or tag key.
- Must accept `tagIds` as a repeatable query filter. Tag filters compose with the text query and other filters; they must not replace or rewrite the text query.
- Must return matching tagged assets when `tagIds` are supplied and the text query is empty, so clients can browse all assets by selected tags.
- Must include the same compact assigned tag objects on each returned asset summary that asset list and detail responses expose.
- Must not match archived tags.
- Must not use archived tag assignments to satisfy `tagIds` filters.
- Must preserve the existing tenant, inventory, lifecycle, checkout, and authorization filters before evaluating tag matches.
- Web and mobile clients must present tag-backed search matches with user-facing labels such as `Tag`, not raw transport field names like `tag_display_name` or `tag_key`.

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
- Read the Homebox tag list when available and use it to resolve sparse item tag references to canonical tag names and colors.
- Preserve tag names.
- Preserve tag color when Homebox provides one. Accepted Homebox color values are `#RRGGBB` or `RRGGBB`; imported colors must normalize to uppercase `#RRGGBB`.
- Normalize invalid or absent colors to no color.

Import planning must include reusable tag definitions and per-asset tag key assignments. Import execution must create or reuse tag definitions before creating assets, then assign imported assets to the created or reused tags.

Import preview UI must show planned tag definitions, including normalized keys and color swatches when colors are present, so users can review Homebox tag preservation before starting an import.

Import job counts must include:

- `tags`
- `tagsCreated`
- `tagsExisting`

## User Experience

Web and mobile clients must show assigned tags as compact chips in asset list and detail views.
Web home recently added cards and mobile home recently changed cards must also show assigned tags as compact chips when space allows.
When a search result matched because of a tag display name or tag key, web and mobile clients must present that match as `Tag` in result metadata instead of exposing raw transport field names such as `tag_display_name` or `tag_key`.

Tag chips must:

- Show the display name.
- Use the tag color as the chip's visible color treatment when a color is present, not only as a tiny swatch. The chip must remain readable and accessible against user-provided colors.
- Keep the asset title, photo, kind, parent/location, checkout state, and lifecycle state visually higher priority than tags.
- Collapse gracefully on narrow screens without causing row height jumps or text overlap.
- In compact list and card contexts, clients may show the first few assigned tags plus a `+N` overflow chip instead of rendering every tag.
- In contexts where the tag chip is not nested inside another interactive row or control, clicking or tapping a tag chip must search or browse the current inventory for that tag without using the tag label as replacement search text when the tag ID is known.

Asset create and edit flows must let users:

- Select existing active tags.
- Remove assigned tags.
- Create a new tag inline with display name and optional color where the UI already supports editing asset metadata.
- Choose the optional tag color with a color-picker affordance on clients that support native color input; text entry may remain as a fallback for platforms without a native picker.

Web tag selection lists and web or mobile tag filter option lists must sort active tags alphabetically by display name using locale-aware, case-insensitive collation. Assigned tag chips may preserve the order supplied by the asset when that order is used for compact overflow or otherwise communicates content order.

Tag selectors with more than twelve available tags must use progressive disclosure: show the first twelve naturally sorted options initially, preserve the selected-tag summary, and provide an explicit control to show or hide the complete list.

Web and mobile clients must load active inventory tags through client adapter boundaries, map API tag DTOs into client domain models, and submit complete `tagIds` lists on asset create and update. Clients must not treat generated API DTOs as UI domain models.

Clients must reconcile pending inline tag drafts against known inventory tags by normalized key before calling the tag creation API. If a matching active tag is already known locally, the save must reuse that tag ID instead of issuing a duplicate create request.

Before creating inline tag drafts, clients must trim draft names, discard empty or keyless drafts, normalize valid colors to uppercase `#RRGGBB`, and omit invalid stale draft colors rather than sending a tag create request that the API will reject.

The first inline creation behavior may create the tag immediately before saving the asset draft. If asset save then fails, the created tag may remain available in the inventory; the UI must keep that state visible by refreshing the active tag list.

Tag controls must remain secondary to the asset title, kind, parent/location, photo, and checkout state.

Mobile Browse must keep the first viewport focused on inventory content rather than summoning the keyboard. The search field placeholder or adjacent affordance must make clear that tags are searched alongside asset and location text. The primary `All`, `Places`, `Containers`, and `Items` scope control must remain visible outside secondary Filters. Lifecycle, availability, and tag browse controls must be disclosed through the compact Filters control; Sort must use its own control. Tag browse suggestions must be sorted alphabetically by display label so the filter sheet is predictable.
Web and mobile tag browse controls must behave as multi-select filters over the current result set. Selecting or clearing a tag must not change the text in the search field. Selected tags must compose with the current text query and other filters, and more than one tag may be selected at once.
Web tag browse filters must use durable route state with repeatable tag identifiers so refresh, back navigation, and shared links preserve the selected tag filter set without replacing the text query.
When mobile opens search from a known tag chip with selected tag IDs, the search text input must not auto-focus. Tag-driven navigation is a browse/filter entry point, not a text-entry entry point.
Mobile filter controls must use consistent titled groups and shared option controls. The secondary filter groups are `Tags`, `Status`, and `Availability`; option copy must use consistent noun or adjective labels such as `Active`, `Archived`, `Any`, `Checked out`, and `Available`. Applied tags must remain visible by display name as removable tokens when the sheet is closed. Filter selection must use accessible selected semantics and a non-color state indicator.

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
