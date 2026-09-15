# Custom-field editors — 24 source axes

R031/R033/R039/R041, source0065b1f7 plus M167/M168. Reviewed production route
composition, CustomizationEditorScreen/Fields, editor draft and command mapping,
ManageCustomFields, settings choice rows, and mounted customization scenarios.
Shared editor findings M58/M59/M165/M166 remain applicable. This is source review,
not native acceptance. Platform rationale follows the platform-interaction review
spec: local single choices use pickers; immutable values are static; navigation
has a distinct destination; submission and lifecycle operations are commands.

| Axis | Evidence and remaining acceptance |
| --- | --- |
| Task | Create or edit a field definition. Name, value type and applicability are presented together; technical key is disclosed separately. |
| Navigation | Household/inventory wrappers pass scope explicitly. Native Back returns to the scoped collection through replacement; inherited management changes to the household editor. Native stack return still needs acceptance. |
| Selection | Creation has native Type and Applies to pickers. Existing type/options/targets are static where immutable. Additional targets use checkmarked multi-selection rows; long target sets and native semantics need runtime review. |
| Modality | Dirty exit and lifecycle actions use focused-resource system confirmations. M167 includes unsubmitted option text in dirty protection. |
| Layout | Shared inset column, grouped sections, keyboard-adjusting scroll, and native Save. Long option/target lists and bottom controls need native geometry evidence. |
| Adaptation | Single-column form; no assumed tablet layout certification. Normal text first, enlarged text still pending. |
| Typography | Labels, selected values and validation use shared text styles. Long field/option names and wrapping require native captures. |
| Appearance | Semantic colors and native pickers/commands; custom multi-select rows remain visible review targets. Dark/light rendering not accepted from source. |
| Localization | English labels; generated key normalization and validation can require manual key correction. Non-Latin names and RTL input/layout remain runtime work. |
| Imagery | No content images; checkmarks and disclosure chevrons reinforce selected/expanded state. |
| Targets | Shared controls define minimum sizes. Actual native picker, Back, option commands and target rows need touch/assistive evidence. |
| Gestures | Explicit commands exist; dirty state disables navigation gestures and removal is intercepted. Native gesture dismissal still pending. |
| Keyboard | AppTextInput and shared keyboard policy. M167 protects unsubmitted option text. Shared single-line native text loss remains unresolved. |
| Accessibility | Named fields, static immutable values, checked multi-select states and inline live validation. Reading order, disclosure state and error announcements remain runtime work. |
| Motion | No custom timed animation in field controls. System navigation and reduced-motion behavior unverified. |
| Content | Existing enum options and targets explain immutability. M167 explains add-or-clear before Save; stable key is secondary detail. |
| Search | No editor search. Target list is inline with no search; assess real large collections before accepting discoverability. |
| Loading | Scoped context/definition/type reads, labeled progress and stale-resource rejection. Permission-refresh and partial-list behaviors have mounted coverage. |
| Recovery | Inline Retry and retained error draft; access changes keep read-only state and permit Refresh access. No source claim of native focus recovery. |
| Editing | M167 protects pending option input; M168 keeps dormant enum options when switching type but excludes them from non-enum creation. Existing options cannot be renamed/removed and applicability cannot narrow. |
| Privacy | Scoped policy gates and permission-denial handling; client policy is not server authorization evidence. Inherited definitions require household management. |
| Notifications | No reminder delivery or permission action in field definition editing. |
| Media | No media capture/upload or permission action. |
| Lifecycle | Focus/resource ownership guards completion/discard. Archive/restore/delete preserve confirmation and locks; native exit/background/deep-link cases remain pending. |

Both new mounted regressions failed before their fixes. All77 customization tests,
TypeScript and mobile structural checks pass on paul; critic found no blockers.
The new native tag-editor journeys exercise shared Back/Save/lifecycle only and
do not establish field picker, target-list or enum-entry native acceptance.
