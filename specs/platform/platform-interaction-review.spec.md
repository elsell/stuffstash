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
must remain runner-only and must not load production services or credentials.
Simulator results do not establish physical capture or server authorization.

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

Search placement diagnostics must isolate registration from screen content. Keep
the existing Place journey and add a runner-only route to the same production
Place fixture, with integrated-button search configured before presentation.
Both routes run the same search, result, clear, cancel and Back assertions; the
production shared search adapter still supplies current handlers and query state.
This comparison changes initial route configuration only. A passing comparison
does not certify a production correction or justify weakening the original gate.
