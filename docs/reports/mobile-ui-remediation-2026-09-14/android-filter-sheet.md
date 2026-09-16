# Android filter sheet inspection (in progress)

## Retained candidate search-clear follow-up

On September16, APK `d9b1a7b5196c970f45131b6d960adea3ae02bc46ea5f4d772dff39ec8a00bb72`
on the same normal-size emulator opens Browse Tags, accepts `Tools`, and exposes
the matching choice with Show results/Back above Gboard. Selecting Tools, deleting
the five search characters, then choosing Back restores the overview with
`1 selected`. The empty search field stayed focused and the full tag list returned.
Hierarchies on paul: `/tmp/filter-{typed,clear,overview}.xml`; inspected screenshot
`/tmp/filter-current.png` on both hosts. This adds search-clear/draft-retention
evidence, not dark appearance, TalkBack or production query acceptance.

Cold-link Apply was not accepted by this fixture's `router.back()` callback because
there was no previous route. The production Browse Filters route instead verifies
the inventory scope and calls `router.dismissTo('/search', params)`; this fixture
result does not establish a production navigation defect. Warm-entry Apply retains
the earlier explicit selected-tag result evidence below.

The same APK was then exercised with Android system night mode enabled. On warm
entry, typing Tools after the field was ready, selecting its row and tapping Show
results above the keyboard returned the explicit `Browse selected tags: audit-tools`
fixture result. The first input command issued during keyboard opening left the
field empty; this observation is retained rather than counted as successful typing.
The inspected dark body keeps search, rows and footer labels visible. Hierarchies:
`/tmp/dark-{tools,result}.xml` on paul; capture `/tmp/dark-tools.png` on both hosts.
That capture precedes the successful second input command. The fixture's white
navigation header lacks production `_layout.tsx`'s palette-backed `headerStyle`,
so it does not prove production dark-header behavior. System night mode was restored
to light after this check; assistive and production-data acceptance remain open.

TalkBack was subsequently enabled on this emulator; `dumpsys accessibility`
confirmed its bound service and touch exploration. A green focus outline appeared
on Filters and then the selected All types menu entry. ADB-injected horizontal
gestures did not establish traversal, and a single injected tap directly opened
the menu. These inputs are insufficient to certify TalkBack gesture activation or
focus order. No spoken-output claim is made. Captures are retained as
`/tmp/talkback-{focus,next,explore}.png` on both hosts. The originally empty enabled
service setting and disabled accessibility state were restored and read back.

### Fixture appearance correction

The runner-only root now supplies production's existing header surface, title
color/weight and resolved-theme StatusBar defaults. Production code is unchanged.
Six fixture-preparation tests, TypeScript and structural checks pass on paul;
critic review found no blocker. Rebuilt APK
`114130346cc27a21d098dc90b662b463a87a9af4562756c1cf2aba17bec1b50c`
shows the corrected [dark header](evidence/android-fixture-header-dark.png) and
[light header](evidence/android-fixture-header-light.png), with matching status
glyphs and readable filter actions. The first cold-start light capture preceded
content readiness; the retained light capture is after restoring light mode from
the loaded dark view. This verifies the Android fixture's theme change, not iOS
appearance or the remaining filter interaction/accessibility checks.

Android16 Pixel6 emulator on paul, normal text, light appearance. The initial
Browse filter sheet at detent0.7 lacks its title and both footer commands;
expanding to1.0 reveals Show results/Cancel. The ADB assertion that Show results
exists fails at initial presentation. [Before](evidence/android-filters-initial-before.png).

Pinned native-stack documents Android form-sheet headers as unsupported and
provides `unstable_sheetFooter` because JS-only footer placement is problematic.
The Android-specific candidate uses that native footer, measured scroll clearance,
an accessible body title, stable presentation and current committed callbacks.
iOS retains its existing direct-scroll layout.
[Guidance](https://reactnavigation.org/docs/native-stack-navigator/).

The first patched warm sheet shows title and both actions at initial0.7:
[capture](evidence/android-filters-native-footer.png). That is rendering evidence,
not full apply/cancel, last-row or keyboard acceptance.
Warm candidate APK SHA-256:
`3ca19be05b5bcd0016650fd96b4beddfef63d4f760ef22c6b5dcdefb96d6fd9a`.
Later root-fallback APK SHA-256:
`7fcf72fc844f3dcc777beeb0123ddaf473c898927705a17b668e29ca4459b942`.
Both are db8d2bb8 synthetic archives with the reviewed icon fix and the then-current
uncommitted filter adapter; neither includes the latest top-inset correction.

Cold-link verification exposed a first-route distinction: Android renders a root
screen rather than a sheet, so the native sheet footer disappears. A full-screen
footer fallback now renders the actions, but its inspected title overlapped the
system area. Top inset reservation has been added and needs a new native build.
Code review also caught competing headerShown=true assignments in the Expiration
route/page; these were removed so the adapter owns Android visibility and iOS uses
stack registration. Native re-verification of those latest changes is pending.

Fourteen focused tests, TypeScript and structural checks passed on paul before the
latest top-inset adjustment. Root footer expectations failed before the fallback;
the first test attempt also exposed a missing useRouter export in the navigation
fake, subsequently corrected. Review has not yet accepted the complete final pass.

Remaining checks: cold-root top inset; root Cancel escape (current production
routes call router.back unconditionally); initial/expanded detents; last tag row;
nested-page search without native sheet headers; draft apply/cancel; keyboard;
Expiration parent rerenders; Android dark appearance and assistive technology.

## Latest native pass

The current APK (`0789c20cf7ab484362c7b168355c0d851a0bd07295e184c729aca5950de06dbc`)
adds top inset, consistent Android header ownership, and root-entry Cancel recovery.
It remains the db8d2bb8 synthetic archive plus icon and filter patches, not a release.
Cold Filters title starts at y128; both actions are visible, and tapping Cancel
returns to the audit index. [Root capture](evidence/android-filter-root-after.png).
Warm initial detent shows title/actions. Scrolling Tags expands the sheet; the last
row is above the footer, can be selected, and Show results returns the confirmed
fixture result `Browse selected tags: audit-last`.

All 1,873 mobile tests in 290 files pass on paul, as do TypeScript and structural
checks. Final critic review found no further source blocker. These checks do not
close full filter acceptance: native Tags search is absent in the headerless Android
sheet (M221). Expiration, keyboard, dark appearance and assistive technology remain.
The white status glyphs in the synthetic root capture are not production evidence:
the fixture omits production StatusBar configuration.

## Keyboard rejection of the footer candidate

Adding in-body search reveals a native crash on focus: RNScreens
`ScreenFooter.onParentLayout` calls `sheetTopInStableState` while the bottom sheet
is settling during keyboard inset changes. AndroidRuntime reports
`IllegalArgumentException: [RNScreens] use of stable-state method for unstable state`.
The process exits to the launcher. The pinned implementation confirms this call
has no dragging/settling branch. The initial/last-row successes above do not
establish keyboard safety, and this candidate must not ship as accepted.

The in-body search has a failing-before/passing-after interaction test; critic also
caught stale page events, now guarded with a keyed input, committed handler ref,
focus cleanup and unmount cleanup. Tests verify hidden events are ignored, current
input recovers on focus, and removed-page events cannot restore old queries.
These guards are not a fix for the native footer crash. The next candidate is an
Android full-screen native stack presentation with a layout-owned footer and
keyboard handling, preserving iOS sheets. It requires spec and native acceptance.

## Replacement presentation and keyboard checks

The replacement uses Android native-stack `card` presentation, a normal title bar,
in-body search and a flex-layout footer. It removes `unstable_sheetFooter` entirely.
Initial native typing no longer crashed, but Gboard covered the footer; an explicit
KeyboardAvoidingView with the native header offset corrected that.

APK SHA-256 `36fe660b24f35a8b9b7d8b2d98f4feac9d03495ae95392807fed71eac5adbe92`
(db8d2bb8 fixture archive plus current icon/filter patches) was tested at normal
size on the same Android16 emulator. Browse and Expiration Tags both accept
`Tools`, narrow to the matching row, and keep primary/Back controls above Gboard.
Browse Back removes search and restores the filter overview. Expiration selection
and Apply return to the prior Browse filter route in the fixture stack. [Keyboard capture](evidence/android-filter-keyboard-after.png).
This is synthetic interaction evidence, not production backend acceptance.

The final full source suite before the KeyboardAvoidingView adjustment passed
1,876 tests in291 files; the final wrapper passed the3 focused filter tests,
TypeScript and structural checks on paul. Broader search/clear/last-row retesting,
dark mode, TalkBack and production navigation remain open; iOS search placement
must still be verified. The earlier native-footer candidate is superseded, not an
accepted fallback.

Replacement-layout follow-up on the same36fe660b APK: after warm Browse entry,
Tags keeps search at y317–443 while the list scrolls independently. The final tag
row is y1812–1949, above Show results y2075–2118. Selecting that row and applying
returns the explicit fixture result `Browse selected tags: audit-last`. This
supersedes the earlier sheet-layout last-row evidence for the new implementation.
Local inspected hierarchies: `/tmp/card-final-row.xml`, `/tmp/card-final-result.xml`.

Date-range check on36fe660b: Android opens its system calendar dialog. Selecting
October20 and OK replaces October15 in the range. Selecting another date and
Cancel preserves October20. Clear date range, then Back, returns the overview with
Any date. These observations cover local draft controls, not backend query results.
[Selected date](evidence/android-date-range-selected.png).
