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

M274–M279 covers filter navigation and density, retained search, contextual detail
actions and Sharing recovery. Source7236f442 in35934741356 passes all12 phone
workflows and11/12 iPad workflows. Reviewed phone detail and Sharing captures show
contextual commands, unclipped completion actions and appropriate grouping.

The last reproduced native defect was iPad Expiration inheriting a580×650 viewport after
Filters. Neither dependency ownership guard corrected it. Sourcef69eae7e removes
those unproven patches and uses standard adaptive iOS modals for both filter tasks;
Android stays a card. Filters previously opened at their largest custom detent.
Keep the verified search patch and all connected viewport/mode/detail/back gates.
Native35937584242 passes all twelve workflows on both devices. Reviewed captures
confirm full Expiration viewport and three reachable modes, plus filter/tag footer
clearance. The explicit medium-sheet fixture
remains an iOS-only diagnostic for other form-sheet consumers.

The follow-up now integrates PR173's final Add/Move selection and stacked search
with its own search focus ownership and detail hierarchy. Review removed a duplicate
empty-photo caption introduced by the merge. The full1989-test mobile suite,
TypeScript, structural checks and10 fixture preparation checks pass; critic review
is complete. These source checks do not establish integrated native acceptance.
The integrated release subset adds three representative Add destination, Add tag
and Move creation workflows to the twelve follow-up workflows for shared search
and return behavior; it retains all existing assertions. Native35941028517 at
b605b6e9 is running this integrated subset. PR174 stays draft until it passes and
its visual review is complete.

Connected Settings readback is a subsequent batch. Its fixture now uses production
cache invalidation;35937082802 confirms the updated row exists but is behind the
native header on both devices. Candidate35bba93c preserves one ScrollView across
loading/ready/retry instead of replacing its root. Source checks pass; native
35939654846 is verifying it. This is a candidate correction, not a proven fix.
This does not block the selection release or expand the frozen follow-up.
PR176 stacks Settings readback and clearer action grouping with native primary
emphasis for Move completion. Its1,993 integrated tests and source checks pass;
native verification is conditional on the existing Settings and follow-up runs.

A subsequent detail-layout candidate addresses M280: independent tablet row caps
produce unrelated trailing edges. One centered content column and gallery viewport
measurement are implemented separately on `codex/mobile-detail-readable-layout`.
Source checks pass; native visual and resized-gallery verification remain open.
This does not expand either frozen batch.

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
- Settings persisted collection readback, physical integrations, assistive behavior
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
