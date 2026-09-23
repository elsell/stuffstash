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
creation-inset and Add destination readiness runs remain release gates.

The filter follow-up run35911930846 passed all six iPad cases and five of six phone
cases, including overview/menu, tag footer, search and Browse/detail return. Phone
Expiration entry renders its list at y750 with height62 in an874-point screen;
results exist but are offscreen. The candidate correction removes the redundant
flex wrapper so FlatList is the native screen's direct scrolling body, preserving
automatic insets and background. The connected result-tap-and-return test stays
unchanged. The precise UIKit transition cause and corrected native behavior remain
unproven; no extra swipe or timeout is used to hide the defect.

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
phone/iPad sharing verification is combined with the filter workflow run.

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


M274 candidate review: native filters run35897564291 passes on phone/iPad at
07f03d61. Matching screenshots confirm the overview menu removes the intermediate
page. Visual acceptance remains open: shorten repetitive menu labels and address
the filter sheet's overall density (M275), especially the iPad footer. Connected
Browse/detail/Back continuity is not established by fixture callback tests.
The iPad density candidate still hides Sort on entry despite tighter rows.
Filters now initially use the native large detent while retaining resizing and
bottom actions; [visual acceptance remains open](evidence/filter-initial-size-results.txt).
[Reviewed evidence](evidence/filter-menu-358975-results.txt).

### Connected filter return correction

M276 reproduced search loss after applying Availability on Android. The shared
native search adapter now distinguishes an active search interaction from header
lifecycle callbacks. Android connected Browse/Filters/detail/Back and Expiration
inheritance pass, as do native clear/type/submit and retained-query checks.
[Evidence](evidence/android-search-ownership-results.txt). iPhone native search/return checks pass in run35909057029. iPad Browse return
and Place search pass; Expiration stops at the known pre-scroll assertion, already
corrected in the combined run35911930846. [Terminal evidence](evidence/filter-connected-359052-results.txt).
Full native acceptance remains open; this follow-up does not broaden frozen M265–M273.


### Detail hierarchy implementation — native acceptance open

M277 now exposes permission-aware Edit in the native header, keeps Move with
location, and pairs availability status with its command. Standalone detail
consumers retain Edit; item/container/place checks cover duplicate removal and
read-only container status. Forty-two focused source tests, TypeScript and the
mobile structural check pass on paul. Critic found duplicate checked-out metadata;
that is corrected with a regression test. Android's first render prompted aligned,
bounded context rows. The revised Android item render has aligned context rows;
native header Edit opens the current asset and Cancel returns with commands intact.
[Android render](evidence/android-detail-contextual-actions.png). Representative Android photo, checked-out/read-only container and editable Place
checks pass; Edit/More/Search coexist on Place, and read-only status retains no
mutation controls. The [photo](evidence/android-detail-photo.png),
[container](evidence/android-detail-checked-out.png),
[read-only](evidence/android-detail-read-only.png) and
[Place](evidence/android-detail-place.png) renders keep M277 visually open:
photo actions still interrupt identity reading, and empty sections/actions push
actual contents too far down. The bundled image proves gallery layout only.
iPhone/iPad acceptance remains open. The corrected native reachability helper
checks Edit inside its header, not below the header with scroll content. This is a follow-up, not a new gate on the frozen selection batch.
