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

## Persistent tab navigation candidate

The user requested persistent Home/Browse navigation on ordinary screens. The
`codex/mobile-persistent-tabs` candidate moves41 ordinary routes into shared tab
stacks, retaining URLs and root modal tasks. The real Expo route expansion test
reproduced the old ownership and passes for both tabs; source guards and fixture
preparation remain checked. Native history, modal return, editor draft retention
and bar clearance are not yet verified. Surface paths reflect the new ownership;
older matrix runtime evidence does not establish this navigation structure.

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
native header on both devices. Run35939654846 disproves stable scroll ownership alone: the phone row still
starts at y24 behind the header. PR176 now reserves an explicit measured iOS
viewport and disables automatic content insets; Android keeps its hierarchy.
The122 affected tests, TypeScript, structural checks and critic review pass.
Native acceptance remains open; entry/return bounds also reject doubled spacing.
This does not block the selection release or expand the frozen follow-up.
PR176 stacks Settings readback and clearer action grouping with native primary
emphasis for Move completion. Its1,993 integrated tests and source checks pass;
run35945430640 passes the phone edited-row return geometry, then persists
Campin after the native field reported Camping. A retained Save event reproduces
stale submission in the mounted editor. The scoped committed-action guard passes
59 Settings tests, TypeScript and structural checks; exact native persistence
acceptance remains open. Keep provider experiments closed. The
failed Settings candidate is superseded, not rerun. Check Settings search/Add as
well as collection readback and the four Move workflows.
The user rejected Move visual acceptance. The next layout replaces its static
subject card with a compact section heading and constrains the iPad list to a
readable column. Search remains visible and completion prominent. Source checks
and critic review pass; native visual acceptance and release remain on hold.

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

[Surfaces](surfaces.json), [axes](axes.json) and [matrix](matrix.csv) enumerate142
surfaces ×24 axes. Matrix classifications describe evidence, not3,408 separate test
requirements. [Findings](findings.md) retain stable IDs and historical evidence.
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
