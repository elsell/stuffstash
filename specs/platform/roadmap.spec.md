# Roadmap Spec

## Integration priority clarification — October 2, 2026

A first-class CLI takes priority over further MCP work. MCP remains a supported
secondary integration. This changes future sequencing only; the complete portable
archive and restore flow remains the immediate delivery priority.

## Current delivery — portable archive and restore, October 2

PR #244 merged at `514a4f37`. It delivers A1 durable ZIP export with selected
original media, A2 validated and explicitly approved restore into a new inventory,
and A3 mobile/web controls and connected round-trip acceptance. The binding scope
is [portable archive](../data-portability/portable-archive.spec.md).

Restore publishes data atomically, then waits for the durable ownership grant and
confirms destination access before offering Open inventory. Delayed finalization
survives worker restarts without republishing data or spinning. Published originals
remain intact after job expiry or a finalization permission failure.

Evidence:
- CI `37032804605` passed at `3609065f`, including required checks and the real
  OIDC/PostgreSQL/Garage/SpiceDB browser journey: ZIP download, upload, review,
  approval, immediate opening and reload. The original photo bytes matched after
  restore, with fresh asset and attachment IDs. Its `connected-browser-evidence`
  artifact retains the fixture ZIP, state sequence and screenshots.
- Domain/application, persistence and real HTTP archive tests cover delayed grants,
  recovery, cancellation after publication, permission failures and scoped access.
  Focused mobile/web tests, type checks and code-critic review passed.
- Native run `37027352935` passed archive upload integrity, rejection/cancellation
  and restore-review/keyboard cases on iPhone17 and iPad mini. iPad legacy JSON
  sharing/cleanup also passed. The iPhone legacy share assertion timed out despite
  the correct file appearing in the system share sheet; phone dismissal/cleanup
  was not reached and remains unverified.
- Android debug build `37027458024` passed. Android runtime and physical document
  provider/recipient behavior remain unverified; simulator evidence does not close
  these obligations.

Release `37033608570` succeeded for **v0.28.0**, including TestFlight build
**162.1**, Apple processing and changelog readback. Paul GitOps commit
`215f7622d23636cd2eb8e479811c9ecb9f0463d5` deployed the published API/web digests.
Flux confirmed that revision, both deployments rolled out, the migration init
container exited successfully, API health returned healthy and web returned HTTP200.
The API ingress now allows the configured 1 GiB archive uploads and 30-minute
transfers. Deployment checks establish runtime readiness, not a new authenticated
production restore; the connected acceptance above ran in isolated CI.

Documentation follow-up #245 also carries the v0.28.0 self-host image pins from
#246, superseding the older pin-only #240/#243. Physical iPhone archive saving,
file picking and restored-photo opening have been requested from the user on162.1.
Keep those results, Android runtime and broader audit acceptance open until verified.

## Frozen follow-up batch — recovery copy and triage

Three scoped gaps follow the archive release: G6 household partial-success guidance
still bypassed the catalog; G6 checkout/return/details/undo availability errors
were shown verbatim by Home and Details; D2 residual-copy classifications did not
capture the reviewed recovery paths and the inventory predates archive delivery.
CI exposed same-tick invitation ID collisions in the seeded web repository. A
fixed-clock regression proves cancellation must preserve the other invitation;
monotonic invitation IDs remove the test's timing dependency. This repairs the
batch's evidence gate without changing production invitation generation.
The fixes preserve recovery state, command dispatch and English wording. Focused
onboarding/checkout/localization checks pass locally; release and native acceptance
are not yet established. Other residual candidates remain unreviewed.

## Next frozen batch — inventory discovery recovery

Three gaps: G6 unavailable selection guidance uses the existing typed/cataloged
error; G6 repeated or over-limit discovery pages use cataloged retry guidance;
D2 residual-copy inventory distinguishes the remaining cancellation control-flow
literal. Preserve selected inventory, request limits, cancellation and tenancy.
This batch is separate from PR #249 delivery and connected-native acceptance.

## Confirmed mobile archive request failure — immediate fix

Physical build165.1 reaches the deployed archive API: job listing returned200,
archive creation returned400, and native restore upload returned201. The mobile
archive control path forwards React Native Request to Expo fetch, which reads a
missing public `body` property and drops JSON. A regression using React Native's
actual Request/Headers implementation reproduces the missing body. Freeze this
batch around create request bodies, restore approval request bodies, and streaming/
authentication regression coverage. Bridge explicit URL/init into Expo fetch;
retain native upload and streaming downloads. This takes priority over the
independent connected sign-in verification draft; do not gate this fix on it.

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
| Completion | G7 visible-image telemetry; G8 domain application-package migration; V1 connected acceptance and performance evidence | Merged PR #216 at f4a8dd48; full CI passed. Browser image delivery/decode and domain service extractions implemented. Remaining obligations continue below. |
| Connected acceptance | V1 real OIDC/PKCE browser sign-in and principal isolation; G7 native image lifecycle measurements; G8 voice vocabulary ownership | Merged PR #218 at a91f9ec0. Final CI 36906940360 passed at 8b1adb7b. Native run 36904211794 at 4ef2f027 passed on iPhone17 and iPad mini. [Retained evidence](../../docs/reports/spec-gap-evidence-2026-10-01/connected/README.md). |

V1 includes authenticated browser journeys, representative connected native
workflows, relevant assistive/adaptation and physical integration checks, and
bounded image/performance measurements. Do not claim missing physical access or
live evidence succeeded. Track the specific remaining acceptance task and obtain
needed user participation only when execution requires it. Three gaps per PR is
a batching rule, not permission to replace missing implementation with scaffolding.

## Current batch completion — G8, G6, V1

- G8: move scoped realtime read-tool orchestration into agent-model application
  ownership; retain existing query authorization, wire contracts and adversarial tests.
- G6: remove remaining Settings diagnostics fallback copy from adapter literals and
  regenerate the residual-message inventory. This does not certify every remaining
  candidate as migrated or broader directional-layout acceptance.
- V1: extend the real OIDC browser journey with UI item creation and actual JSON/CSV
  downloads, including unauthenticated and other-principal export rejection.

PR #221 merged at `31f81eff` before its failing connected export check was required
by GitHub. Its release run36911487216 was cancelled before image/tag publication.
The connected run36910592292 exposed HTTP500: PostgreSQL rejected the new
`inventory.exported` audit action. PR #222 adds migration59 and a domain-action migration guard; merged `39700a21`.
CI36912314042 passed at `b6b07b1c`, including actual JSON/CSV downloads, persisted
export audit history, and other-principal rejection. The connected job is now a
required main check. Release36913294557 succeeded for this completed batch (v0.26.2).
[Retained acceptance](../../docs/reports/spec-gap-evidence-2026-10-01/connected/README.md).

Remaining acceptance includes physical iOS export recipient saving (requested from
user), broader connected native/adaptation workflows, and bounded performance
measurements beyond image callback samples. Remaining G8 orchestration and G6
presentation candidates continue after this batch. Do not gate this batch on those
unrelated remaining requirements or claim they are complete.

## Current delivery and next frozen batch

PR #224 shipped as v0.27.0 (release36915932407). PR #225 merged667a9196
and #227 mergedc5c89fe7 after their full CI passed. Their separate publication
attempts did not complete: GitHub replaced a pending run, and rerun36917837759
failed tag creation because its token lacked workflows permission.

At the user's request, all remaining open PR work was consolidated into #228:
localization guidance, remaining identity/expiration/voice application ownership,
and v0.27.0 image pins. Combined CI36921590281 passed; merged58bb40a8.
PRs #226 and #229 were closed as incorporated. Catch-up release36922285320 succeeded as v0.27.1, including TestFlight
build/upload. Future batches wait for the preceding release to finish before merge.
Native expansion run36920101432 passed on iPhone/iPad at673fb0b5;
[inspected Add/recovery evidence](../../docs/reports/spec-gap-evidence-2026-10-01/native-localization.md).

PR #230 merged0538ddb6 after CI36922781635 passed. Its separate
release36925924885 succeeded as v0.27.2, including TestFlight build/upload. This batch closed three named gaps:
- G6 option copy: migrate previously missed nested option labels and displayed
  fixture-selection fallback labels into the catalog, preserving wire values.
- G6 enforcement: check display properties inside option objects and Svelte
  scripts, with regressions protecting protocol values and styles.
- D2 inventory fidelity: include expressions in component option attributes in
  residual-copy triage and correct the AssetDetail error-classification source path.

PR #232 merged001dfb28; release36929650300 succeeded as v0.27.3,
including TestFlight build/upload and the build changelog. Image-pin PR #235
mergedb3b2a9c3 after approved pull-request checks passed. This batch delivered:
- G6 adapter recovery: catalog surfaced timeout and onboarding failures while
  preserving cancellation and tenancy semantics.
- V1 native expansion: retain inspected en-XA Add/recovery evidence on phone/tablet.
- D2 acceptance status: distinguish this scoped success from still-missing
  physical export, assistive, connected native and broader directional evidence.
  Native search-placement run36923519340 completed seven of eight tests on each
  device; [scoped results and current diagnosis](../../docs/reports/spec-gap-evidence-2026-10-01/native-search/README.md).

PR #233 merged at `47b43d39`; required CI passed. Release36940409190
succeeded with image/TestFlight publication skipped because this batch changed
verification and documentation only. Its three gaps were:
- V1 native search observation and scoped acceptance evidence.
- V1 pending-proposal preservation across close and cancelled reset.
- G6 production-catalog expectations preserving user text and system controls.
Expansion run36930043292 passed all three workflows on both devices. Targeted
run36933539569 passed all three on iPhone and static search on iPad; iPad tag and
expiration launch failed before product assertions. Track those unverified
workflows in [issue #236](https://github.com/elsell/stuffstash/issues/236).

Current priority is normal-English everyday use: broken actions or lost work,
Browse/Details/Edit/Move and filters, then hierarchy and reachable controls.
RTL and broader adaptation remain obligations but do not lead the work queue.

PR #238 merged `30db241f` and shipped v0.27.4, TestFlight build159.1, in
release36954278739. Upload and changelog readback succeeded. The user accepted
the unresolved Sharing keyboard issue for release; #239 owns its follow-up.

October 2 real-browser checks on Paul's isolated v0.27.3 server verified
Browse search/filtering, persisted edits and moves, and Map lookup at normal text.
A newly observed return-context issue awaits user confirmation: Details Back after
cancelling Move returns filtered Browse users to Home and loses their query.
[Current connected evidence](../../docs/reports/spec-gap-evidence-2026-10-01/connected/README.md).
Bounded LAN browser timings now cover four actual image uploads, six detail
reloads and four four-thumbnail Browse navigations, with raw samples and explicit
cache/corpus/automation limits. This does not close native or physical acceptance.
PR #241 merged this evidence at `33bcb67c`. Subsequent connected desktop keyboard
checks verified Filters focus containment/return and keyboard editing persisted
after reload; the linked report retains the focus sequence and its limits.

Normal-text run36941858463 at main `47b43d39` passed Details and Sharing footer
checks on iPhone17 and iPad mini. Inspected Details captures show final content
above the voice accessory and persistent tabs. Sharing captures likewise show the final invitation and footer above
the voice accessory, with the device's tabs visible. These are controlled native fixtures, not authenticated production
journeys. [Current evidence](../../docs/reports/spec-gap-evidence-2026-10-01/core-workflows/README.md).

No new task-blocking defect was established by the focused review. On October 2,
the user approved replacing repeated Sharing cancellation buttons with a trailing
native invitation menu and an isolated authenticated test server on Paul. This
batch preserves cancellation confirmation and permissions; phone/iPad menu and
footer verification precede release. Physical export recipient saving, connected
native sign-in, assistive checks and production performance remain outstanding.

The user authorized releasing PR #238 with [issue #239](https://github.com/elsell/stuffstash/issues/239)
tracked for follow-up. Native run36951622669 at `ef5ba7a0` failed on both
devices because opening invitation actions restored the email keyboard. Initial
run36946987743 passed iPad recovery and both footer checks; its iPhone cancellation
retry was obscured by the keyboard. Remounting after failure and explicitly
blurring the SwiftUI field did not resolve it. Do not repeat unchanged runs or
claim native acceptance. This known issue does not gate the user-authorized
release; regular CI remains required. [Failure evidence](../../docs/reports/spec-gap-evidence-2026-10-01/core-workflows/README.md).

The isolated test server on Paul is available at `https://nsa-lnx-rzn.local:28780`
(API) and port28781 (web), using pinned v0.27.3 images and separate volumes.
`~/stuffstash-audit-server/README.md` on Paul records the private credential and
CA locations. PKCE/refresh, tenant/inventory/item create/read, invitation creation
and cancellation, JSON/CSV export and unrelated-principal denial passed. Hosted
native runners still need a secure network path; server/API checks do not close
connected native acceptance. No production resources were changed.
The current native-audit workflow also clears API/tenant configuration and installs
isolated fixture routes. Connected acceptance therefore requires a separate real
onboarding/sign-in path, explicit test-server certificate trust and a runner network
path; configuring a tunnel alone or rerunning fixtures cannot satisfy it.

The already-running RTL run36936112035 completed successfully on both devices;
no further run was started. Its test result does not imply new visual acceptance.
Issue #236 retains the separate iPad tag/expiration launch evidence gap. Stop
unchanged retries; any later runner investigation needs an outcome-dependent
decision and fixed budget. Product fixes already shipped in v0.27.3.

G8 source review found no further concrete root-policy ownership violation after
#228; construction, cross-domain composition and compatibility facades remain
intentional. Remaining localization candidates and broader native/physical
acceptance stay open. Inventory candidates include technical strings and are not
product-defect counts. Native expansion does not establish search/approval,
physical export recipient saving, or assistive-technology acceptance.

## Delivery and acceptance rules

Historical audits must link source evidence to the audited revision so later
implementation and package moves do not invalidate the recorded baseline.

- Delivery is authorized: complete checks and review, merge, and publish one release
  per frozen batch. Do not hold passing batches for separate merge permission or
  completion of the full audit. PR #213–#216 shipped together in v0.26.0; subsequent batches ship separately.
- GitHub main rules require both `Required checks` and `Connected OIDC browser journey`.
  Verify every frozen batch check before merging; auto-merge is not proof that all
  CI jobs are mandatory.
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

Catch-up release `v0.26.0`, source `f4a8dd48`, run36903326170 succeeded,
including TestFlight upload and build changelog publication. PR #218 merged
`a91f9ec0`; its v0.26.1 release run36907915266 succeeded, including TestFlight upload
and build changelog publication.
These are delivery records, not new physical-device or whole-app acceptance.

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
