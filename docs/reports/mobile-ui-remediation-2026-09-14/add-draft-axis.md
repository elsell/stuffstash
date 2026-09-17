# Add route, main draft and recovery — all 24 axes

R007, S084 and S094, reviewed at eac9c3ac with M197 candidate. Sources:
app/add, AddAssetScreen, AddAssetNameField, AddAssetDraftStore and mounted Add
tests. Parent/tag/type/photo editors retain their separate surface records.

| Axis | Current evidence and remaining acceptance |
| --- | --- |
| Task | Primary capture uses photos/name/parent; optional description/tags live under More details. Save creates an asset; Close preserves the scoped draft. Clear explicitly discards it. |
| Navigation | Native Close/Save header actions use the stable committed-handler adapter. Add Here supplies an initial parent. Save resets the form for another item and offers View; actual header and return behavior remain native checks. |
| Selection | Parent selection is searchable in-place disclosure; tags and type use their dedicated controls. Flat creation/choice commands do not require another generic stack screen. |
| Modality | Add is a native task route; photo source/viewing and color may present system UI. Busy draft operations block route removal. Native dismissal and nested picker return remain unverified. |
| Layout | Main ScrollView uses automatic content/keyboard insets and bottom safe-area padding. Error layout scrolls to the measured header edge. Actual native keyboard/header/error placement remains open. |
| Adaptation | Shared controls and wrapping groups handle available width in source. Phone/iPad normal-size evidence is still required; native typing failures remain unresolved. |
| Typography | Name is a single-line entry; description is multiline. Long user names and labels need runtime verification. Native-owned iOS name entry differs from the controlled fallback and is not certified by React prop checks. |
| Appearance | Shared palette and native Close/Save commands; M197 replaces custom Clear with native destructive semantics. Light/dark and disabled appearance remain pending. |
| Localization | Copy is English; user content is unrestricted. Locale-aware tag matching exists, but RTL and expanded strings are not verified. |
| Imagery | Staged photos support preview/removal/reorder with explicit controls. Empty capture remains available. Permissions and real camera/library behavior require physical checks. |
| Targets | Native header and Clear commands provide platform interaction. Parent/disclosure/tag rows have source minima, not measured bounds. Color target and text-entry findings remain open. |
| Gestures | Drag reorder has explicit move alternatives; More details and parent disclosure are buttons. Idle dismissal preserves draft; busy removal is blocked. Native drag/scroll competition remains pending. |
| Keyboard | Editing uses shared inputs/dismiss control and automatic inset handling. Starting an operation dismisses the keyboard. Existing native input loss is not fixed or explained by M197. |
| Accessibility | Inputs have names; disclosure exposes expanded state; errors use an assertive region plus iOS announcement. Full reading order, possible duplicate announcements and focus after error/reset need runtime checks. |
| Motion | Keyboard and system presentation remain native; parent focus schedules a scroll. Reduce Motion and interrupted scrolling remain unverified. |
| Content | Ready context identifies inventory/tenant; denied creation is explicit. No asset list pagination in the form. Optional metadata has separate loading/error handling. |
| Search | Parent suggestions debounce through the shared scoped hook and require known results before creation. Tag search has separate tests/report. Main draft itself is not a global search surface. |
| Loading | Context and principal must resolve before restoration. Type loading is independent. Saving/parent creation/photo selection lock edits and duplicate dispatch. |
| Recovery | Inline command errors preserve draft, scroll into view and allow correction/retry. Context/type retries do not overwrite dirty fields. Partial tag creation reconciles definitions after failed asset creation. Native error visibility remains pending. |
| Editing | Context-scoped in-memory draft includes title, description, parent, staged photos/tags, unfinished tag/color, type and expiration. Clear empties it; successful Save resets fields while retaining parent. Process-death persistence is not provided. |
| Privacy | Store keys include service/principal/tenant/inventory; screen is resource-keyed. Permission controls hide the form when Add is unavailable. Mounted scope checks are not backend authorization evidence. |
| Notifications | No notification-specific controls here. External interruption while editing remains an app-level lifecycle scenario. |
| Media | Selection cancellation retains draft; source chooser rejects departed callbacks. Draft photos persist only within the current in-memory service context. Native permission and picker focus remain pending. |
| Lifecycle | Ordinary unmount/remount restores the scoped draft. Background/process death and signed-out reconstruction are different cases; do not imply durable storage. Native route-return and typing failures remain in the runtime backlog. |

M197 preserves the existing intentional-clear semantics; it changes the command
adapter and accessible label, not the destructive operation or Close behavior.
The restoration test edits a stored name and unfinished tag/color, clears through
the named command, verifies emptied store values, then creates a fresh draft.
Runtime appearance and reset/focus still require native evidence.

All23 Add checks, TypeScript and mobile structural validation pass remotely on
paul (`/tmp/add-clear-command-reviewed.log`). Critic found no blocker. The initial
follow-up assertion incorrectly read a controlled value prop from iOS's native-owned
name field; it now verifies actual persisted draft state, without claiming native
text rendering. Native input loss remains open.
