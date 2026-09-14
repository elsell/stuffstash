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
