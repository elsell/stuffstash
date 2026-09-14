# Mobile Navigation And Visual Polish Spec

## Purpose

Define the native iOS interaction contract for the Stuff Stash mobile product
after the initial tracer-bullet and voice slices. This spec resolves the
boundary between navigation, temporary tasks, global Voice interaction, and
gesture-driven containment exploration.

## Decisions

- Home and Browse remain the only primary tab destinations.
- Voice remains a prominent persistent bottom accessory. Its visual prominence
  is intentional, but it is a global interaction layer rather than a product
  destination.
- Voice may present one native detent sheet over the current surface. Voice
  must never push an asset, Settings, provider, or other product route while
  its sheet is presented. A linked result must dismiss Voice first and then
  navigate from the underlying surface.
- Map remains a Browse sub-surface with horizontal containment columns,
  breadcrumbs, branch gestures, and progressive path navigation. The gesture
  model is product behavior, not disposable prototype behavior.
- Map must not embed a full asset workspace in a modal. Selecting an asset's
  information action dismisses any transient Map selection surface and opens
  the shared asset workspace as a native stack destination.
- Asset detail, child assets, parent assets, Places, History, Settings, and
  provider-profile detail are hierarchical content and must use native stack
  destinations with standard back navigation.
- Add, Edit, Move, Move Here, Filters, Tenant Switcher, and other bounded
  commit-or-cancel tasks may use native `formSheet` routes. They must not use a
  custom React Native `Modal` for their primary presentation.
- Full-screen photo viewing is the sole full-screen modal exception. It must
  not contain product navigation or another sheet.
- Native platform menus are required for overflow, sort, and short contextual
  action lists. Alerts are reserved for destructive confirmation, permission
  blockers, and authentication/session loss.
- In-component `Modal` is prohibited for product navigation and bounded task
  surfaces. Any exception must be named in this spec and covered by a focused
  structural test.

## Presentation matrix

| Surface | Presentation | May push product routes? | Exception |
| --- | --- | --- | --- |
| Voice | Native detent sheet | No | None |
| Map asset selection | Shared asset stack route | Yes, through normal stack | No embedded workspace |
| Asset detail | Native stack route | Yes | None |
| Add/Edit/Move | Native form sheet | No; finish or dismiss first | None |
| Filters/Switcher | Native form sheet | No | None |
| History/Settings | Native stack route | Yes | None |
| Photo viewer | Full-screen modal | No | Image viewing only |

## Voice behavior

- The Voice accessory remains the strongest secondary control in the bottom
  navigation area and may use a large action target, contextual state, and
  progress indicator.
- Prominence must come from size, placement, and state—not from making Voice a
  competing tab or leaving the tab bar visually ambiguous.
- Opening a grounded asset reference from Voice must preserve the completed
  session state for return, dismiss the Voice sheet, and open the shared asset
  route from the underlying navigation context.
- Opening Voice setup from a recovery state must dismiss Voice before opening
  Settings.
- Parent selection during Voice review must be a native task surface that
  returns to the same review state. It must not create a modal stack.

## Map behavior

- Branch swipes, horizontal column paging, breadcrumbs, and reduced-motion
  behavior remain supported.
- The selected branch and search result must be visually distinct from an
  asset selected for detail.
- Map asset detail uses the shared asset workspace. Map may provide a compact
  native preview only if it contains no edit, move, history, checkout, photo,
  or nested asset navigation controls.
- Map gestures must not steal vertical scrolling from the active column and
  must provide an accessible button path for every gesture-only action.
- Search may expand to a matched path, but must also expose the matched asset
  and its current containment path in user-facing copy.

## Visual and accessibility contract

- Native system typography and Dynamic Type remain the default. Fixed widths,
  forced one-line truncation, and fixed card heights must not hide title,
  placement, status, or primary action information.
- The primary hierarchy is title, placement, state, then supporting metadata.
  Tags and technical classification are subordinate.
- A location path must use a compact readable summary at the current surface;
  it must not auto-scroll away the leading context in a way that hides the
  location root.
- Add and edit forms must keep their primary commit action visible above the
  keyboard and must use keyboard-aware scrolling to the focused field.
- The Add form's parent picker must keep its search field visible while typing
  and provide an independently scrollable results area when the keyboard
  reduces the available sheet height. Creating a new parent must remain
  reachable without relying on content hidden behind the keyboard.
- The Add form must use the available form-sheet height intentionally: its
  header, context, photo affordance, name, parent, details, and commit action
  should read as one compact task rather than a sparse stack surrounded by
  unexplained blank space.
- Every interactive control must have a 44-point minimum target and a
  non-color selected, disabled, loading, and error state.
- Every transient surface must provide an explicit title, dismissal action,
  safe-area handling, focus transfer, and restoration of focus/context on
  dismissal.
- User-facing copy must use domain language (`Name`, `Location`, `Container`,
  `Item`, and `Place`) and must not expose future-work or implementation
  limitation messages.

## Enforcement

- Structural tests must reject product `Modal` usage in navigation and screen
  files except the named full-screen photo viewer and the explicitly reviewed
  system picker exceptions.
- Structural tests must reject Voice route pushes to product destinations and
  Map detail modals containing `AssetDetailView`.
- Mounted behavior tests must prove Voice dismissal before linked navigation,
  Map-to-asset stack navigation, keyboard-safe Add parent selection, and
  preserved Map gesture path state.
- Visual verification must cover regular and inline Voice accessory placement,
  compact and expanded sheets, keyboard-open forms, large Dynamic Type, dark
  appearance, and the Map gesture states on a real native build in CI or on a
  device. This repository must not require a local native build on a
  disk-constrained development host.

## Browse native search (2026-09-13)

- Browse has a nested native stack inside its existing tab, retaining the `/search`
  route and tab identity. A system navigation search bar serves both list and map;
  remove custom in-content search boxes and duplicate Browse titles.
- Use compact native navigation search (updated 2026-09-14), system
  clear/cancel and keyboard Search, no focus on entry. Keep list debounce and
  authorized result/filter/pagination semantics. Map search still finds an item
  and opens its containment path; update it after a short typing pause as well as
  explicit keyboard submission. Native callbacks must submit their event text.
- Carry current text across list/map switches, settling pending list search before
  the switch so delayed callbacks cannot restore the old surface. Existing list
  refinements and map path state remain intact. Clearing search clears map highlight
  without discarding the current map path. Restore route text in the native field.
- A shared native navigation-search adapter owns native clear/cancel/field sync;
  domain search behavior remains in the corresponding screen/application query.
  Do not add dependencies or change backend/security contracts.
- Validate field restoration, clear/submit semantics, switch races, map path
  behavior and existing Browse regressions remotely/CI; run critic before merge
  and follow the signed TestFlight workflow through upload.

### Compact Browse controls (2026-09-14)

- Browse list and map use native integrated-button search at the trailing edge
  of the navigation bar, expanding on interaction. Disable toolbar integration
  so iPhone does not move search into the bottom toolbar. iOS before 26 uses the
  library's native inline fallback; Android retains its native search action.
- Keep existing query restoration, clear/cancel, submit and debounce semantics.
  Search results remain described by the result summary when search is inactive.
- Remove the standalone Expiration pill from the result header: it is navigation,
  not an applied filter. Within expanded Browse Filters, expose standard disclosure
  rows for Expiring soon, Expired and All dates in an Expiration section. Explain
  that these review active items by expiration date. Preserve existing query and
  applied refinement handoff to the expiration workspace.
- Entering expiration from Filters uses the currently displayed draft selections
  and latest search text; cancel pending debounce before navigation. Browse's
  applied filters remain unchanged when returning, while its latest query is
  settled into the route and result state.

### Map navigation-header clearance (2026-09-14)

- Native search makes the iOS header translucent. Browse List continues using
  automatic scroll insets; Map must instead offset its entire fixed header and
  column layout by the native navigation header's measured height.
- Read the reactive header height through @react-navigation/elements 2.9.20,
  promoting the already locked navigation dependency to an explicit dependency.
  Do not hard-code status/search-bar heights. Android's opaque header already
  positions content below itself and receives no additional top padding.
- Map columns and horizontal breadcrumbs must not each apply automatic navigation
  insets. Keep existing bottom dock clearance and horizontal gestures unchanged.

## Native Home and Browse actions (2026-09-14)

- Browse moves Add out of its List/Map content rows into the native trailing
  navigation bar beside compact system search. Keep the List/Map control in
  content. Hide Add when inventory permissions disallow creation; preserve the
  existing Add route and cancel any pending map search before opening it.
- Home gains its own native stack within the existing Home tab, preserving its
  root route through a pathless (home) route group. Keep the inventory dropdown's title/subtitle and switch behavior
  as a custom leading header item. Bound its width and truncate visually while
  keeping full accessible context. Do not add a duplicate Home heading.
- Notifications, Add and Account become trailing native bar items. UIKit owns
  symbols, grouping, backgrounds and hit targets. Use the real iOS 26 bar-item
  badge for unread counts; retain full spoken count and loading/error labels.
  Pre-26 systems retain the accessible count even when native badges are absent.
- Preserve notification registration, inventory-scoped query identity, polling
  and retry-on-open behavior. Present native toolbar data without duplicating
  queries or introducing notification state into domain services.
- Reuse a focused platform header-action adapter. iOS uses UIBarButtonItem via
  native-stack items. Android's native-stack binding lacks that item API, so
  place native Compose IconButtons in its headerRight slot. Non-mobile renderers
  use accessible React Native controls for previews and tests.
- Home content uses automatic scroll insets with no duplicate top safe-area
  padding. Preserve Browse's measured map-header inset. Native bar options must
  update on permission, notification count and scope changes.

## Browse filter sheet (2026-09-14)

- Replace Browse's inline expanded filter panel with a native form sheet,
  initially about 70% height and expandable to full height. Keep a scrollable
  overview of Type, Status, Availability, Tags, Sort and Expiration summary rows.
  Use existing settings disclosure/checkmark patterns; show selections as row
  context and drill into one choice category at a time.
- Reset all changes only the sheet draft. Show results is a prominent full-width
  native bottom action; Cancel/dismiss discards the draft and Back returns from a
  choice page without losing choices. Share the native sheet-action adapter with
  Expiration. Sort lives in this sheet instead of a second Browse toolbar icon;
  indicate Relevance and disable explicit ordering while text/tags drive search.
- Tags use native navigation search and multi-select checkmarks; other categories
  use single-selection rows. Expiration offers its existing date-mode destinations
  using current draft and query. Keep the explanation that expiration reviews
  active assets, even when Browse lifecycle includes archived assets.
- Pass canonical Browse route state plus tenant/inventory identity into the sheet.
  Settle pending search before opening. Applying explicitly replaces every filter
  parameter, including empty/default values, so Reset cannot retain old URL state.
- Resolve tag choices through scoped server-query adapters. Reject wrong inventory,
  tenant or session targets before and after loading, and revalidate scope before
  Apply or expiration navigation. Cross-scope sheets must not display stale choices
  or apply to a newly selected inventory.
- Cancel, native dismissal and scope replacement abort pending Apply verification;
  a late response must not navigate after the sheet closes. Serialize submissions
  synchronously so repeated taps cannot produce duplicate navigation.

## Native tab-header scroll appearance (2026-09-14)

- Home and Browse navigation bars float over scrolling content on iOS. Remove
  opaque custom bar backgrounds and separators: on iOS 26 use the native soft top
  scroll-edge effect, with no additional blur layer; older iOS uses native system
  material blur. Do not simulate scrolling transparency with JS opacity or gradients.
- Preserve system automatic content insets on Home and Browse List, and measured
  header clearance around Browse Map's fixed controls. Android keeps its opaque
  native navigation bar and existing layout.
- This corrects scroll appearance, not navigation placement. Preserve current
  Home inventory selector, permission-aware actions, notifications and compact search.
- Follow Apple's TN3106 navigation-bar appearance guidance and the pinned native
  stack adapter's scrollEdgeEffects API. Do not combine headerBlurEffect with the
  iOS 26 native scroll-edge material.

## Inventory switching remains on Home (2026-09-14)

- Keep the inventory switcher only on Home. Browse returns to its native Browse
  title, Add and compact Search, without a leading inventory control or repeated
  inventory row in content. Clear any prior leading header item explicitly.
- Preserve the native transparent scroll-edge appearance and compact Browse
  controls. Existing scoped queries, permission checks and result error recovery
  remain responsible for loading the selected inventory.

## Home refresh and header fit (2026-09-14)

- Home's native pull-to-refresh control represents only a user-started pull.
  Background dashboard/expiration refetches and post-edit refreshes must not open
  or hold the pull control. Blur/unmount clears its presentation; late completion
  cannot clear a newer pull. Ignore late native pull events while Home is not
  focused. Keep ordinary scroll position when returning.
- Keep refreshing dashboard and expiration data together for an explicit pull;
  retain existing query/error recovery and do not cancel shared background reads
  just to reset the control. Suppress duplicate pulls while one is active.
- Reserve space for Home's native action buttons before sizing its inventory
  selector, including native group/margin allowance. Use explicit 44-point iOS
  bar-item widths. Prioritize Add and Account before Notifications in native item
  order so the primary actions remain available at constrained widths.
- Bound selector width rather than crowding buttons; truncate its text while
  retaining full accessible inventory/tenant names. Add symmetric 12-point inner
  horizontal padding and compact vertical padding. Use a single label at enlarged
  text sizes instead of compressing two lines into the navigation bar.
