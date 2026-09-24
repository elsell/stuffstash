# Mobile UI audit — current state

The comprehensive audit is **incomplete**. Prioritize stable screen structure,
connected everyday tasks, visual coherence, then detailed states. Normal text comes
first. The surface/axis inventory checks omissions; it is not a separate test queue.

## Delivery

Latest verified TestFlight is **0.24.34 (125.1)**. Release35954982498 completed;
changelog job107497620643 verified exact readback after Apple processing.

Earlier release evidence remains in the findings ledger and linked release records.
Current release acceptance below supersedes historical pending statements.

## Persistent tab navigation candidate

PR178 (`codex/mobile-native-patterns-batch`) contains persistent Home/Browse
stacks for41 ordinary routes, bounded native commands, readable detail layout and
expiration search retention. Add/Edit/Move/Filters remain modal tasks. Product
source is35a79fe5;7ee10632 strengthens native observation without changing the app.

The iPhone Settings draft/tab-return check passed in35955427236. Its iPad capture
shows tabs, but nested duplicate accessibility buttons made first-match hit testing
ambiguous. The correction chooses a hittable native match and requires an actual
Browse/Home roundtrip with exact draft readback. Focused run35960677440 verifies
that correction. Combined13-workflow run35958732480 remains the batch's native gate.
Required CI35959208262 passed at35a79fe5; CI35960682844 covers the observation update.
Do not treat earlier route/source checks as native acceptance of all41 screens.

## Current expiration search correction

Android's parameter probe showed Home retaining its route but clearing Kitchen to
an empty query on tab return, before Filters opened. This displayed unfiltered
Camping items; it was not copying Browse's query. Expiration now reuses the shared
native-search interaction guard, retaining its own debounce and pending draft.
The final Android build without the probe passes Home/Browse/Apply with each
query retained. Two regressions reproduced inactive clearing and draft
reseeding, then passed; all2,002 mobile checks, TypeScript and review pass.

Settings draft retention and detail/Move return passed on Android. iOS command
width35954709640 passed9/9 iPad and8/9 phone. Phone Add stopped at keyboard readiness
(line744); its capture shows the keyboard and focused Name field. Keep the workflow
unresolved pending the combined run. Corrected phone/iPad command-width captures were reviewed: retry labels and
contextual commands are readable and bounded (see evidence/native-command-width-review.txt).
Combined13-workflow run35958732480 at35a79fe5 is pending; persistent-tab final-state
coverage and the phone Add workflow remain release gates.

## Accepted corrections and remaining release gates

Released0.24.34 includes M274–M279 filter density/navigation, retained Browse
search, contextual detail actions and Sharing recovery. Settings save/readback
35948277688 passed on both devices, including failed-save retry, exact full-name
persistence, reopening, creation and archiving. Move35951755320 passed all four
workflows on both devices; reviewed selection/creation captures close the scoped
normal-text hierarchy/contrast hold. These checks do not close all Settings design
findings or the comprehensive audit.

The next batch corrects compressed Retry labels by giving the native host the
available width while keeping its button leading-aligned. Reviewed phone/iPad
captures establish readable bounded commands in the named detail states; see
[evidence](evidence/native-command-width-review.txt). Do not repeat provider or
key-delivery experiments. The remaining phone Add keyboard-readiness observation
and combined navigation checks are explicit release gates, not reasons to reopen
accepted button geometry. The focused tab run strengthens final Settings return.
Android light/dark expiration return and Edit/Move cancellation passed; these do
not establish successful Edit/Move execution or app-wide appearance acceptance.

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
