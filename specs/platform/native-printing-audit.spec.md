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

The visible label preview must expose the native image accessibility role and its
localized label. Run 37172865963 rendered the recovered PNG successfully on iPad,
but UIKit exposed it as `Other`, so image-role lookup failed. Preserve the native
image-role acceptance instead of weakening it to accept an untyped element.
