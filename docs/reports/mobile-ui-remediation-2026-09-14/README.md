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

| Workflow | Established evidence | Remaining acceptance |
| --- | --- | --- |
| Browse List/Map | Stable header/switcher, actual scrolling and outer-card tablet geometry passed in35887017924. | Delivered in0.24.32; no unchanged layout rerun. |
| Map/detail hierarchy | M265/M266 scoped phone/iPad acceptance passed in35874901875; Android captures retained. | Preserve in the integrated batch; no claim of whole-detail acceptance. |
| Add/Edit selection | Run35880132749 showed an empty Add name after Tags and Edit Tags behind its modal owner. M271 keeps the current native reappearance seed; M272 presents selection above its owner. | 35890124446 passed Add/Edit Tags on both devices and Add destination on iPad. Phone destination needs native search closed before task Cancel. Unfinished tag disclosure still fails keyboard entry on phone and times out during scrolling on iPad after entering Camping; [retained evidence](evidence/selection-lifecycle-358901-results.txt) keeps those gates open. |
| Move | M270 separates search from creation name and preserves Kind/Create during editing. Android connected creation/move retry passed. | Corrected grouped Move/Move Here native run below; retain exact query/name and selected state assertions. |
| Move Here | M273 replaces the old partial sheet/custom preview with native header/search and checked rows. [Android workflow passed](evidence/move-here-m273-results.txt); retained screenshot inspected. | 35891073312 passed iPad rejected move/retry. Other cases stopped on native radio-role queries or task commands during active search. Corrected harness run35895924050 at e15d3133 is active; no product search changes. |

The integrated source run passed1,968/1,970 tests. The two failures enforced removed
inline Add structure and a hidden Move Here header. Corrected expectations and all
54 other tests in those files pass. TypeScript, structural checks, fixture preparation
and critic review pass. All six CI35891523435 jobs passed at336e17ef. These checks
do not establish native presentation or physical-device behavior.

## Next connected workflow

Source review at b5eee60f of Browse → Filters → results/detail/Back retains in-place
short choices and searchable tag selection. M274 recommends replacing the
three-destination Expiration intermediate page with a native action menu; preserve
inherited filters and original Browse return context. This is outside frozen
M265–M273. Current connected native return/scroll evidence is still required;
source state ownership does not certify the visual journey.

## Detail action hierarchy follow-up

Normal-text phone/iPad detail captures in35907123046 expose a scattered command
hierarchy: prominent full-width Check out, separate Edit/Move, and isolated Add
photos. M277 records the evidence and next design decision in [findings](findings.md).
This is outside the frozen selection batch; preserve task access while reviewing
the complete detail composition rather than merely changing button padding.

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
