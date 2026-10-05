# Roadmap Spec

## Current focus — Full REST CLI parity, October 5, 2026

Implement [the approved CLI parity contract](cli-api-parity.spec.md): all REST
operations, remembered per-server contexts, keyboard-driven prompts, reliable
scripting, STE errors, and minimal measured binary/dependency growth. Chat is
out of scope. The generated SDK is the transport foundation, not proof of
complete command workflows. Full parity is not yet delivered. Existing native
acceptance remains tracked separately and does not block this work.

The draft CLI batch adds private account-bound contexts, searchable scope
selection, directory lifecycle commands, tags, asset workflows, and attachment metadata/lifecycle commands through
the generated SDK. Critical scope and authenticated-request tests, the CLI suite,
and Windows context-file CI have passed. The operation inventory tracks 192
contracts; 94 have their known workflow and field gaps closed. The remaining
98 still need implementation or full contract verification. The scope-override
persistence preference remains open before release. Search transport is implemented;
search command scope defaults await user input.


## Labels, Printers, And CLI: Released — October 4, 2026

Design PR #337 is merged. The integrated implementation now includes registered
printers and connectors, server-rendered QR labels, inventory print defaults,
atomic create-and-print, and manual print/reprint/test-label controls in web,
mobile, and the CLI. Both clients expose queue status, separate connector health
from printer readiness, allow registered media changes, and support explicit
uncertain-outcome recovery. CLI registration, credential rotation, and the
foreground worker use the generated Go SDK. SpiceDB scopes connector principals;
claim tokens fence attempts. Browser PKCE and provider-enabled device-code login
remain available for human commands.

The implementation stack is merged through `27346f77d`; its source tree matches
the integrated tree tested at `4128e468e` and is published in v0.41.0.
Printer retirement atomically cancels safe pending jobs and clears defaults while preserving started/uncertain evidence.
Both clients support one-action default printing and custom copy counts.
Inventory deletion revokes connectors and preserves printing history through
tenant deletion. Critical tests use faithful stateful fakes alongside real
provider and database acceptance; implementation-mirroring tests were excluded.
See [asset labels](../printing/asset-labels.spec.md),
[printer integration](../printing/printer-integration.spec.md), and [CLI](cli.spec.md).

### Web printing usability follow-up

The October 4 website audit prioritizes printer readiness, groups label defaults,
and moves diagnostic fields behind disclosures. Printer setup is an explicit
guided dialog; print history has a separate route that preserves the settings draft. Its refactor protects printer
drafts on dismissal and distinguishes reported output from device confirmation.
See [the scoped audit](../../docs/reports/web-printing-ux-audit.md) for browser
coverage, screenshots, and explicit limitations.

### Verified behavior

- HTTP authorization and isolation checks cover human and connector boundaries.
  Real PostgreSQL checks verify exclusive claims, scoped recovery, atomic rollback,
  and concurrent create retries producing exactly one asset and print job. Both
  create-versus-delete race outcomes pass with real row and foreign-key locks.
  The prior lock-order inversion was reproduced, corrected to lock inventory
  before printer, and reviewed before merge.
- Real-SpiceDB HTTP acceptance verifies permission loss between claim and start,
  rotation without privilege changes, pending/reordered synchronization, and
  retirement outcome/reconciliation. CI runs these alongside adapter grant,
  cross-inventory rejection, and transport-outage checks. Real PostgreSQL also
  verifies concurrent default initialization; these checks do not use mocks.
- Actual CLI browser PKCE with Dex verified S256/state/nonce, an ephemeral
  loopback callback, the separate CLI audience, protected credential storage,
  authorized API reads, and logout removal/denial. The local acceptance used the
  explicit HTTP development opt-in. Device-code login was also exercised against
  Dex and the API, including owner-only credential persistence and generated-SDK
  asset create/list calls. Neither authentication check establishes hardware output.
- Stateful worker checks cover lost responses, durable journals, readiness fencing,
  revocation, and uncertain output. Client checks cover retained request identities,
  scope changes, and fresh acknowledgement after a newer uncertain attempt.
- Connected browser acceptance completed in 41.8 seconds against the production
  web build with real Dex, PostgreSQL, and SpiceDB. It covered connector approval,
  the registered label preset, inventory defaults, checked create-and-print,
  idempotent replay with one queued job, and authorized label resolution while
  denying anonymous and unrelated users. The printer remained offline; this is
  not physical printing evidence. The harness and queued-state screenshot capture
  are recorded in commit `5b7c027be`.
- Registry-generated printer, media, and template pages use the production renderer
  for PNG/PDF examples. Offline drift checks validate owned outputs; QR decoding,
  dimensions, deterministic rendering, and responsive catalog views are verified.
  See [generated printing docs](printing-catalog-docs.spec.md).
- [Release v0.41.0](https://github.com/elsell/stuffstash/releases/tag/v0.41.0)
  published on October 4 at 00:04:07 UTC from
  `27346f77d59db3714ae6ee45be5201a5182af7bb`. All five downloaded CLI archives
  matched their checksum files, release manifest, and GitHub asset digests.
  The released Linux amd64 binary reported the exact tag/commit and
  `usbPrinting: true`; its catalog contained only the Linux USB QL-800/29 × 90 mm
  profile, with wake and physical verification both false. License notices were
  present. Other platforms received artifact integrity checks, not execution checks.
  PR #400 updates the pinned download instructions from those verified assets.

- October 4 hardware regression acceptance: an actual queued test label on Paul's
  Linux USB QL-800 printed, reached API `completed` with one completed copy, and
  cleared its worker journal at 01:03:31 UTC. The adapter candidate drains status
  continuously while writing raster data and treats empty USB IN completions as
  retryable, with bounded backoff. A stateful protocol fake now requires phase
  status draining before raster progress, preventing recurrence of the missing
  duplex behavior. Critical race tests cover shutdown, cancellation, fragmented
  status, queue overflow, and uncertainty without replay. The user confirmed the
  physical output; QR scanning and alignment were not assessed.

### Native printing usability follow-up — October 4, 2026

User-device screenshots exposed obscured scrolling controls, competing tinted
commands, and weak settings hierarchy. The current follow-up uses grouped native
settings/actions and automatic scroll insets; label-render failure recovery must
retry the selected render without repeating a physical handoff. Native run 37176721922 at `00ecca23c` passed the focused iPhone 17 and
iPad mini audit at normal and maximum text. Reviewed captures verify action
clearance, preview recovery, saved defaults and dismissal; see
[the acceptance scope](native-printing-audit.spec.md). PR #409 merged as
`ff03ab2c0` and shipped in v0.42.1. Release run 37178727071 verified
TestFlight build 238.1 at 05:41 UTC on October 4.

The deployed label instance was also uninitialized. An explicit bootstrap restored
public instance readiness and an authenticated CLI PNG render; GitOps now runs the
idempotent bootstrap after migrations, preserving the existing instance identity.

### Native scanner recovery acceptance — October 4, 2026

PR #414 merged as `b05f5bc7d`. The camera adapter retains an OS permission request
while the system dialog makes its task inactive; actual capture remains
foreground-only. Native run 37179475027
at `5492bdec7` passed four cases on iPhone 17 and iPad mini: real OS denial and
paste fallback, invalid/foreign links, retry with old-host resolution, and Cancel
with a late response. Retained links also survive simulated sign-in readiness.
[The iOS report](../../docs/reports/printing-ios-scanner-2026-10-04/README.md)
separates fixture evidence from real authentication. The normal Android APK
separately verified real Dex return and unrelated-account denial; see
[connected Android evidence](../../docs/reports/printing-android-2026-10-04/README.md).

### Hardware and delivery limits

Initial hardware/media support is only Brother QL-800 over USB on Linux with
29 × 90 mm stock (physical profile 29 × 89.8 mm, raster 306 × 991). Registration
trusts the user's loaded-size selection; queued labels retain their original
media. Templates remain independent of printers. Remote wake, additional media,
and a shared hosted QR resolver are outside this delivery.

Actual USB completion is verified for the bounded test above. Physical QR
scanning, label alignment, host udev/service setup, camera image decoding, and
system share/print journeys remain unverified. Source and fixture checks do not
replace those checks. Pending user-device evidence does not block unrelated delivery and
must not be reported as a pass or as closure of existing acceptance work.

## Integration priority clarification — October 2, 2026

A first-class CLI takes priority over further MCP work. MCP remains a supported
secondary integration. This changes future sequencing only; the complete portable
archive and restore flow remains the immediate delivery priority.

## Current delivery and remaining acceptance — October 3

Product delivery through **v0.28.31 is complete**: #326/v0.28.27 Sharing
metadata, #328/v0.28.28 date grouping, #330/v0.28.29 direct invitation cancellation,
#331/v0.28.30 mobile dates and #332/v0.28.31 web timestamps. All exact-source
release workflows succeeded. [Merge/check/release evidence](../../docs/reports/spec-gap-evidence-2026-10-01/deployments/delivery-v0.28.27-v0.28.31.json)
records each batch separately from runtime acceptance.

Paul now serves web v0.28.31 at infra revision
`6ed67e123cd1b2a6d653c6e8c6e35009b7368db7`, with ready replicas and HTTP200
checks ([rollout](../../docs/reports/spec-gap-evidence-2026-10-01/deployments/web-v0.28.31.json)).
API v0.28.26 remains deployed. Infra revision
`07396debd7cc8d23d5ec17b4d1b95df829906821` corrected the voice origin allowlist
behind TLS termination; eight deployed origin/authentication checks pass.
[Voice evidence](../../docs/reports/spec-gap-evidence-2026-10-01/voice-origin/README.md)
records the deployed boundary checks. The user confirmed voice works on
October 4, 2026; the [device checklist](../../docs/reports/user-testing-checklist.md)
records this separately from instrumented evidence, without unspecified build
or device details. #335 records the operator guidance.

Sharing #239 remains open after v0.28.29: initial confirmation presentation has
native evidence, but full recovery remains unverified. Two bounded runs exhausted
the investigation budget; no unchanged third run.
[Evidence](../../docs/reports/spec-gap-evidence-2026-10-01/sharing-direct-confirmation/README.md).

User-device checks are tracked in the [testing checklist](../../docs/reports/user-testing-checklist.md).
Per the user's October 3 instruction, pending user testing is not a delivery or
goal blocker. Continue independent work and keep missing acceptance explicitly
unverified; do not confuse this policy with a passing test or waive required checks.


Historical delivery snapshot through v0.28.26 API publication: product batches #310, #314,
#316 and #318 shipped as v0.28.22–v0.28.25 with successful TestFlight upload and
changelog jobs. Documentation-only #317 merged; its release workflow correctly
skipped publication and created no product tag. #321 was superseded by #322.
#322 merged at `378d5fcbd56e20b5ccbd76732bc387355255c437`; v0.28.26 images
and attestations are published. Its TestFlight completion is tracked separately
in [exact workflow evidence](../../docs/reports/spec-gap-evidence-2026-10-01/deployments/delivery-through-v0.28.26.json).
The earlier [v0.28.20–v0.28.21 record](../../docs/reports/spec-gap-evidence-2026-10-01/deployments/delivery-v0.28.20-v0.28.21.json)
and [preceding releases](../../docs/reports/spec-gap-evidence-2026-10-01/deployments/localization-delivery-through-v0.28.19.json)
retain their original evidence.

At that historical snapshot, Paul's web deployment was v0.28.26, infra revision `de14fb7b1802d9faae5cb1a45633b72a1eef83a3`.
The API rollout was v0.28.26, infra revision `5f2e02c73183d1525b632c89bf3a489cd253b0d6`.
[Web rollout](../../docs/reports/spec-gap-evidence-2026-10-01/deployments/web-v0.28.26.json)
and [API rollout](../../docs/reports/spec-gap-evidence-2026-10-01/deployments/return-cancel-v0.28.26.json)
record pinned images, Flux revision, ready replicas and HTTP200 health checks.
The API fixes return cancellation incorrectly rejected as stale after PostgreSQL
truncated timestamps. Real PostgreSQL reproduces the original failure and verifies
the correction plus scope/stale-edit guards. The user confirmed return
cancellation works on October 4, 2026; the checklist records that user-reported
acceptance without inferring build/device details. API publication and rollout
did not wait for TestFlight. Printer-settings verification is owned by the user
in a separate session; it is not a duplicate test request here.

The full audit remains incomplete. Release success does not close these gaps:

- Local-provider narrow real-model acceptance now passes: instrumented run
  37173848024 at `301496f43` verified diagnostic, lookup and tool-result answer
  with the same pinned 4B configuration. [Evidence](../../docs/reports/spec-gap-evidence-2026-10-01/local-provider/instrumented-4b/README.md)
  is synthetic, not authenticated inventory execution or speech/latency acceptance.
- Localization: the reviewed labels, interpolation and recovery fixes shipped.
  The residual TypeScript snapshot at `a93b524a` has a scoped caller review of
  31 files, retained as exact-string classifications. The subsequent script/template
  review found inline web Browse expiration labels bypassing the catalog; #318
  catalogs those links and guards inline each-block labels. A bounded
  [derived-string review](../../docs/reports/spec-gap-evidence-2026-10-01/localization-derived-review.json)
  classifies five remaining TypeScript files as diagnostics, protocol strings or
  test fixtures. These source reviews do not prove whole-client runtime localization. Existing RTL
  search/proposal artifacts were visually reviewed without another native run;
  [scoped evidence](../../docs/reports/spec-gap-evidence-2026-10-01/native-search/README.md)
  preserves broader mirroring and physical/assistive limits.
- Connected native acceptance: system-auth-browser discovery stopped the iPhone
  investigation before sign-in and isolation checks; no unchanged retry.
- Android connected acceptance verifies real browser sign-in return, persisted
  session, native sign-out and second-principal asset isolation. PR #309 merged
  at `c1c1184b` and release37110696922 succeeded.
- PR #310 merged at `849c26f1` and shipped as v0.28.22. Its
  three gaps are post-setup root navigation, authenticated
  Android archive acceptance and a representative TalkBack journey. The normal
  APK at `a93b524a` verifies fresh household creation returns to Home and an
  authenticated archive export/upload/review/restore opens a restored photo.
  [Scoped runtime evidence](../../docs/reports/spec-gap-evidence-2026-10-01/android-onboarding-archive/README.md)
  records an ADB bridge between the actual downloaded ZIP and system picker;
  native recipient saving is still unverified. TalkBack showed visible focus,
  but complete activation and spoken-output acceptance remain open. Later
  Home/Browse refresh errors are recorded without an unproven diagnosis.
  The navigation fix is shipped; these limits remain audit follow-ups.
- Web workspace creation, editing and shared action recovery now use the existing
  safe localized presenter. PR #311 merged at `42ae0fe3` and shipped as v0.28.21, including TestFlight
  and changelog publication. The [delivery record](../../docs/reports/spec-gap-evidence-2026-10-01/deployments/delivery-v0.28.20-v0.28.21.json)
  also records #309/v0.28.20. Child-dialog recovery and typed-error propagation
  shipped in PR #314/v0.28.23.
- A connected Android normal-text walkthrough at `a93b524a` now verifies Edit
  persistence across relaunch, Move cancel/commit and location restoration, search
  return, and Places-filter return. [Evidence](../../docs/reports/spec-gap-evidence-2026-10-01/core-workflows/android-connected/README.md)
  records the exact APK and isolated server. List/Map control movement awaits
  user confirmation; earlier intermittent refresh errors remain undiagnosed.
- Physical file-provider/recipient and assistive acceptance remain unverified.
  Existing user-confirmed notification delivery is preserved, not reopened.
- The previously recorded iOS sharing-menu keyboard overlap and proposed web
  Back-navigation change retain their existing follow-up/decision status.

Existing connected browser, normal-text native and bounded image timing evidence
remains scoped to its recorded workflows, devices, corpus and sample sizes. It
is not proof of broad physical-device performance or every adaptation state.
The records below preserve historical implementation and acceptance detail;
this section owns current delivery status.

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
- Android debug build `37027458024` passed. Later hosted build and scoped runtime
  acceptance are recorded below; physical document provider/recipient behavior
  remains unverified and is not closed by emulator evidence.

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
Keep physical results and broader audit acceptance open until verified.

## Delivered recovery copy and triage

Three scoped gaps follow the archive release: G6 household partial-success guidance
still bypassed the catalog; G6 checkout/return/details/undo availability errors
were shown verbatim by Home and Details; D2 residual-copy classifications did not
capture the reviewed recovery paths and the inventory predates archive delivery.
CI exposed same-tick invitation ID collisions in the seeded web repository. A
fixed-clock regression proves cancellation must preserve the other invitation;
monotonic invitation IDs remove the test's timing dependency. This repairs the
batch's evidence gate without changing production invitation generation.
The fixes preserve recovery state, command dispatch and English wording. Focused
onboarding/checkout/localization checks pass locally. PR #249 merged at
5961748e after CI37049355677 and Docs37049355682 passed. Release37049926510
succeeded: v0.28.1, TestFlight164.1, changelog readback verified by job110989637980.
Native acceptance remains separate; other residual candidates remain unreviewed.

## Inventory discovery recovery delivery

Three gaps: G6 unavailable selection guidance uses the existing typed/cataloged
error; G6 repeated or over-limit discovery pages use cataloged retry guidance;
D2 residual-copy inventory distinguishes the remaining cancellation control-flow
literal. Preserve selected inventory, request limits, cancellation and tenancy.
PR #251 merged at de87e4ae after CI37052298874 passed. Release37053212190
succeeded as v0.28.2 with TestFlight publication and verified changelog. This is
not connected-native acceptance.

## Archive request compatibility and ZIP security

PR #254 merged at `9f09b902`. Real React Native Request/Headers regression tests
reproduced JSON bodies being lost when forwarded directly to Expo fetch. The
bridge now passes explicit request options for archive creation and restore
approval, preserving authentication, cancellation and streaming downloads.
Release [37058610946](https://github.com/elsell/stuffstash/actions/runs/37058610946)
succeeded for **v0.28.3**, including TestFlight upload and changelog verification
(job111020203512). Physical archive creation, recipient saving and approved
restore on this build still require acceptance; upload success alone is not proof.

PR #255 merged at `a2b2fd15`. ZIP restore does not extract entry names onto the
filesystem and generates fresh destination attachment keys. Adversarial tests
cover traversal names, special files, conflicting local/central headers, and
rejection before restore approval. Required checks job111013509667 in
run37059638237 passed, including archive adapter, HTTP and application packages.
The unrelated conversation-browser job was cancelled and remains unverified.
Release37062636881 succeeded; these changes add tests, not a parser behavior fix.

## Connected native acceptance — investigation stopped

Draft PR #253 compiled the production iPhone client and isolated test target with
real local Dex, SpiceDB and API services. Run37055754739 at `f3759700` passed
startup and anonymous tenant-discovery rejection, then failed while locating the
system authentication browser. It does not distinguish connection failure from
an incorrect system-surface query. Sign-in, persistence, relaunch and second-user
isolation are unverified. The investigation budget is exhausted: no unchanged
hosted retries. Resume only with interactive observation or specific new evidence
that distinguishes those causes. [Durable scoped evidence](../../docs/reports/spec-gap-evidence-2026-10-01/connected-native/README.md).

These three reconciliations close stale delivery status for recovery batches,
archive compatibility/security, and connected-native investigation. They do not
close residual G6 copy migration or V1 physical, connected-native, assistive,
directional-layout and broader performance obligations.

## Local-provider acceptance

The production compatible adapter's controlled protocol tests remain separate
from real-model acceptance. Hosted run37071655138 at `969e2ef8` failed the profile
diagnostic and lookup-argument contract with pinned Ollama 0.9.5/Qwen3 0.6B.
Cleanup passed; final-answer replay was not reached. Keep this deployment
unverified, retain the [bounded failure evidence](../../docs/reports/spec-gap-evidence-2026-10-01/local-provider/README.md),
and do not retry unchanged or gate unrelated releases on it.
The single Qwen3 4B comparison at `3d0ec491` passed the profile diagnostic
(24,886 ms) but failed the required lookup assertion (47,558 ms); tool-result
replay and final answer were not reached. Run37078513169 completed cleanup.
The comparison is stopped with [its evidence retained](../../docs/reports/spec-gap-evidence-2026-10-01/local-provider/capacity-4b/README.md).
The October 4 instrumented run37173848024 at `301496f43` subsequently passed
diagnostic, lookup and tool-result answer with the same model and assertions.
[Retained evidence](../../docs/reports/spec-gap-evidence-2026-10-01/local-provider/instrumented-4b/README.md)
closes this narrow real-model round trip, without explaining the previous failure
or proving general quality, authenticated tool execution or voice latency.

## Android archive acceptance — scoped runtime evidence

The original local attempt at `588e4153` stopped on generated Gradle cache
corruption before app compilation. A fresh hosted build `37080427792` at
`d44531ba49ee840c8ad9e1b991952f95a7efd0a1` succeeded. Its verified APK ran on
Paul's isolated API 36 emulator: native 1 MiB upload integrity/authentication,
redirect/oversized-response rejection, recovery and cancellation passed.
The first approval attempt matched the page title instead of the action. A
selector correction was verified against that saved UI tree, then one bounded
run of the same APK passed preview, close/reopen, approval and the synthetic
destination callback. No product change or rebuild was required. This does not
prove opening a real restored inventory. Physical file-provider, authenticated
native restore and assistive acceptance remain separate. [Scoped runtime evidence](../../docs/reports/spec-gap-evidence-2026-10-01/android-archive/hosted-d44531ba/README.md).

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

## Label client delivery evidence

The independent label download/scanning slice adds all-kind asset label options,
server-rendered PNG/PDF delivery, Browse camera/paste scanning, and retained
label navigation through authentication/server changes. Critical parser,
transport, cancellation, permission, and native action tests are required with
client checks. Native ExpoCamera/ExpoPrint pod lock regeneration uses CI's actual
macOS resolver. The bounded QL-800 output and API completion check is verified
above; it does not establish physical QR decoding or interruption recovery.
Registered printer controls, settings and recovery are delivered, with named
iPhone/iPad settings and label-sheet evidence in
[native printing acceptance](native-printing-audit.spec.md). Camera decoding,
system share/print and assistive-technology acceptance remain separate checks.

## Maintenance

Update this file when batch scope, status, acceptance or material blockers change.
Keep it concise; do not append execution transcripts or duplicate domain specs.
Close a gap only with implementation and the required evidence, or with an explicit
user-approved scope decision. Removing an unmet requirement is not remediation.


## October 4 dependency security remediation

PR #407 shipped remediation for the original 151 Dependabot alerts. Reviewed
upgrades removed affected versions for 148 alerts; three upstream packages
(braces, node-forge, http-cache-semantics) remain version-flagged with reviewed
runtime mitigations and failing-before/passing-after real-library regressions.
The change preserves Expo 55 alignment, adds Metro image-size 2 compatibility,
and migrates the Astro/Starlight and OpenTelemetry pipelines. API/CLI suites,
client type checks, complete web/mobile suites, and fresh iOS/Android Hermes
exports pass. Real SpiceDB authorization integration, dependency-age checks and
custom-domain/project-path browser rendering pass; critic findings are resolved.
PR #407 passed CI and shipped in v0.41.2 and v0.42.0. GitHub reconciliation
leaves the three patched, version-flagged alerts open; do not report zero alerts.
PR #411 also removes the subsequently disclosed grpc-go vulnerable version; its
GitHub alert is closed. Native acceptance and current release rollout remain
separate evidence from dependency checks.
Per-alert evidence lives in `docs/reports/2026-10-04-dependency-security/`.
