# Platform Interaction Review

## Move creation acceptance observations

Run35867595817 retained iPad hierarchy contains the full `Audit crate` value after
an immediate post-typing assertion read `Audit cr`. Wait at most five seconds for
exact field equality before proceeding; never accept partial text. Its phone
journey created and selected the destination, then queried obsolete standalone
`Audit crate` status text. M262 presents `Selected: Audit crate`; assert that exact
status and the selected candidate row, retaining all create/move rejection and
retry assertions. This corrects observations, not production input behavior.
The corrected complete workflow still requires native acceptance. Edit metadata
recovery and tag entry passed on both devices in this run; connected phone Edit
stopped at keyboard readiness before typing and remains unverified by this run.

## Map and empty-detail hierarchy acceptance

M265/M266 follow-up acceptance uses existing Browse and connected Edit fixtures.
Verify root Map has no ancestor button, opening Garage reveals breadcrumbs, and
returning to the root removes them while retaining the column heading. On a
photo-empty item, verify title, Edit and Move precede compact No photos status,
with Add photos reachable in the initial viewport. Include normal-size detail
commands and region recovery to check representative containable consumers and
loading/error distinction. These checks do not certify populated photo paging,
physical photo selection or unrelated Map gestures. Keep this follow-up separate
from the frozen M260–M264 release.

## Custom-field choice acceptance

M02/M11 need native evidence for their distinct control composition. Exercise the
production CustomizationFieldControls in a scrolling create-form fixture: choose
Type and Applies to in place, expose enum controls, select two asset types separated
by scroll, remove one draft selection and preserve the other. Keep visible labels
and verify the same header remains. Controlled draft observers are fixture-only.
This does not certify field persistence, keyboard entry, inherited immutable targets
or assistive behavior; their existing source and workflow evidence remains separate.
Native35821158725 exposes the pickers as `Type, Choose Type. Current value Text`
and the equivalent Applies to label: SwiftUI LabeledContent adds its visible label.
Match the semantic choice label within that composed name, following the existing
Browse choice acceptance pattern; do not assume the custom label is a prefix.
Keep independent visible-label and state-transition assertions. Run35822973764
shows the exact first selection visibly rendered and the row checked on both
platforms, while XCTest reports its passive observer as not hittable. Check passive
state with exact text and visible viewport bounds; reserve hittability assertions
for controls the user activates. Preserve every intermediate selection assertion.
The two focused runs exhausted this acceptance investigation budget. Include the
corrected observation in the next broader batch rather than dispatching another
isolated run; the remaining iOS selection/removal sequence stays unverified.
Use one focused phone/iPad run plus Android acceptance. Investigate a failure only
if it distinguishes a control defect from fixture/observation behavior; retain the
default one-source-pass/two-experiment budget. No production change is presumed.

## Transient success observation order

Native35815492811 phone passed Archive confirmation/cancellation and returned to
Tags, but began querying its4200ms success notice about ten seconds after Archive.
Observe the transient exact notice immediately after the confirmed command, then
wait for the persistent Tags destination. Preserve both assertions and durations;
do not lengthen the product notice or claim the late snapshot proves it appeared.
This harness-only correction needs a focused archive-workflow run; it does not
invalidate the independently passed menu lock/recovery and filter checks.

## Onboarding fixture viewport parity

Production AppServicesContent renders onboarding instead of its navigation-stack
children. The submission fixture previously added a visible audit header and an
in-flow observer label. Its Android keyboard clipping is not yet a confirmed
production-layout defect. Hide that header while this fixture owns the root screen
and position its command observer outside layout flow. Restore production viewport
parity before choosing a keyboard/layout correction. Keep the real form, command,
keyboard and exact submission assertions. Do not ship a layout workaround for
space consumed only by the test harness.

## Onboarding command clearance acceptance

M35 remains a normal-text acceptance gap: existing Connect tests explicitly dismiss
the keyboard before activating the command. Verify the real onboarding form with
the URL keyboard still open, allowing bounded upward scrolling. The full Connect
button must fit below the navigation/safe-area top and above the keyboard, remain
enabled, and submit the exact address on one tap through the controlled command.
Do not substitute keyboard Go or explicit dismissal for this check. A failure
before exact address entry is not command-clearance evidence. This is a separate
follow-up scope from the frozen menu batch; no production layout change is chosen
until the named workflow supplies evidence.

## Retained native menu actions

An open native menu may deliver an item event after its parent locks, removes an
item, replaces the current handler, or unmounts. All menu adapters must resolve
events against committed current menu state, identified by group and item IDs.
Reject a globally disabled menu, disabled/removed item, and retired component;
otherwise invoke the current handler. Never run a captured obsolete callback.
Native item presentation must include the global lock. Controlled Android and
fallback popups close on lock and must not reopen automatically after unlocking.
Retained trigger and accessibility callbacks also obey the current lock and teardown.
Keep iOS/Android native controls, selected/destructive semantics and grouping.
Tests must retain a real rendered adapter callback across these transitions;
source tests do not establish native popup refresh or assistive-tech behavior.

Native acceptance uses a runner-only menu fixture with a bounded delayed lock,
visible lock/activation state, explicit unlock and timer cleanup on removal.
Open the production adapter before the lock, require disabled items or a dismissed
popup, verify no activation, then unlock and execute one fresh command. Run this
on phone/iPad with representative filter choice and destructive-command workflows.
This checks native presentation and delivery; retained-event ownership remains
covered by the mounted callback tests. No production data is mutated.

## Native Sharing access fixtures

Runner-only Sharing scenarios may seed a permissionless scope, one rejected access
read, or one ordinary read failure through controlled ports. They reuse production
query policy and screen controls, never real invitations or credentials. Changing
scenario remounts its owned query client and counters. Native acceptance distinguishes
no retry without permission, Check Again recovering after access failure, and Retry
recovering after an ordinary failure. These fixtures prove presentation/recovery,
not server authorization enforcement.

## Sharing access recovery

The Sharing screen distinguishes denied/unavailable access from an ordinary list
load failure. Hide cached metadata on access failure as before. When the supplied
scope lacks sharing permission, explain that limitation and do not expose a list
retry which cannot refresh that scope; the route guard owns fresh scope discovery.
When the scope still grants sharing but a read returns an access failure, show
Sharing unavailable and Check Again through the existing authorized read path.
Do not claim the precise reason for401/403/404; credentials, access and inventory
availability may all change. Ordinary network errors retain Retry. Verify denial
with cached rows, allowed recovery, and no repository reads without permission.

## Missing-link recovery guidance

The current missing-link error confirms scoped creation without exposing the created
invitation ID. Its guidance must remain accurate after cancellation: if the invitation
is still pending, cancel it; if already cancelled, retry creation. Do not infer the
created ID from email, incomplete paginated lists or the next cancelled row. Do not
clear a newer or unrelated creation error after cancellation. Preserve the submitted
email and historical missing-link notice until a new attempt or focused visit resets
it. This conditional guidance corrects M239 without adding unsafe identity inference.
Native acceptance must inspect the guidance and retained email after successful
cancellation, as well as failure/retry and removal of the cancelled row's command.

## Invitation cancellation interaction

A pending, unexpired invitation exposes its single command directly as a native
Cancel invitation text button below the recipient and status. An overflow menu
adds no choice here and makes failed-link recovery harder to discover. This is a
project task/pattern decision, not a universal prohibition on single-item menus.
Keep destructive confirmation naming the recipient, permission/scope ownership,
independent pending locks, failure recovery and terminal-state removal unchanged.
End keyboard editing before presenting confirmation. Use the existing native
command adapter without adding a shared control API. Verify confirmation (including
Keep Invitation), actual cancellation, concurrent invitation locks and recovery in
mounted tests, then native phone/iPad reachability with the keyboard dismissed.
Do not claim the native menu/keyboard root cause is resolved from source tests.

## Native fixture observation boundaries

Run351041's Add name assertion read Nat immediately after typing, while its
screenshot0.33 seconds later and hierarchy0.82 seconds later both contain the
exact full name. The Add presentation journey must wait at most five seconds for
exact equality before continuing save/recovery; no partial, fuzzy or rewritten
input is acceptable. This bounded native observation is not a general exemption
for controlled-field corruption, which remains a failure in diagnostic fixtures.

The Add-photo journey returns to the audit root with its prior scroll position.
Verify the root navigation bar, then reveal its Browse entry with bounded upward
scrolling and require it to be hittable. Do not require a returning scroll view to
reset to its top unless the product explicitly specifies that behavior. Keep all
photo paging/removal and draft-exit assertions unchanged; native rerun is required.

## Native search acceptance on iPad

Run351041's static integrated-button search, configured without app query handlers,
collapses after focused Clear text on iPad; its retained capture shows the Search
button and no field or keyboard, then the same journey reopens and enters Garage.
The production location flow may follow this observed platform behavior. Require
restored unfiltered choices and either a hittable field or, only on iPad, an absent
field with a hittable Search button contained by its navigation bar. Reopen when
collapsed and retain exact fresh-query, selection, draft retention and Back checks.
Do not relax phone behavior or accept unavailable search. This corrects an overly
specific field-persistence assertion, not an app implementation. A new native run
must still prove the complete production journey. Apple's
[integratedButton API](https://developer.apple.com/documentation/uikit/uinavigationitem/searchbarplacement-swift.enum/integratedbutton)
describes inactive search as a button; the focused-clear transition is observed
runtime evidence, not an explicit promise in that documentation.

Integrated toolbar search on iPad can expose Clear text without a separate Cancel
button. Native audit journeys must verify clearing the query, dismissing the
keyboard, restoring unfiltered content and reaching navigation/actions through
that observed platform pattern. They must not require the phone-only Cancel
affordance or force a particular idle field width. Phone journeys retain Cancel
and collapse assertions. Record screenshots/hierarchy that justify this distinction;
an amended test remains unverified until rerun on its native target.

For notice placement, measure the application, active content, navigation bar and
notice controls from one XCTest hierarchy snapshot per observation. Independent
queries in run350806 returned zero control rectangles while the retained hierarchy
showed valid contained controls. Preserve positive-size, full-containment and
below-navigation checks, the bounded timeout and subsequent actual command taps.
A missing snapshot or element fails that observation; never substitute final
screenshots for the live acceptance gate. Use Apple's public
[snapshot API](https://developer.apple.com/documentation/xcuiautomation/xcuielementsnapshotproviding/snapshot%28%29).

## Purpose

Android conversation review must resize its scroll body above the software keyboard
while keeping Approve/Cancel visible and reachable. Use the existing native stack
form pattern: height-based keyboard avoidance with the native header offset.
iOS retains sheet padding behavior. Verify an actual keyboard-visible capture
and command activation; accessibility bounds alone can include obscured controls.

Voice review decisions belong to one proposed plan within one inventory scope.
Retained Approve/Cancel events must read current committed command/photo drafts,
respect an invalid visible name, and reject hidden or removed review tasks. Mount
a distinct action owner when the plan or scope changes, or when review ends; an
old event must never approve a replacement plan. Preserve controller duplicate
submission locks. Verify workspace-to-transport payloads, refocus, replacement and
teardown with mounted tests before claiming this source repair complete.

Asset Edit, Move and Move Here footer events must resolve the current committed
draft and eligibility within their mounted asset task, reject events while blurred
or removed, and retain the existing duplicate-operation lock. Changing the scoped
asset remounts the task: retained events from its predecessor must never submit
for the replacement. Reuse the focused footer guard with filters rather than
adding another callback cache. Verify actual route command payloads and cancellation
through mounted tests; delayed native event delivery remains a separate runtime check.

Filter footer actions on both mobile platforms must read the current committed
draft and current primary/secondary disabled state. Retained native callbacks
must not apply an earlier valid range after the user makes it invalid, or cancel
the whole task when the current nested-page command is Back. Reject callbacks
while the owning route is blurred and after removal; refocusing permits the
current actions. Share this ownership guard across the iOS and Android layouts.
Verify these retained-event cases through mounted behavior tests; source evidence
does not establish that native platforms delivered a delayed event in practice.

Runner-only native fixture navigation must render the production root's appearance
defaults: palette-backed content and header surfaces, text-colored header titles,
action-colored controls and status glyphs matching the resolved color scheme.
Otherwise a system-theme change can leave a light native header over a dark task
and invalidate appearance observations. Synthetic route/service composition stays
isolated; matching appearance does not make fixture evidence a production-data test.
Inspect both themes on a rebuilt native candidate before accepting this correction.

Choose familiar platform interactions before implementation. Native components alone
are not evidence that a workflow uses the right interaction. This standard covers
Stuff Stash iOS, Android, web, and human-facing documentation surfaces.

## Decisions

- Extend `.codex/skills/stuffstash-ui-design` to design, audit, and implement mobile
  and web workflows. Preserve web candidate guidance as a conditional reference.
- For each changed interaction, record the user task, selected platform pattern,
  source guidance, commit/cancel behavior, state coverage, and adapter used.
- Prefer established in-place choices for small flat selections; use searchable
  selection views for long, hierarchical, descriptive, or multi-selection tasks.
  Navigation, modality, and custom controls each need a task-specific reason.
  This is a project default, not a claim that Apple prohibits selection screens.
- Use current Apple HIG topic families as the iOS audit coverage framework. Include
  foundations, patterns, components, inputs, technologies, and platform adaptation.
  Mark each topic applicable, conditional, or not applicable with a reason; never
  interpret an optional Apple technology as a requirement to add a feature.
- Apply the underlying usability principles to Android and web using their own
  platform conventions. Do not impose iOS-only chrome or gestures on other clients.
- A full audit inventories routes, nested states, shared controls, system entry
  points, and their consuming surfaces. A source review is distinct from a rendered
  review and from interaction verification on a named OS/device/build.
- Findings state observed behavior, user cost, expected interaction, source evidence,
  priority, recommendation, and acceptance scenario. Distinguish source-confirmed
  choices, observed defects, recommendations, and unverified device risks.
- UI reviews must check interaction appropriateness before visual execution and
  code correctness. Shared changes require a consumer impact inventory.
- Native layout/gesture fixes need native runtime evidence for verified claims.
  When runtime access is unavailable, record the gap. Existing release authorization
  remains valid; never describe an untested runtime outcome as verified.
- Use meaningful behavior tests and native acceptance evidence. A test matching a
  prop or screenshot alone cannot prove that a selected pattern fits a user task.
- Record durable audit reports under `docs/reports/`; no production UI changes are
  implied by an audit request. Prioritize remediation for explicit follow-up work.

## Verification

Validate skill metadata, linked references, and topic/surface inventory completeness.
Exercise the revised skill on concrete contrasting interaction examples. Run the
code critic on the policy/skill/report changes. Report any runtime coverage gaps.

## Audit release batching

Prioritize all findings reproducible at normal system text size before work
specific to enlarged text. This includes functional flows, platform interaction
choices, navigation, layout, contrast, keyboard behavior, and recovery. Do not
advance to enlarged-text remediation while known normal-size findings remain
unresolved; continue independent normal-size work while native verification is
pending. Preserve existing accessibility fixes and keep enlarged-text findings
tracked for the subsequent pass. This is the user's remediation sequence, not a
change to accessibility requirements or evidence standards.

During comprehensive remediation, accumulate substantial groups of related fixes
before releasing. Individual fixes remain atomic commits, but are not individual
release boundaries. Review shared consumers and run combined validation for each
batch. An explicit user request may cut an interim release from the validated
subset. An already-running release may finish while the next batch accumulates.
Do not replace completeness or native acceptance with a quota of commits.

Native proposal-selection acceptance uses the production conversation workspace
and destination route screen under synthetic voice/query ports. Share the real
voice sheet options with the fixture; keep its provider above both native routes
so Back exercises retained drafts. Test lookup failure/retry, native search,
immediate destination selection, Back and visible proposal state. Fixture routes
seed their initial proposal once per provider lifetime, not per route mount.
Activate seeding only on first entry to the proposal route; unrelated fixtures
must retain their idle conversation state before that entry.
Closing/reopening a route must retain the conversation; confirming New conversation
must leave it empty without synthetic reseeding. Verify both outcomes positively.
Fixture routes
must remain runner-only and must not load production services or credentials.
Simulator results do not establish physical capture or server authorization.

Move Here native acceptance must cover suggestion recovery, selection, rejected
command draft retention, and successful retry with a positive return destination.
Its isolated command fake rejects the first valid move and accepts the second;
it validates the selected asset and target without contacting production services.
Asset-action removal protection remains registered for the form lifetime, including
the transition from a completed command to route removal. Dispatch idle Move exits
and authorized completions through that guard; retain Edit's discard confirmation
and block pending writes. Do not toggle native removal protection during teardown.

Shared provider-settings loading and failure views must identify the current task
(voice setup, voice stage, profile list/detail, credentials or prompt guidance).
Use the shared labeled progress row; do not substitute an unlabeled spinner or
label unrelated editor failures as Voice Setup. Tenant-context loading also names
its task. Keep the existing scoped query and native retry behavior.

Native acceptance selectors must distinguish duplicate accessibility descendants
from distinct controls. When a recorded hierarchy exposes the same nested text
twice, select its first matching text container explicitly; retain the complete
visibility assertion. Do not weaken geometry or interaction checks to make an
ambiguous selector pass. Record the failed run and require a native rerun.
Run350633 exposes nested duplicate `Recently changed` text nodes in the Home
scroll journey. Select the first text container for that heading while retaining
the measured scroll displacement and stationary, ordered header assertions.
Its iPad capture also exposes the top Home/Browse tab strip as an `Other` container,
not `TabBar`. Scope iPad navigation to the smallest common container containing
both named tab buttons; retain phone TabBar lookup. Verify both buttons are
onscreen/hittable and preserve strip/accessory exclusion, actual tab transitions
and return-state checks. This follows the observed hierarchy, not a requirement
that iPad use phone bottom tabs. Apple documents adaptive tab placement:
https://developer.apple.com/documentation/uikit/elevating-your-ipad-app-with-a-tab-bar-and-sidebar

When a system bar item's accessibility frame is smaller than the recommended
44-point hit region, record the frame failure and separately test delivered
actions near the edges of a centered 44-point square. A frame is not direct
evidence of touch delivery. The Home fixture may count its synthetic notification
callback to observe center and edge taps without navigating away. Keep the
original geometry gate intact until evidence justifies any change to acceptance;
the diagnostic does not certify other actions, devices, or full Home composition.

Extend that independent evidence to Add and Profile through their real Home header
callbacks. Runner-only `/add` and `/settings` destinations identify the dispatched
route; after each center/edge/corner tap, require the correct destination and a
return to the same header. These probe destinations do not stand in for the Add or
Settings workflows. Keep the original frame assertions until those controls have
their own native touch-delivery evidence.
The focused `home-header` workflow selection runs the original geometry/scroll
journey and all three action probes on phone and tablet; it is not full acceptance.

Run35059882579 supplies independent Home action evidence: all three nine-point
probes pass on iPad; phone notification/Profile pass and phone Add completes all
nine destination/return assertions before its screenshot request times out. Both
devices still report a36-point Add accessibility frame. Accordingly, the Home
layout journey must assert nonempty, onscreen, hittable, ordered and stationary
action frames, not equate accessibility-frame size with touch-region size.
Keep raw frames in captured hierarchies and keep all three delivered-touch probes
as separate requirements. Apple's Buttons guidance concerns the hit region:
https://developer.apple.com/design/human-interface-guidelines/buttons
The nine points sample the center and near edges/corners of a44-point square;
they do not measure every point in the region or certify VoiceOver behavior.
The phone capture failure and pending scroll-layout rerun remain recorded failures.

Apply the same frame-versus-delivered-touch distinction to the native tag color
well. Retain the existing frame and ordinary activation journeys; add independent
center and near-edge/corner probes of a44-point square centered on the visible well.
Each tap must open the system picker, then dismiss through the observed platform
affordance and retain the unchanged parent value. Capture the failing probe and
stop that journey to avoid cascading taps into an unexpected presentation. A
passing sample does not certify every point, VoiceOver or color editing itself.

Search placement diagnostics must isolate registration from screen content. Keep
the existing Place journey and add a runner-only route to the same production
Place fixture, with integrated-button search configured before presentation.
Both routes run the same search, result, clear, cancel and Back assertions; the
production shared search adapter still supplies current handlers and query state.
This comparison changes initial route configuration only. A passing comparison
does not certify a production correction or justify weakening the original gate.

When focused proposal-location search collapses after Clear text, compare the
same focused typing/clear gesture against the static native search fixture with
no query-state callbacks. Capture both states and whether a field or reachable
Search button remains, then verify a fresh query can be entered. This is a
diagnostic of platform behavior, not replacement acceptance for the production
proposal-location journey; retain its focused-clear assertion until the cause is
established.

Native text assertions may wait for an exact expected value when a recorded final
capture and hierarchy establish delayed observation after the immediate read.
For the Add tag journey's asset-name entry, run350607 records immediate `T` but
final `Tent`. Use a bounded five-second exact-value expectation after ordinary
unpaced typing; retain the final equality and subsequent draft/save assertions.
Do not replace lost-character checks with prefixes, disable assistance, or treat
the failed original journey as passed. The corrected journey requires a native run.

Proposal-location native acceptance must scroll the conversation's own review
viewport, not the first scroll view in the application (which can be the dimmed
background). Run350592's iPad hierarchy places the location control above the
visible sheet scroll area. Reveal it with bounded, direction-aware gestures in
the scroll view containing that control; require its full frame below the native
header and within the viewport before activation, including after selection and
Back. Preserve native lookup/retry/search and proposal-retention assertions.

Native fixture URL entry uses XCTest's application URL-opening API only inside an
iOS 16.4 availability check. The test target retains its existing minimum OS;
unsupported runtimes fail the URL-entry journey explicitly rather than silently
skipping acceptance. Current iOS 26 runners must execute the same URL and all
subsequent assertions. Run350919 failed compilation at the two unguarded calls;
its fixture jobs provide no runtime acceptance evidence.

Add photo-preview native acceptance must enter through the production Add screen
and its photo-source chooser. The isolated library port returns two distinct IDs
and filenames backed by a bundled image, once per fixture lifetime; subsequent
library requests and camera return no selection. Keep these synthetic selections
out of production adapters. Verify preview paging, removal cancellation/acceptance,
remaining draft thumbnails and Close without saving an asset. Conversation draft
thumbnails alone do not cover the full-screen Add preview consumer.

When onboarding's exact-address wait fails, retain the same timeout and assertion
but attach a failure-only native snapshot listing full text-field labels, values
and frames. Debug hierarchies truncate values and may contain only a failed query
chain. The later screenshot in run350950 shows the full address despite a failed
lookup; it does not prove the value or query state within the acceptance interval.
Diagnostic snapshot failure must be recorded, never converted into a pass.

The iOS Add-preview acceptance journey must positively identify the surviving
filename and thumbnail count after removal, close and reopen the preview, then
remove the final photo and assert automatic return to Add with its empty photo
chooser. Dismiss Add to the audit root without saving. Run it in both the complete
fixture suite and the focused Add-draft suite; recording the scenario is not a
native pass.

Full-screen photo viewing must allow a single tap to reveal or hide commands
without resetting the image's zoom/pan. Keep double-tap zoom distinct from single
tap, cancel delayed tap work on unmount or image change, and preserve platform
system dismissal. This adopts the Photos interaction documented in Apple's iOS26
user guide (https://support.apple.com/en-sg/guide/iphone/iph3d267610/26/ios/26);
retaining zoom while revealing commands is a project usability requirement, not
a quoted HIG rule. Acceptance covers both asset and Add previews: zoom, reveal
commands, close without changing draft/media, double-tap without an extra toggle,
and image-change/unmount without delayed effects. Native verification must cover
iOS and Android separately.

Zoom gestures change image scale only; a single tap controls toolbar visibility.
Newly selected images and loading failures restore commands. Keep visibility in
a stable controller so showing commands does not remount or reset the image.

Photo viewer footer labels and commands must remain readable over any image,
including a zoomed white region. Use the viewer's opaque neutral canvas behind
the complete footer, not only behind its command row; do not rely on image
brightness or a black letterbox for text contrast.

Photo scale, pan position and gesture ownership must survive unrelated React
rerenders for the same image and viewport. Reset them only for an actual image
or geometry change, explicit zoom reset or viewer dismissal. Pending gestures
retire with their owning image; updated callbacks must still reach the current
consumer. Verify retained scale through the installed Android responder as well
as a delayed native capture.

Asset-detail overflow presentation must remain stable across callback-only renders,
just like native header command actions. Its identity changes for title, action
eligibility or disabled state; retained handlers dispatch only the current committed
callback and do nothing after disabling, removing eligibility or unmounting.
Loading/error transitions explicitly clear the installed menu. This prevents
incidental reconfiguration alongside native search; it does not establish M207's
runtime cause. Preserve phone/iPad search-placement acceptance unchanged.
Menu callback ownership is scoped to the serialized asset resource key. Replacing
that resource retires retained native handlers even when the replacement has the
same title and permissions; current callbacks are forwarded only within one scope.
Native search enabled after an unavailable/loading state must initialize its native
field from the current retained query. Disabled fields receive no imperative text
updates. Native edit echoes must not rewrite text during typing; initial enabling
and explicit application query changes are separate synchronization cases.

Move Here native acceptance must activate the accessible candidate row as a button.
Run350950 phone/iPad hierarchies expose `Audit tent, Item, Garage` as one grouped
button, with no standalone `Audit tent` StaticText. Select that observed native
label and require existence, enabled state and hittability before tapping. Preserve
the five-second candidate wait, rejected-command draft retention, retry and positive
return assertions. This corrects test target semantics, not a product recovery fix;
the amended journey remains unverified until native execution.

## Native inbox acceptance fixture

Runner-only notification fixtures reuse the production inbox and application query
with an in-memory notification repository. Include long titles, month/day dates,
read/unread rows and location trails. A separate scenario rejects its first mutation
with permission denied so native acceptance can verify metadata removal and fresh
read recovery. Inspect actual accessory geometry and activate read/unread, mark-all,
filter, item and breadcrumb navigation with return. Fixture destinations label the
resolved target; they do not represent production item/settings content. This
verifies presentation and interaction, not server authorization or APNs delivery.

## Android native menu trigger bounds

The accessible menu-trigger wrapper must size to its native Compose control rather
than stretching into neighboring blank space. Its announced button bounds must
contain an actionable center. Retain native menu selection, labels, disabled state
and compact icon geometry. Validate by tapping the accessibility target center,
not by substituting a text-child coordinate when the declared button misses.
Apply to shared NativeActionMenu consumers; this is a geometry correction, not a
change to menu option or navigation semantics.

Selectable Android menu rows must retain DropdownMenuItem's native onClick event
alongside selected-state accessibility semantics. The pinned Compose component
always owns that click dispatch; a selectable modifier does not replace its event.
Verify actual taps select System/Light/Dark and close the menu. Disabled choices
remain inert and ordinary command items retain their existing activation path.

The controlled settings fixture may display successful appearance-store write count
to prove one save per native selection, including selections with both native click
and accessibility modifier callbacks. This counter is fixture-only.

## Unfinished-tag native observation

Run351121's phone Add-tag assertion reads Camp immediately after typing Camping;
its final screenshot and hierarchy both retain the exact full Camping value. The
unfinished-tag journey must allow at most five seconds for exact equality before
continuing keyboard dismissal, Save blocking, disclosure collapse/reopen and draft
retention checks. Do not weaken those downstream assertions or apply this diagnosis
to controlled-field cases whose final captures still contain incorrect text.

Run351318's phone collapse assertion similarly precedes the collapsed hierarchy.
Observe entry removal and collapsed guidance within five seconds, then observe
the exact retained entry value within five seconds after reopening. Preserve
Save blocking, staging and clear assertions; do not retry the tap or omit the
downstream journey. This is bounded observation of asynchronous UI completion,
not permission to accept a disclosure that fails to finish or loses its draft.

## Managed search and native action coexistence probe

Phone run351121 retains bottom search in Settings while isolated managed search
and production Place search pass. The next controlled comparison must add a native
header action after the existing enable/title-update stages, then verify search
remains in the header and the action activates. This isolates action registration
as a possible contributor without claiming it is the cause. Preserve the earlier
stages unchanged and make no production placement workaround until native evidence
supports it. The fixture action changes only an observable activation count.

## Android color gesture ownership

Retained native color gesture, adjustment-button and accessibility callbacks must use the latest
committed color state, selection callback and enabled state. Disabling or unmounting
the picker retires mutation delivery, including a drag already in progress. Handler
retention must not target an earlier parent callback or compute hue/spectrum changes
from an obsolete value. Verify through the real mounted control with retained
native callback references, then recheck ordinary native dragging and cancellation.

## Android color target sizing

Android color swatches, clear/custom controls, hex input, hue strip and adjustment
buttons must provide at least the shared48dp minimum target, following
https://developer.android.com/guide/topics/ui/accessibility/apps. Existing44dp
fixed boxes do not acquire expanded hit areas automatically in these React Native
controls. Preserve iOS swatch sizing. Verify actual native bounds and actionable
centers at320dp width with normal text, including wrapping and Cancel/Done reachability.

At320dp width, adjustment labels must remain readable without forced mid-word
splitting beside fixed controls. Present each label above its decrement/value/
increment row, keeping the targets full size and the whole editor scrollable.

## Android compact control target probe

Runner-only settings controls may expose isolated real refinement, icon-menu and
ellipsis-menu adapters with observable activation counts. Use native bounds and
actual center/edge taps to determine whether Compose expands declared44dp hosts to
48dp targets. A source literal alone is insufficient to classify native target
failure. This probe must not alter shipped routes or the iOS settings-control fixture.

Confirmed compact Android refinement and menu hosts must provide the shared48dp
minimum, including their React Native accessibility wrapper and Compose button.
Labeled hosts also reserve48dp vertically to avoid constraining native targets.
Keep the iOS adapters unchanged. Verify native bounds and exactly-once center/edge
activation for refinement, sort and overflow commands; inspect Browse and expiration
header consumers and labeled choice menus for clipping.

The runner-only Android header composition probe must mount the production Browse
result-tools row and expiration workspace, including its actual native header
filter slot. Verify48dp targets, commands and absence of clipping at normal text
on320dp and ordinary phone widths. Isolated adapter evidence cannot replace this
consumer check. Keep fixture routes out of production builds.

## Android native command descriptions

NativeSheetActions must preserve each caller's explicit accessible action name,
including expiration filter and voice-review context, while keeping its visible
label, native button role, enabled state and activation semantics. Android's
pinned Expo UI55.0.17 does not expose a Compose content-description modifier.
A reviewed, version-specific pnpm patch may add only that modifier to its native
registry. The project adapter owns the small serialized modifier mapping; no
custom button, React Native accessibility overlay, or dependency upgrade is needed.
Apply Compose semantics without clearing native role/click/disabled semantics.
Follow https://developer.android.com/develop/ui/compose/accessibility/semantics.

Native regression evidence must first demonstrate the missing caller description,
then verify the named native buttons, enabled/disabled behavior, and unchanged
visible labels in expiration and voice/filter consumers. Keep TalkBack speech
acceptance separate from native accessibility-tree inspection. The patch and lock
hash must be committed together and installed through frozen-lockfile validation.

Android autolinking must build the patched `expo-ui` Gradle project from source;
its bundled precompiled Maven artifact does not contain the registry correction.
Scope `expo.autolinking.android.buildFromSource` to that project only. An APK
build that silently reuses the precompiled artifact is not acceptance evidence.

## Native return and delivered-target acceptance

After Add photo dismissal, the audit-index destination must remain usable. Its
probe must choose scroll direction from the target's position relative to the
visible scroll frame, not assume a fixed ordering of growing fixture entries.
Run351214's Browse entry lay below the viewport while the probe swiped downward.
Retain the destination, hittability and running-foreground requirements.

For notification read commands, distinguish native glyph accessibility frames
from delivered hit regions. A24-point icon frame alone neither proves nor disproves
a44-point target. Verify center, four edges and four corners at21-point offsets,
with each real tap producing the intended read-state transition and retaining
the same row. Preserve the full inbox read/unread/navigation workflow separately.


### Native audit execution budget

The full iOS audit job allows120 minutes for dependency setup, compilation,
interaction tests, result-bundle finalization and evidence export. Run35121454700
finished85 iPad tests after roughly91 minutes of total job time but exceeded the
former90-minute limit before the result bundle finalized; its screenshot artifact
was lost. This observed failure justifies the larger bounded budget. Keep all
assertions and failure reporting; a longer job budget is not a test pass. Active
runs retain their original configuration and must not be cancelled or restarted
solely to adopt this change. Check future runs for completed result export and
uploaded evidence as well as test totals.


When a pnpm patch changes a native dependency's installed location, update both
its dependency entry and external-source path in the checked-in iOS Podfile.lock
before native validation. Keep pod versions and checksums unchanged when the
podspec itself is unchanged. The Android-only Expo UI patch changes pnpm's directory
identity on every host, including iOS; its ExpoUI path must follow the frozen
installation. Preserve `pod install --deployment` as the native lock-consistency
gate. Run35130374705 demonstrates the failing path contract before this correction.


Runner-only text-entry comparisons may retain a bounded in-memory event trace to
localize observed character loss. Record native change text/event counts, selection
ranges, and committed React values without scheduling additional renders for each
trace event. Publish only when XCTest explicitly requests a snapshot after typing.
The trace uses synthetic fixture text, is not production logging or telemetry, and
must not change input assistance, value ownership or normal exact-text assertions.
Capture the input value and expected React-observed state before publishing so the
trace button cannot turn blur-induced correction into a passing entry result.
Tracing can affect timing; a trace is diagnostic evidence, not proof of a root cause
or a substitute for the unchanged production entry workflows.

The input-comparison scroll container must deliver handled taps while the keyboard
is visible, so its explicit trace command can publish without a preliminary blur.
Other fixture containers retain their existing behavior. A delivered capture command
must produce its trace within the bounded observation window; missing output must
fail the diagnostic rather than silently skip it. Run35140471580 supplies the
pre-correction evidence: capture taps dismissed the keyboard and no trace was exported.
The focused text-entry selection includes controlled, seeded and system address
comparisons as well as ordinary-name comparisons. It remains diagnostic-only;
the full suite is still required for broader native acceptance.

The runner-only input trace also records React Native key-press events before
text-change events, without updating React state while recording. This separates
received replacement/key events from resulting text values in run351480's missing
character cases. These events are still framework-delivered evidence, not physical
keystroke proof. Keep the bounded buffer, explicit snapshot command, assistance
settings and exact-text assertions unchanged. Verify event order, absence of
trace-induced renders, and bounded retention using the existing render harness.

## Filter footer and keyboard accessory clearance

Native filter commands must remain fully visible above the keyboard and the app's
keyboard-dismiss accessory, not merely accept a center tap through an overlay.
Run35140471580's normal-size phone expiration-search capture shows Back at
Y491–545 while Dismiss keyboard occupies Y485–529. The old hittability-only
assertion passed despite visible overlap; preserve this as a native finding.

The shared sheet boundary calculation must remeasure on settled iOS keyboard
frame/show notifications as well as the anticipated frame change. A native sheet
or input accessory can finish layout after the initial notification. Continue
measuring the unmoved sheet boundary, reject superseded callbacks, clear on hide,
and retain nonoverlapping/floating-keyboard behavior. Do not add an assumed fixed
accessory offset without native geometry evidence. This lifecycle correction is
a candidate, not proof that all overlap causes are resolved.

Strengthen native filter acceptance to require both entire action frames above
the delivered Dismiss keyboard control after layout settles, then exercise Back.
Review Browse and expiration consumers of NativeFilterSheet; Android's separate
IME-aware adapter must keep its existing behavior. Keep normal-size failures ahead
of enlarged-text work and retain the failing capture even if an automation tap
was delivered successfully.

The focused `filters` native selection must include Browse tag search with a
visible keyboard, the last-row selection/apply journey, the in-place availability
menu, expiration date/calendar dismissal, and expiration search with keyboard.
Both searchable consumers must use the same full-action/accessory clearance
assertion, preserve the complete query, and exercise the intended draft/result or
Back behavior. Focused results cannot replace full-suite release acceptance.

After run35154627907, collect fixture-only sheet geometry before another inset
correction: the unmoved bottom edge's measureInWindow coordinates, delivered
keyboard frame and resulting shared boundary calculation. Compare these with XCTest
screen frames. The phone footer ends 62 points below the keyboard container top,
matching the presented sheet origin; this is a coordinate-space hypothesis, not
yet a proven cause. The probe must not change production layout or keyboard
policy and must ignore late measurements after hide or unmount. Install the probe
only for the focused filters selection; full audit routes retain their unmodified
fixture presentation. The probe is an independent event-time sample, not a trace
of the production hook’s applied inset or synchronized XCTest measurement.

Calendar dismissal acceptance must tap the visible dismissal region outside the
expanded date-picker frame. A full-screen PopoverDismissRegion AX element's
default center can lie inside its foreground calendar (phone351546: center201,437
inside calendar46.3,160,320,332). Derive an outside point from current frames,
retain both frames and the chosen point, and require the popover to disappear
before testing the sheet commands. Do not change production dismissal behavior
to accommodate a test that taps the calendar itself. The target must also avoid
underlying commands: iPad351595's outside point landed on Back and returned to
the overview. Prefer the noninteractive part of the current Date range navigation
bar outside the popover, exclude button/title bounds, and assert Date range remains
present after dismissal before exercising Back. Fail with geometry evidence if no
safe target exists; do not guess a device-specific coordinate.

Searchable filter pages must have one owner of the native search options. Updating
the title or selecting a tag must not clear the active search field or its query.
Mounted acceptance must inspect the currently merged navigation options, not a
previous truthy search configuration that may already have been removed. Returning
to the overview must remove search through the search owner's disabled state.

Native keyboard readiness must ignore absent, empty or non-finite key frames
before asking XCTest for hittability. A transient placeholder key must not abort
the comparison before typing; the existing bounded wait must still fail if no
visible interactive key becomes available. This is runner readiness, not a text
entry workaround or a reason to pace ordinary typing.

After staging an Add tag, native acceptance must observe the staged tag, an
existing cleared entry and an enabled Save action. An existing SwiftUI text field
may expose its empty value as nil, an empty string, or its placeholder. Never
treat a nonexistent field as successful clearing; observe the complete staged
state with a bounded wait before continuing to Clear draft.

The runner-only `text-entry-no-provider` diagnostic must execute the same fourteen
text-entry comparisons with AppKeyboardProvider and its accessory entirely absent.
Hiding the accessory alone does not exclude the provider's native input hooks.
The default/full fixture installation must retain the provider. Record actual
provider omission in native evidence; keep exact text, mirror and trace checks,
ordinary typing speed and system-field comparisons unchanged. This comparison is
not a production configuration change or a substitute for full-suite acceptance.

When ordinary color opening misses its existing five-second native observation,
retain the failed state and observe a further bounded fifteen seconds for late
presentation. Record whether the system picker eventually appears, but keep the
original five-second result as the acceptance assertion. Diagnostic waiting must
not silently convert a slow or failed activation into a passing test.

## Bounded release batches during the comprehensive audit

Freeze each release's production source cutoff, changed workflows and critical
regression checklist before acceptance. The comprehensive audit remains a separate,
ongoing workstream; completing every audit finding is not a release prerequisite.
Do not add unrelated remediation to a frozen batch. A correction needed to satisfy
its existing acceptance may extend the cutoff with explicit source/evidence mapping.

Release when the changed workflows and critical regression checks pass. Keep
unrelated existing findings, enlarged-text follow-up and diagnostic-only experiments
tracked separately. A failing diagnostic requires triage against the batch's real
workflow, not automatic promotion to a release blocker and not automatic dismissal.
Security, data loss and failures of required changed workflows remain blockers.
Record the reason and evidence for every failed check's inclusion or exclusion.
Preserve native evidence limits and publish the batch's own TestFlight changelog;
Apple processing and notes readback complete the release.

The frozen batch's last-tag acceptance must observe the row's checked value after
its delivered tap and before Back/Apply. Run351546 iPad delivers touch down/up at
372,700.5, inside the row98,674.5,548,52, yet applies no tag. Retain the selected
state and hierarchy at that boundary to distinguish selection delivery from
navigation/draft loss. Do not infer a clipping defect from the final result alone.

## Invitation email native editing

Phone351567 reorders the complete invitation email in the real Sharing workflow,
even with a seeded uncontrolled React Native field. Use the existing iOS SwiftUI
TextField pattern for this field, with email keyboard/content type, no automatic
capitalization or correction, and the Invitee email accessibility name. Keep native
editing state across parent echo renders; remount only for the existing scope/reset
revision. Failed creation preserves the full email; successful creation resets it.
Disable edits and ignore stale native changes while creation is pending. Android
retains its controlled platform field. Preserve invitation permission checks and
command ownership. Mounted seed/lock/revision tests and existing sharing recovery
tests precede implementation; the native Sharing journey must still verify exact
complete entry, rejection recovery and navigation reachability before release.
This is a field-specific candidate, not a claim to solve all RN input failures.

## Native fixture title parity

The settings collection audit route must use the production Tags navigation title,
not its internal audit-customization route name. Navigation title width competes
with native header controls; a placement failure with different header content is
not sufficient evidence of a production regression. Assert the title in the native
journey before Add/Search and guard fixture/production title parity structurally.
Keep the integrated-button and query/result checks unchanged. This correction does
not explain other search failures whose titles already match; rerun rather than
claiming a production fix from the title change alone.

## Focused release-correction verification

The native audit workflow offers a release-corrections selection containing the
five existing filter journeys, complete Sharing email/rejection recovery, settings
Add/Search with production title parity, Add draft/rejected-save recovery and
ordinary color activation/clear. Run it on phone and iPad with ordinary fixture
layouts; fixture-only geometry probes stay confined to the filters diagnostic.
This selection accelerates changed-workflow verification but does not replace the
frozen55-check matrix, onboarding, retained unchanged-workflow evidence or source
checks. Keep all current native jobs intact. Verify the selector's exact test list
and that every selected Swift test exists before dispatching it.

### Bound individual native audit cases

Enable XCTest test timeouts for all native audit selections and cap each case at
600 seconds, matching Apple's default enabled allowance. Run35159542174 stalled
after an app-launch timeout until the120-minute job limit, leaving an incomplete
result bundle. A stalled case must report failure rather than silently consume the
whole run. This does not guarantee recovery from an unresponsive Xcode process or
replace the outer job timeout. Preserve all assertions, case selection and failed
results; do not retry cases until they pass.

The longest passing case in the retained351567 phone/iPad logs took363.920 seconds,
so a600-second maximum leaves room for the existing complete journeys. This is an
execution guard, not a performance requirement. Active/queued runs retain their
original configuration; do not cancel or restart them to adopt the guard. Verify
future runner acceptance and result export separately from configuration checks.

Apple documents enabling timeouts with `-test-timeouts-enabled YES` and the maximum
allowance option in [executionTimeAllowance](https://developer.apple.com/documentation/xctest/xctestcase/executiontimeallowance).

### Calendar dismissal uses the presented navigation bar

Run351652 iPad retains a visible calendar and Date range page, but the dismissal
target calculation rejects every blank navigation-bar point because it includes
fixture-launcher buttons behind the modal. Scope command exclusions to the Date
range navigation bar, whose bounds already constrain every candidate. Continue
excluding the calendar and navigation title; retain geometry before unwrapping a
target. Calendar dismissal must leave Date range open, with Apply and Back
reachable, and Back must return through Filters to the launcher. The observed
XCTest failure is the regression baseline; this automation correction does not
change product code or establish a native pass until rerun.

### Observe settled native search entry

Phone351652 reports T immediately after typeText(Tools), but both retained
Browse and Expiration teardown hierarchies contain the full Tools query and
matching result. Those two journeys must wait up to five seconds for the exact
native value before checking results and keyboard clearance. Do not retype,
accept partial text or skip the footer assertions. Keep the original failure and
settled hierarchy as evidence; this addresses asynchronous observation, not a
claimed product input fix.

## Native keyboard-window sheet boundary adapter

M249 probe351689 confirms that Fabric's measurement omits the sheet presentation
origin. The iPhone boundary reports812 instead of874 and the317-point inset leaves
Back overlapping the keyboard accessory. This concrete limitation justifies a
narrow native measurement adapter, not custom platform controls.

The pinned RN0.83.6 RCTKeyboardObserver converts native keyboard frames from screen
space into RCTKeyWindow coordinates before publishing KeyboardMetrics. The names
screenX/screenY do not establish screen coordinates. Return the actual boundary
view's frame in its own UIKit window only when that window is RCTKeyWindow. Return
unavailable for detached or other-window views; do not measure a different global
window as a substitute. Never add a device-specific presentation offset. This
window-identity check is necessary to match the producer's coordinate contract.

Expose a typed asynchronous measurement port with an explicit unavailable result.
The existing sheet overlap policy preserves generation guards, hide handling and
settled remeasurement. Keep the measurement view fixed at the unmoved boundary,
not the footer whose position depends on its result. Unavailable or rejected
measurements clear stale insets. Android retains its existing adapter.

Use a local iOS Expo module with the pinned Expo55 and ReactNative dependencies
already provided by the app. Native modules must be autolinked and represented in
the reviewed pod lock before release. The view command executes on the UI thread.
Do not silently fall back to the known incorrect Fabric coordinates.

Acceptance covers translated phone sheets, centered/windowed iPad sheets,
rotation/resizing, keyboard hide/show and late completion after unmount. Native
Browse/Expiration queries, selection and full Apply/Back frames must clear the
keyboard accessory. Pure overlap and mounted adapter tests do not establish
runtime success. M249 stays open until native coordinate/interaction checks pass.

References: [Expo native view commands](https://docs.expo.dev/modules/module-api/#view)
and [local native modules](https://docs.expo.dev/workflow/customizing/).


## Isolating intermittent native color presentation

M51 ordinary first activation remains intermittently absent despite a hittable,
enabled well and successful same-build touch-region checks. Diagnostic comparisons
must use independently launched apps, identical initial unset value, the same
native element tap and five-second presentation assertion. Compare a capture of
the pre-tap screen/hierarchy against no pre-tap capture; record target geometry,
enabled/hittable state and post-tap presentation. Preserve the original failing
workflow. Do not add retry taps or relax its timeout to manufacture acceptance.
These comparisons diagnose observation effects; passing samples alone cannot
close the intermittent production finding or establish causation.


Native runner evidence must export diagnostics directly from an existing xcresult
bundle as well as screenshots. Post-run collection from only booted simulators
can produce no application logs after Xcode shuts down a test simulator; an empty
collection is missing evidence, not proof of no UIKit warnings. Preserve export
errors as artifacts and report them without replacing the native test result.
Use the pinned Xcode xcresulttool; no new dependency or simulator restart is needed.


M51 scroll-interaction diagnostic: retain the normal ScrollView fixture and add
an independently entered variant differing only in scrollEnabled=false. Preserve
content geometry, native picker, initial unset value, ordinary first tap and
five-second assertion. This isolates scrolling behavior, not all scroll-view
internals; a pass does not justify disabling scrolling in real editors. Expose
the variant in the audit launcher only and retain both test outcomes.


M207 search attachment diagnostic: compare the pinned adapter with a runner-only
transformation that applies preferred search placement and toolbar integration
before assigning UINavigationItem.searchController. Retain all source option
values and existing search/navigation assertions. The production dependency patch
and locks remain unchanged. The transformation must refuse unexpected source and
be opt-in through a named workflow selection. Test Place, preconfigured Place,
managed/static placement, settings and filter consumers. A successful sample is
candidate evidence, not comprehensive acceptance or proof of causation.


Native search-button acceptance must verify that tapping Search itself presents
and focuses a usable keyboard before typing. Do not add an unconditional second
field tap as preparation: on an already focused field it can open the native
editing menu and changes the interaction under test. Preserve keyboard readiness,
complete query, filtering and command-clearance assertions. Retain evidence of
any failure instead of claiming the extra tap is proven causal. Provide an
unmodified-adapter selection matching the attachment-order diagnostic journeys
so the same source/test suite can compare the native dependency variants.


After clearing native search, acceptance follows a hittable input or a hittable
collapsed Search control. It must not require UIKit to remove a hidden field from
the accessibility tree. Select the hittable route and verify keyboard readiness,
full re-entered query, filtered results, cancel and navigation return. Do not
replace interaction checks with screenshot appearance or extend timing limits.


Apply the single-activation search acceptance rule consistently to Browse tags,
Place, Settings, static placement and voice location search. Native352287 captures
show the same AutoFill editing menu in Browse and Place after redundant field taps.
Retain ordinary text-input focus taps and explicit keyboard-dismiss/re-entry
coverage. When a search field remains usable after Clear, tap it only if no keyboard
is present; otherwise verify readiness and type. Hidden-field presence is not a
substitute for usable controls or query correctness.


Voice location Clear/re-entry must use the same hittable-input-or-collapsed-button
contract as Place on both device classes. Retain restored locations, full query,
selection, navigation return and unchanged timeout. For unresolved keyboard or
footer checks, attach the individual readiness predicates on failure; screenshots
of visible controls alone do not prove hit-test availability.


Native352366 timing diagnostic: retain the five-second keyboard readiness and
exact-query acceptance deadlines. Record monotonic duration and result for each
predicate evaluation, then attach observations after the wait. Do not add sleeps,
retry interaction, alter keyboard settings or convert eventual state into a pass.
A focused runner-only selection may execute Browse and Expiration filter search
with the attachment-order candidate to isolate these unresolved checks.


Filter-search readiness uses the known English fixture keyboard's t key rather
than enumerating every key: native352409 proved a single enumeration could consume
the entire observation interval. Retain finite/nonempty bounds, hittability, the
five-second waiter and full Tools-query/results/command checks. Other keyboard
journeys retain their existing selection. Failure diagnostics inspect only the
keyboard and selected key, avoiding an additional exhaustive query that itself
failed in352409. This is test infrastructure, not a localization product rule.


M207 production candidate: pin a minimal react-native-screens4.23.0 patch that
applies UINavigationItem preferred placement and toolbar-integration settings
before attaching its searchController. Preserve option values and all platform
availability guards; do not replace native search or change query state. Native
352366 accepts the other six journeys on both devices;352442 accepts Browse and
Expiration typing/filtering/action clearance on both after targeted-key readiness.
Freeze this correction as a separate release batch, requiring the installed patch
to pass the eight search workflows, dependency resolution, shared-header tests and
critical regression checks before TestFlight. Remove the runner-only transform
from candidate validation so it cannot apply twice or hide packaging mistakes.

### Collection geometry acceptance

Measure card width on its outer layout container, not an inset title action.
The iPad combined run 35871152122 measured the title action at 209 points;
AssetCard applies horizontal body padding inside the card, so that observation
cannot establish a violation of the 220-point outer-card minimum. Expose a stable
layout test identifier without grouping or hiding its independently accessible
child actions. Keep the same minimum and column-position assertions on the outer
container. This corrects the observed object rather than relaxing the product
requirement. Native rerun remains required for the corrected measurement.

## Browse scroll gesture acceptance

Run35879482924 passed corrected outer-card geometry on iPad, but the subsequent
synthetic list swipe left Browse for the unmatched root route. This does not
establish a header layout regression. Its previous assertion incorrectly accepted
an absent item after route departure as proof of scrolling. Keep the Browse control
present before judging movement, and originate the list drag in the visible gutter
between the first two cards, away from card navigation targets. Preserve anchor,
actual content movement, Map and return assertions. This one distinguishing run
checks accidental gesture activation versus a repeatable scroll/navigation defect;
do not start keyboard/provider experiments or alter production layout without evidence.

### Connected Browse refinement verification

Verify Browse → Filters → applied results → asset detail → Back, and Browse →
Filters → Expiration → asset detail → Back → Browse using the production route
controllers and navigation. Fixture callbacks alone cannot establish return
context or scroll restoration. Keep application queries injectable at route-screen
boundaries; the production route entrypoint only binds AppServices. Runner-only
fixtures may replace query repositories with deterministic in-memory reads and
must preserve scope validation, query keys, route parameters and commit/cancel
semantics. Use actual matching and filtering in those repositories, not constant
responses that make every selection appear valid. Verify fixture read isolation
and requested filters, then inspect normal-text phone/iPad workflow captures.

### Native search interaction ownership across return

Connected Browse refinement on Android d7d7689b preserved Availability but lost
Camping on return (19 unscoped items instead of18 matching items). The native
adapter currently treats empty text/close callbacks as user commands even when
no search interaction is active. A mounted/focused screen alone does not prove
that a search callback represents editing.

Native search must begin an editing interaction on native focus/open, seed the
retained query, and retire that interaction when the screen loses focus, search
is disabled, or the user closes it. Ignore text/submit/close callbacks outside
that interaction. Within it, typing, deliberate clearing and cancellation retain
their existing semantics. Navigation return must preserve query, filters and
scroll context. Verify the shared adapter with representative Browse and Move
consumers, then the connected native Browse refinement journey; source tests do
not establish the Android root cause or native acceptance alone.

Connected filter native verification must scroll the actual filter body when an
entry is below its visible viewport, retaining hittability before activation.
Run35905242914 confirms the iPad Expiration entry is below the fixed footer; its
check stopped before any scroll. This is not evidence that scrolling fails, nor
acceptance of the sheet's density. Preserve M275 visual review independently.

### Initial filter task size

M275 iPad evidence at cfe9f137 confirms removing duplicate picker padding alone
still leaves Sort and expiration below the opening viewport while both bottom
actions dominate the short sheet. Browse and Expiration filters must initially
use the native large detent, keeping the smaller detent available for deliberate
resizing and retaining the user-requested bottom results/cancel actions. Do not
change unrelated asset sheets or Android stack presentation. Verify ordinary
filter choices and the fixed actions in the initial normal-text phone/iPad view;
retain explicit medium-to-large recovery coverage as a separately configured
fixture. A larger sheet does not by itself certify hierarchy or accessibility.


### Integrated Add destination search observation

Run35941028517 passes15/15 iPad workflows and14/15 iPhone workflows. The remaining
phone failure occurs after the complete `Audit shed` query is entered: the captured
state has no keyboard and a reachable New place command, but the fixture requires
an unconditional Dismiss keyboard tap. Reuse the native-search dismissal helper:
wait for either keyboard absence or a reachable dismissal action; dismiss only
when needed; assert the keyboard is absent before continuing. Keep exact query,
draft retention, creation cancellation, rejection/retry and returned-parent checks.
This is an observation correction, not evidence of a new application fix. Rerun
the affected Add/Move selection subset; preserve the other fourteen passing phone
and all fifteen iPad results without claiming whole-app acceptance.

## Settings save reads the committed draft

Native35945430640 passes the edited-row return geometry on iPhone, but creation
persists Campin after the native field reported Camping. The retained final
collection confirms truncated persisted content, not a missing row or refresh.
Treat this as a data-loss defect. Test the concrete stale Save-handler hypothesis
before changing text providers: a Save event retained before the final edit must
submit the latest committed valid draft, and must reject invalid, pending,
unfocused or unmounted state. Reuse the focused committed-action guard for the
Settings save command. Do not weaken exact native text/persistence assertions or
insert typing delays to make this journey pass. This source correction does not
prove the native cause; the existing readback journey must still pass unchanged.

### Expiration search ownership across tab return

Native Android evidence shows the Home expiration route retaining its identity
while its Kitchen query becomes empty on tab return, before Filters opens. Native
close/empty callbacks outside a search interaction must not clear retained queries.
Expiration must reuse NativeNavigationSearch, whose editing guard owns native
callbacks, while its hook owns debounce and filter-flush semantics. Deliberate
clear/cancel continues to clear the query.

Keep the displayed draft separate from the debounced applied query. Switching tabs
before debounce completes must restore the draft on native focus. Verify this with
the shared adapter and its native field driver, not only the debounce hook.

The native journey enters expiration through in-app actions on each selected tab,
asserts independent Home Kitchen and Browse Camping results before opening Filters,
and preserves both after Apply. External-link setup is separate coverage; this
journey does not certify qualified deep-link behavior. Diagnostic overlays must
not ship. Preserve the existing origin-tab modal return contract.

### Native tab accessibility duplicates

Run35955427236 passes final Settings-tab return on iPhone. The iPad capture shows
both tabs, but the native accessibility tree exposes nested duplicate Home/Browse
buttons with identical bounds; selecting the first match is not proof of the
interactive control. Within the native tab strip, choose a hittable matching
button when available, retaining the first match only to report failure if none
is hittable. Keep the final reachability assertion and add actual Browse/Home
activation with exact Settings draft readback. This observation correction does
not by itself establish iPad task acceptance or justify a production layout change.

### Photo-free grid native comparison

The existing Browse journey fixture may accept photoMix=true to show one real local
image, one photo-bearing asset without a resolved thumbnail, and confirmed photo-free
peers. The normal fixture stays photo-free. Review both grids with List/Map return and
scrolling; source geometry assertions alone do not accept mixed-row density.

The focused Browse native suite verifies compact confirmed-empty card geometry and
then mixed-row title alignment on the same route with a changed fixture variant.
It retains existing responsive width and List/Map scroll-return checks. Photo-free
card height must be smaller than its grid width at normal text; mixed media cards
retain the image area. Capture both states for visual judgment.

### Browse density observation prerequisites

Run35962300873 passes phone Browse checks but the compact twelve-card iPad fixture
fits entirely without scrolling. The scroll-retention scenario must use a distinct
36-asset variant with stable identities; keep sparse and mixed-photo density cases
unchanged. Require actual list movement and retained control placement. The focused
browse-journey selection must include the mixed-photo comparison itself; its prior
omission means that run provides no mixed-photo acceptance evidence.

### Bounded Settings commands

SettingsActionRow issues a command rather than navigating or selecting a value.
Use the shared bounded secondary NativeCommandButton within the existing grouped
row, preserving concise visible labels, optional descriptive accessibility labels,
disabled guards and destructive roles. Preserve navigation/choice row conventions;
do not add button borders to those rows. Group padding must accommodate the native
button without truncation or nested press targets. Apply to filter resets, reminder
retry/discard, device-settings commands and voice profile actions. Verify shared
source consumers and representative native filter/reminder screens before release.
NativeCommandButton accepts a separate accessibility label, defaulting to its visible
label, consistently on iOS, Android and the preview renderer.

Native acceptance for bounded Settings commands includes a runner-only reminder
recovery fixture using the production editor: the first save fails, Retry commits
the retained choice, and Discard restores the saved choice. Include the longer
Open device settings label for layout/activation observation only; a fixture
activation is not evidence that OS settings opened or push permission changed.

Settings dividers align with the following row's content: ordinary text, choice,
switch and command rows use the shared row inset; navigation rows with leading
icons explicitly include the icon width and row gap. Do not apply an icon gutter
to icon-free rows. Stacked layouts retain the ordinary inset. This changes visual
grouping only, preserving interaction and accessibility semantics.

### Native command content sizing

The paired reminder recovery fixture at a0874586 exposed iOS bordered buttons
whose 48-point label minimum acquired native style padding, yielding 62-point
hit frames inside 48-point measured hosts. Retry and Discard overlapped. Keep the
48-point minimum on the outer command, but use the same 32-point label minimum
as primary commands so native padding fits the measured command. Preserve native
bounded styling, multiline growth, and disabled behavior. Acceptance requires the
existing paired recovery test to prove separate hit frames and successful Retry
and Discard on iPhone and iPad; source tests cannot establish native geometry.

### Integrated native batch completion observations

Run35958732480 passes12/13 phone and11/13 iPad workflows. Both pass Add and
independent expiration search return. The phone Retry contents assertion expires
at five seconds, but its final native capture shows Nothing inside yet and no
Retry command. Observe the combined completed state for a bounded fifteen seconds;
require both the command's disappearance and the recovered content without a
second tap. This changes observation tolerance, not the production completion rule.
The focused iPad run35960677440 still fails after the duplicate-element
correction; that correction does not close tab reachability. Its Settings creation reads Camngpi immediately after typing Camping;
this remains an exact-text failure. Observe exact committed text for fifteen seconds
without retyping or replacing it, then retain the exact save/reopen checks. A lasting
wrong value still fails. Do not reopen provider/key-delivery experiments or claim
these observation changes establish native acceptance before the focused run passes.

### Distinguish tab touch and accessibility reachability

Focused run35960677440 retains the exact Settings draft and visibly renders both
iPad tabs, but neither the first-match correction nor waiting establishes native
hit-test reachability. Before the final accessibility assertion, activate the visible
native Browse and Home buttons at their observed centers on iPad, requiring Browse
content and exact draft readback. Derive coordinates from finite on-screen native
button bounds, never fixed screen coordinates. Capture before and after activation.
Keep the separate final isHittable requirement: successful coordinate activation does
not prove accessibility acceptance. A failed destination change identifies a touch
failure; a successful roundtrip with failed hit-test keeps the narrower accessibility
finding open. The focused tab-return selector runs this workflow alone.

### Settings name input retention

Run35961716802 passes contents Retry on both devices and Settings save/readback
on iPad, but iPhone New Tag retains Cngampi after typing Camping and waiting15s.
Do not extend the observation budget or alter typed text. Use the existing
DraftTextField adapter for customization display names (tags, asset types, fields):
iOS owns its editing text and publishes complete drafts; Android preserves its
current adapter. Keep the current validation/name-to-key behavior, busy/access
locks, failed-save retention, tab return and exact save/reopen verification.
Loaded-resource transitions must unmount the old field before seeding a new name.
Multiline descriptions and the explicitly focused technical key remain unchanged.
The native failing workflow is the regression reproduction; source tests cannot
establish correct iOS character order. Rerun Settings readback and persistent-tab
draft return after this correction.

Editing an existing customization definition must preserve its saved stable key
when the native Name field reports either its initial value or a later rename.
Automatic key suggestions apply only during creation until manually overridden.
Replaying an unchanged loaded name must leave the editor clean; renaming and then
restoring the saved name must also restore clean state. Back must not ask to discard
changes solely because a native field appeared. Verify this with a saved display
name that differs from its original key, then retain the native save/reopen gate.

### Resolve native tabs without transient ancestor chains

Run35970467285 passes Settings creation/save/reopen on both devices and the
iPhone tab workflow. iPad completes actual tab switching and exact draft return,
then fails the hit-point waiter. Its log repeatedly retries resolution of Other
ancestors before failing to find Browse, while the final tree retains both native
buttons. The smallest-container locator binds transient ancestor indices and is
not established as reliable. In this isolated fixture, resolve iPad Home/Browse
buttons directly by their unique labels, retaining hittable-candidate selection,
finite on-screen bounds for touch observation, selected state, exact draft checks
and the final both-tabs isHittable requirement. Do not change production layout or
remove the assertion. One focused tab-return run distinguishes locator resolution
from an outstanding native hit-point failure; if it fails, retain that failure
without repeating this locator experiment. Settings readback is already accepted
at5040afd4 and need not be rerun for this test-only change.

### Android appearance lifecycle observation

Before another appearance correction, record whether the native screen/header is
recreated or its existing toolbar replaces/closes search during a light/dark switch.
Use only the marked disposable Android fixture archive and the pinned installed
react-native-screens sources. A runner-only recorder may emit at most256 structured
native diagnostic events: event kind, object identity, search-view identity and
query length, never query contents. Observe fragment creation/start/stop, header
attach/detach/update, menu rebuild, search open/close/focus and text changes. This
recorder is test instrumentation, must never enter production dependencies, and
source backups must permit restoration. Refuse Git checkouts, ambiguous dependency
roots, mismatched source hashes and existing instrumentation before any write.

One replay compares the same open synthetic Camping query before and after a live
appearance change. A new fragment identity directs the fix toward remount ownership;
stable identities plus menu/close events direct it toward native toolbar lifecycle.
If neither appears, retain the failure and inspect native rendering ownership rather
than repeat provider removal, key delivery, palette keys or JS query reseeding.
No appearance fix is accepted from this observation alone.

### Browse route state owns navigation, not adapter identity

Recreating a query adapter without changing Browse route parameters must not clear
the current search, filters or selected List/Map view. Route synchronization reacts
to route values; inventory-scoped query keys continue to isolate inventory results.
Verify a mounted searched Browse survives equivalent adapter replacement and still
applies an actual changed route query. This regression test distinguishes a source
state-reset defect from the remaining native appearance/rendering failure; passing
it alone does not accept Android live appearance.

### Android header appearance after query-ownership correction

The native trace and mounted regression separated query loss from header rendering:
query retention now passes the original Map appearance sequence without native
patches. With that confound corrected, apply the previously observed Android-only
segmented host recreation on semantic palette changes, preserving controlled value
and callbacks. Supply native search text, hint, tint and header-icon semantic colors
on Android. iOS keeps system search colors and host identity. Do not reseed search
on appearance, patch native menus, or recreate the screen. One integrated native
replay must now keep Map available through light/dark/light, search text/results,
continued editing, clear/close and view switching; inspect captures for contrast.
Verify a representative segmented consumer outside the header as well. This is a
new acceptance run after a proven ownership fix, not repetition of a failed reseed.

### Android photo canvas status contrast — rejected activity override

Normal-text API36 photo walkthrough shows black status glyphs on the fixed black
viewer canvas. The Android library's activity-level hide call does not hide them
in this modal. Extending the iOS declarative light-content StatusBar to Android
passes mounted ownership tests but fails the native screenshot; withdraw it.
ReactModalHostView0.83.6 copies activity appearance only at dialog show via
updateSystemAppearance. The next correction must own the dialog's system-bar style
or establish style before presentation with a native lifecycle guarantee, not a
timeout, repeated JS override or global preference. Preserve restoration on Close,
system Back, swipe dismissal and last-photo removal. This P3 appearance finding
remains separate from the verified Browse batch; no Android correction is accepted.

### Photo-free Browse rows align checkout status space

Run35979200078 passes connected Browse behavior on phone/iPad, but reviewed
expiration-return captures show checked-out photo-free cards with lower title
baselines than their peers. Each compact grid row must reserve the same status
space when any peer has a checkout label, using actual label text for native
measurement rather than a fixed height. A peer without checkout remains visually
blank in that slot and must not expose a false checkout label to accessibility.
Rows without checkout remain compact; photo-bearing rows retain square media,
Home row cards remain unchanged, and column changes recompute row membership.
Verify mixed checkout/no-checkout rows, row regrouping, and accessible semantics
before native title-baseline/visual acceptance. This is a project collection-layout
decision; it does not change item checkout state or add a domain concept.

### Place/detail entry clears native navigation chrome

Run35979200078 place-search-collapsed captures on phone/iPad show the initial
asset identity above the visible content edge while No photos is the first visible
body label. AssetDetailView's FlatList must use automatic system content insets,
like Browse, so native search/navigation owns top clearance without guessed header
heights. On initial entry, before scrolling or opening search, verify the asset's
identity header lies completely below navigation chrome and within the viewport.
Preserve search filtering/clear/cancel, ordinary detail scrolling and bottom-tab
return. This acceptance check supplements the existing functional search checks.

### Container details group identity, contents and commands

User capture2026-09-24 shows container checkout under the native tab/accessory
layer, contents commands detached above their heading, and borderless Move items
here. Keep availability/check-out with identity/location, before media and contents;
do not repeat it in the footer. Place Add/Move contents commands directly after
the first contents heading, using bounded native secondary styling for Move.
When a location has an empty/search-empty list, keep its contents commands grouped
with that state. Do not render an empty command group when permission offers none.
Keep ordinary item/location behavior and callbacks intact, and preserve native
automatic insets for navigation, tabs and the voice accessory. Verify final content
can scroll fully above those bars in the real tab shell; isolated fixtures cannot
establish bottom clearance. Inspect ordinary routed scrolling consumers for the
same missing-inset failure, rather than adding guessed universal bottom padding.
Apple layout guidance requires accounting for overlaid navigation controls:
https://developer.apple.com/design/human-interface-guidelines/layout
Container contents use the short heading Contents rather than repeating the full
asset name; separate groups use one spacing interval, avoiding stacked section
padding between the contents commands and empty state. Native Android observation
shows composed commands are reliable in the list header; keep the first section
and its controls there while virtualizing subsequent content rows.

### Sharing participates in persistent-navigation clearance review

Verify the ordinary Sharing destination inside the production tab stacks, with
its actual voice accessory and a populated invitation list long enough to scroll.
Its final invitation-link explanation must scroll
fully into the unobstructed content area. This is a representative ScrollView
consumer of the same layout contract as detail's FlatList. Use delivered tab,
accessory and header bounds on iPhone and iPad rather than fixed bottom padding.
An isolated sharing fixture without tabs cannot establish this acceptance.

Native bottom-clearance fixtures must contain the metadata used by their locator
and enough real contents rows to require scrolling on iPad. An absent synthetic
label is a fixture failure, not proof of product clipping. Detail uses a populated
container with a final Updated label; verify fixture data before dispatch.

### Explicit native audit dispatch

Native simulator workflows are explicitly dispatched for the affected journeys
and frozen revision. Pull-request updates must not automatically start the full
onboarding/fixture matrix: this duplicated focused runs, replayed historical
diagnostics and delayed release acceptance on limited macOS capacity. Ordinary
CI continues source, fixture and type checks; native acceptance remains required
for visual/lifecycle changes and must link the dispatched run and inspected
artifacts. The full sweep remains available through the explicit all selection.
Resolve a dispatched run by event=workflow_dispatch and exact head SHA, never the
most recent workflow result alone; PR and dispatched runs can share a branch.

## Home collection shortcut acceptance

Use a focused Home/Browse fixture with one shared inventory dataset and the actual
production `(tabs)` route paths. A root `/search` test route or a Browse placeholder
cannot prove cross-tab shortcut behavior. Verify a previously refined Map becomes
unfiltered recent List through Home See all; open and return from a real detail;
verify Home Checked out replaces unrelated refinements; verify ordinary tab return
preserves the resulting Browse context. Keep historical tab fixtures unchanged for
their existing scenarios. Run phone/iPad acceptance with the next frozen batch.

#### Dialog-owned photo system bars

The Android photo viewer must use light system-bar glyphs on its fixed dark canvas.
A local Expo native view inside the library's modal owns that dialog's appearance:
apply on window attachment and window focus, after React Native's dialog setup.
Use the attached view's WindowInsetsController on API30+, with system UI appearance
flags on older supported Android. Never reach into the Activity window or use a
JS timer; refuse Activity-root attachment. Dialog disposal leaves the underlying
screen's appearance untouched. This small native adapter is needed because React
Native StatusBar targets the Activity, not this separate dialog. It adds no runtime
dependency. Verify light/dark entry and restoration on Close, Back, swipe and
last-photo removal on the native build; source checks alone do not close M251.

Home shortcut run36013639520 stopped in shared setup on both devices: the focused
production-path fixture correctly opens Home at `/`, while setup expected the
legacy audit-menu button. The retained phone screenshot/tree confirms real Home
and both tabs. Match the named workflow's expected entry before exercising it;
keep all actual shortcut assertions. The legacy `all` arrangement cannot host this
production-path scenario because its root Search/detail fixtures intercept those
paths. Exclude this one scenario explicitly from that arrangement and run it via
`home-collections`; neither selector alone establishes full-app acceptance.

### Connected Settings overview review

Review the Settings root, Account, inventory settings and Diagnostics at normal
text size inside the delivered tab/voice shell. Existing editor and session-action
fixtures do not establish the overview's visual hierarchy or bottom clearance.
Use production screen components with controlled query data; fixture navigation
must be identified as such and does not certify production route callbacks.
Capture root entry and each destination, reveal the final Diagnostics version
above persistent controls, then switch tabs and return to the same destination.
Inspect current-account, current-household and current-inventory context, repeated
labels, row grouping, back behavior and command placement as a whole. Missing
inset props are a source risk, not proof of runtime overlap. Change layout only
when the walkthrough establishes a defect; retain separate acceptance for session
commands and editing rather than repeating those suites unchanged.

## Connected History structure acceptance

Review asset History and a change detail at normal text inside the production tab
stacks, with the persistent voice accessory present. Use a controlled paginated
activity repository and the real History query/screens. Open a change through the
production href, expand Technical details, scroll the last metadata above the
accessory, return to History, and switch tabs/back without losing the destination.
Check the History heading and mode control against the native header. This is a
structure/presentation review; it does not certify server authorization or Revert.
Do not infer clipping from missing source inset props alone. One baseline native
run must distinguish actual header/footer overlap from a safe native automatic
inset, and its screenshots determine whether a product correction is needed.

## Visual evidence contradictions

Before implementing a screenshot-driven correction, review the baseline's complete
visible composition at the named build and state. Element-only screenshots and
accessibility snapshots are supporting diagnostics; when they contradict the full
screen, resolve the observation discrepancy before claiming a rendering defect.
Compare the candidate with the same full-screen baseline state. Preserve useful
counterevidence and withdraw unsupported changes rather than treating a different
renderer or passing test as proof of improvement.

## Remaining inventory collection clearance

The retained `/assets` collection route must keep its heading and final card
commands reachable inside the production tab stack and voice accessory. Home now
uses Browse, but the retained route still has its own scroll owner. Its missing
automatic-inset prop is a source risk, not proof of clipping. Use a controlled
populated repository and the real InventoryAssetsRouteScreen in the production
tab layout. Capture full-screen entry and final-tag positions before asserting
header and bottom-control clearance. Verify the final tag is hittable as well as
geometrically clear. Do not infer overflow from content behind floating chrome
mid-scroll. If the final content is reachable, retain existing ownership; if it
is not, correct this scroll consumer and repeat the same check. Do not change Map
or other consumers that own their own insets based on this observation.

Baseline36045543667 at9afc3fd7 confirms the iPhone final tag remains behind
the persistent bottom controls after scrolling; the iPad baseline passes. The
collection FlatList must request automatic system content-inset adjustment, as
the other ordinary tab-stack lists do, preserving its existing visual padding.
Do not add device-specific bottom heights. Repeat the unchanged native entry and
final-tag acceptance on iPhone and iPad before claiming this correction verified.

## Notification list in persistent tabs

The standalone inbox fixture does not verify production tab ownership. A connected
normal-text check must use NotificationInboxScreen, its application queries and
controlled paginated data in the production tab stack. Open an actual asset detail
and return, confirm local read state, reach pagination above the voice/tab controls,
and reach the final loaded row. Preserve existing query/scope/operation guards.
Use full-screen entry and final captures on phone/iPad. No product change follows
from the coverage gap alone; fix only a reproduced interaction or layout defect.
