# Expiration workspace

## Approved scope (2026-09-12)

The product owner approved the complete recommended mobile and web expiration
experience, including Home preview, full date-grouped workspace, search and
filters, Browse integration, validation, deployment and TestFlight release.
This approval replaces a separate visual direction gate for this scope. Use the
existing native and web components and verify realistic responsive states.
Calendar day grids and additional primary tabs remain deferred. This is inventory
review, independent of notification delivery and read state.

## Query and calendar contract

- Scope every read by authenticated principal, tenant and selected inventory.
  Return only assets and placement metadata the principal may view. Counts must
  use identical authorization and filters to results; never disclose hidden assets.
- Default to active assets with recorded dates. Missing dates do not imply safety.
  Archived assets are excluded. Retained dates on disabled types appear in All
  dates with tracking off, but are excluded from Expiring soon/Expired modes.
- Modes: `soon` (existing personal upcoming state), `expired` (existing expired
  state), `all` (all recorded dates, including retained disabled tracking dates).
  Soon excludes expired items and uses existing per-type/inventory advance days.
  Delivery switches and notification read state never hide inventory facts.
- Preserve day/month precision. Month dates mean the end of the entered month;
  do not display or save an invented exact day. Use existing personal saved
  calendar timezone and injected clock; travel does not change that timezone.
- Support text, custom type, assigned tags, containment location (including
  descendants), and inclusive optional expiration date range filters. Date range
  filters compare the last valid local calendar day of the stored expiration;
  this evaluation must not change its displayed precision. Validate strict day
  range bounds and reject reversed ranges. Text matches item title/description
  through existing search semantics where practical; label its scope accurately.
- Soon and future dated assets sort earliest expiration first. Expired groups
  sort most recently expired first. All dates places expired dates first followed
  by future dates. Stable asset ID resolves ties. Pagination is server-side and
  must not filter only a downloaded client page. Bind cursors to principal/scope,
  normalized filters and evaluation calendar context; reject mismatched cursors.
- Return authoritative scoped counts for soon/expired/all together with a bounded
  page or summary. Keep queries bounded, cancellation-aware and observable through
  ports. Reuse existing expiration descriptions and access policies, with no
  per-card API requests. Failure is not a zero count.
- Read audit follows existing inventory read contracts. New HTTP surfaces require
  adversarial boundary tests before implementation and generated client contracts.

## Home

- Add compact Expiration section beneath inventory context, ahead of Recently
  changed when attention items exist. Use shared photo/name/placement rows.
- Show labeled Expired and Expiring soon counts and at most three preview rows.
  Include both groups when both contain records, avoiding expired backlog hiding
  approaching dates. Tapping a count opens its mode; See all opens All dates.
- With dates but no attention items, retain a compact quiet entry to All dates.
  No dated assets must not create a large onboarding panel. Keep access through
  Browse. Section failures offer retry without blocking the rest of Home.
- Counts are independent of unread notification badge and push permission.

## Complete workspace and navigation

- Title: Expiration. Native hierarchical stack route inside current navigation
  context; never force a tab switch. Shared asset detail opens from a row, and
  Back preserves mode, filters, loaded position and scroll. Web supports ordinary
  links, reloadable scoped routes and browser back/forward.
- Visible modes: Expiring soon, Expired, All dates. Group rows chronologically
  by month and year, with an explicit expired grouping in All dates. Preserve
  exact date and month-only labels; show Expires today through its final day.
- Search and filters support type, tags, location and explicit date range. Active
  filter state is visible and clearable; no matches differs from an empty inventory.
  Filters are a bounded apply/cancel task using existing sheet/form patterns.
- Browse gains the same expiration modes and date ordering, composed with its
  existing filters. Preserve ordinary browse behavior when expiration is unset.
  Use the same query semantics and shared client application/domain presentation.
- No new completion/disposal domain or ambiguous Done/Dismiss item action. Use
  existing authorized detail edit/archive operations and undo.
- Mobile supports full-viewport refresh for empty/short lists, pagination retry,
  system typography, 44-point targets, large text and VoiceOver. Web supports
  keyboard navigation, focus restoration, responsive rows and accessible controls.
  Never communicate state only by color. Long names and paths remain readable.
- Preserve prior data on refresh failure, indicate stale/error state, and prevent
  late requests from replacing another scope or newer filters. Scope caches by
  server/principal/tenant/inventory. Refresh on mutations, preference changes,
  foreground/reconnect and the next relevant calendar boundary.
- Initial loading, initial failure, append failure, empty dates, filtered empty,
  permission loss and tracking-off states must be explicit and testable.

## Delivery and acceptance

Implement through domain-oriented application queries and repository ports,
generated REST clients, and separate frontend adapters/domain models. No new
unpinned dependency. Keep files split by responsibility and run required critic.
Use meaningful failing tests, including multi-page ordering/count consistency,
month precision, timezone boundary, disabled tracking, opposite reminder overrides,
wrong principal/tenant/inventory, denied asset visibility and stale requests.
Exercise realistic small/dense data and narrow/desktop layouts. Native renderer
checks do not establish physical-device acceptance. Run tests remotely/CI, release
through normal atomic PRs and signed TestFlight workflow, apply pinned API/web
images through GitOps and record exact evidence. Notification delivery was confirmed
by the user; tapping repair needs a released native build and user acceptance.

### Initial query execution strategy

Use the existing scoped asset repository with an explicit dated-only read filter.
Scan keyset batches of at most 256 active dated assets, applying personal calendar
state and combined filters in the expiration application package. Keep at most the
requested result page plus lookahead in the sorted selection, and count the full
matching stream. This avoids duplicating personal policy/calendar logic in database
expressions. Bound a request to 1,000 batches and fail visibly if exhausted; never
return partial counts as complete. This is an operational guard, not an inventory
size promise. Add a supporting scoped dated-asset index where measurement warrants.
The acknowledged tradeoff is O(dated assets) count evaluation per request; measure
with dense fixtures before release. All production reads remain cancellation-aware.

### Browse composition

Browse offers Expiration as a date-ordered refinement. Opening it preserves the
submitted text, tags, asset kind and availability filters and the prior Browse
route for Back. Expiration always reviews active assets; label that restriction
when entering from archived/all lifecycle Browse. The full shared expiration
workspace owns its additional type, location and date-range refinements. Kind and
availability remain editable there and clear with other filters. The API accepts
optional `kind` and `checkoutState` and applies them before counts and pagination,
using batch checkout reads. Ordinary Browse sorting and lifecycle are restored on
Back. Date ordering supersedes relevance/updated-time ordering in this refinement.

### Validation and release evidence

- Direct APNs response repair released as v0.23.3 (83.1), workflow 34693391123;
  signing and TestFlight upload succeeded. Physical notification-tap acceptance
  remains separate from upload evidence.
- Expiration HTTP tests cover owner/viewer access, outsiders, unauthenticated and
  malformed tokens, tenant/inventory mismatch, combined filtering, disabled
  tracking, availability transitions, archive, cursor principal/filter/settings
  binding and complete multi-page counts. A 4,097-dated-item in-memory HTTP fixture
  returned exact counts and a 30-item page in 37 ms on the remote validation host.
  This measures the application strategy, not production Postgres latency.
- [Final CI](https://github.com/elsell/stuffstash/actions/runs/34696475443)
  passed all six jobs: required checks, web image, self-host runtime, browser
  journey, iOS dependency lock and PostgreSQL search benchmark. The required suite
  passed API tests, 1,279 mobile tests, 1,119 web tests and 67 SDK tests. Remote
  complete web checks and desktop/Pixel-7-sized browser acceptance passed after
  the final shared-control and visual-token corrections. The browser journey
  covers Home, all dates, pagination, shared detail/Back, filters and reload.
- [PR #101](https://github.com/elsell/stuffstash/pull/101) merged as
  `f7463fdb1e1be12a7c49a8bd29b2f2906a75df95` after required critic review.
  [v0.24.0](https://github.com/elsell/stuffstash/releases/tag/v0.24.0) publishes
  the exact API and web images. Infrastructure commit
  `5afd902c668d6b40a4a4d60109235c89e05b5119` pins them; Flux reported that
  revision Ready and Healthy, with both deployments 1/1 ready. Live `/healthz`
  returned healthy, the new expiration contract was present, the unauthenticated
  expiration endpoint returned 401, and the web root returned 200.
- Signed native delivery for 0.24.0 (84.1) is tracked by the
  [release workflow](https://github.com/elsell/stuffstash/actions/runs/34696764927).
  Its upload step is the delivery evidence; Apple processing and physical-device
  notification-tap/layout acceptance are separate.

## Native control refinement (2026-09-12 device feedback)

- Prefer actual platform controls and familiar native patterns over custom text
  actions. Keep the existing system segmented control for mutually exclusive
  expiration modes, with the native menu fallback at accessibility text sizes.
- Use the native navigation search bar (UISearchController on iOS), including
  system clear/cancel and keyboard Search. Search changes update results after
  a 300ms pause; clear/cancel and keyboard Search apply immediately. Preserve
  route query when returning from detail or filters; cancel pending work on
  unmount and do not allow stale search callbacks to undo newer route filters.
- Put Filters in the native navigation toolbar using
  `line.3.horizontal.decrease.circle` (filled when refinements are active), with
  an accessible label describing active filters. Search text alone does not
  mark the filter button active. Reuse the existing filter form sheet.
- Home status shortcuts use standard full-width disclosure rows with neutral
  labels, trailing numeric counts and chevrons, reusing the shared selection row.
  Whole rows open the corresponding mode; See all still opens All dates.
- Remove the separate Search submit link, in-content Filters link and redundant
  ordering sentence. Put the segmented control in the list header so the list
  is the native inset-adjusted scroll surface. Keep full-height pull to refresh,
  keyboard dismissal, empty/error states and shared asset navigation.
- Validate live-search/cancel/restoration, native header action semantics, and
  Home disclosure navigation with controlled renderer tests. Run remote checks,
  critic and CI before a signed TestFlight release. Device screenshots remain
  distinct from renderer verification.

### Filter-sheet device correction

Use a full-height native form sheet with a system title. Put Apply filters and
Cancel (or Back within a choice page) in a persistent bottom safe-area action
area using platform-native buttons; never place custom touch targets over the
sheet's top drag region. The scroll area has a single standard horizontal inset,
and the footer remains reachable above the keyboard. Choice searches use the
native navigation search bar. Date bounds use system switches for optional bounds
and compact native date pickers, avoiding a permanently expanded inline calendar.
Reset remains a standard grouped action. Applying a reversed range is disabled;
Cancel/dismiss never applies staged changes. Back retains the staged selections.
