# Comprehensive mobile UI audit and remediation

## Scope and completion

The user authorizes a long-running audit and remediation of every mobile surface,
including iOS, declared iPad support, and Android implementations. Expand route
inventory into nested tasks, shared controls, lifecycle states, system entrypoints,
and accessibility/adaptation axes. Maintain an explicit surface-by-axis ledger in
`docs/reports/mobile-ui-remediation-2026-09-14/`.

Each cell records pending, source-reviewed, runtime-verified, finding, or justified
N/A. Track findings through spec, failing test, fix, review, runtime evidence and
release. Source and native-runtime evidence are separate; pending cells prevent an
unqualified claim that the audit is complete. Unavailable device access must be
recorded with attempted access paths, without stopping independent fixes.

Use the platform-interaction review policy and UI design skill. Preserve accepted
Home Add/Notifications/Profile order, Home-only inventory switching, transparent
scroll-edge behavior, and explicitly requested bottom filter completion actions.
Do not change backend/auth boundaries for UI convenience. Android uses its own
native interaction adapters. No unrelated web remediation is part of this scope.

## First remediation pass: short value choices

- Browse Type, Status, Availability and Sort select a value in place using a
  native menu-style picker, within the existing filter draft. Selecting an option
  must not navigate away from Filters. Cancel discards draft; Show results applies.
- Expiration Kind and Availability follow the same pattern. Large/multiple/searchable
  choices retain selection views; do not replace tags or location hierarchy with
  an unbounded menu.
- Reuse the platform choice adapter. Current value, label, disabled and selected
  states must be accessible. Android must use the existing native menu adapter
  rather than the generic expanding custom selection fallback.
- Outer scope-change and asynchronous cancellation protections remain intact.
- During search, show the effective relevance order instead of an inactive saved
  sort. Choice adapters share empty-option semantics: optional empty selection is
  offered once unless the caller supplies its own empty value or disables it.

## Verification

Test selection without navigation, draft/apply/cancel behavior, disabled options,
searchable-choice navigation, platform adapter selected-state behavior, and all
shared consumers. Run checks remotely or in CI under the session constraint against
local builds/tests. Each implementation pass requires the code critic. Release
mobile fixes through TestFlight with changelog under existing authorization.

## Native audit runner

Use macOS GitHub runners with the release workflow's pinned Node, pnpm, Xcode and
CocoaPods versions. Build the committed native application for the simulator in
Release configuration, without signing credentials. An ephemeral XCTest UI target
may be added to the runner checkout; it must not enter the distribution project.
Capture XCTest result bundles and screenshots on success and failure, identified
by source revision and simulator. Begin with genuine unauthenticated onboarding
entry, help, keyboard and reachable completion controls on phone and iPad.

This initial smoke test establishes runner access, not full application coverage.
Authenticated task coverage requires synthetic data through existing ports or a
controlled test backend; no production authentication bypass, user inventory data,
or signing secret may be used to make UI tests convenient. Expand scenarios and
record actual runtime evidence in the surface ledger as the audit proceeds.

Workflow timeouts bound failed builds. Use scheduled sleep intervals while waiting
for GitHub jobs; do not cancel a healthy build because it is slow.

## Pull refresh lifecycle

All list refresh controls represent an explicit pull gesture, not background query
invalidation, polling, pagination or focus reconciliation. Reuse Home's focused
pull lifecycle across inventory assets, location assets, locations, sharing and
expiration, plus existing gesture-driven Browse, Map, detail/history, notifications
and customization collections. Clear presentation on navigation blur and ignore duplicate pulls and
late completion from a previous focus session. Keep query access/error handling
unchanged. Background updates must remain functional without shifting the list or
presenting a pull spinner. Prevent direct query-activity binding mechanically.

## Custom field choices

New custom field Type and Applies to choices use the shared native menu picker,
with the existing immutable edit values retained. A picker must not publish changes
after its containing form becomes read-only, including an option opened before
permissions change. Retain outer field draft/save semantics and validate selected
values against the supplied options.

## Inventory switcher

Present the inventory switcher as a bounded, scrollable native sheet with a native
Close action available during loading, error and ready states. Determine the
current household by tenant identity, never its display name. Preserve the
household/inventory hierarchy and current selection. Empty households explain that
no inventories are available. Prevent duplicate selection requests, report a failed
switch in place, and retain the sheet for retry. Dismiss only after a successful
selection; suppress late navigation after the sheet has unmounted.

## Multiple tag selection

Expiration tag choices expose independent checked states and checkbox semantics.
Users can select and remove multiple tags in the filter draft without replacing
the other selections. Type and location remain single-choice controls.

## Location choice identity

Expiration location choices display and search the known containment path so
same-named places in different rooms are distinguishable. Carry parent ancestry
from the already authorized active inventory tree through the location summary
and view model; do not fetch unrelated inventories or expose unavailable ancestors.
Indicate a partial path when the supplied tree cannot resolve the full ancestry.
Keep filtering by stable location ID and preserve draft/apply behavior.

## Notice replacement lifecycle

Each global notice owns its timer and animation lifecycle. Replacing a notice
starts a fresh display interval. A delayed dismissal from a replaced notice must
not remove its successor. Use stable monotonic notice identity and clean up the
old timer on replacement.

## Accessible feedback

Global notices with actions, warnings or errors remain until dismissed or replaced.
Plain informational/success notices may expire after the normal display interval,
but remain while a screen reader is enabled. Read accessibility preferences
conservatively and respond to changes during display. With Reduce Motion enabled,
show and dismiss without sliding or spring motion, including gesture recovery.
The dismiss control's accessible name contains the actual message. Actions have
explicit labels and at least 48-point targets; enlarged text stacks the action
below the message. On iOS, announce a new message without moving focus; Android
uses the live region. Retain the shared nonmodal banner pattern because the native
alert dialog would interrupt ordinary saved-state/undo feedback.
