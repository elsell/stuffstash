# Mobile UI audit — current state

Latest verified TestFlight: **0.24.28 (118.1)**; [release and changelog verification](https://github.com/elsell/stuffstash/actions/runs/35818700192).
The comprehensive audit remains incomplete. Normal-text user-visible defects take
priority; freeze and release verified batches independently of audit completion.

## Scope and evidence

[Surfaces](surfaces.json), [axes](axes.json), and [coverage matrix](matrix.csv)
enumerate142 surfaces ×24 axes =3,408 cells:2,593 source-reviewed,576 finding,
198 not-applicable and41 runtime-partial. These classify evidence, not distinct
unresolved defects or individual test requirements. Verify shared controls,
representative consumers and critical workflows; test additional consumers where
composition or ownership differs. [Findings](findings.md) retain stable IDs.

[Latest full native evidence](native-full-352471.md): phone74/92, iPad84/92 fixture
cases pass. All16 released search checks pass. Other failures include product,
enlarged-text and deliberately varied diagnostic cases; the aggregate does not
establish whole-app acceptance. Physical-device-only behavior, Android and broader
adaptation/assistive-technology coverage remain open in the surface reports.

## Current diagnosis and decisions — September 23

| Defect | Established facts | Decision and next acceptance |
| --- | --- | --- |
| M51: custom color picker sometimes does not open | Ordinary well taps intermittently leave the parent unchanged. Disabling scrolling is insufficient. RGB retention and disabled-state fixes already passed. [Evidence](native-color-scroll-352164.md). | Diagnostic budget exhausted. Use a standard UIKit button with explicitly owned system color-picker presentation; both prior well paths showed missed opening. Verify first-tap opening, selection, dismissal, parent draft retention, clear and lock/unlock on phone/iPad. Retain presets. |
| Text loss in RN input comparisons | Complete key sequences can yield missing/reordered JS values. Verified provider-free iPad still fails; pacing or removing assistance is not a general correction. [Consolidated evidence](native-text-entry-352471.md). | Diagnostic budget exhausted. Keep existing native Add-name, invitation-email and onboarding adapters. Verify actual editing workflows and fix a reproduced consumer with the established native field pattern; do not replace every input because an isolated comparison fails. Preserve external reset, draft, disabled and submission semantics. |

Run35803226783 completed: phone12/14 and iPad9/14. Its
[exact outcomes](native-text-entry-358032-results.csv) reproduce the known isolated
RN input failures; this does not change the decision above. No further trace
instrumentation is planned.

M51 current candidate: aa9fbd9e uses a standard UIKit button and explicitly
owned UIColorPickerViewController, including popover/sheet dismissal, lock/removal
cleanup and retired-callback guards. The decorative swatch retains visual selection
while the command keeps its native tint. Source checks and critic pass;
CI35812502532 passes all six jobs. [Draft PR161](https://github.com/elsell/stuffstash/pull/161)
is not released. Duplicate PR-triggered CI/full audit runs were canceled in favor
of the already-running exact-commit CI and focused acceptance, not treated as passes.

[Native35812501085](native-color-358125-results.csv) accepts9/9 iPad and7/9 phone.
Both devices pass ordinary opening, name, clear, lock/unlock, RGB edits and Settings
save/back. Phone fails coordinate probe0 and the Add exact-value waiter. The
[Add final hierarchy](evidence/phone-add-final-value-358125.txt) contains complete
Tent after the timeout; this is not evidence of text loss or a passing deadline.
The [coordinate-tap hierarchy](evidence/phone-color-tap-358125.txt) still shows the
parent, so that opening remains unverified. Earlier well outcomes remain in
[native358096](native-color-358096-results.csv).

Decision: retain this implementation, keep PR161 draft, and stop control-swapping
or unchanged reruns. The investigation budget remains exhausted. Native phone
coordinate activation and timely Add observation remain explicit release blockers;
no remaining job is running for this candidate. Other audit findings remain tracked
independently. Do not turn these partial results into a release acceptance claim.

Frozen color candidate scope (not this release batch): explicit system color-picker presentation in tag editing,
current-owner selection events, disabled-state handling and optional color reset.
Acceptance covers ordinary first taps and touch area, RGB channel retention,
dismissal, lock/unlock, Add draft retention and the Settings save workflow on
phone/iPad. Text-entry comparisons and other audit findings are outside this batch.
The TestFlight note will be: “Improved custom tag color selection with the iOS
system color picker, while preserving your draft and preset colors.”

## Current release — menu action ownership

Sourceb75ef748 fixes retained native menu callbacks executing obsolete actions or
reopening after unlock. M54 in [findings](findings.md) holds the current diagnosis.
Production scope: current committed action/trigger ownership on all adapters,
global/item disabled state, removed items and teardown, controlled popup dismissal.
Policy-only follow-up e3159d32 carries the bounded-investigation rules independently
of the frozen color candidate. No other product fixes are entering this batch.

Remote mounted regressions, TypeScript, structural checks and code critic pass.
[Native35815492811](https://github.com/elsell/stuffstash/actions/runs/35815492811)
and [CI35815495522](https://github.com/elsell/stuffstash/actions/runs/35815495522)
testb75ef748. All six CI jobs pass. Native results are terminal: phone2/3, iPad3/3.
[Exact results](native-menu-358154-results.csv) preserve the phone archive-notice failure. Android uses the
existing disposable audit tree and emulator, with a scripted lock/recovery journey.
Android lock/recovery and applied Browse choice pass; retained evidence is linked
from M54. Phone/iPad menu and filter checks also pass. Phone archive returned to
Tags but queried its4200ms notice about ten seconds after confirmation. The
corrected observation order passes the focused archive journey on both devices in
[native35816959842](https://github.com/elsell/stuffstash/actions/runs/35816959842),
retaining the exact notice and destination assertions. All six final CI jobs pass.
PR162 merged asc31638fc; [release35818700192](https://github.com/elsell/stuffstash/actions/runs/35818700192)
completed successfully. Upload succeeded September23 at04:55:29UTC; Apple
processing and exact-build changelog readback passed at04:59:56UTC. This releases
the frozen menu fix, not the held color candidate or the later fixture-only PR164.

Acceptance: shared menu open/lock/dismiss/unlock/execute behavior, representative
filter selection and destructive-command recovery on phone/iPad, plus Android
popup/choice behavior. No new color/input diagnostic runs. TestFlight note:
“Fixed menus accepting outdated actions and reopening unexpectedly after a task unlocks.”

## Accepted follow-up — onboarding command clearance

M35's earlier fixture inserted a navigation header that production onboarding does
not have. The corrected fixture preserves the production viewport and keeps its
observer outside layout flow. Android passes full Connect clearance above the
visible keyboard and exact one-tap URL submission without a product change.
[Corrected iOS acceptance35818325756](https://github.com/elsell/stuffstash/actions/runs/35818325756)
passes2/2 on phone and iPad: full keyboard-open Connect clearance, exact one-tap
submission and keyboard Go. Reviewed [phone](evidence/phone-onboarding-keyboard-clearance-358183.png)
and [iPad](evidence/ipad-onboarding-keyboard-clearance-358183.png) captures agree.
The prior header-bearing phone failure does not establish a shipped layout defect.
PR164 merged as2d9efd02, retaining the regression and evidence. No production
workaround is needed for M35. Larger fonts and other configurations
retain their separate audit scope.

Android missing-photo recovery also passes on the same existing audit APK:
readable error, native Retry, retry failure and Close returning to the parent.
[Capture](evidence/android-photo-recovery-retry.png) and
[returned hierarchy](evidence/android-photo-recovery-closed.xml) retain evidence.
The initial script's final lookup used iOS title casing; inspecting the captured
Android uppercase label confirms return, without an unnecessary rerun. This does
not establish successful image retry, zoom, backgrounding or assistive behavior.

## Current release batch — Move destination recovery

PR165 fixes M254: confirmed creations are merged with current search results by
ID, filtered by query and retained as the selected destination. The failing mounted
regression,90 focused tests, TypeScript, structural checks and critic establish the
source correction. Android APKc62b07ae passes existing selection, native Container
choice, rejected creation/retry, no duplicate creation offer, selected new row,
rejected Move retention and successful exact-payload retry. [Reviewed capture](evidence/android-m254-retained.png)
and [returned hierarchy](evidence/android-m254-returned.xml) retain evidence.

Native35826352235 stopped phone scenarios during Xcode launch and iPad Move in a
key-enumeration precheck. Its scoped observer correction preserved field hittability,
keyboard presence, one typing attempt and exact text. Native35828644735 then exposed
M255 in the real Move journey: Audit crate became Ae on phone and A on iPad before
creation. [Exact outcomes](evidence/native-move-field-358286-results.txt). CI atbcbb83ed
passes. These are failed acceptance runs, not release evidence.

Current decision: the investigation budget is exhausted. Reuse Add's proven
SwiftUI draft field for iOS Move query; do not repeat input/provider/timing
comparisons. The shared adapter preserves native editing state; Move advances a
query revision only after successful creation to apply the returned canonical
name. Android keeps its existing input. Review requested an actual Move regression
for canonical naming, unchanged field through typing/rejection and reset after
success; it is added and its negative control fails without revision advancement.
All1,925 remote tests (306 files), TypeScript and structural checks pass. Native
acceptance35831661267 passes2/2 on both phone and iPad: exact one-attempt text,
creation/movement recovery and the representative Add regression. All six CI jobs
pass at827fd875. [Terminal evidence](evidence/native-move-add-358316-results.txt)
and retained captures establish scoped acceptance. PR165 merged as00e8e032;
a sleeping collector owns release observation. TestFlight availability is not yet
verified. No further field-choice rerun is needed.

### Custom-field choices — scoped acceptance complete

M02/M11's real create-form composition passes all intermediate selections on both
phone and iPad in35828644735: Type/Applies to choices, first/last target selection,
removing the first and retaining the last. [Reviewed phone capture](evidence/phone-field-retained-target-358286.png). Android APKbe77f1fa passed the same
sequence; [capture](evidence/android-field-choices-retained.png) and
[hierarchy](evidence/android-field-choices-retained.xml) retain evidence. Earlier
label and passive-observer failures remain recorded in commits984cba79/2f421bf4
and their retained evidence. Field persistence, saved-target immutability and
assistive modes were not part of this scenario; existing source tests retain those
boundaries. Do not rerun this accepted shared-control composition unchanged.

For each concrete correction, run one focused native acceptance pass. If it fails,
use the specific failed gate to choose the next correction; do not reopen broad
experiments without naming competing causes and the decision each outcome changes.
Use [the investigation policy](../../../specs/platform/mobile-comprehensive-ui-audit.spec.md).
Do not reset budgets across task continuations or weaken exact-value acceptance.

This is the sole current summary. Older report checkpoints are durable historical
evidence; their pending-job and unreleased-candidate statements are not current
status. Update this summary in place rather than adding another checkpoint report.
