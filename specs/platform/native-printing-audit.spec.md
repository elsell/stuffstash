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
