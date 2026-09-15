# Add tag selection: source review

S089 at6982be03. All24 axes below were considered against AddAssetScreen,
AddAssetDraftStore and the existing asset-tag/platform interaction specs. This
pass is source evidence, not native acceptance. The task is selecting or creating
tags inside an unsaved asset; Add's context-scoped draft storage differs from Edit's
explicit discard flow, so its recovery contract must be preserved independently.

| Axis | Observation and required follow-up |
| --- | --- |
| Task | Tags are secondary metadata behind More details. Selection and creation stay in the form. Review the separate search and creation fields with realistic inventories before merging their meanings. |
| Navigation | More details is in-place disclosure, not a new stack. It conditionally unmounts the picker: M96 loses unfinished entry. |
| Selection | Existing choices expose selected state and a checkmark. Pending definitions show a remove icon but no selected-state declaration. Verify assistive output and intended action semantics. |
| Modality | iOS color uses the native picker; Android uses the documented fallback. Nested presentation and cancellation still need runtime checks. |
| Layout | Tag controls live in the Add scroll inside details. Save is a native header action. Large-text reachability and label bounds remain pending. |
| Adaptation | Choices wrap as a group, names truncate to one line. Narrow phone, iPad and enlarged text remain unverified. |
| Typography | Explicit styles scale but truncate long names; similarly prefixed names may become indistinguishable. This is an unverified usability risk, not a measured overflow. |
| Appearance | Shared palette and translucent tag color treatment are used. Contrast and selected-state legibility still require measurement. |
| Localization | Search lowercases using the locale but does not trim its input; results preserve input order. Expanded strings, locale ordering and RTL are pending. |
| Imagery | Checkmarks identify selected existing tags; an X removes pending definitions. Color choice has named accessible swatches. Verify actual labels and distinguishability. |
| Targets | Existing M40 records48-point minima and native Add tag. Actual hit regions and overlapping/scrolled states remain pending. |
| Gestures | Explicit disclosure, selection, removal and Add commands exist. Scroll versus keyboard gestures and native color dismissal remain pending. |
| Keyboard | Search and new-name fields use shared inputs. Focus, keyboard return and retained typing after More details need native verification. |
| Accessibility | Existing selected state is declared. Pending-definition removal has an inferred name and no explicit remove label; full reading order and control purpose need acceptance review. |
| Motion | No tag-specific animation. Shared keyboard and sheet behavior still require Reduce Motion verification. |
| Content | With no query, only selected existing tags appear. This differs from Edit and the general initial-twelve disclosure contract; inspect the intended Add discovery pattern before choosing a correction. |
| Search | Local case-insensitive substring matching exposes matches alongside selected choices. There is no result count or explicit no-match message. Whitespace and broad matches need review. |
| Loading | Tags arrive with Add context. No independent tag-loading state inside the picker. Context error/retry and existing draft retention need native checks. |
| Recovery | M96: unfinished name/color are local and absent from stored AddAssetDraft. Overlong names disable Add without feedback; Edit's M93 correction is not shared here. |
| Editing | Existing IDs and staged definitions are route state and persisted. Unstaged entry is omitted from Save and is lost on details collapse. Correct this without changing Add's context-scoped resume behavior. |
| Privacy | Draft storage keys include service scope, principal, tenant and inventory. This source pass is not adversarial authentication or storage-boundary verification. |
| Notifications | No direct notification controls. Interruption and resuming the Add draft still apply. |
| Media | The tag picker has no attachment acquisition. Other More details/Add media controls are separate surfaces. |
| Lifecycle | Context-scoped draft persistence includes staged tags but excludes unfinished entry. Background/return and scope changes require acceptance after M96 repair. |

M96 is confirmed by local state in AssetTagPicker, conditional rendering under
showDetails, and the route/store fields. No device reproduction is claimed. Next:
route-own and persist unfinished entry; guard Save with useful feedback; test
collapse/reopen, close/resume, clear and staging atomically. Then review Add's
selection discovery against the general tag contract and exercise native behavior.

M96 now has a correction candidate with route-owned, scoped entry persistence,
collapse/reopen guidance and guarded Save. Nineteen remote Add tests plus type and
structural checks passed. This supersedes the local-entry observations above;
native acceptance and the separate discovery/overlong-name review remain open.

M93 now also has an Add correction candidate: resolver-driven overlong-name
feedback retains the entered value and clears after correction. Eighteen remote
Add/resolver tests plus type/structural checks passed; native feedback is pending.
