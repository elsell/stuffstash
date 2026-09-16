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
