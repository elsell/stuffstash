# Inventory switcher review

Source064fbaf9; R056 and S080–S083. Reviewed tenant-switcher route, root sheet
configuration, TenantSwitcherSheetScreen, IdentityLabel and SelectInventoryCommand.
This is a source review plus the focused remote regression evidence, not native
visual or assistive-technology acceptance.

## Task, navigation and selection

Switching inventory changes the context used by subsequent tasks. The current
inventory is marked selected; choosing another household first narrows its inventory
list without yet committing a different inventory. A bounded hierarchical chooser
is appropriate for this relationship. Do not flatten household/inventory names
into an unqualified menu: households can share a name and inventories have roles.
The existing M07 test verifies household identity rather than name matching.

The sheet supports medium/full detents, a scrolling body and native Close. The
household list returns to inventories when a household is chosen. Empty households
have an explicit message. Retry stays in the sheet after load failure; selection
failure retains the chooser and gives a safe retry message. This review does not
claim that every custom command or row already uses the best native adapter.
M80 replaces Switch household/Back and load Retry with NativeCommandButton.
The full-width command now sits below the wrapping heading; native acceptance
remains pending.

## Ownership and recovery

M78 covers focus departure during selection, both orderings of completion/refocus,
and a fresh selection afterward. Pending selection blocks a duplicate request;
Close aborts the initiating request and dismisses. Focus cancellation is not a
rollback promise. M79 ensures repository acceptance still notifies the scoped
selection observer even if the caller cancels before completion. Initially canceled
or rejected requests do not publish selection. Repository/API authorization is
unchanged and requires its separate boundary evidence.

The UI reads cached dashboard data through the inventory-scoped query adapter.
Initial loading and failure are distinct from cached content; selection has separate
busy/error state. M80 loading copy now says “inventories,” matching the current task. Background-refetch
failure with retained data has no explicit notice here; evaluate its effect on
stale membership before treating it as a proven user-facing failure.

## Layout and accessibility gates

Rows expose selected/disabled state and inventory names; labels and metadata are
separate from the role badge. The body scrolls and uses automatic inset adjustment.
These properties do not prove correct VoiceOver grouping or native target geometry.
M80 lets the household heading wrap. Long names and
large type need visual inspection, particularly beside Switch household and the
nonshrinking role badge. Existing full-width row targets are not proof that their
contents fit. Light/dark, contrast, RTL, narrow/iPad windows, reduced transparency,
VoiceOver/TalkBack and interruption by push navigation remain pending.

Critic-reviewed implementation evidence for M78/M79 is in findings.md. No native
switcher run has been inferred from the unrelated sheet-diagnostic scenarios.

## Complete axis follow-up at fc93c151 plus M184

This follow-up inspects the current route, sheet, query and selection command. It
extends the earlier source review, not the native acceptance evidence.

| Axis | Source decision and remaining evidence |
| --- | --- |
| Task | Current household narrows inventory choices; selecting inventory commits app context. Names and roles remain visible. |
| Navigation | Household selection stays inside the sheet; inventory success returns through native Back. Verify entry from Home and dismissal on device. |
| Selection | Hierarchical, potentially duplicate-named resources justify a chooser instead of a small flat menu. Current inventory has a checkmark/selected state. |
| Modality | Native bounded sheet with Close; Close cancels pending presentation work. Verify drag dismissal. |
| Layout | Safe-area wrapper excludes top, native header owns it; scroll body adjusts insets automatically. Native footer/last-row bounds remain unverified. |
| Adaptation | Household title wraps; rows use flexible labels and fixed checkmark/role regions. Long names, narrow windows and tablet geometry remain open. |
| Typography | Labels and secondary metadata use explicit font sizes; native wrapping/large text remain open. |
| Appearance | Shared semantic palette and native commands; selected checkmark is not color-only. Contrast and materials require captures. |
| Localization | English labels, plural inventories and domain date summaries; RTL and long localized strings remain open. |
| Imagery | Shared identity symbols plus text labels; no photos required. Missing icon behavior belongs to the shared adapter. |
| Targets | Inventory/household rows declare minimum height62; commands use native adapter. Actual targets are not proved by constants. |
| Gestures | Standard scroll and sheet dismissal; explicit Close and household Back alternatives. |
| Keyboard | N/A: no editable field in the chooser. |
| Accessibility | Inventory rows expose name, selected, disabled and busy; household rows expose selected state. Native grouping/focus and announcements remain unverified. |
| Motion | No custom animation; native sheet transitions and spinner need reduced-motion review. |
| Content | Household name, inventory names, roles and update summaries form the hierarchy. Empty household has explicit copy. No pagination implemented; high collection counts need runtime evidence. |
| Search | N/A for current defined chooser; no search requirement or large-collection evidence established. |
| Loading | Initial named spinner, cached content and pending selection are separate. A pending selection disables inventory choices and household mode command. |
| Recovery | Initial load Retry and safe selection error; failed selection stays open. Background stale-membership notice remains an unverified risk described above. |
| Editing | No editable draft. Household browsing does not switch inventory until a row is chosen. Cancellation does not promise rollback of accepted repository changes. |
| Privacy | Scoped dashboard hook suppresses denied data. Existing command and repository determine authorization; UI review does not replace boundary tests. |
| Notifications | N/A: no OS scheduling or notification entry owned here. |
| Media | N/A: no capture, picker or playback. |
| Lifecycle | M184 retires selection callbacks per focused visit and session/command identity. Old callbacks cannot begin work after blur/refocus. Close retires immediately. Existing cancellation/completion tests remain. Native cold entry/backgrounding and return remain open. |

M184's mounted RED case reproduced a retained callback starting a selection after
blur. Its fix preserves fresh selections after return. A second case verifies
immediate retirement on Close and duplicate Close suppression, before native blur.
All10 focused cases, TypeScript and structural checks pass remotely on paul
(`/tmp/switcher-retained-reviewed.log`); critic found no blockers. Native acceptance
remains open.
