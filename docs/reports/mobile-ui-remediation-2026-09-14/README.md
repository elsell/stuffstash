# Mobile UI audit — current state

The comprehensive audit is **incomplete**. Prioritize stable screen structure,
connected everyday tasks, visual coherence, then detailed states. Normal text comes
first. The surface/axis inventory checks omissions; it is not a separate test queue.

## Delivery

Latest verified TestFlight is **0.24.35 (126.1)**. PR178/179 merged into
main262f27f6; integrated CI35976251247 passed. Release35976929170 completed and
changelog job107568033135 verified Apple readback. [Release evidence](evidence/release-126.txt).

This batch delivers persistent Home/Browse stacks for41 ordinary routes, bounded
native commands, detail hierarchy, compact photo-free Browse rows, expiration
search retention, Settings name/key corrections and filter/footer fixes. Add,
Edit, Move and Filters remain modal tasks. Scoped acceptance:

- Persistent tab switching/draft return35973624529 passed phone/iPad after correcting
  the transient ancestor locator; all hit-test assertions remain. Settings
  creation/save/reopen35970467285 passed both devices.
- Browse density35964382077 passed3/3 both; filters35965504659 passed6/6 both.
- Paired native Settings command hit regions/recovery35971374386 passed both after
  sizing correction. Captures reviewed for spacing and legibility.
- Add/expiration35958732480 and contents Retry35961716802 passed both. Android
  connected Edit/save/reopen and Move/save/reopen passed on the named audit APK.
- Integrated2,006 mobile tests, TypeScript, structural checks and review passed.

Evidence: [tabs](evidence/persistent-tab-touch.txt),
[commands](evidence/bounded-settings-actions.txt),
[density](evidence/photo-free-grid-review.txt),
[Android connected tasks](evidence/android-connected-edit-move.txt).
These scoped results do not certify every route or the whole app.

## Current follow-up

Android Browse now retains List/Map and search through light/dark changes on the
named emulator APK. Corrected adapter-driven route resets and native control
appearance; reviewed captures, continued editing/clear/close, return to List and
Notifications selection pass.26 focused tests, TypeScript, structural checks and
review pass. PR181 remains draft. Connected iOS Browse/filter regression35979200078
at77c4738d passed6/6 on both phone and iPad. Reviewed return captures exposed
misaligned compact-card titles with mixed checkout status; the row-space correction
passes Android native alignment and48 focused tests. Place entry also hid its
identity under native chrome; automatic detail insets now have68 passing detail
tests and Android entry/Edit/save/Move/return verification. Frozen iOS run35987800972 at3c0ba0ff
replaced canceled pending35983649762 on an independent verification branch.
iPad passed14/16: reviewed captures confirm Place identity below the header and
aligned Browse titles. Its32pt Move-items command fails the44pt gate; PR182
addresses that control. Add destination stopped at entry readiness although the
final capture shows its correctly labeled field; the recovery workflow remains
unverified. Phone result remains pending.

PR182 groups availability with identity and places bounded contents commands under
a short heading.80 focused tests, TypeScript, structural checks and review pass;
Android capture confirms both commands after rejecting a recycled-row candidate
that hid them. iOS run35994194799 atb891a13e verifies hierarchy and real-tab footer
clearance. Later test-only8499d7c2 adds populated Sharing inside production tabs;
it has not run natively. Sharing/list missing-inset source risks are not confirmed
runtime defects. [Current diagnosis and evidence](evidence/container-organization.txt).

Broad refresh35980094051 atc7c44b45: phone completed121 tests with14 failures.
Failures include historical controlled-input and sheet diagnostics, enlarged-text
cases, color first-tap and the same32pt Move-items command. Do not treat them as14
new product defects or repeat the diagnostic experiments.
iPad completed121 tests with6 failures, including Add keyboard readiness.
[Terminal results and decisions](evidence/native-full-359800.txt).
The broad sweep is not a prerequisite for this bounded release. [Detail evidence](evidence/detail-entry-insets.txt),
[alignment evidence](evidence/checkout-row-alignment.txt),
[Android appearance evidence](evidence/android-live-appearance-header.txt).

## Separate unresolved decisions

- **M20 dismissal timing:** current iPhone onboarding refresh timed out awaiting
  keyboard disappearance; final capture/tree shows it dismissed and address intact.
  iPad passed. Timing remains unverified; no input rewrite or unchanged rerun.
  [Evidence](evidence/onboarding-current-dismissal.txt).

- **M251 Android photo status:** normal-text native review confirms dark status
  glyphs on the black photo canvas. Activity-level styling does not fix the modal;
  that candidate was withdrawn. Close/return works. Next correction needs dialog
  ownership. [Bounded evidence](evidence/android-photo-status.txt).

- **M51 color selection:** [PR161](https://github.com/elsell/stuffstash/pull/161)
  remains draft at aa9fbd9e. The explicit system color-picker candidate passes
  ordinary opening on both devices, but phone coordinate activation and timely Add
  observation remain unverified in35812501085. Investigation budget is exhausted:
  retain the implementation and the exact failed gates; do not swap controls or
  repeat unchanged diagnostics. [Results](native-color-358125-results.csv).
- **Text-input comparisons:** provider removal and paced entry did not establish
  a general correction. Fix reproduced consumers using the established native
  draft field, preserving reset and ownership semantics. Do not repeat the same
  provider/key experiments. [Consolidated evidence](native-text-entry-352471.md).
- Remaining Settings design findings, physical integrations, assistive behavior
  and wider device adaptations remain tracked in the full findings and surface
  reports. Passed fixture workflows do not close them.

## Coverage and evidence limits

[Surfaces](surfaces.json), [axes](axes.json) and [matrix](matrix.csv) enumerate144
surfaces ×24 axes. Matrix classifications describe evidence, not3,456 separate test
requirements. The two recent selection routes are now included; `unreviewed`
marks axes that their scoped workflow evidence does not establish. [Findings](findings.md) retain stable IDs and historical evidence.
The older [whole-workflow review](everyday-workflow-review.md) records design
rationale; this file supplies current acceptance status.

The last full native sweep, [352471](native-full-352471.md), passed phone74/92 and
iPad84/92 fixture cases. It predates subsequent fixes and is neither a current
failure count nor whole-app certification. Verify shared controls once, representative
compositions and critical connected workflows; add coverage when ownership differs.


This is the sole current status summary. Update it in place. Keep exact diagnoses,
decisions and durable evidence; historical pending-run statements elsewhere are not
current state. Sleeping scripts collect terminal job results without unchanged
status narration. Do not weaken acceptance to make a batch pass.
