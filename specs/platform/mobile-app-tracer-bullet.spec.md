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

This spec defines camera behavior only for attaching still photos during the Add flow. The first functional realtime voice query slice is specified in `specs/agent-model/mobile-realtime-voice-query.spec.md`. This spec does not define video capture, release signing, TestFlight, EAS builds, or production mobile distribution. Those behaviors must be introduced through their own specs before implementation.

## Decisions

- The first mobile app must live under `apps/mobile`.
- The first mobile app must use Expo, React Native, and TypeScript.
- The first mobile app must target Expo SDK 55 so Expo Router Native Tabs can use the native bottom accessory API for persistent conversational entry points while remaining testable in Expo Go on supported clients.
- The first screen must be driven by application-layer state loaded through a mobile API adapter.
- The app must not require an Expo account for the first local validation path.
- Physical iPhone validation for Expo SDK 55 may use a local Expo development build when the App Store Expo Go client does not yet support the required SDK version.
- The local development build must use `expo-dev-client` and must be installable from a connected Mac/iPhone through local Xcode signing before relying on EAS or TestFlight.
- The app must not add native modules beyond Expo-compatible navigation, development-client dependencies, Expo FileSystem for durable non-secret onboarding profile storage, Expo Auth Session/Web Browser/Secure Store for mobile OIDC authentication, Expo Clipboard for explicit one-time invitation-link copy actions, and Expo Audio for the specified realtime voice query slice in the first local validation path.
- The mobile API adapter must use the generated `@stuff-stash/api-client` package rather than hand-written endpoint fetches.
- Expo local development may seed mobile runtime configuration through Expo public environment variables:
  - `EXPO_PUBLIC_STUFF_STASH_API_BASE_URL`
  - `EXPO_PUBLIC_STUFF_STASH_TENANT_ID`
  - `EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN`, an HTTPS origin used to register and validate universal invitation links; the custom `stuffstash://invitations/accept` route remains available for local development.
- Optional mobile developer diagnostics must be configurable through `EXPO_PUBLIC_STUFF_STASH_VOICE_DIAGNOSTICS_ENABLED`, and the Expo app config must mirror that value into `extra.stuffStash` so native development builds and dev-client reloads see the same runtime seed values as the JavaScript bundle.
- The local-dev token value is a development-only credential for the API's local-dev auth mode. Production mobile authentication is defined in `specs/identity-access/mobile-oidc-authentication.spec.md`.
- Expo public environment variables are development defaults only. The app must not require them for first launch once onboarding exists.
- On first launch without a saved connection profile, mobile must show an onboarding flow before the tab shell:
  - The first onboarding step asks for a Stuff Stash instance URL.
  - The URL must be normalized and saved in durable app-local storage so the user does not need to type it again on later launches.
  - Durable connection profile storage must contain only non-secret profile metadata, such as the instance URL and selected tenant ID. Development tokens and future production credentials must not be written to this profile file.
  - After the instance URL is saved, onboarding must guide the user through the OIDC SSO flow specified in `specs/identity-access/mobile-oidc-authentication.spec.md`.
  - After the instance URL is saved, onboarding must guide the user to create a tenant if the authenticated principal has no usable tenant.
  - After tenant creation, onboarding must guide the user to create an initial inventory if that tenant has no usable inventory.
  - After initial inventory creation, the app must enter the regular Home/Browse tab shell backed by the newly created tenant and inventory.
- A saved connection profile may include the selected tenant ID. Until durable selected-inventory preferences are specified, inventory selection may remain session-scoped and default to the first visible inventory for the selected tenant.
- Settings must expose a way to revisit or reset the saved instance connection profile. Changing the profile must rebuild mobile application services instead of requiring an app reinstall.
- Production mobile onboarding must use the OIDC flow defined in `specs/identity-access/mobile-oidc-authentication.spec.md`. The local-development auth token path may remain available only as an explicit development fallback.
- Mobile must provide one shared feedback surface for transient status and error
  messages. Screens and workflow components should use this shared notice surface
  for action results, refresh failures when stale content remains usable,
  background upload failures, save confirmations, and recoverable non-blocking
  problems instead of inventing screen-local toast, alert, or status patterns.
- Shared mobile notices must render as native-feeling compact top banners below
  the safe-area/status-bar region so they remain visible while forms and the
  keyboard are active. Users must be able to dismiss a notice by tapping it or
  swiping it upward. Notices must animate in and out with short, native-feeling
  motion, including auto-dismiss and gesture-dismiss paths. Notices must support
  semantic tones for success, info, warning, and error; concise text; and at
  most one action such as `Retry`, `Undo`, or `Sign in`.
- Blocking native dialogs remain appropriate for destructive confirmations,
  permission blockers, and authentication/session loss. Dialogs must use calm,
  direct copy and one obvious primary action. They must not be used for ordinary
  saved-state acknowledgement or recoverable background failures.
- When a mounted authenticated surface receives an authentication-required
  result, mobile must clear stale secure auth state, preserve the saved instance
  URL and tenant hint, show a native session-expired dialog, and return to the
  onboarding sign-in step. It must not leave the user on a generic `Could not
  load` screen that requires manual app reset.
- The first navigation shell must use the iOS and Android system tab bar through Expo Router Native Tabs.
- React Navigation JavaScript bottom tabs are not sufficient for the first mobile shell because they do not render the iPhone-native tab bar.
- The mobile app must use the current approved brand glyph asset from `docs/public/brand/stuff-stash-glyph.png` for local app identity and in-app brand marks until a mobile-specific app icon asset is specified.
- Mobile color tokens must follow `specs/platform/brand-guidelines.spec.md`:
  - System grays and whites for most task surfaces.
  - Charcoal and dusty blue from the current logo direction for brand identity and calm selection accents.
  - A familiar system-like blue for primary actions, links, and interactive emphasis.
  - Amber only for real warning or attention states.
- Mobile task typography must use platform-native system typography. A custom wordmark font must not be added to mobile until a pinned mobile-compatible font asset and dependency strategy are specified.
- Mobile product UI icons must use pinned icon packages rather than text initials or generic letter badges. Mobile surfaces that display tenant or inventory names must include the corresponding semantic icon next to that name so the tenant boundary remains visually distinct from inventory selection, except the compact Home navigation context may distinguish inventory and tenant through title/subtitle hierarchy, a disclosure indicator, and an explicit screen-reader label without repeating brand or domain icons.
- The first tab set must use the same primary navigation language as the web home-hub candidate while preserving enough native bottom-bar width for the persistent Voice accessory:
  - `Home` for inventory overview.
  - `Browse` for list browsing, search, filtering, location-first browsing, and full-inventory Map exploration inside the configured tenant.
  - Location-first browsing must live inside `Browse` as the `Places` scope instead of consuming a separate top-level tab. Full-inventory containment exploration must live inside Browse as the separate `Map` sub-surface.
- `Add` is a command that starts a creation workflow, not a peer navigation destination. It must open from a consistently labeled 44-by-44-point-or-larger toolbar action on Home and Browse into a native stack or sheet route, and must not consume a bottom-tab slot.
- `Settings` must remain available as a native stack route from a trailing account affordance on Home rather than consuming a bottom-tab slot. The account affordance may use the authenticated principal's avatar or a familiar person symbol and must expose an explicit `Open account and settings` accessibility label.
- A persistent far-right `Voice` native bottom accessory must provide the first conversational inventory entry point using a microphone symbol.
- Voice must not be modeled as a sixth native tab and must not use the `search` tab role. The search role is reserved for search semantics and may keep system-controlled Search labeling or overflow behavior.
- The first Voice accessory must use Expo Router Native Tabs `NativeTabs.BottomAccessory` on SDK 55 or later. The platform owns whether the accessory is presented in `inline` or `regular` placement; the app must not model Voice as a tab or rely on the search tab role to force placement.
- The Voice accessory must render as a compact contextual voice tray in regular placement and an icon-only microphone control in inline placement. In regular placement it may show preview-only state, such as ready, listening, or review-needed, plus a contextual label for the current app surface.
- The Voice accessory and full Voice route must share the same preview state so the accessory can represent low-friction voice activity, plan review, and cancellation without being only a button that opens a page.
- Until realtime conversational transport is implemented, the `Voice` tab must render a deterministic UX preview only. It must not request microphone permissions, capture audio, stream realtime events, call model providers, create action plans through the API, or execute inventory commands.
- Voice, sharing, audit history, or account management tabs beyond this UX preview must wait until their behavior is specified.
- The first `Home` surface must be a calm, activity-focused native dashboard:
  - A compact top navigation context showing the selected inventory as the primary title, the tenant as secondary context, and a disclosure indicator without a bordered container or repeated Stuff Stash brand mark.
  - The inventory context and trailing account affordance must share one navigation-height alignment grid. They must not use mismatched standalone card heights.
  - The inventory context must expose an explicit screen-reader label such as `Current inventory Household, tenant Home. Switch inventory` and a 44-by-44-point-or-larger target.
  - A native platform sheet tenant and inventory switcher opened from that context control, sized to fit its content when the platform supports sheet detents.
  - The switcher must make the tenant boundary visible with the current tenant as a prominent top section and a smaller `Switch tenant` action.
  - Tapping `Switch tenant` must show the API-visible tenant list. Selecting a tenant must return the sheet to the inventory list scoped to that tenant.
  - Inventory switching that changes the selected inventory for mobile queries and commands during the current app session.
  - A `Recently changed` section near the top with a bounded set of compact shared asset entries and a `See all` action. Entries must include API-derived relative update context so the section communicates why each asset is recent.
  - A `Checked out` attention section that appears only when assets are checked out and uses compact shared asset entries with direct Return actions instead of large image-dominant cards. A quiet empty sentence must not reserve a full section when there is nothing checked out.
  - No permanent Locations or Places preview. Place discovery belongs to the `Places` scope in Browse; an empty-inventory onboarding state may offer a contextual create-place or browse-places action.
- Home must not show dashboard metric tiles. The inventory home workspace is a browse and recency surface, not an analytics dashboard.
- Home must use whitespace, semantic background hierarchy, and restrained separators instead of wrapping the navigation context and every content group in one-point rounded borders. Filled action color must remain reserved for actions; normal metadata and noninteractive decoration must not compete with links.
- Home typography must use platform system styles and Dynamic Type, with no repeated black-weight treatment. Inventory context and section headings may use semibold or bold hierarchy; asset titles should use semibold; supporting recency, placement, and checkout metadata should use regular or medium weight.
- Home section headings must expose heading semantics. `See all`, `View all`, Add, inventory switching, account/settings, and Return must each provide explicit context-rich accessibility labels and at least a 44-by-44-point target.
- Home must remain usable at the largest supported accessibility text sizes without clipping primary labels or relying on fixed-width horizontal cards. Meaningful icons must scale with their associated text where appropriate.
- The Home scroll container must derive enough bottom inset from the native tab bar and persistent Voice accessory for the final row and action to scroll fully above both floating surfaces in regular and minimized states.
- The Home recent-assets section must open asset detail routes. Its `See all` action must open a native stack asset list for the selected inventory.
- The Home recent-assets section must use the API asset recency contract, requesting assets in updated-descending order and deriving labels from API-provided timestamps. Mobile must not infer recency from asset IDs or page through an entire inventory only to sort locally.
- The Home recent-assets section must include all asset kinds in API-provided recency order. After a successful Add flow, switching to Home or pulling to refresh must make the newly created item eligible to appear ahead of older locations and containers.
- Home loading may use a centered progress state. A first-load failure must expose a direct `Retry` action; a refresh failure while prior content remains usable must preserve that content and use the shared notice surface.
- The mobile app may keep Expo Router Native Tabs as the platform-native bottom navigation mechanism, but the on-screen content hierarchy must follow the web candidate's mobile hub layout.
- Inventory switching in the first mobile slice must operate over the authenticated principal's API-visible tenants and inventories. It must not imply tenant membership management.
- The configured tenant ID remains a local-development default selection hint only. It must not prevent other API-visible tenants from appearing in the mobile switcher.
- The selected inventory may be held in memory for the first Expo Go validation slice. Durable user preferences must wait for a settings or local persistence spec.
- The first mobile parity pass must preserve the same mental model as the web candidate:
  - Tenant contains inventories.
  - Inventory contains locations and assets.
  - Locations are assets with kind `location`.
  - Recently added, full asset lists, and search results open the same asset-detail language.
  - Add creates one asset in the selected inventory and optional parent asset.
- Mobile asset detail must be an asset workspace, not a read-only card or a pile of unrelated buttons:
  - A photo-first hero area must support multiple visible photo positions, stable placeholder space, and an obvious `Add photos` affordance.
  - Items, containers, and locations must use the same shared asset-detail
    photo gallery and full-screen viewer. Containable assets must not suppress
    the gallery-level add-photo affordance and recreate it later as a separate
    maintenance control.
  - The first implementation may show a locally ordered carousel/strip for available attachment thumbnails and newly selected upload drafts. Full cover-photo selection, persisted attachment reordering, and attachment deletion require a future media-management API slice.
  - The asset's own identity must read as the current leaf/detail page, not as another browse card. The title, kind, optional custom type, overflow action, metadata, and primary actions should live in unframed page sections rather than a bordered rounded card surface.
  - Card styling is reserved for browse/search/home entry points and repeated contained-asset rows. The detail page may use native-feeling separators, grouped rows, and inline toolbars, but must not wrap the core asset identity in a standalone card that visually competes with browse results.
  - The workspace must show title, kind, optional custom type, location path, lifecycle state, and updated-at metadata without hiding the relevant next action below low-value chrome.
  - Asset detail identity must prioritize the user-provided title before kind, custom type, tags, and other classification metadata. Kind and custom type must use secondary typography rather than appearing as higher-priority badges above the title.
  - The native navigation bar and page identity must not present the same asset title as two competing headings. The page owns the prominent asset title; the navigation bar should use concise route context and a trailing overflow action.
  - The asset's structured parent trail must appear directly below identity as clickable breadcrumb-style placement context, before description and routine metadata. Each breadcrumb must use its parent asset ID from the application view model and open that parent asset's workspace, matching asset-card breadcrumb behavior. Titles containing `/` or other separator-like characters must not be parsed to derive breadcrumb structure.
  - Root-level items and containers must use calm user-facing placement copy such as `No location`; asset detail must not expose `Inventory root` as the primary placement label. Root-level locations must omit the missing-placement treatment or show quiet top-level context because the place itself is not missing a location.
  - Normal states such as active lifecycle and available checkout must not dominate the metadata hierarchy. Checked-out and archived states must remain clearly visible near the relevant action, while updated-at metadata may remain quiet and secondary.
  - Item details must expose clear maintenance actions for `Edit`, `Move`, and `Add photos` without implying unsupported spatial behavior. The first implementation may render those as a compact utility toolbar rather than a filled primary button when no single maintenance action is universally primary.
  - Generic asset maintenance actions such as `Edit`, `Move`, and `Add photos` must not visually outrank spatial work on container and location details. For containers and locations, the filled system-blue primary action belongs to the most relevant spatial action, initially `Add item here`; generic maintenance actions should use quiet native-feeling utility controls or overflow placement. Location movement must be labeled `Move place` anywhere it could be confused with `Move items here`.
  - The visible maintenance toolbar must render only applicable actions. It must not keep unavailable checkout, return, edit, move, or photo actions as disabled peers merely to preserve a fixed row shape.
  - A visible utility toolbar must contain no more than three maintenance actions, keep every target at least 44 by 44 points, and adapt without truncating labels at supported Dynamic Type sizes. Less-frequent actions belong in the overflow.
  - Checkout and return are primary asset availability actions once the API exposes `specs/assets/asset-checkout.spec.md`. An active portable asset that is not checked out must expose a direct `Check out` action. Location assets are non-portable and must not expose checkout. Any existing asset with an open checkout must expose a direct `Return` action, including archived assets when the detail surface can show them for recovery. Optional details must not block the fastest path; the UI may offer `Add details` as a secondary follow-up after checkout.
  - Asset detail must render only the currently applicable availability action. `Check out` and `Return` must never appear together as enabled/disabled peers, and the visible label must use the verb phrase `Check out`.
  - Item availability actions must be visually distinct from quiet maintenance actions. Container and location workspaces may keep their filled primary treatment for `Add item here`, but any applicable checkout or return action must still remain direct and visible near the asset's availability state.
  - A recently used, searched, scanned, or browsed asset must be check-outable from a fresh app launch in no more than three primary user actions after the app is ready.
  - Asset cards, recent assets, search results, location lists, map rows, and asset detail must show a compact checked-out indicator when an asset has an open checkout, without implying that the asset moved out of its normal location.
  - Asset cards and recent-asset cards must prioritize the asset name first and placement second. They must show the asset title above the asset's parent location trail, and the parent trail must render as left-aligned breadcrumb chips in a horizontally scrollable row where type chips would otherwise compete for attention. The breadcrumb row must default to the most specific parent when the trail is longer than the available card width. If an asset has no parent trail, the card must not show an empty breadcrumb row. Kind and custom type may remain available in detail views, empty-photo placeholders, or supporting metadata, but they must not occupy the primary card chip row.
  - Mobile card breadcrumbs must be backed by an ordered list of parent location segments from the application view model, including each parent asset ID and title, not by splitting a formatted path string. Parent or asset titles may contain `/` or other separator-like characters without changing breadcrumb structure.
  - Mobile card breadcrumb chips must be tappable links to their parent asset workspace across Home, selected-inventory lists, Search results, and location-scoped asset lists. Non-immediate parent chips should have slightly lower visual weight than the immediate parent, but this distinction must use accessible color and type weight rather than opacity.
  - Mobile card breadcrumb chips should preserve enough text to identify real household locations and containers. Individual breadcrumb segments must not use very short truncation caps; if truncation is needed for extreme names, the cap should be broad enough for ordinary names such as seasonal storage bins, cabinets, and closets, with horizontal scrolling handling longer parent paths.
  - General browse cards must not show photo-readiness chips or low-value last-updated metadata in the visible card body. The compact Home recent variant must show its API-derived relative update label because recency is the reason the entry appears. When an asset is checked out, the checked-out status chip belongs as a compact image overlay or equally clear compact-row status treatment, not next to the location breadcrumb row as if it were placement metadata.
  - Home recent and checked-out entries, selected-inventory lists, Search results, and location-scoped lists must render through the same shared asset-card component or shared asset-entry primitives. The shared implementation may expose standard-card and compact-row variants, optional recency metadata, optional tag visibility, and an optional footer action, but media and placeholder treatment, title, checked-out status, parent breadcrumbs, and tag presentation must not be independently reimplemented by Home. Compact Home rows may omit long-form description and search-match metadata.
  - A mobile workspace load that hydrates the selected inventory's active placement tree must reuse that same ordered asset data when mapping checked-out, browse, Search, and location results. Home must load its workspace and checked-out summaries through one required dashboard snapshot port rather than an optional compatibility fallback or a second workspace load. A paginated card request must not trigger a second unbounded traversal of the selected inventory solely to rebuild breadcrumbs.
  - Each asset card must expose one clearly named screen-reader action for opening the asset. Visually separate image and metadata hit regions may invoke the same action, but redundant regions must be hidden from accessibility while breadcrumb, tag, and footer actions remain independently operable.
  - Location paths must remain easy to scan on mobile. The current location should be presented as left-aligned breadcrumb-style context rather than a right-aligned table value when the detail page needs compact metadata.
  - Lifecycle, audit, destructive, and other less-frequent secondary actions must live behind an overflow/action-sheet style control instead of occupying the visible utility or spatial action rows.
  - The overflow must expose lifecycle actions when available and a `History` action for read-authorized assets.
  - The overflow must describe itself as general asset actions rather than lifecycle-only actions when it also contains history commands. Logically related history, lifecycle, and irreversible actions must remain ordered and visually distinguishable, and permanent deletion must retain native destructive treatment and explicit confirmation.
  - Asset detail photo presentation must size from the available viewport instead of fixed minimum widths, preserve a stable aspect ratio on narrow phones and larger screens, and expose multiple-photo position without treating the first attachment as a persisted cover photo. Raw file names belong in the full-screen viewer rather than over the normal workspace image.
  - The no-photo state must keep stable media space while presenting one clear add-photo affordance. The workspace must avoid repeating equivalent `Add photos` controls at the same hierarchy level.
  - Asset detail typography and controls must use platform-native system typography, support Dynamic Type without hiding primary labels, and avoid relying on repeated heavy or black weights for hierarchy. Meaningful controls must keep at least 44-by-44-point targets and sufficient separation.
  - Mobile appearance must support `Light`, `Dark`, and `System` as typed user preferences. `System` is the default and follows the current device appearance; an explicit `Light` or `Dark` preference overrides the device appearance for both React Native content and native navigation chrome.
  - The appearance preference is non-secret device-local configuration. It must persist behind a mobile application port and filesystem adapter, remain independent of tenant, inventory, principal, authentication session, and instance connection profile, and survive connection-profile reset.
  - Settings must expose `Light`, `Dark`, and `System` as one mutually exclusive appearance control. Changing the selection must apply immediately without restarting the app, preserve the user's current route and workflow state, and persist for the next launch.
  - Every mobile surface must consume semantic appearance tokens rather than hard-coded light values, including onboarding, native stack and tab chrome, Settings, Home, Add, Browse, asset cards and detail, native sheets, shared notices and dialogs, tenant switching, provider settings, and voice surfaces. Full-screen photo viewing may intentionally remain image-first black in both appearances when its text and controls retain sufficient contrast.
  - Light, dark, increased-contrast-light, and increased-contrast-dark variants must provide semantic surface, text, control-boundary, focus, action, warning, success, and danger roles. Text and meaningful control boundaries must meet the project's WCAG 2.2 AA contrast requirement in every supported appearance. User photos and status meaning must not be inverted or communicated by color alone.
  - Asset detail loading, denied, not-found, and recoverable failure states must not collapse into one dead-end message. Recoverable load failures must expose a direct retry action, and denied or missing assets must use safe copy without leaking unauthorized resource existence.
  - Large contained-asset workspaces must keep primary spatial and maintenance actions reachable and use virtualization or another measured bounded-rendering strategy rather than an unbounded child map inside the page scroll container.
  - Opening a location workspace must use one dedicated mobile application
    repository operation for its active containment workspace. It must not load
    the default inventory summary and then repeat a full Map traversal. The
    adapter may paginate the existing authorized active-asset list once to
    reconstruct the subtree, but it must resolve card thumbnails only for the
    selected location and assets presented in that location workspace; full
    attachment detail remains scoped to the selected asset.
  - The same workspace component must be reusable across Home recent assets, selected-inventory asset lists, Search results, and Location asset lists.
- Mobile asset detail History browsing must be read-only except for the explicit authorized `Revert change` command, and it must remain scoped to the current inventory and asset:
  - The user-facing action is `History`. On iPhone it must open as a native stack destination above the asset workspace rather than a form sheet because it is browsable hierarchical content, not a short modal task. Larger platforms may use an equivalent native adaptive presentation without replacing standard navigation gestures.
  - The mobile app must call the generated asset activity endpoint with the tenant ID, inventory ID, and asset ID from the loaded asset workspace. It must not rediscover scope through the currently selected or default inventory and must not scan broader inventory audit pages.
  - History defaults to `changes` and provides an accessible native filter menu for `Changes` and `All events`. Read audit records remain available in `All events` but must not bury state changes in the default view.
  - The history list must be cursor-paginated newest-first and grouped with localized date context. It must support initial loading, empty changes, empty all-events, refreshing with retained content, loading another page, recoverable pagination failure, permission denied, and safe missing-asset states.
  - History rows must use concise human language derived from typed actions and safe structured changes, such as `Renamed to “Dog Toys”` or `Edited name and tags`. Raw action strings must not be mechanically title-cased for primary user copy.
  - Tapping a history row must push a native detail destination with safe changes, actor, exact localized time, source, an optional authorized `Revert change` action, and a collapsed technical-details group for safe request and audit identifiers.
  - History must use the Git Revert mental model: `Revert change` applies a new compensating command for the selected operation and creates new History. It must not be presented as restoring the whole item to a historical point in time.
  - A historical revert must use a native confirmation that predicts the selected operation's effect in user language, states that only the selected change is reversed, and uses `Revert Change` as the confirmation label. Field edits must name the affected fields; creation, movement, lifecycle, checkout, and return reversals must name their specific archive, location, lifecycle, or checkout outcome. The action must remain available only while the API reports the operation as available.
  - Successful historical reversal must report `Change reverted` and refresh History. If later state makes the operation stale, mobile must preserve the current item state and explain: `This item changed afterward, so this change can’t be safely reverted.`
  - A direct edit save must return to the asset workspace, refresh it, and use the shared notice surface to show `Saved “<title>”`. When the API returns an available operation, the notice must offer one `Undo` action and announce saved/undo state to assistive technology.
  - The detail workspace must display safe audit metadata only: action label, source, actor display label resolved from the API's safe user profile when available, principal ID fallback, occurred-at label, request ID when present, and safe metadata values returned by the API.
  - The History list must prioritize the human action and actor/time context over raw target identifiers. Individual records must be compact native-feeling list rows rather than cards or a decorative custom timeline. Rows that navigate to detail must use disclosure semantics.
  - The mobile UI must not expose raw provider prompts, raw voice transcripts, credentials, storage keys, blob paths, authorization internals, or other sensitive implementation detail.
  - Mobile audit metadata rows must be allowlisted in the application query before rendering. The first allowlist may include user-facing movement, title/name, lifecycle/status, kind/type, attachment file name, content type, file size, and count summaries. Unknown metadata keys and values that look like prompts, transcripts, credentials, storage paths, blob keys, authorization internals, or provider internals must be omitted from History detail.
  - Every interactive control and row must retain at least a 44-by-44-point target, support Dynamic Type through accessibility sizes without hiding the change, expose context-rich VoiceOver labels, preserve sufficient light/dark/increased-contrast contrast, and remain understandable with reduced motion enabled. Decorative event icons must be hidden from accessibility.
  - Pull to refresh must preserve the current view selection. Pagination progress must be unobtrusive, and a failed next page must retain loaded history with a direct retry action.
- Mobile asset detail checkout history must be read-only and scoped to the current inventory and asset:
  - The mobile app must call the generated asset checkout history endpoint for the current tenant, inventory, and asset instead of scanning broader inventory audit pages.
  - The checkout history surface must display checkout, return, safe actor labels, timestamps, and bounded details.
  - Checkout history must not expose raw provider prompts, raw voice transcripts, credentials, authorization internals, hidden inventory data, or audit metadata not needed for the checkout history use case.
- Mobile asset detail edit must be a platform-native stack sheet or pushed form backed by mobile application commands:
  - On Expo Router/native-stack builds that support `formSheet`, edit must use the route-level native sheet presentation, native grabber, and platform dismissal semantics rather than a custom transparent modal.
  - Edit sheets must not allow gesture dismissal to silently discard dirty edits. Until native prevent-remove confirmation is wired for form-sheet gestures, edit sheets may disable gesture dismissal and require the visible cancel action so the discard confirmation remains authoritative.
  - If a platform cannot render a native sheet, the fallback must be the closest platform-native modal screen, not a custom bottom card inside a dimmed overlay.
  - Users may edit title and description.
  - Kind changes remain unavailable until the API exposes a safe conversion/promotion command; the UI may display kind as read-only helper context.
  - The edit surface must show the current kind and custom type, when present, as read-only context so users understand why the editable fields are limited.
  - Save must remain unavailable until the title is non-empty and at least one editable field differs from the loaded asset.
  - Save and discard handling must use the same normalized edit-state rules. Leading and trailing whitespace in editable fields is not meaningful and must not trigger dirty-state warnings or be submitted to the update command.
  - Save must call the API update asset endpoint through a mobile application command and generated API-client adapter.
  - Canceling with unsaved changes must ask for confirmation.
  - After save, the asset workspace must refresh from the application query so server validation, audit-backed updates, updated-at labels, and downstream lists converge.
  - After successful edit, move, archive, or restore operations, the workspace must show a concise native status message near the primary actions and still refresh from the server. This status message must not replace photo upload status, lifecycle confirmations, or safe error alerts.
- Mobile asset movement must be a dedicated placement picker:
  - Move and move-here pickers must use the same native stack sheet presentation rules as edit and audit surfaces.
  - The picker must show the current location path and the proposed destination before saving.
  - Users may search selectable parent candidates by title/path, and result rows must show the candidate title, kind hint, and structured path breadcrumb from the parent lookup result.
  - The move destination picker must not present item assets as valid destinations until item-to-container promotion exists as a real application command.
  - Users may choose `No parent` to move the asset to the inventory root.
  - Users may create a missing destination inline as either a location or a container in the current inventory, then immediately select it as the destination. The picker must expose a compact native kind choice so rooms and broad places become locations while boxes, shelves, bins, cabinets, and similar storage objects become containers.
  - Inline destination creation should preserve spatial context by creating the new destination under the asset's current parent when the asset has one, then selecting the new destination for the move. Root-level assets may create root-level destinations.
  - The picker must reject moving an asset into itself. Server validation remains authoritative for cycles, invalid parent kind, archived parents, cross-tenant, and cross-inventory attempts.
  - After move, the asset workspace must refresh and show the new path.
- Mobile asset detail photo management must use the existing authorized attachment APIs:
  - `Add photos` must use the same native camera/library chooser and selected-photo model as the Add and voice flows.
  - Detail photo upload must prefer the direct-upload path and may use the explicit local-development sentinel fallback, keeping transport details behind adapters.
  - Existing asset photos must render as an ordered horizontal strip backed by attachment metadata and authorized thumbnail references. Dense cards and lists should use the `small` thumbnail variant. Asset detail hero photos should use `medium` thumbnails so the preview is visibly sharper than card media without downloading original bytes.
  - Tapping an existing photo must open the shared full-screen photo viewer with the current photo, position within the asset's photo set, safe file metadata when available, including original file name, media type, and friendly file size, explicit next/previous controls when multiple photos exist, swipe navigation, double-tap zoom in/out, and a removal action when the caller can edit the asset. The viewer should use the `large` thumbnail variant by default, not the card-sized thumbnail and not the original full-size attachment.
  - The shared full-screen photo viewer must be used across persisted asset photos and draft Add-flow photos. Its chrome should follow the iOS Photos mental model: image-first presentation, dark background, swipe and double-tap gestures, and primary actions grouped in a bottom toolbar rather than split between top and bottom bars.
  - Removing an existing photo must call the attachment hard-delete endpoint through a mobile application command and generated API-client adapter, with native confirmation before mutation.
  - Upload progress/status must be visible per selected photo at the asset workspace, including pending, uploading, attached, and failed states. Failed uploads must offer retry without claiming the asset update failed.
  - Photo upload progress rows must be derived from the selected-photo order and updated only by matching progress events so stale progress from a previous selection cannot rewrite the visible retry state.
  - The first implementation may upload selected photos in the visible order. Persisted attachment reordering and cover-photo selection require a future media-management API slice because the current attachment API does not expose ordering or cover-image commands. The detail workspace may show position labels, next/previous carousel affordances, and local reorder controls for the currently loaded photo strip, but must not imply that reordered attachment order has been saved when no persistence command exists.
- Mobile asset detail must expose the asset lifecycle controls currently supported by the API for items, containers, and locations:
  - Active assets show an `Archive` action.
  - Archived assets show a `Restore` action and a destructive `Delete permanently` action.
  - Archive, restore, and permanent delete must call the generated API client through mobile application ports and commands. UI code must not call generated DTO clients directly.
  - Archive, restore, and permanent delete must use native confirmation before mutation. Permanent delete must be framed as irreversible and must not share the same visual weight as ordinary edit or move actions.
  - The lifecycle overflow must name the current asset and separate reversible lifecycle actions from the irreversible permanent-delete action using native destructive styling where available.
  - Lifecycle confirmation copy must name the asset and explain the consequence of the selected operation. Permanent delete confirmation must state that the asset itself and its photos are removed while audit history remains.
  - Archive and restore must refresh the asset detail view from the application query after success so lifecycle state, updated-at labels, and downstream lists converge with API state.
  - Permanent delete must navigate away from the deleted asset detail after success because the asset is no longer readable.
  - API validation failures, including attempts to archive or delete assets with active children or restore assets whose parent is archived, must be shown as safe user-facing errors that name the attempted action and suggest resolving contained assets or parent state without client-side lifecycle workarounds.
- Mobile asset detail for containers and locations must show contained assets directly below the workspace metadata:
  - The first implementation may derive immediate children from the selected inventory summary already loaded by the mobile application query.
  - Child rows must use the same image-first asset language and open the same asset workspace route, but contained rows should be compact enough for spatial scanning inside a known parent rather than reusing the full browse/search card chrome.
  - Contained child presentation must avoid repeating redundant global metadata such as the current parent path when the surrounding workspace already establishes the container or location context.
  - Location workspaces must separate direct child locations and containers as
    navigable spaces from descendant items found anywhere below the location.
    Descendant item rows must retain their structured path relative to the
    current location so an item inside a nested container remains findable and
    understandable. Container workspaces may continue to present immediate
    children only.
  - When a location workspace contains at least 20 combined space and item
    rows, it must expose one compact inline contents search field that filters
    both sections by title and relative path without leaving the workspace.
    Smaller locations must not spend permanent vertical space on this control.
    The search result state must preserve section headings, counts, and a clear
    no-match recovery action.
  - Contained children must have deterministic presentation ordering. The first ordering groups containers and locations before items so nested places remain easy to scan, then sorts each group by the user-visible title and asset ID as a stable tiebreaker.
  - Container and location workspaces should elevate spatial actions over generic asset maintenance: `Add item here` should be the primary spatial action before the contents section, while generic `Edit` and target-specific movement actions remain available as quiet maintenance controls after the contained-assets workspace. Photo management belongs to the shared gallery.
  - Contained-assets headings should anchor the spatial context to the current asset, such as `Inside Garage cabinet` for containers and `In Garage` or `Items in Garage` for locations, with the count as secondary text. A heading must remain adjacent to the rows or empty state it labels; action stacks must not separate them.
  - Active container/location workspaces with edit permission must keep `Add item here` and `Move items here` available as spatial actions whether the container is empty or already has contents.
  - Empty container/location states must still explain that nothing is inside yet and reinforce the same `Add item here` and `Move items here` affordances.
  - `Move items here` must not offer the current container/location itself or assets already directly inside it as movable candidates. Candidate filtering must prefer explicit candidate parent identity over the bounded contained-assets summary when available. Deeper cycle prevention remains server-authoritative until the mobile client has a tree-aware move planner.
  - If the initial bounded move-here candidate set becomes empty after filtering, the sheet must invite search instead of claiming that no movable matches exist.
  - `Add item here` must be available only for active containers or locations that the user may edit. It must navigate to Add with the current container or location preselected as the parent. This route-scoped parent prefill may update only the placement fields of the current Add draft so typed title, description, selected photos, and details state are not lost.
  - Add parent prefill route parameters must be encoded and decoded through one typed mobile route contract. The Add route must ignore incomplete, unsupported, item-kind, or otherwise inapplicable parent-prefill parameters instead of showing an unverified parent as selected.
  - Recursive tree editing and bulk move flows remain future work.
- Browse must be a combined list, search, and map inventory surface for the selected inventory:
  - Browse must contain separate `List` and `Map` sub-surfaces rather than treating Map as a visual layout toggle for list results.
  - The `List` sub-surface owns search-first browsing, scopes, secondary filters, a separate sort control, paginated asset cards, and Places rows.
  - Selecting a Places row must open that location asset's shared detail
    workspace. Places must not fork users into a separate legacy location-only
    list whose identity, photos, actions, and containment language differ from
    the asset workspace. Existing location-scoped deep links may remain as
    compatibility routes for contained-asset browsing.
  - The `Map` sub-surface owns full-inventory containment exploration for the selected inventory. It must always show the selected inventory's containment structure instead of inheriting the current list scope, sort, or search-result subset.
  - With an empty query it must browse selected-inventory assets through the API asset list endpoint.
  - With a non-empty query it must call the API search endpoint through the generated API client wrapper.
  - Activating the Browse tab must not focus the search field or summon the keyboard. Browse is a content-first destination; text entry begins only after the user activates search or follows an explicit text-search entry point. Tag, scope, checkout, lifecycle, and other filter-driven navigation must never auto-focus search.
  - The focused search field must have a visible focus state that is consistent with the Stuff Stash brand tokens.
  - Browse must keep the title `Browse`. The header must show the selected inventory as quiet context so a multi-inventory user can tell which inventory is being searched without leaving the page.
  - Browse refinement controls must follow native search hierarchy: the search field comes first, followed by an always-visible mutually exclusive scope control for `All`, `Places`, `Containers`, and `Items`. Scope is browse navigation and must not be hidden inside the secondary filter disclosure.
  - The refinement/tool region and the first result row must use the standard 16-point section gap. Smaller within-control gaps must not collapse this boundary or vary based on whether applied-filter tokens are present.
  - Secondary filters must be disclosed through one compact `Filters` control. The control must show an active-filter count only when non-default lifecycle, availability, or tag refinements are applied. Default state must not be narrated as a list of active filters.
  - Applied secondary filters must remain visible as removable, user-labeled tokens when the filter surface is closed. Tag filters must use display names rather than only a tag count. `Clear all` must be available when more than one secondary filter is active.
  - The expanded filter surface must use a native sheet or equivalent platform-appropriate temporary surface with separate `Status`, `Availability`, and `Tags` groups. It must maintain draft filter state until the user applies it, support reset and cancellation, and keep each meaningful control target at least 44 by 44 points.
  - `Availability` is the user-facing group label for checkout filtering, with `Any`, `Available`, and `Checked out` options. The mobile Browse UI must not label the group `Checkout`.
  - Sort is not a filter. Browse must expose sort through a separate compact control or native menu. `Recently changed` is the default user-facing label. Until the API exposes a more meaningful alternative such as alphabetical order, the deterministic ID order must be labeled `Default order`, not `Stable`.
  - `Places` must replace the previous top-level `Locations` tab by rendering location-first rows inside Browse. Places rows must show the location title, photo or stable placeholder, contained asset count, and recent contained assets when available.
  - It must render result rows using the same image-first asset card language as `Recently changed`.
  - The default non-Places List presentation must remain a two-column photo-first asset grid on ordinary phone widths. The grid must preserve useful title and placement text on iPhone 15-class widths, adapt safely for larger Dynamic Type or narrower windows, and must not show photo-readiness or last-updated badges.
  - It must lazy-load additional result pages as the user scrolls instead of assuming the first API page contains the whole inventory.
  - Pull-to-refresh must reload the first page for the current query, filters, and sort.
  - Applying scope, filters, or sort must keep the last successful results visible while the next request is pending. A compact progress state may indicate replacement loading, but the grid must not blank or present an empty state until the replacement request succeeds.
  - Recoverable first-load and replacement-load failures must provide a direct retry action. Pagination failures must preserve loaded results and expose a retry affordance near the list footer. A failure to load tag options must not silently remove the Tags filter without status or recovery.
  - Result copy must not present the number of currently loaded paginated rows as the complete result total. Until the API exposes a total, use copy such as `20 shown` or omit the count.
  - Browse empty states must distinguish an empty inventory, a text search with no matches, and refinements with no matches. Empty-inventory copy must offer the existing Add flow; search and filter empty states must offer clear-search or clear-filter recovery.
  - Browse mode must support lifecycle filtering (`Active`, `Archived`, `All`), scope filtering (`All`, `Places`, `Containers`, `Items`), and API-backed sorting (`Recently changed`, `Default order`). Alphabetical title sorting must wait until the API exposes that sort contract.
  - Search mode must support lifecycle and kind filters. Search result order remains API relevance order until the search API exposes explicit sort controls.
  - Sort controls may remain visible in search mode if clearly disabled or described by compact state language; the client must not fake sorted search by loading every result page locally.
  - Mobile route parsing must preserve every supported Browse entry parameter, including scope, query, selected tag IDs, lifecycle state, checkout state, and sort. Home `View all`, tag chips, and deep links must resolve through the same typed applied-filter state.
  - Search should refine results after a short debounce while the user types when API latency and request cancellation make that responsive. Submitting the keyboard search action must remain supported. Clearing text must return immediately to browse results without changing selected tags or other applied filters.
  - Browse interactive controls and cards must provide visible pressed states. Selected filters and scope must not rely on color alone, and List/Map must expose selected tab or segmented-control semantics to assistive technologies.
  - Browse colors must resolve through the same app-wide semantic appearance preference and palette as every other mobile surface. Browse must not maintain an independent theme switch or cache. Focus indicators and meaningful control boundaries must meet the light, dark, and increased-contrast requirements.
- Browse Map must be a native-feeling horizontal containment map for non-destructive inventory exploration:
  - Map must be presented as its own `Map` tab or segmented sub-surface inside Browse, alongside the normal `List` browse surface.
  - Map must use native platform components wherever they can provide the intended behavior, including native search fields, segmented or tab controls, sheets, action menus, pull-to-refresh where applicable, system typography, accessibility semantics, and platform scrolling physics. Custom UI is acceptable for the horizontal containment column mechanics when native components cannot express the interaction cleanly.
  - Map must not be called a graph in product language because Stuff Stash containment is a tree: each asset has at most one parent, and containment cycles are forbidden by the domain model.
  - Map must use the selected inventory as the root and expose the full active containment tree for that inventory, including locations, containers, and items. It must not show only locations or only assets returned by the current list page.
  - Every visible Map row must belong to the selected inventory and must be loadable by the Map detail sheet with full parity to the normal asset detail workspace. The selected-inventory summary and selected active Map tree are peer asset-detail sources; Map detail must not be modeled or implemented as a lesser read-only path.
  - Map must not inherit `List` scope filters such as `All`, `Places`, `Containers`, and `Items`, because those filters would make the structure incomplete and undermine orientation.
  - Map search may help users jump to a matching asset and expand the path to that asset, but it must not replace the map with a filtered result list. Clearing search must leave the map structure intact.
  - Map must use horizontally scrollable columns where each column represents one containment level. The previous and next columns may peek at the left and right edges so users understand that they can move across the open path.
  - Horizontal movement between open columns must be smooth and gesture-friendly. Users must be able to move backward and forward through the open path by horizontal swipe or scroll, not only by tapping breadcrumbs.
  - The horizontal column rail may use a custom controlled pager instead of a native horizontal scroll view when native scroll physics would compete with row-level branch gestures. The controlled pager must still preserve native-feeling tracking, snapping, reduced-motion behavior, native vertical list scrolling inside each column, and pull-to-refresh where applicable.
  - When a horizontal swipe begins on the main body of a container or location
    row and moves left, Map may treat that gesture as branch selection instead
    of horizontal map scrolling. The row card should move left and reveal a
    right-pointing chevron under the row's right side as immediate feedback.
    Once the pull reveals enough of the chevron, the row should become the
    selected branch immediately and the remaining pull should smoothly drive the
    horizontal map toward that branch's next column. Releasing should settle the
    map to the next column; it must not be the primary activation moment. The row
    should spring back without selecting when the gesture never crosses the
    reveal threshold. This row gesture must not replace normal horizontal
    scrolling when the drag begins outside a containing row, when the row is a
    leaf item, or when the user swipes right to move back up the tree.
  - After a containing-row swipe crosses the reveal threshold and is committed
    to branch navigation, vertical list scrolling in the map columns must be
    temporarily locked until that swipe completes. This prevents slight diagonal
    finger movement during the committed horizontal travel from moving the
    source column vertically. Before the reveal threshold, normal vertical list
    scrolling must remain available.
  - Programmatic navigation, including selecting a row or tapping a breadcrumb, must animate smoothly to the relevant column unless the user has requested reduced motion.
  - Deeper containment columns should enter and leave with subtle native-feeling motion instead of flashing in and out. The first implementation should use a small spatial slide plus fade for normal motion, inspired by platform navigation/shared-axis transitions, and reduce that to a fade or instant update when reduced motion is enabled.
  - Breadcrumbs must remain visible, clickable, and synchronized with the active column. Tapping a breadcrumb must move to that level without collapsing unrelated history unless the destination is intentionally reset.
  - The main body of a row must select that asset as the current branch and reveal its immediate children in the next column when the asset is a location or container.
  - Rows that are part of the current expanded path must remain visually
    distinguished from sibling rows so users can identify the active branch when
    scrolling backward through earlier columns.
  - The info affordance on every row must open a native sheet with full asset-detail workspace parity for that asset at any depth, without navigating away from the map or changing the open path. The sheet must use the same detail presentation, photo management, lifecycle actions, refresh behavior, and contained-asset/spatial actions as the normal asset detail route wherever those actions are supported by the current mobile application ports. Map detail must not degrade into a read-only intermediate summary.
  - Map rows must use the available screen space efficiently. They should be compact enough to show several assets per viewport while still providing useful scan information such as title, kind, concise placement or note text, child count when relevant, and a clear info affordance.
  - Map rows must not reserve large empty media or decorative areas when no user photo is available. Photo or kind placeholders should be stable and modest so row text remains useful.
  - Map row density must remain readable and touchable on mobile. Compactness must not reduce primary row and info affordance hit targets below platform accessibility expectations.
  - Leaf item rows must not navigate away from Map by default. They should open the same full-parity detail sheet used by the row info affordance so exploration remains non-destructive and low context-switch.
  - Container and location detail sheets should expose spatial actions such as `Add item here` and `Move items here` when the current principal may edit that inventory and the asset is active.
  - Map detail sheets may still navigate to existing native edit, move, move-here, add, and audit routes for workflows that are already implemented as route-level native sheets. Returning from those routes should refresh the map structure so Browse Map does not show stale information.
  - Map state should remember the open path for the selected inventory during the current app session. It must not persist path state across tenants, inventories, principals, or rebuilt application service compositions until a durable preference spec exists.
  - Map must render large inventories with virtualization or another measured strategy. It must not render an unbounded recursive tree with nested scroll views.
  - If the current API cannot provide a complete containment summary efficiently, the first implementation may assemble the active map from the selected inventory summary already loaded by the mobile application query, but the UI must be structured so a future tree-aware or lazy child-loading port can replace that data source without rewriting row presentation.
  - Empty containers and locations must still appear in Map so the full structure remains visible. Their next column should show a calm empty state and keep spatial actions available when permitted, including an `Add item here` action that opens the native add flow with that empty container or location preselected as the parent when the current principal may create and edit assets there.
  - Archived assets are out of the first Map surface unless the API and product spec define a clear archived-structure visibility mode. Map must not silently mix archived nodes into the normal active containment map.
  - Map interactions must remain read-only until the user chooses an explicit spatial action or full detail action. Expanding, scrolling, searching, breadcrumbs, and opening the full-parity detail sheet must not mutate inventory state.
  - Map UI code must consume application view models assembled from mobile application ports. It must not depend on generated API DTOs or construct domain entities directly.
- Search and browse results must render as image-first asset cards and open the same mobile asset detail view as other asset entry points.
- Location cards must open a location-scoped asset list for the selected inventory. The first location list may be assembled from the selected inventory asset summary by selecting assets whose current location or immediate parent is the selected location.
- Location-scoped asset lists must render image-first asset cards and open the same mobile asset workspace as search results.
- Search result details, location-scoped asset lists, and location asset details must be native stack routes above the native tab shell so iOS owns the standard edge-swipe back gesture.
- Swipe-back must not be the only navigation path. Native stack back affordances or visible Back controls remain required for accessibility and users who do not use gestures.
- Image-first asset cards must reserve stable media space even when the API has no asset photo URL. When no user photo is available, the card must show a calm type/kind placeholder rather than a decorative illustration.
- When API asset attachments exist, mobile asset cards, recent asset cards, location cards, and asset workspace hero areas must render active attachment thumbnails through the generated API client wrapper. Authorization details remain adapter-owned infrastructure and must not leak generated DTOs into UI components.
- The selected-inventory asset list must render image-first asset cards and open the same mobile asset workspace as search results and location-scoped lists.
- Add must call the API create asset endpoint through the generated API client wrapper for base asset fields supported by the current API: kind, title, description, and optional parent asset ID.
- Add is a fallback for moments when voice is not usable and must optimize for low-friction capture over taxonomy-first data entry.
- Add must not ask users to choose between `item` and `container` before capture. New non-location assets start as items; container behavior is inferred later when another asset is placed inside them. If the current API still requires `kind`, the mobile command may send `item` by default while keeping the UI free of the item/container distinction.
- Add may still allow explicit location creation because locations remain user-facing places, but location creation must be expressed through parent placement language rather than a top-level required type selector.
- Add must be photo-first without consuming excessive vertical space. The first interactive region after inventory context must be a compact `Photos` strip whose first square is always an add-photo tile with a plus icon. Tapping the add tile must open the native platform source chooser for camera or photo library. The strip must preview selected photos and support removing photos. Camera and library must not appear as two always-visible sibling buttons in the form.
- Adding or capturing photos must return to the Add form with the new photos in the strip. It must not automatically open the full-screen preview carousel.
- Tapping a selected photo in the Add photo strip must open a full-screen preview carousel for the selected draft photos. The preview must support native-feeling swipe navigation across all selected photos, double-tap zoom, pinch zoom, swipe-down dismissal, closing back to the Add form, and removing the current photo from the draft after confirmation.
- Selected photos must support user-controlled ordering before save. Because draft photos already live in a horizontal strip, the Add form must provide a reorder mode that does not fight normal strip scrolling. The command must upload photos in the user-visible order.
- Description must be hidden behind a secondary `More details` disclosure by default. Users must be able to add a description when needed, but the base Add path must not require encountering a multiline field.
- Save must remain reachable at the bottom of the Add sheet without being obscured by the native bottom navigation. Because Add is designed to stay compact, the first implementation should prefer an in-sheet bottom action over an absolute footer that can collide with the tab bar.
- Add must preserve in-flight draft state for the current app session and selected inventory context. Navigating away from and back to Add must not clear the name, description, selected parent, selected photos, details disclosure, or recently created/selected parent unless the user successfully saves or explicitly clears the draft. Draft state must not bleed across tenants, inventories, principals, or service compositions.
- Add draft persistence must live behind a mobile application draft-store port supplied by bootstrap composition. UI modules must not use module-level singleton draft maps for preservation.
- Add must expose an explicit `Clear draft` action inside secondary details/actions so users can intentionally discard in-flight work without making the primary capture path feel destructive or cluttered.
- After a successful save, Add should keep the most recently created or selected parent preselected for the next asset so batch entry into the same box, shelf, room, or container is low-friction.
- The parent picker must be labeled with user-facing placement language such as `Put in`, not implementation language such as `Location` or `parent asset ID`.
- `No parent` must be the label for top-level inventory placement. `Inventory root` must not appear in the Add form.
- The parent picker must search every searchable asset in the selected inventory, including locations, containers, and items. It must not be restricted to locations.
- Parent search results must render inside a collapsed select-menu style control rather than as a full always-visible list. Empty-query results should show a bounded set of recent and likely parents from current inventory context only after the control is opened.
- If a typed parent name has no exact match, Add must show the create-parent affordance directly below the parent input, above `No parent` and all search results, so it is the first available action while the keyboard is open.
- Creating a missing parent from the Add form must be one action. The first implementation may create it as a `location` until the API supports explicit parent-intent selection, but the UI must make the creation affordance visible before the user finishes typing. After creation, the newly created place must be inserted into the open picker results, selected immediately, and acknowledged inline with a `Place created` state so the user can see what happened.
- If the user selects an item as the parent, Add must clearly communicate that Stuff Stash will promote that item into a container for this placement. The create command must perform the parent promotion through application services and persistence unit-of-work behavior rather than client-side mutation so server validation, authorization, audit, and downstream list state remain authoritative.
- The Add form must not fully enumerate every parent candidate in large inventories. Parent candidates must be query-bounded, recent/contextual, and compact.
- Add must support selected-photo management before save: add photos from the device library, preview selected photos, reorder selected photos, remove selected photos, and upload the remaining photos as asset attachments after the asset is created. The first slice may use the existing JSON base64 attachment endpoint and must keep picker/transport details behind application/adapters.
- Mobile photo selection must declare iOS photo-library and camera usage in the Expo config and generated native project before invoking native photo or camera capture. The microphone usage string may be present only to satisfy the Expo image-picker module's native plist validation; the Add flow must not expose video capture or microphone behavior until those behaviors are specified.
- The Add form must provide an iOS keyboard accessory dismissal control for text inputs, including the multiline description field, so the keyboard cannot trap users away from lower sheet controls. Scrolling may also dismiss the keyboard where the platform supports it, but it must not be the only dismissal mechanism.
- The mobile API adapter must map API effective access metadata into mobile inventory role and capability state at the adapter boundary. It must use inventory `access.relationship` for display labels and inventory `access.permissions` for workflow affordances such as whether Add is reachable. The API remains the authorization authority for every state-changing operation.
- Local mobile validation datasets may live under `.stuffstash/seed-data/` and `.stuffstash/seed-media/`, must be ignored by Git, and must seed the in-memory API only through public REST endpoints for tenants, inventories, assets, and attachments.
- Local mobile validation against the in-memory API should disable HTTP rate limiting with `STUFF_STASH_HTTP_RATE_LIMIT_ENABLED=false`; the mobile browsing path may legitimately issue many API and thumbnail requests from one device while the app is being iterated locally.
- Settings must be production-shaped even in the Expo Go validation slice:
  - The Settings root must use a native-feeling grouped list hierarchy instead of presenting each category as an equally weighted card. The root is navigation, not a dashboard, and must keep account, preferences, tenant administration, connection, diagnostics, and about information in distinct groups.
  - Settings categories must open dedicated native stack destinations with disclosure-row semantics. The root must not repeat the Stuff Stash wordmark or a second in-content `Settings` title beneath the native navigation title.
  - It must show the current authenticated principal from the API when available.
  - The account destination must show a human-meaningful principal label and may expose the opaque principal ID only as secondary diagnostic detail. It must expose `Sign out` as a separate operation that clears secure authentication state while preserving the saved server URL and tenant hint for the next sign-in.
  - The appearance destination must expose `System`, `Light`, and `Dark` as one checkmarked single-selection list. `System` must be presented first and remain the default. The list must reflow without truncation at every supported Dynamic Type size.
  - The connection destination must use `Stuff Stash server` as the primary user-facing term while allowing `instance` in explanatory self-hosting copy. It must show the current server URL and expose `Change server` separately from `Sign out`.
  - `Change server` must state that the operation signs the user out and clears the saved server and tenant hint on this device without deleting server data. Its native confirmation must use the explicit action label `Change Server`, not a generic continuation label.
  - It must show About information for Stuff Stash, including the mobile app version.
  - Clearly labeled diagnostics such as the configured API base URL, local-development authentication mode, and opaque principal ID must live in a dedicated Diagnostics destination instead of sharing a group with normal user actions.
  - Voice setup is tenant administration. The Settings root must expose it only when the selected tenant's access metadata grants `configure`; viewers and inventory-only administrators must not be sent to a provider-profile authorization failure.
  - The Voice row and destination must identify the tenant whose shared voice behavior is being configured. The root may show a compact readiness value such as `Ready` or `Needs attention`, but loading voice readiness must not block the rest of Settings.
  - Settings and every destination must expose a direct retry action for recoverable first-load failures. Root Settings must not use pull-to-refresh for device-local preferences and mostly static account or diagnostic information.
  - Settings section titles must expose heading semantics, navigable rows must have disclosure semantics and context-rich accessibility labels, and all controls must retain at least 44-by-44-point targets. Single-line labels, values, choices, and actions must be vertically centered within their row targets; multiline content may grow the row while preserving balanced vertical padding.
  - Settings layouts must remain comfortably readable from the smallest supported content size through `accessibility-extra-extra-extra-large`. Horizontal controls, label-value pairs, status pills, and action groups must stack at accessibility sizes rather than splitting words, clipping primary labels, or forcing multicolumn text.
  - It must not duplicate the Home inventory switcher, expose unfinished account controls, or show unavailable feature panels. Inventory-scoped administration rows may identify the selected inventory when their behavior is specified below.
- Settings must expose a voice provider profile readiness surface before voice capture depends on tenant-managed profiles:
  - The surface must load safe tenant-scoped provider profile metadata through a mobile application port and generated API-client adapter.
  - It must show capability, provider kind, display name, lifecycle state, credential status, model name when present, last-tested state, and safe prompt-template presence for language inference profiles.
  - It may run the API's safe provider diagnostic probe for a selected profile and show only the safe status, message, and tested-at timestamp returned by the API.
  - It must never show raw credentials, sealed credential data, provider account details, provider-specific realtime URLs, raw prompts, raw transcripts, raw model responses, raw audio, or generated speech.
  - Mobile may create first-pass recommended profiles for the existing API-supported provider contracts so a tenant can be configured from the phone:
    - Gemini API-key speech-to-text using capability `speech_to_text`, provider kind `gemini`, model `gemini-2.5-flash-lite`, and credential purpose `api_key`.
    - Gemini API-key language inference using capability `language_inference`, provider kind `gemini`, model `gemini-2.5-flash-lite`, credential purpose `api_key`, and optional prompt-template editing.
    - Google Cloud Text-to-Speech using capability `text_to_speech`, provider kind `gemini`, runtime options for `languageCode` and `voiceName`, and credential purpose `oauth_bearer` until an API-key-backed speech synthesis adapter is specified.
  - Mobile profile creation must keep advanced provider fields conservative and explicit. It must not ask for provider secrets until after the profile record exists, because credential replacement is a separate API operation that seals raw material server-side.
  - Mobile credential entry must send the raw credential only to the credential replacement command, keep it in component state only for the active edit session, clear it after completion or cancellation, and never persist, log, echo, or include it in diagnostics.
  - Mobile may enable, disable, or archive profiles through explicit user actions. Archive remains a destructive action that requires native confirmation.
  - Arbitrary endpoint editing, runtime option editing beyond the first recommended profile controls, and custom provider-kind creation remain future mobile work.
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

## Inventory Sharing And Invitation Links

- Settings must expose `Sharing` only when the selected inventory grants `share`; direct deep links without permission must show a calm denied state.
- Sharing must use a dedicated native stack destination, not an inline Settings card or modal nested in the root list.
- The destination must list safe invitation metadata and support viewer/editor invitation creation through a mobile-owned application port.
- A created invitation must show the complete link once with `Copy link` and native `Share invitation` actions, expiration context, and a warning that the link cannot be recovered after leaving.
- Incoming `/invitations/accept` routes must remain outside the tab hierarchy and survive connection onboarding and native OIDC sign-in.
- The acceptance destination must follow the state, security, accessibility, theme, and post-accept navigation requirements in `specs/identity-access/mobile-oidc-authentication.spec.md`.

## Verification

- `pnpm --dir apps/mobile check` must type-check the mobile scaffold.
- `pnpm --dir apps/mobile test` must run focused mobile application tests.
- The Expo development server should start with `pnpm --dir apps/mobile start`.
- iPhone verification is manual: install Expo Go, run the mobile dev server, scan the QR code, and confirm the Stuff Stash home screen appears.
