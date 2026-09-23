# Comprehensive mobile UI audit and remediation

## Investigation and acceptance discipline

Prioritize confirmed user-visible defects at normal text size. Keep one current
diagnosis per active defect with established facts, excluded causes, implementation
decision, remaining budget and acceptance gates; link durable evidence instead of
duplicating checkpoint narratives. Default investigation budget: one source review
and two native experiments. Every experiment must distinguish named competing
causes and state what each outcome would change. Exhaustion requires a concrete
implementation/adapter decision or documented external limitation, not another
equivalent run. Budgets persist across task continuations.

Current text-entry and color investigations have exhausted that diagnostic budget.
No further broad provider/accessory/pacing/scrolling comparisons are authorized by
the audit plan. Prioritize concrete corrections and unchanged user-workflow acceptance;
retain completed diagnostic evidence without restarting equivalent runs. Keep diagnostic-only
failures distinct from confirmed shipped-workflow failures.

Validate shared control behavior across relevant states, representative consumers
and critical end-to-end workflows. The3,408 coverage cells do not require3,408 tests.
Differences in composition, lifecycle, security or edit ownership justify extra
consumer checks. Keep spec-first development, meaningful regressions, code critic
review and native verification. Sleeping scripts own terminal-state collection;
manual status polling must not duplicate a healthy collector.

## Move destination text entry

The real Move journey reproduced lost query characters on phone and iPad after
one complete typing attempt. Reuse the proven SwiftUI draft text-field adapter
used by Add on iOS; keep the existing Android text input. Native text owns its
editing buffer so ordinary query updates do not echo stale JavaScript values back
into it. A successful creation can replace the query with the canonical returned
name: advance an explicit draft query revision to reset the native field then.
Pending/denied operations remain noneditable and late changes remain guarded.
Do not change providers, typing speed or keyboard assistance to hide the failure.
Retain exact one-attempt typing and creation/movement recovery acceptance.

## Newly created Move destinations

During a Move form visit, successful destination creation must immediately add
that confirmed destination to the visible matches and select it, even if the
lookup cache still contains an empty result. The matching name/kind/parent must
no longer offer creation. Retain confirmed creations during that form visit,
filter them by the current normalized title query, and merge lookup results by
asset ID with server values taking precedence. Changing the query must not leak
unrelated local matches or erase the selected destination. Failed creation adds
nothing; failed Move retains the selection. Closing the form drops this local
supplement. Backend authorization and duplicate rules remain authoritative.

## Move destination acceptance coverage

Use the real Move route with synthetic core, lookup, create and move ports in the
isolated audit runner. Cover choosing an existing destination, switching to a new
container through the native Kind menu, a rejected create retaining the exact
query/kind, successful creation selecting that destination, rejected Move retaining
that selection, and successful retry returning. The fake must validate exact
creation and movement payloads, so returning alone cannot pass an incorrect move.
Require a hittable query field, keyboard presence, one typing attempt and exact
retained text; do not make enumeration of individual keyboard key hit regions a
prerequisite for this editing journey. Keep this separate from the already accepted
Move-here journey. Normal-size
phone/iPad and Android checks precede enlarged-text work; real persistence and
permission boundaries remain independently tested. No production route is added.

## Invitation native acceptance coverage

The runner-only native suite must render the real invitation acceptance screen
with synthetic preview/acceptance ports and the production Invitation header.
At normal text size on phone and iPad, verify a long inventory name, visible
access, reachable Join and Not now, dismissal without accepting, explicit joining,
and retained accepted access when opening fails. Retrying Open must not accept the
invitation again. Keep production routes untouched outside the isolated runner.
This fixture does not certify external link intake, authentication, authorization,
or physical-device transitions; those retain their separate acceptance boundaries.

## Home within the production tab shell

Standalone header probes do not establish composition with the native tabs and
voice accessory. The isolated runner must also reuse the three production tab
layouts unchanged from its saved route tree, supplying synthetic Home data and a
clearly identified Browse placeholder. Mount them at a distinct `audit-tabs` path
with the same folder depth, avoiding an ambiguous second root index while keeping
relative imports intact. Verify normal-size Home action order,
voice entry reachability, bottom-row clearance, scrolling and Home→Browse→Home
return without a stuck pull indicator on phone and iPad. The placeholder verifies
tab transitions only, not Browse content/search. Do not replace the existing
standalone action probes or claim real audio/provider/device coverage from this
composition fixture. Runner diagnostics must not overlay the content being checked.

## Asset action eligibility and retained drafts

Edit and both Move forms must honor the current core view's edit/move capability,
including read-only access and archived lifecycle, on direct entry and refresh.
When unavailable, retain the draft in the mounted form, disable fields/choices
and mutation commands, and explain why no action can be submitted. Cancel remains
available and keeps existing discard protection. Restored eligibility may resume
the retained draft. Previously captured mutation and draft-change callbacks must
consult current committed eligibility, not the permission captured when rendered.
Already-submitted mutations remain authorized by the server; this UI guard is not
an authorization substitute. Edit must recheck eligibility after asynchronous tag
reconciliation and before submitting the update. Shared route tests must exercise
denied direct entry, revocation, archive, retained draft recovery and stale callbacks.

## Asset action loading and error exits

Edit, Move and Move-here must expose a native Close command while the asset core
is loading or unavailable. These states have no draft to discard. Close returns
to the previous route, or Home when there is no back destination. Retry remains
available after failure; a slow or failed read must not trap a user in an Edit
sheet whose swipe dismissal is disabled. Ready-state draft and operation guards
remain authoritative once editing begins.

## Staged tag removal in Edit

New staged tag chips in Edit perform removal, unlike existing-tag choices that
toggle assignment. Announce the action as Remove new tag followed by its name;
do not describe this one-way draft removal as a selected toggle. Preserve other
staged tags, assigned tags and the item's edited fields. Apply the same command
semantics already used by Add; native assistive-technology output remains a
separate verification requirement.

## Add draft photo removal

Add's photo rail must place a native removal command below each preview, outside
the image and its preview/reorder hit region. Name commands by visible photo
position (Remove photo 1, etc.) and update those positions after removal/reorder.
Removing a draft photo retains the other photos and unfinished item fields. The
command shares the draft-operation lock, including while a system picker is open.
Do not use the existing 28-point overlay as the removal target. Verify native
phone/iPad layout and activation separately from mounted behavior tests.

## Native search placement comparison

Run350465 phone captures show a bottom search field despite integratedButton and
allowToolbarIntegration=false. Retain the production Search-button acceptance
assertions; do not accept the bottom field as equivalent to requested placement.
Add a runner-only minimal comparison with identical search options supplied at
route registration before presentation. It must capture idle geometry and verify
the search button lies in the navigation bar before expanding. This distinguishes
baseline native configuration from the more complex production mounting path;
it does not prove a cause or replace production search/result/navigation tests.

## Onboarding command controls

Connect/sign-in, household/inventory creation and sign-out/start-over reuse the
shared native command adapter. Primary submission is prominent; start-over is a
standard command. Keep command labels visible while pending and show separately
labeled progress. Preserve required-value readiness, keyboard submission,
synchronous duplicate locking, partial-setup recovery and safe late completion.
Native phone/iPad keyboard and reachability acceptance must be rerun after this
migration; prior custom-button captures cannot certify the new control.

## Push setup command names

When successful device setup changes the action to Open device settings, its
accessible name must identify that new action too. It must not continue announcing
setup while opening Settings. Activation must open Settings without re-registering
the device or changing inventory preferences; launch failure/retry retains the
existing visit-owned recovery behavior.

## Scope and completion

The user authorizes a long-running audit and remediation of every mobile surface,
including iOS, declared iPad support, and Android implementations. Expand route
inventory into nested tasks, shared controls, lifecycle states, system entrypoints,
and accessibility/adaptation axes. Maintain an explicit surface-by-axis ledger in
`docs/reports/mobile-ui-remediation-2026-09-14/`.

Each cell records pending, source-reviewed, runtime-partial, runtime-verified,
finding, or justified N/A. Runtime-partial identifies observed scenarios without
claiming the entire axis is verified; a reproduced unresolved defect is a finding. Track findings through spec, failing test, fix, review, runtime evidence and
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
selection; suppress late navigation after the sheet has unmounted. Household rows
show the count of inventories belonging to that household, using “1 inventory”
and “0 inventories” or plural counts as appropriate.

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

Services-gate integration must be exercisable without replacing platform modules
or React hooks. Keep native composition construction in AppServicesContext and
extract the actual mounted gate/feedback wiring behind injected onboarding,
profile-store and composition-factory dependencies. The gate owns startup,
completion, sign-out, server change and session-expiry transitions. Controlled
ports must verify notice/action invalidation through those real transitions and
that notice context changes do not repeat startup or composition construction.
This extraction preserves existing authentication and push-cleanup ordering.

Global notices belong to a services context. The app's feedback provider receives
the current ready composition's service-scope identity, or a disconnected identity
while loading/onboarding. Changing identity immediately hides old notices and
invalidates their action callbacks and retained notice publishers, including an
old publisher used after returning to the same identity. A scope transition must
not clear a newly published notice from the new context. Unmount also invalidates
captured actions. Ordinary same-context navigation preserves completion notices
and View/Undo. Native blocking dialogs remain a separate interaction; this notice
ownership rule must not restart onboarding or rebuild services when context changes.

Each pending invitation exposes a native contextual actions menu, with a named
destructive Cancel invitation command instead of an ambiguous X. Cancellation is
infrequent and irreversible; retain the confirmation naming the recipient before
submitting. Each scope/invitation has an independent pending lock and visible
Cancelling… status. Duplicate confirmations must not submit twice. One completion
must not re-enable a different pending invitation, and failed rows retain retry.
Confirmations captured before leaving the focused scope must
not start a new cancellation after departure. Already-authorized operations may
finish and update their scoped cache.

The isolated native audit must exercise the actual Sharing screen at normal text
size with controlled invitation and link-action ports. Cover unavailable creation
link recovery, retained email, retry to a usable link, copy failure/retry, and
cancellation failure/retry. Assert feedback stays below the native header and
fully within the scroll viewport, commands can be reached, and Back remains usable.
Retain screenshots and accessibility hierarchies. Fixture actions must not send
invitations, copy secrets to the system clipboard, or open a real share destination.

Sharing creation and one-time-link commands use the shared native command adapter:
primary emphasis for Create Invitation, ordinary text commands for Copy link and
Share invitation. Keep the system share sheet as the destination. While a link
copy/share operation is pending, both link commands are disabled and duplicate
activation is ignored; show a textual pending label. Completion re-enables the
commands without losing the link. A new creation resets link-operation ownership,
and a late result cannot unlock or annotate a newer operation. Creation retains
its existing draft lock and command guard. Native appearance and keyboard
reachability require current-build verification; using the adapter is not proof.

Sharing's link copy result and link copy/share failures belong beside the current
one-time link, inside its scroll content. Cancellation failures belong beside the
affected invitation with its retry command retained. These task-owned messages
must not cover native navigation with a global banner. Clear link feedback when
another link action or invitation creation begins; clear cancellation feedback
when that cancellation is retried. Clear both on focus/scope change, and reject
late results from a previous focused session or replaced link. Keep the system
share sheet and destructive cancellation confirmation. Other global notice
consumers require their own positioning review; this does not certify them.

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

The native settings fixture must verify reminder mode selection in place: choose
Custom, observe its timing controls and saved mode, then restore Use defaults
and observe inherited mode without a navigation transition.

Keyboard readiness must find an actual hittable key, not assume the first
accessibility key is interactive: UIKit also exposes zero-size padding keys.
Keep full displayed/submitted text assertions after this readiness condition.

Native draft-photo fixtures use bundled synthetic imagery and the production photo
strip. Verify numbered removal sends the intended photo identity, independent
horizontal scrolling, Add callback, and read-only preview retention. These fixtures
must not invoke a camera, library, upload, or real conversation service.

Use Apple's XCTest accessibility audit on named visible states to check hit
regions, labels, traits, contrast, Dynamic Type support and clipping. Preserve
all reported issues; any future suppression needs a specific documented false
positive. Record viewport/device/build scope and do not interpret a visible-state
audit as coverage of offscreen controls or the complete platform.
See https://developer.apple.com/documentation/accessibility/performing-accessibility-audits-for-your-app.

The runner-only Add fixture may catch React render errors and display their
message and component stack as diagnostic UI, so a fixture composition failure
can be distinguished from an application/native crash. This must remain outside
production routes and must not replace a failing assertion with success.

After the phone direct-scroll diagnostic passed and both wrapped variants lost
the body, compare two bottom-action structures before changing production filters:
a direct scroll view with a sibling overlaid footer, and a direct scroll view with
an in-content footer. Both must retain visible rows and reachable bottom actions
on phone and tablet. A successful isolated layout still requires production
expansion, keyboard and long-list verification.

For repeated character loss, compare the existing controlled/uncontrolled React
Native input fixtures with a SwiftUI TextField from the already-pinned Expo UI
adapter. Keep the same URL keyboard, disabled correction/capitalization and XCTest
whole-string typing. Assert both displayed native value and observed callback
value. Preserve the production onboarding assertion; a diagnostic comparison
cannot establish a production fix or justify silently slowing/shortening input.

Native assertions following a React state transition wait for the expected
accessible state within a bounded timeout. For read-only photo previews, wait for
Add/Remove disappearance before asserting retained images; do not infer failure
from an immediate stale snapshot or skip the final-state assertions.

### Native query readiness diagnostics

The Add and checkout-history native fixtures must expose query readiness when
synthetic repositories do not reach their expected screen state. Record only
fixture query keys, status/fetch status, presence of data, observer counts, online
and focus state. Never expose query data, errors containing arbitrary responses,
authentication values or production configuration. Keep diagnostics runner-only;
retain existing readiness and interaction assertions. The diagnostic component
must not alter connectivity, pre-seed the cache, bypass the scoped query path or
turn a loading screen into a passing acceptance result.

### Onboarding address input candidate

After run34917318548 preserved the entire URL through the SwiftUI comparison on
both phone and tablet while production onboarding still lost characters, use the
already-pinned Expo SwiftUI TextField for the iOS server-address field. This is a
single-line URL entry task, not a search or navigation task. Keep the visible field
label, URL keyboard, disabled autocorrection/capitalization, native Go submission,
and pending disabled state. Native text owns its editing value; React receives
changes for validation and submission without replacing text on each keystroke.
Help/error renders must preserve the draft; leaving and restarting setup remounts
an empty field. Android retains its existing platform text-input path. Other
onboarding fields are outside this candidate until evidence requires changing them.

The production full-address typing and submission assertions remain unchanged.
Source tests do not establish that native typing, keyboard avoidance, focus or
large text works; require the complete production onboarding and submission
scenarios on phone and tablet before closing M14.

The shared iOS keyboard accessory must dismiss the actual native first responder,
including both React Native and SwiftUI text fields. Use the already-pinned keyboard
controller's native responder dismissal, which does not depend on React Native's
focused-input registry. Preserve the explicit accessory command and its no-submit,
no-edit semantics for every consuming form. Retain the fixture assertion that taps
Dismiss keyboard and waits for keyboard disappearance, in addition to scroll and
Go submission scenarios.

When XCTest's debug hierarchy truncates the query-readiness accessibility value,
attach that diagnostic element's explicit value as a separate text attachment.
Use a stable test identifier rather than relying on a particular native element
type. When React Native exposes nested text elements without the outer value,
retain the same safe snapshot in the diagnostic's accessibility label as a
fallback. Choose a complete JSON object so a truncated value cannot hide a
complete label. Keep the visible short label and geometry unchanged.
Do not widen the UI or render raw query data to compensate. Keep the existing
safe diagnostic schema, screen layout and failing acceptance assertions intact.

If tapping the accessible color-picker row fails to open the system picker, retain
that assertion and add an independent comparison targeting the visible trailing
color well using the captured native element bounds. The iPad run34919776387
hierarchy exposes one 704-by-36-point button spanning label and well, while its
screenshot shows the well at the trailing edge. Do not infer a successful picker
from a tap or replace the failed row-activation result with the comparison's result.

## Native choice-label text adaptation

The shared iOS menu picker must preserve the full visible field label when
Dynamic Type or a narrow available width requires wrapping. Keep the native
LabeledContent/Picker interaction and allow the label's intrinsic vertical size
to increase; do not shrink the text, truncate it, or hide the label to make room.
This applies to every shared picker consumer, including filter Availability,
settings, customization fields and reminder mode.

Apple's Typography guidance recommends adapting layout for larger text, including
stacking where needed: https://developer.apple.com/design/human-interface-guidelines/typography.
Run34920888328's phone accessibility audit identifies Availability as potentially
clipped at larger sizes. Preserve that failing native audit and add an explicit
accessibility-size label/menu scenario; source tests cannot verify text layout.

## Native choice events during locked editing

The shared choice adapters must reject selection events while disabled, even
if the native menu was already open when the parent locked editing. Native visual
disabling is not a substitute for guarding the callback boundary. When editing
resumes, valid selection events must be delivered normally. This preserves the
parent's draft during pending saves; it does not change application authorization.

## iPad onboarding drag diagnosis

Run34923022927 reaches native iPad portrait UI, preserves the full address, and
leaves Connect visible, but the left-margin downward drag does not dismiss the
keyboard. Retain this scenario. Add an independent drag inside the centered form
column to distinguish gesture-region behavior from a general dismissal failure.
Record both outcomes and screenshots; a passing comparison does not erase the
original failure or certify all onboarding keyboard behavior.

### Native dependent-query comparison

Keep the failing production Add scenario and add an independent runner-only cold
inventory-query comparison. It uses the production provider/hook and a fresh cache
with local deterministic inventory/resource ports. Verify an initial scoped query
and a second query enabled by its result become visible without user interaction.
This isolates query readiness from Add draft/navigation composition. Retain cache
readiness diagnostics; a comparison pass does not certify or replace Add acceptance.

### Add presentation comparison

Keep the original Add form-sheet scenario and introduce a runner-only navigation
card route exporting the identical Add fixture. Run the same full-string draft,
rejected Save, pending commands and Close assertions through both routes. No
preseeded cache or query bypass is permitted. Compare presentation/entry before
changing production query behavior; a passing comparison does not clear the
original failing Add scenario. Generated routes remain isolated from release builds.

### Accessibility issue attribution

Native accessibility audits must retain a per-issue description, optional element
hierarchy and screenshot when XCTest identifies an issue. A missing element is
reported as unavailable, not treated as a false positive. The diagnostic handler
must return false so XCTest continues to report the issue as a failure; it must
not suppress an audit category. This distinguishes broad audit predictions from
explicit enlarged-text interaction results and avoids attributing an unidentified
clipping issue to whichever control was fixed most recently.

### Add sheet header configuration comparison

The native Add audit must compare the existing full-height sheet with the same Add
fixture in a full-height sheet whose native header is declared visible with its
known title before presentation. This isolates changing header visibility during
presentation from the existing card-versus-sheet comparison. Preserve the original
routes and complete draft/typing/rejected-save assertions. The comparison uses cold
query state and the same application component; it must not preload resources.
Fixture preparation must remain isolated from production routes.

Run34965113594 on iPad reaches the focused form in the preconfigured-header sheet,
while the hidden-header sheet remains loading with zero query observers. Apply the
preconfigured native header and known Add item title to production Add before
presentation. Keep the same full-height form sheet and the screen-owned Close,
Save and busy/dismissal rules. Preserve the hidden-header diagnostic for comparison.
This is a readiness candidate based on native evidence, not proof that typing or
rejected-save recovery passes: the configured-header case separately fails exact
text retention. Current phone/iPad acceptance remains required.

### M20 — full-width onboarding scroll content

Run34932076384 on iPad repeats the outer-margin dismissal failure while the
inside-column comparison passes. Preserve the centered 600-point form width,
but apply that constraint to a child form rather than the scroll content itself.
The scroll content must fill the viewport so blank margins participate in the
same native scroll/keyboard gesture surface. Keep native keyboard dismissal;
do not add a competing tap catcher or manual gesture recognizer. Both existing
native drag cases, phone keyboard reachability, and iPad landscape layout remain
acceptance gates. The failing native drag is the regression baseline; source
layout and unit checks alone cannot establish its resolution.

### Onboarding start-over completion ownership

A pending Sign out and start over command may finish its authorized teardown after
its screen unmounts. Its former screen must not then clear local fields or invoke
onStartOver/onStateChange callbacks that could replace the new destination. Use
the same mounted-generation ownership as Connect/Create completion. While the
initiating screen remains mounted, successful reset still returns to connection;
failed reset preserves the current form and reports a retryable error. This is a
UI completion guard, not cancellation of sign-out or profile cleanup.

### Onboarding required-field readiness

Connect/Create is unavailable while a required value is blank or whitespace-only.
Connection requires a server address; household setup requires both names; first
inventory requires its name. Show a concise visible explanation of the missing
value rather than an unexplained disabled action. Keep entered values and enable
the action as soon as required text is present. URL syntax and server validation
remain at the application boundary on submission, so malformed nonempty input
still receives a specific error. Apply readiness to both the primary button and
keyboard submission; pending operations retain their existing duplicate guard.
Start over remains available independently of missing names. This uses native
text entry and an ordinary command state; it does not introduce another screen.

### Checkout-history readability comparison

Keep the existing StaticText hit-testing scenario as a diagnostic. Run349289 iPad
shows the first checkout note visibly inside the sheet while its StaticText
isHittable fails; tappability alone is not a readability oracle. Add a comparison
that checks the complete note bounds inside the sheet scroll viewport below its
navigation bar, captures screenshots for occlusion/readability review, expands,
loads older content using that sheet's scroll view, and closes. Continue requiring
actual hit targets for commands. Geometry is not proof of VoiceOver access,
contrast or unclipped text rendering. Do not clear M61 merely from the comparison
passing or remove the old failing scenario without resolving its evidence.

### Time-zone search vocabulary

The reminder time-zone chooser searches both its readable city/region label and
the underlying IANA identifier, case-insensitively and with outer whitespace
trimmed. Partial identifiers such as America/New must find America/New_York;
requiring a complete valid identifier is not sufficient search behavior. Preserve
the saved choice, native navigation search, empty-result feedback, the bounded
initial list and explicit valid-zone fallback when the runtime list is unavailable.
Typing/searching alone never changes the saved time zone.

### Inventory switcher focus ownership

Selection callbacks belong to the focused visit that rendered them. Retained
callbacks must not start a selection after departure, including after returning
to the same switcher. A fresh visit provides fresh selections. Close must ignore
inactive visits and retire its current visit immediately before navigation.

Selection completion belongs to the switcher's uninterrupted focus session.
When the sheet loses focus, abort its request signal and suppress late navigation
or selection-error feedback, even if it remains mounted or regains focus before
completion. Retain the pending guard until that request settles so a second
selection cannot overlap it. After settlement, a fresh focused selection works.
Do not assume blur reverses an inventory selection already accepted by the port.

### Inventory switcher command controls

Use the existing native command adapter for Switch household/Back and load Retry.
Keep the hierarchical selection behavior and request guards. Its full-width native
Host belongs below the household heading, not squeezed alongside it in a row.
Allow the heading to wrap; do not rely on toolbar-style single-line truncation in
the sheet body. Loading names the inventory task rather than the internal tenant
concept. Verify native narrow/large-text layout before claiming visual acceptance.

### Native inventory switcher fixture

Add a runner-only fixture using the production switcher screen, its production
sheet detents and native command adapters. Use a preloaded synthetic dashboard
for presentation coverage (not cold-query acceptance), two households, and a local
selection port that rejects once then succeeds. Verify household drilldown,
selection failure/retry, return to the fixture menu, and explicit Close. Retain
screenshots. Do not touch real inventory selection or production services.

### History comparison query resolution

Run34937278231 exposes nested StaticText wrappers with the same checkout note
label. The bounds comparison must resolve an explicit first matching note inside
its History scroll view before reading geometry, including the older-page note.
Run34965113594's iPad capture confirms the note is visible while the legacy
StaticText hit-test fails; the scoped bounds/pagination/dismissal case passes.
Retire that duplicate tappability diagnostic and retain the full bounds-based
journey. Static text need not be an interactive command. Continue asserting actual
command hit targets, pagination, expansion and Close on the native runtime.

Native integrated search may collapse after clearing an unfocused search field.
After clearing, require the full result collection to return. Accept the native
collapsed state only when the search field is absent and the Search button is
reachable. Reopen it, enter a fresh query and exercise explicit cancellation while
focused; require the query field and keyboard to disappear, the full collection
to return, and native More/Back to remain reachable. Never require an obsolete
Cancel button after UIKit has already closed search.

### Landscape capture cross-check

The iPad landscape app attachment in run349372 has landscape pixel dimensions
and EXIF orientation8, while its rendered view disagrees with the recorded window
and form geometry. Capture XCUIScreen.main alongside the app attachment for named
landscape checkpoints. Preserve both originals and existing geometry assertions;
this comparison must not transform screenshots or declare layout verified merely
because hit-testing passed.

### Settings recovery commands

Root Settings, detail loading failures and the shared refresh notice must use the
existing NativeCommandButton adapter for Retry and Retry refresh. These issue an
in-place command; they do not choose a value or navigate. Preserve Account and
Connection recovery links, retained data, query retry behavior and error copy.
Place the full-width native Host in vertical content. Inspect every shared-notice
consumer; native error-state layout remains a separate verification requirement.

The same command rule covers scoped household/inventory Settings, customization
collection/editor load retries, the editor's Refresh access command, and provider
and voice setup load retries. Refresh access still rechecks permission through its
existing application flow and retains the read-only draft. The shared refresh
notice also serves Sharing, customization, provider lists/editors, scoped Settings
and voice setup; inspect those vertical placements and preserve their behavior.

### Add save errors remain in the presented form

A failed Add save must show a persistent, announced error inside the Add form,
retain the draft and restore Save and Close. Do not rely on the root notice
overlay for a failure in a native sheet: it can render behind that sheet.
Show the error before form fields and scroll it into view after failure; the
next save or deliberate edit can clear the stale failure. Keep ordinary
background refresh separate from this recovery. Native acceptance must inspect
the visible error, retained text and reachable retry/close, not only AX existence.

The same form-owned error rule applies to Add parent creation, camera and photo-library failures. Use the operation-specific heading; cancellation without an error stays silent. Beginning another operation clears stale failure feedback.

### Onboarding disappearance observation

For onboarding help collapse, use XCTest's dedicated waitForNonExistence
with the existing five-second deadline instead of wrapping an element existence
query in a generic predicate waiter. Record screenshot/hierarchy evidence when
the assertion times out; a later closed snapshot does not prove timely collapse.
Do not count unreached keyboard steps as passed. This is a test observation
change, not a production repair.

### Ordinary-keyboard text-entry comparisons

Runner-only diagnostics must compare seeded single-line and multiline RN inputs
with default keyboard/correction settings, alongside the existing URL and SwiftUI
comparisons. Use identical full-string entry and verify native text plus the
application-observed value. Do not replace failing product scenarios, disable
autocorrection in production, slow typing to obtain a pass, or equate URL-field
success with general text-entry acceptance. Native execution is required.

### Remaining Settings recovery consumers

Expiration filter loading recovery, notification settings recovery, Sharing and
Voice permission-check recovery, invitation retry and invitation pagination must
use the shared native command adapter in their existing vertical containers.
Preserve access decisions, retry destinations, cancellation and pagination
disabling. This is a presentation change, not a change to access policy. Remove
shared custom retry styling only after every consumer has migrated.

### Photo removal failure above the full-screen viewer

A rejected photo removal requires a native acknowledgment alert above the active
viewer, rather than a root notice obscured by the full-screen modal. This
operation-specific failure interrupts because the user needs to know the
confirmed destructive request did not succeed. Retain the photo and enable retry
after completion; acknowledgment must not perform another removal. Suppress
late failures after the route's operation owner is gone. No authorization or
removal service boundary changes. Native modal layering remains an acceptance gate.

Native photo-recovery coverage must compose the real AssetPhotoViewerSheet and
AppFeedback native dialog adapter with a synthetic rejected removal. Verify the
confirmation, failure alert, acknowledgment, retained viewer and retry controls,
then close. This fixture establishes modal layering only; production command
ownership and authorization remain covered by their separate boundary tests.

### Pinned photo-viewer motion patch

Use a pnpm content-hashed patch against react-native-image-viewing0.2.2 rather
than an unreviewed version upgrade. Its chrome animation hook must start with
motion suppressed, honor live Reduce Motion changes, ignore stale initial reads,
stop ongoing movement when reduction becomes enabled, and keep animation values
stable across rerenders. Re-enable optional animation only after an explicit
platform value permits it. Cleanup subscriptions/late reads. Preserve the public
viewer API and zoom controls. Test the installed patched hook through the existing
native fake; source assertions alone cannot prove its behavior. Native zoom and
preference changes remain required. The independent image-load recovery gap M85
is not resolved by this patch.

All workspace installation contexts, including the web container build, must include the pinned patch file before pnpm install. The lockfile must change only the patch declaration and affected dependency identity; retain existing package versions.

### Failed photo loading acceptance

A failed image decode or dimension request must replace loading with a readable
Photo unavailable state and explicit Retry, while keeping Close reachable.
Retry starts a fresh image/dimension attempt, preserves request headers and ignores
late events from older attempts; failure must not auto-loop network requests.
Runner diagnostics use a deliberately missing bundled-file sibling, not external
URLs or user media. Verify error/Retry/Close native reachability; separate controlled
loading tests must prove a new request and successful recovery. A native click
on a still-failing fixture alone does not prove retry semantics.

The pinned viewer dimension adapter accepts an explicit retry generation. Each
source or generation change clears old dimensions and starts a fresh native size
request; obsolete completions cannot affect the active photo. Remove its URI-only
JavaScript dimensions cache so a changed source/header context cannot inherit a
previous request's result. Native image caching remains the platform's concern.
This internal lifecycle repair alone does not establish visible error/retry UI.

Both platform image items must share attempt ownership for dimension and decode
completion. A decode error is terminal for that attempt even if a late load event
arrives. Retry remounts the native image and restarts dimensions without modifying
the source URI or headers. The failure replaces the image's zoom/gesture surface
with a centered, readable message and native Retry button; the viewer's existing
Close and photo navigation stay outside this replacement. Do not add another
screen, confirmation, or automatic retry loop for this recoverable read failure.

The dependency exposes an optional error renderer so Stuff Stash can reuse its
native command-button adapter. Restore viewer chrome when the current photo fails;
background/preloaded photo failures must not interrupt another photo's zoom state.
Keep image source objects stable across unrelated wrapper rerenders.

### Add error reveal beneath native chrome

Run34944106312 shows the inline error heading partly covered by the Add navigation
bar after programmatic scrolling. Reveal the form's top using the measured native
header height on iOS, rather than content offset zero. Keep automatic safe-area
adjustment and allow that bounded negative programmatic offset; the pinned RN
Fabric scrollTo implementation otherwise clamps it using raw contentInset rather
than UIKit's adjustedContentInset. Android retains offset zero. Do not add fixed
header padding or a guessed navigation-bar height. The native rejection scenario
must assert the complete error heading below the navigation bar and inside the
sheet, in addition to retained text and reachable actions. Source tests verify the
scroll command against changing header measurements; native geometry is separate.

References: React Native ScrollView scrollToOverflowEnabled and
contentInsetAdjustmentBehavior (https://reactnative.dev/docs/scrollview), and
UIKit adjustedContentInset
(https://developer.apple.com/documentation/uikit/uiscrollview/adjustedcontentinset).

### Asset command completion ownership

Checkout, return, archive, restore and permanent deletion share the existing
asset-screen operation owner with photo changes. Acquire the synchronous guard
before invoking a command so repeated callbacks cannot submit twice before React
renders the busy state. A confirmation callback captured for another asset must
not start work. After unmounting or replacing the asset screen, completed commands
may finish in the domain, but must not navigate, show screen-owned feedback,
refresh the replacement screen or clear its busy state. Keep explicit deletion
confirmation, domain authorization, and command/audit behavior unchanged.

### Native recovery in asset action sheets

Edit, Move and Move here query failures use the existing native command adapter.
Name the failed resource in Retry labels so simultaneous type/tag failures remain
distinguishable. Metadata failures are compact inline messages alongside the
retained form, not additional expanding full-screen error panels. Keep the blocking
asset-load error separate from supplementary metadata recovery. Manual retry must
not reset dirty fields, selected destinations or staged tags. Preserve queries,
permissions and command behavior; native reachability still requires verification.

A runner-only Edit fixture starts with both metadata queries failed and succeeds
on each explicit retry. At the largest accessibility text size, verify both native
retry commands and Cancel are reachable, then verify the original asset name after
recovery. Local behavior tests separately exercise a dirty draft. The fixture must
remain isolated from production routes and cannot perform a real save.

### Keyboard accessory isolation comparison

The seeded ordinary and URL fixtures both use the same AppTextInput wrapper.
Their different strings/keyboard settings do not isolate that wrapper. Add a
runner-only seeded URL comparison that removes AppKeyboardAccessory before focus,
while retaining the provider, input props, string and typing cadence. Verify the
accessory is absent. This isolates the extender's presence, not the entire keyboard
controller; its pinned native implementation reloads input views when attaching
an accessory, which is a hypothesis to test, not an established cause of lost text.
Do not remove production keyboard controls based only on this source observation.


### Contained-workspace recovery and native-pattern follow-up

The contained-items audit includes the shared detail route reached from Browse
and Map. Unknown contents must not be described as an empty collection.
Initial loading and failed loads without data must suppress empty-result claims
for both contents and photos;
a failed refresh must preserve any available contents. Contents and photo query
errors must have persistent region-level status and independent native retry
commands, rather than requiring a pull gesture or relying on a root notice. This does not change repository authorization or query scope.

The older inline contained-search contract requires a separate spec revision
before adopting native navigation search. That revision must preserve title/path
matching, section counts and no-match recovery, address search ownership on both
entry points, and include native focus/close checks. See M88/M89 and the24-axis
contained-items report; neither finding is considered fixed by this source review.


M89 native acceptance uses the real progressive detail route with controlled first
failures for contents and photos, followed by independent successful retries.
The runner verifies absent empty claims before recovery, reachable native retry
commands at largest accessibility text, independent error removal and Back after
recovery. Current Map info pushes the ordinary asset route; the fixture must not
be described as a Map sheet test. Real Map return preservation remains a separate
acceptance scenario. No fixture mutations or production credentials are used.


M88 search revision adopts NativeNavigationSearch for location contents at the
existing20-row threshold. Route-owned query state clears on asset replacement
or loss of eligibility. The adapter explicitly removes search configuration when
disabled. Existing Browse/Map/Expiration callers retain enabled behavior by default.
Native verification must cover header coexistence, expansion, cancellation and
return, in addition to query-fake tests for filtering and ownership. Spatial
command modernization remains a separate part of M88.

Native place-search acceptance uses the shared detail route with20 known item
rows and successful independent queries. Verify the integrated search button,
scoped field, filtering, native clear, settled full results, keyboard dismissal
and Back. Check the More actions control before and after search. This fixture
does not certify Map-path return, all text sizes, or spatial command styling.

M88 detail commands reuse NativeCommandButton with optional primary prominence.
Default commands retain their existing native text-button appearance. Add item here
uses native primary styling; Move items here and maintenance remain quiet. Direct
item Check out/Return retains primary prominence, while contained-workspace
availability remains quiet. Authorization-derived visibility, missing-handler and
pending disabling, action ordering and route destinations must be preserved.
SwiftUI and Compose own button appearance and label measurement; preview styling
is not evidence of native rendering. Shared adapter consumers need regression
checks, and native large-text/permission-state verification remains required.

Native detail-command acceptance includes an editable container with create/edit
permissions and a long title at largest accessibility text. Check Add item here,
Move items here, Check out and maintenance commands after scrolling: unobstructed
hit targets, minimum44-point height, horizontal containment and primary width.
Capture each visible command region for multiline/visual review. This is layout
acceptance only; downstream mutations and route destinations remain covered by
separate behavior/native journeys. The fixture does not execute mutations.

The Add rejected-draft native assertion must resolve one matching accessibility
element before reading its frame. RN can expose the same rejected-body text on
nested parent/child StaticText nodes; ambiguity is an instrumentation error, not
proof that the error layout failed. Keep the heading-below-navigation and
heading-before-body geometry assertions after resolving the first matching node.

M87 native follow-up: supplementary Edit metadata errors and their retry commands
belong inside the same scrolling form as the editable fields. They must not
consume fixed space above that form or displace its persistent Cancel/Save actions.
Preserve independent retries and the dirty draft. Scroll containment is a source
regression check; native largest-text acceptance must still verify label geometry,
scroll reachability and dismissal. Moving content alone does not establish a fix
for overlapping hosted native labels.

Edit native acceptance may scroll each retry into full view before tapping; it
must not require all metadata to fit simultaneously at accessibility text sizes.
Require the complete retry frame inside the visible form, and Cancel hittable
before and after scrolling/retry. Capture each region for label-overlap review.

### Native command height comparison

A runner-only diagnostic compares the shipping standard native command with an
otherwise equivalent SwiftUI button that requests ideal vertical size on the
outer button as well as its label. Hold label, width, font category and host
configuration constant; use a narrow240-point region and largest accessibility
text. Capture both variants with a following text boundary. This isolates the
outer sizing hypothesis without changing production controls. Completion of the
diagnostic is not acceptance: inspect label bounds and adjacent text before
choosing a repair. Neither variant invokes domain mutations.

### Move-here suggestion recovery

An unknown suggestions collection is not an empty search result. Move here must
show its no-movable-matches state only after a result is available for the current
query. Keep cached candidates on refresh failure and provide retry within the
results scroll region. Retrying preserves the query and selection. Loading/error
status must not occupy a separate fixed region above the form. Native large-text
layout and keyboard/dismissal remain separate verification requirements.

### Native asset-form completion controls

Edit, Move and Move here use NativeSheetActions for their persistent completion
and Cancel controls. Their KeyboardAvoidingView owns keyboard overlap; pass the
container ownership mode to the iOS adapter. Extend the shared adapter with an
optional secondaryDisabled flag, default false, so mutations disable both actions
without changing filter forms where an invalid primary still permits dismissal.
Guard callbacks as well as platform disabled state. Keep native primary emphasis,
content-driven heights and full-width stacked actions. Verify footer reachability
at narrow/large-text sizes; stacking is not proof the whole sheet fits.

A runner-only Move-here recovery fixture fails the first lookup for each search
query, then returns one known movable asset after explicit retry. Exercise the
production sheet at largest accessibility text: enter a query, observe the local
error without an empty-match claim, retry to the actual candidate, retain the
query and dismiss through Cancel. No move is performed. Require scrolling rather
than assuming every result fits; failure to reach controls remains an audit
finding, not a reason to reduce the text size.

Region-recovery native captures must reveal each retry and its settled empty
state completely inside the visible detail scroll region, below navigation.
Existence or a partly hittable control does not establish readable layout. Capture
photo and contents recovery separately; preserve positive settled-state waits and
independent-query assertions. Screenshot evidence must not certify offscreen text.

### Destination creation requires known suggestions

The Move form must not infer that a destination is new from an unavailable
suggestions collection. Hide inline creation until current-query results exist;
keep retry/loading status within the results region. Preserve the query, selected
destination and existing move action during lookup failure. Cached results may
still support the existing same-kind/title/parent duplicate check; this is not a
global uniqueness guarantee. Explicit retry can restore creation after a known
empty result. Native layout acceptance remains required.

### Move form reflow

Move and Move here must place title, help, placement preview, query, status and
selection content in one scrolling form. Keep only the native completion actions
outside this scroll region. Do not cap the results to280points while surrounding
text remains fixed; available space must adapt to sheet height, keyboard and text
size. Keep one keyboard-avoidance owner and preserve search/selection behavior.
Source containment tests verify ownership; native checks verify actual reachability.

### Edit tag-name validation

When inline tag resolution rejects a name as too long, explain how to recover
next to the name field instead of only disabling Add tag. Use concise user copy
without exposing byte-count implementation details. Keep the typed value, color,
selected tags and asset draft; clear the message when the name is valid. Use the
existing application resolver as the validation authority. Empty untouched input
does not need an error. Existing color validation remains with the color picker.

### Provider task completion ownership

Provider creation, credential replacement, prompt guidance, connection tests and
lifecycle commands belong to the focused profile task that initiated them. Leaving
and returning creates a new presentation session. Authorized mutations may finish,
but their old completions must not navigate, publish notices or refresh the new
screen. Keep synchronous duplicate guards and busy input protection until the
pending command settles. A captured archive confirmation must not start a command
after the initiating task loses focus or changes identity. Preserve domain/cache
mutation observers, credential secrecy, draft retention on failure, and existing
focused success behavior. Reuse one provider presentation-session hook across these
consumers; do not cancel an authorized mutation merely because navigation changed.

Successful credential replacement must still clear the submitted secret in its
own keyed form after blur. Focus controls presentation, not secret cleanup. A
replacement profile/form must retain its independently owned input.

The same visit boundary applies to voice-stage service selection, connection test
and enable commands. Key it by the current provider query scope and capability,
so replacing a tenant or stage invalidates earlier presentation callbacks.
Preserve the in-place service picker, synchronous pending guard and scoped mutation
observer. Do not show success/errors or explicitly reload a replacement stage
from a departed operation; focused success and retry must continue to work.

### Provider replacement form commands and removal

Credential and prompt editors use the existing native header Save adapter, with
normal navigation Back rather than duplicate in-form Cancel/Save buttons. Save is
available only for nonblank required text, except server ADC which needs no secret.
Both rendering and command entry enforce this readiness. Keep a local announced
failure near the input, retain failed drafts, and clear stale errors when editing
or retrying. Preserve scoped success notices for the return destination.

Use the native navigation removal guard for dirty drafts and pending saves. A
pending save blocks removal; an unsaved draft offers Keep Editing and Discard in a
native alert. Confirmations are single-use and owned by the focused keyed form.
Successful save disarms removal protection before invoking the return callback.
A late confirmation or save from a replaced form/visit cannot leave the new task.
Secrets remain transient and successful replacement clears the submitted value.

## Provider editor native acceptance fixtures

Runner-only provider fixtures must render the production credential and prompt
screens through real application commands and controlled repository ports. Use
synthetic replacement input only; do not load a session or contact a provider.
The first replacement fails locally and the next succeeds. Native journeys at
normal text size must verify disabled empty Save, typing, Back/Keep Editing,
retained input after failure, retry and successful navigation. A separate dirty
Discard journey must verify return without saving. Keep phone/iPad keyboard,
header and navigation evidence distinct from fixture installation or type checks.

## Shared notice native placement regression

A runner-only notice fixture must use the production AppFeedbackProvider and real
native navigation, both pushed and presented as a full-height sheet. Show a
persistent synthetic notice with an action and retain its screenshot/hierarchy.
The notice's full bounds must be below the native navigation bar and within the
visible application bounds; Back, notice action and dismissal remain reachable.
Executing the action must update a visible fixture result exactly once. These
checks intentionally expose M103 before a placement repair. Do not add a guessed
header offset or replace the nonmodal handoff with an alert to satisfy the test.
The sheet fixture supplies a native Close action; the pushed screen uses Back.
Measure the full animated notice using a stable test identifier without changing
its accessibility grouping, and compare with the active content viewport rather
than the entire iPad window. Wait for presentation geometry before asserting bounds.

## Notice presentation in the active screen

Keep notice data and action ownership in the service-scoped provider, but render
ready-app notices in the focused leaf screen's native content layer. Root and
nested navigation containers must not create duplicate presenters. Preserve the
screen's existing direct scroll child: add the presenter as a sibling rather than
introducing a flex wrapper that changes native sheet measurement. Header overlap
must follow the navigator's actual measured header height when transparent;
ordinary headers already reserve their content area. Headerless screens use safe
area insets. Disconnected onboarding retains its root presenter. Do not guess a
constant navigation height or move domain action ownership to individual routes.
Mounted acceptance must prove one presenter, background hiding, surviving
same-service navigation and action ownership, and live measured-header updates.
Native push/sheet regressions and Home/Browse transparent-header captures remain
required before claiming M103 closed.
Notice lifetime and accessibility announcement are shared with its data: changing
focused presenters must not restart expiry or reannounce the same notice. Entry
animation occurs on its first presentation only. Actionable/warning/error notices
and screen-reader feedback retain their existing persistence policy.

## Reminder completion belongs to the initiating visit

Reminder mode, timing and time-zone controls must suppress result feedback and
completion navigation from a save initiated in an earlier focused visit, including
blur followed by return before settlement. Keep the save lock until that operation
settles; a late completion must not unlock another operation. Preserve draft and
retry behavior for a failure in the current visit and allow a fresh action after
return. Reuse the focused-task presentation primitive across provider and reminder
controls; component-only owners reset on unmount, and command/resource owners also
reset when their identity changes. This is presentation ownership, not cancellation
or reversal of an authorized server mutation.
A departed save must not leave an optimistic value appearing saved without recovery.
When it settles, mode/timing controls reconcile to the latest known saved policy,
even if a parent refresh returns unchanged values. Current-visit failures keep the
attempted selection and explicit retry/discard. Do not discard edits made after an
operation unlocks or use a stale callback's captured policy for reconciliation.

### Footer-appearance diagnostic body sizing

The normal-size footer diagnostic must allocate remaining sheet height explicitly
to its scroll body before auditing native button appearance. Run349725's iPad
capture shows footer commands but no body controls; its failure is missing body,
not established dark contrast. Match the explicit flex scroll sizing used by the
passing footer-layout diagnostic, preserving the same production footer adapter.
Re-run the native entry/appearance scenario; source layout changes cannot certify
that the body or contrast is fixed. Preserve the failing capture.

### Native audit job evidence budget

Allow90 minutes for each macOS native-audit job, including dependency setup,
compilation, the complete fixture suite and evidence export. Run349725's phone
job exceeded the previous60-minute budget while its49-test suite and artifact
upload completed, leaving a cancelled job conclusion despite retained results.
Keep individual XCTest interaction waits bounded and retain failures. This changes
job capacity, not acceptance criteria; do not cancel or restart existing live runs
when changing the workflow budget.

### Input-comparison fixture reachability

Native input comparisons must scroll toward the actual field position, whether
it is above or below the viewport. Run349725's phone system-address fixture shows
the input below the viewport while the test repeatedly scrolls toward the top;
its pre-typing failure cannot diagnose SwiftUI or React Native text fidelity.
Use one bounded geometry-aware reveal for ordinary and address comparisons.
Require the complete field inside the visible scroll area and hittable before
focusing; keep keyboard readiness, typing speed and full-string assertions intact.
Preserve this baseline as a test-procedure failure, distinct from actual malformed
product text captured after typing.

The Expiration native search journey must open the integrated Search button before
typing, record its collapsed state, and assert the full entered query and filtered
choices alongside existing keyboard/footer reachability. Do not retain the old
assumption that the search field is permanently visible after M117.

Run349789 still rendered no footer-diagnostic body after explicit flex sizing.
The next diagnostic uses NativeFilterSheet's direct scroll body and measured
opaque footer, retaining NativeSheetActions through that shared composition.
Measure commands against the actual footer region rather than an enclosing root
View. This removes the diagnostic's nested-scroll difference from working native
filter layouts; it does not establish button contrast until runtime reaches it.

Run349789's iPad Return-details assertion read a partial value immediately after
typeText, but its final screenshot and hierarchy both show the complete requested
value. That journey must wait at most5 seconds for the exact value before failing,
without retyping, changing keyboard speed, or accepting a prefix. Keep subsequent
failed-save/retry assertions unchanged. This evidence does not excuse Sharing's
separately retained truncated email, nor establish Return acceptance before rerun.

Native tag clearance verification must locate the tag by its accessible identifier
without assuming checkbox rows have the Button element type. It must scroll the
sheet containing that tag, not the background route's first scroll view. Preserve
full-row/footer bounds, action reachability, selection and applied-ID assertions.

Text-entry comparison fixtures must replace the fixture menu with a dedicated
scroll page when opened. Inserting them earlier in the long menu can leave their
native field offscreen and confound text-entry evidence with menu scroll position.
Keep actual controls, ownership modes, keyboard options and exact full-speed
entry assertions unchanged; preserve an explicit return action.

Native fixture navigation must preserve production's explicit `Back` label. The
fixture menu title is diagnostic content and must not become a long inherited
Back label that competes with search and detail actions. Run34992079258 phone
place search shows a bottom field instead of the expected integrated button;
its fixture uses `Native UI audit` as Back text while production uses `Back`.
Correct that configuration drift without relaxing the integrated-button assertion.
This is a controlled fixture correction, not proof that Back width caused the
placement or that production search has passed native acceptance.

Expiration type-change confirmation belongs to the focused editor visit and the
exact asset, draft, type definitions and enabled state that opened it. A retained
confirmation after blur/refocus, replacement, settings refresh, draft change,
disabling or unmount must not replace current edits or clear their expiration.
A valid confirmation preserves other draft fields and applies once. Reuse the
existing presentation-ownership primitive; no new navigation or dialog style.

Stored-photo removal confirmation is valid only for the current focused viewer,
selected photo, collection and removal availability. Closing/reopening the viewer,
changing selection/collection, losing remove access, starting removal, leaving the
screen or unmounting invalidates retained confirmation callbacks. Valid acceptance
calls the existing removal command once for the confirmed photo. Preserve the
existing destructive native alert and server-side authorization boundary.

Account sign-out and server-change confirmations belong to their settings query
and focused visit. Retained acceptance after leaving and returning must not start
a session-changing action. A confirmation may start its command once. If a
started command fails after departure, it must release its pending lock without
publishing an error into the new visit. Current-visit failures remain retryable.
This does not change session termination, persistence or authorization behavior.

Account identity recovery must distinguish initial failure from failed refresh.
When no principal has loaded, describe unavailable account details without claiming
cached values are displayed. A failed refresh with retained principal data may
explain that previous details remain. Retry and sign-out stay available in either
case; loaded-provider and other retained-settings recovery keeps its existing copy.

Account and Connection commands use NativeCommandButton within their existing
settings groups. The native command's accessible name matches its visible Sign Out
or Change Server label (and pending label); the adjacent selectable value row
continues to identify the account/server, and the confirmation repeats the subject.
Do not wrap native buttons in a second accessibility control to reproduce the old
custom row label. Preserve disabled pending behavior, confirmations and retry.

Runner-only Account/Connection acceptance composes the production settings screens
with fake principal/diagnostic ports and a session action that rejects once, then
returns to the fixture menu. It must never sign out or change a real server.
At normal text size verify visible command bounds, native Cancel without mutation,
first-confirmation error with retry enabled, second confirmation returning to the
menu, and Back reachability. This provides control/recovery evidence, not proof of
real authentication-provider/session teardown or VoiceOver operation.

Query-provider lifecycle acceptance must include effect replay with mounted query
consumers, not only measurement collectors. A replayed setup/cleanup/setup must
leave active observers attached and reads usable, while actual replacement and
unmount still cancel and clear the departed client. Native run349983 Add remained
Loading inventory with idle zero-observer queries; effect replay is a hypothesis
to test, not an established cause of that native capture.


### Non-crashing keyboard accessory comparison

Run349983 proves the removal-based comparison aborts at native input focus:
KeyboardExtender reads the first child of an empty content view. Preserve this
failure as evidence; do not infer a production typing cause. For the runner-only
comparison, retain the existing accessory and its content, disabling attachment
through the pinned native extender's enabled property before focus. The shared
adapter defaults to enabled for every production consumer. Disabled state must
also hide its accessibility controls and intercept no touches. Keep the original
full-string and observed-value assertions, and require no visible Dismiss keyboard
command. This compares attachment enabled versus disabled, not the presence of
the native observer or the entire keyboard controller. Native execution must prove
the comparison no longer crashes and reaches typing before drawing conclusions.


### Draft-photo confirmation ownership

Add's photo removal alert belongs to the visible draft photo selection and current
navigation visit. A retained confirmation must not remove photos or close/change
the viewer after selection, collection, close/reopen, navigation, busy-state change, or unmount.
A current confirmation executes once, retaining the existing next-photo selection
or closing after removal of the only photo. Reuse the presentation ownership
mechanism; keep draft-photo presentation separate from the large Add screen.


### Edit discard confirmation ownership

The native Edit discard alert belongs to the current asset, draft and focused visit.
A retained acceptance after a draft change, blur/refocus or unmount must not navigate
away. A valid acceptance leaves once; pending saves remain protected by the existing
operation lock. Preserve Keep editing and the current draft.


### Provider archive confirmation consumption

Each native provider Archive confirmation permits one attempt. Failure retains
retry through a fresh confirmation; replaying the old acceptance after the request
settles must not issue another archive request. Preserve current-visit checks and
pending-operation exclusion.


### Asset sheet mutation completion ownership

Edit, Move, Move Here and destination creation retain their draft lock until the
request settles, but completion presentation belongs to the initiating focused
visit. A mounted sheet that lost focus and returned must not receive a stale
completion notice, navigation, error alert or destination/draft replacement.
Background work may settle normally; unlock the still-mounted form afterward so
the current visit can continue. Preserve current-visit success, failure and retry.


### Provider profile task controls

Provider list Add Profile and detail Replace Credential/Prompt Guidance open
substantial editor destinations and must use the existing settings navigation row
with disclosure, rather than a command row. Creating a recommended draft, testing
a connection, enabling/disabling and opening Archive confirmation are commands
and must reuse NativeCommandButton. Accessible command names match their visible
labels; the profile heading and archive confirmation carry subject context.
Creation labels name the action (Create plus service name). Preserve pending
labels, disabled navigation during work, single-use confirmation and retry. The
archive confirmation retains its destructive semantic; no shared adapter changes.
Native spacing, command sizing and assistive navigation require runtime acceptance.

### Representative Home header acceptance

The native audit must exercise Home with Add, Notifications and Profile together,
a long inventory name, production header appearance and no root back button.
Use production Home and expiration content with controlled domain data sufficient
to scroll. Assert actual content movement, left-to-right action order in the
English fixture, contained and hittable action bounds, selector separation, and
stable header position after scrolling. Retain before/after screenshots for visual
scroll-edge review. Geometry assertions do not prove material transparency or the
full production tab/accessory composition; retain those acceptance limits.

### Item-type query recovery in Add and Edit

An initial failed item-type query must show actionable failure and native Retry,
without simultaneously claiming expiration settings are still loading. Preserve
the name, description and other draft values while retrying. A retry may show
loading while no data is available; success restores the type/date editor. Keep
cached usable type data visible during transient refresh failures. Access failure
continues to hide unavailable data through the existing scoped query boundary.

An expanded item-type search with no matches must say so beside the query. Clearing
or changing the query restores matching choices without changing the selected type
or expiration draft. Keep this descriptive, searchable selection in the current
form; do not add a navigation destination for the empty state.

Native notice acceptance must query the exposed accessible notice container and
validate its message, not require separate static-text children that may be grouped
for assistive reading. Keep whole-notice bounds, command reachability and retry
assertions; a locator correction must not remove the behavior being verified.

The compact iOS tag color well must retain a44-point minimum actionable target.
Use the platform's larger control size and minimum frame constraints rather than
scaling its drawing or widening an inactive label. Verify touch delivery independently
of the accessibility frame: the system color well exposes28/36-point AX bounds
while run350695 opens and dismisses it from all nine center/edge/corner probes
on both devices. Retain nonempty, onscreen, compact AX bounds and ordinary activation
checks, plus all nine delivered-touch probes. Record the smaller AX frames rather
than failing touch acceptance solely from their size. Sampled probes do not prove
every point in a44-point square, VoiceOver operation, or color editing; retain
those evidence limits and the earlier activation failures. React wrapper dimensions
alone never establish native touch acceptance.

### Focused native diagnosis

Manual native-audit dispatch may select the two color-picker interaction cases
on both supported simulator devices. This shortens feedback on a changed native
adapter while a full audit proceeds. The selection is a fixed workflow choice,
not caller-supplied shell or XCTest arguments. Default/manual All and pull-request
runs still execute both complete onboarding and fixture suites. Focused runs
record their selection with revision artifacts and are diagnostic evidence only;
they never satisfy full-batch native acceptance or release readiness. Keep existing
live runs and their evidence intact.

### Photo upload recovery audit acceptance

S101 must exercise the mounted asset workspace with the real upload command and
a controlled repository: one selected photo fails while another attaches, Retry
resubmits only the failed selection without reopening the picker, and successful
recovery removes retry/progress while preserving the asset. This complements
command-level partial-failure checks; it does not establish native geometry,
VoiceOver announcements, or physical camera/library permissions.

### Root session callback ownership

Authentication-required callbacks belong to the composition that created them.
After replacement or root teardown, retained callbacks must not expire credentials
or replace the current screen with old connection onboarding. Expiry completion
and dialog actions must also ignore a superseded composition. Verify the mounted
gate with the real onboarding command and controlled auth/profile ports across
sign-out, server change and expiry followed by a new completed session.

Composition ownership ends when sign-out, server change or expiry successfully
returns to onboarding, including the interval before another sign-in completes.
Failed push cleanup retains the current composition. Expiry dialogs need only a
dismiss action, not a callback that mutates the next session’s prompt state.

### Ordinary text entry isolation

For the English URL-keyboard onboarding fixture, readiness may query the observed
`q` key directly rather than enumerating every key and testing zero-sized padding
elements. Keep the same keyboard-existence, hittability and timeout requirements.
This optimizes test observation only; it does not prove keyboard readiness or text
fidelity until the named native build passes. Other keyboard layouts require their
own appropriate readiness target rather than inheriting an English key assumption.

Add separate controlled and uncontrolled paced-injection diagnostic cases using
the same assisted input and string, entering one character per XCTest call. Keep
all existing whole-string cases and their exact assertions unchanged. This is an
explicit timing comparison, never a replacement acceptance test: a paced pass
cannot close a failed whole-string case or justify slowing production entry.
Record its distinct screenshot names and include it in the text-entry diagnostic
workflow. Do not infer that an upstream issue matches without the same conditions.

Include a runner-only SwiftUI TextField comparison with default text assistance,
the same whole-string injection and native/application-value assertions. It must
not inherit URL keyboard or autocorrection-disabling modifiers from the existing
address comparison. This separates the native field path from React Native text
input without changing production fields or relaxing any original failing case.
Passing the comparison does not establish physical typing or prove the cause.

The iOS Add name field may use the already-pinned SwiftUI TextField as a scoped
candidate after the default-assisted comparison passes on phone and iPad. Keep
ordinary text assistance, its visible Name label and accessible Asset name. Seed
once per existing name revision; restored drafts and Clear draft deliberately
remount the field, while typing, metadata refresh and rejected saves do not. Keep
application state as the save/validation owner and disable changes while busy.
Android retains its current field. Do not generalize this migration to multiline,
search, generated-key or externally controlled fields. Unchanged Add whole-string
typing, failed-save recovery, clear and restored-draft native journeys must pass
before claiming the candidate solves product text entry.
The focused `add-draft` workflow selection runs the existing Add typing/recovery,
navigation-stack, preconfigured-header and unfinished-tag journeys unchanged on
phone and tablet. It is a diagnostic subset, not full native acceptance.

The same scoped native name adapter may serve Add's inline new-tag name after
run35058684319 passed phone asset-name entry but visibly truncated `Camping` to
`Cing` in the separate tag field. Preserve its normal text assistance. Native
editing owns its mount seed; application state still owns validation, draft
persistence and staged tags. Staging a valid tag explicitly resets the field;
invalid staging, color changes and removing another staged tag must not reset it.
Collapsing/reopening details restores the unfinished name. Clear draft and scope
restoration replace the seed deliberately. Keep the unchanged native unfinished-tag
journey as acceptance; adapter tests alone cannot establish typing fidelity.

Shared native navigation search options must settle across navigation-context
updates. Query and caller callback changes update current committed handlers and
native text, without reconstructing unchanged header presentation. Enabled state
and placeholder changes still update presentation. Old events after disabling or
unmount must do nothing; return to a focused enabled route restores interaction.
This prevents repeated header reconfiguration; it does not by itself prove the
dynamic search placement failure is resolved. Retain native placement acceptance.

The native tag color well must expose one accessible “Choose any color” name.
Use the native ColorPicker label as its naming source; do not repeat it through
an accessibility-label modifier. Retained350504 hierarchy duplicates the name.
Keep a separate native exact-name assertion so naming acceptance does not mask
or depend on the currently failing picker-activation journey.

Onboarding's whole-string address check must observe completion with a bounded
five-second exact-value predicate before retaining its exact equality assertion.
Run350549 recording shows complete text after the initial partial-value snapshot,
with focus retained; typeText returning is not sufficient completion evidence.
Do not retry typing, substitute paced injection, accept substrings or suppress a
timeout. Capture the resulting state whether the predicate succeeds or fails.

The onboarding dismissal gesture must read each required frame once per gesture
and use those local bounds for its origin, containment assertion and destination.
Repeated XCUI frame resolution can stall for tens of seconds on the retained
350549 phone run. Keep the gesture coordinates, keyboard-disappearance deadline
and following action/draft assertions unchanged; do not cache across gestures or
silently treat delayed teardown screenshots as passing acceptance.

The native audit must compare the same ordinary text with the default keyboard:
uncontrolled baseline, controlled value, uncontrolled without keyboard assistance,
and uncontrolled without the app keyboard accessory. Retain the baseline and exact
native/observed-value assertions; do not slow typing, disable correction globally,
or replace product inputs merely to pass automation. Accessory isolation keeps its
host mounted and uses its enabled flag. These are diagnostic fixtures, not proposed
product behavior. A fixed `text-entry` workflow selection may run these comparisons
and the multiline baseline on both devices; full/manual-All/PR acceptance remains
unchanged. Record failures and their actual build before inferring a root cause.

If the controlled ordinary-input comparison fails, also compare that controlled
input with assistance disabled and with the accessory disabled. A passing
uncontrolled variant cannot isolate either variable in a failing controlled field.
Keep all baseline cases and exact-value assertions in the focused selection.

Definition/tag collection header changes require native search activation, exact
filtering, clear/return, and Add reachability on phone and iPad. A synthetic tag
collection may exercise the shared screen with real query/policy adapters and a
controlled repository; this is not production authorization or pagination evidence.

### Android filter sheet geometry

Browse and Expiration filters must expose their current page title and both commit
and return/cancel commands at every allowed Android sheet detent. The pinned native
stack does not display headers inside Android form sheets, and an absolute JS
footer tracks the expanded content height. Use its Android `unstable_sheetFooter`
adapter for these commands and an accessible title inside the scroll body. Keep
footer space reserved for the last selectable row. Footer presentation must settle
across navigation renders while callbacks read the latest committed draft, disabled
state and teardown; do not freeze the initial draft. Preserve the existing iOS
direct-scroll layout. Native acceptance covers initial/expanded detents, nested
tags, apply/cancel, last-row scrolling, keyboard and system Back.
If filters are the first route after a cold link, Android presents them as a root
screen rather than a sheet. In that case reserve a normal full-screen footer;
do not rely on the sheet-only footer slot. Hide native header chrome in this
Android adapter so the accessible body title remains singular in both cases.
Cold-root content must reserve the top system inset. Route bodies must not reapply
headerShown=true over the Android sheet adapter on query/page rerenders; iOS gets
its native header from the root stack registration.

Browse and Expiration filter Cancel must work in loading, failure and ready states.
With history, return Back without applying the draft. Without history, replace
the root Home route (`/`) rather than leaving an inert Cancel. Browse must cancel
its pending navigation operation before leaving; neither fallback applies filters
or trusts a stale inventory identity.

### Android native header and vector compatibility

Android header actions retain native Compose IconButton controls, accessible names,
48dp hosts and the requested Add/Notifications/Profile order. Only a positive unread
count may instantiate BadgedBox; its native fallback otherwise draws an unintended
dot. Zero/absent counts must render the icon directly.

The pinned Expo UI55.0.17 XML vector loader supports pathData and fillColor but
ignores stroke properties and fillType. Shared Android vectors must therefore use
filled contours supported by this adapter. Guard these assets mechanically against
unsupported stroke/fillType attributes, with rejecting and accepting fixtures.
The mobile structural pre-commit hook also runs for XML asset edits.
Review every consumer (headers, conversation commands, notification read state,
filter/sort). Native acceptance requires recognizable icons, correct badges and
working actions on normal-size Android, not merely a successful APK build.
Use Google's filled Material24px icon contours from reviewed repository revision
`40a7a292a79d9394157e1ea24f83d52d5e17c556`; retain source mapping and Apache2.0
license alongside the local vectors. Convert only opaque path contours to Android
XML, omitting SVG canvas paths marked fill=none; no runtime asset download.

### Notice geometry diagnostics

Notice placement acceptance must retain its full-rectangle containment requirement.
When timed geometry checks disagree with final captures, record the last sampled
header, content, app and control rectangles and each control's containment result
in the failure message. Do not infer a corrected product layout from a later
screenshot or relax the bounds to make an intermittent observation pass.

### Isolated Android runtime preparation

Fixture installation outside GitHub Actions requires an explicitly disposable
source archive: `MOBILE_AUDIT_ARCHIVE_ROOT` must resolve to the installer's root,
which must contain a regular, non-symlink `.mobile-audit-archive` marker with exact
content `disposable-mobile-audit` plus newline. Refuse any `.git` file/directory
in that root or an ancestor, and require `RUNNER_TEMP` outside the archive for
the production-route backup. Retain the explicit fixtures suite requirement and
refusal to overwrite a previous backup. A marked archive is never a distributable
release checkout. Ordinary local checkouts remain rejected without modifying routes.
Before any route backup or deletion, reject symlinks in every route-path component
below the archive root so an archive cannot redirect mutation into another checkout.

Android simulator preparation may run on the authorized remote validation host,
outside the repository and existing generated Android directories. Bootstrap the
Linux command-line SDK tools from Google's numbered15859902 archive, verifying
SHA-256 `4e4c464f145a7512b57d088ac6c278c03c9eea610886b35a5e0804e74eedf583`
before extraction or execution. The `_latest` filename suffix is part of that
numbered, checksum-pinned artifact; do not resolve an unversioned latest download.
Record subsequent emulator/system-image package revisions and checksums before
installing them. Check disk capacity and KVM access before provisioning a virtual
device. Installing tools or booting an emulator does not establish app acceptance;
record actual build, device, Android version and exercised interactions separately.

For the isolated Android preparation, any further Android CLI use must invoke the
captured implementation directly (version1.0.16261425), verifying SHA-256
`847e24a7d1711561a8739629b59c6e09b5a80dbfd98045d6ce7c661f46ecbc81`
first. Do not rerun the SDK wrapper's automatic CLI download. Record the initial
bootstrap's unpinned execution as a historical supply-chain gap; capturing a pin
afterward constrains subsequent use but does not retroactively verify that step.

The disposable Android build uses the pinned React Native0.83.6 catalog's API36,
build-tools36.0.0 and NDK27.1.12297006. The generated Gradle9.0.0 wrapper must verify
its distribution against published SHA-256
`8fad3d78296ca518113f3d29016617c7f9367dc005f932bd9d93bf45ba46072b`.
Keep build caches outside the source checkout and restrict the emulator build to
x86_64. A debug-signed audit APK is synthetic evidence only and must never enter
the distribution pipeline.

Provision the isolated build's Java17 toolchain explicitly rather than relying on
React Native's incompatible Foojay0.5.0/Gradle9 automatic resolver. Temurin17.0.16+8
Linux x64 HotSpot archive must pass SHA-256
`166774efcf0f722f2ee18eba0039de2d685b350ee14d7b69e6f83437dafd2af1`
before extraction/execution. This is a validation-host toolchain, not a change to
the app dependency graph.

Android filter selection pages must expose search inside the sheet body: native
navigation headers are unsupported for Android form sheets. Reuse AppTextInput
(the platform TextInput/EditText adapter) as a labelled single-line search field
with the search IME action, no automatic capitalization or correction, and live
filtering. iOS keeps native navigation search. The shared sheet owns placement;
searchable callers supply query handlers, not platform-specific layout. Leaving a
selection page removes search and preserves existing draft/Back semantics.

The keyboard acceptance run rejects the Android native sheet-footer candidate:
react-native-screens 4.23.0 calls a stable-state footer calculation during keyboard
settling and crashes. Android filters therefore use a full-screen native stack
route with its normal title bar, an in-body search field, and a layout-owned bottom
action region. A flexible scroll body reserves that region without overlaying
choices. This preserves all filter functionality and draft semantics; iOS retains
its current detented sheet. Do not patch or upgrade dependencies to hide this
failure without a separately reviewed dependency change.

Native fixture setup must distinguish the synthetic audit menu from production
navigation. Run350702's Add test tapped an audit-menu label but opened the unrelated
inventory-query fixture. Isolated Add comparisons must enter their explicitly named
fixture URLs with XCUIApplication.open, then assert Add content before interacting.
This does not establish Home-to-Add navigation; retain the separate Home header
and production-configured Add scenarios. Hidden-header Add variants remain
configuration diagnostics, not evidence that the production preconfigured header
regressed. Capture color target bounds and pre-tap appearance when investigating
an ordinary activation failure alongside passing delivered-touch probes.

M207's next native comparison must isolate delayed search registration. A synthetic
route starts without search, enables the production NativeNavigationSearch adapter
on an explicit command, captures its placement, then changes only the route title
and captures placement again. Record both geometries before requiring header
placement; do not drop the original production Place/settings assertions. A title
change is a diagnostic, not a production timing workaround. No speculative native
library patch is accepted from source inspection alone.

Android Move Here at its initial0.6 detent renders its form without Move/Cancel.
Edit, Move and Move Here share the same sheet sizing approach. These asset action
routes must use Android full-screen native-stack presentation with native title
bars and layout-owned actions, matching the corrected filter approach. Their
existing dirty/operation guards and commands remain authoritative. Shared action
forms must resize for the Android keyboard using the actual native header height;
iOS keeps its existing sheet detents, header visibility and keyboard behavior.
Verify each consumer before claiming shared runtime acceptance.

The Android asset-action migration must not expose an unguarded Edit exit. Header
Back, hardware Back and navigator removal must use Edit's existing discard
confirmation, with current draft/visit ownership, while pending operations block
removal. Confirmed discard and successful Save authorize exactly one removal;
Keep editing preserves the draft. Apply this at the navigation boundary, not only
by hiding the header button.

On Android full-screen asset actions, use the native header for the task title;
do not repeat Edit/Move task headings in the scroll body. Preserve contextual
asset names, instructions and previews. iOS headerless sheets retain body titles.

Android Compose controls must follow the resolved in-app appearance preference,
including when it differs from the device theme. All project Compose hosts share
one appearance-aware adapter; do not rely on each host's system-theme default.
Preserve Material enabled/disabled colors and existing interaction semantics.
Acceptance must include switching light/dark while controls remain mounted,
readable enabled secondary actions, disabled-command non-execution and enabled
primary execution. Source propagation alone does not prove rendered contrast.

Android Add must use a full-screen native-stack card so its Close and Save header
commands remain visible. Android form sheets do not render the requested header.
Keep the existing retained-draft and busy-operation guards, save eligibility and
return behavior; iOS keeps the full-height native Add sheet. Production route and
the production-equivalent native audit fixture must share these options.

Closing Add must return to the previous route when history exists and replace the
root with Home otherwise. Reuse the same navigation-only return policy as filter
cancellation; callers retain their own draft, cancellation and pending-work guards.
The isolated Add native fixture must exercise the same policy. A retained Add draft
must not be cleared merely because Close chooses the Home fallback.

Inventory switching and checkout history require visible native titles and Close
commands. On Android they use full-screen stack cards because form sheets omit
those requested headers; iOS retains their existing detents. Share inventory
switcher presentation between production and the native fixture. Checkout history's
shared presentation also hosts Home return details, whose pending-return and exit
guards must remain intact and receive focused regression coverage.

Inventory switcher Close and successful selection must also return Home when no
back destination exists. Preserve immediate callback retirement and abort handling
when leaving; failed selection stays open for retry.

Return details must not toggle native removal protection in the same commit that
clears its task and pops the route. Keep the route's removal guard installed for
its lifetime: active tasks delegate attempted exits to their existing close policy;
completed/absent tasks dispatch the original removal action. Native acceptance must
positively observe the destination screen and surviving app process after Save or
Cancel; disappearance of the task alone is insufficient and can hide a crash.

Android Conversation requires its native Close and New conversation header actions.
Use a full-screen native-stack card rather than an Android form sheet, which omits
that header. Preserve iOS detents, retained conversation state, media pause on Close,
and confirmation before discarding an unfinished proposal through New conversation.
Verify context placement and proposal actions after the presentation change.

## Full-screen photo viewer command visibility

The fixed black photo canvas requires a self-contained native filled Retry command
so the app's light appearance cannot put dark text directly on black. Retain the
same retry callback and unavailable-photo state. Use the existing safe-area-aware
viewer toolbar as the single explicit Close affordance; suppress the image library's
redundant default header, which overlaps Android status icons. Preserve swipe/system
Back dismissal and removal/paging controls. Verify both asset and draft-photo consumers.

### Invitation email keyboard semantics

Sharing email entry uses the email keyboard and email autofill, without automatic
capitalization, spelling correction or spell-check underlining. Email local parts
are identifiers, not prose. Keep native draft ownership and successful-create/scope
reset behavior. Run350950's unchanged exact-email native assertion fails visibly;
changing keyboard traits is a candidate correction, not proof that missing typed
characters are resolved. Preserve that native gate and submission/retry checks.


### Enum option editing acceptance

Normal-text acceptance of custom-field choices does not establish option editing.
Exercise the production controls with a duplicate option: enter its complete text
once, invoke Add, retain the exact draft with the duplicate explanation, clear it,
enter a new multiword option once, add it, and observe the canonical option plus an
empty entry field. Remove only that new draft option and preserve the existing one.
Commands must remain reachable with the keyboard present; the existing form and
Back destination must survive. Keep saved-option immutability in its already
accepted separate scenario instead of replaying unrelated picker/target workflows.

The iOS enum option field uses the shared native draft adapter after run35836383102
retained only `r` from `ready` on both devices. Ordinary typing, validation errors
and busy-state changes keep the native editing instance. A successful Add advances
a local reset revision and clears the field; Android keeps its controlled input
instance. Changing the enclosing loaded resource already unmounts controls through
the editor loading state. Type disclosure re-entry seeds the retained pending value.
Keep the field's validation accessibility hint alongside its visible error. Keep
the native modifier sequence stable when the hint appears or clears: supply an
empty hint rather than inserting/removing its modifier. Expo55.0.17 wraps modifiers
structurally and reseeds TextField on appearance; validation must not recreate that
subtree and replace the user's draft. Run35839289171's iPad duplicate rejection
ended with an empty field; verify the unchanged recovery workflow after correction.

Use one focused native pass on phone and iPad. A failed exact-value assertion in
this production consumer selects the established native draft-field adapter;
it does not reopen provider-removal, key-delivery or pacing experiments. If typing
and recovery pass, retain the existing implementation and record scoped acceptance.


Text-entry acceptance reads the actual native field value immediately after the
single typing action, before Add, using XCTest's direct value assertion as in the
Sharing email workflow. Do not treat a later teardown snapshot as a pass. The
nested five-second diagnostic waiter is not a product response-time requirement:
run35842653888 recorded one4.26-second false evaluation on iPad and complete text
at teardown, while phone passed the whole recovery sequence. Remove that timing
wrapper for enum text/clear assertions; preserve exact strings, keyboard-open Add,
duplicate retention, normalized creation, successful reset and selective removal.

### Shared Settings name and save recovery acceptance

The representative Settings editor must accept a complete multiword name change,
not only a one-character append. Exercise the actual routed editor with a controlled
repository: the first exact-payload save is rejected, the full name stays editable,
and retry with the same payload returns to the collection with its success notice.
The fixture uses the production header, keyboard container, application managers
and editor; rejection belongs in the repository fake, not in alternate UI logic.
Keep existing successful-save, dirty-back and lifecycle cases intact. A real typing
failure selects native draft ownership for this shared single-line editor family;
no repeat provider-removal or typing-speed experiment is needed.

### Normal-text Asset Edit recovery acceptance

Verify the routed production Edit form at the default text size before pursuing
its enlarged-text failures. Enter one complete multiword name while asset-type
and tag metadata reads have failed, dismiss editing normally, retry both metadata
loads and retain the exact draft. A rejected Save must retain the name and permit
Keep editing; deliberate Discard returns to the originating screen. Use the existing
controlled metadata/command fixture. This verifies recovery, not successful backend
persistence. Keep enlarged-text scenarios and their unresolved findings separate.

### Normal-text Edit tag acceptance

Run the existing routed tag draft/disclosure sequence at default text size, sharing
its assertions with enlarged-text coverage. Verify exact one-attempt tag entry,
Save disabled while unstaged, Cancel/Keep editing retention, accepted Add clearing
the input, native expansion/collapse retaining selected extras, and explicit
Discard removing the editor before the audit launcher is considered restored.
This verifies local draft handling and selection, not successful server persistence.
A reproduced entry failure selects the established native draft adapter; do not
repeat provider or typing-speed experiments.

### Normal-text detail and Move Here acceptance

Before enlarged-text remediation, reuse existing detail action geometry, independent
photo/contents recovery, and Move Here query/suggestion retry scenarios at default
text size. Retain exact query input, separate error/empty outcomes, native command
reachability and return to the launcher with the destination removed. Run these
representative consumers together; no assertion of mutation persistence, assistive
coverage or whole-app acceptance follows from them. Keep historical enlarged-text
results independent.

For the Edit name append scenario, readiness targets the first required space key
using the existing named-key helper, rather than enumerating the whole keyboard
in a timed predicate. Native35847685581 stops before typing on iPad after one
4.68-second enumeration; its retained capture shows focused input and keyboard.
This is not proof of input loss or a passing readiness deadline. Preserve exact
one-attempt input, metadata retry, rejected Save and discard checks in acceptance.

Tag reachability acceptance uses ordinary native scroll swipes, not repeated short
press-and-drag gestures that can leave a multi-detent sheet at its starting height.
Run35848704289 failed on both devices before typing with the correct form scroll
owner, selected tag below its viewport, and no progression after18 short drags.
One follow-up uses full native swipes with unchanged tag visibility, draft and
selection assertions; if reachability still fails, investigate the sheet layout
as a product defect rather than continue gesture tuning.

### Move Here native query retention (M257)

Move Here reuses the shared native DraftTextField on iOS so ordinary query, loading,
selection and error renders preserve the native editing buffer. Android keeps its
existing controlled input. Asset/tenant/inventory ownership already keys the route
form and resets the query when that owner changes; do not add resets to lookup or
selection updates. Preserve read-only and busy guards, suggestion retry, selection
and cancellation behavior. Native35850832085 is the failing regression: one Tent
entry leaves T on both phone and iPad before any suggestion action. Acceptance
requires the exact full query and successful retry with that query retained.

### Edit new-tag native draft retention (M258)

Edit uses the shared native DraftTextField for new tag names on iOS, with a
flex-width owner beside the color field. Ordinary typing, invalid input, selection,
color and disclosure changes retain the same editing buffer. Only an accepted tag
resolution that clears inputs advances the native field revision; Android retains
its controlled field instance. Preserve staged tags, validation, busy/read-only
guards and route ownership resets. Native35853305160 reached the tag field on both
devices, then phone reduced one Camping entry to C. iPad stopped before typing in
whole-keyboard enumeration; readiness should observe the stable space key without
weakening the exact text assertion. Acceptance requires full text, accepted clearing,
staged tag presence and retained selection through disclosure.
