# Comprehensive mobile UI audit and remediation

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
