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
source35a79fe5 is the integrated baseline; the current candidate also replaces
the confirmed failing customization Name control with the existing DraftTextField.

Focused run35960677440 completed: iPhone3/3 passed; iPad2/3 passed.
The iPad Settings draft remains exactly `Tools emergency` and both tabs are visibly
present, but the revised hittable-match query still fails after return. Duplicate
first-match selection alone is therefore insufficient to explain the failure.
Keep the release gate open. The final native tree places the tab strip at y32–76
and the Tag navigation bar at y32–140; this overlap is an observation, not proof
of touch interception. Next verification must distinguish actual tap interception
from an accessibility hit-test discrepancy by exercising the visible tab and
requiring the destination change and exact draft return. Do not remove the
reachability requirement or repeat the same selector-only adjustment.
See [focused evidence](evidence/persistent-tabs-35960677440.txt).
Required CI35961721094 passed at3fc47eaf. Completion run35961716802 passes Retry
on both devices and Settings readback on iPad; phone New Tag still fails exact name
entry (Cngampi rather than Camping). Do not treat
route/source checks as native acceptance of all41 screens.

## Current expiration search correction

Android's parameter probe showed Home retaining its route but clearing Kitchen to
an empty query on tab return, before Filters opened. This displayed unfiltered
Camping items; it was not copying Browse's query. Expiration now reuses the shared
native-search interaction guard, retaining its own debounce and pending draft.
The final Android build without the probe passes Home/Browse/Apply with each
query retained. Two regressions reproduced inactive clearing and draft
reseeding, then passed; all2,002 mobile checks, TypeScript and review pass.

Settings draft retention and detail/Move return passed on Android. iOS command
width35954709640 passed9/9 iPad and8/9 phone. Phone Add initially stopped at keyboard readiness
(line744); the subsequent combined run passed Add on both devices. Corrected phone/iPad command-width captures were reviewed: retry labels and
contextual commands are readable and bounded (see evidence/native-command-width-review.txt).
Combined13-workflow run35958732480 at35a79fe5 passed12/13 phone and11/13 iPad.
Add and expiration retention passed both. Follow-up35961716802 closes Retry recovery.
Its phone capture confirms lasting wrong character order after15s; this is not a
transient observation issue. The Name-only native draft correction preserves current
validation, key derivation, resource loading, locks and rejected-save drafts. Existing
2,002 source tests, TypeScript, structural checks and review pass. Exact native
save/reopen and tab-return remain gates; no more typing delays/provider experiments.

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
key-delivery experiments. Add now passes on both devices. Remaining native gates are iPad Settings tab
reachability and exact Settings create/readback after the Name control correction. These do not
reopen accepted button geometry.
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
