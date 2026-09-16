# Android asset action inspection

Android16 Pixel6 emulator, normal text, light appearance, synthetic data on paul.
Move Here initially used a0.6 sheet detent and rendered recovery/search/preview
without either footer action. [Before](evidence/android-move-initial-before.png).

Edit, Move and Move Here now use Android native stack cards; iOS sheets retain their
existing detents/header behavior. The shared form wrapper resizes for the keyboard
using native header height. These are the only production consumers of those forms.

Review found that exposing Edit's native header would bypass explicit Cancel's
unsaved-draft confirmation. The shared operation guard now delegates idle Edit
removal to its existing guarded close handler. Pending work still prevents exit;
authorized Discard and successful Save can dispatch the original removal. Two new
GO_BACK/POP tests failed first, then passed with Keep editing/double-confirm checks.

APK d93b7c66552adbff511fa487813dc06e91091c02d98f86d238f6ab4430f4fcc3
uses the db8d2bb8 isolated fixture baseline plus recorded icon/filter/action patches.
Native hardware Back and header Back both display Discard changes after editing
the name. Keep editing retains `Audit tentchanged`; Discard returns to the audit
index. [Native confirmation](evidence/android-edit-native-discard.png).
Move Here displays Move here/Cancel immediately, and after typing `tent` keeps both
above Gboard (text bounds y1223–1266 and1370–1413). The failed suggestion state
remains independently recoverable.

The full mobile suite passes1,879 tests/291 files on paul;133 focused tests,
TypeScript and structural checks also pass. Critic accepted the navigation guard.
A final Android-only duplicate-body-title adjustment passed TypeScript/structural
checks and awaits rebuilt native verification; the initial native heading assertion
failed with two copies. Neither this report nor that suite certifies the remaining
Move destination/create flow, all Edit metadata states, dark mode, TalkBack or iOS.

Final title-only rebuildab582c2f1b91992a82880f4ce077411696a471abc11b596ce5dee1c5fbc1b749
passes the native assertion: exactly one Move something here heading and both
Move here/Cancel present. The earlier failing two-heading assertion is retained in
`/tmp/asset-heading-red.log`. Critic found no blocker in that final adjustment.

## Edit metadata and rejected-save recovery

On the final ab582c2f APK, both initial metadata errors expose separate retry
commands. Retrying types clears its error while retaining the tags error. After
changing the name to `Audit tentchanged`, retrying tags clears its error without
losing that draft. Scrolling exposes the tag-color controls and Add tag above the
Save/Cancel commands. Save reaches the fixture's intentionally rejected command
and displays Could not save changes; dismissing that alert preserves the edited
name and both footer commands. The inspected
[recovered form](evidence/android-edit-recovered.png) confirms the lower controls
remain visually separate from the actions. This verifies error recovery through
synthetic ports; successful backend persistence and every metadata variant remain
outside this sample. XML snapshots are retained at
`/tmp/edit-recovery-android.xml`, `/tmp/edit-retry-android.xml`,
`/tmp/edit-bottom-android.xml`, `/tmp/edit-save-error-android.xml` and
`/tmp/edit-recovered-android.xml`.

## Move Here completion teardown (M231)

The September16 normal-text sample recovered suggestions, selected Audit tent,
rejected the command and preserved the selection/preview and both actions.
[Retained selection](evidence/android-movehere-rejected-retained.png). Cancel
returned to the audit index. The fixture previously rejected every command;
it now checks the expected source/target IDs, rejects once and accepts retry.

That successful retry exposed a real native crash at07:40:21 emulator time,
PID11572: `ScreenStackFragment added into a non-stack container` from
`ScreenStackHeaderConfig.onUpdate`. The destination assertion failed and the
captured hierarchy was the Android launcher. This is a failed candidate, not
successful navigation. Log: `/tmp/movehere-success-crash.log`.

The shared asset-operation guard now stays registered throughout form lifetime,
redispatches authorized completions and idle Move exits, preserves Edit discard
confirmation, and blocks pending writes. Rebuilt APK
`e37351994c5f6f398188fcdd829cddb499f6d1501afc4e34ccc2c16d29eddd28`
passes the same failure/retention/retry sequence and returns to Native UI audit;
app PID11822 remains alive. Hierarchies:
`/tmp/movehere-fixed-{entry,retry,selected,error,retained,return}.xml`.

All38 shared action behavior tests, TypeScript and structural checks pass on paul.
Tests include pending native removal and idle removal after rejected Move/Move Here.
Critic accepted source semantics and required positive destination interaction in
the new normal-text iOS regression. That Swift test is pending macOS execution;
iOS idle dismissal gestures, Move destination creation, successful Edit and real
backend persistence remain separate unverified scenarios. No production inventory
was changed. Build log: `/tmp/android-move-guard-build.log`; source checks:
`/tmp/asset-removal-guard-{tests,check,structural}.log`.
