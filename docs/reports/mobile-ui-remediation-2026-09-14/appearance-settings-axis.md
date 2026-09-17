# Appearance settings interaction review

R025, source55fb2cd9, September15. Sources: appearance route,
AppearanceSettingsScreen, AppearancePicker, native choice adapters,
AppearanceContext, AppearancePreferenceController and JSON preference adapter.
The inline Settings consumer shares the picker; this table does not certify the
whole Settings screen. Shared materials review remains in appearance-axis.md.

| Axis | Source evidence and acceptance limits |
| --- | --- |
| Task | Three exclusive appearance choices use an in-place native menu. No extra navigation is needed to change the value. The legacy detail route exposes the same picker. |
| Navigation | Detail route inherits stack Back and title Appearance. Inline Settings needs no navigation. Native Back and menu dismissal still need verification. |
| Selection | System, Light and Dark map to typed preferences; invalid or unchanged values are ignored. Menu carries current selection; source tests cover persistence and rollback. |
| Modality | Choice uses a native menu, not a new modal editor. iPad popover anchoring and dismissal remain native checks. |
| Layout | Detail screen scrolls with shared settings spacing. Native Host measures vertical content and fills width. Safe area and row geometry remain unverified on devices. |
| Adaptation | iOS normal-size choice is labeled content; larger text stacks the label. Android uses native action menu. Compact width and tablet presentation remain pending. |
| Typography | Native picker text wraps through fixedSize modifiers. Shared explanatory footer uses settings styles. Normal-size long localized labels require testing; enlarged-text remediation is later. |
| Appearance | Provider resolves explicit or system preference, applies native scheme and semantic palette. Shared contrast handling is audited separately. Live transitions with an open menu need native evidence. |
| Localization | Labels/footer are English. System follows OS appearance, independent of timezone. Translated lengths and RTL alignment remain unverified. |
| Imagery | No photos or custom decorative icons. Selection checkmarks belong to platform menus. Native selected-state rendering remains pending. |
| Targets | Native picker host has minHeight48. Actual menu trigger and option hit regions require runtime verification. |
| Gestures | Selection and dismissal use native menu interaction; no custom gesture is required. Assistive alternatives need runtime testing. |
| Keyboard | No text input. External keyboard menu selection and escape/return behavior are not established. |
| Accessibility | Native trigger has Choose appearance label and current selection; footer explains System. VoiceOver/TalkBack traversal and announcement of changed appearance remain pending. |
| Motion | No custom animation in the picker. OS appearance/menu transitions and Reduce Motion must be checked natively. |
| Content | Three static options need no pagination, search or remote list. No inventory content is shown. |
| Search | No search is needed for three fixed choices. No filtering state is owned by this surface. |
| Loading | Root waits for appearance hydration. Preference changes apply optimistically and persistence is serialized; no blocking spinner. Cold-start fallback and failed storage load remain native observations. |
| Recovery | Latest failed save rolls back through provider and offers notice; user can choose again. M127 suppresses obsolete failure feedback. Actual error placement remains pending. |
| Editing | Immediate preference has no Save/Cancel draft. Queued writes retain order; a failed older write does not roll back a newer preference. Current failure can revert to persisted preference. |
| Privacy | Device-local preference has no inventory/account data. JSON adapter validates loaded values. This review adds no authentication boundary or server request. |
| Notifications | No notification task is owned here. Notification-driven departure must obey M127 visit boundary; physical push interruption remains unverified. |
| Media | No camera, upload or audio operation is owned here. Covering global voice interactions remain a native interruption scenario. |
| Lifecycle | M127 binds failure to current selection and focused visit, including blur/refocus before rejection. Already-started persistence continues. Native focus delivery and app background transitions remain pending. |

Four picker tests and four controller cases cover the ownership/persistence slice;
source evidence is not native acceptance. No whole-app appearance pass is claimed.

## Inline Settings selection follow-up

S104 reviewed at f80a4b7e against the same24 axes above. SettingsScreen renders
AppearancePicker directly inside SettingsSection with separators; it does not
wrap it in a navigation row. The dedicated route adds an explanatory System
footer; inline selection has no such footer. Its three familiar labels remain
visible through the native picker. This difference is not proof of a defect.

Current provider/store review confirms serialized persistence, latest-failure
rollback and native scheme application. The trigger rejects invalid/unchanged
choices and suppresses failed-save notices after a newer choice or departure.
The table's route Back statements apply only to R025; S104 selection stays in
Settings. All other shared control axes apply with the Settings section as parent.
Native geometry, screen-reader focus and open-menu appearance changes remain
pending. Preserve the earlier limited native selection result without promoting
the rest of this surface to runtime verified.

## Android normal-text menu acceptance and M243/M244

APK31457c4d reproduced two independent failures: a stretched accessible trigger
whose center missed the visible button, then a visible Dark option tap that left
System selected and the menu open. NativeActionMenu now sizes its wrapper around
the measured control and retains DropdownMenuItem onClick for selectable items.
The selected modifier remains for accessibility; disabled guards are unchanged.
Reviewed shared consumers: Android NativeChoicePicker (including reminder mode),
AssetOverflowMenu, ExpirationWorkspaceScreen and AssetHistoryRouteScreen.

Candidate APK SHA256
`b76e227a0c215b08da22f929024514b4b0721f9781e1d209811aaff877d1b4c1`
was installed on Pixel6/API36,1080×2400,420dpi,font1. Native target-center taps now
open the menu. Dark, Light and System changes close it and persist exactly once:
a fixture-only store counter records1,2,3. Back dismisses without changing Dark.
An open System menu adapts when the OS changes to dark. Light mode was restored.
The underlying picker, menu, Compose host and appearance-provider source hashes
matched HEAD before the candidate patch; the APK remains a selectively patched
historical disposable tree, not a full HEAD build.

[Dark selected](evidence/android-appearance-dark-selected.png) and
[open-menu system change](evidence/android-appearance-open-theme-change.png).
Unstyled black diagnostic text belongs to this runner-only fixture; these captures
do not certify production Settings layout or its explanatory text. Native TalkBack
selection, keyboard input and iPad/iOS checks remain separate.

Nine focused menu/choice/appearance checks, TypeScript, six fixture installer tests
and structural validation pass on paul. Critic found no blocker and specifically
requested exactly-once native validation, now satisfied by the fixture counter.
Native reproduction/check scripts remain `/tmp/verify-appearance-menu.py` and
`/tmp/verify-appearance-choices.py` on paul, along with captures and XML.

The shared reminder-mode consumer also passes actual Off and Use defaults selection
and immediate result reconciliation in the same APK (`/tmp/verify-reminder-menu.py`).
