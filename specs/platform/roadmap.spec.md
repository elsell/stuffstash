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
| Documentation | D1 inaccurate roadmap; D2 historical evidence/spec status drift; D3 ambiguous audit coverage and delivery status | Merged PR #213. |
| Capabilities 1 | G1 JSON/CSV inventory export; G2 authenticated external MCP read tools; G3 executable OpenAI-compatible/local providers | Merged PR #214. CI and native export share/cancel/cleanup passed; recipient saving remains V1. Native run 36862474560 at d8dd2c45. |
| Capabilities 2 | G4 web inventory conversation; G5 approved conversational asset/custom-field edits; G6 localization infrastructure and client migration | Merged PR #215. G4/G5 implemented; G6 catalogs, client copy migration and copy gate delivered. Native RTL Add/recovery passed on iPhone/iPad in run 36894135591; [inspected evidence](../../docs/reports/spec-gap-evidence-2026-10-01/native-localization.md). Remaining: residual copy migration and broader directional-layout coverage. |
| Completion | G7 visible-image telemetry; G8 domain application-package migration; V1 connected acceptance and performance evidence | PR #216 delivers scoped image outcomes and browser decode/delivery verification, domain application extractions including response safety and action-plan review projection. Full CI 36899689933 passed at 045570f3; merge reconciliation receives its own CI. Remaining: realtime orchestration, native image measurements, and connected/physical acceptance. Unfinished connected OIDC test work is preserved outside this batch. |

V1 includes authenticated browser journeys, representative connected native
workflows, relevant assistive/adaptation and physical integration checks, and
bounded image/performance measurements. Do not claim missing physical access or
live evidence succeeded. Track the specific remaining acceptance task and obtain
needed user participation only when execution requires it. Three gaps per PR is
a batching rule, not permission to replace missing implementation with scaffolding.

## Next frozen batch — V1, G7, G8

Continue from merged PR #216 with three scoped deliverables: V1 real browser
OIDC/PKCE sign-in, automatic workspace provisioning and principal isolation;
G7 representative native image lifecycle measurements; G8 remaining realtime
application orchestration migration. This batch does not claim broader physical
acceptance or all localization work is complete. Release only its verified changes.

## Delivery and acceptance rules

- Delivery is authorized: complete checks and review, merge, and publish one release
  per frozen batch. Do not hold passing batches for separate merge permission or
  completion of the full audit. The current PR #213–#216 catch-up ships together.
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
