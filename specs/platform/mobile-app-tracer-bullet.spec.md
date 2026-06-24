# Mobile App Tracer Bullet Spec

## Purpose

Stuff Stash needs a minimal mobile app slice so the team can verify Expo Go on a real iPhone before investing in mobile product workflows.

## Scope

This spec covers the first mobile application foundation and the first mobile parity pass for the approved web home-hub candidate:

- A separate Expo React Native app in the monorepo.
- A TypeScript entry point that composes domain, application, adapter, and UI layers.
- First mobile surfaces backed by application queries, commands, and repository ports.
- A native bottom tab navigation shell whose labels match the web candidate's mobile navigation language.
- A mobile app frame that mirrors the approved web home-hub candidate's mobile structure.
- A mobile API adapter that consumes the generated `@stuff-stash/api-client` package behind application ports.
- Local development commands for Expo Go.

This spec does not define camera behavior, functional voice capture, realtime conversational transport, release signing, TestFlight, EAS builds, or production mobile distribution. Those behaviors must be introduced through their own specs before implementation.

## Decisions

- The first mobile app must live under `apps/mobile`.
- The first mobile app must use Expo, React Native, and TypeScript.
- The first mobile app must target Expo SDK 55 so Expo Router Native Tabs can use the native bottom accessory API for persistent conversational entry points while remaining testable in Expo Go on supported clients.
- The first screen must be driven by application-layer state loaded through a mobile API adapter.
- The app must not require an Expo account for the first local validation path.
- Physical iPhone validation for Expo SDK 55 may use a local Expo development build when the App Store Expo Go client does not yet support the required SDK version.
- The local development build must use `expo-dev-client` and must be installable from a connected Mac/iPhone through local Xcode signing before relying on EAS or TestFlight.
- The app must not add native modules beyond Expo-compatible navigation and development-client dependencies for the first local validation path.
- The mobile API adapter must use the generated `@stuff-stash/api-client` package rather than hand-written endpoint fetches.
- Expo Go local development must receive mobile runtime configuration through Expo public environment variables:
  - `EXPO_PUBLIC_STUFF_STASH_API_BASE_URL`
  - `EXPO_PUBLIC_STUFF_STASH_TENANT_ID`
  - `EXPO_PUBLIC_STUFF_STASH_DEV_TOKEN`
- The local-dev token value is a development-only credential for the API's local-dev auth mode. Production authentication must wait for the mobile authentication spec.
- Until the API exposes a principal-scoped tenant discovery endpoint, the first real-API mobile slice may use `EXPO_PUBLIC_STUFF_STASH_TENANT_ID` to identify the tenant whose inventories should be listed.
- The first navigation shell must use the iOS and Android system tab bar through Expo Router Native Tabs.
- React Navigation JavaScript bottom tabs are not sufficient for the first mobile shell because they do not render the iPhone-native tab bar.
- The mobile app must use the current approved brand glyph asset from `docs/public/brand/stuff-stash-glyph.png` for local app identity and in-app brand marks until a mobile-specific app icon asset is specified.
- Mobile color tokens must follow `specs/platform/brand-guidelines.spec.md`:
  - System grays and whites for most task surfaces.
  - Charcoal and dusty blue from the current logo direction for brand identity and calm selection accents.
  - A familiar system-like blue for primary actions, links, and interactive emphasis.
  - Amber only for real warning or attention states.
- Mobile task typography must use platform-native system typography. A custom wordmark font must not be added to mobile until a pinned mobile-compatible font asset and dependency strategy are specified.
- Mobile product UI icons must use pinned icon packages rather than text initials or generic letter badges. Mobile surfaces that display tenant or inventory names must include the corresponding semantic icon next to that name so the tenant boundary remains visually distinct from inventory selection, except the compact Home inventory context control may show the tenant as muted inline prefix text before a slash and use the inventory icon only.
- The first tab set must use the same primary navigation language as the web home-hub candidate while preserving enough native bottom-bar width for the persistent Voice accessory:
  - `Home` for inventory overview.
  - `Search` for asset lookup inside the configured tenant.
  - `Add` for creating a base item, container, or location asset through the API.
  - `Locations` for location-first browsing.
- `Settings` must remain available as a native stack route from the Home surface rather than consuming a bottom-tab slot, because the bottom bar should reserve primary workflow tabs for inventory activity and leave the far-right accessory eligible for native inline placement.
- A persistent far-right `Voice` native bottom accessory must provide the first conversational inventory entry point using a microphone symbol.
- Voice must not be modeled as a sixth native tab and must not use the `search` tab role. The search role is reserved for search semantics and may keep system-controlled Search labeling or overflow behavior.
- The first Voice accessory must use Expo Router Native Tabs `NativeTabs.BottomAccessory` on SDK 55 or later. The platform owns whether the accessory is presented in `inline` or `regular` placement; the app must not model Voice as a tab or rely on the search tab role to force placement.
- The Voice accessory must render as a compact contextual voice tray in regular placement and an icon-only microphone control in inline placement. In regular placement it may show preview-only state, such as ready, listening, or review-needed, plus a contextual label for the current app surface.
- The Voice accessory and full Voice route must share the same preview state so the accessory can represent low-friction voice activity, plan review, and cancellation without being only a button that opens a page.
- Until realtime conversational transport is implemented, the `Voice` tab must render a deterministic UX preview only. It must not request microphone permissions, capture audio, stream realtime events, call model providers, create action plans through the API, or execute inventory commands.
- Voice, sharing, audit history, or account management tabs beyond this UX preview must wait until their behavior is specified.
- The first `Home` surface must mirror the mobile layout of the web home-hub candidate:
  - A sticky top inventory context control showing selected inventory and tenant.
  - A native platform sheet tenant and inventory switcher opened from that context control, sized to fit its content when the platform supports sheet detents.
  - The switcher must make the tenant boundary visible with the current tenant as a prominent top section and a smaller `Switch tenant` action.
  - Tapping `Switch tenant` must show the API-visible tenant list. Selecting a tenant must return the sheet to the inventory list scoped to that tenant.
  - Inventory switching that changes the selected inventory for mobile queries and commands during the current app session.
  - A horizontally scrolling recent-assets ticker near the top with up to 10 most recently added or changed assets and a `See all` action.
  - A small preview of top-level location cards with a `View all` action.
  - No account affordance until there is a specified account, profile, or authentication interaction.
- Home must not show dashboard metric tiles. The inventory home workspace is a browse and recency surface, not an analytics dashboard.
- The Home locations preview must intentionally show only a few top-level locations. The full location browser remains the `Locations` tab.
- The Home recent-assets ticker must open asset detail routes. Its `See all` action must open a native stack asset list for the selected inventory.
- Until the API exposes sortable created/updated timestamps in the mobile summary contract, the mobile app may treat the selected inventory asset summary order as the most-recent order and may display the API-provided `updatedAtLabel`.
- The mobile app may keep Expo Router Native Tabs as the platform-native bottom navigation mechanism, but the on-screen content hierarchy must follow the web candidate's mobile hub layout.
- Inventory switching in the first mobile slice must operate over the authenticated principal's API-visible tenants and inventories. It must not imply tenant membership management.
- The configured tenant ID remains a local-development default selection hint only. It must not prevent other API-visible tenants from appearing in the mobile switcher.
- The selected inventory may be held in memory for the first Expo Go validation slice. Durable user preferences must wait for a settings or local persistence spec.
- The first mobile parity pass must preserve the same mental model as the web candidate:
  - Tenant contains inventories.
  - Inventory contains locations and assets.
  - Locations are assets with kind `location`.
  - Recently added, full asset lists, and search results open the same asset-detail language.
  - Add creates one asset in the selected inventory and optional parent location/container.
- The first mobile asset detail view must support selecting an API-backed asset and showing a read-only detail route based on the web asset-detail candidate:
  - A photo-first hero area.
  - Asset kind and optional custom type badges.
  - Title and description.
  - Location, lifecycle state, and updated-at metadata.
  - Edit and move affordances that are visibly unavailable until state-changing commands are specified.
- The read-only asset detail view must remain in the mobile UI/application layer. It must not introduce editing commands, move commands, media upload, camera behavior, or archive behavior.
- Search must call the API search endpoint through the generated API client wrapper and render asset results using the same asset row language as `Recently added`.
- Search results must render as image-first asset cards and open the same read-only mobile asset detail view as other asset entry points.
- Location cards must open a location-scoped asset list for the selected inventory. The first location list may be assembled from the selected inventory asset summary by selecting assets whose current location or immediate parent is the selected location.
- Location-scoped asset lists must render image-first asset cards and open the same read-only mobile asset detail view as search results.
- Search result details, location-scoped asset lists, and location asset details must be native stack routes above the native tab shell so iOS owns the standard edge-swipe back gesture.
- Swipe-back must not be the only navigation path. Native stack back affordances or visible Back controls remain required for accessibility and users who do not use gestures.
- Image-first asset cards must reserve stable media space even when the API has no asset photo URL. When no user photo is available, the card must show a calm type/kind placeholder rather than a decorative illustration.
- When API asset attachments exist, mobile asset cards, recent asset cards, location cards, and read-only asset detail views must render the first active attachment thumbnail through the generated API client wrapper. Authorization details remain adapter-owned infrastructure and must not leak generated DTOs into UI components.
- The selected-inventory asset list must render image-first asset cards and open the same read-only mobile asset detail view as search results and location-scoped lists.
- The first mobile asset detail view must be reusable across Home recent assets, selected-inventory asset lists, Search results, and Location asset lists. It remains read-only until editing, move, media, and archive commands are specified.
- Add must call the API create asset endpoint through the generated API client wrapper for base asset fields supported by the current API: kind, title, description, and optional parent asset ID.
- Add must not fully enumerate every location as the primary location picker. It must provide a typed location field that shows a small set of matching existing locations and keeps `Inventory root` available without turning large inventories into long forms.
- If the typed location field does not exactly match an existing location, Add must let the user create a minimal location asset with one action, then use that new location as the parent for the asset being created.
- Add must support selected-photo management before save: add photos from the device library, preview selected photos, remove selected photos, and upload the remaining photos as asset attachments after the asset is created. The first slice may use the existing JSON base64 attachment endpoint and must keep picker/transport details behind application/adapters.
- The mobile API adapter must map API effective access metadata into mobile inventory role and capability state at the adapter boundary. It must use inventory `access.relationship` for display labels and inventory `access.permissions` for workflow affordances such as whether Add is reachable. The API remains the authorization authority for every state-changing operation.
- Local mobile validation datasets may live under `.stuffstash/seed-data/` and `.stuffstash/seed-media/`, must be ignored by Git, and must seed the in-memory API only through public REST endpoints for tenants, inventories, assets, and attachments.
- Local mobile validation against the in-memory API should disable HTTP rate limiting with `STUFF_STASH_HTTP_RATE_LIMIT_ENABLED=false`; the mobile browsing path may legitimately issue many API and thumbnail requests from one device while the app is being iterated locally.
- Settings must be production-shaped even in the Expo Go validation slice:
  - It must show the current authenticated principal from the API when available.
  - It must show About information for Stuff Stash, including the mobile app version.
  - It may show clearly labeled developer diagnostics such as the configured API base URL and local-development authentication mode.
  - It must not show current inventory context, tenant switching state, unfinished account controls, sharing management, unavailable feature panels, or OIDC mobile authentication until those behaviors are specified.
- The first mobile voice UX preview must preserve the conversational inventory specs' mental model:
  - It must show the current tenant and inventory context.
  - It must represent the journey as user utterance, transcript, assistant interpretation, structured action-plan review, approval, and cancellation.
  - It must communicate that execution is unavailable without presenting unfinished production controls as working.
  - It must keep all data local and deterministic until realtime ports and action-plan APIs exist.
  - The persistent accessory may preview voice activity and review state, but it must not request microphone permissions, record audio, call realtime services, or execute approved plans until the supporting ports and APIs are specified.
- Mobile source code must use hexagonal organization from the start:
  - `src/domain/` owns client-side domain models, value objects, and domain calculations.
  - `src/application/` owns use cases, query services, ports, and view-model assembly.
  - `src/adapters/` owns infrastructure implementations of application ports.
  - `src/bootstrap/` owns dependency wiring.
  - `src/ui/` owns React Native screens, components, and theme tokens.
- UI code must consume application view models or commands. It must not construct domain entities directly and must not depend on generated API DTOs.
- Navigation code belongs under `src/app/` and `src/ui/navigation/` and may depend on Expo Router, UI screens, and application services supplied by bootstrap wiring.
- API data must live behind an adapter implementing an application port so generated transport DTOs do not leak into domain or UI code.
- Domain concepts used in the mobile app must match existing Stuff Stash language, including inventories, assets, locations, containers, lifecycle state, and media readiness.

## Requirements

- Mobile dependencies must be pinned exactly in `apps/mobile/package.json`.
- Mobile dependency versions must be recorded in `specs/platform/tooling-versions.spec.md` before use.
- `pnpm --dir apps/mobile start` must start the Expo development server.
- The root package must expose a convenience script for the mobile development server.
- The root package must expose a convenience script for mobile tests.
- The first screen must make it obvious that the app launched successfully in Expo Go while showing real API-backed Stuff Stash product concepts when API configuration is present.
- The first screen must support loading, ready, and error states at the UI boundary.
- Each implemented tab screen must be backed by an application query, not hard-coded UI data.
- Asset detail preview data must be assembled by application queries from domain asset summaries.
- Search, location asset lists, and asset detail views must be backed by application view models rather than UI-local DTO mapping.
- Read-only and browsing mobile surfaces must support native pull-to-refresh when they show API-backed state, including Home, Search results after a query, Locations, selected-inventory asset lists, location asset lists, asset detail, and Settings.
- Pull-to-refresh must re-run the relevant application query through existing ports and adapters. It must not bypass application queries or call generated API DTOs directly from UI code.
- State-changing forms such as Add must not add pull-to-refresh until a future form reset or draft-recovery behavior is specified.
- Add and search behavior must be covered by focused tests using fakes rather than mocks.
- Application behavior must be covered by focused tests using fakes rather than mocks.

## Verification

- `pnpm --dir apps/mobile check` must type-check the mobile scaffold.
- `pnpm --dir apps/mobile test` must run focused mobile application tests.
- The Expo development server should start with `pnpm --dir apps/mobile start`.
- iPhone verification is manual: install Expo Go, run the mobile dev server, scan the QR code, and confirm the Stuff Stash home screen appears.
