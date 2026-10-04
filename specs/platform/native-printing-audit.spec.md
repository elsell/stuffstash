# Native printing interaction acceptance

The named `printing-settings-and-labels` native audit mounts production printing
settings, defaults, printer detail, history and label screens in the real native
tab stacks with the persistent voice accessory. Label options must open from a
tab-hosted entry into the production root `assets/[assetId]/label` form sheet,
using `createAssetNativeSheetOptions(...).add` and `LabelTaskHeader`. A regular
tab-stack label screen is not valid evidence for sheet geometry or dismissal. Stateful in-memory repositories
provide printer readiness, revision-checked settings persistence and label render
failure followed by recovery. No hardware output or external printer call occurs.

Acceptance verifies the first and final actionable rows clear native navigation,
tabs and voice chrome; label first/last actions and Cancel clear the sheet header
and bottom bounds, and the rendered image fits the sheet width (including iPad); changing the automatic-print default and using the native
header Save persists the current draft when reopening; retrying a failed preview
renders again and produces a visible preview. Capture normal and enlarged text
layouts on the workflow's iPhone and iPad matrix. Source/fixture checks are separate
from the macOS runtime evidence and cannot establish a native pass themselves.

The visible label preview must declare the native image accessibility role and its
localized label. The pinned React Native runtime categorizes the rendered preview
as `Other` in XCTest even with the explicit image role. The fixture locates its
unique accessible label across element categories and verifies positive dimensions
and containment within the actual sheet. This proves rendered geometry, not
VoiceOver traits; do not claim assistive-technology acceptance from XCTest category
or source props alone. Preserve the explicit image role in production.

## Verified runtime evidence

[Run 37176721922](https://github.com/elsell/stuffstash/actions/runs/37176721922)
passed on iPhone 17 and iPad mini (A17 Pro) for source
`00ecca23c65b6589ef6098d9ba999193e783b5f0`. Normal and maximum-text captures
verify grouped export rows without overlap, preview containment, reachable final
Print and Cancel actions, settings navigation and native Save persistence.
The exported screenshots and hierarchies are retained as workflow artifacts.
This fixture does not invoke physical output or establish VoiceOver, camera
decoding, system share-sheet or physical printer acceptance. The existing voice
accessory's maximum-text clipping remains a separate shared limitation.

## Remaining print-task coverage (October 4 follow-through)

The earlier verified fixture covered settings and label export, not the separate
registered-printer form, quick-print fallback, reprint, print-job detail or uncertain
output resolution. Those surfaces require their own source and native acceptance;
the earlier pass must not be presented as evidence for them.

Registered printing uses a bounded task with an explicit native Cancel action.
Transfer the grouped choice hierarchy from Settings and the single task completion
from the reference layout reset: printer/readiness, layout/reference/copies, preview
and recovery, then one padded primary Print command. Flat choices retain native
pickers; reference and acknowledgement use inset native switch rows. Copies uses the pinned Expo SwiftUI Stepper on iOS, because the common task is
incrementing one label to two without a keyboard. Its range is positive safe
integers, not an invented server copy limit. Android retains an explicitly
labeled numeric TextInput with keyboard dismissal and validation. Native stepper
events must honor current locked state and teardown. Preview width comes from its measured containing
sheet, not the device window. Changing selection invalidates the old preview. A failed preview offers Retry that
re-renders the current draft; it must not refresh the catalog and reset choices.

Print-job detail has an inset status/progress heading and grouped contextual
Reprint, Cancel and Refresh actions. Unknown-output reporting is a separate grouped
form retaining acknowledgement, idle guard and immutable retry identity. Quick
print and reprint recovery reuse these controls and retain their current request
ownership and no-duplicate-submission guarantees. Cancel dismisses the UI and does
not cancel a possibly submitted physical job; only the explicit job Cancel action
does that. Native review must include these destinations on phone/tablet, enlarged
text, keyboard entry/dismissal, preview failure/retry and task cancellation. Source
checks alone do not establish those runtime results.

The route retains the scoped print draft while focus-bound authorization hides
its child. Restoring access retains printer, layout, reference and copies; changing
scope resets it. A retired printer or unavailable template blocks preview/submit
instead of silently replacing the user's destination with the new default. Pending
ambiguous submissions still use their immutable request selection.

### Print route presentation parity

The production root explicitly registers `assets/[assetId]/print` with the same
bounded `createAssetNativeSheetOptions(...).add` form-sheet presentation used by
label options. Quick-print fallback and manual print options therefore have known
sheet geometry and a native Cancel owner rather than inheriting a root card.
The complete-print native fixture must use that exact presentation, including
phone/iPad bounds. Print-job inspection and its reprint mode remain a stack
screen; reprint Cancel returns to its job without changing the job's outcome.
