# Platform Interaction Review

## Direct UIKit optional tag color

M51's investigation budget is exhausted. Replace the current SwiftUI-hosted color
control with an Expo view adapter around Apple's UIColorWell, retaining the system
picker and existing in-place optional tag-color task. This is an implementation
decision supported by repeated opening failures, not a claim that SwiftUI caused
them. See Apple's UIColorWell and color-well guidance linked in the audit evidence.

The native control has a44-point minimum hit area, a single descriptive accessible
name, no alpha editing, and native enabled/disabled state. Its wrapper must not
duplicate accessibility or capture touches. Opening or dismissing without selection
must not invent a persisted color. Valid selection emits uppercase six-digit RGB
to the current parent draft; rounded byte conversion preserves unchanged channels.
Parent color changes, clearing and permission locks update the control without
emitting user changes. Existing presets and Android behavior remain unchanged.

Expose the UIColorWell itself as the single accessible button, rather than its
internal default-named child. Hide the accompanying visual text from accessibility
traversal. Native acceptance must retain the descriptive name and exercise actual
activation; changing the test to accept the internal generic label is insufficient.
The adapter owns the well's accessible-name getter so UIKit selection updates
cannot replace it with the generic system name. Preserve UIKit's activation and
enabled-state behavior; do not proxy taps or traverse private subviews.
For the Add draft regression, keyboard readiness checks the next intended key
(T for Tent, C for Camping) instead of enumerating every key. Preserve the same
five-second deadline, actual typing and exact complete-value assertions.

Use a focused local `color-well` module with ExpoModulesCore; keep it separate from
sheet geometry. Detect its actual native view before choosing the adapter so older
binaries retain the existing fallback. Test null/invalid selection, external reset,
disabled native events, RGB conversion and shared Add/Edit/Settings consumers.
One focused phone/iPad acceptance run must cover first-tap opening, selection,
dismissal, clear, draft retention and lock/unlock before release. A failed gate
drives a concrete correction, not another broad comparison.

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


Native352470 observes a true keyboard-readiness evaluation taking4.0277 seconds,
but XCTWaiter reports timeout after its initial1.0397-second scheduling delay.
Evaluate observation predicates immediately, then spend only the remainder of the
same five-second budget waiting when false. A matching evaluation completed after
the total deadline must still time out. Apply this to keyboard/query and filter
clearance observation; preserve every state/geometry/hit-test condition. Record
eager and subsequent evaluation durations to expose future observation cost.
