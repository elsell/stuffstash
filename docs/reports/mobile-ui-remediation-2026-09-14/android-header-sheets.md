# Android inventory, history and Return details

M228: native normal-size inspection of inventory switching and checkout history
found their requested titles/Close commands absent from Android form sheets.
Android now uses native-stack cards; iOS retains the existing detents. Inventory
presentation is shared by production and fixture. Checkout history's shared options
also affect Home Return details, so that consumer was explicitly exercised.

The [inventory switcher](evidence/android-switcher-header.png) and
[checkout history](evidence/android-history-header.png) show the restored headers
on APK256fdd0c03d828b6a52c2b1e16eb234e26ca6e5cb5ccb924295afbbc71c9666e.
Native root-switcher Close reaches the fixture root index. Mounted tests additionally
reproduce then fix root selection/Close back-only navigation using the shared Home
fallback. Existing callback retirement, cancellation and failed-selection retry
remain covered. Checkout-history Close permits return to the preceding screen.

## Return details crash discovered and resolved in this candidate

Return details has visible Cancel return/Save above Gboard at normal size
([keyboard capture](evidence/android-return-keyboard.png)). Its first synthetic Save
fails and preserves the note. Retrying initially crashed APK256fdd at07:15:47:
`ScreenStackFragment.canNavigateBack` reports a non-stack container during header
update. The negative-only assertion that Return details disappeared initially
passed; inspecting the positive destination revealed the Android launcher. That
result was rejected. No successful-navigation claim is based on disappearance.

M229: the route toggled `usePreventRemove(Boolean(task))` in the same completion
commit that clears the task and pops the route. The candidate keeps the removal
guard installed for the route lifetime. Active tasks still own close/undo behavior;
a completed/absent task redispatches the original removal action. This avoids that
native configuration transition without a delay or bypass of active-task guards.

Rebuilt APK45060ca1ec90320b951701c8e53356d75e85f57d53bf621581bb948c65848b3f
passes the same keyboard-edit, failed Save, retained note and successful retry
journey. The destination positively contains Recently changed, no Return details,
and no Return Audit drill. App PID10308 remains alive with no fatal entry for that
process. A fresh Cancel return journey positively restores Return Audit drill and
Recently changed on Home. This is fixture verification, not real backend mutation.
The iOS Save-recovery test now also requires runningForeground and Home content;
its next native execution is pending.

## Evidence and limits

Both presentation tests failed first. The full mobile suite passes1,884 tests across
292 files before the final stable-removal correction; its3 focused route tests,
TypeScript and structural checks pass afterward. Code critic found no confirmed
blocker. Builds/tests run on paul. Native iOS, enlarged text, TalkBack, all history
pagination states and real inventory selection remain outside these samples.

Retained evidence: `/tmp/android-inventory-switcher.xml`,
`/tmp/android-checkout-history.xml`, `/tmp/android-switcher-{after,closed}.xml`,
`/tmp/android-history-after.xml`, `/tmp/android-return-crash.log`,
`/tmp/android-return-stable-{entry,error,saved,cancelled}.xml`,
`/tmp/android-return-after-fix-crash-buffer.log`, `/tmp/android-header-batch-full.log`,
`/tmp/return-stable-guard-{check,structural,build}.log`.

## History recovery, pagination and root Close follow-up

September16, Android16 Pixel6, normal text/light appearance, APK
`e37351994c5f6f398188fcdd829cddb499f6d1501afc4e34ccc2c16d29eddd28`:
three history records remain visible while the asset-name query fails. Scrolling
reveals the separate name Retry and Load older checkouts commands. Retry removes
the error; scrolling back positively confirms the recovered title Audit ladder.
Load older appends [Older audit checkout](evidence/android-history-older-page.png)
and its return note while retaining the preceding records. The exhausted cursor
removes the load-more command. Both older notes fit above the system bottom area.

Native Close returns to the audit index after pagination. A separate force-stop
and cold URL entry, without launching the index first, also exposes Close and
returns positively to the root audit index. These exercise the production route
with synthetic read ports; production asset-detail focus restoration, denied
reads, failed continuation, larger text and TalkBack remain outside this sample.
The existing duplicated Returned title/status is still tracked for presentation
and assistive-reading review; this sample does not certify all history styling.

Evidence: `/tmp/history-pagination-{entry,bottom,recovered,older,name,close}.xml`,
`/tmp/history-root-{entry,close}.xml`, `/tmp/history-pagination-older.png`.
No source implementation changed for this verification.
