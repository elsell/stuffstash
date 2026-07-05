# Web Inventory Workspace Spec

## Purpose

Stuff Stash needs a production-grade web inventory workspace direction before the disposable tracer-bullet screens are replaced.

This spec captures the approved web UI direction from the Stuff Stash UI design workshop candidate. It defines the first real product shell, navigation model, browse flow, add flow, search placement, tenant and inventory context behavior, and asset/location surfaces for the SvelteKit web app.

## Scope

This spec covers:

- Authenticated web application shell behavior.
- Desktop side navigation and top header behavior.
- Mobile header and bottom navigation behavior.
- Tenant and inventory context switching.
- Inventory home workspace.
- Location browsing and location-contained asset list behavior.
- Asset detail view behavior.
- URL-addressable workspace routes and deep links.
- Add item/container/location interaction.
- Photo attachment affordances during creation.
- Search placement and high-level search expectations.
- Accessibility, responsive, and frontend engineering expectations for these surfaces.

This spec does not define:

- The public marketing site.
- The mobile native application UI.
- Final visual design tokens beyond the existing brand guidance.
- Complete search result ranking.
- Complete edit, move, archive, restore, delete, undo, sharing, audit, custom field, or custom asset type screens.
- Backend API behavior that is already owned by domain and platform API specs.
- Direct upload implementation details, which are owned by `specs/media/media-attachments.spec.md`.

## Source Context

This spec is grounded in:

- `specs/platform/ui-design-workshop.spec.md`.
- `specs/platform/brand-guidelines.spec.md`.
- `specs/platform/client-technology.spec.md`.
- `specs/platform/web-frontend-tracer-bullet.spec.md`.
- `specs/assets/asset-model.spec.md`.
- `specs/assets/containment-model.spec.md`.
- `specs/locations/location-model.spec.md`.
- `specs/media/media-attachments.spec.md`.
- `specs/search/search.spec.md`.
- `specs/identity-access/tenant-inventory-access.spec.md`.

The approved direction was explored in a temporary SvelteKit candidate at `/private/tmp/stuffstash-ui-candidate-home-hub`.

## Product Principles

- The web workspace must feel like a sharp personal tool for a technically adept homeowner, not warehouse software, an admin console, a marketing page, or a novelty AI surface.
- The first screen after sign-in must be the usable product surface.
- The primary day-to-day jobs are finding assets, browsing locations, and adding assets quickly.
- The interface must not use explanatory product-pitch panels or third-person UI narration. State and actions should be apparent through layout, labels, hierarchy, and affordances.
- Photos are a primary recognition layer. The UI should frame user inventory photos clearly and avoid competing decorative visuals.
- A missing photo must render as an explicit kind fallback. The web app must never reuse, inherit, or visually carry over another asset's photo for an asset, container, or location that does not have its own primary photo.
- Tenant and inventory context must always be knowable because tenant is the top-level security boundary, but the UI must not overemphasize technical tenancy language during ordinary work.
- The home workspace must stay focused. Sharing, audit history, and recent activity are not primary home panels.

## Domain Language

The UI must use user-facing household language while preserving domain correctness:

- `Tenant` is the top-level security boundary and may appear in context switchers and settings where needed.
- `Inventory` is the primary collection boundary inside a tenant.
- `Location` is user-facing language for place-like assets backed by `asset.kind = location`.
- `Container` is a movable asset that can contain other assets.
- `Item` is a normal thing the user wants to find, move, document, or share.
- `Asset` may be used in details, settings, and technical-adjacent surfaces when it refers to item/container/location records as a group.

The UI must not expose persistence, SpiceDB, OIDC, generated DTOs, storage keys, bucket names, audit internals, or implementation-specific tenancy phrasing.

## Web Shell

The authenticated web app must use a persistent product shell.

## URL And Deep-Link Model

Every durable web workspace destination must be addressable through a stable URL path, not only in component-local state.

The first canonical URL model is:

- `/` for the selected inventory home.
- `/tenants/{tenantId}/inventories/{inventoryId}` for an inventory home.
- `/tenants/{tenantId}/inventories/{inventoryId}/locations` for top-level location browsing.
- `/tenants/{tenantId}/inventories/{inventoryId}/locations/{locationAssetId}` for a focused location view.
- `/tenants/{tenantId}/inventories/{inventoryId}/locations/{locationAssetId}/edit` for the location edit state when edit is available.
- `/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}` for asset detail.
- `/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/edit` for the asset edit state when edit is available.
- `/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/move` for the asset move state when move is available.
- `/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/archive` for active asset archive confirmation when archive is available.
- `/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/restore` for archived asset restore confirmation when restore is available.
- `/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/delete` for the asset delete confirmation state when delete is available.
- `/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/delete` for the attachment delete confirmation state when delete is available.
- `/tenants/{tenantId}/inventories/{inventoryId}/search` for search.
- `/tenants/{tenantId}/inventories/{inventoryId}/settings` for inventory settings.
- `/tenants/{tenantId}/inventories/{inventoryId}/settings/{section}` for a focused inventory settings section.
- `/tenants/{tenantId}/inventories/{inventoryId}/settings/access/invitations/{invitationId}/expire` for an invitation expire confirmation when expire is available.
- `/tenants/{tenantId}/inventories/{inventoryId}/settings/access/invitations/{invitationId}/cancel` for an invitation cancel confirmation when cancel is available.
- `/tenants/{tenantId}/inventories/{inventoryId}/settings/access/invitations/{invitationId}/delete` for an invitation delete confirmation when delete is available.
- `/tenants/{tenantId}/inventories/{inventoryId}/settings/fields/asset-types/{customAssetTypeId}/archive` for a custom asset type archive confirmation when archive is available.
- `/tenants/{tenantId}/inventories/{inventoryId}/settings/fields/field-definitions/{customFieldDefinitionId}/archive` for a custom field definition archive confirmation when archive is available.
- `/tenants/{tenantId}/inventories/{inventoryId}/import` for import.
- `/tenants/{tenantId}/inventories/{inventoryId}/import/{source}` for a focused import source, initially `legacy-homebox` or `legacy-homebox-csv`.
- `/tenants/{tenantId}/inventories/{inventoryId}/add/{kind}` for add item, container, or location.

The web app may accept `/inventories/{inventoryId}` and descendant paths as compatibility aliases for an inventory that is visible in the current tenant context. When a compatibility alias can be resolved, the app should replace the browser URL with the canonical tenant-scoped path.

The route path owns durable navigation state. Query parameters may own transient filters such as:

- `lifecycle=active|archived`.
- `q={search query}`.
- `mode=fuzzy|exact`.
- `parent={assetId}` for add item, container, or location routes that should open with an existing location or container selected as the parent destination.
- `invitationStatus=all|pending|accepted|revoked|cancelled|expired` for the access settings invitation filter.
- `auditScope=inventory|tenant` for the activity settings audit scope filter.

Deep links must preserve tenant and inventory boundaries:

- If the requested tenant is visible to the principal, the app should select it before rendering the route state.
- If the requested inventory is visible in the selected tenant context, the app should select it before rendering the route state.
- If the requested tenant is not visible to the principal, the app must show a calm unavailable or setup state rather than rendering stale local data.
- If the requested inventory is not visible in the current tenant context, the app must show a calm unavailable or setup state rather than rendering stale local data.
- Unavailable-route recovery controls that return to the selected inventory home must expose a canonical `href` while preserving ordinary in-app navigation.
- A location deep link must only open an asset whose kind is `location`.
- A location edit deep link must normalize to the same API-backed asset detail edit workflow used for editing the underlying location asset.
- An asset deep link must load the selected asset through the repository port and API adapter.
- Asset action deep links must not leave the URL in an action state when the action is unavailable. The app must normalize to asset detail or show an unavailable state.
- Asset action panels that materially change data, such as edit, move, archive confirmation, restore confirmation, and delete confirmation, are durable route states.
- Route-opened asset action panels must appear before secondary detail, file, and danger-zone sections and scroll into view so their heading and controls are visible when focus moves to the panel.
- Settings access actions that materially change invitations, such as expire, cancel, and delete, must use durable confirmation route states instead of immediate row-button mutations.
- Settings actions that materially change reusable schema, such as archiving custom asset types or custom field definitions, must use durable confirmation route states instead of immediate icon-button mutations.
- Unsupported paths under a valid inventory route must fall back to the inventory home without crashing and normalize the browser URL to that inventory home.

Navigation controls must update the URL when they change durable workspace state, and browser back/forward controls must restore the corresponding workspace state.

Durable navigation items and durable action menu choices must expose canonical `href` values for their route states even when the app intercepts ordinary same-window clicks for client-side navigation.
Asset detail controls that open durable action panels must expose canonical `href` values for their action routes.

Desktop:

- A sticky left side navigation must be visible at large viewport widths.
- A top header must remain available for global search and add actions.
- The side navigation must not include a Search item when search is already globally available in the top header.
- The side navigation must contain durable destinations, not duplicate global actions.
- The profile entry belongs at the bottom of the side navigation.

Mobile:

- The desktop side navigation must collapse away.
- The top header must be compact and must not contain the global search bar or the add button.
- Mobile must use bottom navigation for primary reachable actions.
- The bottom navigation must include Search and a central Add action.
- The central Add action must open the same add tray behavior as desktop.
- The Places or Locations bottom-navigation destination must route to the top-level locations browse route, not merely duplicate Home.

## Tenant And Inventory Context

The product shell must show the current inventory and tenant context without making tenancy the main content.

Desktop:

- The side navigation must include a compact context switcher near the top.
- The switcher must occupy a single sidebar row when closed.
- The switcher trigger row must show both the current inventory name and tenant name.
- Opening the switcher should show a popover with the inventories inside the current tenant.
- The popover must include a right-aligned `Switch Tenant` action.
- `Switch Tenant` must show a tenant list.
- Selecting a tenant must keep the popover open and replace the tenant list with that tenant's inventories.
- The switcher must not show one combined dropdown containing all inventories from all tenants.
- The switcher must not include a tenant/inventory search field in the approved first direction.
- The switcher must not include a separate duplicated current tenant card, persistent inventory list, or duplicate inventory settings link.
- Inventory choices must expose selected state accessibly and show relationship metadata when available.
- Inventory choices must expose canonical inventory home `href` values while preserving ordinary in-app switching behavior.
- Tenant choices must expose selected state accessibly and show the number of inventories in that tenant.

Mobile:

- The compact header context control should open a bottom sheet or equivalent mobile-appropriate context switcher.
- The mobile context switcher should follow the same tenant-first behavior as desktop.
- The mobile context trigger should remain a single row; the sheet may use section labels, selected check affordance, identity labels, metadata, and optional role pill.
- The mobile context switcher must not require a search field for the approved first direction.
- Mobile context switcher backdrop controls must expose a clear accessible close name and must not inherit ordinary button chrome.
- The open mobile context switcher sheet must render above the backdrop and bottom navigation so the inventory choices are visually and pointer-accessible.
- The open mobile context switcher sheet must make the route content and mobile bottom navigation inert and hidden from assistive technology while leaving the sheet itself available.

## Desktop Header

The desktop top header must prioritize:

- Global inventory search.
- Add action.
- User/account affordance only when not already clear from the side navigation.

Search:

- Search must be front and center on desktop.
- Search must be available across primary web pages.
- Search should feel closer to Google Drive than a command palette: a visible field that accepts ordinary asset/location/container terms.
- The dedicated search route must preserve the same autocomplete affordance as the global header search, including keyboard access to suggestions and direct opening of suggested assets.
- On the dedicated desktop search route, the page search field is the primary search affordance and the desktop header must not also render the global search field.
- Search must be scoped to the selected tenant and inventory unless a future search spec defines cross-inventory behavior.
- Search must preserve tenant and inventory authorization boundaries.

Add:

- The Add action must live in the desktop header.
- The Add action should use a compact menu pattern similar to GitHub's create button.
- The user must be able to choose `Item`, `Container`, or `Location`.
- The add dialog/tray must still allow changing the selected kind after opening.
- Visible add-tray dismissal controls must expose canonical `href` values back to the workspace route that ordinary in-app close restores.
- Add must be disabled or replaced by an explicit denied state for inventories where the user lacks create-asset permission.
- Header and mobile Add controls must expose a perceivable disabled reason when add creation is unavailable or no inventory is selected.
- Add deep links must not silently render the ordinary workspace when creation is unavailable. They must show a calm denied state or normalize to a non-action route.
- Modal add and edit surfaces must make the background workspace inert and hidden from assistive technology while the modal is open.
- On mobile, the add tray must behave like a focused sheet with usable viewport height and fixed completion controls so save/cancel actions remain reachable while long custom-field or parent-picker content scrolls.
- Parent-picker destinations in mobile add and edit trays must scroll with enough bottom clearance that focused or selected destination controls are not hidden under sticky action bars.

## Mobile Navigation

Mobile bottom navigation must provide reachable primary actions without duplicating the desktop header.

The approved first mobile bottom navigation direction is:

- Home.
- Search.
- Add as the central primary action.
- Locations or equivalent browse destination when a full route exists.
- Settings or inventory/settings access when it exists.
- The mobile Add control must expose the same durable add-action URL and unavailable-state semantics as the desktop Add control.

Mobile must not show a desktop-style global search bar in the header when Search is already in bottom navigation.

## Inventory Home Workspace

The inventory home workspace must stay focused on one or two primary concerns:

- Browse top-level locations.
- Recently added assets, when useful and low-clutter.

The home workspace must not include primary panels for:

- Sharing.
- Recent activity/audit feed.
- Technical tenant details.
- Product explanation or feature narration.
- "Needs attention" in the first approved direction.

Sharing and activity belong in inventory settings, asset detail, or future focused pages unless a later spec gives them a stronger home-workspace role.

Top-level location browsing:

- Locations should be the main browse surface.
- A top-level locations browse route must exist for navigation surfaces that need to prioritize places separately from the home workspace.
- Location cards or tiles should use photos when available.
- Location cards must open a focused location view.
- The UI must support long location names, missing photos, and empty inventories.

Recently added:

- Recently added assets should appear before location browsing as a compact horizontal rail when active assets exist.
- Recently added must not dominate the page or compete with search/add.
- Recently added rows must open the asset detail view.
- Recently added rows, archived asset rows, location tiles, and add-location actions must expose canonical `href` values while preserving ordinary in-app navigation behavior.
- Home add-location controls, including empty-inventory calls to action, must respect the selected inventory's create permission. When creation is unavailable they must render as explicit disabled or denied states instead of live deep links.
- Empty-inventory home states must not imply that item creation is blocked when the inventory can accept root-level items. They should recommend creating locations for better browsing while still exposing a route-backed add-item action when creation is allowed.

## Location View

Clicking a top-level location must open a focused location view.

The first location view must include:

- Back navigation to the locations/home view.
- Location title.
- Location photo when available.
- Location description.
- Asset count for the assets visible in that location scope.
- A scannable list of assets inside the location.

The asset list must:

- Support items and containers.
- Show asset photo or a kind icon fallback.
- Show title.
- Show optional custom asset type label when available.
- Show a short description.
- Show the local containment trail within the location.
- Open the asset detail view when an asset row is selected.
- Expose canonical `href` values for the location back destination, current location edit destination, nested location rows, and contained asset rows.

The location view must not become a dashboard. It should answer: "What is in this place?"

## Asset Detail View

Selecting an asset must open an asset detail view.

The first asset detail view must include:

- Back navigation to the previous location list when opened from a location.
- Back navigation must expose a canonical `href` matching the ordinary in-app back destination.
- Asset action panel cancellation must expose a canonical `href` back to the asset detail or focused location route that the ordinary in-app close action restores.
- A prominent asset photo area with a kind icon fallback.
- Asset title.
- Asset kind.
- Optional custom asset type label.
- Description.
- Location trail.
- Lifecycle state.
- Updated timestamp or equivalent saved-state clue.
- Primary actions for `Edit` and `Move` as visible affordances.
- A secondary actions affordance for future actions.

Desktop asset detail layout:

- The primary photo/gallery must sit on the left or top-left at roughly 320-420px wide when viewport space allows.
- The asset title, location trail, lifecycle state, kind, custom type, and primary actions must sit beside the photo/gallery.
- Description and custom fields belong below the hero identity area.
- Photos must be presented near the hero as asset identity media, with a thumbnail rail when more than one image is available.
- Disabled asset-detail photo upload actions must expose a perceivable reason, including missing edit access, inactive lifecycle state, save-in-progress state, or no supported image upload types.
- The first implementation may choose the first active image attachment as the primary photo until explicit primary-photo selection is specified.
- Non-image attachments such as receipts, manuals, PDFs, and supporting documents must be visually separated from photos and appear lower than the primary photo/gallery.

Mobile asset detail layout:

- The primary photo/gallery must appear first.
- Asset title and location trail must appear immediately after the photo/gallery so the user can confirm identity without a long scroll.
- Photo upload must expose one visible mobile affordance near the primary asset actions; a gallery-local upload action may remain on desktop but must not create duplicate adjacent mobile `Add photo` controls.
- `Edit`, `Move`, and photo upload actions must remain close to the title and identity area.
- The photo thumbnail rail must appear before non-image documents.
- The photo area must not consume so much vertical space that the title and location are hidden for common mobile viewport heights.

The detail view is the preferred home for asset-level actions such as edit, move, archive, sharing-related actions, attachment management, and future custom field editing.

Asset detail must support:

- Missing photos.
- Long titles.
- Long location trails.
- Viewer or denied states for edit-only actions.
- Archived state when lifecycle views expose archived assets.

Asset detail loading and actions must use real API-backed boundaries:

- Opening an asset detail must load the selected asset by ID through the frontend repository port and API adapter rather than relying only on the current list row.
- Asset detail API responses must include the same safe primary photo summary used by asset list and search responses when the asset has an active image attachment.
- Frontend API adapters must treat primary photo summaries as belonging only to the exact tenant, inventory, and asset response that carried them. If a response does not include a primary photo summary for an item, container, or location, the mapped frontend asset must not include a photo.
- The API adapter must call the generated client wrapper for `GET /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}` and map the DTO into the frontend asset domain model.
- Editing asset title and description must go through a repository update method backed by the generated client wrapper.
- Moving an asset must update `parentAssetId` through the same API-backed update path and must use valid parent targets from the current inventory, not free-form IDs.
- Edit and move affordances must require the exact `edit_asset` permission from the selected inventory access metadata.
- Save success, save failure, loading, and denied states must be explicit in the detail workflow.
- Svelte components must not import generated SDK DTOs or call generated client methods directly for detail, edit, or move behavior.

## Add Flow

The add workflow must have equal product weight with find/browse.

Entry points:

- Desktop: Add action in the top header.
- Mobile: central Add action in the bottom navigation.

The add surface:

- Must open as a modal, tray, or sheet appropriate to viewport.
- Must let the user choose or change `Item`, `Container`, or `Location`.
- Must reflect the selected kind in the tray heading, title/name prompt, placeholder, and primary save action so add-location and add-container routes do not read like item-only forms.
- The kind selector must be compact. It must not use large stacked cards.
- The tray must show a compact live summary of the selected kind, parent destination, and photo count so users can confirm what will be created without rescanning the whole form.
- On mobile, the live summary must remain a single compact row so it does not displace the primary create fields or completion controls. The destination value must receive more horizontal priority than the kind and photo-count values, and full summary values must remain available to assistive technology when visible text is elided.
- Must collect name/title.
- Must collect a valid parent target when required.
- Saving a new location must land on the canonical focused location route for the created location, `/tenants/{tenantId}/inventories/{inventoryId}/locations/{locationAssetId}`, rather than the generic asset detail route.
- Add routes may include a `parent` query parameter to preselect an existing valid location or container parent. Location view add actions must use this route-backed preselection rather than component-local-only state.
- If an add route includes an invalid `parent` query parameter, the app must normalize the URL to the same add route without the invalid parent rather than silently saving to a different destination than the URL implies.
- Parent target selection must use a picker/search over valid location/container targets, not a free text field that implies invalid foreign keys can be saved.
- Parent target selection must be search-first when many locations or containers are available: it must always show the current destination, expose an inventory root action only when the current destination is not already root, and avoid rendering an unfiltered stack of every possible parent by default.
- The parent picker must support quick creation when the user realizes the parent location/container does not exist yet.
- Quick parent creation in add flows must be an explicit opt-in section, not always-visible secondary fields that compete with the common add path.
- Quick creation must be explicit and must preserve authorization, validation, and audit expectations when implemented against the real API.

Photos:

- Photos are first-class and low-friction in the add flow.
- Photos must be optional. A user must be able to save an asset without adding a photo.
- The add surface should expose camera and upload actions.
- Photo actions must be grouped as one attachment section with accessible group labeling, supported media guidance, and selected-photo count/status.
- The add photo section must associate supported-media guidance, selected-photo status, and validation errors with the attachment group so assistive technology users hear the current state before acting.
- The first approved web direction supports JPEG, PNG, and WebP image affordances consistent with the media spec.
- Selected photos should show thumbnails and allow removal before save.
- Selected photo previews must be exposed as a named list with individual remove actions.
- Selected photo preview images must expose the selected file name as their accessible image text so users can identify each thumbnail.
- Invalid or oversized selected photos must block save until removed or corrected.
- Attachment size and supported type rules must come from configuration or a media policy boundary rather than scattered hard-coded checks.

Saved state:

- Successful create must show concise saved feedback.
- When quick-creating a parent and asset together, saved feedback must make both outcomes clear.
- Saved feedback for photo uploads must be count-aware and must preserve quick-created parent context when both happen in one save.
- If some photo uploads succeed and others fail, saved feedback must report both the uploaded count and the failed count.
- Future real implementation must produce audit history through application behavior, not UI-only state.

## Search

Search is a primary workflow.

Desktop:

- Search belongs in the top header.
- Search does not belong in the side navigation for the approved first direction.

Mobile:

- Search belongs in bottom navigation or an equivalent mobile primary action surface.
- Search does not need to occupy header space.

Search behavior:

- Search should resolve to authorized assets, containers, and locations in the selected inventory.
- Search should provide autocomplete-style suggestions from visible inventory assets while preserving the repository-backed search action as the authoritative result source.
- Autocomplete suggestions and local/demo search results must rank exact and prefix title matches before looser title, description, or custom-type matches.
- Search suggestions and search results must show an asset image thumbnail when the asset has its own primary photo, and must show the same explicit kind fallback used elsewhere when it does not.
- If the API says an asset has its own primary photo but the web adapter cannot fetch or materialize the authenticated thumbnail, the frontend asset model must preserve that state as an unavailable photo rather than treating the asset as unphotographed. Thumbnail surfaces, including search suggestions and results, must show the explicit kind fallback with a visible unavailable-photo badge and expose that state to assistive technology.
- Assets created with a successfully uploaded photo must appear in subsequent search results with that asset's own primary photo thumbnail.
- Search result rows should open asset or location detail/list surfaces. Location-kind results and suggestions must route to the focused location URL rather than generic asset detail.
- Search suggestions and result rows must expose canonical destination `href` values while preserving ordinary in-app navigation behavior.
- Global header search and the dedicated search page must use a shared suggestion-list composition so thumbnail behavior, kind labels, route links, focus state, and ordinary list/button semantics do not drift.
- No-results and denied states must be explicit and calm. Submitted searches with no results must name the query so the user can tell the search ran. Focused autocomplete fields with a non-empty query and no suggestions must expose calm no-suggestion feedback instead of appearing inert.
- Search must not bypass tenant, inventory, lifecycle, or authorization boundaries.

## Consistent Controls

The workspace must use consistent controls for repeated interaction patterns:

- Lifecycle and search-mode filters must use a segmented tab/filter control rather than unrelated pressed buttons.
- Segmented filters that correspond to durable route query state, such as lifecycle and search mode, must expose canonical `href` values while preserving ordinary in-app filtering behavior.
- Durable navigation must use nav links/buttons with clear current state.
- Transient menus such as the desktop Add menu must expose durable item `href`s, move focus into the menu when opened, close on Escape or focus leaving the menu, and restore focus to their trigger when dismissed.
- Desktop side navigation must group primary inventory destinations separately from utility workflows, expose the current destination with `aria-current="page"`, and avoid presenting secondary workflows as an undifferentiated stack.
- Icon-only controls must have accessible names.
- Creation and edit controls must use the local shadcn-style button, input, select/tabs, label, textarea, and dialog/sheet primitives or product-specific compositions over those primitives.
- Parent, type, and custom-field pickers must have keyboard-reachable controls, screen-reader labels, and visible selected state.
- The add surface parent picker must support filtering valid parent targets and must expose each picker group with honest grouped-control semantics rather than unlabeled piles of buttons.
- Parent target pickers must avoid unfiltered all-parent stacks by default; suggested destinations, search results, current selection, and empty states must be visually distinct.
- Parent target pickers must show the selected destination as a compact summary with target kind and containment trail, expose a clear action when a non-root target is selected, and show result counts without rendering every possible parent before search. When inventory root is already selected, the root summary is the selected-state surface and must not be duplicated as a separate pressed option.
- Parent target pickers must offer a compact bounded set of suggested destinations before search, initially no more than four suggestions, excluding the already-selected destination because it is shown in the current destination summary.
- Parent target search results may start with the same compact visible limit, but overflow states must expose an explicit keyboard-reachable action to reveal the remaining matches without changing the query or hiding the current destination.
- Parent target pickers must prefer locations before containers when choosing unfiltered suggestions and ordered search results.
- Parent target search results must rank exact and prefix title matches before looser title or containment-trail matches within the same target kind.
- Parent target search result counts must be announced politely to assistive technology users when the query changes.
- Parent target search results must present target kind, title, and containment trail in each selectable row so locations and containers are distinguishable while scanning.
- Parent target kind/trail labels must omit empty containment trails and must not render dangling separators.
- Parent target search results must group locations and containers with named grouped-control semantics for assistive technology users.
- Parent target pickers must keep suggested destinations and search results bounded by the component's visible limit.
- Asset detail edit and move panels must use the same grouped-control and searchable parent-target patterns as the add surface.
- Add and move parent-target selection must be implemented through a shared workspace picker component so filtering, selected-state language, empty states, and grouped-control semantics do not drift between creation and asset-detail actions.
- Settings relationship selectors, status filters, and audit scope filters must use the shared segmented-control composition rather than one-off pressed-button groups.
- Settings status and scope filters that correspond to durable settings subsection state, such as access invitation status and activity audit scope, must expose canonical `href` values while preserving ordinary in-app filtering behavior.
- Settings access invitation rows must keep identity, relationship/status metadata, status badge, and row actions visually distinct at desktop and mobile widths instead of compressing them into a crowded single line.
- Route-backed segmented-control options must expose link semantics with canonical `href` values, `aria-current` for the selected option, and the same visible selected state as button-backed options.
- Custom field target pickers must expose visible selected state, `aria-pressed` state, and a calm empty state when no custom asset types are eligible.
- Custom asset type, enum, and custom field target choice grids must use a shared workspace choice-grid composition so selected-state semantics, disabled behavior, empty state copy, and button styling do not drift between add, edit, and settings surfaces.
- Import options for images, insecure TLS, and private-network access must use the shared binary-option composition with visible on/off state, clear option copy, and honest switch or checkbox semantics.
- Import source choices that correspond to distinct import workflows must expose canonical `href` values while preserving ordinary in-app source switching behavior.
- Route-backed controls rendered as disabled links must remove their `href`, expose `aria-disabled`, leave the tab order, and visually match native disabled buttons through the shared button primitive.
- Form errors, denied actions, loading states, and saved feedback must be perceivable to assistive technologies.
- Passive saved/status feedback must not intercept pointer interaction with dialogs, sheets, or workspace controls.
- Disabled primary actions in workspace forms must have a visibly unavailable treatment rather than relying only on opacity over the active primary color.

## Inventory Settings

Inventory settings is the preferred location for inventory-level secondary workflows, including:

- Sharing and access management.
- Inventory metadata.
- Activity/audit views when exposed.
- Tenant/inventory administrative details.

These workflows must not be primary home panels in the approved first direction.

Inventory settings must be structured as focused sections rather than one long mixed surface:

- `overview` for inventory and tenant summary.
- `access` for sharing and access management.
- `fields` for custom asset types and custom fields.
- `activity` for audit/history when exposed.
- `administration` for tenant or inventory administrative actions and denied states.
- The settings section navigator must behave like navigation, not a generic filter bar: each section control must expose a canonical `href`, current section state, icon, title, and short description.
- The settings section navigator must remain compact and scannable on desktop, and collapse into a compact mobile pattern on narrow screens that exposes all available sections without clipping labels or consuming the first viewport before the active settings task.
- Settings surfaces must collapse before controls, panels, or invitation lists force horizontal page overflow at tablet and narrow desktop widths.
- The settings content area should restate the active section with a concise heading and context so the user can confirm where they are after deep linking.
- The settings page shell must avoid duplicating inventory and role context that is already present in the workspace chrome or the active settings panel. The top heading should name the task surface, while detailed inventory, tenant, and relationship values belong in the overview or focused settings content.

Settings section navigation must be URL-addressable through `/settings/{section}`. Unknown settings sections must resolve to `overview` and normalize to the canonical `/settings` overview URL rather than leaving an unsupported section slug in the browser.

Settings section navigation must use route-backed links with `aria-current` for the active section rather than pressed-button filter semantics. On desktop, it may render as a compact vertical section rail with short descriptions when that improves scanability. On mobile, the same sections may collapse into a compact wrapping grid or horizontal strip above the active section, as long as all sections remain discoverable, section labels remain visible, and section descriptions remain available to assistive technology without forcing a tall card grid.

## Reusable Workspace Controls

Repeated segmented controls must be implemented through a reusable workspace control component rather than repeated ad hoc markup. The first reusable control must support:

- A typed list of options.
- A visible selected state.
- `role="group"` with a clear accessible label.
- `aria-pressed` on each button-backed option.
- Link semantics, `aria-current`, and visible selected state on each route-backed option.
- Disabled options where needed.
- Consistent wrapping, spacing, focus, and mobile behavior.

Search suggestions must use honest ordinary button/list semantics unless a future pass implements a complete combobox pattern. Partial combobox/listbox/menu ARIA is not allowed.

## Responsive Behavior

Desktop:

- Use horizontal space for a persistent side nav, a stable header, denser lists, and detail layouts.
- Avoid filling wide screens with low-value stacked boxes.
- Keep the side nav sticky.

Mobile:

- Use one-column layouts for location and asset views.
- Keep bottom navigation visible and reachable.
- Keep add/search reachable without making the header tall.
- Fixed bottom chrome such as mobile navigation, saved/status toasts, and local-auth/demo banners must not occlude the final reachable content in the main workspace.
- Mobile workspace pages and sheets must share a named bottom-clearance rhythm so long settings, import, add, edit, and picker content can be scrolled fully above fixed bottom navigation, sticky action rows, and safe-area insets.
- Mobile workspace navigation and primary interaction controls should provide at least a 44px touch target wherever the control is not intentionally inline text.
- Desktop and mobile add trays must reserve visible space for the tray heading and save/cancel row while the form body scrolls internally.
- Mobile add-tray content must scroll in a body region that ends above its action row. The save/cancel action row must remain reachable without covering parent picker, quick-parent, custom field, or photo controls.
- Use sheets/trays for transient creation and context switching.
- Ensure bottom navigation does not cover primary form actions; account for safe area insets.

## Accessibility

The web workspace must meet WCAG 2.2-aligned expectations:

- Keyboard navigation for nav, context switchers, search, add menus, modals, sheets, lists, and detail actions.
- Focus trapping and Escape close behavior for dialogs and sheets.
- Visible focus states.
- Proper labels for icon buttons.
- Semantic landmarks for navigation and main content.
- Dialogs must have accessible names.
- Asset rows and location cards must have clear accessible names.
- Images used as decoration may have empty alt text; meaningful image content must be represented by adjacent text or appropriate alt text.
- Touch targets must be large enough on mobile.
- Text must not overflow buttons, cards, nav items, list rows, or trays.
- Permission denied, validation error, empty, and no-results states must be screen-reader perceivable.

## Frontend Architecture

The implementation must preserve the frontend architecture defined by `specs/platform/client-technology.spec.md`.

Required boundaries:

- Svelte route files own route-level loading, page composition, and route-local state.
- Product components own focused workflow surfaces such as shell, context switcher, home browse, location list, asset detail, add tray, search panel, and settings panels.
- Reusable generic controls must use the local shadcn-style component foundation.
- Generated OpenAPI DTOs must stay behind frontend API adapters.
- Frontend domain models must describe product concepts used by screens.
- Runtime configuration must own API/auth/provider URLs and feature flags.
- Observability must use domain-oriented helper/port events, not raw `console.log`, `print`, or one-off diagnostics.
- Meaningful UI states must use typed concepts or enums, not loose strings.
- Multi-step asset workflow transitions, such as create-with-quick-parent, create-with-photo-upload, and local workspace asset replacement, must live in focused application helpers rather than accumulating in the product shell component.
- Files must stay cohesive. Do not create broad "god" files that mix route state, mock data, adapter mapping, observability, UI primitives, and product components.

The first promoted implementation should split at least these concerns:

- App shell.
- Desktop side navigation.
- Desktop top header.
- Mobile header.
- Mobile bottom navigation.
- Tenant/inventory switcher.
- Inventory home.
- Location browser.
- Location asset list.
- Asset detail.
- Add asset flow.
- Parent target picker.
- Photo attachment picker.
- Search surface.
- Inventory settings.
- Frontend domain types.
- API adapter mapping.
- Observability events.

## First Promotion Implementation Shape

The first promoted `apps/web` implementation must replace the disposable tracer-bullet page composition with production-shaped frontend boundaries from the start.

Required implementation split:

- Frontend domain models live outside Svelte components and represent product concepts such as tenant, inventory, asset, asset kind, lifecycle state, selected photos, workspace mode, and search result.
- UI-facing data access must go through a frontend repository port with domain-oriented methods.
- The REST adapter must use the checked generated TypeScript API client package and map API DTOs into frontend domain objects at the adapter boundary.
- Svelte components must not import generated schema types or API DTOs directly.
- Route files may compose authentication, runtime configuration, repository construction, and top-level page state, but must not become the place where transport mapping, containment derivation, visual components, and API calls all accumulate.
- Workspace-specific derivation such as top-level locations, contained asset lists, valid parent targets, and containment trails must live in focused application helpers.
- Parent target picker suggestion, search ranking, result limiting, and location/container grouping must live in focused application helpers rather than component-local derivation.
- Parent target picker result-count, destination-count, suggestion-count, no-match, overflow, and no-target presentation must live in focused application helpers rather than component-local conditional copy.
- Parent target destination metadata labels must live in focused application helpers rather than component-local string assembly.
- Add surface control options, kind-specific labels, destination summaries, and photo-count summaries must live in focused application helpers rather than component-local derivation.
- Add form summary labels, section labels, description placeholders, and quick-parent labels must live in focused application helpers rather than component-local conditional copy.
- Add surface quick-parent validation copy must live in focused application helpers rather than component-local conditional copy.
- Add photo-picker action labels, input labels, selected-list labels, supported-image type derivation, and supported-format copy must live in focused application helpers rather than component-local conditional copy.
- Home and location browse href derivation must live in focused application helpers rather than component-local route string assembly.
- Home workspace heading, lifecycle filter options, empty-state copy, and create-location denied copy must live in focused application helpers rather than component-local conditional copy.
- Focused location empty-state copy, add-item action label, and create-item denied copy must live in focused application helpers rather than component-local conditional copy.
- Shell navigation destination grouping, labels, descriptions, current-state rules, context-switcher option metadata and href derivation, add-menu option labels and href derivation, and import-source href derivation must live in focused application helpers rather than component-local route string assembly.
- Context switcher trigger labels, active tenant label, and empty-inventory copy must live in focused application helpers rather than component-local fallback copy.
- Product-shell route recovery, add-close, normalization, and detail-back route derivation must live in focused application helpers rather than accumulating in the workspace shell component.
- Workspace unavailable-route and no-inventory setup presentation must live in focused application helpers rather than component-local conditional copy.
- Search query execution, search state normalization, and autocomplete-style suggestion derivation must live in focused application helpers rather than accumulating in the product shell component.
- Search filter option composition, labels, and href derivation must live in focused application helpers rather than component-local route string assembly.
- Search result and suggestion href derivation must live in focused application helpers rather than component-local route string assembly.
- Search panel loading, first-run, empty-result, and error presentation must live in focused application helpers rather than component-local conditional copy.
- Route-backed control click interception must use a shared helper so ordinary in-app clicks, modified clicks, non-primary clicks, and already-prevented events behave consistently across navigation, filters, suggestions, settings actions, and dialogs.
- Add route availability, route-backed parent preselection, and invalid parent-route normalization must live in focused application helpers rather than accumulating in the product shell component.
- Settings access invitation action availability and canonical action/cancel href derivation must live in focused application helpers rather than component-local route string assembly.
- Settings access invitation action labels, destructive tone, disabled state, row option metadata, and confirmation copy must live in focused application helpers rather than component-local conditionals.
- Settings access missing-context, denied, and operation-error presentation must live in focused application helpers rather than component-local conditional copy.
- Settings access list loading and empty-state presentation must live in focused application helpers rather than component-local conditional copy.
- Settings activity audit loading, empty, denied, error, and audit-row presentation must live in focused application helpers rather than component-local conditional copy.
- Settings activity audit rows must default to human-readable action, actor, source, target type, and time labels. Raw principal IDs, target IDs, request IDs, provider identifiers, and metadata values must be secondary technical details rather than the primary scan line.
- Settings overview and administration panel headings, row labels, unavailable values, and disabled action copy must live in focused application helpers rather than component-local conditional copy.
- Asset detail action availability and canonical asset/attachment action href derivation must live in focused application helpers rather than component-local route string assembly.
- Asset detail description fallback, edit-action unavailable copy, and file-list empty-state presentation must live in focused application helpers rather than component-local conditional copy.
- Settings customization archive action and cancel href derivation must live in focused application helpers rather than component-local route string assembly.
- Settings customization archive confirmation title, description, target label, unavailable copy, and disabled state must live in focused application helpers rather than component-local conditionals.
- Settings customization missing-context, denied, and operation-error presentation must live in focused application helpers rather than component-local conditional copy.
- Settings section, access-status, and audit-scope navigation option composition and href derivation must live in focused application helpers rather than component-local route string assembly.
- Settings shell title, context summary, live section announcement, overview context, and missing-inventory presentation must live in focused application helpers rather than component-local conditional copy.
- Settings section navigation helpers should expose stable icon identifiers rather than importing Svelte icon components into application helpers.
- Settings access relationship selector option composition must live in focused application helpers backed by canonical frontend-domain relationship values rather than component-local arrays.
- Settings customization scope, field type, applicability, and target option composition must live in focused application helpers backed by canonical frontend-domain values rather than component-local arrays.
- Asset-detail loading and attachment refresh orchestration must live in focused application helpers rather than accumulating in the product shell component.
- Loaded asset-detail results must be applied through focused application helpers so workspace asset replacement, selected asset state, attachment state, and mode transitions do not accumulate in the product shell.
- Asset-detail photo gallery derivation, supported media-type checks, and photo-upload unavailable reasons must live in focused application helpers rather than component-local derivation.
- Asset-detail photo gallery empty-state copy and unsupported media upload error copy must live in focused application helpers rather than component-local conditional copy.
- Import source options, source summaries, apply-status copy, preview source summaries, message tone derivation, ready-state checks, and import request construction must live in focused application helpers rather than component-local derivation.
- Import missing-inventory, denied, first-run preview, planned-count, and applied-result presentation must live in focused application helpers rather than component-local conditional copy.
- Import preview/apply message detail formatting and apply-message section presentation must live in focused application helpers rather than component-local conditional copy.
- Domain-oriented frontend observability must be represented through an explicit helper or port, even when the first implementation records events only in memory.

The first promotion may include a local seeded adapter for unauthenticated browser review and for unavailable backend operations, but it must be truthful:

- Seeded data must be isolated behind the same repository port as the API adapter.
- Seeded behavior must not be presented as saved backend state.
- Operations backed by unavailable API capabilities must show unavailable, disabled, local-demo, or otherwise explicit state rather than pretending production persistence exists.
- Once the corresponding API operations are exposed through the generated client package, the API adapter must replace seeded behavior for those production paths.

Authenticated workspace loading must use real API discovery:

- The API adapter must call `GET /me/tenants` to discover active tenants visible to the authenticated principal.
- The initial selected tenant should be the first visible tenant unless the user has already selected a tenant in the current browser session.
- The API adapter must list inventories through `GET /tenants/{tenantId}/inventories` for the selected tenant.
- The initial selected inventory should be the first visible inventory unless the user has already selected an inventory in the current browser session and it is still visible.
- The API adapter must list active assets for the selected inventory through the generated client wrapper.
- Tenant and inventory names in the context switcher must come from API responses, not placeholder labels.
- Edit/add affordances must derive from effective inventory permissions in the API response, not from hard-coded editor assumptions.
- Add/create asset affordances must require `create_asset`; broader edit affordances may use `edit_asset` or a separate edit capability.
- If the authenticated user has no visible tenants or no visible inventories, the workspace must show the existing create/setup empty state rather than local seeded data.

Tenant and inventory switching must preserve the API boundary:

- Switching tenants must go through a frontend repository port method instead of requiring components to know or synthesize inventory IDs.
- Selecting a tenant must load that tenant's inventories from `GET /tenants/{tenantId}/inventories`.
- If the newly selected tenant has visible inventories, the first visible inventory should become selected and active assets should be loaded for that inventory.
- If the newly selected tenant has no visible inventories, the workspace must keep the tenant selected, clear the selected inventory, and show the setup empty state.
- Creating from an empty selected tenant must create an inventory in that tenant rather than creating a second tenant.
- Creating from a no-tenant state may create the first tenant and starter inventory together.
- The empty-tenant create action must be hidden or replaced by an explicit denied state when the selected tenant does not grant inventory creation permission.
- Switching inventories within the selected tenant must continue to load active assets through the generated client wrapper.
- The current tenant and inventory selection may be remembered only for the current browser session until a dedicated preferences spec exists.
- Mobile must expose the same tenant-first context switching path through the compact header or an equivalent mobile menu.
- Svelte components must not hold generated DTOs, call generated client methods directly, or construct tenant/inventory fallbacks outside the frontend domain model.

## Data And API Expectations

The UI direction depends on existing domain capabilities:

- Tenant and inventory listing.
- Asset listing with active lifecycle default.
- Location-like assets through `asset.kind = location`.
- Asset containment hierarchy.
- Asset detail reads.
- Asset creation.
- Asset update/move.
- Search across authorized assets.
- Media attachment upload/list/detail where exposed.
- Inventory access and sharing where exposed.

If an API operation is not available during implementation, the UI must expose a truthful loading, unavailable, or disabled state rather than fake saved behavior in production code.

Temporary candidates may use realistic mock data, but promoted web implementation must go through adapter boundaries and generated API contracts.

## Verification

Before this direction is promoted into `apps/web`:

- Update this spec first when the approved behavior changes.
- Implement behind the existing SvelteKit web architecture.
- Run Svelte type checking.
- Run web tests.
- Run the shadcn foundation check after generic primitive changes.
- Browser-level smoke tests must exercise the authenticated workspace shell through runtime config, stored session state, and API-boundary responses rather than relying on removed unauthenticated demo data.
- Run browser-level smoke tests for:
  - desktop shell load,
  - mobile shell load,
  - tenant/inventory switching,
  - global search entry,
  - add menu/tray open,
  - add item without photo,
  - add item with selected photo preview,
  - location card to location list,
  - asset row to asset detail,
  - route-backed settings sections and import source choices,
  - asset edit and move action deep links,
  - unavailable asset action deep-link normalization,
  - back navigation from asset detail to location list,
  - viewer denied edit/add state when applicable.
- Run accessibility checks for dialogs, context switchers, nav, and list/detail flows.
- Run the code critic agent after implementation and fix or explicitly defer findings.

## Non-Goals For First Promotion

- A complete activity feed on the home page.
- "Needs attention" dashboard panels.
- Cross-tenant global search.
- Cross-inventory combined switcher search.
- Full custom field editing UI.
- Full custom asset type management UI.
- Full direct-upload production UX.
- Offline create queues.
- Voice or conversational command execution.
- Public marketing/landing pages.

## Open Questions

- Should desktop eventually support a split-pane location list and asset detail for faster scanning?
- Should location lists default to photo rows, compact rows, or a user-selectable density?
- How should archived assets appear in location views?
- What is the final mobile order and labeling for bottom navigation once Search, Locations, Settings, and future voice interaction all exist?
- Which asset detail actions belong in the first promoted implementation versus later asset management iterations?
- How should expiration-oriented custom fields surface later without turning the home page into a noisy dashboard?
