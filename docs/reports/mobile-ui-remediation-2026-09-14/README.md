# Mobile UI audit — current state

The comprehensive audit is **incomplete**. Prioritize stable screen structure,
connected everyday tasks, visual coherence, then detailed states. Normal text comes
first. The surface/axis inventory checks omissions; it is not a separate test queue.

## Delivery

Latest verified TestFlight is **0.24.37 (128.1)**. PR184 merged8bdcc289;
release36020469905 and iOS upload107707509064 passed. Apple changelog readback
was verified2026-09-24 at16:01UTC. Home collection shortcuts now use Browse with
explicit collection criteria; Android photo dialog contrast is corrected within
the recorded API36 scope. [Release evidence](evidence/release-128.txt).
Prior127 delivered Browse retention/alignment and clearer container detail grouping
with tab/voice clearance. [Prior release](evidence/release-127.txt).

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

## Home collection batch acceptance

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
stale-state observation findings were corrected. Native36013639520 stopped in fixture setup; corrected36017420318 passes on both
devices at3f7485e3. Reviewed captures show the correct List/tab,25 recent items and6
checked-out items with only the intended filter. CI36017423895 passes the same
revision. PR184 merged8bdcc289; release36020469905 delivered128.1 with verified notes. [Evidence](evidence/home-collections.txt).

Separate follow-up: Home's inventory switcher now uses available header space up
to320pt instead of always capping at180pt, avoiding unnecessary truncation when
only Profile is visible. Viewer regression failed before the change; six width
checks, structural checks and critic review pass. Native36018428143 at452265ab passes five cases on each device; reviewed captures
confirm one-action/three-action spacing and stable actions on scroll. This is
outside PR184 and is not released. [Evidence](evidence/home-header-space.txt).

## Next connected review

Settings overview review confirmed iPhone root labels crowded out by long values,
Appearance flush with the group edge, and repetitive scope subtitles. Candidate
uses descriptive subtitles, shared inset rows and scope context once.83 focused
tests, TypeScript,12 preparation tests, structural checks and critic review pass.
Native36019282915 reached Diagnostics but stopped on duplicate selectable-text
AX nodes; the locator is corrected without relaxing geometry checks. Corrected36023217961 passes all three cases on each device. Reviewed final
captures confirm hierarchy, footer clearance and tab return; Account/Connection
recovery checks also pass. PR186 is accepted for release, not yet delivered.
[Diagnosis and evidence](evidence/settings-overview.txt).

## Current workflow investigation

History baseline5068ae30 adds a production-tab walkthrough. Native36027545902
is running; watcher58712 writes `/tmp/history-native-result.json`. Android review
confirmed dense multi-field paragraphs in the list, while final metadata remains
reachable. Candidatea07a25c5 replaces those paragraphs with changed-field summaries
and separates Before/After values in detail.23 focused tests, TypeScript,13 fixture
preparation tests, structural checks and critic review pass; corrected Android
captures are reviewed. iOS header/footer geometry remains unconfirmed; do not
change insets based only on source. [Current diagnosis](evidence/history-structure.txt).

PR186 final CI36026968447 passes all six checks; auto-merge was enabled.
Release watcher session16214 writes `/tmp/settings-release-terminal.json` when
its matching main release terminates. Do not start duplicate watchers or treat
an observation timeout as a release failure.

## Separate unresolved decisions

- **M20 dismissal timing:** current iPhone onboarding refresh timed out awaiting
  keyboard disappearance; final capture/tree shows it dismissed and address intact.
  iPad passed. Timing remains unverified; no input rewrite or unchanged rerun.
  [Evidence](evidence/onboarding-current-dismissal.txt).

- **M251 Android photo status:** dialog-owned native adapter shipped in128.1 and shows
  readable light glyphs on the dark canvas in reviewed API36 captures. Close and
  Android Back restore the underlying screen in light/dark appearance. Final-photo
  disposal/restoration also passes both. Reviewed captures are retained with the
  evidence. Swipe replay failed both directions and remains unverified after the
  bounded investigation; older Android is not certified. No Activity-level style workaround was restored. [Evidence](evidence/android-photo-status.txt).

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
