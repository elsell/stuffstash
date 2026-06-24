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

## Tenant And Inventory Context

The product shell must show the current inventory and tenant context without making tenancy the main content.

Desktop:

- The side navigation must include a compact context switcher near the top.
- The switcher trigger should show the current inventory name and tenant name.
- Opening the switcher should first show the current tenant as a compact header row and the inventories inside that tenant.
- The tenant header row must include a right-aligned `Switch Tenant` action.
- `Switch Tenant` must show a tenant list.
- Selecting a tenant must keep the switcher open and replace the inventory list with that tenant's inventories.
- The switcher must not show one combined dropdown containing all inventories from all tenants.
- The switcher must not include a tenant/inventory search field in the approved first direction.
- The switcher must not include a separate duplicated "current tenant" card above the tenant header.
- Inventory settings may be reachable from the switcher.

Mobile:

- The compact header context control should open a bottom sheet or equivalent mobile-appropriate context switcher.
- The mobile context switcher should follow the same tenant-first behavior as desktop.
- The mobile context switcher must not require a search field for the approved first direction.

## Desktop Header

The desktop top header must prioritize:

- Global inventory search.
- Add action.
- User/account affordance only when not already clear from the side navigation.

Search:

- Search must be front and center on desktop.
- Search must be available across primary web pages.
- Search should feel closer to Google Drive than a command palette: a visible field that accepts ordinary asset/location/container terms.
- Search must be scoped to the selected tenant and inventory unless a future search spec defines cross-inventory behavior.
- Search must preserve tenant and inventory authorization boundaries.

Add:

- The Add action must live in the desktop header.
- The Add action should use a compact menu pattern similar to GitHub's create button.
- The user must be able to choose `Item`, `Container`, or `Location`.
- The add dialog/tray must still allow changing the selected kind after opening.
- Add must be disabled or replaced by an explicit denied state for inventories where the user lacks edit permission.

## Mobile Navigation

Mobile bottom navigation must provide reachable primary actions without duplicating the desktop header.

The approved first mobile bottom navigation direction is:

- Home.
- Search.
- Add as the central primary action.
- Locations or equivalent browse destination when a full route exists.
- Settings or inventory/settings access when it exists.

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
- Location cards or tiles should use photos when available.
- Location cards must open a focused location view.
- The UI must support long location names, missing photos, and empty inventories.

Recently added:

- Recently added assets may appear below the location browse surface.
- Recently added must not dominate the page or compete with search/add.
- Recently added rows must open the asset detail view.

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

The location view must not become a dashboard. It should answer: "What is in this place?"

## Asset Detail View

Selecting an asset must open an asset detail view.

The first asset detail view must include:

- Back navigation to the previous location list when opened from a location.
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

The detail view is the preferred home for asset-level actions such as edit, move, archive, sharing-related actions, attachment management, and future custom field editing.

Asset detail must support:

- Missing photos.
- Long titles.
- Long location trails.
- Viewer or denied states for edit-only actions.
- Archived state when lifecycle views expose archived assets.

## Add Flow

The add workflow must have equal product weight with find/browse.

Entry points:

- Desktop: Add action in the top header.
- Mobile: central Add action in the bottom navigation.

The add surface:

- Must open as a modal, tray, or sheet appropriate to viewport.
- Must let the user choose or change `Item`, `Container`, or `Location`.
- The kind selector must be compact. It must not use large stacked cards.
- Must collect name/title.
- Must collect a valid parent target when required.
- Parent target selection must use a picker/search over valid location/container targets, not a free text field that implies invalid foreign keys can be saved.
- The parent picker must support quick creation when the user realizes the parent location/container does not exist yet.
- Quick creation must be explicit and must preserve authorization, validation, and audit expectations when implemented against the real API.

Photos:

- Photos are first-class and low-friction in the add flow.
- Photos must be optional. A user must be able to save an asset without adding a photo.
- The add surface should expose camera and upload actions.
- The first approved web direction supports JPEG, PNG, and WebP image affordances consistent with the media spec.
- Selected photos should show thumbnails and allow removal before save.
- Invalid or oversized selected photos must block save until removed or corrected.
- Attachment size and supported type rules must come from configuration or a media policy boundary rather than scattered hard-coded checks.

Saved state:

- Successful create must show concise saved feedback.
- When quick-creating a parent and asset together, saved feedback must make both outcomes clear.
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
- Search result rows should open asset or location detail/list surfaces.
- No-results and denied states must be explicit and calm.
- Search must not bypass tenant, inventory, lifecycle, or authorization boundaries.

## Inventory Settings

Inventory settings is the preferred location for inventory-level secondary workflows, including:

- Sharing and access management.
- Inventory metadata.
- Activity/audit views when exposed.
- Tenant/inventory administrative details.

These workflows must not be primary home panels in the approved first direction.

## Responsive Behavior

Desktop:

- Use horizontal space for a persistent side nav, a stable header, denser lists, and detail layouts.
- Avoid filling wide screens with low-value stacked boxes.
- Keep the side nav sticky.

Mobile:

- Use one-column layouts for location and asset views.
- Keep bottom navigation visible and reachable.
- Keep add/search reachable without making the header tall.
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
- Domain-oriented frontend observability must be represented through an explicit helper or port, even when the first implementation records events only in memory.

The first promotion may include a local seeded adapter for unauthenticated browser review and for unavailable backend operations, but it must be truthful:

- Seeded data must be isolated behind the same repository port as the API adapter.
- Seeded behavior must not be presented as saved backend state.
- Operations backed by unavailable API capabilities must show unavailable, disabled, local-demo, or otherwise explicit state rather than pretending production persistence exists.
- Once the corresponding API operations are exposed through the generated client package, the API adapter must replace seeded behavior for those production paths.

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
