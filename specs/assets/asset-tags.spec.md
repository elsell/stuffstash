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

Cross-platform tag-management navigation and interaction behavior are defined by `specs/platform/client-settings-management.spec.md`. That client spec does not add tenant tags, tag inheritance, restore, or hard delete.

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
- Choose the optional tag color through the browser-native color input on web, the native SwiftUI color picker supplied by the pinned Expo UI integration on iOS, and an accessible project-owned full-spectrum picker on Android. The Android picker must reach arbitrary valid `#RRGGBB` colors and include labeled hex entry as a fallback because the pinned Expo SDK has no equivalent native Android picker. Mobile must not reduce the full picker to a fixed set of swatches.
- Use optional labeled quick swatches without losing access to the full picker, and clear an existing color through an explicit accessible action. Hex text entry may be offered as an advanced web fallback but must not replace the browser-native input.

Tag-management collection rows on web and mobile must reserve a leading circular color slot for every tag. Colored tags use a filled circle; tags without a color use an empty outlined circle with sufficient contrast. The reserved slot keeps names aligned across mixed-color lists. The tag name must be the row's only visible text; the row must not repeat the visible indicator with a `No color`, color-name, or hex-value subtitle. The circle or row must expose a text accessibility value such as `Blue color` or `No color`, and tag identity, selection, and state must never rely on color alone.

The iOS bridge must use pinned `@expo/ui 55.0.17`, justified in `specs/platform/tooling-versions.spec.md`. Review must cover Expo SDK compatibility, iOS behavior, accessibility, maintenance posture, and supply-chain risk. The client must not claim that this package supplies a native Android color picker.

Web tag selection lists and web or mobile tag filter option lists must sort active tags alphabetically by display name using locale-aware, case-insensitive collation. Assigned tag chips may preserve the order supplied by the asset when that order is used for compact overflow or otherwise communicates content order.

In mobile Edit, collapsed options include the first twelve naturally sorted active tags plus any selected active tags outside that subset. Pending new definitions remain visible. Native Show all tags / Show fewer tags commands expand or collapse without changing selected IDs, pending definitions, or unrelated draft fields.

Tag selectors with more than twelve available tags must use progressive disclosure: show the first twelve naturally sorted options initially, preserve the selected-tag summary, and provide an explicit control to show or hide the complete list.

Web and mobile clients must load active inventory tags through client adapter boundaries, map API tag DTOs into client domain models, and submit complete `tagIds` lists on asset create and update. Clients must not treat generated API DTOs as UI domain models.

Clients must reconcile pending inline tag drafts against known inventory tags by normalized key before calling the tag creation API. If a matching active tag is already known locally, the save must reuse that tag ID instead of issuing a duplicate create request.

Before creating inline tag drafts, clients must trim draft names, discard empty or keyless drafts, normalize valid colors to uppercase `#RRGGBB`, and omit invalid stale draft colors rather than sending a tag create request that the API will reject.

The first inline creation behavior may create the tag immediately before saving the asset draft. If asset save then fails, the created tag may remain available in the inventory; the UI must keep that state visible by refreshing the active tag list.

Tag controls must remain secondary to the asset title, kind, parent/location, photo, and checkout state.

Mobile Browse must keep the first viewport focused on inventory content rather than summoning the keyboard. The search field placeholder or adjacent affordance must make clear that tags are searched alongside asset and location text. Asset Type (`All`, `Places`, `Containers`, and `Items`), lifecycle, availability, and tag browse controls must be disclosed through the compact Filters control; Sort must use its own control. Type must participate in the same draft, apply, reset, cancellation, count, removable-token, and durable-route model as the other applied filters. Tag browse suggestions must be sorted alphabetically by display label so the filter sheet is predictable.
Web and mobile tag browse controls must behave as multi-select filters over the current result set. Selecting or clearing a tag must not change the text in the search field. Selected tags must compose with the current text query and other filters, and more than one tag may be selected at once.
Web tag browse filters must use durable route state with repeatable tag identifiers so refresh, back navigation, and shared links preserve the selected tag filter set without replacing the text query.
When mobile opens search from a known tag chip with selected tag IDs, the search text input must not auto-focus. Tag-driven navigation is a browse/filter entry point, not a text-entry entry point.
Mobile filter controls must use consistent titled groups and shared option controls. The filter groups are `Type`, `Status`, `Availability`, and `Tags`; option copy must use consistent noun or adjective labels such as `All`, `Items`, `Active`, `Archived`, `Any`, `Checked out`, and `Available`. Applied tags must remain visible by display name as removable tokens when the sheet is closed. Filter selection must use accessible selected semantics and a non-color state indicator.

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
- Web and mobile tag-manager row rendering with mixed colored and uncolored tags, including the invariant leading-slot alignment, outlined no-color state, non-color accessible value, contrast, and large-text behavior.
- Web tag create/edit using native color-input semantics for arbitrary selection and explicit clearing, including keyboard and screen-reader behavior.
- iOS tag create/edit using the native SwiftUI picker and Android tag create/edit using the accessible full-spectrum project picker with labeled hex fallback, both covering arbitrary selection, explicit clearing, and VoiceOver/TalkBack behavior; fixed quick swatches may be tested only as an additional shortcut.
- Visual verification of mixed colored and uncolored manager rows and the open picker in light, dark, increased-contrast, narrow/reflow, and large accessibility text presentations where the platform supports them.


### Native mobile color editing audit correction

When the native iOS color picker is available, expose it directly beside the
optional quick swatches. Its selection updates the parent tag draft; the parent
editor owns Save and cancellation. Do not require opening a custom intermediate
color panel or confirming a second project-owned Done/Cancel pair. Clearing
remains explicit and opening the picker must not replace an absent color.
Android and runtimes without the native picker retain the documented accessible
spectrum/hex fallback with its local draft. Disabled editors reject late changes.

The mobile Edit tag picker applies inline tag resolution as one draft update:
selected existing tag IDs and pending new-tag definitions change together. A
normalized match to an existing tag selects that tag without creating another;
updating pending definitions must not restore the previous selected IDs. Preserve
all previously selected tags and unrelated draft fields.

Native Edit tag-disclosure acceptance must use isolated inventory choices exceeding
twelve, preserve an initially selected option outside the first twelve, select
another hidden option, collapse and confirm both selected states remain exposed.
At the largest accessibility text size, reveal disclosure and completion controls
fully within the visible form. Discarding the test draft must return without any
production mutation. Capture expanded and collapsed states; this supplements the
route-level Save retention regression.

Mobile Edit must retain the inline tag name and color in its route-owned draft
until Add tag stages them or the user clears them. Nonblank unstaged input counts
as unsaved work for Cancel/discard. Save must not silently omit that input: keep
Save unavailable and explain beside the entry that the tag must be added or the
entry cleared. Add tag remains the explicit staging action and atomically clears
its entry while retaining unrelated edits. Whitespace-only entry is not dirty.

Native unfinished-entry acceptance starts with an otherwise unchanged Edit draft:
type an inline tag name, dismiss the keyboard, Cancel and Keep editing, then
verify the name remains. Save stays disabled with the Add/clear explanation
visible. Add tag clears the entry and enables saving the staged change. The
runner must verify keyboard readiness and exact input, and discard rather than
write to any real inventory.
