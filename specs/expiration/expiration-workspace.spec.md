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
