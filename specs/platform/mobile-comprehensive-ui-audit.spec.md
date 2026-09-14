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

The native audit may additionally build a runner-only fixture application. Its
route root must be installed only in an ephemeral GitHub Actions checkout, with
an explicit fixture suite selection and an untouched copy of the production
routes retained. The ordinary onboarding suite still uses the production root.
Fixture routes compose production UI components with synthetic props and in-memory
ports, using the real Expo Router stack, native control adapters, theme, keyboard,
and feedback providers. They must not import production credentials or connect to
an inventory service. No fixture entrypoint, authentication exception, or runtime
fixture flag may enter the distribution bundle. Include fixture sources in
TypeScript checks and record the suite separately in native artifacts.

Begin fixture runtime coverage with Browse/expiration filter sheets and feedback;
verify menu selection stays in place, bottom actions remain hittable, draft values
reach the caller, and date-range and tag pages retain accessible return actions.
This is component/native integration evidence, not authenticated end-to-end coverage.

Workflow timeouts bound failed builds. Use scheduled sleep intervals while waiting
for GitHub jobs; do not cancel a healthy build because it is slow.

Native keyboard checks must use the platform's actual dismissal gesture: iOS
interactive dismissal drags downward from scroll content above the keyboard.
An upward whole-application swipe is not evidence of a broken dismissal. Assert
the entered field value as well as keyboard presence so input loss cannot pass
unnoticed. Preserve failed-run artifacts and distinguish test-procedure failures
from confirmed application defects.

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

## Custom-field target drafts

Apply the saved-versus-draft target distinction from client settings management.
Keep persisted targets immutable; allow unsaved additions to be deselected in both
create and edit. Preserve permissions and scoped eligible-type loading.

## Native fixture findings

Short menu choices must retain a visible field label beside the current value.
On iOS use SwiftUI LabeledContent around the native menu picker when it is outside
a native Form; a centered value without its field name is not an acceptable row.
Accessibility may combine the native field name with its explicit action label;
runtime tests should locate that semantic label without requiring an exact
platform-generated concatenation.

Expiration filter sheets use a bounded View/ScrollView arrangement like Browse.
Their text search belongs to the native navigation bar and date entry uses native
date controls, so a root KeyboardAvoidingView is unnecessary. Keep automatic
keyboard insets on the scroller and verify actual sheet content plus bottom
actions on phone and iPad. Native run 34887652455 showed a blank body with the
previous keyboard-avoiding root; treat the replacement as unverified until rerun.


The native fixture suite also compares controlled and uncontrolled text entry
using the same project text-input primitive and URL keyboard. This is a diagnostic
for M14; neither a fixture pass nor an upstream issue alone establishes the cause
of lost onboarding characters. Keep the production full-address assertion intact.

M19 diagnostics retain the production full-height filter fixture and add the same
screen at medium/large detents in a separate runner-only route. A diagnostic pass
must not replace or conceal failure at the production presentation. Export native
hierarchy alongside screenshots to compare scroll and footer geometry.

Run 34903318947 completed the detent comparison: medium/large showed the
iPhone body while full-only did not. The next candidate shares the medium/large
production filter configuration with fixtures and additionally checks expansion.
Prior full-only failure remains recorded in the audit evidence.

Native settings-control fixtures exercise the shared Appearance menu, direct
iOS color picker opening/closing and explicit color clearing, and compact asset
expiration date entry/clearing. These isolate native controls; parent editor
cancel/save and actual Add-sheet navigation require their own runtime scenarios.

An isolated onboarding fixture uses the real screen and command with in-memory
API/auth/profile ports to assert the complete native-entered address reaches the
authentication command. It must not contact an external service. Production-root
onboarding continues to test real layout and keyboard behavior independently.

The Add fixture must mount the real Add screen at its production full-height
form-sheet presentation with controlled in-memory query and command ports. Check
native header visibility, complete typed draft submission, disabled Save/Close
while pending, failure retention, and dismissal recovery. A rejected fake save
is intentional and must never create an asset in an external inventory.

Keyboard-dismissal audit gestures must originate inside the actual native
ScrollView bounds and above the keyboard/accessory, rather than infer the scroll
edge from the keyboard position. Export hierarchy to verify those bounds. A
corrected gesture must still demonstrate dismissal; changing coordinates alone
does not resolve the finding.

Native typing scenarios must wait for the keyboard and a hittable key before
injecting text. Preserve exact displayed and submitted value assertions; do not
mask lost characters with replacement typing or weaker matching. Calendar
popover dismissal must target outside its bounds and verify dismissal before
checking the underlying action's reachability. A failed test gesture is not
evidence that a product control is unreachable.

The form-sheet diagnostic compares a direct ScrollView, the same ScrollView
inside a background/flex container, and that container with the existing native
action footer. All use identical synthetic rows and full-height native sheet
presentation. This fixture isolates layout integration; it must not replace the
production filter or count as filter completion. Require visible, hittable rows
and export native hierarchy for each independently executed variant.

Native fixture selectors must use observed accessibility labels, including the
system color picker's lowercase close label. The onboarding command-submission
fixture dismisses the keyboard through its explicit accessory before submission;
keyboard-obscured action reachability remains a separate layout finding and may
not be considered fixed by changing the test sequence.

Native audit artifacts must retain xcodebuild standard error as well as standard
output. After a run, collect StuffStash-named diagnostic reports from the ephemeral
runner and recent simulator logs restricted to the StuffStash process. Collection
runs after failures and must not change a failed test result to success or prevent
already available screenshots/results from being uploaded. These synthetic,
credential-free audit jobs retain artifacts for the existing fourteen-day window.
