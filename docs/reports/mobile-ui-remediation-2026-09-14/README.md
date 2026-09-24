# Mobile UI audit — current state

The comprehensive audit is **incomplete**. Prioritize stable screen structure,
connected everyday tasks, visual coherence, then detailed states. Normal text comes
first. The surface/axis inventory checks omissions; it is not a separate test queue.

## Delivery

Latest verified TestFlight is **0.24.34 (125.1)**. Release35954982498 completed;
changelog job107497620643 verified exact readback after Apple processing.

Earlier release evidence remains in the findings ledger and linked release records.
Current release acceptance below supersedes historical pending statements.

## Merged release batch — PR178/179

Both PRs are merged into main262f27f6. Integrated CI35976251247 passed;
release35976929170 is pending terminal verification. The scoped evidence below
is accepted for this batch; unrelated findings remain open.

`codex/mobile-native-patterns-batch` at5040afd4 contains persistent Home/Browse
stacks for41 ordinary routes, bounded native commands, readable detail layout,
expiration search retention and Settings name/key corrections. Add/Edit/Move/Filters
remain modal tasks. Required CI35970458935 passed. Run35970467285 now passes
Settings creation/save/reopen on both devices and the complete iPhone tab workflow.
iPad touch switching and exact draft return pass, but the final hit-point waiter
fails while its ancestor locator repeatedly retries resolution. Test-only579fb767
uses direct native tab labels and preserves every acceptance assertion. Focused
35973624529 now passes both devices, with final captures reviewed. CI35973630755
also passes. The follow-up is merged and final integrated CI passed. Do not repeat the old
ancestor-query experiment or infer acceptance of all41 routes.

The integrated baseline35958732480 passed Add and expiration on both devices;
follow-up35961716802 closed contents Retry on both. Native Name avoids keystroke
reordering, and5040afd4 preserves existing stable keys on name initialization/edit.
Their exact create/save/reopen verification now passes on both iOS devices.

Android connected detail/Move/tab return, Settings draft retention and expiration
query/filter return passed. Reviewed iOS command-width captures show readable
bounded recovery actions. Android successful Edit/save/reopen and Move/save/reopen
also pass on the named audit APK: [connected evidence](evidence/android-connected-edit-move.txt).
These scoped results do not certify the whole app.

## Follow-up — PR179

`codex/mobile-photo-free-browse` is stacked on PR178. Sparse photo-free Browse rows
compact without collapsing mixed-media rows. Native35964382077 passed all three
List/Map geometry and return checks on both devices; Android sparse/mixed return
also passed. Evidence: [Browse density](evidence/photo-free-grid-review.txt).

Shared Settings commands use bounded native buttons. Filter35965504659 passed
six workflows on each iOS device; reviewed date, overview and last-tag captures
show clear commands and footer clearance. Android ordinary/320dp filter reset and
reminder Retry/Discard passed. Icon-free separators now use the normal row inset.

The paired reminder test35967919825 exposed overlapping iOS command hit areas.
1661a240 corrects sizing; native35971374386 now passes Retry, Discard, separate hit
frames and long-label activation on both devices. Captures reviewed for spacing
and legibility: [command evidence](evidence/bounded-settings-actions.txt).
Integrated source passes2,006 mobile tests, TypeScript, structural checks and
review. The branch includes PR178's product fixes and current test-only locator
correction. Its scoped native gates and the base tab gate now pass; final integrated CI passed;
TestFlight delivery remains unverified.

## Current follow-up

Android live appearance still loses native Browse controls. A bounded lifecycle
trace excludes native screen recreation in the reproduced Map sequence. A separate
mounted regression proves query-adapter replacement incorrectly resets Browse
state; its correction passes12 behavior tests and TypeScript. Android native replay now retains the query through dark/light changes, continued
typing and clear. Missing List/Map and poor search contrast remain unaccepted. [Single diagnosis](evidence/android-live-appearance-header.txt).

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
