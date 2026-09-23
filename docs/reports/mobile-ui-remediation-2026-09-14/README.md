# Mobile UI audit — current state

The comprehensive audit is **incomplete**. Prioritize stable screen structure,
connected everyday tasks, visual coherence, then detailed states. Normal text comes
first. The surface/axis inventory checks omissions; it is not a separate test queue.

## Delivery and frozen batch

Latest verified TestFlight: **0.24.32 (123.1)** — M260–M264, with Apple processing
and exact [changelog readback](evidence/workflow-release-358899-results.txt).
[PR171](https://github.com/elsell/stuffstash/pull/171) merged as690ee8e4;
release35889902851 succeeded. Final Browse run35887017924 passed on phone/iPad,
and required CI passed on ddcb700c before merge.

The next batch is frozen to **M265–M273**, [draft PR173](https://github.com/elsell/stuffstash/pull/173):
Map/detail hierarchy, optional Add tag creation, Add/Edit tag selection, Add
destination selection, independent Move creation drafts, retained native text,
selection presentation above modal editors, and consistent Move Here controls.
See the [batch contract](../../../specs/platform/mobile-selection-batch.spec.md).
Unrelated findings do not block either release.

## Current diagnosis and next decisions

Browse List/Map structure shipped in0.24.32 after native verification. The frozen
M265–M273 selection batch is tracked independently in PR173; its current Move
grouped-creation and Add destination entry runs remain release gates. The preceding Move run passed all four workflows on both devices.

The filter follow-up run35911930846 passed all six iPad cases and five of six phone
cases, including overview/menu, tag footer, search and Browse/detail return. Phone
Expiration entry renders its list at y750 with height62 in an874-point screen;
results exist but are offscreen. The candidate correction removes the redundant
flex wrapper so FlatList is the native screen's direct scrolling body, preserving
automatic insets and background. The connected result-tap-and-return test stays
unchanged. The precise UIKit transition cause and corrected native behavior remain
unproven; no extra swipe or timeout is used to hide the defect. Search ownership preserves queries across filter application; the connected Android journey and scoped native search checks are recorded in [the evidence](evidence/android-search-ownership-results.txt).

## Detail action hierarchy follow-up

M277 now groups Move with Location and availability with its command, keeps Edit
in the native header, pairs photo status with Add photos, and groups contents
commands directly before the list. Empty sibling sections no longer push actual
contents below the fold; empty places and no-result searches retain one useful
recovery state. Android representative captures show the revised
[photo row](evidence/android-detail-photo.png) and
[place contents](evidence/android-detail-place.png). Populated-photo, checked-out,
read-only and place header checks pass on the rebuilt Android fixture. The bundled
image establishes gallery layout, not real-photo crop quality.

Phone/iPad verification of the final composition remains open. This is outside
the frozen selection batch and is not a prerequisite for its release. Source
checks and Android review do not establish iOS visual acceptance.

Sharing review also corrected inline feedback alignment and grouped Share/Copy
completion actions (M278). Android controlled recovery and visual review pass;
phone/iPad sharing verification is combined with the filter and detail workflows.

Run35921268590 passes nine of twelve cases on both devices, including all six
filter/search/return workflows, Sharing recovery, Map context and empty-photo
hierarchy. This establishes the corrected phone Expiration entry/return workflow.
Three detail assertions stop on observation mismatches: old empty-state copy,
a36-point system toolbar frame treated as a custom body button, and duplicate
nested AX text nodes at the same Availability bounds. Matching screenshots and
accessibility trees show the expected empty state and one visual heading. Correct
those observations, require valid heading geometry before deduplication, and retain
all recovery, bounds and permission checks. Only the five detail/hierarchy cases
need another native run; do not repeat the nine unchanged passed workflows.

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
