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
| Add/Edit selection | Run35880132749 showed an empty Add name after Tags and Edit Tags behind its modal owner. M271 keeps the current native reappearance seed; M272 presents selection above its owner. | Grouped [35890124446](https://github.com/elsell/stuffstash/actions/runs/35890124446): name retention, visible/tappable selection, Cancel/Done, destination creation/retry and unfinished tag disclosure. |
| Move | M270 separates search from creation name and preserves Kind/Create during editing. Android connected creation/move retry passed. | Corrected grouped Move/Move Here native run below; retain exact query/name and selected state assertions. |
| Move Here | M273 replaces the old partial sheet/custom preview with native header/search and checked rows. [Android workflow passed](evidence/move-here-m273-results.txt); retained screenshot inspected. | [35891073312](https://github.com/elsell/stuffstash/actions/runs/35891073312): normal-text search, refinement/clear, lookup recovery, rejected move/retry and return on phone/iPad. |

The integrated source run passed1,968/1,970 tests. The two failures enforced removed
inline Add structure and a hidden Move Here header. Corrected expectations and all
54 other tests in those files pass. TypeScript, structural checks, fixture preparation
and critic review pass. All six CI35891523435 jobs passed at336e17ef. These checks
do not establish native presentation or physical-device behavior.

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
