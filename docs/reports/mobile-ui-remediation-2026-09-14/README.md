# Mobile UI audit — current state

The comprehensive audit is **incomplete**. Prioritize stable screen structure,
connected everyday tasks, visual coherence, then detailed states. Normal text comes
first. The surface/axis inventory checks omissions; it is not a separate test queue.

## Delivery

Latest verified TestFlight is **0.24.36 (127.1)**. Combined PR182 merged as
main0bb1f24e; candidate CI36007310510 and release36008181746 passed. Apple changelog
readback was verified2026-09-24 at14:18UTC. [Release evidence](evidence/release-127.txt).
This delivers Browse retention/alignment, detail header insets and clearer
container identity/contents grouping with bounded commands. The prior persistent
tabs, filters and editing batch remains recorded in [release126](evidence/release-126.txt).

## Shipped batch acceptance

PR182 included PR181: Browse search retention/appearance, neighboring card
alignment, detail header insets and grouped container actions. Standard CI36000041858,
focused source tests, TypeScript, structural checks and critic review pass.

Native acceptance: Browse/filter workflows35979200078 pass6/6 both. Frozen
35987800972 passes phone15/16 and iPad14/16: its32pt Move-items target is corrected
by PR182 and passes in hierarchy35994194799 (six existing cases pass both).
The initial footer test had missing fixture metadata, not demonstrated clipping.
Corrected populated Detail/Sharing35999231112 at28bc9b3c passes2/2 both; iPhone
needed one retry after Xcode failed before app launch. Reviewed screenshots show
final content above delivered tabs/voice chrome. Product code is unchanged since
that native revision. Sharing needs no speculative inset change on this evidence.
[Evidence](evidence/container-organization.txt).

Android appearance and connected Edit/Move checks pass on the named audit APK.
[Appearance](evidence/android-live-appearance-header.txt),
[alignment](evidence/checkout-row-alignment.txt),
[detail entry](evidence/detail-entry-insets.txt).

Frozen iPad Add entry readiness remains unverified; phone Add/Move recovery passes.
No product rewrite or broader acceptance is inferred from the timing failure.
Recent-assets/invitation inset risks and unrelated findings remain in the audit.
Broad refresh35980094051 completed121 cases per device (phone14 failures, iPad6),
including old diagnostics and enlarged-text cases; it is not a release gate for
this verified batch. [Terminal decisions](evidence/native-full-359800.txt).

## Next batch in progress

Home Recently changed → See all now targets existing Browse List with explicit
all-active/recent ordering and clears prior query/tags/kind/availability criteria.
This removes the everyday entry to the older reduced-function asset grid; its
legacy route remains available. Home navigation test failed before implementation;
Home28, route6 and mounted Browse13 tests pass. TypeScript/structural checks pass;
critic found no source blocker. Native connected acceptance remains required:
start with refined Browse, return Home, use See all, open an asset and return;
verify selected tab, criteria, scroll clearance and ordinary tab context retention.
Checked out now uses the same explicit reset contract, differing only in availability.
Home28 and shared fixture3 tests pass; fixture preparation11 tests pass. Focused
`home-collections` acceptance uses production tab paths and shared real screens,
with no competing root Search route or placeholder. Critic's root-header and
stale-state observation findings were corrected. Native execution remains pending.
Include these in the next meaningful batch, not isolated releases.

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
