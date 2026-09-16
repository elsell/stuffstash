# Asset Edit route — 24-axis source review

R009 atdc901a08 plus M201/M209 follow-up. Reviewed the edit route/service wiring,
ActionAsset, EditAssetForm, useAssetSheetOperation, EditAssetSheet, draft helpers,
native sheet options and native action adapter. Tag-specific evidence lives in
edit-tags-axis.md. This is source coverage, not a native pass.

| Axes | Evidence and remaining acceptance |
| --- | --- |
| Task, navigation, modality | A multi-field draft justifies an editor sheet. Cancel protects dirty and unstaged input with a scoped discard alert; Save returns to Details. Swiping is disabled. M209 adds Close during initial load/error, using Back or Home for direct entry. |
| Selection, editing | Kind and existing custom type are fixed; eligible items can assign a type. Name/description, expiration and tag draft are independent of refreshed core data. Save trims fields and checks valid expiration/unstaged tags. Successful partial tag creation is reconciled after failure. |
| Layout, adaptation, keyboard | Safe-area frame encloses a keyboard-avoiding container, scrollable form and native footer. Automatic keyboard insets also apply to the ScrollView. Native overlap/double-adjustment, detent sizes and long content remain runtime checks; the title is within the form. |
| Typography, appearance | Semantic colors, wrapping form labels and native commands. Long chip text truncates to one line; actual contrast and large-text geometry remain unverified. Normal-size acceptance comes first. |
| Localization, imagery | English copy; user name/description/tags may be long or RTL. No photo editing in this sheet; tag color/checkmarks supplement text. |
| Targets, gestures, accessibility | Named text fields and native Save/Cancel, retry and disclosure commands. Tag choices expose selection; staged removal now names its distinct action (M201). Explicit Cancel avoids relying on gestures. Device focus order and actual target geometry remain open. |
| Motion, content, search | No editor-specific animation; system sheet/keyboard transitions need Reduce Motion review. Tag disclosure retains selected extras. No whole-editor search task; tag discovery remains limited as documented in edit-tags-axis.md. |
| Loading, recovery | Core gates the form; types/tags load independently and expose retries within the form. M209 prevents a loading/error dead end. Save failure keeps the draft and presents a native error. Existing controlled-text and color runtime failures are not resolved by these checks. |
| Privacy, lifecycle | Core query scopes by inventory and suppresses access-error data. M210 now makes readable view-only/archived assets read-only, preserves drafts and Cancel, and rejects retained callbacks. Busy mutations guard removal and late completion; scoped discard callbacks reject obsolete drafts. Process death is not durable draft recovery. |
| Notifications, media | No notification entry, camera, upload or audio in Edit. Interruption/resume belongs to lifecycle acceptance, not a separate media feature. |

M209 applies to the shared Edit/Move/Move-here loading shell. The ready forms keep
their existing operation/dismissal behavior. Mounted tests establish navigation
commands, not actual native sheet placement. M210 has13 adversarial route cases
across all three forms: initial denial, revocation/archive, retained callbacks,
recovery and Edit eligibility changing during tag reconciliation. It leaves
server authorization unchanged. Native permission-transition acceptance is pending.

M201/M209 validation:64 related mounted tests, TypeScript and mobile structural
checks pass on paul (`/tmp/action-exit-reviewed.log`), including normal Back and
direct-entry Home fallback. Critic found no blocker. Native command reachability
and announcement remain unverified.
