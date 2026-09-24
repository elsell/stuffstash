# Mobile UI audit — current state

The comprehensive audit is **incomplete**. Prioritize stable screen structure,
connected everyday tasks, visual coherence, then detailed states. Normal text comes
first. The surface/axis inventory checks omissions; it is not a separate test queue.

## Delivery

Latest verified TestFlight is **0.24.33 (124.1)**, with Apple processing and
exact [changelog readback](evidence/selection-release-359389-results.txt).

**M265–M273 is merged** in [PR173](https://github.com/elsell/stuffstash/pull/173),
e0d63ad5. Release35938987211 completed with eight TestFlight notes. Required
CI35936948141 passed at4b98eac2. Final Add35936945060 passes phone/iPad, including
search, cancel/reopen, destination creation/retry and bounded form spacing; final
creation and returned-draft captures were reviewed. Move35920806955 passes all
four workflows on both devices; tag/draft35920079319 passes three on both devices.
Android grouped creation and recovery were reviewed. Add's scoped design hold was
closed. Move remains visually open following user feedback; functional passes do
not override that judgment. Current source already has grouped native rows,
toolbar actions and stacked iPhone/iPad search; assess those actual captures before
repeating the same redesign. The comprehensive audit and physical-device checks
remain incomplete.

## Current follow-up diagnosis and decisions

The release hold remains open for Move's visual acceptance. PR174 consolidates
M274–M279 plus Settings current-save/readback and Move hierarchy/contrast into one
release candidate. The combined source passes1,995 mobile checks, TypeScript,
structural checks and ten fixture-preparation checks. Its product source matches da791006; the additional Add test
observation fix does not change product code. M274–M279 covers filter navigation/density, retained search, contextual detail actions and
Sharing recovery. Standard adaptive filters corrected the iPad viewport defect;
do not repeat rejected dependency ownership patches. Integrated35941028517 passed
14/15 phone and15/15 tablet cases. The conditional keyboard-dismiss observation
fix16501db7 passed focused Add/Move35945429162 on both devices; required
CI35945384623 passed. Phone/iPad Add creation and return captures are reviewed: native actions are
reachable, the item draft survives and the new destination is selected.

Settings current-save246edefd passes connected readback35948277688 on both devices,
including exact full-name persistence after failure/retry, reopen/create/archive.
Phone and tablet collection/reopened-editor captures confirm header clearance and
complete names. This closes those scoped defects, not all Settings design findings.

Move8e79997c passes all four workflows on both devices in35947975650. Visual review
found the custom section header too faint. Contrast correctionda791006 is in
35951755320 on `codex/mobile-move-contrast`; native acceptance remains open.

Detail M280 at a7ea4ac5 passed five of six cases per device in35948366265, including
resized gallery selection. Photo Retry failed to disappear on phone; contents Retry
failed on tablet. Preserve the assertions. Candidate615485f5 adds bounded native
buttons by default and puts identity before photo recovery. Run35951671413 passes
all nine phone workflows; iPad has two application-launch failures before Add and
region recovery, with seven workflow passes. The phone capture nevertheless shows
Retry contents compressed into a three-line oval. Correction5dc08f9c gives the
native host the available width and keeps the button leading-aligned inside it;
normal-text Retry geometry assertions accompany the existing completion checks.
Its1,992 source tests, TypeScript, structural checks and review pass;35954709640
is the required new native check. Preserve the sizing diagnosis; do not rediscover
provider or key-delivery hypotheses. This candidate is not visually accepted.

Persistent tabs a1d7d825 is in35951026134. Source checks establish route ownership,
not native retention. A subsequent source correction carries the originating tab
through Expiration Filters because root modal segments cannot disambiguate shared
routes. That correction still needs a native roundtrip. Android, dark appearance,
long command labels and the wider surface audit remain open. The button/detail and tab candidates remain outside this frozen release batch.

## Separate unresolved decisions

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
