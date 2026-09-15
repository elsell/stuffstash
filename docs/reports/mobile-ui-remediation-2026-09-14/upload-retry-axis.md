# Asset photo upload recovery

S101 source review at `c8b460c2`; mounted acceptance added in this pass.
Sources: AssetDetailRouteScreen, AssetDetailView, AssetPhotoUploadProgressPresentation,
AddAssetPhotosCommand, NativeCommandButton, PhotoSelectionQuery and
ExpoPhotoSelectionProvider. This covers existing-asset uploads, not Add draft
photos, stored-photo removal or remote image download retry.

| Axis | Evidence and remaining acceptance |
| --- | --- |
| Task | Retry is a command on failed selections, in the current asset workspace. No new screen or confirmation is needed for the same attempted upload. |
| Navigation | Failure and retry retain the asset screen. Source selection uses the existing system chooser; retry does not reopen it. |
| Selection | Initial library/camera selection is separate from retry. Failed photo objects retain their URI and original upload data; successful selections are excluded from retry. |
| Modality | Retry itself presents no modal. Permission/source failures use shared feedback. Picker presentation, dismissal and return remain native acceptance work. |
| Layout | Status and progress live in the parent scroll content, with no separate absolute footer. Filenames flex beside a status label; long names and bottom accessory clearance require native captures. |
| Adaptation | No upload-owned viewport dimensions. iPhone/iPad and compact-window reachability are unverified. |
| Typography | Filenames and status text wrap without a line cap. Normal long-name layout precedes enlarged-text review. |
| Appearance | Shared palette supplies text, border, failure and accent colors. Failed status is also textual. Actual contrast and dark/light rendering remain open. |
| Localization | English count/status labels and original filenames are displayed. Counts distinguish singular/plural; localized grammar and RTL row order are not established. |
| Imagery | This status region uses filenames rather than thumbnails; attached images are owned by the gallery. No inferred image-preview requirement. |
| Targets | M131 replaces custom Retry with NativeCommandButton. Its minimum-height declaration does not prove the native hit target or scroll reachability. |
| Gestures | Explicit Add photos and Retry commands exist. No gesture-only upload recovery. System picker gestures belong to separate media acceptance. |
| Keyboard | No text entry in this region. A keyboard left from a previous editor, hardware focus and native return are not verified here. |
| Accessibility | Retry has a native command label. Per-photo states are plain text; automatic announcement and reading order remain unverified. Workspace status has a polite live region, whereas the photo summary does not declare one. Do not infer reliable upload announcements from the workspace status. |
| Motion | No region-owned animated progress. It reports discrete pending/uploading/attached/failed states. Shared native control feedback still requires Reduce Motion acceptance. |
| Content | Rows preserve selected order; updates match both index and filename. The library picker has no configured selection limit, and rows are unvirtualized. Large multi-selection performance remains unverified. |
| Search | N/A: upload recovery operates on already selected files, not a searchable resource collection. |
| Loading | A synchronous operation guard prevents duplicate concurrent attempts. Upload progress is distinct from workspace mutation status and query refetching. Success clears progress after photo reconciliation. |
| Recovery | Partial failure preserves only failed selections for Retry. Mounted test with the real command proves only that failed file is retried, successful file is not duplicated, picker is not reopened, and success clears Retry/progress. |
| Editing | Uploads attach immediately rather than editing an asset draft. Picker cancellation adds nothing. Failed selections are in-memory and are cleared on asset change; no offline queue or durable restoration is claimed. |
| Privacy | The scoped repository remains the upload boundary. Camera acquisition requests permission; library uses the system picker without broad-library access. This source audit does not replace authorization or physical-device permission checks. |
| Notifications | N/A: this upload status does not schedule or handle system notifications. |
| Media | Retries use original selections without reacquisition. Transport, expired local URIs, limited library access and real camera interruption require device evidence. |
| Lifecycle | Existing asset-operation ownership suppresses completion after asset change/teardown. Mounted tests cover obsolete selection/results and concurrent asset uploads. Blur/background and process death remain open; teardown safety is not full lifecycle acceptance. |

## Evidence

The new mounted journey and existing related cases pass remotely on paul:
40 tests across2 files, TypeScript and the mobile structural check. No production
code changed in this pass. M131 remains pending native acceptance; no new rendered
failure is inferred from source. Screen-reader announcement behavior is explicitly
unverified and must be exercised rather than counted as an accessibility pass.

Apple's progress-indicators page was requested on2026-09-15 but returned only a
JavaScript-required shell. No additional claim about its contents is made here:
https://developer.apple.com/design/human-interface-guidelines/progress-indicators
