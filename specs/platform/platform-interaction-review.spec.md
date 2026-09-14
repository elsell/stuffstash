# Platform Interaction Review

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
