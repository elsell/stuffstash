# Mobile UI audit — current state

The comprehensive audit is **incomplete**. Prioritize stable screen structure,
connected everyday tasks, visual coherence, then detailed states. Normal text comes
first. The surface/axis inventory checks omissions; it is not a separate test queue.

## Delivery

Latest verified TestFlight is **0.24.34 (125.1)**. Release35954982498 completed;
changelog job107497620643 verified exact readback after Apple processing.

Earlier release evidence remains in the findings ledger and linked release records.
Current release acceptance below supersedes historical pending statements.

## Frozen release — PR178

`codex/mobile-native-patterns-batch` at5040afd4 contains persistent Home/Browse
stacks for41 ordinary routes, bounded native commands, readable detail layout,
expiration search retention and Settings name/key corrections. Add/Edit/Move/Filters
remain modal tasks. Required CI35970458935 passed. The remaining release gate is
native settings-tab-return35970467285 on iPhone and iPad, including exact creation,
save/reopen and tab-return behavior. Do not infer acceptance of all41 routes.

The integrated baseline35958732480 passed Add and expiration on both devices;
follow-up35961716802 closed contents Retry on both. Native Name uses a draft field
to avoid keystroke reordering. Subsequent35965781821 passed iPad tab return and
exposed false dirty state after reopening a renamed setting on both devices.
5040afd4 preserves existing stable keys during name initialization/editing; a
rendered regression reproduced the failure, then64 focused checks, TypeScript,
structural checks and review passed. The iPhone tab check in35965781821 stopped
on screenshot acquisition after exact draft return, not a demonstrated touch bug.
The current native gate must close both remaining checks before release.

Android connected detail/Move/tab return, Settings draft retention and expiration
query/filter return passed. Reviewed iOS command-width captures show readable
bounded recovery actions. These scoped results do not certify the whole app.

## Follow-up — PR179

`codex/mobile-photo-free-browse` is stacked on PR178. Sparse photo-free Browse rows
compact without collapsing mixed-media rows. Native35964382077 passed all three
List/Map geometry and return checks on both devices; Android sparse/mixed return
also passed. Evidence: [Browse density](evidence/photo-free-grid-review.txt).

Shared Settings commands use bounded native buttons. Filter35965504659 passed
six workflows on each iOS device; reviewed date, overview and last-tag captures
show clear commands and footer clearance. Android ordinary/320dp filter reset and
reminder Retry/Discard passed. Icon-free separators now use the normal row inset.

The paired reminder test35967919825 then exposed iOS button hit-area overlap:
62-point styled buttons occupied48-point hosts.1661a240 reduces label minimums
while retaining48-point outer commands. Four adapter checks, TypeScript,
structural checks and critic review pass. Native35971374386 must establish separate
hit areas, successful Retry/Discard and long-label activation before acceptance.
See [command evidence](evidence/bounded-settings-actions.txt). This follow-up does
not expand PR178's release gates. Its Settings name/key corrections are integrated.

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
