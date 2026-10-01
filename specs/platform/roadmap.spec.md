# Roadmap Spec

## Current objective — October 1, 2026

Close the documentation, implementation and verification gaps in
`docs/reports/spec-implementation-audit-2026-10-01.md`. That report is a dated
baseline at `dabe2839`, not a mutable backlog. This file owns current delivery
status and sequencing. Domain specs continue to own behavior.

The user authorizes implementation of all found gaps, largest capabilities first,
in delivery batches of three named gaps per PR. Correct documentation first.
Use atomic commits within each PR. Update the relevant behavior spec before code;
write only critical behavior, recovery and adversarial boundary tests. Existing
required checks and code-critic review remain mandatory. Do not add low-value
implementation-mirroring tests or weaken security coverage to meet a batch size.

## Frozen batch sequence

| Batch | Three gaps | State |
| --- | --- | --- |
| Documentation | D1 inaccurate roadmap; D2 historical evidence/spec status drift; D3 ambiguous audit coverage and delivery status | PR #213; required checks passed, merge pending |
| Capabilities 1 | G1 JSON/CSV inventory export; G2 authenticated external MCP read tools; G3 executable OpenAI-compatible/local providers | PR #214 ready for review: CI and native export share/cancel/cleanup passed at d8dd2c45 (run 36862474560); recipient saving remains V1 |
| Capabilities 2 | G4 web inventory conversation; G5 approved conversational asset/custom-field edits; G6 localization infrastructure and client migration | Draft PR #215. G4 typed web conversation and G5 approved detail/customization changes implemented; CI 36869156630 passed at 9f9de6e7, including English/expanded/RTL browser journeys. G6 shared catalogs, direct-render copy gate and history/voice/search/settings presentation migration implemented. Native expanded/RTL Add recovery case prepared. CI 36871896807 passed at 1b4bad65 after the dependency-order fix. Expanded/RTL native acceptance remains open: run 36871890499 failed before the fixture entry appeared on iPhone; the artifact records a startup JavaScript fatal through RCTExceptionsManager. A missing-Intl regression reproduces failure at translator initialization; mobile now installs pinned FormatJS prerequisites, plural/list APIs and English data before initialization. Six localization tests and typecheck pass; critic approved. Native confirmation is still required. Run 36871894730 is not accepted. Consolidated presentation migration covers mobile/web controls, accessibility labels, recovery and administrative status maps. Move headings, checkout status and location photo styling use explicit state instead of English labels. Full mobile: 2,070 passed; web: 1,128 passed with four timeouts, all four passing in a single-worker rerun. Follow-up affected tests and both typechecks passed. Application-generated presentation and safe validation messages now use catalogs; 365 mobile and 279 web application tests pass, both typechecks pass. Root parent suggestions use identity rather than the English root label. Complete expiration state/precision messages replace concatenated fragments. Remaining copy classification and native locale acceptance remain |
| Completion | G7 visible-image telemetry; G8 domain application-package migration; V1 connected acceptance and performance evidence | In progress on codex/completion-gaps-batch. Web image lifecycle consumers now record actual surface/requested variant and load/error/cancel outcomes without URLs; 61 affected tests plus the thumbnail-context integration pass, typecheck passes. Native inventory/gallery/upload images now use a shared observer context and source-scoped lifecycle; 11 initial critical tests passed, then 23 focused tests passed after updating two legacy shallow assertions for the measured image component. Full mobile run passed 2,072 cases with those two assertion failures subsequently resolved. Native typecheck passes. G7 fullscreen wiring and runtime measurements, G8 migration, and V1 connected evidence remain open |

V1 includes authenticated browser journeys, representative connected native
workflows, relevant assistive/adaptation and physical integration checks, and
bounded image/performance measurements. Do not claim missing physical access or
live evidence succeeded. Track the specific remaining acceptance task and obtain
needed user participation only when execution requires it. Three gaps per PR is
a batching rule, not permission to replace missing implementation with scaffolding.

## Delivery and acceptance rules

- Freeze each PR's three gaps and critical workflow/regression checks. Unrelated
  existing findings remain tracked and do not gate a verified frozen release.
- Preserve the user's normal-text priority: structure and stable navigation,
  everyday tasks, visual hierarchy, then detailed/adaptation states.
- Ask before expanding work into newly suspected product/design defects. The
  October 1 gaps and previously confirmed defects are already authorized.
- Shared-control tests plus representative consumers and critical workflows are
  preferred over one test for every surface/axis cell.
- Keep one current diagnosis per investigation, specify what an experiment will
  distinguish, and make a concrete implementation decision within a bounded budget.
- Long-running CI waits use sleeping scripts that collect terminal outcomes;
  avoid repeated unchanged status checks and polling commentary.
- Preserve disk space. Native builds run on macOS CI; do not build iOS locally.
- Source tests, native fixtures, physical-device acceptance, upload, processing,
  and deployment are distinct evidence. Record exact revisions and scope.

## Current implementation baseline

These are implemented, not new backlog: API rate limiting (default enabled),
SpiceDB LookupResources-based search visibility, S3 direct uploads, canonical
invitation links and client acceptance, durable Homebox live/CSV imports, web
media/actions, conversation workflow/case/run administration, asset-scoped
undo/redo, scoped server state, expiration workspace and notification adapters.
The user has verified notification delivery; historical APNs setup notes do not
reopen that issue. Source presence does not certify every runtime boundary.

Implementation and verification status for G1–G8 is recorded in the batch table.
The audit baseline predates the open PRs; do not present its original absence
findings as the current branch state. Unmerged delivery and missing acceptance
evidence remain explicit rather than being treated as completed releases.

Deliberate scope limits are not defects: offline writes/sync, cross-inventory
moves, multi-inventory plans, calendar-grid expiration, non-asset undo, whole-asset
time travel, destructive field-schema changes, arbitrary-server invitation app
handoff, model photo consumption and Android store distribution. If later work
expands these limits, update its domain/security contract first.

## Latest verified release evidence

Main baseline: `dabe2839599c80c5db0ed22c8792adef1265ce95` (PR209), tag `v0.25.3`.
Release run36664629076 completed successfully; iOS upload job109728590645 and
changelog job109732833589 succeeded. Native run36661826231 passed its selected
Browse header/filter workflows on iPhone17 and iPad mini at source38e493d9.
This records the workflow outcome, not a new physical-device or whole-app review.

PR210 merged8308690e. Release36657966242 succeeded. Its source/native evidence
includes sheet notice dismissal, Add keyboard and title copying; the failed
pushed-notice Back check ended with Settings/Siri foregrounded and is unresolved
verification evidence, not an established app defect. Do not simply relabel it
passed or retry unchanged until green.

## Audit and historical evidence

`docs/reports/mobile-ui-remediation-2026-09-14/README.md` owns mobile evidence
interpretation; its matrix is an omission inventory, not a defect count or release
queue. Current inventory:147 surfaces,24 axes,3,528 cells. Historical evidence
must retain its revision/date and may not override later acceptance or this
current sequencing. Old roadmap checkpoint narratives remain in Git history.

## Maintenance

Update this file when batch scope, status, acceptance or material blockers change.
Keep it concise; do not append execution transcripts or duplicate domain specs.
Close a gap only with implementation and the required evidence, or with an explicit
user-approved scope decision. Removing an unmet requirement is not remediation.
