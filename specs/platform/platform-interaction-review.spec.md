# Platform Interaction Review

## Native search acceptance on iPad

Integrated toolbar search on iPad can expose Clear text without a separate Cancel
button. Native audit journeys must verify clearing the query, dismissing the
keyboard, restoring unfiltered content and reaching navigation/actions through
that observed platform pattern. They must not require the phone-only Cancel
affordance or force a particular idle field width. Phone journeys retain Cancel
and collapse assertions. Record screenshots/hierarchy that justify this distinction;
an amended test remains unverified until rerun on its native target.

## Purpose

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
