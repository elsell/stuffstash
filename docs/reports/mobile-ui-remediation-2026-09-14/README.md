# Mobile UI audit — current state

Latest verified TestFlight: **0.24.27 (116.2)**; [release and changelog evidence](release-batch-116.md).
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
| M51: custom color picker sometimes does not open | Ordinary well taps intermittently leave the parent unchanged. Disabling scrolling is insufficient. RGB retention and disabled-state fixes already passed. [Evidence](native-color-scroll-352164.md). | Diagnostic budget exhausted. Implement a direct UIKit color-well adapter using the system picker, replacing the current SwiftUI-hosted path for this control. Verify first-tap opening, selection, dismissal, parent draft retention, clear and lock/unlock on phone/iPad. Retain presets. |
| Text loss in RN input comparisons | Complete key sequences can yield missing/reordered JS values. Verified provider-free iPad still fails; pacing or removing assistance is not a general correction. [Consolidated evidence](native-text-entry-352471.md). | Diagnostic budget exhausted. Keep existing native Add-name, invitation-email and onboarding adapters. Verify actual editing workflows and fix a reproduced consumer with the established native field pattern; do not replace every input because an isolated comparison fails. Preserve external reset, draft, disabled and submission semantics. |

Run35803226783 completed: phone12/14 and iPad9/14. Its
[exact outcomes](native-text-entry-358032-results.csv) reproduce the known isolated
RN input failures; this does not change the decision above. No further trace
instrumentation is planned.

M51 candidate: direct UIKit adapter implemented, with current-owner event handling
and native enabled state. Remote validation passes1,921 tests across305 files,
TypeScript and structural checks; regressions were observed failing before fixes.
Code critic has no remaining source blocker. CI35805414085 passes all six jobs,
including compiled Swift RGB checks and the committed native dependency lock.
Native35805411868 at4e906d30 passes1/9 on each device. Seven color checks stop
before activation because UIKit exposes its internal button as “Color” alongside
the separate visible text ([hierarchy](evidence/phone-color-accessibility-358054.txt)).
Correction: expose the well itself as the named accessible button and hide the
visual label from traversal; retain exact native acceptance. The Add check also
stops before typing: its exhaustive key query finds readiness but takes5.5seconds
([timing](evidence/phone-add-keyboard-readiness-358054.txt)). Check the intended
T/C key instead, retaining the five-second deadline and exact text assertions.
Remote1,921 tests, TypeScript and11 fixture checks pass; critic reports no source
blocker. Corrected native activation/disabled-state acceptance remains required.
This candidate is not released.

Frozen release scope: direct system color-well presentation in tag editing,
current-owner selection events, disabled-state handling and optional color reset.
Acceptance covers ordinary first taps and touch area, RGB channel retention,
dismissal, lock/unlock, Add draft retention and the Settings save workflow on
phone/iPad. Text-entry comparisons and other audit findings are outside this batch.
The TestFlight note will be: “Improved custom tag color selection with the iOS
system color picker, while preserving your draft and preset colors.”

For each concrete correction, run one focused native acceptance pass. If it fails,
use the specific failed gate to choose the next correction; do not reopen broad
experiments without naming competing causes and the decision each outcome changes.
Use [the investigation policy](../../../specs/platform/mobile-comprehensive-ui-audit.spec.md).
Do not reset budgets across task continuations or weaken exact-value acceptance.

This is the sole current summary. Older report checkpoints are durable historical
evidence; their pending-job and unreleased-candidate statements are not current
status. Update this summary in place rather than adding another checkpoint report.
