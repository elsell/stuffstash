# Add tag selection: source review

S089 initially at6982be03, revisited at327391bf plus M201. All24 axes below were considered against AddAssetScreen,
AddAssetDraftStore and the existing asset-tag/platform interaction specs. This
pass is source evidence, not native acceptance. The task is selecting or creating
tags inside an unsaved asset; Add's context-scoped draft storage differs from Edit's
explicit discard flow, so its recovery contract must be preserved independently.

| Axis | Observation and required follow-up |
| --- | --- |
| Task | Tags are secondary metadata behind More details. Selection and creation stay in the form. Review the separate search and creation fields with realistic inventories before merging their meanings. |
| Navigation | More details is in-place disclosure, not a new stack. M96 moved unfinished entry above the conditionally mounted picker, preserving it on collapse. |
| Selection | Existing choices expose selected state and a checkmark. Pending definitions are removal commands, not persisted-tag selection toggles; M201 names that action explicitly. Verify actual assistive output. |
| Modality | iOS color uses the native picker; Android uses the documented fallback. Nested presentation and cancellation still need runtime checks. |
| Layout | Tag controls live in the Add scroll inside details. Save is a native header action. Large-text reachability and label bounds remain pending. |
| Adaptation | Choices wrap as a group, names truncate to one line. Narrow phone, iPad and enlarged text remain unverified. |
| Typography | Explicit styles scale but truncate long names; similarly prefixed names may become indistinguishable. This is an unverified usability risk, not a measured overflow. |
| Appearance | Shared palette and translucent tag color treatment are used. Contrast and selected-state legibility still require measurement. |
| Localization | Shared tag-choice presentation trims and lowercases search and naturally orders the initial choices. Expanded strings, locale ordering and RTL remain native checks. |
| Imagery | Checkmarks identify selected existing tags; an X removes pending definitions. Color choice has named accessible swatches. Verify actual labels and distinguishability. |
| Targets | Existing M40 records48-point minima and native Add tag. Actual hit regions and overlapping/scrolled states remain pending. |
| Gestures | Explicit disclosure, selection, removal and Add commands exist. Scroll versus keyboard gestures and native color dismissal remain pending. |
| Keyboard | Search and new-name fields use shared inputs. Focus, keyboard return and retained typing after More details need native verification. |
| Accessibility | Existing selected state is declared. M201 gives pending-definition commands the name Remove new tag {name} and disabled state. Full reading order, actual spoken output and visual truncation remain native checks. |
| Motion | No tag-specific animation. Shared keyboard and sheet behavior still require Reduce Motion verification. |
| Content | M94 aligns Add with the initial-twelve disclosure and selected extras contract. Native Show all/fewer handles larger collections. |
| Search | Local case-insensitive substring matching exposes matches alongside selected choices; search trims whitespace. No matching tags is explicit. Search itself does not create or stage a tag. |
| Loading | Tags arrive with Add context. No independent tag-loading state inside the picker. Context error/retry and existing draft retention need native checks. |
| Recovery | M96 retains unfinished name/color in the scoped draft; M93 explains overlong-name rejection without clearing the input. Save guides users to stage or clear unfinished entry, including when details are collapsed. |
| Editing | Existing IDs, staged definitions and unfinished entry are stored with the route draft. M201 removal preserves other staged tags and unfinished entry. Clear draft and successful Save have separate reset behavior. |
| Privacy | Draft storage keys include service scope, principal, tenant and inventory. This source pass is not adversarial authentication or storage-boundary verification. |
| Notifications | No direct notification controls. Interruption and resuming the Add draft still apply. |
| Media | The tag picker has no attachment acquisition. Other More details/Add media controls are separate surfaces. |
| Lifecycle | Context-scoped in-memory persistence includes staged tags and unfinished entry after M96. It is not process-death storage. Native background/return and scope changes remain acceptance work. |

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

Native M96 scenario added against the existing configured-header Add fixture:
exact asset-name and tag-name entry after keyboard readiness, Save enabled then
blocked, collapse/reopen retention with visible collapsed guidance, staging, clear
and close. It performs no Save; context-scoped remount remains proven only by the
route regression. This scenario uses standard text size; large-text Add/tag checks
remain outstanding. Structural checks and critic review passed; native execution
and screenshot inspection are pending.

M94 discovery follow-up: Add now uses the same initial twelve naturally ordered
choices and selected extras as Edit, with native disclosure. Trimmed search
refines matching choices without removing selections; unmatched queries have
explicit feedback. The earlier search-only discovery observation is superseded.
Twenty-six remote Add/Edit tests and type/structural checks passed; native
interaction acceptance remains pending.

M201 revisits the remaining inferred removal label. The staged chip now explicitly
names removal of a new tag, preserving the distinction from deleting a saved tag.
The existing draft workflow case failed first because this named command was absent;
it now verifies removal of one of two staged tags while keeping the other and an
unfinished name, then saves the remaining tag. Native spoken output is not inferred
from a React label assertion.

Current follow-through:39 tests across5 Add/color files, TypeScript and structural
checks pass remotely on paul (`/tmp/add-tag-remove-green.log`). Critic found no blocker.
